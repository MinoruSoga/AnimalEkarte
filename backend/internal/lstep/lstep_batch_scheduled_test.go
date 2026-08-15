package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/reservation"
)

type scheduledNoShowRepository struct {
	candidatesAt time.Time
	markedAt     time.Time
}

func (r *scheduledNoShowRepository) FindNoShowCandidates(
	context.Context,
	uint64,
) ([]model.Reservation, error) {
	return nil, errors.New("legacy no-show query must not be used by scheduled execution")
}

func (r *scheduledNoShowRepository) FindNoShowCandidatesAt(
	_ context.Context,
	_ uint64,
	evaluatedAt time.Time,
) ([]model.Reservation, error) {
	r.candidatesAt = evaluatedAt
	return []model.Reservation{{ID: 41}}, nil
}

func (r *scheduledNoShowRepository) MarkNoShow(
	context.Context,
	uint64,
	uint64,
) (reservation.NoShowTransition, error) {
	return reservation.NoShowTransition{}, errors.New("legacy no-show CAS must not be used by scheduled execution")
}

func (r *scheduledNoShowRepository) MarkNoShowAt(
	_ context.Context,
	_ uint64,
	_ uint64,
	evaluatedAt time.Time,
) (reservation.NoShowTransition, error) {
	r.markedAt = evaluatedAt
	return reservation.NoShowTransition{
		Changed:        true,
		PreviousStatus: model.ReservationStatusConfirmed,
	}, nil
}

type scheduledDormantRepository struct {
	capturedAt time.Time
}

func (r *scheduledDormantRepository) FindDormantOwnerEntries(
	context.Context,
	uint64,
	int,
) ([]medicalrecord.DormantOwnerEntry, error) {
	return nil, nil
}

func (r *scheduledDormantRepository) FindDormantOwnerEntriesCursor(
	context.Context,
	uint64,
	int,
	uint64,
	int,
) ([]medicalrecord.DormantOwnerEntry, error) {
	return nil, errors.New("legacy dormant query must not be used by scheduled execution")
}

func (r *scheduledDormantRepository) FindDormantOwnerEntriesCursorAt(
	_ context.Context,
	_ uint64,
	_ int,
	_ uint64,
	_ int,
	evaluatedAt time.Time,
) ([]medicalrecord.DormantOwnerEntry, error) {
	r.capturedAt = evaluatedAt
	return []medicalrecord.DormantOwnerEntry{{OwnerID: 55, DaysSince: 200}}, nil
}

type scheduledAuditSpy struct {
	metadata any
	err      error
}

func (s *scheduledAuditSpy) LogLstepOperationWithMetadata(
	_ context.Context,
	_ uint64,
	_ *uint64,
	_ string,
	_ string,
	_ *uint64,
	metadata any,
) error {
	s.metadata = metadata
	return s.err
}

func TestRunNoShowCheckAllClinicsAt_UsesScheduledTimeAndStableRunIdentity(t *testing.T) {
	scheduledAt := time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC)
	reservationRepo := &scheduledNoShowRepository{}
	noShowAudit := &batchNoShowAuditTxLogger{}
	audit := &scheduledAuditSpy{}
	service := NewLstepBatchService(
		reservationRepo,
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 7}}, nil
			},
		},
		&batchMockMedRecordRepo{},
		audit,
		&mockLstepSettingsService{},
		nil,
		batchImmediateTransactor{},
		noShowAudit,
	)

	result := service.RunNoShowCheckAllClinicsAt(
		context.Background(),
		scheduledAt,
		"animalekarte-scheduler-v1:1785200400000:no_show",
	)

	assert.Equal(t, BatchRunResult{Processed: 1, Succeeded: 1}, result)
	assert.Equal(t, scheduledAt, reservationRepo.candidatesAt)
	assert.Equal(t, scheduledAt, reservationRepo.markedAt)
	require.Len(t, noShowAudit.entries, 1)
	assert.Equal(t, scheduledAt, noShowAudit.entries[0].EvaluatedAt)
	assert.Equal(
		t,
		"animalekarte-scheduler-v1:1785200400000:no_show",
		noShowAudit.entries[0].BatchRunID,
	)
	meta, ok := audit.metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "animalekarte-scheduler-v1:1785200400000:no_show", meta["batch_run_id"])
	assert.Equal(t, scheduledAt.UnixMilli(), meta["scheduled_time"])
}

func TestRunDormantDetectionAllClinicsAt_UsesScheduledTimeForSelection(t *testing.T) {
	scheduledAt := time.Date(2026, 7, 24, 17, 0, 0, 0, time.UTC)
	medRecordRepo := &scheduledDormantRepository{}
	service := NewLstepBatchService(
		&batchMockReservationRepo{},
		&batchMockTagSyncSvc{},
		&mockClinicRepository{
			findAllFn: func(context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 8}}, nil
			},
		},
		medRecordRepo,
		&scheduledAuditSpy{},
		&mockLstepSettingsService{},
		nil,
		batchImmediateTransactor{},
		&batchNoShowAuditTxLogger{},
	)

	result := service.RunDormantDetectionAllClinicsAt(
		context.Background(),
		scheduledAt,
		"animalekarte-scheduler-v1:1785258000000:dormant",
	)

	assert.Equal(t, BatchRunResult{Processed: 1, Succeeded: 1}, result)
	assert.Equal(t, scheduledAt, medRecordRepo.capturedAt)
}

