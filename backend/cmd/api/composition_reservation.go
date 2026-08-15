package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/trimming"
)

type reservationRepositories struct {
	Reservations            reservation.ReservationStore
	ReservationTypes        reservation.ReservationTypeRepository
	ReservationTypeGroups   reservation.ReservationTypeGroupRepository
	ReservationTypeLiff     reservation.ReservationTypeLiffRepository
	ReservationStaff        reservation.ReservationStaffRepository
	ReservationSchedules    reservation.ReservationScheduleRepository
	ReservationAdmin        reservation.ReservationAdminRepository
	LineReservationSettings reservation.LineReservationSettingRepository
	UnavailableTimes        reservation.ReservationTypeUnavailableTimeRepository
	AvailableSlots          reservation.ReservationTypeAvailableSlotRepository
	TypeOccupations         reservation.ReservationTypeOccupationRepository
	Occupations             staff.OccupationRepository
}

func newReservationRepositories(
	db *gorm.DB,
	staffRepository staff.StaffRepository,
	shiftEntries staff.ShiftEntryRepository,
	occupations staff.OccupationRepository,
) reservationRepositories {
	return reservationRepositories{
		Reservations:            reservation.NewReservationRepository(db),
		ReservationTypes:        reservation.NewReservationTypeRepository(db),
		ReservationTypeGroups:   reservation.NewReservationTypeGroupRepository(db),
		ReservationTypeLiff:     reservation.NewReservationTypeLiffRepository(db),
		ReservationStaff:        reservation.NewReservationStaffRepository(db, staffRepository),
		ReservationSchedules:    reservation.NewReservationScheduleRepository(db, shiftEntries),
		ReservationAdmin:        reservation.NewReservationAdminRepository(db),
		LineReservationSettings: reservation.NewLineReservationSettingRepository(db),
		UnavailableTimes:        reservation.NewReservationTypeUnavailableTimeRepository(db),
		AvailableSlots:          reservation.NewReservationTypeAvailableSlotRepository(db),
		TypeOccupations:         reservation.NewReservationTypeOccupationRepository(db),
		Occupations:             occupations,
	}
}

type reservationMedicalRecords interface {
	AutoCreateFromReservation(
		ctx context.Context,
		clinicID uint64,
		reservation *model.Reservation,
	)
	DeleteDraftFromReservation(ctx context.Context, clinicID, reservationID uint64)
}

type reservationServiceDependencies struct {
	Transactor      reservation.Transactor
	StaffDeleter    reservation.ReservationStaffDeleter
	LineCustomers   reservation.LiffLineCustomerRepository
	Owners          owner.LookupRepository
	TrimmingCourses trimming.TrimmingCourseRepository
	TrimmingOptions trimming.TrimmingOptionRepository
	TrimmingDetails trimming.AppointmentTrimmingDetailRepository
	Vaccinations    medicalrecord.VaccinationRepository
	MedicalRecords  reservationMedicalRecords

	NotificationConfig *reservation.ReservationNotificationConfig
	EncryptCredential  func(value string) (string, error)
	DecryptCredential  func(ctx context.Context, value string) string
	NewLineMessenger   func(channelToken string) reservation.LinePusher
	SendMail           func(
		ctx context.Context,
		cfg reservation.SMTPConfig,
		from, to string,
		message []byte,
	) error
}

type reservationComposition struct {
	Repositories            reservationRepositories
	Reservations            reservation.ReservationService
	ReservationTypes        reservation.ReservationTypeService
	ReservationTypeGroups   reservation.ReservationTypeGroupService
	ReservationTypeLiff     reservation.ReservationTypeLiffService
	ReservationStaff        reservation.ReservationStaffService
	ReservationSchedules    reservation.ReservationScheduleService
	ReservationAdmin        reservation.ReservationAdminService
	LineReservationSettings reservation.LineReservationSettingService
	Liff                    reservation.LiffService
	Notifier                reservation.ReservationNotifier
	DrainNotifications      func()
	medicalRecords          reservationMedicalRecords
}

type reservationMasterServices struct {
	types        reservation.ReservationTypeService
	groups       reservation.ReservationTypeGroupService
	typeLiff     reservation.ReservationTypeLiffService
	staff        reservation.ReservationStaffService
	schedules    reservation.ReservationScheduleService
	lineSettings reservation.LineReservationSettingService
}

type reservationBookingServices struct {
	reservations reservation.ReservationService
	admin        reservation.ReservationAdminService
	liff         reservation.LiffService
}

func newReservationComposition(
	repositories reservationRepositories,
	dependencies reservationServiceDependencies,
) reservationComposition {
	notifier := newReservationNotifier(
		repositories.LineReservationSettings,
		dependencies,
	)
	masters := newReservationMasterServices(repositories, dependencies)
	booking := newReservationBookingServices(
		repositories,
		dependencies,
		notifier,
	)

	return reservationComposition{
		Repositories:            repositories,
		Reservations:            booking.reservations,
		ReservationTypes:        masters.types,
		ReservationTypeGroups:   masters.groups,
		ReservationTypeLiff:     masters.typeLiff,
		ReservationStaff:        masters.staff,
		ReservationSchedules:    masters.schedules,
		ReservationAdmin:        booking.admin,
		LineReservationSettings: masters.lineSettings,
		Liff:                    booking.liff,
		Notifier:                notifier,
		DrainNotifications:      reservationNotificationDrainer(notifier),
		medicalRecords:          dependencies.MedicalRecords,
	}
}

