package staff_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type blockingAssignmentRaceReservationRepository struct {
	reservation.ReservationRepository
	created     chan struct{}
	release     chan struct{}
	createdOnce sync.Once
}

func (r *blockingAssignmentRaceReservationRepository) Create(
	ctx context.Context,
	appointment *model.Reservation,
) error {
	if err := r.ReservationRepository.Create(ctx, appointment); err != nil {
		return err
	}
	r.createdOnce.Do(func() {
		close(r.created)
	})
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type observedAssignmentRaceStaffUpdateLocker struct {
	staffpkg.Repository
	started      chan struct{}
	returned     chan struct{}
	startedOnce  sync.Once
	returnedOnce sync.Once
}

func (r *observedAssignmentRaceStaffUpdateLocker) LockActiveByIDForUpdate(
	ctx context.Context,
	staffID uint64,
) (*model.Staff, error) {
	r.startedOnce.Do(func() {
		close(r.started)
	})
	staff, err := r.Repository.LockActiveByIDForUpdate(ctx, staffID)
	r.returnedOnce.Do(func() {
		close(r.returned)
	})
	return staff, err
}

type blockingAssignmentRaceClinicLookup struct {
	*clinicdomain.Repository
	locked     chan struct{}
	release    chan struct{}
	lockedOnce sync.Once
}

func (r *blockingAssignmentRaceClinicLookup) LockActiveByID(
	ctx context.Context,
	clinicID uint64,
) (*model.Clinic, error) {
	clinic, err := r.Repository.LockActiveByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	r.lockedOnce.Do(func() {
		close(r.locked)
	})
	select {
	case <-r.release:
		return clinic, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type observedAssignmentRaceReservationStaffRepository struct {
	reservation.ReservationStaffRepository
	started      chan struct{}
	returned     chan struct{}
	startedOnce  sync.Once
	returnedOnce sync.Once
}

func (r *observedAssignmentRaceReservationStaffRepository) FindByID(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	inWriteTransaction := persistence.TxFromContext(ctx) != nil
	if inWriteTransaction {
		r.startedOnce.Do(func() {
			close(r.started)
		})
	}
	staff, err := r.ReservationStaffRepository.FindByID(ctx, clinicID, staffID)
	if inWriteTransaction {
		r.returnedOnce.Do(func() {
			close(r.returned)
		})
	}
	return staff, err
}

type assignmentRaceFixture struct {
	db                *gorm.DB
	sourceClinic      *model.Clinic
	targetClinic      *model.Clinic
	staff             *model.Staff
	reservationType   *model.ReservationType
	staffRepo         staffpkg.Repository
	assignmentRepo    staffpkg.StaffClinicAssignmentRepository
	shiftRepo         staffpkg.ShiftEntryRepository
	clinicRepo        *clinicdomain.Repository
	reservationRepo   reservation.ReservationStore
	reservationStaff  reservation.ReservationStaffRepository
	transactor        persistence.Transactor
	reservationCreate *reservation.CreateManualReservationInput
}

func setupStaffAssignmentReservationRaceTest(t *testing.T) *assignmentRaceFixture {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ShiftEntry{},
		&model.ReservationType{},
		&model.StaffReservationCapability{},
		&model.Reservation{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE appointments, staff_reservation_capabilities, shift_entries, staff_clinic_assignments, staffs, reservation_types, clinics, companies CASCADE",
	).Error)
	require.NoError(t, db.Exec("DROP INDEX IF EXISTS ux_test_staff_assignment_reservation_race").Error)
	require.NoError(t, db.Exec(`
		CREATE UNIQUE INDEX ux_test_staff_assignment_reservation_race
		ON staff_clinic_assignments (staff_id, clinic_id)
	`).Error)

	company := &model.Company{Name: "staff assignment reservation race company"}
	require.NoError(t, db.Create(company).Error)
	sourceClinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "staff assignment reservation race source",
		IsActive:  true,
	}
	targetClinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "staff assignment reservation race target",
		IsActive:  true,
	}
	require.NoError(t, db.Create(sourceClinic).Error)
	require.NoError(t, db.Create(targetClinic).Error)
	staff := &model.Staff{
		ClinicID:           sourceClinic.ID,
		Name:               "staff assignment reservation race doctor",
		IsActive:           true,
		ReservationVisible: true,
		StaffType:          model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)

	staffRepo := staffpkg.NewRepository(db)
	assignmentRepo := staffpkg.NewStaffClinicAssignmentRepository(db)
	require.NoError(t, assignmentRepo.Create(context.Background(), &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: sourceClinic.ID,
		IsMain:   true,
	}))
	reservationType := &model.ReservationType{
		ClinicID: sourceClinic.ID,
		Name:     "staff assignment reservation race type",
		IsActive: true,
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.Create(reservationType).Error)
	require.NoError(t, db.Create(&model.StaffReservationCapability{
		ClinicID:          sourceClinic.ID,
		StaffID:           staff.ID,
		ReservationTypeID: reservationType.ID,
	}).Error)

	reservationRepo := reservation.NewReservationRepository(db)
	reservationStaff := reservation.NewReservationStaffRepository(db, staffRepo)
	transactor := persistence.NewTransactor(db)
	route := "reception"
	start := time.Date(2026, 7, 25, 9, 0, 0, 0, time.UTC)
	return &assignmentRaceFixture{
		db:               db,
		sourceClinic:     sourceClinic,
		targetClinic:     targetClinic,
		staff:            staff,
		reservationType:  reservationType,
		staffRepo:        staffRepo,
		assignmentRepo:   assignmentRepo,
		shiftRepo:        staffpkg.NewShiftEntryRepository(db),
		clinicRepo:       clinicdomain.NewClinicRepository(db),
		reservationRepo:  reservationRepo,
		reservationStaff: reservationStaff,
		transactor:       transactor,
		reservationCreate: &reservation.CreateManualReservationInput{
			ClinicID:          sourceClinic.ID,
			StartTime:         start,
			EndTime:           start.Add(30 * time.Minute),
			VisitType:         model.VisitTypeRevisit,
			ReservationTypeID: reservationType.ID,
			DoctorID:          &staff.ID,
			Status:            model.ReservationStatusPending,
			Source:            model.ReservationSourceManual,
			ReservationRoute:  &route,
		},
	}
}

