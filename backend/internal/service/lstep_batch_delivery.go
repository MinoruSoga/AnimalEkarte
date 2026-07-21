package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

func (s *lstepBatchService) runDeliveryTriggersForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if s.lstepDeliveryTrigger == nil {
		return 0, nil
	}
	asOf := time.Now()
	type batchFn func(context.Context, uint64, time.Time) (int, []error)
	triggers := []batchFn{
		s.lstepDeliveryTrigger.TriggerFirstVisitFollowUp3D,
		s.lstepDeliveryTrigger.TriggerFirstVisitFollowUp7D,
		s.lstepDeliveryTrigger.TriggerNextVisitReminder,
		s.lstepDeliveryTrigger.TriggerVaccineDeadline60,
		s.lstepDeliveryTrigger.TriggerVaccineDeadline30,
		s.lstepDeliveryTrigger.TriggerBirthdayMessage,
		s.lstepDeliveryTrigger.TriggerDormantPrevention180,
		s.lstepDeliveryTrigger.TriggerDormantPrevention210,
		s.lstepDeliveryTrigger.TriggerDormantPrevention240,
		s.lstepDeliveryTrigger.TriggerDormantPrevention365,
		s.lstepDeliveryTrigger.TriggerFilariaAlert,
		s.lstepDeliveryTrigger.TriggerFleaTickAlert,
		s.lstepDeliveryTrigger.TriggerFoodRefillReminder,
	}
	totalFired := 0
	var allErrs []error
	for _, fn := range triggers {
		n, errs := fn(ctx, clinicID, asOf)
		totalFired += n
		allErrs = append(allErrs, errs...)
	}
	return totalFired, allErrs
}

// RunDeliveryTriggerBatchAllClinics は全クリニックの自動配信トリガーバッチを実行する。
// 仕様 §6.4 により配信時刻は 10:00 JST 固定。
func (s *lstepBatchService) RunDeliveryTriggerBatchAllClinics(ctx context.Context) error {
	nowHour := s.nowFn().In(config.JST).Hour()

	clinics, err := s.clinicRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger batch: failed to fetch clinics", "error", err)
		return apperrors.Wrap(err, "failed to fetch clinics for delivery trigger batch")
	}

	for i := range clinics {
		clinic := &clinics[i]
		if nowHour != deliveryTriggerHourJST {
			continue
		}
		if s.settingsSvc != nil {
			enabled, syncErr := s.settingsSvc.IsSyncEnabled(ctx, clinic.ID)
			if syncErr != nil {
				slog.ErrorContext(ctx, "delivery trigger batch: failed to check sync enabled", "clinic_id", clinic.ID, "error", syncErr)
				continue
			}
			if !enabled {
				continue
			}
		}
		count, errs := s.runDeliveryTriggersForClinic(ctx, clinic.ID)
		if len(errs) > 0 {
			slog.ErrorContext(ctx, "delivery trigger batch: partial errors", "clinic_id", clinic.ID, "error_count", len(errs))
		}
		if count > 0 {
			slog.InfoContext(ctx, "delivery trigger batch: fired triggers", "clinic_id", clinic.ID, "count", count)
			if err := s.auditSvc.LogLstepOperationWithMetadata(ctx, clinic.ID, nil,
				"batch_delivery_trigger", "clinic", &clinic.ID,
				map[string]any{
					"operation":       "batch_delivery_trigger",
					"processed_count": count,
					"error_count":     len(errs),
				},
			); err != nil {
				slog.WarnContext(ctx, "audit log failed for delivery trigger batch", "error", err, "clinic_id", clinic.ID)
			}
		}
	}
	return nil
}

// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行する（02:00 JST バッチ）。
