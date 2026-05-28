package service

import (
	"context"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
)

// removeStaleTagsByPrefixes はキャッシュ内で指定プレフィックスに一致する古いタグを Lステップから解除する（ISSUE-006）。
// 同一カテゴリ1タグ保持ルールのための前処理として呼ぶ。
func (s *lstepTagSyncService) removeStaleTagsByPrefixes(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string, prefixes []string, skipTags map[string]struct{}) bool {
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for stale cleanup", "error", err)
		return false
	}
	apiFailed := false
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
				} else {
					_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
				}
				break
			}
		}
	}
	return apiFailed
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