func newReservationMasterServices(
	repositories reservationRepositories,
	dependencies reservationServiceDependencies,
) reservationMasterServices {
	reservationTypes := reservation.NewReservationTypeService(
		repositories.ReservationTypes,
		repositories.UnavailableTimes,
		repositories.TypeOccupations,
		repositories.Occupations,
		repositories.ReservationTypeGroups,
		repositories.AvailableSlots,
	)

	return reservationMasterServices{
		types:     reservationTypes,
		groups:    reservation.NewReservationTypeGroupService(repositories.ReservationTypeGroups),
		typeLiff:  reservation.NewReservationTypeLiffService(repositories.ReservationTypeLiff, repositories.Reservations),
		staff:     reservation.NewReservationStaffService(repositories.ReservationStaff, dependencies.Transactor, dependencies.StaffDeleter),
		schedules: reservation.NewReservationScheduleService(repositories.ReservationSchedules),
		lineSettings: reservation.NewLineReservationSettingService(
			repositories.LineReservationSettings,
			dependencies.EncryptCredential,
			dependencies.DecryptCredential,
		),
	}
}

func newReservationBookingServices(
	repositories reservationRepositories,
	dependencies reservationServiceDependencies,
	notifier reservation.ReservationNotifier,
) reservationBookingServices {
	return reservationBookingServices{
		reservations: newReservationService(repositories, dependencies),
		admin:        newReservationAdminService(repositories, dependencies),
		liff:         newReservationLiffService(repositories, dependencies, notifier),
	}
}

func newReservationService(
	r reservationRepositories,
	d reservationServiceDependencies,
) reservation.ReservationService {
	return reservation.NewReservationServiceWithAvailabilityAndType(
		r.Reservations,
		r.ReservationTypes,
		d.Transactor,
		r.ReservationStaff,
		r.UnavailableTimes,
		r.AvailableSlots,
	)
}

func newReservationAdminService(
	r reservationRepositories,
	d reservationServiceDependencies,
) reservation.ReservationAdminService {
	return reservation.NewReservationAdminServiceWithMedicalRecord(
		r.ReservationAdmin,
		r.Reservations,
		r.ReservationTypes,
		d.Transactor,
		r.ReservationStaff,
		r.UnavailableTimes,
		r.AvailableSlots,
		d.MedicalRecords,
	)
}

func newReservationLiffService(
	r reservationRepositories,
	d reservationServiceDependencies,
	notifier reservation.ReservationNotifier,
) reservation.LiffService {
	return reservation.NewLiffServiceWithType(
		r.LineReservationSettings,
		r.ReservationTypeLiff,
		r.ReservationTypes,
		r.ReservationStaff,
		r.ReservationSchedules,
		r.ReservationAdmin,
		d.LineCustomers,
		d.Owners,
		d.Transactor,
		r.Reservations,
		notifier,
		r.UnavailableTimes,
		r.AvailableSlots,
		r.TypeOccupations,
		d.TrimmingCourses,
		d.TrimmingOptions,
		d.TrimmingDetails,
		d.Vaccinations,
		d.MedicalRecords,
	)
}

func newReservationNotifier(
	settings reservation.LineReservationSettingRepository,
	dependencies reservationServiceDependencies,
) reservation.ReservationNotifier {
	config := dependencies.NotificationConfig
	if config == nil {
		config = &reservation.ReservationNotificationConfig{}
	}
	return reservation.NewReservationNotificationService(
		config,
		settings,
		dependencies.DecryptCredential,
		dependencies.NewLineMessenger,
		dependencies.SendMail,
	)
}

func reservationNotificationDrainer(
	notifier reservation.ReservationNotifier,
) func() {
	return func() {
		if notifier != nil {
			notifier.Wait()
		}
	}
}

type reservationStaffAssignments interface {
	FindAllByStaffID(
		ctx context.Context,
		staffID uint64,
	) ([]model.StaffClinicAssignment, error)
}

type reservationHandlerDependencies struct {
	StaffAssignments  reservationStaffAssignments
	LiffAuth          gin.HandlerFunc
	LiffRateLimit     func(limit int) gin.HandlerFunc
	LinkLiffAccount   gin.HandlerFunc
	RequirePermission reservation.PermissionMiddleware
}

func (c reservationComposition) newHandler(
	dependencies reservationHandlerDependencies,
) *reservation.Handler {
	return reservation.NewHandler(
		reservation.NewReservationTypeHandler(
			c.ReservationTypes,
			c.ReservationTypes,
			c.ReservationTypes,
			c.ReservationTypes,
		),
		reservation.NewReservationTypeGroupHandler(c.ReservationTypeGroups),
		reservation.NewReservationTypeLiffHandler(c.ReservationTypeLiff),
		reservation.NewReservationStaffHandler(c.ReservationStaff),
		reservation.NewReservationScheduleHandler(c.ReservationSchedules),
		reservation.NewReservationHandler(
			c.Reservations,
			c.medicalRecords,
			c.Liff,
			dependencies.StaffAssignments,
		),
		reservation.NewReservationAdminHandler(
			c.ReservationAdmin,
			dependencies.StaffAssignments,
		),
		reservation.NewLineReservationSettingHandler(c.LineReservationSettings),
		reservation.NewLiffHandler(c.Liff, dependencies.StaffAssignments),
		dependencies.LiffAuth,
		dependencies.LiffRateLimit,
		dependencies.LinkLiffAccount,
		dependencies.RequirePermission,
	)
}
