package service

// be9_2c_r1_mock_carriers_test.go — BE9-2C R①でreservationへ移動したtestが定義していた共有mockの
// carrier複製（残留consumer: appointment/liff/cross_tenant系test。liff(R⑤)/appointment(R④)移動時に解消）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockUnavailableTimeRepository は ReservationTypeUnavailableTimeRepository のテスト用モック
type mockUnavailableTimeRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error)
	createFn   func(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockUnavailableTimeRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeUnavailableTime{}, nil
}

func (m *mockUnavailableTimeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeUnavailableTime{ID: id}, nil
}

func (m *mockUnavailableTimeRepository) Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}

func (m *mockUnavailableTimeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// mockAvailableSlotRepository は ReservationTypeAvailableSlotRepository のテスト用モック実装
type mockAvailableSlotRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	createFn   func(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockAvailableSlotRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeAvailableSlot{}, nil
}

func (m *mockAvailableSlotRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeAvailableSlot{ID: id, ClinicID: clinicID}, nil
}

func (m *mockAvailableSlotRepository) Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error {
	if m.createFn != nil {
		return m.createFn(ctx, slot)
	}
	return nil
}

func (m *mockAvailableSlotRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// mockReservationStaffRepository — R②移動test（reservation_staff_service_test.go）由来のcarrier複製。
type mockReservationStaffRepository struct {
	findAllFn                              func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                             func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	createFn                               func(ctx context.Context, staff *model.Staff, clinicID uint64) error
	updateFieldsFn                         func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                               func(ctx context.Context, clinicID, id uint64) error
	countUsageByStaffIDFn                  func() (int64, error)
	swapSortOrderFn                        func(ctx context.Context, clinicID, id uint64, direction string) error
	findExcludedReservationTypesFn         func(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
	findExcludedReservationTypesByStaffIDs func(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	replaceExcludedReservationTypesFn      func(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error
	findCapabilitiesFn                     func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error)
	findCapabilitiesByStaffIDsFn           func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
	replaceCapabilitiesFn                  func(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error
	supportsReservationTypeFn              func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockReservationStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockReservationStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) Create(ctx context.Context, staff *model.Staff, clinicID uint64) error {
	return m.createFn(ctx, staff, clinicID)
}

func (m *mockReservationStaffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil
}

func (m *mockReservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockReservationStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	if m.countUsageByStaffIDFn != nil {
		return m.countUsageByStaffIDFn()
	}
	return 0, nil
}

func (m *mockReservationStaffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if m.swapSortOrderFn != nil {
		return m.swapSortOrderFn(ctx, clinicID, id, direction)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, staffID)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesByStaffIDs != nil {
		return m.findExcludedReservationTypesByStaffIDs(ctx, staffIDs)
	}
	return []model.StaffReservationExclusion{}, nil
}

func (m *mockReservationStaffRepository) UpdateExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64, courseIDs []uint64) error {
	if m.replaceExcludedReservationTypesFn != nil {
		return m.replaceExcludedReservationTypesFn(ctx, clinicID, staffID, courseIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilities(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesFn != nil {
		return m.findCapabilitiesFn(ctx, clinicID, staffID)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesByStaffIDsFn != nil {
		return m.findCapabilitiesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return []model.StaffReservationCapability{}, nil
}

func (m *mockReservationStaffRepository) UpdateReservationCapabilities(ctx context.Context, clinicID, staffID uint64, typeIDs []uint64) error {
	if m.replaceCapabilitiesFn != nil {
		return m.replaceCapabilitiesFn(ctx, clinicID, staffID, typeIDs)
	}
	return nil
}

func (m *mockReservationStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}

// mockReservationTypeFinder — R③移動test（reservation_capacity_test.go）由来のcarrier複製。
type mockReservationTypeFinder struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

// dateInDays/newSettingForValidation — R③移動test（reservation_validators_test.go）由来のcarrier複製（liff系cross_tenant用）。
// newSettingForValidation は業務ルール検証テスト用の設定を返す。
// デフォルト: 営業時間 09:00-19:00 / 休憩 12:00-13:00 / 予約窓 2〜30日
func newSettingForValidation() *model.LineReservationSetting {
	return &model.LineReservationSetting{
		Status:                "running",
		ClosedWeekdays:        []byte("[]"),
		ClosedDates:           []byte("[]"),
		NationalHolidayClosed: false,
		BusinessHours:         []byte(`{"start":"0900","end":"1900"}`),
		BreakHours:            []byte(`[{"start":"1200","end":"1300"}]`),
		BookingWindowMinDays:  2,
		BookingWindowMaxDays:  30,
	}
}

// dateInDays は今日から n 日後の 00:00 JST を返す。
func dateInDays(n int) time.Time {
	now := time.Now().In(config.JST)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, config.JST)
	return today.AddDate(0, 0, n)
}

func ptrU64(v uint64) *uint64 { return &v }

func (m mockReservationTypeFinder) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return m.findByIDFn(ctx, clinicID, id)
}
