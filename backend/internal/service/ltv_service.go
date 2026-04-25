package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

// cpmStageToAPI は内部CPMStageをAPIレスポンス名（task doc準拠）に変換する。
func cpmStageToAPI(stage CPMStage) string {
	switch stage {
	case CPMStageEncounter:
		return "encounter"
	case CPMStageGrowing:
		return "growing"
	case CPMStageCore:
		return "core"
	case CPMStageNoah:
		return "noah"
	case CPMStageSpot:
		return "spot"
	default: // CPMStageDormant
		return "dormant"
	}
}

// apiToCPMStage はAPIフィルタ名を内部CPMStageに変換する。
func apiToCPMStage(name string) (CPMStage, bool) {
	switch name {
	case "encounter":
		return CPMStageEncounter, true
	case "growing":
		return CPMStageGrowing, true
	case "core":
		return CPMStageCore, true
	case "noah":
		return CPMStageNoah, true
	case "spot":
		return CPMStageSpot, true
	case "dormant":
		return CPMStageDormant, true
	default:
		return "", false
	}
}

// maskLineUserID はLINE User IDをマスクする（例: "U1234abcd" → "U****abcd"）。
func maskLineUserID(id string) string {
	if len(id) < 4 {
		return "U****"
	}
	return "U****" + id[len(id)-4:]
}

// OwnerLTVItem はLTV一覧の1件。
type OwnerLTVItem struct {
	OwnerID          uint64
	OwnerName        string
	LineUserIDMasked *string
	HasLine          bool
	TotalAmount      int64
	TotalVisitCount  int64
	AnnualVisitCount int64
	LastVisitDate    *time.Time
	FirstVisitDate   *time.Time
	CPMStageAPI      string
}

// ListOwnerLTVInput はListOwnerLTV の入力パラメータ。
type ListOwnerLTVInput struct {
	Sort           string
	MinTotalAmount *int64
	MaxTotalAmount *int64
	MinVisitCount  *int64
	CPMStageFilter string // encounter/growing/core/spot/noah/dormant or ""
	LineLinked     bool
	Page           int
	PerPage        int
}

// ListOwnerLTVResult はListOwnerLTV の結果。
type ListOwnerLTVResult struct {
	Total   int
	Page    int
	PerPage int
	Items   []OwnerLTVItem
}

// SyncLtvTagsInput はSyncLtvTags の入力パラメータ。
type SyncLtvTagsInput struct {
	TagName        string
	MinTotalAmount *int64
	DryRun         bool
}

// SyncLtvTagsResult はSyncLtvTags の結果。
type SyncLtvTagsResult struct {
	DryRun  bool
	Total   int
	Synced  int
	Skipped int
}

// LtvService はLTV集計・一括タグ同期の業務ロジックインターフェース（BE-010）。
type LtvService interface {
	ListOwnerLTV(ctx context.Context, clinicID uint64, input ListOwnerLTVInput) (*ListOwnerLTVResult, error)
	SyncLtvTags(ctx context.Context, clinicID uint64, input SyncLtvTagsInput) (*SyncLtvTagsResult, error)
}

type ltvService struct {
	repo         repository.LtvRepository
	tagCacheRepo repository.LstepTagCacheRepository
}

// NewLtvService は LtvService を初期化して返す。
func NewLtvService(repo repository.LtvRepository, tagCacheRepo repository.LstepTagCacheRepository) LtvService {
	return &ltvService{repo: repo, tagCacheRepo: tagCacheRepo}
}

func (s *ltvService) ListOwnerLTV(ctx context.Context, clinicID uint64, input ListOwnerLTVInput) (*ListOwnerLTVResult, error) {
	var cpmFilter CPMStage
	if input.CPMStageFilter != "" {
		stage, ok := apiToCPMStage(input.CPMStageFilter)
		if !ok {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("cpm_stage %q は無効です（encounter/growing/core/spot/noah/dormant のいずれかを指定してください）", input.CPMStageFilter))
		}
		cpmFilter = stage
	}

	rows, err := s.repo.FindOwnerLTV(ctx, repository.FindOwnerLTVParams{
		ClinicID:       clinicID,
		Sort:           input.Sort,
		MinTotalAmount: input.MinTotalAmount,
		MaxTotalAmount: input.MaxTotalAmount,
		MinVisitCount:  input.MinVisitCount,
		LineLinked:     input.LineLinked,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner ltv", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find owner ltv")
	}

	now := time.Now()
	var items []OwnerLTVItem
	for _, row := range rows {
		daysSinceVisit := -1
		if row.LastVisitDate != nil {
			daysSinceVisit = int(now.Sub(*row.LastVisitDate).Hours() / 24)
		}

		stage := CalculateCPMStage(CPMData{
			TotalVisitCount:  row.TotalVisitCount,
			AnnualVisitCount: row.AnnualVisitCount,
			DaysSinceVisit:   daysSinceVisit,
			LTVAmount:        row.TotalAmount,
		})

		if cpmFilter != "" && stage != cpmFilter {
			continue
		}

		item := OwnerLTVItem{
			OwnerID:          row.OwnerID,
			OwnerName:        row.OwnerName,
			HasLine:          row.LineUserID != nil && *row.LineUserID != "",
			TotalAmount:      row.TotalAmount,
			TotalVisitCount:  row.TotalVisitCount,
			AnnualVisitCount: row.AnnualVisitCount,
			LastVisitDate:    row.LastVisitDate,
			FirstVisitDate:   row.FirstVisitDate,
			CPMStageAPI:      cpmStageToAPI(stage),
		}
		if row.LineUserID != nil && *row.LineUserID != "" {
			masked := maskLineUserID(*row.LineUserID)
			item.LineUserIDMasked = &masked
		}
		items = append(items, item)
	}

	total := len(items)

	// Go-side pagination
	perPage := input.PerPage
	if perPage <= 0 {
		perPage = 50
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage
	if offset >= total {
		items = nil
	} else {
		end := offset + perPage
		if end > total {
			end = total
		}
		items = items[offset:end]
	}

	return &ListOwnerLTVResult{
		Total:   total,
		Page:    page,
		PerPage: perPage,
		Items:   items,
	}, nil
}

func (s *ltvService) SyncLtvTags(ctx context.Context, clinicID uint64, input SyncLtvTagsInput) (*SyncLtvTagsResult, error) {
	// 禁止プレフィックスチェック（BE-019と共通の isAutoManagedTag を使用）
	if isAutoManagedTag(input.TagName) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("tag_name %q は自動管理タグのため使用できません", input.TagName))
	}

	rows, err := s.repo.FindOwnerLTV(ctx, repository.FindOwnerLTVParams{
		ClinicID:       clinicID,
		MinTotalAmount: input.MinTotalAmount,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner ltv for sync", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find owner ltv for sync")
	}

	result := &SyncLtvTagsResult{DryRun: input.DryRun, Total: len(rows)}
	for _, row := range rows {
		if row.LstepOptOut {
			result.Skipped++
			continue
		}
		if row.LineUserID == nil {
			result.Skipped++
			continue
		}

		if !input.DryRun {
			if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, row.OwnerID, input.TagName, "manual"); upsertErr != nil {
				slog.ErrorContext(ctx, "failed to upsert ltv tag", "owner_id", row.OwnerID, "error", upsertErr)
				result.Skipped++
				continue
			}
		}
		result.Synced++
	}

	return result, nil
}