func TestRunDeliveryTriggerBatchAllClinicsAt_UsesScheduledTimeForEveryTrigger(t *testing.T) {
	scheduledAt := time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC) // 10:00 JST
	var captured []time.Time
	trigger := &mockLstepDeliveryTriggerBatch{
		triggerFn: func(_ string, _ uint64, asOf time.Time) (int, []error) {
			captured = append(captured, asOf)
			return 1, nil
		},
	}
	service := &lstepBatchService{
		clinicRepo: &mockClinicRepository{
			findAllFn: func(context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 9}}, nil
			},
		},
		settingsSvc:          &mockLstepSettingsService{},
		auditSvc:             &scheduledAuditSpy{},
		lstepDeliveryTrigger: trigger,
	}

	result := service.RunDeliveryTriggerBatchAllClinicsAt(
		context.Background(),
		scheduledAt,
		"animalekarte-scheduler-v1:1785200400000:delivery",
	)

	assert.Equal(t, BatchRunResult{Processed: 13, Succeeded: 13}, result)
	require.Len(t, captured, 13)
	for _, asOf := range captured {
		assert.Equal(t, scheduledAt, asOf)
	}
}

func TestRunDeliveryTriggerBatchAllClinicsAt_MissingTriggerFailsClosed(t *testing.T) {
	service := &lstepBatchService{
		clinicRepo: &mockClinicRepository{
			findAllFn: func(context.Context) ([]model.Clinic, error) {
				return []model.Clinic{{ID: 9}}, nil
			},
		},
		settingsSvc: &mockLstepSettingsService{},
		auditSvc:    &scheduledAuditSpy{},
	}

	result := service.RunDeliveryTriggerBatchAllClinicsAt(
		context.Background(),
		time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC),
		"animalekarte-scheduler-v1:1785200400000:delivery",
	)

	assert.Equal(t, BatchRunResult{Processed: 1, Failed: 1}, result)
}

func TestScheduledBatchResult_ReportsPreprocessingAndPartialFailures(t *testing.T) {
	t.Run("clinic fetch failure is one failed unit", func(t *testing.T) {
		service := newBatchService(
			&batchMockReservationRepo{},
			&batchMockTagSyncSvc{},
			&mockClinicRepository{
				findAllFn: func(context.Context) ([]model.Clinic, error) {
					return nil, errors.New("database unavailable")
				},
			},
			&batchMockMedRecordRepo{},
		)

		result := service.RunNoShowCheckAllClinicsAt(
			context.Background(),
			time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC),
			"run",
		)

		assert.Equal(t, BatchRunResult{Processed: 1, Failed: 1}, result)
	})

	t.Run("per clinic and audit errors make a successful item partial", func(t *testing.T) {
		audit := &scheduledAuditSpy{err: errors.New("audit unavailable")}
		service := NewLstepBatchService(
			&scheduledNoShowRepository{},
			&batchMockTagSyncSvc{},
			&mockClinicRepository{
				findAllFn: func(context.Context) ([]model.Clinic, error) {
					return []model.Clinic{{ID: 1}}, nil
				},
			},
			&batchMockMedRecordRepo{},
			audit,
			&mockLstepSettingsService{},
			nil,
			batchImmediateTransactor{},
			&batchNoShowAuditTxLogger{},
		)

		result := service.RunNoShowCheckAllClinicsAt(
			context.Background(),
			time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC),
			"run",
		)

		assert.Equal(t, BatchRunResult{Processed: 2, Succeeded: 1, Failed: 1}, result)
	})

	t.Run("settings read error is not reported as success", func(t *testing.T) {
		service := NewLstepBatchService(
			&scheduledNoShowRepository{},
			&batchMockTagSyncSvc{},
			&mockClinicRepository{
				findAllFn: func(context.Context) ([]model.Clinic, error) {
					return []model.Clinic{{ID: 1}}, nil
				},
			},
			&batchMockMedRecordRepo{},
			&scheduledAuditSpy{},
			&mockLstepSettingsService{
				isSyncEnabledFn: func(context.Context, uint64) (bool, error) {
					return false, errors.New("settings unavailable")
				},
			},
			nil,
			batchImmediateTransactor{},
			&batchNoShowAuditTxLogger{},
		)

		result := service.RunNoShowCheckAllClinicsAt(
			context.Background(),
			time.Date(2026, 7, 24, 1, 0, 0, 0, time.UTC),
			"run",
		)

		assert.Equal(t, BatchRunResult{Processed: 1, Failed: 1}, result)
	})
}
