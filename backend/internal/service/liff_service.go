package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	settingRepo         repository.LineReservationSettingRepository
	typeLiffRepo        repository.ReservationTypeLiffRepository
	typeRepo            reservationTypeFinder
	staffRepo           repository.ReservationStaffRepository
	scheduleRepo        repository.ReservationScheduleRepository
	adminRepo           repository.ReservationAdminRepository
	reservationRepo     repository.ReservationRepository
	customerRepo        repository.LineCustomerRepository
	ownerRepo           repository.OwnerRepository
	validators          ReservationValidators
	notifier            ReservationNotifier
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository // BE-117
	availableSlotRepo   repository.ReservationTypeAvailableSlotRepository
	occupationRepo      repository.ReservationTypeOccupationRepository // BE-117
	trimmingCourseRepo  repository.TrimmingCourseRepository            // BE-120
	trimmingOptionRepo  repository.TrimmingOptionRepository            // BE-120
	trimmingDetailRepo  repository.AppointmentTrimmingDetailRepository // BE-120
	vaccinationRepo     repository.VaccinationRepository
}

func NewLiffServiceWithType(
	settingRepo repository.LineReservationSettingRepository,
	typeLiffRepo repository.ReservationTypeLiffRepository,
	typeRepo reservationTypeFinder,
	staffRepo repository.ReservationStaffRepository,
	scheduleRepo repository.ReservationScheduleRepository,
	adminRepo repository.ReservationAdminRepository,
	customerRepo repository.LineCustomerRepository,
	ownerRepo repository.OwnerRepository,
	tx repository.Transactor,
	reservationRepo repository.ReservationRepository,
	notifier ReservationNotifier,
	unavailableTimeRepo repository.ReservationTypeUnavailableTimeRepository,
	availableSlotRepo repository.ReservationTypeAvailableSlotRepository,
	occupationRepo repository.ReservationTypeOccupationRepository,
	trimmingCourseRepo repository.TrimmingCourseRepository,
	trimmingOptionRepo repository.TrimmingOptionRepository,
	trimmingDetailRepo repository.AppointmentTrimmingDetailRepository,
	vaccinationRepo repository.VaccinationRepository,
) LiffService {
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
		validators:          NewReservationValidators(tx, reservationRepo, typeRepo, trimmingCourseRepo, trimmingOptionRepo),
		notifier:            notifier,
		unavailableTimeRepo: unavailableTimeRepo,
		availableSlotRepo:   availableSlotRepo,
		occupationRepo:      occupationRepo,
		trimmingCourseRepo:  trimmingCourseRepo,
		trimmingOptionRepo:  trimmingOptionRepo,
		trimmingDetailRepo:  trimmingDetailRepo,
		vaccinationRepo:     vaccinationRepo,
	}
}