func newAssignmentRaceService(
	fixture *assignmentRaceFixture,
	staffRepo staffpkg.Repository,
	clinicRepo staffpkg.StaffAssignmentClinicLookup,
) staffpkg.Service {
	return staffpkg.NewService(
		staffRepo,
		nil,
		fixture.assignmentRepo,
		fixture.reservationRepo,
		fixture.shiftRepo,
		nil,
		nil,
		nil,
		clinicRepo,
		fixture.transactor,
	)
}

func newAssignmentRaceReservationService(
	fixture *assignmentRaceFixture,
	reservationRepo reservation.ReservationRepository,
	reservationStaff reservation.ReservationStaffRepository,
) reservation.ReservationService {
	return reservation.NewReservationServiceWithAvailabilityAndType(
		reservationRepo,
		reservation.NewReservationTypeRepository(fixture.db),
		fixture.transactor,
		reservationStaff,
		nil,
	)
}

func TestStaffSetAssignmentsAndReservationCreate_ReservationWinnerPreservesAssignmentDatabase(t *testing.T) {
	fixture := setupStaffAssignmentReservationRaceTest(t)
	blockingReservationRepo := &blockingAssignmentRaceReservationRepository{
		ReservationRepository: fixture.reservationRepo,
		created:               make(chan struct{}),
		release:               make(chan struct{}),
	}
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() {
			close(blockingReservationRepo.release)
		})
	})
	observedStaffRepo := &observedAssignmentRaceStaffUpdateLocker{
		Repository: fixture.staffRepo,
		started:    make(chan struct{}),
		returned:   make(chan struct{}),
	}
	reservationSvc := newAssignmentRaceReservationService(
		fixture,
		blockingReservationRepo,
		fixture.reservationStaff,
	)
	staffSvc := newAssignmentRaceService(fixture, observedStaffRepo, fixture.clinicRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reservationDone := make(chan error, 1)
	go func() {
		_, err := reservationSvc.Create(ctx, fixture.reservationCreate)
		reservationDone <- err
	}()
	select {
	case <-blockingReservationRepo.created:
	case <-ctx.Done():
		t.Fatal("reservation create did not reach its blocking write")
	}

	assignmentDone := make(chan error, 1)
	go func() {
		assignmentDone <- staffSvc.SetClinicAssignments(ctx, &staffpkg.SetClinicAssignmentsInput{
			StaffID: fixture.staff.ID,
			ClinicIDs: []uint64{
				fixture.targetClinic.ID,
			},
			AuthorizedClinicIDs: []uint64{
				fixture.sourceClinic.ID,
				fixture.targetClinic.ID,
			},
		})
	}()
	select {
	case <-observedStaffRepo.started:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not attempt the staff update lock")
	}
	select {
	case <-observedStaffRepo.returned:
		t.Fatal("assignment replacement escaped the reservation transaction share lock")
	case <-time.After(100 * time.Millisecond):
	}

	releaseOnce.Do(func() {
		close(blockingReservationRepo.release)
	})
	require.NoError(t, <-reservationDone)
	assignmentErr := <-assignmentDone
	require.Error(t, assignmentErr)
	assert.True(t, apperrors.IsConflict(assignmentErr), "unexpected assignment error: %v", assignmentErr)

	assignments, err := fixture.assignmentRepo.FindByStaffID(context.Background(), fixture.staff.ID)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	assert.Equal(t, fixture.sourceClinic.ID, assignments[0].ClinicID)
	var appointmentCount int64
	require.NoError(t, fixture.db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND doctor_id = ?", fixture.sourceClinic.ID, fixture.staff.ID).
		Count(&appointmentCount).Error)
	assert.Equal(t, int64(1), appointmentCount)
}

