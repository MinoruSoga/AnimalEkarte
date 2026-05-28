package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *lstepTagSyncService) SyncVaccineDeadlineTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for vaccine deadline tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	vaccinations, err := s.vacRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccinations for deadline tag", "error", err)
		return apperrors.Wrap(err, "failed to find vaccinations")
	}

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for vaccine deadline tag", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
	}
	now := time.Now()
	deadline := now.AddDate(0, 0, thresholds.VaccineDeadline)
	deadlineSoon := false
	var earliestNextDate *time.Time
	for i := range vaccinations {
		nd := vaccinations[i].NextDate
		if nd == nil {
			continue
		}
		if !nd.Before(now) && !nd.After(deadline) {
			deadlineSoon = true
			if earliestNextDate == nil || nd.Before(*earliestNextDate) {
				copied := *nd
				earliestNextDate = &copied
			}
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if deadlineSoon {
		if addErr := client.AddTag(ctx, lineUserID, PrevVaccineDeadlineTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add vaccine deadline tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add vaccine deadline tag")
		}
		vaccineReason := ""
		if earliestNextDate != nil {
			vaccineReason = fmt.Sprintf("次回期限: %s", earliestNextDate.Format("2006-01-02"))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, PrevVaccineDeadlineTag, "auto", vaccineReason); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert vaccine deadline tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, PrevVaccineDeadlineTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove vaccine deadline tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, PrevVaccineDeadlineTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncFilariaTag はフィラリア検査・予防薬処方に基づき PREV_フィラリア未完了 を同期する（FEAT-379）。
// 犬を飼育する飼い主のみ対象。検査・処方どちらかが完了していればタグを解除する。
