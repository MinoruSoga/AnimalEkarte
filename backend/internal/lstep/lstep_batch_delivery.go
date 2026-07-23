package lstep

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
)

func (s *lstepBatchService) runDeliveryTriggersForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	asOf := time.Now()
	if s.nowFn != nil {
		asOf = s.nowFn()
	}
	return s.runDeliveryTriggersForClinicAt(ctx, clinicID, asOf)
}

func (s *lstepBatchService) runDeliveryTriggersForClinicAt(
	ctx context.Context,
	clinicID uint64,
	asOf time.Time,
) (int, []error) {
	if s.lstepDeliveryTrigger == nil {
		return 0, nil
	}
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
	asOf := time.Now()
	if s.nowFn != nil {
		asOf = s.nowFn()
	}
	_, err := s.runDeliveryTriggerBatchAllClinicsWithResult(ctx, asOf, nil)
	return err
}

// RunDeliveryTriggerBatchAllClinicsAt evaluates the fixed 10:00 JST gate and
// every trigger target against the durable scheduler's immutable timestamp.
func (s *lstepBatchService) RunDeliveryTriggerBatchAllClinicsAt(
	ctx context.Context,
	scheduledAt time.Time,
	runID string,
) BatchRunResult {
	if s.lstepDeliveryTrigger == nil {
		slog.ErrorContext(ctx, "delivery trigger batch: delivery trigger service is not configured")
		return failedBatchRunResult()
	}
	scheduledAt = scheduledAt.UTC()
	result, _ := s.runDeliveryTriggerBatchAllClinicsWithResult(
		ctx,
		scheduledAt,
		scheduledBatchMetadata(scheduledAt, runID),
	)
	return result
}

func (s *lstepBatchService) runDeliveryTriggerBatchAllClinicsWithResult(
	ctx context.Context,
	asOf time.Time,
	metadata map[string]any,
) (BatchRunResult, error) {
	if asOf.In(config.JST).Hour() != deliveryTriggerHourJST {
		return BatchRunResult{}, nil
	}

	return s.runBatchAllClinicsWithResult(
		ctx,
		"delivery trigger batch",
		"delivery trigger batch",
		"fired triggers",
		"batch_delivery_trigger",
		metadata,
		func(ctx context.Context, clinicID uint64) (int, []error) {
			return s.runDeliveryTriggersForClinicAt(ctx, clinicID, asOf)
		},
	)
}

// RunDormantDetectionAllClinics は全クリニックに対して休眠検知を実行する（02:00 JST バッチ）。
