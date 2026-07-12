package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- local minimal mocks scoped to this file (avoids depending on concurrently-edited
//      lstep_batch_service_test.go / medical_record_service_test.go / clinic_service_test.go) ----

type dormantMockMedicalRecordRepository struct {
	findDormantOwnerEntriesFn       func(ctx context.Context, clinicID uint64, minDaysSince int) ([]repository.DormantOwnerEntry, error)
	findDormantOwnerEntriesCursorFn func(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]repository.DormantOwnerEntry, error)
}

func (m *dormantMockMedicalRecordRepository) FindAll(_ context.Context, _ []uint64, _ repository.MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
	return nil, 0, nil
}
func (m *dormantMockMedicalRecordRepository) FindByID(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) Create(_ context.Context, _ *model.MedicalRecord) error {
	return nil
}
func (m *dormantMockMedicalRecordRepository) Update(_ context.Context, _, _ uint64, _ map[string]any, _ *int) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *dormantMockMedicalRecordRepository) DeleteDraftByAppointmentID(_ context.Context, _, _ uint64) error {
	return nil
}

// LockByIDForUpdate は X-11 finalize-lock テスト用に FindByID と同じ挙動へ委譲する。
func (m *dormantMockMedicalRecordRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.FindByID(ctx, clinicID, id)
}
func (m *dormantMockMedicalRecordRepository) CountByPetID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *dormantMockMedicalRecordRepository) FindFirstVisitDateByPetID(_ context.Context, _, _ uint64) (*time.Time, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) CountEstimatesByMedicalRecordID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *dormantMockMedicalRecordRepository) FindOwnerVisitSummary(_ context.Context, _, _ uint64) (*repository.OwnerVisitSummary, error) {
	return &repository.OwnerVisitSummary{}, nil
}
func (m *dormantMockMedicalRecordRepository) FindLatestByOwner(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindDormantOwnerEntries(ctx context.Context, clinicID uint64, minDaysSince int) ([]repository.DormantOwnerEntry, error) {
	if m.findDormantOwnerEntriesFn != nil {
		return m.findDormantOwnerEntriesFn(ctx, clinicID, minDaysSince)
	}
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindDormantOwnerEntriesCursor(ctx context.Context, clinicID uint64, minDaysSince int, afterOwnerID uint64, limit int) ([]repository.DormantOwnerEntry, error) {
	if m.findDormantOwnerEntriesCursorFn != nil {
		return m.findDormantOwnerEntriesCursorFn(ctx, clinicID, minDaysSince, afterOwnerID, limit)
	}
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindOwnersByFirstVisitDate(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindOwnersByLastVisitDays(_ context.Context, _ uint64, _ int, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) FindOwnersByNextVisitRecommended(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
func (m *dormantMockMedicalRecordRepository) CountByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
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
func (m *dormantMockClinicRepository) FindByStaffID(_ context.Context, _ uint64) ([]model.Clinic, error) {
	return nil, nil
}
func (m *dormantMockClinicRepository) FindByID(_ context.Context, _ uint64) (*model.Clinic, error) {
	return nil, nil
}
func (m *dormantMockClinicRepository) FindCompany(_ context.Context) (*model.Company, error) {
	return nil, nil
}
func (m *dormantMockClinicRepository) Create(_ context.Context, _ *model.Clinic) error { return nil }
func (m *dormantMockClinicRepository) Update(_ context.Context, _ uint64, _ map[string]any) error {
	return nil
}
func (m *dormantMockClinicRepository) Delete(_ context.Context, _ uint64) error { return nil }
func (m *dormantMockClinicRepository) CountOwnersByClinicID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *dormantMockClinicRepository) CountStaffByClinicID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *dormantMockClinicRepository) CountBlockingReferencesByClinicID(_ context.Context, _ uint64) ([]repository.ClinicDependencyCount, error) {
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

func newDormantTagSyncWrapper(fn func(ctx context.Context, clinicID, ownerID uint64, daysSinceLastVisit int, thresholds model.DormantThresholds) error) LstepTagSyncService {
	return &dormantTagSyncServiceWrapper{mockLstepTagSyncService: &mockLstepTagSyncService{}, syncDormantTagsWithThresholdsFn: fn}
}

func newDormantBatchService(
	medRecordRepo repository.MedicalRecordRepository,
	tagSyncSvc LstepTagSyncService,
	clinicRepo repository.ClinicRepository,
	auditSvc AuditService,
	settingsSvc LstepSettingsService,
) LstepBatchService {
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return []repository.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return []repository.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
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
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return []repository.DormantOwnerEntry{{OwnerID: 10, DaysSince: 200}}, nil
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

	t.Run("nil settingsSvc processes all clinics without the enabled check (legacy path)", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		medRecordRepo := &dormantMockMedicalRecordRepository{
			findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
				return nil, nil
			},
		}
		svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), clinicRepo, &mockAuditService{}, nil)
		err := svc.RunDormantDetectionAllClinics(context.Background())
		assert.NoError(t, err)
	})
}

// ---- DetectDormantOwners error propagation (exercised indirectly above, direct case here) ----

func TestLstepBatchService_DetectDormantOwners_FindError(t *testing.T) {
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.DetectDormantOwners(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestLstepBatchService_DetectDormantOwners_ThresholdsError(t *testing.T) {
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, _ uint64, _ int) ([]repository.DormantOwnerEntry, error) {
			return []repository.DormantOwnerEntry{{OwnerID: 1, DaysSince: 200}}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getDormantThresholdsFn: func(_ context.Context, _ uint64) (model.DormantThresholds, error) {
			return model.DormantThresholds{}, errors.New("thresholds db error")
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, settingsSvc)
	count, errs := svc.DetectDormantOwners(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

// TestLstepBatchService_DetectDormantOwners_PaginatesAcrossMultiplePages verifies
// PERF-FOLLOWUP-02 cursor pagination: when the first page is exactly full (pageSize), a second
// page fetch is issued using the last entry's OwnerID as the cursor, and entries from both pages
// are processed with no duplicates/no skips.
func TestLstepBatchService_DetectDormantOwners_PaginatesAcrossMultiplePages(t *testing.T) {
	fetchCalls := make([]uint64, 0, 3)
	medRecordRepo := &dormantMockMedicalRecordRepository{
		findDormantOwnerEntriesCursorFn: func(_ context.Context, _ uint64, _ int, afterOwnerID uint64, limit int) ([]repository.DormantOwnerEntry, error) {
			fetchCalls = append(fetchCalls, afterOwnerID)
			assert.Equal(t, dormantBatchPageSize, limit)
			switch afterOwnerID {
			case 0:
				entries := make([]repository.DormantOwnerEntry, dormantBatchPageSize)
				for i := range entries {
					entries[i] = repository.DormantOwnerEntry{OwnerID: uint64(i + 1), DaysSince: 200}
				}
				return entries, nil
			case uint64(dormantBatchPageSize):
				return []repository.DormantOwnerEntry{{OwnerID: uint64(dormantBatchPageSize + 1), DaysSince: 200}}, nil
			default:
				return nil, nil
			}
		},
	}
	svc := newDormantBatchService(medRecordRepo, newDormantTagSyncWrapper(nil), &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.DetectDormantOwners(context.Background(), 1)
	assert.Empty(t, errs)
	assert.Equal(t, dormantBatchPageSize+1, count, "entries from both pages must be processed")
	assert.Equal(t, []uint64{0, uint64(dormantBatchPageSize)}, fetchCalls, "cursor must advance using the last entry's OwnerID of the previous page, no duplicates/no skips")
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
