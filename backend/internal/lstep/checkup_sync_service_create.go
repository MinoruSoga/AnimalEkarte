package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *checkupSyncService) CreateCheckupSync(ctx context.Context, clinicID uint64, input CreateCheckupSyncInput, actorID *uint64) (*CreateCheckupSyncResult, error) {
	if IsSystemManagedTag(input.TagName) {
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
		return nil, apperrors.Wrap(findErr, "failed to fetch owners")
	}

	result := &CreateCheckupSyncResult{FailedOwnerIDs: []uint64{}}

	// PERF-3: 生存ペット数を 1 クエリで一括取得する（CountLivingByOwner N+1 解消）。opt-out
	// 済みオーナーは生存ペット判定が不要なため、事前にスコープを絞り込む（BE-refactor.md E-10:
	// partitionCheckupSyncTargets 内でも opt-out 判定を行うが、DB クエリのスコープを決める
	// ためにここで一度だけ算出する）。
	candidateIDs := make([]uint64, 0, len(owners))
	for _, owner := range owners {
		if !owner.LstepOptOut {
			candidateIDs = append(candidateIDs, owner.ID)
		}
	}
	var livingPetCounts map[uint64]int64
	if len(candidateIDs) > 0 {
		var countErr error
		livingPetCounts, countErr = s.petRepo.CountLivingByOwnerIDs(ctx, clinicID, candidateIDs)
		if countErr != nil {
			return nil, apperrors.Wrap(countErr, "failed to count living pets")
		}
	}

	targets, skipStats, failedIDs := partitionCheckupSyncTargets(ctx, owners, input.OwnerIDs, livingPetCounts)
	result.FailedOwnerIDs = append(result.FailedOwnerIDs, failedIDs...)
	result.FailedCount += len(failedIDs)
	result.SkippedCount += skipStats.OptOut + skipStats.NoLivingPet + skipStats.LineUnlinked

	for _, t := range targets {
		if err := s.applyCheckupTag(ctx, client, clinicID, t.OwnerID, t.LineUserID, input.TagName); err != nil {
			result.FailedOwnerIDs = append(result.FailedOwnerIDs, t.OwnerID)
			result.FailedCount++
			continue
		}
		result.SuccessCount++
	}

	s.logAndAuditCheckupExecute(ctx, clinicID, actorID, input, len(input.OwnerIDs), result, skipStats)

	return result, nil
}

// checkupSyncTarget はタグ適用対象1件（オーナーIDとLINE User ID）（BE-refactor.md E-10）。
type checkupSyncTarget struct {
	OwnerID    uint64
	LineUserID string
}

// checkupSyncSkipStats は CreateCheckupSync のスキップ理由別カウント（BE-refactor.md E-10）。
type checkupSyncSkipStats struct {
	OptOut       int
	NoLivingPet  int
	LineUnlinked int
}

// partitionCheckupSyncTargets は要求された owner ID を「タグ適用対象（targets）」
// 「スキップ理由別カウント（skipStats）」「見つからなかった ID（failedIDs）」に分類する
// （BE-refactor.md E-10: CreateCheckupSync のskip分類ループ本体の純粋抽出）。
// livingPetCounts は呼び出し元が opt-out 済みを除いた候補IDに対して一括取得済みのマップ
// であること（opt-out 済みオーナーは本関数内で先に除外されるため参照されない）。
// 優先度: 見つからない > opt-out > 生存ペットなし > LINE未連携（preview 側の
// deriveExclusionReason と同順、ISSUE-007）。
func partitionCheckupSyncTargets(ctx context.Context, owners []*model.Owner, requestedIDs []uint64, livingPetCounts map[uint64]int64) (targets []checkupSyncTarget, skipStats checkupSyncSkipStats, failedIDs []uint64) {
	foundIDs := make(map[uint64]struct{}, len(owners))
	for _, o := range owners {
		foundIDs[o.ID] = struct{}{}
	}
	for _, ownerID := range requestedIDs {
		if _, ok := foundIDs[ownerID]; !ok {
			slog.ErrorContext(ctx, "checkup sync: owner not found", "owner_id", ownerID)
			failedIDs = append(failedIDs, ownerID)
		}
	}

	for _, owner := range owners {
		if owner.LstepOptOut {
			skipStats.OptOut++
			continue
		}
		if livingPetCounts[owner.ID] == 0 {
			skipStats.NoLivingPet++
			continue
		}
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			skipStats.LineUnlinked++
			continue
		}
		targets = append(targets, checkupSyncTarget{OwnerID: owner.ID, LineUserID: *owner.LineUserID})
	}
	return targets, skipStats, failedIDs
}

// applyCheckupTag は1オーナーへのタグ付与（AddTag→成功時UpsertTag）を実行する
// （BE-refactor.md E-10）。AddTag 失敗時のみエラーを返す。UpsertTag 失敗は non-fatal
// （キャッシュ更新のみ、ログ記録して継続、AddTag 自体は成功扱い）。
func (s *checkupSyncService) applyCheckupTag(ctx context.Context, client lstepapi.Client, clinicID, ownerID uint64, lineUserID, tagName string) error {
	if addErr := client.AddTag(ctx, lineUserID, tagName); addErr != nil {
		slog.ErrorContext(ctx, "checkup sync: failed to add lstep tag", "error", addErr, "owner_id", ownerID)
		return addErr
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "manual", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "checkup sync: failed to upsert tag cache", "error", upsertErr, "owner_id", ownerID)
	}
	return nil
}

// logAndAuditCheckupExecute は一括タグ付与の実行結果を slog + audit_logs に記録する
// （BE-refactor.md E-10: CreateCheckupSync の観測性テール抽出）。
func (s *checkupSyncService) logAndAuditCheckupExecute(ctx context.Context, clinicID uint64, actorID *uint64,
	input CreateCheckupSyncInput, requestedCount int, result *CreateCheckupSyncResult, skipStats checkupSyncSkipStats) {
	// ISSUE-007: スキップ理由内訳を監査ログとして残す（opt-out / 生存ペットなし / LINE未連携）。
	slog.InfoContext(ctx, "checkup sync executed",
		"clinic_id", clinicID,
		"checkup_type", input.CheckupType,
		"tag_name", input.TagName,
		"requested_count", requestedCount,
		"success_count", result.SuccessCount,
		"skipped_count", result.SkippedCount,
		"failed_count", result.FailedCount,
		"skipped_opt_out", skipStats.OptOut,
		"skipped_no_living_pet", skipStats.NoLivingPet,
		"skipped_line_unlinked", skipStats.LineUnlinked,
	)

	// ISSUE-010: 一括タグ付与の実行件数とスキップ理由内訳を audit_logs.metadata に永続化する。
	if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinicID, actorID,
		"checkup_sync", "owner", nil,
		map[string]any{
			"operation":             "checkup_sync_execute",
			"checkup_type":          input.CheckupType,
			"tag_name":              input.TagName,
			"requested_count":       requestedCount,
			"success_count":         result.SuccessCount,
			"skipped_count":         result.SkippedCount,
			"failed_count":          result.FailedCount,
			"skipped_opt_out":       skipStats.OptOut,
			"skipped_no_living_pet": skipStats.NoLivingPet,
			"skipped_line_unlinked": skipStats.LineUnlinked,
		},
	); err != nil {
		slog.WarnContext(ctx, "audit log failed for checkup sync execute", "error", err, "clinic_id", clinicID)
	}
}
