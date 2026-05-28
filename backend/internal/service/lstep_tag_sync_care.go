package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
// refill_due_* タグを更新する（BE-009）。duration_days < 7 の場合は prescribed_at + 1 日を使用する。
// 最新の refill_due が現在日時を過ぎている場合は refill_due_* タグをすべて削除して終了する。
func (s *lstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for prescription tag sync", "error", err)
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
		slog.ErrorContext(ctx, "failed to build lstep client for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	prescriptions, err := s.prescriptionRepo.FindActiveByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load prescriptions for tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load prescriptions")
	}

	// 補充推奨日の計算: prescribed_at + duration_days - 7。duration_days < 7 なら prescribed_at + 1
	var latestRefillDue *time.Time
	for i := range prescriptions {
		p := &prescriptions[i]
		var refill time.Time
		if p.DurationDays < 7 {
			refill = p.PrescribedAt.AddDate(0, 0, 1)
		} else {
			refill = p.PrescribedAt.AddDate(0, 0, p.DurationDays-7)
		}
		if latestRefillDue == nil || refill.After(*latestRefillDue) {
			t := refill
			latestRefillDue = &t
		}
	}

	// 旧 refill_due_* タグをキャッシュ経由で削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, tagPrefixRefillDue) {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale refill_due tag", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete refill_due tag cache", "error", delErr, "tag", c.TagName)
			}
		}
	}

	// 補充推奨日が未来にある場合のみ新タグを付与
	if latestRefillDue == nil || !latestRefillDue.After(time.Now()) {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := tagPrefixRefillDue + latestRefillDue.Format("2006-01-02")
	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add refill_due tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add refill_due tag")
	}
	if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); upsertErr != nil {
		slog.ErrorContext(ctx, "failed to upsert refill_due tag cache", "error", upsertErr)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

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

// conditionTagMapFromMappings は DB レコードを疾患コード→タグ名マップに変換する（純粋関数）。
func conditionTagMapFromMappings(mappings []*model.LstepConditionTagMapping) map[string]string {
	m := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		m[mapping.ConditionCode] = mapping.TagName
	}
	return m
}

// SyncChronicConditionTags は慢性疾患タグを差分同期する（BE-012）。
func (s *lstepTagSyncService) SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for chronic condition tag sync", "error", err)
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
		slog.ErrorContext(ctx, "failed to build lstep client for chronic condition tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 慢性疾患コードマッピングを DB から取得
	var conditionMap map[string]string
	if s.tagConfigRepo != nil {
		mappings, loadErr := s.tagConfigRepo.FindAllConditionTagMappings(ctx)
		if loadErr != nil {
			slog.ErrorContext(ctx, "failed to load condition tag mappings", "error", loadErr)
			return apperrors.Wrap(loadErr, "failed to load condition tag mappings")
		}
		conditionMap = conditionTagMapFromMappings(mappings)
	} else {
		conditionMap = map[string]string{}
	}

	// 目標タグセットを構築
	activeTags := make(map[string]bool, len(activeConditionCodes))
	for _, code := range activeConditionCodes {
		if tag, ok := conditionMap[code]; ok {
			activeTags[tag] = true
		}
	}

	// 現在の chronic_* キャッシュを取得
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find cached tags for chronic sync", "error", err)
		return apperrors.Wrap(err, "failed to find cached tags")
	}

	existingSet := make(map[string]bool, len(cached))
	apiFailed := false
	for _, t := range cached {
		if !strings.HasPrefix(t.TagName, tagPrefixChronic) {
			continue
		}
		existingSet[t.TagName] = true

		// アクティブでないタグを解除
		if activeTags[t.TagName] {
			continue
		}
		if rmErr := client.RemoveTag(ctx, lineUserID, t.TagName); rmErr != nil {
			slog.WarnContext(ctx, "failed to remove chronic tag via lstep api", "tag", t.TagName, "error", rmErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, t.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to delete chronic tag cache", "tag", t.TagName, "error", delErr)
		}
	}

	// 未付与のアクティブタグを付与
	for tagName := range activeTags {
		if existingSet[tagName] {
			continue
		}
		if addErr := client.AddTag(ctx, lineUserID, tagName); addErr != nil {
			slog.WarnContext(ctx, "failed to add chronic tag via lstep api", "tag", tagName, "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tagName, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert chronic tag cache", "tag", tagName, "error", upsertErr)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// ResyncOwnerCheckupTags は飼い主の生存健診記録から checkup_done_* / next_checkup_* タグを再構築する（ISSUE-004）。
// 健診の更新・削除後に呼び出すこと。同一健診種別では最新検査日のみ保持。next_checkup は最遠の next_date のみ保持。
func (s *lstepTagSyncService) ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for checkup resync", "error", err)
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
		if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale checkup tag", "error", delErr, "tag", c.TagName)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
			continue
		}
		if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delErr != nil {
			slog.ErrorContext(ctx, "failed to delete stale checkup tag cache", "error", delErr, "tag", c.TagName)
		}
	}

	for tag := range newTagSet {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add checkup tag on resync", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add checkup tag %s", tag))
		}
		if upsertErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); upsertErr != nil {
			slog.ErrorContext(ctx, "failed to upsert checkup tag cache on resync", "error", upsertErr, "tag", tag)
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
		tagSet[tagPrefixNextCheckup+latestNext.Format("2006-01-02")] = struct{}{}
	}
	return tagSet
}
