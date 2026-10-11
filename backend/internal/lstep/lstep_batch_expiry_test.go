package lstep

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

type expiryBatchRepository struct {
	batchMockReservationRepo
	findFn func(context.Context, uint64, time.Time, uint64, int) ([]uint64, error)
}

func (r *expiryBatchRepository) FindOwnerIDsWithExpiredTentativeTagCursor(ctx context.Context, clinicID uint64, asOf time.Time, after uint64, limit int) ([]uint64, error) {
	if r.findFn != nil {
		return r.findFn(ctx, clinicID, asOf, after, limit)
	}
	return nil, nil
}

func TestDeliveryBatch_RepairsExpiredTentativeTags(t *testing.T) {
	for _, later := range []bool{false, true} {
		t.Run(map[bool]string{false: "remove when no current appointment", true: "replace with existing later booking"}[later], func(t *testing.T) {
			day1 := time.Date(2026, 10, 11, 1, 0, 0, 0, time.UTC)
			day2 := day1.AddDate(0, 0, 1)
			future := day1.AddDate(0, 0, 2)
			expired := false
			tags := map[string]bool{}
			lineID := "mock-only"
			client := &mockLstepAPIClient{
				addTagFn:    func(_ context.Context, _, tag string) error { tags[tag] = true; return nil },
				removeTagFn: func(_ context.Context, _, tag string) error { delete(tags, tag); return nil },
			}
			syncer := &lstepTagSyncService{
				settingsSvc: &mockLstepSettingsService{},
				ownerRepo: &mockOwnerRepository{findByIDFn: func(_ context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(2), ownerID)
					return &model.Owner{ID: ownerID, LineUserID: &lineID}, nil
				}},
				buildClientFn: func(context.Context, uint64) (lstepapi.Client, error) { return client, nil },
				nextVisitAppointmentRepo: &mockNextVisitAppointmentRepo{findNextAppointmentForOwnerFn: func(context.Context, uint64, uint64) (*model.Reservation, error) {
					if !expired {
						return &model.Reservation{ID: 5, Status: model.ReservationStatusTentative, IntendedDate: &day1}, nil
					}
					if later {
						return &model.Reservation{ID: 6, Status: model.ReservationStatusConfirmed, StartTime: future}, nil
					}
					return nil, nil
				}},
				tagCacheRepo: &mockLstepTagCacheRepository{findByOwnerFn: func(context.Context, uint64, uint64) ([]*model.LstepTagCache, error) {
					var out []*model.LstepTagCache
					for tag := range tags {
						out = append(out, &model.LstepTagCache{TagName: tag})
					}
					return out, nil
				}},
			}
			require.NoError(t, syncer.SyncNextVisitTag(context.Background(), 1, 2))
			require.True(t, tags["next_visit_2026-10-11"])
			expired = true
			repo := &expiryBatchRepository{findFn: func(_ context.Context, clinic uint64, asOf time.Time, after uint64, limit int) ([]uint64, error) {
				assert.Equal(t, uint64(1), clinic)
				assert.Equal(t, day2, asOf)
				if after == 0 {
					return []uint64{2}, nil
				}
				return nil, nil
			}}
			batch := &lstepBatchService{
				reservationRepo: repo,
				tagSyncSvc:      &batchMockTagSyncSvc{syncNextVisitTagFn: syncer.SyncNextVisitTag},
				clinicRepo:      &mockClinicRepository{findAllFn: func(context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil }},
				settingsSvc:     &mockLstepSettingsService{}, auditSvc: &scheduledAuditSpy{},
				lstepDeliveryTrigger: &mockLstepDeliveryTriggerBatch{},
			}
			result := batch.RunDeliveryTriggerBatchAllClinicsAt(context.Background(), day2, "expiry-regression")
			assert.Equal(t, BatchRunResult{Processed: 1, Succeeded: 1}, result)
			assert.False(t, tags["next_visit_2026-10-11"], "daily job must repair date-passage without an appointment mutation")
			if later {
				assert.True(t, tags["next_visit_2026-10-13"])
			} else {
				assert.Empty(t, tags)
			}
		})
	}
}

