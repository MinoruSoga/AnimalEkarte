package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- local minimal mocks scoped to this file (avoids depending on concurrently-edited
//      lstep_batch_service_test.go / medical_record_service_test.go / clinic_service_test.go) ----

type dormantMockMedicalRecordRepository struct {
	findDormantOwnerEntriesFn       func(ctx context.Context, clinicID uint64, minDaysSince int) ([]medicalrecord.DormantOwnerEntry, error)
	findDormantOwnerEntriesCursorFn func(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error)
}

func (m *dormantMockMedicalRecordRepository) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]medicalrecord.DormantOwnerEntry, error) {
	if m.findDormantOwnerEntriesFn != nil {
		return m.findDormantOwnerEntriesFn(ctx, clinicID, minDaysSince)
	}
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error) {
	if m.findDormantOwnerEntriesCursorFn != nil {
		return m.findDormantOwnerEntriesCursorFn(ctx, clinicID, minDaysSince, afterOwnerID, limit)
	}
	return nil, nil
}

type dormantMockClinicRepository struct {
	findAllFn func(ctx context.Context) ([]model.Clinic, error)
}

func (m *dormantMockClinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

// dormantTagSyncServiceWrapper embeds mockLstepTagSyncService (defined in the stable
// lstep_lifecycle_service_test.go) and overrides SyncDormantTagsWithThresholds, which the
// base mock hardcodes to always succeed with no injection hook.
type dormantTagSyncServiceWrapper struct {
	*mockLstepTagSyncService
	syncDormantTagsWithThresholdsFn func(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error
}

func (m *dormantTagSyncServiceWrapper) SyncDormantTagsWithThresholds(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error {
	if m.syncDormantTagsWithThresholdsFn != nil {
		return m.syncDormantTagsWithThresholdsFn(ctx, clinicID, ownerID, daysSinceLastVisit, thresholds)
	}
	return nil
}

func newDormantTagSyncWrapper(fn func(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error) lstepBatchTagSyncer {
	return &dormantTagSyncServiceWrapper{mockLstepTagSyncService: &mockLstepTagSyncService{}, syncDormantTagsWithThresholdsFn: fn}
}

// newDormantBatchService は具象型を返す（B-5: detectDormantOwners の unexport に伴い、
// テストが interface 外の非公開メソッドを直接呼ぶため）。
func newDormantBatchService(
	medRecordRepo lstepBatchMedicalRecordRepository,
	tagSyncSvc lstepBatchTagSyncer,
	clinicRepo lstepBatchClinicRepository,
	auditSvc lstepBatchAuditService,
	settingsSvc lstepBatchSettingsService,
) *lstepBatchService {
	return &lstepBatchService{
		medRecordRepo: medRecordRepo,
		tagSyncSvc:    tagSyncSvc,
		clinicRepo:    clinicRepo,
		auditSvc:      auditSvc,
		settingsSvc:   settingsSvc,
	}
}

// ---- RunDormantDetectionAllClinics ----

func TestLstepBatchService_RunDormantDetectionAllClinics(t *testing.T) {
	t.Run("returns wrapped error when fetching clinics fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return nil, errors.New("db error") },
		}
		svc := newDormantBatchService(&dormantMockMedicalRecordRepository{}, newDormantTagSyncWrapper(nil), clinicRepo, &mockAuditService{}, &mockLstepSettingsService{})
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.Error(t, err)
	})

	t.Run("skips clinics where sync-enabled check fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				t.Fatal("DetectDormantOwners must not run for a clinic whose sync-enabled check failed")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") },
		}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err, "per-clinic errors must not fail the whole batch")
	})

	t.Run("skips clinics with sync disabled", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				t.Fatal("DetectDormantOwners must not run for a clinic with sync disabled")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("processes enabled clinics and persists audit metadata when count>0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return []medicalrecord.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(action string) { auditCalled = true; assert.Equal(t, "batch_dormant_detect", action) }}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, auditSvc, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
		assert.True(t, auditCalled, "audit metadata must be persisted when count>0")
	})

	t.Run("does not persist audit metadata when count==0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(_ string) { auditCalled = true }}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, auditSvc, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
		assert.False(t, auditCalled)
	})

	t.Run("audit log failure is logged but does not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return []medicalrecord.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditSvc := &spyLstepBatchAuditService{err: errors.New("audit write failed")}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, auditSvc, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("tag sync errors are logged per-clinic but do not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return []medicalrecord.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		tagSync := newDormantTagSyncWrapper(func(_ context.Context, _, _ uint64, _ int, _ model.DormantThresholds) error {
			return errors.New("tag sync failed")
		})
		svc := newDormantBatchService(medRecordRepo, tagSync, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("nil settingsSvc fails closed", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
				return nil, nil
			},
		}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, &mockAuditService{}, nil)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.Error(t, err)
	})
}

