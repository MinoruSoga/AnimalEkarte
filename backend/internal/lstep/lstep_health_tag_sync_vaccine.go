package lstep

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// syncVaccineDeadlineTagImpl はワクチン次回予定日が VaccineDeadlineDays 以内なら PREV_ワクチン期限 を
// 同期する（FEAT-379）。事前取得済み thresholds を使って処理する（PERF-1 N+1 解消用）。
//
//nolint:gocritic // hugeParam: thresholds は HealthPreventionThresholds を値型で受ける
func (s *lstepTagSyncService) syncVaccineDeadlineTagImpl(ctx context.Context, clinicID, ownerID uint64, thresholds model.HealthPreventionThresholds) error {
	return s.syncVaccineDeadlineTagWithInputs(ctx, clinicID, ownerID, thresholds, nil)
}

// syncVaccineDeadlineTagWithInputs は preloadedVaccinations が非 nil なら再取得しない（G2F-02 page bulk）。
//
//nolint:gocritic // hugeParam: thresholds は HealthPreventionThresholds を値型で受ける
func (s *lstepTagSyncService) syncVaccineDeadlineTagWithInputs(
	ctx context.Context,
	clinicID, ownerID uint64,
	thresholds model.HealthPreventionThresholds,
	preloadedVaccinations *[]model.Vaccination,
) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, PrevVaccineDeadlineTag)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var vaccinations []model.Vaccination
	if preloadedVaccinations != nil {
		vaccinations = *preloadedVaccinations
	} else {
		vaccinations, err = s.vacRepo.FindByOwner(ctx, clinicID, ownerID)
		if err != nil {
			slog.ErrorContext(ctx, "failed to find vaccinations for deadline tag", "error", err)
			return apperrors.Wrap(err, "failed to find vaccinations")
		}
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

	vaccineReason := ""
	if earliestNextDate != nil {
		vaccineReason = fmt.Sprintf("次回期限: %s", earliestNextDate.Format(time.DateOnly))
	}

	// DEC-35 / G2B-01: Remove 失敗も err 伝播（silent success 禁止）。
	if deadlineSoon {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevVaccineDeadlineTag, "vaccine deadline", vaccineReason, true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, PrevVaccineDeadlineTag, "vaccine deadline", vaccineReason, false); err != nil {
			return err
		}
	}

	s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	return nil
}

// SyncFilariaTag はフィラリア検査・予防薬処方に基づき PREV_フィラリア未完了 を同期する（FEAT-379）。
// 犬を飼育する飼い主のみ対象。検査・処方どちらかが完了していればタグを解除する。
