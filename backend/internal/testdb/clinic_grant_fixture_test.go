package testdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestSeedDualClinicGrantFixture_SurvivesLaggingClinicSequence(t *testing.T) {
	db := SetupTestDB(t)
	require.NoError(t, EnsureAutoMigrated(db, &model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{}))

	ctx := context.Background()
	company := &model.Company{Name: "seq-lag company"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	var maxID uint64
	require.NoError(t, db.WithContext(ctx).Raw(`SELECT COALESCE(MAX(id), 0) FROM clinics`).Scan(&maxID).Error)
	explicitID := maxID + 100000
	explicit := &model.Clinic{ID: explicitID, CompanyID: company.ID, Name: "seq-lag clinic", IsActive: true}
	require.NoError(t, db.WithContext(ctx).Create(explicit).Error)
	// Force sequence behind MAX(id) the way explicit-ID inserts do.
	require.NoError(t, db.Exec(`SELECT setval(pg_get_serial_sequence('clinics', 'id'), 1, true)`).Error)

	fx := SeedDualClinicGrantFixture(t, db, "seq-lag dual")
	require.NotEqual(t, uint64(0), fx.ClinicA)
	require.NotEqual(t, uint64(0), fx.ClinicB)
	require.NotEqual(t, explicit.ID, fx.ClinicA)
	require.NotEqual(t, explicit.ID, fx.ClinicB)
}
