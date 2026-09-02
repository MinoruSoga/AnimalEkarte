package lstep

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// removeStaleTagsByPrefixes はキャッシュ内で指定プレフィックスに一致する古いタグを Lステップから解除する（ISSUE-006）。
// 同一カテゴリ1タグ保持ルールのための前処理として呼ぶ。
//
// 戻り値の契約（LSA-11）:
//   - cache read 失敗: (true, err) — 呼び出し元はカテゴリ API（AddTag/RemoveTag）を一切呼ばず失敗を伝播する（fail-closed）。
//   - RemoveTag 部分失敗: (true, nil) — 耐久的な failure accounting は行い、desired AddTag は既存どおり継続可能。
//   - 成功: (false, nil)
func (s *lstepTagSyncService) removeStaleTagsByPrefixes(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string, prefixes []string, skipTags map[string]struct{}) (apiFailed bool, err error) {
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		// LSA-11: cache load failure is observable and must stop every category API call.
		slog.ErrorContext(ctx, "failed to load tag cache for stale cleanup", "error", err)
		return true, apperrors.Wrap(err, "failed to load tag cache for stale cleanup")
	}
	for _, c := range cached {
		if _, skip := skipTags[c.TagName]; skip {
			continue
		}
		for _, pfx := range prefixes {
			if strings.HasPrefix(c.TagName, pfx) {
				if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
					slog.ErrorContext(ctx, "failed to remove stale tag", "error", delErr, "tag", c.TagName)
					s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
					apiFailed = true
					break
				}
				if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delCacheErr != nil {
					slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", c.TagName)
				}
				break
			}
		}
	}
	return apiFailed, nil
}

// resolveSyncTargetOwner は飼い主のオプトアウト状態と LINE 連携有無を確認し、
// 同期先の lineUserID を解決する（ISSUE dup-lstep-tag-apply Phase 1）。
// ok=false の場合は呼び出し元でスキップする（optOut またはLINE未連携）。
func (s *lstepTagSyncService) resolveSyncTargetOwner(ctx context.Context, clinicID, ownerID uint64, tagLabel string) (lineUserID string, ok bool, err error) {
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out", "error", err, "tag", tagLabel)
		return "", false, apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return "", false, nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return "", false, nil
	}
	return *owner.LineUserID, true, nil
}

// resolveSyncTarget は shouldSkipSync（クリニック単位の同期無効判定）を含めた
// 同期先解決のフル版。Phase 1 では追加のみ（vaccine/checkup 等 Phase 2+ で使用）。
func (s *lstepTagSyncService) resolveSyncTarget(ctx context.Context, clinicID, ownerID uint64, tagLabel string) (lineUserID string, ok bool, err error) {
	skip, err := s.shouldSkipSync(ctx, clinicID)
	if err != nil {
		// shouldSkipSync が既に apperrors.Wrap 済みのため、ここでは再ラップしない（G3-1 P2a）。
		slog.ErrorContext(ctx, "failed to check lstep sync enabled", "error", err, "tag", tagLabel, "clinic_id", clinicID)
		return "", false, err
	}
	if skip {
		return "", false, nil
	}
	return s.resolveSyncTargetOwner(ctx, clinicID, ownerID, tagLabel)
}

// mappingsFor は cached が非nilならそれをそのまま返し、nilなら tagName の
// LstepTagCodeMapping を取得する（BE-refactor.md E-7: batch からの hoist 対応の
// Sync*WithMappings 系関数冒頭ブロック統合）。label はログメッセージ
// 「failed to find tag code mappings for <label>」の可変部（呼び出し元ごとの既存文言を再現）。
func (s *lstepTagSyncService) mappingsFor(ctx context.Context, clinicID uint64, tagName, label string, cached []*model.LstepTagCodeMapping) ([]*model.LstepTagCodeMapping, error) {
	if cached != nil {
		return cached, nil
	}
	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, tagName)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for "+label, "error", err)
		return nil, apperrors.Wrap(err, "failed to find tag code mappings")
	}
	return mappings, nil
}

