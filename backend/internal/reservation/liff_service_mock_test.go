package reservation

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
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

func (m *mockLiffSettingRepository) FindAll(_ context.Context) ([]model.LineReservationSetting, error) {
	return nil, nil
}

func (m *mockLiffSettingRepository) Save(_ context.Context, _ uint64, _ *model.LineReservationSetting) error {
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

func (m *mockLiffTypeRepository) CountChildrenByParentID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockLiffTypeRepository) Create(_ context.Context, _ *model.ReservationType) error {
	return nil
}

func (m *mockLiffTypeRepository) Update(_ context.Context, clinicID, id uint64, _ map[string]any) (*model.ReservationType, error) {
	return &model.ReservationType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockLiffTypeRepository) Delete(_ context.Context, _, _ uint64) error { return nil }

func (m *mockLiffTypeRepository) DeleteWithDependencyChecks(_ context.Context, _, _ uint64, _ reservationTypeUsageChecker) error {
	return nil
}

func (m *mockLiffTypeRepository) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

// --- mockLiffStaffRepository ---

type mockLiffStaffRepository struct {
	findAllFn                                func(ctx context.Context, clinicID uint64) ([]model.Staff, error)
	findByIDFn                               func(ctx context.Context, clinicID, id uint64) (*model.Staff, error)
	findExcludedReservationTypesFn           func(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error)
	findExcludedReservationTypesByStaffIDsFn func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error)
	findCapabilitiesByStaffIDsFn             func(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error)
	supportsReservationTypeFn                func(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error)
}

func (m *mockLiffStaffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Staff, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Staff, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffStaffRepository) LockForMutation(
	_ context.Context,
	clinicID, id uint64,
) (*model.Staff, error) {
	return &model.Staff{ID: id, ClinicID: clinicID}, nil
}

func (m *mockLiffStaffRepository) Create(_ context.Context, _ *model.Staff, _ uint64) error {
	return nil
}

func (m *mockLiffStaffRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}

func (m *mockLiffStaffRepository) Delete(_ context.Context, _, _ uint64) error { return nil }
func (m *mockLiffStaffRepository) CountUsageByStaffID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockLiffStaffRepository) UpdateSortOrder(_ context.Context, _, _ uint64, _ string) error {
	return nil
}

