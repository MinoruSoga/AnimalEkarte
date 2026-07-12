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

	if len(input.OwnerIDs) == 0 {
		// PERF-M3: 早期 return でも「誰が・いつ・何を操作したか」の証跡を残す。
		if auditErr := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
			"checkup_sync", "owner", nil,
			map[string]any{
				"operation":    "checkup_sync_execute",
				"checkup_type": input.CheckupType,
				"tag_name":     input.TagName,
				"owner_count":  0,
			},
		); auditErr != nil {
			slog.WarnContext(ctx, "audit log failed for checkup sync execute (empty owner_ids)", "error", auditErr, "clinic_id", clinicID)
		}
		return &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}, nil
	}

	// PERF-3: オーナー情報を 1 クエリで一括取得する（FindByID N+1 解消）。
	//
	// H-2 コントラクト変更: FindByIDs / CountLivingByOwnerIDs の DB エラーは
	// 全体 HTTP 500 を返す（以前は per-owner FailedOwnerIDs に積んでいた）。
	// 理由: DB 障害時に partial success を返すと「処理済み」と誤解させるリスクがある。
	// 全件失敗した場合はリトライで安全に再試行できる（AddTag は冪等）。
	owners, findErr := s.ownerRepo.FindByIDs(ctx, clinicID, input.OwnerIDs)
	if findErr != nil {
		slog.ErrorContext(ctx, "checkup sync: failed to fetch owners", "error", findErr)
		return nil, apperrors.Wrap(findErr, "failed to fetch owners")
	}

	// 要求されたが見つからなかったオーナーを failed に積む。
	foundIDs := make(map[uint64]struct{}, len(owners))
	for _, o := range owners {
		foundIDs[o.ID] = struct{}{}
	}

	result := &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}
	// ISSUE-007: 監査要件 — スキップ理由内訳をカウントしてログに残す。
	skippedOptOut := 0
	skippedNoLivingPet := 0
	skippedLineUnlinked := 0

	for _, ownerID := range input.OwnerIDs {
		if _, ok := foundIDs[ownerID]; !ok {
			slog.ErrorContext(ctx, "checkup sync: owner not found", "owner_id", ownerID)
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, ownerID)
			result.FailedCount++
		}
	}

	// ISSUE-007: スキップ判定は preview 側の deriveExclusionReason と同じ優先度で行う。
	//   優先度: opt-out > 生存ペットなし > LINE未連携
	// opt-out 済みオーナーをスキップし、生存ペット判定の対象を絞り込む。
	candidateIDs := make([]uint64, 0, len(owners))
	for _, owner := range owners {
		if owner.LstepOptOut {
			skippedOptOut++
			result.SkippedCount++
			continue
		}
		candidateIDs = append(candidateIDs, owner.ID)
	}

	// PERF-3: 生存ペット数を 1 クエリで一括取得する（CountLivingByOwner N+1 解消）。
	var livingPetCounts map[uint64]int64
	if len(candidateIDs) > 0 {
		var countErr error
		livingPetCounts, countErr = s.petRepo.CountLivingByOwnerIDs(ctx, clinicID, candidateIDs)
		if countErr != nil {
			slog.ErrorContext(ctx, "checkup sync: failed to count living pets", "error", countErr)
			return nil, apperrors.Wrap(countErr, "failed to count living pets")
		}
	}

	// candidate オーナーをルックアップマップに変換。
	ownerMap := make(map[uint64]struct{ LineUserID *string }, len(owners))
	for _, o := range owners {
		ownerMap[o.ID] = struct{ LineUserID *string }{LineUserID: o.LineUserID}
	}

	for _, ownerID := range candidateIDs {
		if livingPetCounts[ownerID] == 0 {
			skippedNoLivingPet++
			result.SkippedCount++
			continue
		}

		lineUserID := ownerMap[ownerID].LineUserID
		if lineUserID == nil || *lineUserID == "" {
			skippedLineUnlinked++
			result.SkippedCount++
			continue
		}

		if addErr := client.AddTag(ctx, *lineUserID, input.TagName); addErr != nil {
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
	if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
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
	); err != nil {
		slog.WarnContext(ctx, "audit log failed for checkup sync execute", "error", err, "clinic_id", clinicID)
	}

	return result, nil
}
