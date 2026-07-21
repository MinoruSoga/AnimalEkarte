package reservation

// service_mocks_test.go — mocks_shared_test.go（service残存・liff/occupation等の残留consumer共有）
// からの複製（BE9-2C R①: def残存→移動先で再宣言する規約。liff系(R⑤)移動時に集約解消）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockOccupationRepository は repository.OccupationRepository のテスト用共有モック（BE7-15 正本）。
type mockOccupationRepository struct {
	findAllFn                  func(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	findByIDFn                 func(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	createFn                   func(ctx context.Context, occupation *model.Occupation) error
	updateFieldsFn             func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
	reorderErr                 error
	countUsageByOccupationIDFn func(ctx context.Context, clinicID, occupationID uint64) (int64, error)
}

func (m *mockOccupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.Occupation{}, nil
}

func (m *mockOccupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Occupation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOccupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, occupation)
	}
	return nil
}

func (m *mockOccupationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Occupation{}, nil
}

func (m *mockOccupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockOccupationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockOccupationRepository) CountUsageByOccupationID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByOccupationIDFn != nil {
		return m.countUsageByOccupationIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

// mockReservationTypeOccupationRepository は repository.ReservationTypeOccupationRepository のテスト用共有モック（BE7-15 正本）。
type mockReservationTypeOccupationRepository struct {
	findAllFn         func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	findByIDFn        func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	createFn          func(ctx context.Context, o *model.ReservationTypeOccupation) error
	deleteFn          func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
	countByStaffIDsFn func(ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time) (map[string]int64, error)
}

func (m *mockReservationTypeOccupationRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeOccupation{}, nil
}

func (m *mockReservationTypeOccupationRepository) FindByID(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return &model.ReservationTypeOccupation{ID: 1, ClinicID: clinicID, ReservationTypeID: reservationTypeID, OccupationID: occupationID}, nil
}

func (m *mockReservationTypeOccupationRepository) Create(ctx context.Context, o *model.ReservationTypeOccupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, o)
	}
	return nil
}

func (m *mockReservationTypeOccupationRepository) Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return nil
}

func (m *mockReservationTypeOccupationRepository) CountWorkingStaffByReservationTypeIDs(ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time) (map[string]int64, error) {
	if m.countByStaffIDsFn != nil {
		return m.countByStaffIDsFn(ctx, clinicID, reservationTypeID, dates)
	}
	result := make(map[string]int64, len(dates))
	for _, d := range dates {
		result[d.Format("2006-01-02")] = 1
	}
	return result, nil
}

// mockTransactor は service/trimming_service_test.go の同名モックの複製（def残存→再宣言規約）。
type mockTransactor struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}

// mockReservationRepository — appointment_service_test.go（R④残留）定義の複製（def残存→再宣言規約）。
type mockReservationRepository struct {
	findAllFn                          func(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error)
	findByIDFn                         func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	lockAndFindByIDFn                  func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	createFn                           func(ctx context.Context, reservation *model.Reservation) error
	updateFieldsFn                     func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
	deleteFn                           func(ctx context.Context, clinicID, id uint64) error
	countMedicalRecordsByReservationID func(ctx context.Context, reservationID uint64) (int64, error)
	countOnDutyDoctorsFn               func(ctx context.Context, clinicID uint64, date time.Time) (int64, error)
	countConflictsFn                   func(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error)
	countByTypeAndStartTimeFn          func(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
	assertOwnerInClinicFn              func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn             func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	assertLineCustomerInClinicFn       func(ctx context.Context, clinicID, lineCustomerID uint64) error
}

func (m *mockReservationRepository) FindAll(ctx context.Context, clinicIDs []uint64, page, limit int, date, startDate, endDate *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Reservation, int64, error) {
	return m.findAllFn(ctx, clinicIDs, page, limit, date, startDate, endDate, status, source, petID, ownerID)
}

func (m *mockReservationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
	return nil, nil
}

func (m *mockReservationRepository) Create(ctx context.Context, reservation *model.Reservation) error {
	return m.createFn(ctx, reservation)
}

func (m *mockReservationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockReservationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountMedicalRecordsByReservationID(ctx context.Context, _, reservationID uint64) (int64, error) {
	if m.countMedicalRecordsByReservationID != nil {
		return m.countMedicalRecordsByReservationID(ctx, reservationID)
	}
	return 0, nil
}

func (m *mockReservationRepository) AcquireBookingLock(_ context.Context, _ uint64) error {
	return nil
}

func (m *mockReservationRepository) LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.lockAndFindByIDFn != nil {
		return m.lockAndFindByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockReservationRepository) CountOnDutyDoctors(ctx context.Context, clinicID uint64, date time.Time) (int64, error) {
	if m.countOnDutyDoctorsFn != nil {
		return m.countOnDutyDoctorsFn(ctx, clinicID, date)
	}
	return 1, nil
}

func (m *mockReservationRepository) CountConflicts(ctx context.Context, clinicID uint64, start, end time.Time, excludeID *uint64) (int64, error) {
	if m.countConflictsFn != nil {
		return m.countConflictsFn(ctx, clinicID, start, end, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	if m.countByTypeAndStartTimeFn != nil {
		return m.countByTypeAndStartTimeFn(ctx, clinicID, reservationTypeID, startTime, excludeID)
	}
	return 0, nil
}

func (m *mockReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockReservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerInClinicFn != nil {
		return m.assertOwnerInClinicFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockReservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerInClinicFn != nil {
		return m.findPetOwnerInClinicFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockReservationRepository) AssertLineCustomerInClinic(ctx context.Context, clinicID, lineCustomerID uint64) error {
	if m.assertLineCustomerInClinicFn != nil {
		return m.assertLineCustomerInClinicFn(ctx, clinicID, lineCustomerID)
	}
	return nil
}

func (m *mockReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

// okTrimmingCourseRepo / okTrimmingOptionRepo — service/cross_tenant_master_fk_write_test.go の
// 同名builderの最小複製（view型版・def残存→再宣言規約）。
type mockTrimmingCourseFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
}

func (m *mockTrimmingCourseFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

type mockTrimmingOptionFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
}

func (m *mockTrimmingOptionFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func okTrimmingCourseRepo() trimmingCourseFinder {
	return &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func okTrimmingOptionRepo() trimmingOptionFinder {
	return &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}

// rejectTrimmingCourseRepo — service側同名builderのview型版複製。
func rejectTrimmingCourseRepo(ownedID uint64) trimmingCourseFinder {
	return &mockTrimmingCourseFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_course", "foreign")
		}
		return &model.TrimmingCourse{ID: id, IsActive: true}, nil
	}}
}

func rejectTrimmingOptionRepo(ownedID uint64) trimmingOptionFinder {
	return &mockTrimmingOptionFinder{findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingOption, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("trimming_option", "foreign")
		}
		return &model.TrimmingOption{ID: id, IsActive: true}, nil
	}}
}
