package service

import (
	"context"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

func (s *checkupSyncService) PreviewCheckupSync(ctx context.Context, clinicID uint64, input *PreviewCheckupSyncInput, actorID *uint64) (*PreviewCheckupSyncResult, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input is nil")
	}
	rows, err := s.repo.FindCheckupSyncPreview(ctx, &repository.FindCheckupSyncPreviewParams{
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
		slog.ErrorContext(ctx, "failed to find checkup sync preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to find checkup sync preview")
	}

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v1 thresholds for checkup preview", "error", err, "clinic_id", clinicID)
		return nil, apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	owners := make([]CheckupSyncPreviewOwner, 0, len(rows))
	lineLinkedCount := 0
	optOutCount := 0
	noLivingPetCount := 0
	eligibleCount := 0

	for i := range rows {
		row := &rows[i]
		// ISSUE-009: CPM ステージは集計値ベースの後段フィルタとして適用する。
		// SQL ではなく service 層で判定することで、タグ同期側 CalculateCPMStage と同じロジックを共有する。
		cpmStage := computeCPMStageFromRow(row, thresholds)
		if input.CPMStage != "" && string(cpmStage) != input.CPMStage {
			continue
		}

		hasLine := row.LineUserID != nil && *row.LineUserID != ""
		hasLivingPet := row.LivingPetCount > 0
		if hasLine {
			lineLinkedCount++
		}
		if row.LstepOptOut {
			optOutCount++
		}
		if !hasLivingPet {
			noLivingPetCount++
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

		exclusionReason := deriveExclusionReason(hasLine, row.LstepOptOut, hasLivingPet)
		if exclusionReason == nil {
			eligibleCount++
		}

		owners = append(owners, CheckupSyncPreviewOwner{
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
		})
	}

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
		"total_count", len(owners),
		"eligible_count", eligibleCount,
		"line_linked_count", lineLinkedCount,
		"opt_out_count", optOutCount,
		"no_living_pet_count", noLivingPetCount,
	)

	// ISSUE-010: 抽出条件と件数を audit_logs.metadata に永続化する（slog の構造化ログを DB にも残す）。
	//   resource_id は対象者リストが多次元のため使えない（個別 owner_id 単一参照は意味を持たない）。
	//   よって metadata に集計値・条件をまとめて格納する。
	_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync_preview", "owner", nil,
		buildCheckupSyncPreviewMetadata(input, len(owners), eligibleCount, lineLinkedCount, optOutCount, noLivingPetCount),
	)

	return &PreviewCheckupSyncResult{
		Owners:           owners,
		TotalCount:       len(owners),
		EligibleCount:    eligibleCount,
		LineLinkedCount:  lineLinkedCount,
		OptOutCount:      optOutCount,
		NoLivingPetCount: noLivingPetCount,
	}, nil
}
