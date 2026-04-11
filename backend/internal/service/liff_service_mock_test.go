package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ================================================================
// モック実装: liff_service_test 専用
// ================================================================

// --- mockLineReservationSettingRepository ---

type mockLiffSettingRepository struct {
	findByClinicIDFn func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
}

func (m *mockLiffSettingRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.findByClinicIDFn != nil {
		return m.findByClinicIDFn(ctx, clinicID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockLiffSettingRepository) Upsert(_ context.Context, _ *model.LineReservationSetting) error {
	return nil
}

// --- mockLiffTypeRepository ---

type mockLiffTypeRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

func (m *mockLiffTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockLiffTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffTypeRepository) Create(_ context.Context, _ *model.ReservationType) error { return nil }

func (m *mockLiffTypeRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}

func (m *mockLiffTypeRepository) Delete(_ context.Context, _, _ uint64) error { return nil }

func (m *mockLiffTypeRepository) SwapSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

// --- mockLiffStaffRepository ---

type mockLiffStaffRepository struct {
	findAllByClinicIDFn        func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                 func(ctx context.Context, id uint64) (*model.Staff, error)
	findExcludedReservationTypesFn func(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error)
}

func (m *mockLiffStaffRepository) FindAllByClinicID(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	if m.findAllByClinicIDFn != nil {
		return m.findAllByClinicIDFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffStaffRepository) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}

func (m *mockLiffStaffRepository) Update(_ context.Context, _ uint64, _ map[string]any) error {
	return nil
}

func (m *mockLiffStaffRepository) SoftDelete(_ context.Context, _ uint64) error { return nil }

func (m *mockLiffStaffRepository) SwapSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

func (m *mockLiffStaffRepository) FindExcludedReservationTypes(ctx context.Context, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, staffID)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) FindExcludedReservationTypesByStaffIDs(_ context.Context, _ []uint64) ([]model.StaffReservationExclusion, error) {
	return nil, nil
}

func (m *mockLiffStaffRepository) ReplaceExcludedReservationTypes(_ context.Context, _ uint64, _ []uint64) error {
	return nil
}

// --- mockLiffScheduleRepository ---

type mockLiffScheduleRepository struct {
	findByDateFn func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
}

func (m *mockLiffScheduleRepository) FindByMonth(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
	return nil, nil
}

func (m *mockLiffScheduleRepository) FindBreaksByEntryIDs(_ context.Context, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	return nil, nil
}

func (m *mockLiffScheduleRepository) FindByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, staffID, date)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffScheduleRepository) FindBreaksByEntryID(_ context.Context, _ uint64) ([]model.ShiftEntryBreak, error) {
	return nil, nil
}

func (m *mockLiffScheduleRepository) Upsert(_ context.Context, _ *model.ShiftEntry, _ []model.ShiftEntryBreak) error {
	return nil
}

func (m *mockLiffScheduleRepository) DeleteByDate(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

// --- mockLiffAdminRepository ---

type mockLiffAdminRepository struct {
	findByDayFn         func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Appointment, error)
	findByCustomerIDFn  func(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error)
	cancelByIDFn        func(ctx context.Context, clinicID, customerID, id uint64) error
	findByIDForNotifyFn func(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
}

func (m *mockLiffAdminRepository) FindByMonth(_ context.Context, _ uint64, _ int, _ time.Month) ([]model.Appointment, error) {
	return nil, nil
}

func (m *mockLiffAdminRepository) FindByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Appointment, error) {
	if m.findByDayFn != nil {
		return m.findByDayFn(ctx, clinicID, date)
	}
	return nil, nil
}

func (m *mockLiffAdminRepository) Create(_ context.Context, _ *model.Appointment) error {
	return nil
}

func (m *mockLiffAdminRepository) SoftDelete(_ context.Context, _, _ uint64) error { return nil }

func (m *mockLiffAdminRepository) FindByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Appointment, error) {
	if m.findByCustomerIDFn != nil {
		return m.findByCustomerIDFn(ctx, clinicID, customerID)
	}
	return nil, nil
}