func TestStaffSetAssignmentsAndReservationCreate_AssignmentWinnerPreventsUnassignedReservationDatabase(t *testing.T) {
	fixture := setupStaffAssignmentReservationRaceTest(t)
	blockingClinicRepo := &blockingAssignmentRaceClinicLookup{
		Repository: fixture.clinicRepo,
		locked:     make(chan struct{}),
		release:    make(chan struct{}),
	}
	var releaseOnce sync.Once
	t.Cleanup(func() {
		releaseOnce.Do(func() {
			close(blockingClinicRepo.release)
		})
	})
	observedReservationStaff := &observedAssignmentRaceReservationStaffRepository{
		ReservationStaffRepository: fixture.reservationStaff,
		started:                    make(chan struct{}),
		returned:                   make(chan struct{}),
	}
	staffSvc := newAssignmentRaceService(fixture, fixture.staffRepo, blockingClinicRepo)
	reservationSvc := newAssignmentRaceReservationService(
		fixture,
		fixture.reservationRepo,
		observedReservationStaff,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	assignmentDone := make(chan error, 1)
	go func() {
		assignmentDone <- staffSvc.SetClinicAssignments(ctx, &staffpkg.SetClinicAssignmentsInput{
			StaffID: fixture.staff.ID,
			ClinicIDs: []uint64{
				fixture.targetClinic.ID,
			},
			AuthorizedClinicIDs: []uint64{
				fixture.sourceClinic.ID,
				fixture.targetClinic.ID,
			},
		})
	}()
	select {
	case <-blockingClinicRepo.locked:
	case <-ctx.Done():
		t.Fatal("assignment replacement did not acquire the target clinic lock")
	}

	reservationDone := make(chan error, 1)
	go func() {
		_, err := reservationSvc.Create(ctx, fixture.reservationCreate)
		reservationDone <- err
	}()
	select {
	case <-observedReservationStaff.started:
	case <-ctx.Done():
		t.Fatal("reservation create did not attempt its in-transaction staff share lock")
	}
	select {
	case <-observedReservationStaff.returned:
		t.Fatal("reservation create escaped the assignment replacement transaction lock")
	case <-time.After(100 * time.Millisecond):
	}

	releaseOnce.Do(func() {
		close(blockingClinicRepo.release)
	})
	require.NoError(t, <-assignmentDone)
	reservationErr := <-reservationDone
	require.Error(t, reservationErr)
	assert.True(t, apperrors.IsNotFound(reservationErr), "unexpected reservation error: %v", reservationErr)

	assignments, err := fixture.assignmentRepo.FindByStaffID(context.Background(), fixture.staff.ID)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	assert.Equal(t, fixture.targetClinic.ID, assignments[0].ClinicID)
	var appointmentCount int64
	require.NoError(t, fixture.db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND doctor_id = ?", fixture.sourceClinic.ID, fixture.staff.ID).
		Count(&appointmentCount).Error)
	assert.Zero(t, appointmentCount)
}
