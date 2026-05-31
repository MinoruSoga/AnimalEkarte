package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncCheckupTag は健診記録の作成・更新時に checkup_done_{typeID}_{YYYY-MM}/next_checkup_* タグを同期する（BE-008）。
// 同一健診種別の古い checkup_done タグを解除してから新タグを付与する。next_checkup_* は最新1件のみ。
func (s *lstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for checkup tag sync", "error", err)
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
		slog.ErrorContext(ctx, "failed to build lstep client for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 同一健診種別の古い checkup_done タグ + next_checkup_* タグをキャッシュ経由で削除
	stalePrefix := fmt.Sprintf("%s%d_", tagPrefixCheckupDone, checkupTypeID)
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for checkup tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, stalePrefix) || strings.HasPrefix(c.TagName, tagPrefixNextCheckup) {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale checkup tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete checkup tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// checkup_done_{typeID}_{YYYY-MM} タグを付与
	checkupTag := fmt.Sprintf("%s%d_%s", tagPrefixCheckupDone, checkupTypeID, checkupDate.Format("2006-01"))
	if addErr := client.AddTag(ctx, lineUserID, checkupTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add checkup tag", "error", addErr, "tag", checkupTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add checkup tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, checkupTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert checkup tag cache", "error", upsertErr)
	}

	// next_checkup_YYYY-MM-DD タグを付与（設定時のみ）
	if nextDate != nil {
		nextTag := tagPrefixNextCheckup + nextDate.Format("2006-01-02")
		if addErr := client.AddTag(ctx, lineUserID, nextTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add next_checkup tag", "error", addErr, "tag", nextTag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add next_checkup tag")
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, nextTag, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert next_checkup tag cache", "error", upsertErr)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