func (m *mockLiffStaffRepository) FindAllExcludedReservationTypes(ctx context.Context, clinicID, staffID uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesFn != nil {
		return m.findExcludedReservationTypesFn(ctx, clinicID, staffID)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) FindAllExcludedReservationTypesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationExclusion, error) {
	if m.findExcludedReservationTypesByStaffIDsFn != nil {
		return m.findExcludedReservationTypesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) UpdateExcludedReservationTypes(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}

func (m *mockLiffStaffRepository) FindAllReservationCapabilities(_ context.Context, _, _ uint64) ([]model.StaffReservationCapability, error) {
	return nil, nil
}

func (m *mockLiffStaffRepository) FindAllReservationCapabilitiesByStaffIDs(ctx context.Context, clinicID uint64, staffIDs []uint64) ([]model.StaffReservationCapability, error) {
	if m.findCapabilitiesByStaffIDsFn != nil {
		return m.findCapabilitiesByStaffIDsFn(ctx, clinicID, staffIDs)
	}
	return nil, nil
}

func (m *mockLiffStaffRepository) UpdateReservationCapabilities(_ context.Context, _, _ uint64, _ []uint64) error {
	return nil
}

func (m *mockLiffStaffRepository) SupportsReservationType(ctx context.Context, clinicID, staffID, reservationTypeID uint64) (bool, error) {
	if m.supportsReservationTypeFn != nil {
		return m.supportsReservationTypeFn(ctx, clinicID, staffID, reservationTypeID)
	}
	return true, nil
}

// --- mockLiffScheduleRepository ---

type mockLiffScheduleRepository struct {
	findByDateFn                    func(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error)
	findAllByStaffIDsAndDateRangeFn func(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error)
	findAllBreaksByEntryIDsFn       func(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error)
}

func (m *mockLiffScheduleRepository) FindAllByMonth(_ context.Context, _, _ uint64, _ string) ([]model.ShiftEntry, error) {
	return nil, nil
}

func (m *mockLiffScheduleRepository) FindAllByStaffIDsAndDateRange(ctx context.Context, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error) {
	if m.findAllByStaffIDsAndDateRangeFn != nil {
		return m.findAllByStaffIDsAndDateRangeFn(ctx, clinicID, staffIDs, from, to)
	}
	return nil, nil
}

func (m *mockLiffScheduleRepository) FindAllBreaksByEntryIDs(ctx context.Context, clinicID uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
	if m.findAllBreaksByEntryIDsFn != nil {
		return m.findAllBreaksByEntryIDsFn(ctx, clinicID, entryIDs)
	}
	return nil, nil
}

func (m *mockLiffScheduleRepository) FindAllByDate(ctx context.Context, clinicID, staffID uint64, date time.Time) (*model.ShiftEntry, error) {
	if m.findByDateFn != nil {
		return m.findByDateFn(ctx, clinicID, staffID, date)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffScheduleRepository) FindAllBreaksByEntryID(_ context.Context, _, _ uint64) ([]model.ShiftEntryBreak, error) {
	return nil, nil
}

func (m *mockLiffScheduleRepository) Save(
	_ context.Context,
	_ uint64,
	_ *model.ShiftEntry,
	_ []model.ShiftEntryBreak,
) (*model.ShiftEntry, []model.ShiftEntryBreak, bool, error) {
	return nil, nil, false, nil
}

func (m *mockLiffScheduleRepository) Delete(_ context.Context, _, _ uint64, _ time.Time) error {
	return nil
}

// --- mockLiffAdminRepository ---

type mockLiffAdminRepository struct {
	findByDayFn                 func(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error)
	findByCustomerIDFn          func(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	cancelByIDFn                func(ctx context.Context, clinicID, customerID, id uint64) error
	findByIDForNotifyFn         func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	findTimeRangesByDateRangeFn func(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error)
}

func (m *mockLiffAdminRepository) FindAllByMonth(_ context.Context, _ uint64, _ int, _ time.Month) ([]model.Reservation, error) {
	return nil, nil
}

func (m *mockLiffAdminRepository) FindAllByDay(ctx context.Context, clinicID uint64, date time.Time) ([]model.Reservation, error) {
	if m.findByDayFn != nil {
		return m.findByDayFn(ctx, clinicID, date)
	}
	return nil, nil
}

func (m *mockLiffAdminRepository) FindTimeRangesByDateRange(ctx context.Context, clinicID uint64, from, to time.Time) ([]model.Reservation, error) {
	if m.findTimeRangesByDateRangeFn != nil {
		return m.findTimeRangesByDateRangeFn(ctx, clinicID, from, to)
	}
	return nil, nil
}

func (m *mockLiffAdminRepository) Create(_ context.Context, _ *model.Reservation) error {
	return nil
}

func (m *mockLiffAdminRepository) SoftDelete(_ context.Context, _, _ uint64) error { return nil }

func (m *mockLiffAdminRepository) FindAllByCustomerID(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
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

func (m *mockLiffAdminRepository) FindByIDForNotify(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	if m.findByIDForNotifyFn != nil {
		return m.findByIDForNotifyFn(ctx, clinicID, id)
	}
	return nil, nil
}

// --- mockLiffCustomerRepository ---

type mockLiffCustomerRepository struct {
	findByIDFn               func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	updateAdditionalFieldsFn func(ctx context.Context, clinicID, id uint64, fields []byte) error
	updateOwnerLinkFn        func(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
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

func (m *mockLiffCustomerRepository) UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error {
	if m.updateOwnerLinkFn != nil {
		return m.updateOwnerLinkFn(ctx, clinicID, id, ownerID)
	}
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

// --- mockLiffOwnerRepository ---

type mockLiffOwnerRepository struct {
	findByNameAndPhoneFn func(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error)
}

func (m *mockLiffOwnerRepository) FindAll(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}

func (m *mockLiffOwnerRepository) FindByID(_ context.Context, _, _ uint64) (*model.Owner, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffOwnerRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffOwnerRepository) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffOwnerRepository) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffOwnerRepository) FindByNameAndPhone(ctx context.Context, clinicID uint64, name, phone string) (*model.Owner, error) {
	if m.findByNameAndPhoneFn != nil {
		return m.findByNameAndPhoneFn(ctx, clinicID, name, phone)
	}
	return nil, nil
}

func (m *mockLiffOwnerRepository) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}

func (m *mockLiffOwnerRepository) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}

func (m *mockLiffOwnerRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLiffOwnerRepository) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mockLiffOwnerRepository) FindByLineUserID(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mockLiffOwnerRepository) FindAllWithLineUserID(_ context.Context, _ uint64) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLiffOwnerRepository) FindAllWithLineUserIDCursor(_ context.Context, _, _ uint64, _ int) ([]model.Owner, error) {
	return nil, nil
}
func (m *mockLiffOwnerRepository) UpdateLineUserID(_ context.Context, _, _ uint64, _ *string) error {
	return nil
}

func (m *mockLiffOwnerRepository) UpdateLineFollowedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}

