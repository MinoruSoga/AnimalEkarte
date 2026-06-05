package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncDormantTags は最終来院からの経過日数に基づき dormant_* タグを差分同期する（BE-005）。
func (s *lstepTagSyncService) SyncDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for dormant tag sync", "error", err)
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
		slog.ErrorContext(ctx, "failed to build lstep client for dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 付与すべき dormant タグを決定（閾値は clinic_settings から取得、Q21）
	thresholds, tErr := s.settingsSvc.GetDormantThresholds(ctx, clinicID)
	if tErr != nil {
		slog.ErrorContext(ctx, "SyncDormantTags: failed to get dormant thresholds", "clinic_id", clinicID, "error", tErr)
		return apperrors.Wrap(tErr, "failed to get dormant thresholds")
	}
	var targetTag string
	switch {
	case daysSinceLastVisit < 0 || daysSinceLastVisit >= thresholds.Stage365:
		targetTag = "dormant_365d"
	case daysSinceLastVisit >= thresholds.Stage240:
		targetTag = "dormant_240d"
	case daysSinceLastVisit >= thresholds.Stage210:
		targetTag = "dormant_210d"
	case daysSinceLastVisit >= thresholds.Stage180:
		targetTag = "dormant_180d"
	}

	// 240日以上の休眠では CPM 休眠ステージタグ（cpm_dormant）も付与する
	needsCpmDormant := targetTag == "dormant_240d" || targetTag == "dormant_365d"

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for dormant sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}

	apiFailed := false
	for _, c := range cached {
		if isDormantTag(c.TagName) && c.TagName != targetTag {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale dormant tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delCacheErr != nil {
					slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", c.TagName)
				}
			}
		}
		// cpm_dormant は 240d 未満の場合に削除する
		if c.TagName == "cpm_dormant" && !needsCpmDormant {
			if delErr := client.RemoveTag(ctx, lineUserID, "cpm_dormant"); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale cpm_dormant tag", "error", delErr)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, "cpm_dormant"); delCacheErr != nil {
					slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", "cpm_dormant")
				}
			}
		}
	}

	if targetTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if addErr := client.AddTag(ctx, lineUserID, targetTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add dormant tag", "error", addErr, "tag", targetTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add dormant tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, targetTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert dormant tag cache", "error", upsertErr)
	}

	// 240日以上の場合は cpm_dormant も付与する
	if needsCpmDormant {
		if addErr := client.AddTag(ctx, lineUserID, "cpm_dormant"); addErr != nil {
			slog.ErrorContext(ctx, "failed to add cpm_dormant tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add cpm_dormant tag")
		} else if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, "cpm_dormant", "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert cpm_dormant tag cache", "error", upsertErr)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncVisitDormantTags は最終来院経過日数に基づき VISIT_* タグを差分同期する（FEAT-377）。
// VISIT タグは重複付与可（複数閾値を同時保持する）。
// visitDormantTagsForDays は経過日数に対して付与すべき VISIT_* タグのスライスを返す純粋関数。
func visitDormantTagsForDays(days int) []string {
	type th struct {
		tag string
		min int
	}
	thresholds := []th{
		{tag: visitTag120, min: 120},
		{tag: visitTag180, min: 180},
		{tag: visitTag220, min: 220},
		{tag: visitTag240, min: 240},
	}
	var tags []string
	for _, t := range thresholds {
		if days > t.min {
			tags = append(tags, t.tag)
		}
	}
	return tags
}

func (s *lstepTagSyncService) SyncVisitDormantTags(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for VISIT dormant tag sync", "error", err)
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
		slog.ErrorContext(ctx, "failed to build lstep client for VISIT dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	type visitThreshold struct {
		tag  string
		days int
	}
	thresholds := []visitThreshold{
		{tag: visitTag120, days: 120},
		{tag: visitTag180, days: 180},
		{tag: visitTag220, days: 220},
		{tag: visitTag240, days: 240},
	}

	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for VISIT dormant sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	cachedSet := make(map[string]struct{}, len(cached))
	for _, c := range cached {
		cachedSet[c.TagName] = struct{}{}
	}

	apiFailed := false
	for _, th := range thresholds {
		shouldHave := daysSinceLastVisit > th.days
		_, hasCached := cachedSet[th.tag]
		if shouldHave {
			if addErr := client.AddTag(ctx, lineUserID, th.tag); addErr != nil {
				slog.ErrorContext(ctx, "failed to add VISIT dormant tag", "error", addErr, "tag", th.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, th.tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert VISIT dormant tag cache", "error", cacheErr, "tag", th.tag)
			}
		} else if hasCached {
			if delErr := client.RemoveTag(ctx, lineUserID, th.tag); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove VISIT dormant tag", "error", delErr, "tag", th.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, th.tag); delCacheErr != nil {
				slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", th.tag)
			}
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
