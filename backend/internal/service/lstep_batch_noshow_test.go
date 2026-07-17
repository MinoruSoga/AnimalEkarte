package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// noShowMockReservationRepository is a local, complete ReservationRepository mock scoped to
// this file (avoids depending on concurrently-edited appointment_service_test.go /
// lstep_batch_service_test.go, which define similarly-shaped mocks under different names).
type noShowMockReservationRepository struct {
	findNoShowCandidatesFn func(ctx context.Context, clinicID uint64) ([]model.Reservation, error)
	updateFn               func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
}

func (m *noShowMockReservationRepository) FindAll(_ context.Context, _ []uint64, _, _ int, _, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *noShowMockReservationRepository) FindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *noShowMockReservationRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *noShowMockReservationRepository) Create(_ context.Context, _ *model.Reservation) error {
	return nil
}
func (m *noShowMockReservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{ID: id}, nil
}
func (m *noShowMockReservationRepository) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *noShowMockReservationRepository) AcquireBookingLock(_ context.Context, _ uint64) error {
	return nil
}
func (m *noShowMockReservationRepository) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, nil
}
func (m *noShowMockReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}
func (m *noShowMockReservationRepository) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) CountByTypeAndStartTime(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *noShowMockReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}
func (m *noShowMockReservationRepository) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}
func (m *noShowMockReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}
func (m *noShowMockReservationRepository) FindNoShowCandidates(ctx context.Context, clinicID uint64) ([]model.Reservation, error) {
	if m.findNoShowCandidatesFn != nil {
		return m.findNoShowCandidatesFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *noShowMockReservationRepository) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *noShowMockReservationRepository) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *noShowMockReservationRepository) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

// newNoShowBatchService は具象型を返す（B-5: detectNoShowReservations の unexport に伴い、
// テストが interface 外の非公開メソッドを直接呼ぶため）。
func newNoShowBatchService(reservationRepo *noShowMockReservationRepository, clinicRepo *dormantMockClinicRepository, auditSvc AuditService, settingsSvc LstepSettingsService) *lstepBatchService {
	return &lstepBatchService{
		reservationRepo: reservationRepo,
		clinicRepo:      clinicRepo,
		auditSvc:        auditSvc,
		settingsSvc:     settingsSvc,
	}
}

// ---- RunNoShowCheckAllClinics ----

func TestLstepBatchService_RunNoShowCheckAllClinics(t *testing.T) {
	t.Run("returns wrapped error when fetching clinics fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return nil, errors.New("db error") },
		}
		svc := newNoShowBatchService(&noShowMockReservationRepository{}, clinicRepo, &mockAuditService{}, &mockLstepSettingsService{})
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.Error(t, err)
	})

	t.Run("skips clinics where sync-enabled check fails", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				t.Fatal("DetectNoShowReservations must not run when sync-enabled check failed")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, errors.New("db error") },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("skips clinics with sync disabled", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				t.Fatal("DetectNoShowReservations must not run for a clinic with sync disabled")
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("processes enabled clinics and persists audit metadata when count>0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(action string) { auditCalled = true; assert.Equal(t, "batch_no_show_detect", action) }}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.True(t, auditCalled)
	})

	t.Run("does not persist audit metadata when count==0", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return nil, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditCalled := false
		auditSvc := &spyLstepBatchAuditService{onLogWithMetadata: func(_ string) { auditCalled = true }}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
		assert.False(t, auditCalled)
	})

	t.Run("audit log failure is logged but does not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		auditSvc := &spyLstepBatchAuditService{err: errors.New("audit write failed")}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, auditSvc, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("per-reservation update errors are logged but do not fail the batch", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return []model.Reservation{{ID: 5}}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Reservation, error) {
				return nil, errors.New("update failed")
			},
		}
		settingsSvc := &mockLstepSettingsService{
			isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return true, nil },
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, settingsSvc)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})

	t.Run("nil settingsSvc processes all clinics without the enabled check (legacy path)", func(t *testing.T) {
		clinicRepo := &dormantMockClinicRepository{
			findAllFn: func(_ context.Context) ([]model.Clinic, error) { return []model.Clinic{{ID: 1}}, nil },
		}
		reservationRepo := &noShowMockReservationRepository{
			findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
				return nil, nil
			},
		}
		svc := newNoShowBatchService(reservationRepo, clinicRepo, &mockAuditService{}, nil)
		err := svc.RunNoShowCheckAllClinics(context.Background())
		assert.NoError(t, err)
	})
}

// ---- DetectNoShowReservations direct cases ----

func TestLstepBatchService_DetectNoShowReservations_FindError(t *testing.T) {
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectNoShowReservations(context.Background(), 1)
	assert.Equal(t, 0, count)
	assert.NotEmpty(t, errs)
}

func TestLstepBatchService_DetectNoShowReservations_Success(t *testing.T) {
	var capturedFields map[string]any
	reservationRepo := &noShowMockReservationRepository{
		findNoShowCandidatesFn: func(_ context.Context, _ uint64) ([]model.Reservation, error) {
			return []model.Reservation{{ID: 5}, {ID: 6}}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Reservation, error) {
			capturedFields = fields
			return &model.Reservation{ID: 5, Status: model.ReservationStatusNoShow}, nil
		},
	}
	svc := newNoShowBatchService(reservationRepo, &dormantMockClinicRepository{}, &mockAuditService{}, &mockLstepSettingsService{})
	count, errs := svc.detectNoShowReservations(context.Background(), 1)
	assert.Equal(t, 2, count)
	assert.Empty(t, errs)
	assert.Equal(t, model.ReservationStatusNoShow, capturedFields["status"])
}
