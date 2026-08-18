package reservation

// liff_mock_carriers_test.go — BE9-2C R⑤: 移動した liff 系 test が使う、def 残存
// （trimming/line/vaccination=未移行domain）の共有 mock の再宣言複製（def残存→再宣言規約）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockTrimmingCourseRepository struct {
	findAllFn              func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	createFn               func(ctx context.Context, course *model.TrimmingCourse) error
	updateFieldsFn         func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error)
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	countUsageByCourseIDFn func(ctx context.Context, clinicID, courseID uint64) (int64, error)
	reorderErr             error
}

func (m *mockTrimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return m.createFn(ctx, course)
}

func (m *mockTrimmingCourseRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingCourse, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockTrimmingCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockTrimmingCourseRepository) CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error) {
	if m.countUsageByCourseIDFn != nil {
		return m.countUsageByCourseIDFn(ctx, clinicID, courseID)
	}
	return 0, nil
}

type mockTrimmingDetailRepository struct {
	findByAppointmentIDFn func(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
	createFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	updateFn              func(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
	setOptionsFn          func(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error
}

func (m *mockTrimmingDetailRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
	if m.findByAppointmentIDFn != nil {
		return m.findByAppointmentIDFn(ctx, clinicID, appointmentID)
	}
	return &model.AppointmentTrimmingDetail{AppointmentID: appointmentID}, nil
}

func (m *mockTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if m.createFn != nil {
		return m.createFn(ctx, detail)
	}
	return nil
}

func (m *mockTrimmingDetailRepository) Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, detail)
	}
	return nil
}

func (m *mockTrimmingDetailRepository) SetOptions(ctx context.Context, clinicID, appointmentID uint64, optionIDs []uint64) error {
	if m.setOptionsFn != nil {
		return m.setOptionsFn(ctx, clinicID, appointmentID, optionIDs)
	}
	return nil
}

type mockTrimmingOptionRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	createFn            func(ctx context.Context, option *model.TrimmingOption) error
	updateFieldsFn      func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	reorderErr          error
	countRecordsByOptFn func(ctx context.Context, clinicID, optionID uint64) (int64, error)
}

func (m *mockTrimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	return m.createFn(ctx, option)
}

func (m *mockTrimmingOptionRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.TrimmingOption, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockTrimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockTrimmingOptionRepository) CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error) {
	if m.countRecordsByOptFn != nil {
		return m.countRecordsByOptFn(ctx, clinicID, optionID)
	}
	return 0, nil
}

type mockVaccinationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	findByOwnerFn  func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	createFn       func(ctx context.Context, vaccination *model.Vaccination) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, search string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return m.createFn(ctx, vaccination)
}

func (m *mockVaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockVaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockVaccinationRepository) FindOwnersByVaccineDeadline(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}
