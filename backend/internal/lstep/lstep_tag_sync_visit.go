package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// SyncVisitCompletionTags は診療完了時の来院・LTV タグを同期する（BE-004）。
func (s *lstepTagSyncService) SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "visit completion")
	if err != nil {
		return err
	}
	if !ok {
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

	tags := buildVisitTags(summary, ltv)

	// 同一カテゴリ1タグ保持: 古い ltv_amount_* / visit_count_annual_* / first_visit_* / last_visit_* を解除（ISSUE-006）
	newTagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		newTagSet[t] = struct{}{}
	}
	apiFailed, err := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"ltv_amount_", "visit_count_annual_", "first_visit_", "last_visit_"}, newTagSet)
	if err != nil {
		// LSA-11: cache-read failure — zero category API calls, surface the error.
		return err
	}

	for _, tag := range tags {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tag, "visit completion", "", true); err != nil {
			return err
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
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "visit completion dormant", "", false); err != nil {
				apiFailed = true
			}
		}
	}
	// レガシータグ（名前変更前との互換）
	for _, staleTag := range []string{"dormant", "noshow", "reserved"} {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, staleTag, "visit completion legacy", "", false); err != nil {
			apiFailed = true
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildVisitTags は来院サマリーから付与するタグ一覧を生成する。
func buildVisitTags(summary *medicalrecord.OwnerVisitSummary, ltv int64) []string {
	var tags []string
	if summary.FirstVisitAt != nil {
		tags = append(tags, "first_visit_"+summary.FirstVisitAt.Format(time.DateOnly))
	}
	if summary.LastVisitAt != nil {
		tags = append(tags, "last_visit_"+summary.LastVisitAt.Format(time.DateOnly))
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
