package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/repository"
)

// autoManagedPrefixes は自動管理タグのプレフィックス一覧。手動付与・解除を禁止する（BE-019）。
var autoManagedPrefixes = []string{
	"cpm_", "last_visit_", "first_visit_", "ltv_",
	"visit_count_", "vaccine_", "checkup_", "next_checkup_", "refill_due_",
	"next_visit_", "reserved_", "canceled_visit",
	"no_show_", "breed_",
	"sex_", "pet_birthday_", "birth_year_",
	"has_dog", "has_cat", "has_both",
	"spay_neutered", "intact", "chronic_", "dormant_",
	"cert_sent_", "post_surgery_", "post_discharge_",
	"pet_deceased_",
}

// isAutoManagedTag は指定タグが自動管理タグかどうかを判定する。
func isAutoManagedTag(tagName string) bool {
	for _, prefix := range autoManagedPrefixes {
		if tagName == prefix || strings.HasPrefix(tagName, prefix) {
			return true
		}
	}
	return false
}

// OwnerTagsResult は飼い主タグ一覧の取得結果。
type OwnerTagsResult struct {
	LineUserID  *string
	IsLinked    bool
	LstepOptOut bool
	Tags        []string
	FetchedAt   time.Time
}

// BulkAddOwnerTagResult は一括タグ付与の結果。
type BulkAddOwnerTagResult struct {
	SyncedCount    int
	SkippedCount   int
	FailedOwnerIDs []uint64
}

// LstepTagService は飼い主タグの手動 CRUD インターフェース（BE-019）。
type LstepTagService interface {
	// GetOwnerTags は飼い主の現在のLステップタグ一覧を返す。LINE未連携時は空リスト。
	GetOwnerTags(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error)
	// AddOwnerTag は飼い主に手動でタグを付与する。
	AddOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
	// RemoveOwnerTag は飼い主から手動でタグを解除する。冪等（存在しないタグは正常終了）。
	RemoveOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
	// BulkAddOwnerTag は複数飼い主に同一タグをベストエフォートで付与する。
	// LINE未連携/opt-outはスキップ扱い（エラーにしない）。
	BulkAddOwnerTag(ctx context.Context, clinicID uint64, ownerIDs []uint64, tagName string, actorID *uint64) (*BulkAddOwnerTagResult, error)
}

type lstepTagService struct {
	settingsSvc  LstepSettingsService
	ownerRepo    repository.OwnerRepository
	tagCacheRepo repository.LstepTagCacheRepository
	auditSvc     AuditService
}

// NewLstepTagService は LstepTagService を初期化して返す。
func NewLstepTagService(
	settingsSvc LstepSettingsService,
	ownerRepo repository.OwnerRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	auditSvc AuditService,
) LstepTagService {
	return &lstepTagService{
		settingsSvc:  settingsSvc,
		ownerRepo:    ownerRepo,
		tagCacheRepo: tagCacheRepo,
		auditSvc:     auditSvc,
	}
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効（is_sync_enabled=false）または API キー未設定の場合は nil, nil を返す。
func (s *lstepTagService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	enabled, err := s.settingsSvc.IsSyncEnabled(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check lstep sync enabled")
	}
	if !enabled {
		return nil, nil
	}
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

func (s *lstepTagService) GetOwnerTags(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find owner")
	}

	result := &OwnerTagsResult{
		LineUserID:  owner.LineUserID,
		IsLinked:    owner.LineUserID != nil && *owner.LineUserID != "",
		LstepOptOut: owner.LstepOptOut,
		Tags:        []string{},
		FetchedAt:   time.Now(),
	}

	// LINE User ID 未連携は空リスト（エラーではない）
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return result, nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	// API 未設定時は DB キャッシュからフォールバック
	if client == nil {
		cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
		if cacheErr != nil {
			slog.ErrorContext(ctx, "failed to find tag cache", "error", cacheErr, "owner_id", ownerID)
			return result, nil
		}
		for _, c := range cached {
			result.Tags = append(result.Tags, c.TagName)
		}
		return result, nil
	}

	tags, err := client.GetUserTags(ctx, *owner.LineUserID)
	if err != nil {
		if lstep.IsUserNotFound(err) {
			return result, nil
		}
		slog.ErrorContext(ctx, "failed to get lstep tags", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return nil, apperrors.WrapInternalServerError("Lステップ API からタグ取得に失敗しました")
	}

	if tags != nil {
		result.Tags = tags
	}
	return result, nil
}

func (s *lstepTagService) AddOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	if isAutoManagedTag(tagName) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("タグ %q は自動管理タグのため手動付与できません", tagName))
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LstepOptOut {
		return apperrors.WrapForbidden(fmt.Sprintf("飼い主 %d はLステップ配信をオプトアウトしています", ownerID))
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return apperrors.WrapNotFound("owner", fmt.Sprintf("%d:line_user_id", ownerID))
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return apperrors.WrapInvalidInput("Lステップ API が設定されていません")
	}

	if err := client.AddTag(ctx, *owner.LineUserID, tagName); err != nil {
		slog.ErrorContext(ctx, "failed to add lstep tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "tag", tagName)
		return apperrors.Wrap(err, "failed to add tag")
	}

	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "manual"); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert tag cache after add", "error", cacheErr, "tag", tagName)
	}

	_ = s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "add_tag", "owner", &ownerID)

	return nil
}

