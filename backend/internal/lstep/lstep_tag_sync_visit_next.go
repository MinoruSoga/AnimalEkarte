package lstep

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SyncNextVisitTag は次回来院推奨日タグを同期する（BE-006）。
// 最新カルテの next_visit_recommended_date を参照し、古い next_visit_* タグを差し替える。
// date が nil の場合はすべての next_visit_* タグを削除する。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "next visit")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 最新カルテの次回来院推奨日を取得
	latest, err := s.medRecordRepo.FindLatestByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find latest medical record for next visit tag", "error", err)
		return apperrors.Wrap(err, "failed to find latest medical record")
	}

	// 既存の next_visit_* タグをキャッシュから取得して削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if !strings.HasPrefix(c.TagName, tagPrefixNextVisit) {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "next visit", "", false); err != nil {
			apiFailed = true
			continue
		}
	}

	// 新しい next_visit_YYYY-MM-DD タグを付与（日付が設定されている場合のみ）
	if latest == nil || latest.NextVisitRecommendedDate == nil {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := tagPrefixNextVisit + latest.NextVisitRecommendedDate.Format(time.DateOnly)
	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, newTag, "next visit", "", true); err != nil {
		return err
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
