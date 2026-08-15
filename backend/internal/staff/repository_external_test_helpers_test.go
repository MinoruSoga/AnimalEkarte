package staff_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestMain(m *testing.M) {
	code := m.Run()
	testdb.CloseSharedTestDB()
	os.Exit(code)
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.SetupTestDB(t)
}

func ensureAutoMigrated(db *gorm.DB, models ...any) error {
	return testdb.EnsureAutoMigrated(db, models...)
}

func seedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
	t.Helper()
	company := &model.Company{Name: "staff domain repository tests"}
	require.NoError(t, db.WithContext(context.Background()).Create(company).Error)
	for _, clinicID := range clinicIDs {
		clinic := &model.Clinic{
			ID:        clinicID,
			CompanyID: company.ID,
			Name:      fmt.Sprintf("staff domain clinic %d", clinicID),
			IsActive:  true,
		}
		require.NoError(t, db.WithContext(context.Background()).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(clinic).Error)
	}
}

func makeStaffClinicAssignment(t *testing.T, db *gorm.DB, staffID, clinicID uint64) {
	t.Helper()
	assignment := &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}
	require.NoError(t, db.WithContext(context.Background()).Create(assignment).Error)
}

func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
	t.Helper()
	seedClinicsForFK(t, db, clinicID)
	staff := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		StaffType: model.StaffTypeDoctor,
		IsActive:  true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	return staff
}

func makeOccupation(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
) *model.Occupation {
	t.Helper()
	occupation := &model.Occupation{
		ClinicID: clinicID,
		Name:     name,
		IsActive: true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(occupation).Error)
	return occupation
}

func makeStaffWithOccupation(
	t *testing.T,
	db *gorm.DB,
	clinicID, occupationID uint64,
	name string,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID:     clinicID,
		Name:         name,
		StaffType:    model.StaffTypeDoctor,
		OccupationID: &occupationID,
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(staff).Error)
	return staff
}
