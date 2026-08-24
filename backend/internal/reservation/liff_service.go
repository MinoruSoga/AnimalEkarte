package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// LiffService はLIFF公開APIのビジネスロジックインターフェース
type LiffService interface {
	GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error)
	GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error)
	GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
	GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
	CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error)
	GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error
	GetHealthCard(ctx context.Context, clinicID, customerID uint64) (*HealthCardResult, error)
}

type liffService struct {
	settingRepo         lineReservationSettingFinder
	typeLiffRepo        ReservationTypeLiffRepository
	typeRepo            reservationTypeFinder
	staffRepo           ReservationStaffRepository
	scheduleRepo        ReservationScheduleRepository
	adminRepo           ReservationAdminRepository
	reservationRepo     ReservationRepository
	customerRepo        LiffLineCustomerRepository
	ownerRepo           liffOwnerRepo
	validators          ReservationValidators
	notifier            ReservationNotifier
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository // BE-117
	availableSlotRepo   ReservationTypeAvailableSlotRepository
	occupationRepo      ReservationTypeOccupationRepository // BE-117
	trimmingCourseRepo  trimmingCourseFinder                // BE-120
	trimmingOptionRepo  trimmingOptionFinder                // BE-120
	trimmingDetailRepo  liffTrimmingDetailRepo              // BE-120
	vaccinationRepo     liffVaccinationRepo
	medicalRecord       medicalRecordAutoCreator
}

func NewLiffServiceWithType(
	settingRepo lineReservationSettingFinder,
	typeLiffRepo ReservationTypeLiffRepository,
	typeRepo reservationTypeFinder,
	staffRepo ReservationStaffRepository,
	scheduleRepo ReservationScheduleRepository,
	adminRepo ReservationAdminRepository,
	customerRepo LiffLineCustomerRepository,
	ownerRepo liffOwnerRepo,
	tx Transactor,
	reservationRepo ReservationRepository,
	notifier ReservationNotifier,
	unavailableTimeRepo ReservationTypeUnavailableTimeRepository,
	availableSlotRepo ReservationTypeAvailableSlotRepository,
	occupationRepo ReservationTypeOccupationRepository,
	trimmingCourseRepo trimmingCourseFinder,
	trimmingOptionRepo trimmingOptionFinder,
	trimmingDetailRepo liffTrimmingDetailRepo,
	vaccinationRepo liffVaccinationRepo,
	holidayFinder clinicHolidayFinder,
	medicalRecords ...medicalRecordAutoCreator,
) LiffService {
	var medicalRecord medicalRecordAutoCreator
	if len(medicalRecords) > 0 {
		medicalRecord = medicalRecords[0]
	}
	return &liffService{
		settingRepo:         settingRepo,
		typeLiffRepo:        typeLiffRepo,
		typeRepo:            typeRepo,
		staffRepo:           staffRepo,
		scheduleRepo:        scheduleRepo,
		adminRepo:           adminRepo,
		reservationRepo:     reservationRepo,
		customerRepo:        customerRepo,
		ownerRepo:           ownerRepo,
		validators:          NewReservationValidators(tx, reservationRepo, typeRepo, staffRepo, trimmingCourseRepo, trimmingOptionRepo, trimmingDetailRepo, holidayFinder),
		notifier:            notifier,
		unavailableTimeRepo: unavailableTimeRepo,
		availableSlotRepo:   availableSlotRepo,
		occupationRepo:      occupationRepo,
		trimmingCourseRepo:  trimmingCourseRepo,
		trimmingOptionRepo:  trimmingOptionRepo,
		trimmingDetailRepo:  trimmingDetailRepo,
		vaccinationRepo:     vaccinationRepo,
		medicalRecord:       medicalRecord,
	}
}
