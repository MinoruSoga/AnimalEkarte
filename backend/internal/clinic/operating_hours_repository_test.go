package clinic

// operating_hours_repository_test.go — real-DB test for the clinic_operating_hours read port.
// Applies 001 to a disposable database (testdb.SetupIsolated001TestDB); the
// former 008 migration body is part of 001 since the 2026-10-23 consolidation,
// like internal/model/clinic_operating_hours_realdb_test.go. Skipped under -short.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupOperatingHoursRepositoryDB(t *testing.T) *gorm.DB {
	t.Helper()
	// 001_init.sql carries the former 008 DDL (2026-10-23 consolidation).
	db := testdb.SetupIsolated001TestDB(t)
	return db
}

func seedOperatingHoursClinic(t *testing.T, db *gorm.DB, name string) uint64 {
	t.Helper()
	var companyID, clinicID uint64
	require.NoError(t, db.Raw(`INSERT INTO companies (name) VALUES (?) RETURNING id`, name).Scan(&companyID).Error)
	require.NoError(t, db.Raw(`INSERT INTO clinics (company_id, name) VALUES (?, ?) RETURNING id`, companyID, name).Scan(&clinicID).Error)
	return clinicID
}

func insertOperatingHoursRow(t *testing.T, db *gorm.DB, clinicID uint64, closedDates string, holidayClosed bool) {
	t.Helper()
	require.NoError(t, db.Exec(`
		INSERT INTO clinic_operating_hours
			(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
		VALUES (?, 'clinic-operating-hours/v1', ?::jsonb, ?::jsonb, ?)`,
		clinicID, operatingHoursSevenClosedDays, closedDates, holidayClosed).Error)
}

func TestOperatingHoursRepository_FindByClinicID(t *testing.T) {
	db := setupOperatingHoursRepositoryDB(t)
	repo := NewOperatingHoursRepository(db)
	ctx := context.Background()

	clinicA := seedOperatingHoursClinic(t, db, "oh-repo-a")
	clinicB := seedOperatingHoursClinic(t, db, "oh-repo-b")
	clinicC := seedOperatingHoursClinic(t, db, "oh-repo-c")

	t.Run("absent row is found=false without an error and without a complemented row", func(t *testing.T) {
		row, found, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Equal(t, model.ClinicOperatingHours{}, row)
	})

	insertOperatingHoursRow(t, db, clinicA, `[{"date":"2026-11-03","reason":"臨時休診"}]`, true)
	insertOperatingHoursRow(t, db, clinicB, `[]`, false)

	t.Run("stored row is returned with its stored version and clinic scope", func(t *testing.T) {
		row, found, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, clinicA, row.ClinicID)
		assert.EqualValues(t, 1, row.Version)
		assert.Equal(t, model.ClinicOperatingHoursSchemaV1, row.SchemaVersion)
		assert.True(t, row.NationalHolidayClosed)

		settings, err := LoadOperatingHoursSettings(row)
		require.NoError(t, err)
		dates, err := settings.ClosedDates()
		require.NoError(t, err)
		require.Len(t, dates, 1)
		assert.Equal(t, "臨時休診", dates[0].Reason)
	})

	t.Run("another clinic's row is never returned", func(t *testing.T) {
		row, found, err := repo.FindByClinicID(ctx, clinicB)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, clinicB, row.ClinicID)
		assert.False(t, row.NationalHolidayClosed)

		_, found, err = repo.FindByClinicID(ctx, clinicC)
		require.NoError(t, err)
		assert.False(t, found, "a clinic without its own row is unconfigured even though other clinics are configured")
	})

	t.Run("database failure propagates and is never reported as absent", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		_, found, err := repo.FindByClinicID(canceled, clinicA)
		require.Error(t, err)
		assert.False(t, found)
	})

	t.Run("read joins the caller's ambient transaction", func(t *testing.T) {
		rollbackProbe := errors.New("rollback operating hours ambient probe")
		txCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		err := persistence.NewTransactor(db).WithTx(txCtx, func(ambient context.Context) error {
			require.NoError(t, persistence.DBOrTx(ambient, db).Exec(`
				INSERT INTO clinic_operating_hours
					(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
				VALUES (?, 'clinic-operating-hours/v1', ?::jsonb, '[]'::jsonb, false)`,
				clinicC, operatingHoursSevenClosedDays).Error)
			_, found, readErr := repo.FindByClinicID(ambient, clinicC)
			require.NoError(t, readErr)
			assert.True(t, found, "uncommitted ambient write must be visible to the participating read")
			return rollbackProbe
		})
		require.ErrorIs(t, err, rollbackProbe)

		_, found, err := repo.FindByClinicID(ctx, clinicC)
		require.NoError(t, err)
		assert.False(t, found, "rolled back")
	})
}

func TestOperatingHoursReadService_RealDBCorruptJSONBIsDistinctFromAbsent(t *testing.T) {
	db := setupOperatingHoursRepositoryDB(t)
	clinicID := seedOperatingHoursClinic(t, db, "oh-repo-corrupt")
	// jsonb satisfies the 008 CHECKs (array, 7 elements) yet is not a contract-shaped weekly_hours.
	require.NoError(t, db.Exec(`
		INSERT INTO clinic_operating_hours
			(clinic_id, schema_version, weekly_hours, closed_dates, national_holiday_closed)
		VALUES (?, 'clinic-operating-hours/v1', '[1,2,3,4,5,6,7]'::jsonb, '[]'::jsonb, false)`, clinicID).Error)

	service := NewOperatingHoursReadService(NewOperatingHoursRepository(db))
	_, err := service.Get(context.Background(), clinicID)
	requireCorrupt(t, err)
}
