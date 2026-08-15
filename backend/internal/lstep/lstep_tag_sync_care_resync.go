package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ResyncOwnerCheckupTags は飼い主の生存健診記録から checkup_done_* / next_checkup_* タグを再構築する（ISSUE-004）。
// 健診の更新・削除後に呼び出すこと。同一健診種別では最新検査日のみ保持。next_checkup は最遠の next_date のみ保持。
func (s *lstepTagSyncService) ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "checkup resync")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for checkup resync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for resync", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	newTagSet := buildLatestCheckupTagSet(checkups)

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for checkup resync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if !strings.HasPrefix(c.TagName, tagPrefixCheckupDone) && !strings.HasPrefix(c.TagName, tagPrefixNextCheckup) {
			continue
		}
		if _, keep := newTagSet[c.TagName]; keep {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "checkup resync", "", false); err != nil {
			apiFailed = true
			continue
		}
	}

	for tag := range newTagSet {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tag, "checkup resync", "", true); err != nil {
			return err
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildLatestCheckupTagSet は健診記録から「種別ごとの最新検査日」「最遠の next_date」のタグ集合を返す。
func buildLatestCheckupTagSet(checkups []model.Checkup) map[string]struct{} {
	latestByType := make(map[uint64]time.Time)
	var latestNext *time.Time
	for i := range checkups {
		c := &checkups[i]
		if cur, ok := latestByType[c.CheckupTypeID]; !ok || c.Date.After(cur) {
			latestByType[c.CheckupTypeID] = c.Date
		}
		if c.NextDate != nil {
			if latestNext == nil || c.NextDate.After(*latestNext) {
				t := *c.NextDate
				latestNext = &t
			}
		}
	}
	tagSet := make(map[string]struct{}, len(latestByType)+1)
	for typeID, date := range latestByType {
		tagSet[fmt.Sprintf("%s%d_%s", tagPrefixCheckupDone, typeID, date.Format("2006-01"))] = struct{}{}
	}
	if latestNext != nil {
		tagSet[tagPrefixNextCheckup+latestNext.Format(time.DateOnly)] = struct{}{}
	}
	return tagSet
}
