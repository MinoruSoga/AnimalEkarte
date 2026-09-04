package seedlogin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestApplyUpsertsAccountsAndIsIdempotent(t *testing.T) {
	db := setupLoginSeedDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx := context.Background()

	applied, err := Apply(ctx, sqlDB)
	require.NoError(t, err)
	require.Equal(t, 40, applied)

	assertLoginCatalog(t, db, SharedPassword)

	applied, err = Apply(ctx, sqlDB)
	require.NoError(t, err)
	require.Equal(t, 40, applied)
	assertLoginCatalog(t, db, SharedPassword)
}

func TestApplyFailsWhenClinicMissing(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.StaffPermissionGroup{},
	))
	testdb.Truncate(t, db,
		"staff_permission_groups",
		"staff_clinic_assignments",
		"staffs",
		"accounts",
		"permission_groups",
		"clinics",
		"companies",
	)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = Apply(context.Background(), sqlDB)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "002_master required")
}

func TestApplyFailsWhenStaffLinkedToDifferentAccount(t *testing.T) {
	db := setupLoginSeedDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx := context.Background()

	_, err = Apply(ctx, sqlDB)
	require.NoError(t, err)

	spec := Catalog()[0]
	other := model.Account{
		Email:        "other-login-seed@example.test",
		PasswordHash: "not-a-real-hash-value",
		IsActive:     true,
	}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", spec.StaffID).Update("account_id", other.ID).Error)

	_, err = Apply(ctx, sqlDB)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "linked to a different account")
}

func setupLoginSeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.PermissionGroup{},
		&model.StaffPermissionGroup{},
	))
	testdb.Truncate(t, db,
		"staff_permission_groups",
		"staff_clinic_assignments",
		"staffs",
		"accounts",
		"permission_groups",
		"clinics",
		"companies",
	)

	company := &model.Company{Name: "login-seed-test"}
	require.NoError(t, db.Create(company).Error)
	for clinicID := uint64(1); clinicID <= 4; clinicID++ {
		clinic := &model.Clinic{ID: clinicID, CompanyID: company.ID, Name: "clinic"}
		require.NoError(t, db.Clauses(clause.OnConflict{DoNothing: true}).Create(clinic).Error)
		group := &model.PermissionGroup{ClinicID: clinicID, Name: PermissionGroupName, IsActive: true}
		require.NoError(t, db.Create(group).Error)
	}
	return db
}

func assertLoginCatalog(t *testing.T, db *gorm.DB, password string) {
	t.Helper()
	for _, spec := range Catalog() {
		var staff model.Staff
		require.NoError(t, db.Unscoped().First(&staff, spec.StaffID).Error)
		assert.True(t, staff.IsActive)
		assert.False(t, staff.DeletedAt.Valid)
		require.NotNil(t, staff.AccountID)

		var account model.Account
		require.NoError(t, db.Unscoped().First(&account, *staff.AccountID).Error)
		assert.Equal(t, spec.Email, account.Email)
		assert.True(t, account.IsActive)
		assert.False(t, account.IsSystemAdmin)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)))

		var assignment model.StaffClinicAssignment
		require.NoError(t, db.Where("staff_id = ? AND clinic_id = ?", spec.StaffID, spec.ClinicID).First(&assignment).Error)
		assert.True(t, assignment.IsMain)

		var linked int64
		require.NoError(t, db.Model(&model.StaffPermissionGroup{}).
			Where("staff_id = ?", spec.StaffID).
			Count(&linked).Error)
		assert.Equal(t, int64(1), linked)
	}
}
