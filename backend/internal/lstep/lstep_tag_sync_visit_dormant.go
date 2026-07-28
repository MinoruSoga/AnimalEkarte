package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncDormantTagsWithThresholds は最終来院からの経過日数に基づき dormant_* タグを差分同期する（BE-005）。
// 事前取得済みの閾値を使う（PERF-2 N+1 解消用）。
func (s *lstepTagSyncService) SyncDormantTagsWithThresholds(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "dormant")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for dormant tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 付与すべき dormant タグを決定（閾値はパラメータで受け取る）
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
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "dormant", "", false); err != nil {
				apiFailed = true
			}
		}
		// cpm_dormant は 240d 未満の場合に削除する
		if c.TagName == "cpm_dormant" && !needsCpmDormant {
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, "cpm_dormant", "cpm dormant", "", false); err != nil {
				apiFailed = true
			}
		}
	}

	if targetTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, targetTag, "dormant", "", true); err != nil {
		return err
	}

	// 240日以上の場合は cpm_dormant も付与する
	if needsCpmDormant {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, "cpm_dormant", "cpm dormant", "", true); err != nil {
			return err
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
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "VISIT dormant")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, th.tag, "VISIT dormant", "", true); err != nil {
				apiFailed = true
				continue
			}
		} else if hasCached {
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, th.tag, "VISIT dormant", "", false); err != nil {
				apiFailed = true
				continue
			}
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