// thresholdsFor は cached が非nilならそれをそのまま返し、nilなら健康維持しきい値を取得する
// （BE-refactor.md E-7）。label はログメッセージ
// 「failed to get health prevention thresholds for <label>」の可変部。
func (s *lstepTagSyncService) thresholdsFor(ctx context.Context, clinicID uint64, label string, cached *model.HealthPreventionThresholds) (model.HealthPreventionThresholds, error) {
	if cached != nil {
		return *cached, nil
	}
	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for "+label, "error", err, "clinic_id", clinicID)
		return model.HealthPreventionThresholds{}, apperrors.Wrap(err, "failed to get health prevention thresholds")
	}
	return thresholds, nil
}

// applyTagState は desired の真偽に応じて Lステップタグを付与/解除し、キャッシュを更新する。
// desired=true: AddTag→失敗時 notifyAPIFailure + wrap されたエラーを返す。成功時 UpsertTag（失敗は non-fatal）。
// desired=false: RemoveTag→失敗時 notifyAPIFailure + wrap されたエラーを返す。成功時 DeleteTag（失敗は non-fatal）。
// エラーを return するかどうか（呼び出し元で apiFailed に丸めるか）は呼び出し元の判断に委ねる。
func (s *lstepTagSyncService) applyTagState(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID, tag, actionLabel, reason string, desired bool) error {
	if desired {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add "+actionLabel+" tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add "+actionLabel+" tag")
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", reason); upsertErr != nil {
			slog.WarnContext(ctx, "failed to upsert tag cache (non-fatal)", "tag", tag, "owner_id", ownerID, "error", upsertErr)
		}
		return nil
	}
	if delErr := client.RemoveTag(ctx, lineUserID, tag); delErr != nil {
		slog.ErrorContext(ctx, "failed to remove "+actionLabel+" tag", "error", delErr, "tag", tag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(delErr, "failed to remove "+actionLabel+" tag")
	}
	if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, tag); delCacheErr != nil {
		slog.WarnContext(ctx, "failed to delete tag cache (non-fatal)", "tag", tag, "owner_id", ownerID, "error", delCacheErr)
	}
	return nil
}

// notifyAPIFailure は Lステップ API 呼び出し失敗時にカウンターを更新し、
// 閾値（lstepSyncErrorThreshold）に達した場合は EXCL_カルテ連携エラー タグを付与する。
// エラーはログのみ — 呼び出し元に伝播させない（best-effort）。
func (s *lstepTagSyncService) notifyAPIFailure(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	newCount, err := s.errorCounterRepo.IncrementFailure(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to increment lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if newCount < lstepSyncErrorThreshold {
		return
	}
	if addErr := client.AddTag(ctx, lineUserID, lstepErrorTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add EXCL error tag", "error", addErr, "owner_id", ownerID)
		return
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, lstepErrorTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert EXCL error tag cache", "error", upsertErr)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// FEAT-377: CPM V2 / LTV 上位 20% / VISIT / PET / 除外タグ同期
// ─────────────────────────────────────────────────────────────────────────────

// notifyAPISuccess は Lステップ API 呼び出し成功時にエラーカウンターをリセットし、
// EXCL_カルテ連携エラー タグを解除する（カウンターが 0 の場合は noop）。
// エラーはログのみ — 呼び出し元に伝播させない（best-effort）。
func (s *lstepTagSyncService) notifyAPISuccess(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	if s.errorCounterRepo == nil {
		return
	}
	counter, err := s.errorCounterRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return
		}
		slog.ErrorContext(ctx, "failed to find lstep error counter", "error", err, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if counter.FailureCount == 0 {
		return
	}
	if removeErr := client.RemoveTag(ctx, lineUserID, lstepErrorTag); removeErr != nil {
		slog.ErrorContext(ctx, "failed to remove EXCL error tag", "error", removeErr, "owner_id", ownerID)
		return
	}
	if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, lstepErrorTag); delErr != nil {
		slog.ErrorContext(ctx, "failed to delete EXCL error tag cache", "error", delErr)
	}
	if resetErr := s.errorCounterRepo.ResetFailure(ctx, clinicID, ownerID); resetErr != nil {
		slog.ErrorContext(ctx, "failed to reset lstep error counter", "error", resetErr, "owner_id", ownerID)
	}
}