func (m *mockLiffAdminRepository) CancelByID(ctx context.Context, clinicID, customerID, id uint64) error {
	if m.cancelByIDFn != nil {
		return m.cancelByIDFn(ctx, clinicID, customerID, id)
	}
	return nil
}

func (m *mockLiffAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	if m.findByIDForNotifyFn != nil {
		return m.findByIDForNotifyFn(ctx, clinicID, id)
	}
	return nil, nil
}

// --- mockLiffCustomerRepository ---

type mockLiffCustomerRepository struct {
	findByIDFn               func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	updateAdditionalFieldsFn func(ctx context.Context, clinicID, id uint64, fields []byte) error
}

func (m *mockLiffCustomerRepository) FindAll(_ context.Context, _ uint64) ([]model.LineCustomer, error) {
	return nil, nil
}

func (m *mockLiffCustomerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffCustomerRepository) UpdateOwnerLink(_ context.Context, _, _ uint64, _ *uint64) error {
	return nil
}

func (m *mockLiffCustomerRepository) FindOrCreateByLineUserID(_ context.Context, _ uint64, _, _ string) (*model.LineCustomer, error) {
	return nil, nil
}

func (m *mockLiffCustomerRepository) UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error {
	if m.updateAdditionalFieldsFn != nil {
		return m.updateAdditionalFieldsFn(ctx, clinicID, id, fields)
	}
	return nil
}

// --- mockLiffValidators ---

type mockLiffValidators struct {
	validateAndCreateFn func(ctx context.Context, input *CreateReservationInput) (*model.Appointment, error)
}

func (m *mockLiffValidators) ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Appointment, error) {
	if m.validateAndCreateFn != nil {
		return m.validateAndCreateFn(ctx, input)
	}
	return &model.Appointment{ID: 1, ClinicID: input.ClinicID}, nil
}

// --- mockLiffNotifier ---

type mockLiffNotifier struct {
	notifyCreatedFn   func(ctx context.Context, appt *model.Appointment, customer *model.LineCustomer)
	notifyCancelledFn func(ctx context.Context, appt *model.Appointment, customer *model.LineCustomer)
}

func (m *mockLiffNotifier) NotifyCreated(ctx context.Context, appt *model.Appointment, customer *model.LineCustomer) {
	if m.notifyCreatedFn != nil {
		m.notifyCreatedFn(ctx, appt, customer)
	}
}

func (m *mockLiffNotifier) NotifyCancelled(ctx context.Context, appt *model.Appointment, customer *model.LineCustomer) {
	if m.notifyCancelledFn != nil {
		m.notifyCancelledFn(ctx, appt, customer)
	}
}

// ================================================================
// テスト共通ヘルパー
// ================================================================

// newLiffSvc はテスト用の liffService を直接生成する。
func newLiffSvc(
	setting *mockLiffSettingRepository,
	course *mockLiffTypeRepository,
	staff *mockLiffStaffRepository,
	schedule *mockLiffScheduleRepository,
	admin *mockLiffAdminRepository,
	customer *mockLiffCustomerRepository,
	validators *mockLiffValidators,
	notifier *mockLiffNotifier,
) *liffService {
	return &liffService{
		settingRepo:  setting,
		typeLiffRepo:   course,
		staffRepo:    staff,
		scheduleRepo: schedule,
		adminRepo:    admin,
		customerRepo: customer,
		validators:   validators,
		notifier:     notifier,
	}
}

// liffDefaultSetting はテスト用デフォルト予約設定を返す。
func liffDefaultSetting() *model.LineReservationSetting {
	return &model.LineReservationSetting{
		ID:                      1,
		ClinicID:                3,
		Status:                  "active",
		LiffID:                  "2009755544-abcdefgh",
		BusinessHours:           json.RawMessage(`{"start":"0900","end":"1900"}`),
		BreakHours:              json.RawMessage(`[{"start":"1200","end":"1300"}]`),
		BookingWindowMinDays:    2,
		BookingWindowMaxDays:    30,
		CalendarMonths:          2,
		TimeSlotIntervalMinutes: 15,
		TimeSlotMode:            "minimize_gaps",
		NoStaffMode:             "first_available",
	}
}
