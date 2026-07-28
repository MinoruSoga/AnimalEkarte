package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SyncCheckupTag は健診記録の作成・更新時に checkup_done_{typeID}_{YYYY-MM}/next_checkup_* タグを同期する（BE-008）。
// 同一健診種別の古い checkup_done タグを解除してから新タグを付与する。next_checkup_* は最新1件のみ。
func (s *lstepTagSyncService) SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "checkup")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "checkup", "", false); err != nil {
				apiFailed = true
				continue
			}
		}
	}

	// checkup_done_{typeID}_{YYYY-MM} タグを付与
	checkupTag := fmt.Sprintf("%s%d_%s", tagPrefixCheckupDone, checkupTypeID, checkupDate.Format("2006-01"))
	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, checkupTag, "checkup", "", true); err != nil {
		return err
	}

	// next_checkup_YYYY-MM-DD タグを付与（設定時のみ）
	if nextDate != nil {
		nextTag := tagPrefixNextCheckup + nextDate.Format(time.DateOnly)
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, nextTag, "next_checkup", "", true); err != nil {
			return err
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
