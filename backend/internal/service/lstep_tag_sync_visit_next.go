package service

import (
	"context"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncNextVisitTag は次回来院推奨日タグを同期する（BE-006）。
// 最新カルテの next_visit_recommended_date を参照し、古い next_visit_* タグを差し替える。
// date が nil の場合はすべての next_visit_* タグを削除する。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

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
		if strings.HasPrefix(c.TagName, tagPrefixNextVisit) {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove next_visit tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete next_visit tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 新しい next_visit_YYYY-MM-DD タグを付与（日付が設定されている場合のみ）
	if latest == nil || latest.NextVisitRecommendedDate == nil {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := tagPrefixNextVisit + latest.NextVisitRecommendedDate.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add next_visit tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add next_visit tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert next_visit tag cache", "error", upsertErr)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
