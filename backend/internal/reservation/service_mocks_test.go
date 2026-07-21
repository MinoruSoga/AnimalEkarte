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

// okTrimmingCourseRepo / okTrimmingOptionRepo — service/cross_tenant_master_fk_write_test.go の
// 同名builderの最小複製（view型版・def残存→再宣言規約）。
type mockTrimmingCourseFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
}

func (m *mockTrimmingCourseFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseFinder) FindAll(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
	return nil, nil
}

type mockTrimmingOptionFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
}

func (m *mockTrimmingOptionFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionFinder) FindAll(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
	return nil, nil
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
