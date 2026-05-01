package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// hasVaccineDeadlineSoon は vaccinations の中に NextDate が now から days 日以内のものがあれば true を返す。
// pure function — モックなしでテスト可能。
func hasVaccineDeadlineSoon(vaccinations []model.Vaccination, now time.Time, days int) bool {
	deadline := now.AddDate(0, 0, days)
	for i := range vaccinations {
		nd := vaccinations[i].NextDate
		if nd == nil {
			continue
		}
		// now <= NextDate <= deadline
		if !nd.Before(now) && !nd.After(deadline) {
			return true
		}
	}
	return false
}

// SyncHealthcheckTags は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 を同期する（FEAT-379）。
// SPEC-002 Q5 確定待ち: HealthCheckupCodes が空なら noop。
func (s *lstepTagSyncService) SyncHealthcheckTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	if len(HealthCheckupCodes) == 0 {
		return nil
	}
	// TODO: SPEC-002 Q5 確定後に実装
	return nil
}

// SyncAnnual4CheckupTag は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 を付与する（FEAT-379）。
// SPEC-002 Q5 確定待ち: HealthCheckupCodes が空なら noop。
func (s *lstepTagSyncService) SyncAnnual4CheckupTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	if len(HealthCheckupCodes) == 0 {
		return nil
	}
	// TODO: SPEC-002 Q5 確定後に実装
	return nil
}

// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内なら
// PREV_ワクチン期限 を付与し、それ以外は解除する（FEAT-379）。
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

	deadlineSoon := hasVaccineDeadlineSoon(vaccinations, time.Now(), VaccineDeadlineDays)

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
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, PrevVaccineDeadlineTag, "auto"); cacheErr != nil {
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
// SPEC-002 Q5 確定待ち: FilariaTestCodes / FilariaPrescriptionCodes が共に空なら noop。
func (s *lstepTagSyncService) SyncFilariaTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	if len(FilariaTestCodes) == 0 && len(FilariaPrescriptionCodes) == 0 {
		return nil
	}
	// TODO: SPEC-002 Q5 確定後に実装
	return nil
}

// SyncFleaTickTag はノミ・マダニ駆除薬処方に基づき PREV_ノミダニ対象 を同期する（FEAT-379）。
// SPEC-002 Q5 確定待ち: FleaTickPrescriptionCodes が空なら noop。
func (s *lstepTagSyncService) SyncFleaTickTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	if len(FleaTickPrescriptionCodes) == 0 {
		return nil
	}
	// TODO: SPEC-002 Q5 確定後に実装
	return nil
}

// SyncFoodPurchaseTag はフード購入履歴に基づき LTV_フード購入あり を同期する（FEAT-379）。
// SPEC-002 Q5 確定待ち: FoodPurchaseCodes が空なら noop。
func (s *lstepTagSyncService) SyncFoodPurchaseTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	if len(FoodPurchaseCodes) == 0 {
		return nil
	}
	// TODO: SPEC-002 Q5 確定後に実装
	return nil
}

// SyncSpecialCheckupCandidateTag は専門検診候補に HLTH_専門検診候補 タグを付与する（FEAT-379）。
// SPEC-002 Q6 確定待ち: 常に noop。
func (s *lstepTagSyncService) SyncSpecialCheckupCandidateTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	// TODO: SPEC-002 Q6 確定後に実装
	return nil
}

// SyncHealthPreventionTagsForClinic は指定クリニックの全飼い主に対して
// 健診・予防・物販タグを一括同期する（FEAT-379 バッチエントリポイント）。
func (s *lstepTagSyncService) SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{apperrors.Wrap(err, "failed to check lstep sync enabled")}
	} else if skip {
		return 0, nil
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	var errs []error
	count := 0
	for i := range owners {
		ownerID := owners[i].ID
		syncFns := []struct {
			name string
			fn   func() error
		}{
			{"SyncHealthcheckTags", func() error { return s.SyncHealthcheckTags(ctx, clinicID, ownerID) }},
			{"SyncAnnual4CheckupTag", func() error { return s.SyncAnnual4CheckupTag(ctx, clinicID, ownerID) }},
			{"SyncVaccineDeadlineTag", func() error { return s.SyncVaccineDeadlineTag(ctx, clinicID, ownerID) }},
			{"SyncFilariaTag", func() error { return s.SyncFilariaTag(ctx, clinicID, ownerID) }},
			{"SyncFleaTickTag", func() error { return s.SyncFleaTickTag(ctx, clinicID, ownerID) }},
			{"SyncFoodPurchaseTag", func() error { return s.SyncFoodPurchaseTag(ctx, clinicID, ownerID) }},
			{"SyncSpecialCheckupCandidateTag", func() error {
				return s.SyncSpecialCheckupCandidateTag(ctx, clinicID, ownerID)
			}},
		}
		ownerFailed := false
		for _, sf := range syncFns {
			if syncErr := sf.fn(); syncErr != nil {
				slog.ErrorContext(ctx, "health-prevention batch: sync failed",
					"clinic_id", clinicID, "owner_id", ownerID, "method", sf.name, "error", syncErr)
				errs = append(errs, apperrors.Wrap(syncErr, sf.name))
				ownerFailed = true
			}
		}
		if !ownerFailed {
			count++
		}
	}
	return count, errs
}
