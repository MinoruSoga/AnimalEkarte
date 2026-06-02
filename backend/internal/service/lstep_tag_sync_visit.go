package service

import (
	"context"
	"fmt"
	"log/slog"

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
				if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, c.TagName); delCacheErr != nil {
					slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", c.TagName)
				}
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
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, staleTag); delCacheErr != nil {
				slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", staleTag)
			}
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
