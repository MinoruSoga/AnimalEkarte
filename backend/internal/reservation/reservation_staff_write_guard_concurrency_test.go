package reservation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type reservationStaffWriteOperation string

const (
	reservationStaffWriteStandardCreate reservationStaffWriteOperation = "standard_create"
	reservationStaffWriteStandardUpdate reservationStaffWriteOperation = "standard_update"
	reservationStaffWriteAdminCreate    reservationStaffWriteOperation = "admin_create"
)

type reservationStaffCompetingMutation string

const (
	reservationStaffMutationDeleteStaff      reservationStaffCompetingMutation = "delete_staff"
	reservationStaffMutationRemoveAssignment reservationStaffCompetingMutation = "remove_assignment"
)

type pausingReservationStaffWriteGuard struct {
	delegate ReservationStaffWriteGuard
	locked   chan struct{}
	release  chan struct{}
	once     sync.Once
}

func (g *pausingReservationStaffWriteGuard) FindByID(
	ctx context.Context,
	clinicID, staffID uint64,
) (*model.Staff, error) {
	return g.delegate.FindByID(ctx, clinicID, staffID)
}

func (g *pausingReservationStaffWriteGuard) SupportsReservationType(
	ctx context.Context,
	clinicID, staffID, reservationTypeID uint64,
) (bool, error) {
	supports, err := g.delegate.SupportsReservationType(
		ctx,
		clinicID,
		staffID,
		reservationTypeID,
	)
	if err != nil {
		return false, err
	}
	g.once.Do(func() {
		close(g.locked)
		<-g.release
	})
	return supports, nil
}

type reservationStaffWriteGuardFixture struct {
	clinicID        uint64
	staffID         uint64
	assignmentID    uint64
	reservationType uint64
	reservationID   uint64
	originalStart   time.Time
	originalEnd     time.Time
	targetStart     time.Time
	targetEnd       time.Time
}

func makeReservationStaffWriteGuardClinic(t *testing.T, db *gorm.DB) *model.Clinic {
	t.Helper()

	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}))
	company := &model.Company{Name: "予約書込競合ガード法人"}
	require.NoError(t, db.Create(company).Error)

	clinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "予約書込競合ガード医院",
	}
	require.NoError(t, db.Create(clinic).Error)
	return clinic
}

func setupReservationStaffWriteGuardFixture(
	t *testing.T,
	db *gorm.DB,
	operation reservationStaffWriteOperation,
) reservationStaffWriteGuardFixture {
	t.Helper()

	clinicID := makeReservationStaffWriteGuardClinic(t, db).ID
	staff := makeDoctor(t, db, clinicID, "予約書込競合ガード担当者")
	assignment := &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   true,
	}
	require.NoError(t, db.Create(assignment).Error)
	reservationType := makeReservationType(t, db, clinicID)
	require.NoError(t, db.Create(&model.StaffReservationCapability{
		ClinicID:          clinicID,
		StaffID:           staff.ID,
		ReservationTypeID: reservationType.ID,
	}).Error)

	fixture := reservationStaffWriteGuardFixture{
		clinicID:        clinicID,
		staffID:         staff.ID,
		assignmentID:    assignment.ID,
		reservationType: reservationType.ID,
		originalStart:   time.Date(2027, 7, 25, 9, 0, 0, 0, time.UTC),
		originalEnd:     time.Date(2027, 7, 25, 9, 30, 0, 0, time.UTC),
		targetStart:     time.Date(2027, 7, 25, 10, 0, 0, 0, time.UTC),
		targetEnd:       time.Date(2027, 7, 25, 10, 30, 0, 0, time.UTC),
	}
	if operation != reservationStaffWriteStandardUpdate {
		return fixture
	}

	current := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         fixture.originalStart,
		EndTime:           fixture.originalEnd,
		VisitType:         model.VisitTypeRevisit,
		ReservationTypeID: reservationType.ID,
		DoctorID:          &staff.ID,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.Create(current).Error)
	fixture.reservationID = current.ID
	return fixture
}

type reservationStaffWriteResult struct {
	reservation *model.Reservation
	err         error
}

func runReservationStaffWriteGuardOperation(
	ctx context.Context,
	db *gorm.DB,
	operation reservationStaffWriteOperation,
	fixture *reservationStaffWriteGuardFixture,
	guard ReservationStaffWriteGuard,
) (*model.Reservation, error) {
	reservationRepo := NewReservationRepository(db)
	tx := testNewTransactor(db)

	switch operation {
	case reservationStaffWriteStandardCreate:
		route := "reception"
		service := NewReservationServiceWithAvailabilityAndType(
			reservationRepo,
			nil,
			tx,
			guard,
			nil,
		)
		return service.Create(ctx, &CreateManualReservationInput{
			ClinicID:          fixture.clinicID,
			StartTime:         fixture.targetStart,
			EndTime:           fixture.targetEnd,
			VisitType:         model.VisitTypeRevisit,
			ReservationTypeID: fixture.reservationType,
			DoctorID:          &fixture.staffID,
			Status:            model.ReservationStatusPending,
			Source:            model.ReservationSourceManual,
			ReservationRoute:  &route,
		})
	case reservationStaffWriteStandardUpdate:
		service := NewReservationServiceWithClinicHolidays(
			reservationRepo,
			nil,
			tx,
			guard,
			nil,
			nil,
			nil,
			openDayHolidayFinder(),
		)
		return service.Update(
			ctx,
			fixture.clinicID,
			fixture.reservationID,
			&UpdateReservationInput{
				StartTime: &fixture.targetStart,
				EndTime:   &fixture.targetEnd,
				DoctorID:  &fixture.staffID,
			},
		)
	case reservationStaffWriteAdminCreate:
		service := NewReservationAdminServiceWithClinicHolidays(
			NewReservationAdminRepository(db),
			reservationRepo,
			nil,
			tx,
			guard,
			nil,
			nil,
			openDayHolidayFinder(),
		)
		return service.Create(ctx, fixture.clinicID, &CreateReservationAdminInput{
			StartTime:         fixture.targetStart,
			EndTime:           fixture.targetEnd,
			VisitType:         string(model.VisitTypeRevisit),
			ReservationTypeID: fixture.reservationType,
			DoctorID:          &fixture.staffID,
		})
	default:
		return nil, fmt.Errorf("unknown reservation write operation: %s", operation)
	}
}