func (m *mockLiffOwnerRepository) UpdateLineBlockedAt(
	_ context.Context,
	_, _ uint64,
	_ string,
	_ time.Time,
) (bool, error) {
	return true, nil
}

func (m *mockLiffOwnerRepository) FindByIDs(_ context.Context, _ uint64, _ []uint64) ([]*model.Owner, error) {
	return nil, nil
}

// --- mockLiffReservationRepository ---

type mockLiffReservationRepository struct {
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error)
}

func (m *mockLiffReservationRepository) FindAll(_ context.Context, _ []uint64, _, _ int, _, _, _ *time.Time, _, _ *string, _, _ *uint64) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockLiffReservationRepository) FindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffReservationRepository) FindByIDForClinics(_ context.Context, _ []uint64, _ uint64) (*model.Reservation, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffReservationRepository) Create(_ context.Context, _ *model.Reservation) error {
	return nil
}

func (m *mockLiffReservationRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Reservation, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Reservation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockLiffReservationRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLiffReservationRepository) ExistsByReservationTypeID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockLiffReservationRepository) ExistsByStaffID(_ context.Context, _, _ uint64) (bool, error) {
	return false, nil
}

func (m *mockLiffReservationRepository) CountMedicalRecordsByReservationID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) AcquireBookingLock(_ context.Context, _ uint64) error {
	return nil
}

func (m *mockLiffReservationRepository) LockAndFindByID(_ context.Context, _, _ uint64) (*model.Reservation, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockLiffReservationRepository) HasDoctorConflict(_ context.Context, _, _ uint64, _, _ time.Time, _ *uint64) (bool, error) {
	return false, nil
}

func (m *mockLiffReservationRepository) CountOnDutyDoctors(_ context.Context, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) CountConflicts(_ context.Context, _ uint64, _, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) CountByTypeAndStartTime(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) CountByCustomerAndDateRange(_ context.Context, _, _ uint64, _, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) CountByDateAndSource(_ context.Context, _ uint64, _ time.Time, _ model.ReservationSource) (int64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) FindAllByCategory(_ context.Context, _ uint64, _ model.ReservationTypeCategory, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
	return nil, 0, nil
}

func (m *mockLiffReservationRepository) FindNoShowCandidates(_ context.Context, _ uint64) ([]model.Reservation, error) {
	return nil, nil
}

