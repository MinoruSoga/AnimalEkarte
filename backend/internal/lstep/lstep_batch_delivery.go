package lstep

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
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
	type batchEntry struct {
		triggerType string
		fn          batchFn
	}
	// G2B-03 Option B: run high-priority triggers first (priority ASC) so demote is rare.
	// Clinic-specific overrides are still enforced via applySuppression at process time.
	triggers := []batchEntry{
		{model.TriggerTypeFirstVisitFollowUp3D, s.lstepDeliveryTrigger.TriggerFirstVisitFollowUp3D},
		{model.TriggerTypeFirstVisitFollowUp7D, s.lstepDeliveryTrigger.TriggerFirstVisitFollowUp7D},
		{model.TriggerTypeNextVisitReminder, s.lstepDeliveryTrigger.TriggerNextVisitReminder},
		{model.TriggerTypeVaccineDeadline60, s.lstepDeliveryTrigger.TriggerVaccineDeadline60},
		{model.TriggerTypeVaccineDeadline30, s.lstepDeliveryTrigger.TriggerVaccineDeadline30},
		{model.TriggerTypeBirthdayMessage, s.lstepDeliveryTrigger.TriggerBirthdayMessage},
		{model.TriggerTypeDormantPrevention180, s.lstepDeliveryTrigger.TriggerDormantPrevention180},
		{model.TriggerTypeDormantPrevention210, s.lstepDeliveryTrigger.TriggerDormantPrevention210},
		{model.TriggerTypeDormantPrevention240, s.lstepDeliveryTrigger.TriggerDormantPrevention240},
		{model.TriggerTypeDormantPrevention365, s.lstepDeliveryTrigger.TriggerDormantPrevention365},
		{model.TriggerTypeFilariaAlert, s.lstepDeliveryTrigger.TriggerFilariaAlert},
		{model.TriggerTypeFleaTickAlert, s.lstepDeliveryTrigger.TriggerFleaTickAlert},
		{model.TriggerTypeFoodRefillReminder, s.lstepDeliveryTrigger.TriggerFoodRefillReminder},
	}
	sort.SliceStable(triggers, func(i, j int) bool {
		return defaultDeliveryTriggerPriority(triggers[i].triggerType) < defaultDeliveryTriggerPriority(triggers[j].triggerType)
	})
	totalFired := 0
	var allErrs []error
	for _, entry := range triggers {
		n, errs := entry.fn(ctx, clinicID, asOf)
		totalFired += n
		allErrs = append(allErrs, errs...)
	}
	return totalFired, allErrs
}

func defaultDeliveryTriggerPriority(triggerType string) int {
	if p, ok := model.DefaultTriggerPriorities[triggerType]; ok {
		return p
	}
	return model.DefaultPriorityFallback
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