func attemptReservationStaffCompetingMutation(
	db *gorm.DB,
	mutation reservationStaffCompetingMutation,
	fixture *reservationStaffWriteGuardFixture,
) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// The lock holder is paused until this statement returns. PostgreSQL's
		// lock timeout therefore gives a deterministic blocked/not-blocked
		// result without timing a goroutine with sleep or time.After.
		if err := tx.Exec("SET LOCAL lock_timeout = '100ms'").Error; err != nil {
			return err
		}

		var result *gorm.DB
		switch mutation {
		case reservationStaffMutationDeleteStaff:
			result = tx.Model(&model.Staff{}).
				Where("id = ?", fixture.staffID).
				Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))
		case reservationStaffMutationRemoveAssignment:
			result = tx.Model(&model.StaffClinicAssignment{}).
				Where("id = ?", fixture.assignmentID).
				Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))
		default:
			return fmt.Errorf("unknown competing mutation: %s", mutation)
		}
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("competing mutation changed %d rows", result.RowsAffected)
		}
		return nil
	})
}

func TestReservationWrites_SerializeStaffDeleteAndAssignmentRemoval(t *testing.T) {
	tests := []struct {
		name      string
		operation reservationStaffWriteOperation
		mutation  reservationStaffCompetingMutation
	}{
		{
			name:      "standard create blocks staff delete",
			operation: reservationStaffWriteStandardCreate,
			mutation:  reservationStaffMutationDeleteStaff,
		},
		{
			name:      "standard create blocks assignment removal",
			operation: reservationStaffWriteStandardCreate,
			mutation:  reservationStaffMutationRemoveAssignment,
		},
		{
			name:      "standard update blocks staff delete",
			operation: reservationStaffWriteStandardUpdate,
			mutation:  reservationStaffMutationDeleteStaff,
		},
		{
			name:      "standard update blocks assignment removal",
			operation: reservationStaffWriteStandardUpdate,
			mutation:  reservationStaffMutationRemoveAssignment,
		},
		{
			name:      "admin create blocks staff delete",
			operation: reservationStaffWriteAdminCreate,
			mutation:  reservationStaffMutationDeleteStaff,
		},
		{
			name:      "admin create blocks assignment removal",
			operation: reservationStaffWriteAdminCreate,
			mutation:  reservationStaffMutationRemoveAssignment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupReservationRepoTestDB(t)
			fixture := setupReservationStaffWriteGuardFixture(t, db, tt.operation)
			locked := make(chan struct{})
			release := make(chan struct{})
			guard := &pausingReservationStaffWriteGuard{
				delegate: NewReservationStaffRepository(db, nil),
				locked:   locked,
				release:  release,
			}

			writeDone := make(chan reservationStaffWriteResult, 1)
			go func() {
				reservation, err := runReservationStaffWriteGuardOperation(
					context.Background(),
					db,
					tt.operation,
					&fixture,
					guard,
				)
				writeDone <- reservationStaffWriteResult{
					reservation: reservation,
					err:         err,
				}
			}()

			select {
			case <-locked:
			case result := <-writeDone:
				require.NoError(t, result.err)
				require.Fail(t, "reservation write completed before the guard acquired its locks")
			}

			mutationErr := attemptReservationStaffCompetingMutation(db, tt.mutation, &fixture)
			close(release)
			writeResult := <-writeDone

			require.Error(t, mutationErr)
			assert.True(
				t,
				strings.Contains(mutationErr.Error(), "55P03") ||
					strings.Contains(mutationErr.Error(), "lock timeout"),
				"expected PostgreSQL lock timeout, got: %v",
				mutationErr,
			)
			require.NoError(t, writeResult.err)
			require.NotNil(t, writeResult.reservation)
			require.NotNil(t, writeResult.reservation.DoctorID)
			assert.Equal(t, fixture.staffID, *writeResult.reservation.DoctorID)
			assert.WithinDuration(t, fixture.targetStart, writeResult.reservation.StartTime, 0)

			var persisted int64
			require.NoError(t, db.Model(&model.Reservation{}).
				Where(
					"clinic_id = ? AND doctor_id = ? AND reservation_type_id = ? AND start_time = ?",
					fixture.clinicID,
					fixture.staffID,
					fixture.reservationType,
					fixture.targetStart,
				).
				Count(&persisted).Error)
			assert.Equal(t, int64(1), persisted)

			var activeStaff, activeAssignment int64
			require.NoError(t, db.Model(&model.Staff{}).
				Where("id = ?", fixture.staffID).
				Count(&activeStaff).Error)
			require.NoError(t, db.Model(&model.StaffClinicAssignment{}).
				Where("id = ?", fixture.assignmentID).
				Count(&activeAssignment).Error)
			assert.Equal(t, int64(1), activeStaff)
			assert.Equal(t, int64(1), activeAssignment)
		})
	}
}