func (m *mockLiffReservationRepository) AssertOwnerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockLiffReservationRepository) FindPetOwnerInClinic(_ context.Context, _, _ uint64) (uint64, error) {
	return 0, nil
}

func (m *mockLiffReservationRepository) FindPetByIDInClinic(_ context.Context, _, petID uint64) (*model.Pet, error) {
	return &model.Pet{ID: petID, Status: model.PetStatusAlive}, nil
}


func (m *mockLiffReservationRepository) AssertLineCustomerInClinic(_ context.Context, _, _ uint64) error {
	return nil
}

// --- mockLiffValidators ---

type mockLiffValidators struct {
	validateAndCreateFn func(ctx context.Context, input *CreateReservationInput) (*model.Reservation, error)
}

func (m *mockLiffValidators) ValidateAndCreate(ctx context.Context, input *CreateReservationInput) (*model.Reservation, error) {
	if m.validateAndCreateFn != nil {
		return m.validateAndCreateFn(ctx, input)
	}
	return &model.Reservation{ID: 1, ClinicID: input.ClinicID}, nil
}

// --- mockLiffNotifier ---

type mockLiffNotifier struct {
	notifyCreatedFn   func(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
	notifyCancelledFn func(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer)
}

func (m *mockLiffNotifier) NotifyCreated(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer) {
	if m.notifyCreatedFn != nil {
		m.notifyCreatedFn(ctx, appt, customer)
	}
}

func (m *mockLiffNotifier) NotifyCancelled(ctx context.Context, appt *model.Reservation, customer *model.LineCustomer) {
	if m.notifyCancelledFn != nil {
		m.notifyCancelledFn(ctx, appt, customer)
	}
}

func (m *mockLiffNotifier) Wait() {}

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
	return newLiffSvcWithDeps(
		setting,
		course,
		staff,
		schedule,
		admin,
		&mockLiffReservationRepository{},
		customer,
		&mockLiffOwnerRepository{},
		validators,
		notifier,
	)
}

func newLiffSvcWithDeps(
	setting *mockLiffSettingRepository,
	course *mockLiffTypeRepository,
	staff *mockLiffStaffRepository,
	schedule *mockLiffScheduleRepository,
	admin *mockLiffAdminRepository,
	reservation *mockLiffReservationRepository,
	customer *mockLiffCustomerRepository,
	owner *mockLiffOwnerRepository,
	validators *mockLiffValidators,
	notifier *mockLiffNotifier,
) *liffService {
	return &liffService{
		settingRepo:         setting,
		typeLiffRepo:        course,
		staffRepo:           staff,
		scheduleRepo:        schedule,
		adminRepo:           admin,
		reservationRepo:     reservation,
		customerRepo:        customer,
		ownerRepo:           owner,
		validators:          validators,
		notifier:            notifier,
		unavailableTimeRepo: &mockLiffUnavailableTimeRepository{},
		occupationRepo:      &mockReservationTypeOccupationRepository{},
	}
}

// mockLiffUnavailableTimeRepository は ReservationTypeUnavailableTimeRepository のテスト用スタブ
type mockLiffUnavailableTimeRepository struct {
	findAllFn func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
}

func (m *mockLiffUnavailableTimeRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeUnavailableTime{}, nil
}
func (m *mockLiffUnavailableTimeRepository) FindByID(_ context.Context, _, _ uint64) (*model.ReservationTypeUnavailableTime, error) {
	return nil, nil
}
func (m *mockLiffUnavailableTimeRepository) Create(_ context.Context, _ *model.ReservationTypeUnavailableTime) error {
	return nil
}
func (m *mockLiffUnavailableTimeRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

// liffDefaultSetting はテスト用デフォルト予約設定を返す。
func liffDefaultSetting() *model.LineReservationSetting {
	return &model.LineReservationSetting{
		ID:                      1,
		ClinicID:                3,
		Status:                  "active",
		LiffID:                  "1234567890-abcdefgh",
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