func TestExpiredTentativeRepair_PagingAndRetry(t *testing.T) {
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	fail := true
	var calls []uint64
	var cursors []uint64
	repo := &expiryBatchRepository{findFn: func(_ context.Context, clinic uint64, got time.Time, after uint64, limit int) ([]uint64, error) {
		assert.Equal(t, uint64(3), clinic)
		assert.Equal(t, asOf, got)
		assert.Equal(t, 200, limit)
		cursors = append(cursors, after)
		if after == 0 {
			return []uint64{2, 2, 5}, nil
		}
		if after == 5 {
			return []uint64{9}, nil
		}
		return nil, nil
	}}
	batch := &lstepBatchService{reservationRepo: repo, tagSyncSvc: &batchMockTagSyncSvc{syncNextVisitTagFn: func(_ context.Context, clinic, owner uint64) error {
		assert.Equal(t, uint64(3), clinic)
		calls = append(calls, owner)
		if owner == 5 && fail {
			return errors.New("mock removal failed")
		}
		return nil
	}}}
	count, errs := batch.reconcileExpiredTentativeTags(context.Background(), 3, asOf)
	assert.Equal(t, 2, count)
	assert.Len(t, errs, 1)
	assert.Equal(t, []uint64{2, 5, 9}, calls)
	assert.Equal(t, []uint64{0, 5, 9}, cursors)
	fail = false
	calls = nil
	cursors = nil
	count, errs = batch.reconcileExpiredTentativeTags(context.Background(), 3, asOf)
	assert.Equal(t, 3, count)
	assert.Empty(t, errs)
	assert.Contains(t, calls, uint64(5), "no persisted cursor may hide a previously failed owner")
}

func TestExpiredTentativeRepair_FailsClosed(t *testing.T) {
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name   string
		repo   lstepBatchReservationRepository
		syncer lstepBatchTagSyncer
		cancel bool
	}{
		{name: "missing repository", syncer: &batchMockTagSyncSvc{}},
		{name: "missing capability", repo: &batchMockReservationRepo{}, syncer: &batchMockTagSyncSvc{}},
		{name: "missing syncer", repo: &expiryBatchRepository{}},
		{name: "lookup error", repo: &expiryBatchRepository{findFn: func(context.Context, uint64, time.Time, uint64, int) ([]uint64, error) {
			return nil, errors.New("lookup")
		}}, syncer: &batchMockTagSyncSvc{}},
		{name: "nonadvancing cursor", repo: &expiryBatchRepository{findFn: func(context.Context, uint64, time.Time, uint64, int) ([]uint64, error) { return []uint64{0}, nil }}, syncer: &batchMockTagSyncSvc{}},
		{name: "cancelled", repo: &expiryBatchRepository{}, syncer: &batchMockTagSyncSvc{}, cancel: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel()
			}
			batch := &lstepBatchService{reservationRepo: tt.repo, tagSyncSvc: tt.syncer}
			count, errs := batch.reconcileExpiredTentativeTags(ctx, 1, asOf)
			assert.Zero(t, count)
			require.Len(t, errs, 1)
		})
	}
}

func TestDeliveryBatch_ExpiryClinicGateAndFailureCounters(t *testing.T) {
	asOf := time.Date(2026, 10, 12, 1, 0, 0, 0, time.UTC)
	var queried []uint64
	audit := &expiryRecordingAudit{}
	batch := &lstepBatchService{
		reservationRepo: &expiryBatchRepository{findFn: func(_ context.Context, clinic uint64, _ time.Time, after uint64, _ int) ([]uint64, error) {
			queried = append(queried, clinic)
			if after == 0 {
				return []uint64{2}, nil
			}
			return nil, nil
		}},
		tagSyncSvc: &batchMockTagSyncSvc{syncNextVisitTagFn: func(_ context.Context, clinic, owner uint64) error {
			assert.NotEqual(t, uint64(2), clinic)
			assert.Equal(t, uint64(2), owner)
			if clinic == 1 {
				return errors.New("remove failed")
			}
			return nil
		}},
		clinicRepo:  &mockClinicRepository{findAllFn: func(context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}, {ID: 2}, {ID: 3}}, nil }},
		settingsSvc: &mockLstepSettingsService{isSyncEnabledFn: func(_ context.Context, clinic uint64) (bool, error) { return clinic != 2, nil }},
		auditSvc:    audit, lstepDeliveryTrigger: &mockLstepDeliveryTriggerBatch{},
	}
	result := batch.RunDeliveryTriggerBatchAllClinicsAt(context.Background(), asOf, "retry-run")
	assert.Equal(t, BatchRunResult{Processed: 2, Succeeded: 1, Failed: 1}, result)
	require.NoError(t, result.Validate())
	assert.Equal(t, []uint64{1, 1, 3, 3}, queried)
	require.Len(t, audit.entries, 2)
	assert.Equal(t, 1, audit.entries[0]["error_count"])
	assert.Equal(t, "retry-run", audit.entries[0]["batch_run_id"])
	assert.Equal(t, asOf.UnixMilli(), audit.entries[0]["scheduled_time"])
}

type expiryRecordingAudit struct{ entries []map[string]any }

func (s *expiryRecordingAudit) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _ string, _ string, _ *uint64, metadata any) error {
	s.entries = append(s.entries, metadata.(map[string]any))
	return nil
}
