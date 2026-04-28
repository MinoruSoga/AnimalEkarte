package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CheckupSyncPreviewOwner はプレビュー一覧の1件。
type CheckupSyncPreviewOwner struct {
	OwnerID       uint64
	OwnerName     string
	PetNames      []string
	LastVisitDate *time.Time
	HasLine       bool
	IsOptOut      bool
	CurrentTags   []string
}

// PreviewCheckupSyncInput はPreviewCheckupSyncの入力パラメータ。
type PreviewCheckupSyncInput struct {
	CheckupType     string
	Species         string
	LastVisitBefore *time.Time
	LastVisitAfter  *time.Time
}

// PreviewCheckupSyncResult はPreviewCheckupSyncの結果。
type PreviewCheckupSyncResult struct {
	Owners          []CheckupSyncPreviewOwner
	TotalCount      int
	LineLinkedCount int
}

// CreateCheckupSyncInput はCreateCheckupSyncの入力パラメータ。
type CreateCheckupSyncInput struct {
	CheckupType string
	OwnerIDs    []uint64
	TagName     string
}

// CreateCheckupSyncResult はCreateCheckupSyncの結果。
type CreateCheckupSyncResult struct {
	SuccessCount   int
	SkippedCount   int
	FailedCount    int
	FailedOwnerIDs []uint64
}

// CheckupSyncService は健診対象者抽出・一括タグ連携の業務ロジックインターフェース（BE-004）。
type CheckupSyncService interface {
	PreviewCheckupSync(ctx context.Context, clinicID uint64, input PreviewCheckupSyncInput) (*PreviewCheckupSyncResult, error)
	CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error)
}

type checkupSyncService struct {
	repo         repository.CheckupSyncRepository
	ownerRepo    repository.OwnerRepository
	tagCacheRepo repository.LstepTagCacheRepository
	settingsSvc  LstepSettingsService
	auditSvc     AuditService
}

// NewCheckupSyncService は CheckupSyncService を初期化して返す。
func NewCheckupSyncService(
	repo repository.CheckupSyncRepository,
	ownerRepo repository.OwnerRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	settingsSvc LstepSettingsService,
	auditSvc AuditService,
) CheckupSyncService {
	return &checkupSyncService{
		repo:         repo,
		ownerRepo:    ownerRepo,
		tagCacheRepo: tagCacheRepo,
		settingsSvc:  settingsSvc,
		auditSvc:     auditSvc,
	}
}

func (s *checkupSyncService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

func (s *checkupSyncService) PreviewCheckupSync(ctx context.Context, clinicID uint64, input PreviewCheckupSyncInput) (*PreviewCheckupSyncResult, error) {
	rows, err := s.repo.FindCheckupSyncPreview(ctx, repository.FindCheckupSyncPreviewParams{
		ClinicID:        clinicID,
		Species:         input.Species,
		LastVisitBefore: input.LastVisitBefore,
		LastVisitAfter:  input.LastVisitAfter,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkup sync preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find checkup sync preview")
	}

	owners := make([]CheckupSyncPreviewOwner, 0, len(rows))
	lineLinkedCount := 0

	for _, row := range rows {
		hasLine := row.LineUserID != nil && *row.LineUserID != ""
		if hasLine {
			lineLinkedCount++
		}

		var petNames []string
		if row.PetNamesCSV != "" {
			petNames = strings.Split(row.PetNamesCSV, ",")
		} else {
			petNames = []string{}
		}

		var currentTags []string
		if hasLine {
			cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, row.OwnerID)
			if cacheErr != nil {
				slog.ErrorContext(ctx, "failed to load tag cache for preview", "error", cacheErr, "owner_id", row.OwnerID)
				currentTags = []string{}
			} else {
				currentTags = make([]string, 0, len(cached))
				for _, c := range cached {
					currentTags = append(currentTags, c.TagName)
				}
			}
		} else {
			currentTags = []string{}
		}

		owners = append(owners, CheckupSyncPreviewOwner{
			OwnerID:       row.OwnerID,
			OwnerName:     row.OwnerName,
			PetNames:      petNames,
			LastVisitDate: row.LastVisitDate,
			HasLine:       hasLine,
			IsOptOut:      row.LstepOptOut,
			CurrentTags:   currentTags,
		})
	}

	return &PreviewCheckupSyncResult{
		Owners:          owners,
		TotalCount:      len(owners),
		LineLinkedCount: lineLinkedCount,
	}, nil
}

func (s *checkupSyncService) CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
	if isAutoManagedTag(input.TagName) {
		return nil, apperrors.WrapInvalidInput("tag_name は自動管理タグのため使用できません")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, apperrors.WrapInvalidInput("Lステップ API が設定されていません")
	}

	result := &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}
	for _, ownerID := range input.OwnerIDs {
		owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if findErr != nil {
			slog.ErrorContext(ctx, "checkup sync: owner not found", "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}
		if owner.LstepOptOut || owner.LineUserID == nil || *owner.LineUserID == "" {
			result.SkippedCount++
			continue
		}

		if addErr := client.AddTag(ctx, *owner.LineUserID, input.TagName); addErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to add lstep tag", "error", addErr, "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}

		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, input.TagName, "manual"); upsertErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to upsert tag cache", "error", upsertErr, "owner_id", ownerID)
		}
		result.SuccessCount++
	}

	_ = s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "checkup_sync", "owner", nil)

	return result, nil
}
