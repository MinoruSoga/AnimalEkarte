package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// systemManagedPrefixes はシステム固定の自動管理タグプレフィックス一覧（BE-019）。
// B/C1/C2/C3 カテゴリは DB 管理 (lstep_auto_managed_prefixes) に移行済み。
var systemManagedPrefixes = []string{
	"CPM_", "LTV_", "VISIT_", "PET_", "HLTH_", "PREV_", "EXCL_",
	"cpm_", "last_visit_", "first_visit_", "ltv_",
	"visit_count_", "dormant_", "has_dog", "has_cat", "has_both",
}

// isSystemManagedTag はシステム固定プレフィックスに一致するか判定する。
func isSystemManagedTag(tagName string) bool {
	for _, prefix := range systemManagedPrefixes {
		if tagName == prefix || strings.HasPrefix(tagName, prefix) {
			return true
		}
	}
	return false
}

// IsSystemManagedTag reports whether a tag is reserved for automatic management.
// It is exported for the remaining checkup-sync consumer during BE9-2C.
func IsSystemManagedTag(tagName string) bool {
	return isSystemManagedTag(tagName)
}

// isAutoManagedTagWithPrefixes はシステム固定 + DB 登録プレフィックスに一致するか判定する（純粋関数）。
func isAutoManagedTagWithPrefixes(tagName string, dbPrefixes []*model.LstepAutoManagedPrefix) bool {
	if isSystemManagedTag(tagName) {
		return true
	}
	for _, p := range dbPrefixes {
		if tagName == p.Prefix || strings.HasPrefix(tagName, p.Prefix) {
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

// LstepTagService は飼い主タグの手動 CRUD インターフェース（BE-019）。
type LstepTagService interface {
	// GetOwnerTags は飼い主の現在のLステップタグ一覧を返す。LINE未連携時は空リスト。
	GetOwnerTags(ctx context.Context, clinicID, ownerID uint64) (*OwnerTagsResult, error)
	// AddOwnerTag は飼い主に手動でタグを付与する。
	AddOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
	// RemoveOwnerTag は飼い主から手動でタグを解除する。冪等（存在しないタグは正常終了）。
	RemoveOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error
}

type lstepTagService struct {
	settingsSvc   LstepSettingsService
	ownerRepo     tagOwnerFinder
	tagCacheRepo  LstepTagCacheRepository
	auditSvc      lstepAuditLogger
	tagConfigRepo LstepTagConfigRepository
	// buildClientFn is a test hook to inject a Client without the real deploy write gate.
	buildClientFn func(ctx context.Context, clinicID uint64) (lstep.Client, error)
}

// NewLstepTagService は LstepTagService を初期化して返す。
// tagConfigRepo が nil の場合はシステム固定プレフィックスのみチェックする。
func NewLstepTagService(
	settingsSvc LstepSettingsService,
	ownerRepo tagOwnerFinder,
	tagCacheRepo LstepTagCacheRepository,
	auditSvc lstepAuditLogger,
	tagConfigRepo LstepTagConfigRepository,
) LstepTagService {
	return &lstepTagService{
		settingsSvc:   settingsSvc,
		ownerRepo:     ownerRepo,
		tagCacheRepo:  tagCacheRepo,
		auditSvc:      auditSvc,
		tagConfigRepo: tagConfigRepo,
	}
}

// isAutoManagedTag はシステム固定 + DB 登録プレフィックスに一致するか判定する。
// tagConfigRepo が nil の場合はシステム固定プレフィックスのみチェックする。
func (s *lstepTagService) isAutoManagedTag(ctx context.Context, tagName string) (bool, error) {
	if isSystemManagedTag(tagName) {
		return true, nil
	}
	if s.tagConfigRepo == nil {
		return false, nil
	}
	prefixes, err := s.tagConfigRepo.FindAllAutoManagedPrefixes(ctx)
	if err != nil {
		return false, apperrors.Wrap(err, "failed to load auto managed prefixes")
	}
	return isAutoManagedTagWithPrefixes(tagName, prefixes), nil
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効（is_sync_enabled=false）または API キー未設定の場合は nil, nil を返す。
func (s *lstepTagService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	if s.buildClientFn != nil {
		return s.buildClientFn(ctx, clinicID)
	}
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
	return newLstepAPIClient(apiKey, baseURL), nil
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
			// LSA-10: do not disguise cache DB failure as "zero tags".
			return nil, apperrors.Wrap(cacheErr, "failed to load lstep tag cache")
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
		return nil, apperrors.WrapInternalServerError("Lステップ API からタグ取得に失敗しました")
	}

	if tags != nil {
		result.Tags = tags
	}
	return result, nil
}

func (s *lstepTagService) AddOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	managed, err := s.isAutoManagedTag(ctx, tagName)
	if err != nil {
		return apperrors.Wrap(err, "failed to check auto managed tag")
	}
	if managed {
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
		return apperrors.Wrap(err, "failed to add tag")
	}

	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "manual", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert tag cache after add", "error", cacheErr, "tag", tagName)
	}

	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "add_tag", "owner", &ownerID); err != nil {
		slog.WarnContext(ctx, "audit log failed for add tag", "error", err, "owner_id", ownerID, "tag", tagName)
	}

	return nil
}

func (s *lstepTagService) RemoveOwnerTag(ctx context.Context, clinicID, ownerID uint64, tagName string, actorID *uint64) error {
	managed, err := s.isAutoManagedTag(ctx, tagName)
	if err != nil {
		return apperrors.Wrap(err, "failed to check auto managed tag")
	}
	if managed {
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
		return apperrors.Wrap(err, "failed to remove tag")
	}

	if cacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, tagName); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to delete tag cache after remove", "error", cacheErr, "tag", tagName)
	}

	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "remove_tag", "owner", &ownerID); err != nil {
		slog.WarnContext(ctx, "audit log failed for remove tag", "error", err, "owner_id", ownerID, "tag", tagName)
	}

	return nil
}
