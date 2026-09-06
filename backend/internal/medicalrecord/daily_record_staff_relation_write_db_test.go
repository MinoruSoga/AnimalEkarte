package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type dailyRecordStaffRelationDBFixture struct {
	db              *gorm.DB
	clinicA         uint64
	hospitalization *model.Hospitalization
	validStaff      *model.Staff
	foreignStaff    *model.Staff
	unassignedStaff *model.Staff
	service         DailyRecordService
}

func setupDailyRecordStaffRelationDBFixture(t *testing.T) *dailyRecordStaffRelationDBFixture {
	t.Helper()

	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.Hospitalization{},
		&model.DailyRecord{},
		&model.VitalRecord{},
		&model.CareLog{},
		&model.StaffNote{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			care_logs,
			staff_notes,
			vital_records,
			daily_records,
			hospitalizations,
			staff_clinic_assignments,
			staffs,
			pets,
			animal_species,
			owners
		CASCADE
	`).Error)
	// Persistent test databases may still carry the pre-BUG-404 timestamptz type.
	require.NoError(t, db.Exec(`
		ALTER TABLE care_logs
		ALTER COLUMN "time" TYPE time USING "time"::time
	`).Error)
	require.NoError(t, db.Exec(`
		ALTER TABLE staff_notes
		ALTER COLUMN "time" TYPE time USING "time"::time
	`).Error)

	// Other DB tests insert explicit fixture IDs. Keep serial sequences ahead of
	// those rows before this fixture relies on auto-generated company/clinic IDs.
	require.NoError(t, db.Exec(`
		SELECT setval(
			pg_get_serial_sequence('companies', 'id'),
			COALESCE((SELECT MAX(id) + 1 FROM companies), 1),
			false
		)
	`).Error)
	require.NoError(t, db.Exec(`
		SELECT setval(
			pg_get_serial_sequence('clinics', 'id'),
			COALESCE((SELECT MAX(id) + 1 FROM clinics), 1),
			false
		)
	`).Error)

	company := &model.Company{Name: "daily relation DB fixture"}
	require.NoError(t, db.Create(company).Error)
	clinicA := &model.Clinic{CompanyID: company.ID, Name: "daily relation clinic A"}
	clinicB := &model.Clinic{CompanyID: company.ID, Name: "daily relation clinic B"}
	require.NoError(t, db.Create(clinicA).Error)
	require.NoError(t, db.Create(clinicB).Error)

	species := &model.AnimalSpecies{Name: "daily relation species", IsActive: true}
	require.NoError(t, db.Create(species).Error)
	owner := &model.Owner{ClinicID: clinicA.ID, Name: "daily relation owner"}
	require.NoError(t, db.Create(owner).Error)
	pet := &model.Pet{
		ClinicID: clinicA.ID, OwnerID: owner.ID,
		AnimalSpeciesID: species.ID, Name: "daily relation pet",
	}
	require.NoError(t, db.Create(pet).Error)
	hospitalization := &model.Hospitalization{
		ClinicID: clinicA.ID, OwnerID: owner.ID, PetID: pet.ID,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		EndDate:             time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		Status:              model.HospitalizationStatusAdmitted,
	}
	require.NoError(t, db.Create(hospitalization).Error)

	// As with production authorization, the active assignment is authoritative;
	// the valid staff's primary clinic is intentionally the other clinic.
	validStaff := createDailyRelationStaff(t, db, clinicB.ID, "daily assigned staff")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: validStaff.ID, ClinicID: clinicA.ID,
	}).Error)
	foreignStaff := createDailyRelationStaff(t, db, clinicB.ID, "daily foreign staff")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: foreignStaff.ID, ClinicID: clinicB.ID,
	}).Error)
	unassignedStaff := createDailyRelationStaff(t, db, clinicA.ID, "daily unassigned staff")

	service := NewDailyRecordServiceWithRelationValidation(
		NewDailyRecordRepository(db),
		NewHospitalizationRepository(db),
		reservation.NewReservationRepository(db),
		staffdomain.NewRepository(db),
		staffdomain.NewStaffClinicAssignmentRepository(db),
		persistence.NewTransactor(db),
	)
	return &dailyRecordStaffRelationDBFixture{
		db: db, clinicA: clinicA.ID, hospitalization: hospitalization,
		validStaff: validStaff, foreignStaff: foreignStaff,
		unassignedStaff: unassignedStaff, service: service,
	}
}

func createDailyRelationStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID: clinicID, Name: name,
		IsActive: true, StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	return staff
}

func TestDB_DailyRecordServiceRejectsPollutedStaffRelations(t *testing.T) {
	fixture := setupDailyRecordStaffRelationDBFixture(t)
	ctx := context.Background()
	rejectedStaff := []*model.Staff{fixture.foreignStaff, fixture.unassignedStaff}

	t.Run("vital record", func(t *testing.T) {
		date := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
		temperature := 38.1
		for _, staff := range rejectedStaff {
			got, err := fixture.service.AddVitalRecord(
				ctx,
				fixture.clinicA,
				fixture.hospitalization.ID,
				date,
				&CreateVitalRecordInput{
					Time:        time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC),
					Temperature: &temperature, StaffID: &staff.ID,
				},
			)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, countDailyRelationRows(t, fixture.db, &model.VitalRecord{}))
			assert.Zero(
				t,
				countDailyRelationRows(
					t,
					fixture.db,
					&model.DailyRecord{},
					"hospitalization_id = ? AND date = ?",
					fixture.hospitalization.ID,
					date,
				),
				"rejected staff must not leave a parent daily record",
			)
		}

		got, err := fixture.service.AddVitalRecord(
			ctx,
			fixture.clinicA,
			fixture.hospitalization.ID,
			date,
			&CreateVitalRecordInput{
				Time:        time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC),
				Temperature: &temperature, StaffID: &fixture.validStaff.ID,
			},
		)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.VitalRecords, 1)
		require.NotNil(t, got.VitalRecords[0].StaffID)
		assert.Equal(t, fixture.validStaff.ID, *got.VitalRecords[0].StaffID)
		assert.EqualValues(t, 1, countDailyRelationRows(t, fixture.db, &model.VitalRecord{}))
	})

	t.Run("care log", func(t *testing.T) {
		date := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
		for _, staff := range rejectedStaff {
			got, err := fixture.service.AddCareLog(
				ctx,
				fixture.clinicA,
				fixture.hospitalization.ID,
				date,
				&CreateCareLogInput{
					Time: "09:00:00", Type: string(model.CareLogTypeFood),
					Status: string(model.CareLogStatusCompleted), StaffID: &staff.ID,
				},
			)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, countDailyRelationRows(t, fixture.db, &model.CareLog{}))
			assert.Zero(
				t,
				countDailyRelationRows(
					t,
					fixture.db,
					&model.DailyRecord{},
					"hospitalization_id = ? AND date = ?",
					fixture.hospitalization.ID,
					date,
				),
				"rejected staff must not leave a parent daily record",
			)
		}

		got, err := fixture.service.AddCareLog(
			ctx,
			fixture.clinicA,
			fixture.hospitalization.ID,
			date,
			&CreateCareLogInput{
				Time: "10:00:00", Type: string(model.CareLogTypeFood),
				Status:  string(model.CareLogStatusCompleted),
				StaffID: &fixture.validStaff.ID,
			},
		)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.CareLogs, 1)
		require.NotNil(t, got.CareLogs[0].StaffID)
		assert.Equal(t, fixture.validStaff.ID, *got.CareLogs[0].StaffID)
		assert.EqualValues(t, 1, countDailyRelationRows(t, fixture.db, &model.CareLog{}))
	})

	t.Run("staff note", func(t *testing.T) {
		date := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)
		for _, staff := range rejectedStaff {
			got, err := fixture.service.AddStaffNote(
				ctx,
				fixture.clinicA,
				fixture.hospitalization.ID,
				date,
				&CreateStaffNoteInput{
					Time: "09:00:00", Content: "rejected", StaffID: &staff.ID,
				},
			)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, countDailyRelationRows(t, fixture.db, &model.StaffNote{}))
			assert.Zero(
				t,
				countDailyRelationRows(
					t,
					fixture.db,
					&model.DailyRecord{},
					"hospitalization_id = ? AND date = ?",
					fixture.hospitalization.ID,
					date,
				),
				"rejected staff must not leave a parent daily record",
			)
		}

		got, err := fixture.service.AddStaffNote(
			ctx,
			fixture.clinicA,
			fixture.hospitalization.ID,
			date,
			&CreateStaffNoteInput{
				Time: "10:00:00", Content: "accepted", StaffID: &fixture.validStaff.ID,
			},
		)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.StaffNotes, 1)
		require.NotNil(t, got.StaffNotes[0].StaffID)
		assert.Equal(t, fixture.validStaff.ID, *got.StaffNotes[0].StaffID)
		assert.EqualValues(t, 1, countDailyRelationRows(t, fixture.db, &model.StaffNote{}))
	})
}

func countDailyRelationRows(
	t *testing.T,
	db *gorm.DB,
	target any,
	where ...any,
) int64 {
	t.Helper()
	query := db.Model(target)
	if len(where) > 0 {
		query = query.Where(where[0], where[1:]...)
	}
	var count int64
	require.NoError(t, query.Count(&count).Error)
	return count
}