func (s *lstepTagService) RemoveOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	if isAutoManagedTag(tagName) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("タグ %q は自動管理タグのため手動解除できません", tagName))
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}

	// LINE 未連携は冪等に正常終了
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	if err := client.RemoveTag(ctx, *owner.LineUserID, tagName); err != nil {
		if lstep.IsUserNotFound(err) {
			return nil
		}
		slog.ErrorContext(ctx, "failed to remove lstep tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "tag", tagName)
		return apperrors.Wrap(err, "failed to remove tag")
	}

	if cacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, tagName); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to delete tag cache after remove", "error", cacheErr, "tag", tagName)
	}

	_ = s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "remove_tag", "owner", &ownerID)

	return nil
}

func (s *lstepTagService) BulkAddOwnerTag(ctx context.Context, clinicID uint64, ownerIDs []uint64, tagName string, actorID *uint64) (*BulkAddOwnerTagResult, error) {
	if isAutoManagedTag(tagName) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("タグ %q は自動管理タグのため手動付与できません", tagName))
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, apperrors.WrapInvalidInput("Lステップ API が設定されていません")
	}

	result := &BulkAddOwnerTagResult{FailedOwnerIDs: []uint64{}}
	for _, ownerID := range ownerIDs {
		owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if findErr != nil {
			slog.ErrorContext(ctx, "bulk tag: owner not found", "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			continue
		}
		if owner.LstepOptOut || owner.LineUserID == nil || *owner.LineUserID == "" {
			result.SkippedCount++
			continue
		}

		if addErr := client.AddTag(ctx, *owner.LineUserID, tagName); addErr != nil {
			slog.ErrorContext(ctx, "bulk tag: failed to add lstep tag", "error", addErr, "owner_id", ownerID, "tag", tagName)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			continue
		}

		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "manual"); cacheErr != nil {
			slog.ErrorContext(ctx, "bulk tag: failed to upsert tag cache", "error", cacheErr, "owner_id", ownerID, "tag", tagName)
		}
		result.SyncedCount++
	}

	// ISSUE-010: LTV/CPM 一括同期や健診一括タグ付与の件数を audit_logs.metadata に永続化する。
	failedCount := len(result.FailedOwnerIDs)
	_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"bulk_add_tag", "owner", nil,
		map[string]any{
			"operation":       "bulk_add_tag",
			"tag_name":        tagName,
			"requested_count": len(ownerIDs),
			"synced_count":    result.SyncedCount,
			"skipped_count":   result.SkippedCount,
			"failed_count":    failedCount,
		},
	)

	return result, nil
}
