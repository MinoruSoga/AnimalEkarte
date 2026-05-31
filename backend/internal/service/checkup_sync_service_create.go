package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *checkupSyncService) CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
	if isSystemManagedTag(input.TagName) {
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
	// ISSUE-007: 監査要件 — スキップ理由内訳をカウントしてログに残す。
	skippedOptOut := 0
	skippedNoLivingPet := 0
	skippedLineUnlinked := 0

	for _, ownerID := range input.OwnerIDs {
		owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if findErr != nil {
			slog.ErrorContext(ctx, "checkup sync: owner not found", "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}

		// ISSUE-007: スキップ判定は preview 側の deriveExclusionReason と同じ優先度で行う。
		//   優先度: opt-out > 生存ペットなし > LINE未連携
		// API を直接叩かれた場合でも死亡ペットのみの飼い主を確実に除外し、誤配信を防ぐ。
		if owner.LstepOptOut {
			skippedOptOut++
			result.SkippedCount++
			continue
		}

		livingPetCount, countErr := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
		if countErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to count living pets", "error", countErr, "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}
		if livingPetCount == 0 {
			skippedNoLivingPet++
			result.SkippedCount++
			continue
		}

		if owner.LineUserID == nil || *owner.LineUserID == "" {
			skippedLineUnlinked++
			result.SkippedCount++
			continue
		}

		if addErr := client.AddTag(ctx, *owner.LineUserID, input.TagName); addErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to add lstep tag", "error", addErr, "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
			continue
		}

		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, input.TagName, "manual", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to upsert tag cache", "error", upsertErr, "owner_id", ownerID)
		}
		result.SuccessCount++
	}

	// ISSUE-007: スキップ理由内訳を監査ログとして残す（opt-out / 生存ペットなし / LINE未連携）。
	slog.InfoContext(ctx, "checkup sync executed",
		"clinic_id", clinicID,
		"checkup_type", input.CheckupType,
		"tag_name", input.TagName,
		"requested_count", len(input.OwnerIDs),
		"success_count", result.SuccessCount,
		"skipped_count", result.SkippedCount,
		"failed_count", result.FailedCount,
		"skipped_opt_out", skippedOptOut,
		"skipped_no_living_pet", skippedNoLivingPet,
		"skipped_line_unlinked", skippedLineUnlinked,
	)

	// ISSUE-010: 一括タグ付与の実行件数とスキップ理由内訳を audit_logs.metadata に永続化する。
	_ = s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync", "owner", nil,
		map[string]any{
			"operation":             "checkup_sync_execute",
			"checkup_type":          input.CheckupType,
			"tag_name":              input.TagName,
			"requested_count":       len(input.OwnerIDs),
			"success_count":         result.SuccessCount,
			"skipped_count":         result.SkippedCount,
			"failed_count":          result.FailedCount,
			"skipped_opt_out":       skippedOptOut,
			"skipped_no_living_pet": skippedNoLivingPet,
			"skipped_line_unlinked": skippedLineUnlinked,
		},
	)

	return result, nil
}
