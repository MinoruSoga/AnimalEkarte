package lstep

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupSyncPreviewTimeout bounds PreviewCheckupSync wall time (BUG-032).
// Local SQL (not external LSTEP) still needs a deadline so the UI can recover.
const CheckupSyncPreviewTimeout = 15 * time.Second

// CheckupSyncPreviewOwnerCap is the max owners returned after CPM post-filter.
// Matches FE CHECKUP_SYNC_OWNER_LIMIT.
const CheckupSyncPreviewOwnerCap = 100

func (s *checkupSyncService) PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input is nil")
	}
	// BUG-032: enforce deadline so unbounded SQL cannot hang the request forever.
	previewCtx, cancel := context.WithTimeout(ctx, CheckupSyncPreviewTimeout)
	defer cancel()

	rows, err := s.repo.FindCheckupSyncPreview(previewCtx, &FindCheckupSyncPreviewParams{
		ClinicID:            clinicID,
		Species:             input.Species,
		LastVisitBefore:     input.LastVisitBefore,
		LastVisitAfter:      input.LastVisitAfter,
		MinAgeYears:         input.MinAgeYears,
		MaxAgeYears:         input.MaxAgeYears,
		HasChronicCondition: input.HasChronicCondition,
		MinTotalAmount:      input.MinTotalAmount,
		MinAnnualVisitCount: input.MinAnnualVisitCount,
		LastCheckupBefore:   input.LastCheckupBefore,
		LastCheckupAfter:    input.LastCheckupAfter,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(previewCtx.Err(), context.DeadlineExceeded) {
			// BUG-030: do not surface opaque 500 — timeout is a recoverable client condition.
			slog.ErrorContext(ctx, "checkup sync preview timed out", "error", err, "clinic_id", clinicID, "timeout", CheckupSyncPreviewTimeout)
			return nil, apperrors.WrapInvalidInput("健診対象者プレビューがタイムアウトしました。条件を絞り込んで再試行してください")
		}
		slog.ErrorContext(ctx, "failed to find checkup sync preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find checkup sync preview")
	}

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(previewCtx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v1 thresholds for checkup preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	// N+1回避: LINE連携済み owner_id を先に収集し、タグキャッシュを1クエリで一括取得する。
	lineLinkedOwnerIDs := make([]uint64, 0, len(rows))
	for i := range rows {
		row := &rows[i]
		if row.LineUserID != nil && *row.LineUserID != "" {
			lineLinkedOwnerIDs = append(lineLinkedOwnerIDs, row.OwnerID)
		}
	}
	tagCacheByOwner, err := s.tagCacheRepo.FindByOwners(previewCtx, clinicID, lineLinkedOwnerIDs)
	if err != nil {
		// 一括取得の失敗は non-fatal: 全員分のタグを空扱いにしてプレビューは継続する
		// （per-owner版と異なり、失敗が全体に及ぶ点は仕様として G7-2 で固定した挙動差）。
		slog.ErrorContext(ctx, "failed to batch load tag cache for preview", "error", err, "clinic_id", clinicID)
		tagCacheByOwner = map[uint64][]*model.LstepTagCache{}
	}

	owners := make([]CheckupSyncPreviewOwner, 0, len(rows))
	var counters checkupPreviewCounters

	for i := range rows {
		if len(owners) >= CheckupSyncPreviewOwnerCap {
			break
		}
		row := &rows[i]
		// ISSUE-009: CPM ステージは集計値ベースの後段フィルタとして適用する。
		// SQL ではなく service 層で判定することで、タグ同期側 CalculateCPMStage と同じロジックを共有する。
		cpmStage := computeCPMStageFromRow(row, thresholds)
		if input.CPMStage != "" && string(cpmStage) != input.CPMStage {
			continue
		}
		owner, delta := buildPreviewOwner(row, cpmStage, tagCacheByOwner)
		owners = append(owners, owner)
		counters.LineLinked += delta.LineLinked
		counters.OptOut += delta.OptOut
		counters.NoLivingPet += delta.NoLivingPet
		counters.Eligible += delta.Eligible
	}

	s.logAndAuditCheckupPreview(ctx, clinicID, actorID, input,
		len(owners), counters.Eligible, counters.LineLinked, counters.OptOut, counters.NoLivingPet)

	return &PreviewCheckupSyncResult{
		Owners:           owners,
		TotalCount:       len(owners),
		EligibleCount:    counters.Eligible,
		LineLinkedCount:  counters.LineLinked,
		OptOutCount:      counters.OptOut,
		NoLivingPetCount: counters.NoLivingPet,
	}, nil
}

// checkupPreviewCounters は PreviewCheckupSync のper-owner分類ループが集計する統計群
// （BE-refactor.md E-9）。
type checkupPreviewCounters struct {
	LineLinked  int
	OptOut      int
	NoLivingPet int
	Eligible    int
}

// buildPreviewOwner は1オーナー分の行データから CheckupSyncPreviewOwner を構築し、
// 分類カウンタへの加算量を返す（BE-refactor.md E-9: PreviewCheckupSync のper-owner
// 分類ループ本体の純粋抽出）。呼び出し元は CPM ステージフィルタで除外されなかった行のみを
// 渡すこと（フィルタで除外された行はどのカウンタも増加させない、という元の挙動を保つため
// フィルタ自体はこの関数の外で行う）。
func buildPreviewOwner(row *CheckupSyncPreviewRow, cpmStage CPMStage, tagCacheByOwner map[uint64][]*model.LstepTagCache) (CheckupSyncPreviewOwner, checkupPreviewCounters) {
	var counters checkupPreviewCounters

	hasLine := row.LineUserID != nil && *row.LineUserID != ""
	hasLivingPet := row.LivingPetCount > 0
	if hasLine {
		counters.LineLinked++
	}
	if row.LstepOptOut {
		counters.OptOut++
	}
	if !hasLivingPet {
		counters.NoLivingPet++
	}

	var petNames []string
	if row.PetNamesCSV != "" {
		petNames = strings.Split(row.PetNamesCSV, ",")
	} else {
		petNames = []string{}
	}

	var currentTags []string
	if hasLine {
		cached := tagCacheByOwner[row.OwnerID]
		currentTags = make([]string, 0, len(cached))
		for _, c := range cached {
			currentTags = append(currentTags, c.TagName)
		}
	} else {
		currentTags = []string{}
	}

	exclusionReason := deriveExclusionReason(hasLine, row.LstepOptOut, hasLivingPet)
	if exclusionReason == nil {
		counters.Eligible++
	}

	owner := CheckupSyncPreviewOwner{
		OwnerID:             row.OwnerID,
		OwnerName:           row.OwnerName,
		PetNames:            petNames,
		LastVisitDate:       row.LastVisitDate,
		HasLine:             hasLine,
		IsOptOut:            row.LstepOptOut,
		HasLivingPet:        hasLivingPet,
		ExclusionReason:     exclusionReason,
		CurrentTags:         currentTags,
		MinPetAgeYears:      row.MinPetAgeYears,
		MaxPetAgeYears:      row.MaxPetAgeYears,
		HasChronicCondition: row.HasChronicCondition,
		CPMStage:            string(cpmStage),
		TotalAmount:         row.TotalAmount,
		AnnualVisitCount:    row.AnnualVisitCount,
		LastCheckupDate:     row.LastCheckupDate,
	}
	return owner, counters
}

// logAndAuditCheckupPreview はプレビュー結果の統計を slog + audit_logs に記録する
// （BE-refactor.md E-9: PreviewCheckupSync の観測性テール抽出）。
func (s *checkupSyncService) logAndAuditCheckupPreview(ctx context.Context, clinicID uint64, actorID *uint64,
	input *PreviewCheckupSyncInput, totalCount, eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount int) {
	// ISSUE-005 / ISSUE-009: 抽出条件と除外理由ごとの件数を残す（監査ログ要件）。
	slog.InfoContext(ctx, "checkup sync preview extracted",
		"clinic_id", clinicID,
		"checkup_type", input.CheckupType,
		"species", input.Species,
		"last_visit_before", input.LastVisitBefore,
		"last_visit_after", input.LastVisitAfter,
		"min_age_years", input.MinAgeYears,
		"max_age_years", input.MaxAgeYears,
		"has_chronic_condition", input.HasChronicCondition,
		"cpm_stage", input.CPMStage,
		"min_total_amount", input.MinTotalAmount,
		"min_annual_visit_count", input.MinAnnualVisitCount,
		"last_checkup_before", input.LastCheckupBefore,
		"last_checkup_after", input.LastCheckupAfter,
		"total_count", totalCount,
		"eligible_count", eligibleCount,
		"line_linked_count", lineLinkedCount,
		"opt_out_count", optOutCount,
		"no_living_pet_count", noLivingPetCount,
	)

	// ISSUE-010: 抽出条件と件数を audit_logs.metadata に永続化する（slog の構造化ログを DB にも残す）。
	//   resource_id は対象者リストが多次元のため使えない（個別 owner_id 単一参照は意味を持たない）。
	//   よって metadata に集計値・条件をまとめて格納する。
	if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync_preview", "owner", nil,
		buildCheckupSyncPreviewMetadata(input, totalCount, eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount),
	); err != nil {
		slog.WarnContext(ctx, "audit log failed for checkup sync preview", "error", err, "clinic_id", clinicID)
	}
}
