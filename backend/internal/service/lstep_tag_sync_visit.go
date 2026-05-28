package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/repository"
)

// SyncVisitCompletionTags は診療完了時の来院・LTV タグを同期する（BE-004）。
func (s *lstepTagSyncService) SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for visit tags sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	ltv, err := s.accountRepo.SumPaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sum paid amount", "error", err)
		return apperrors.Wrap(err, "failed to sum paid amount")
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID
	tags := buildVisitTags(summary, ltv)

	// 同一カテゴリ1タグ保持: 古い ltv_amount_* / visit_count_annual_* / first_visit_* / last_visit_* を解除（ISSUE-006）
	newTagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		newTagSet[t] = struct{}{}
	}
	apiFailed := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"ltv_amount_", "visit_count_annual_", "first_visit_", "last_visit_"}, newTagSet)

	for _, tag := range tags {
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add visit tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add visit tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert tag cache", "error", cacheErr, "tag", tag)
		}
	}

	// 来院完了でキャッシュ経由のdormantタグを削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for visit completion cleanup", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	for _, c := range cached {
		if isDormantTag(c.TagName) || isVisitDormantTag(c.TagName) || c.TagName == "cpm_dormant" {
			if delErr := client.RemoveTag(ctx, lineUserID, c.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove dormant tag on visit completion", "error", delErr, "tag", c.TagName)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
			}
		}
	}
	// レガシータグ（名前変更前との互換）
	for _, staleTag := range []string{"dormant", "noshow", "reserved"} {
		if delErr := client.RemoveTag(ctx, lineUserID, staleTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale legacy tag", "error", delErr, "tag", staleTag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, staleTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildVisitTags は来院サマリーから付与するタグ一覧を生成する。
func buildVisitTags(summary *repository.OwnerVisitSummary, ltv int64) []string {
	var tags []string
	if summary.FirstVisitAt != nil {
		tags = append(tags, "first_visit_"+summary.FirstVisitAt.Format("2006-01-02"))
	}
	if summary.LastVisitAt != nil {
		tags = append(tags, "last_visit_"+summary.LastVisitAt.Format("2006-01-02"))
	}
	tags = append(tags, ltvBracketTag(ltv), visitCountAnnualTag(summary.AnnualCount))
	return tags
}

func ltvBracketTag(ltv int64) string {
	switch {
	case ltv >= 80_000:
		return "ltv_amount_8"
	case ltv >= 50_000:
		return "ltv_amount_5"
	case ltv >= 20_000:
		return "ltv_amount_2"
	default:
		return "ltv_amount_0"
	}
}

func visitCountAnnualTag(count int64) string {
	switch {
	case count >= 10:
		return "visit_count_annual_10"
	case count >= 5:
		return "visit_count_annual_5"
	case count >= 3:
		return "visit_count_annual_3"
	case count >= 2:
		return "visit_count_annual_2"
	default:
		return "visit_count_annual_1"
	}
}

// isDormantTag は dormant 系タグかどうかを判定する。
func isDormantTag(tag string) bool {
	return tag == "dormant_180d" || tag == "dormant_210d" || tag == "dormant_240d" || tag == "dormant_365d"
}

// isVisitDormantTag は VISIT_* 系タグかどうかを判定する（FEAT-377）。
func isVisitDormantTag(tag string) bool {
	return tag == visitTag120 || tag == visitTag180 || tag == visitTag220 || tag == visitTag240
}

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

// SyncCPMStageTag は CPM ステージタグを同期する（BE-011）。
// cpm_version が "v2" のクリニックは SyncCPMStageTagV2 に委譲する（Q19）。
func (s *lstepTagSyncService) SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error {
	version, vErr := s.settingsSvc.GetCPMVersion(ctx, clinicID)
	if vErr != nil {
		slog.ErrorContext(ctx, "failed to get cpm_version for CPM dispatch", "error", vErr)
		return apperrors.Wrap(vErr, "failed to get cpm_version")
	}
	if version == "v2" {
		return s.SyncCPMStageTagV2(ctx, clinicID, ownerID)
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	ltv, err := s.accountRepo.SumPaidByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to sum paid amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to sum paid amount")
	}

	maxSingle, err := s.accountRepo.MaxSingleVisitAmountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get max single visit amount for CPM", "error", err)
		return apperrors.Wrap(err, "failed to get max single visit amount")
	}

	thresholds, err := s.settingsSvc.GetCPMV1Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v1 thresholds", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get cpm v1 thresholds")
	}

	daysSince := -1
	if summary.LastVisitAt != nil {
		daysSince = int(time.Since(*summary.LastVisitAt).Hours() / 24)
	}
	firstVisitDaysSince := 0
	if summary.FirstVisitAt != nil {
		firstVisitDaysSince = int(time.Since(*summary.FirstVisitAt).Hours() / 24)
	}

	stage := CalculateCPMStage(CPMData{
		TotalVisitCount:      summary.TotalCount,
		AnnualVisitCount:     summary.AnnualCount,
		DaysSinceVisit:       daysSince,
		LTVAmount:            ltv,
		FirstVisitDaysSince:  firstVisitDaysSince,
		MaxSingleVisitAmount: maxSingle,
		Thresholds:           thresholds,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	// 旧ステージタグをすべて削除してから新ステージを付与
	apiFailed := false
	for _, old := range allCPMStages {
		if string(old) == string(stage) {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

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
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName)
			}
		}
		// cpm_dormant は 240d 未満の場合に削除する
		if c.TagName == "cpm_dormant" && !needsCpmDormant {
			if delErr := client.RemoveTag(ctx, lineUserID, "cpm_dormant"); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove stale cpm_dormant tag", "error", delErr)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
			} else {
				_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, "cpm_dormant")
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

// SyncCPMStageTagV2 は来院回数ベース V2 CPM ステージタグを同期する（FEAT-377）。
func (s *lstepTagSyncService) SyncCPMStageTagV2(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for CPM V2 tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	summary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for CPM V2", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	v2Thresholds, err := s.settingsSvc.GetCPMV2Thresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get cpm v2 thresholds", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get cpm v2 thresholds")
	}

	stage := CalculateCPMStageV2(CPMStageV2Input{
		TotalVisitCount: summary.TotalCount,
		CPMV2Thresholds: v2Thresholds,
	})

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID
	apiFailed := false
	for _, old := range allCPMV2Stages {
		if old == stage {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, string(old)); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old CPM V2 stage tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, string(old))
		}
	}

	if addErr := client.AddTag(ctx, lineUserID, string(stage)); addErr != nil {
		slog.ErrorContext(ctx, "failed to add CPM V2 stage tag", "error", addErr, "stage", stage)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, "failed to add CPM V2 stage tag")
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, string(stage), "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert CPM V2 stage tag cache", "error", cacheErr)
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
// それ以外から解除する（FEAT-377）。
func (s *lstepTagSyncService) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{err}
	} else if skip {
		return 0, nil
	}

	revenues, err := s.accountRepo.FindOwnersByAnnualRevenue(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find revenues", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by annual revenue")}
	}

	topN := 0
	if len(revenues) > 0 {
		topN = (len(revenues)*20 + 99) / 100
	}
	topOwnerIDs := make(map[uint64]struct{}, topN)
	for i := 0; i < topN; i++ {
		topOwnerIDs[revenues[i].OwnerID] = struct{}{}
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return 0, []error{err}
	}
	if client == nil {
		return 0, nil
	}

	var errs []error
	count := 0
	for i := range owners {
		owner := &owners[i]
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			continue
		}
		lineUserID := *owner.LineUserID
		_, isTop := topOwnerIDs[owner.ID]
		if isTop {
			if addErr := client.AddTag(ctx, lineUserID, ltvTop20Tag); addErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to add tag", "owner_id", owner.ID, "error", addErr)
				s.notifyAPIFailure(ctx, client, clinicID, owner.ID, lineUserID)
				errs = append(errs, apperrors.Wrap(addErr, fmt.Sprintf("failed to add LTV top20 tag for owner %d", owner.ID)))
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, owner.ID, ltvTop20Tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to upsert tag cache", "owner_id", owner.ID, "error", cacheErr)
			}
		} else {
			cached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, owner.ID)
			if cacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to load cache", "owner_id", owner.ID, "error", cacheErr)
				errs = append(errs, apperrors.Wrap(cacheErr, fmt.Sprintf("failed to load tag cache for owner %d", owner.ID)))
				continue
			}
			hasTag := false
			for _, c := range cached {
				if c.TagName == ltvTop20Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
				count++
				continue
			}
			if delErr := client.RemoveTag(ctx, lineUserID, ltvTop20Tag); delErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to remove tag", "owner_id", owner.ID, "error", delErr)
				s.notifyAPIFailure(ctx, client, clinicID, owner.ID, lineUserID)
				errs = append(errs, apperrors.Wrap(delErr, fmt.Sprintf("failed to remove LTV top20 tag for owner %d", owner.ID)))
				continue
			}
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, owner.ID, ltvTop20Tag); delCacheErr != nil {
				slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to delete tag cache", "owner_id", owner.ID, "error", delCacheErr)
			}
		}
		s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
		count++
	}
	return count, errs
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
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, th.tag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