// ---- DetectDormantOwners error propagation (exercised indirectly above, direct case here) ----

func TestLstepBatchService_DetectDormantOwners_FindError(t *testing.T) {
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectDormantOwners(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestLstepBatchService_DetectDormantOwners_ThresholdsError(t *testing.T) {
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			return []medicalrecord.DormantOwnerEntry{{OwnerID: 1, DaysSince: 200}}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getDormantThresholdsFn: func(_ context.Context, _ uint64) (model.DormantThresholds, error) {
			return model.DormantThresholds{}, errors.New("thresholds db error")
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, settingsSvc)
	count, errs := svc.detectDormantOwners(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestLstepBatchService_DetectDormantOwners_NilSettingsFailsClosed(t *testing.T) {
	tagSyncCalled := false
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]medicalrecord.DormantOwnerEntry, error) {
			return []medicalrecord.DormantOwnerEntry{{OwnerID: 1, DaysSince: 200}}, nil
		},
	}
	tagSync := newDormantTagSyncWrapper(func(_ context.Context, _, _ uint64, _ int, _ model.DormantThresholds) error {
		tagSyncCalled = true
		return nil
	})
	svc := newDormantBatchService(medRecordRepo, tagSync, &dormantMockClinicRepository{}, &mockAuditService{}, nil)

	count, errs := svc.detectDormantOwners(context.Background(), 1)

	assert.Zero(t, count)
	assert.NotEmpty(t, errs)
	assert.False(t, tagSyncCalled)
}

// TestLstepBatchService_DetectDormantOwners_PaginatesAcrossMultiplePages verifies
// PERF-FOLLOWUP-02 cursor pagination: when the first page is exactly full (pageSize), a second
// page fetch is issued using the last entry's OwnerID as the cursor, and entries from both pages
// are processed with no duplicates/no skips.
func TestLstepBatchService_DetectDormantOwners_PaginatesAcrossMultiplePages(t *testing.T) {
	fetchCalls := make([]uint64, 0, 3)
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, afterOwnerID uint64, limit int) ([]medicalrecord.DormantOwnerEntry, error) {
			fetchCalls = append(fetchCalls, afterOwnerID)
			assert.Equal(t, lstepDormantBatchPageSize, limit)
			switch afterOwnerID {
			case 0:
				entries := make([]medicalrecord.DormantOwnerEntry, lstepDormantBatchPageSize)
				for i := range entries {
					entries[i] = medicalrecord.DormantOwnerEntry{OwnerID: uint64(i + 1), DaysSince: 200}
				}
				return entries, nil
			case uint64(lstepDormantBatchPageSize):
				return []medicalrecord.DormantOwnerEntry{{OwnerID: uint64(lstepDormantBatchPageSize + 1), DaysSince: 200}}, nil
			default:
				return nil, nil
			}
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectDormantOwners(context.Background(), 1)
	assert.Empty(t, errs)
	assert.Equal(t, lstepDormantBatchPageSize+1, count, "entries from both pages must be processed")
	assert.Equal(t, []uint64{0, uint64(lstepDormantBatchPageSize)}, fetchCalls, "cursor must advance using the last entry's OwnerID of the previous page, no duplicates/no skips")
}

// spyLstepBatchAuditService is a minimal AuditService spy scoped to this file (avoids relying
// on a specific concurrently-edited spy from another domain's test file).
type spyLstepBatchAuditService struct {
	mockAuditService
	err               error
	onLogWithMetadata func(action string)
}

func (s *spyLstepBatchAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, action, _ string, _ *uint64, _ any) error {
	if s.onLogWithMetadata != nil {
		s.onLogWithMetadata(action)
	}
	return s.err
}
