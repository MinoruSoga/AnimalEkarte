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

func TestApplyReplacesStaleGeneralMembership(t *testing.T) {
	db := setupLoginSeedDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx := context.Background()

	_, err = Apply(ctx, sqlDB)
	require.NoError(t, err)

	spec := Catalog()[0]
	require.Equal(t, PermissionGroupExecutive, spec.PermissionGroupName)

	var general model.PermissionGroup
	require.NoError(t, db.Where("clinic_id = ? AND name = ?", spec.ClinicID, PermissionGroupGeneral).First(&general).Error)
	require.NoError(t, db.Where("staff_id = ?", spec.StaffID).Delete(&model.StaffPermissionGroup{}).Error)
	require.NoError(t, db.Create(&model.StaffPermissionGroup{StaffID: spec.StaffID, GroupID: general.ID}).Error)

	_, err = Apply(ctx, sqlDB)
	require.NoError(t, err)

	var names []string
	require.NoError(t, db.Raw(`
		SELECT pg.name
		  FROM permission_groups pg
		  JOIN staff_permission_groups spg ON spg.group_id = pg.id
		 WHERE spg.staff_id = ?
		 ORDER BY pg.name
	`, spec.StaffID).Scan(&names).Error)
	require.Len(t, names, 4)
	for _, name := range names {
		assert.Equal(t, PermissionGroupExecutive, name)
	}
}

func TestApplyFailsWhenExecutiveGroupMissing(t *testing.T) {
	db := setupLoginSeedDB(t)
	require.NoError(t, db.Where("name = ?", PermissionGroupExecutive).Delete(&model.PermissionGroup{}).Error)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = Apply(context.Background(), sqlDB)
	require.Error(t, err)
	assert.Contains(t, err.Error(), PermissionGroupExecutive)
}

func TestApplyAssignsHayashiToAllCatalogClinics(t *testing.T) {
	db := setupLoginSeedDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)

	_, err = Apply(context.Background(), sqlDB)
	require.NoError(t, err)

	hayashi := Catalog()[0]
	require.True(t, hayashi.AssignAllCatalogClinics)
	assertAssignedClinics(t, db, hayashi)

	general := Catalog()[1]
	require.False(t, general.AssignAllCatalogClinics)
	assertAssignedClinics(t, db, general)
}

func TestApplyRetiresExtraClinicAssignment(t *testing.T) {
	db := setupLoginSeedDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx := context.Background()

	_, err = Apply(ctx, sqlDB)
	require.NoError(t, err)

	var company model.Company
	require.NoError(t, db.First(&company).Error)
	extraClinic := model.Clinic{ID: 99, CompanyID: company.ID, Name: "extra"}
	require.NoError(t, db.Create(&extraClinic).Error)

	hayashi := Catalog()[0]
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  hayashi.StaffID,
		ClinicID: extraClinic.ID,
		IsMain:   false,
	}).Error)

	_, err = Apply(ctx, sqlDB)
	require.NoError(t, err)

	var extra model.StaffClinicAssignment
	require.NoError(t, db.Unscoped().Where("staff_id = ? AND clinic_id = ?", hayashi.StaffID, extraClinic.ID).First(&extra).Error)
	assert.True(t, extra.DeletedAt.Valid)
	assertAssignedClinics(t, db, hayashi)
}

func TestApplyUpsertsOperatorSystemAdminFromEnv(t *testing.T) {
	db := setupLoginSeedDB(t)
	t.Setenv(operatorEnvEmail, "stg-operator@example.test")
	t.Setenv(operatorEnvName, "Operator")
	t.Setenv(operatorEnvPassword, "OperatorPass1")
	sqlDB, err := db.DB()
	require.NoError(t, err)
	ctx := context.Background()

	applied, err := Apply(ctx, sqlDB)
	require.NoError(t, err)
	require.Equal(t, 40, applied)

	var account model.Account
	require.NoError(t, db.Where("email = ?", "stg-operator@example.test").First(&account).Error)
	assert.True(t, account.IsSystemAdmin)
	assert.True(t, account.IsActive)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte("OperatorPass1")))
	assert.False(t, AcceptSharedPassword("staging", account.Email, SharedPassword))

	var staff model.Staff
	require.NoError(t, db.Where("account_id = ?", account.ID).First(&staff).Error)
	assert.Equal(t, "Operator", staff.Name)
	assert.False(t, staff.ReservationVisible)
	assertAssignedClinics(t, db, AccountSpec{
		StaffID:                 staff.ID,
		ClinicID:                catalogClinicIDs()[0],
		PermissionGroupName:     PermissionGroupExecutive,
		AssignAllCatalogClinics: true,
	})

	applied, err = Apply(ctx, sqlDB)
	require.NoError(t, err)
	require.Equal(t, 40, applied)
	var again model.Account
	require.NoError(t, db.Where("email = ?", "stg-operator@example.test").First(&again).Error)
	assert.Equal(t, account.ID, again.ID)
	assert.True(t, again.IsSystemAdmin)
}

func TestApplyRejectsOperatorCatalogEmail(t *testing.T) {
	db := setupLoginSeedDB(t)
	catalog := Catalog()[0]
	t.Setenv(operatorEnvEmail, catalog.Email)
	t.Setenv(operatorEnvName, "Operator")
	t.Setenv(operatorEnvPassword, "OperatorPass1")
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = Apply(context.Background(), sqlDB)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "demo catalog")
}

func setupLoginSeedDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv(operatorEnvEmail, "")
	t.Setenv(operatorEnvName, "")
	t.Setenv(operatorEnvPassword, "")
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
		for _, name := range []string{PermissionGroupExecutive, PermissionGroupGeneral} {
			group := &model.PermissionGroup{ClinicID: clinicID, Name: name, IsActive: true}
			require.NoError(t, db.Create(group).Error)
		}
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
		assertAssignedClinics(t, db, spec)

		var groupRows []struct {
			ClinicID uint64
			Name     string
		}
		require.NoError(t, db.Raw(`
			SELECT pg.clinic_id, pg.name
			  FROM permission_groups pg
			  JOIN staff_permission_groups spg ON spg.group_id = pg.id
			 WHERE spg.staff_id = ?
			 ORDER BY pg.clinic_id
		`, spec.StaffID).Scan(&groupRows).Error)
		wantClinics := assignmentClinicIDs(spec)
		require.Len(t, groupRows, len(wantClinics))
		gotClinics := make([]uint64, 0, len(groupRows))
		for _, row := range groupRows {
			assert.Equal(t, spec.PermissionGroupName, row.Name)
			gotClinics = append(gotClinics, row.ClinicID)
		}
		assert.Equal(t, wantClinics, gotClinics)
	}
}

func assertAssignedClinics(t *testing.T, db *gorm.DB, spec AccountSpec) {
	t.Helper()
	var assignments []model.StaffClinicAssignment
	require.NoError(t, db.Where("staff_id = ?", spec.StaffID).Order("clinic_id").Find(&assignments).Error)
	want := assignmentClinicIDs(spec)
	require.Len(t, assignments, len(want))
	mainCount := 0
	got := make([]uint64, 0, len(assignments))
	for _, assignment := range assignments {
		got = append(got, assignment.ClinicID)
		if assignment.IsMain {
			mainCount++
			assert.Equal(t, spec.ClinicID, assignment.ClinicID)
		}
	}
	assert.Equal(t, 1, mainCount)
	assert.Equal(t, want, got)
}
