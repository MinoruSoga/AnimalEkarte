package auth_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestCurrentAccessStaffReaderDB_ReadsIdentityRow(t *testing.T) {
	db, _ := setupPermissionAuditRollbackDB(t)
	reader := authdomain.NewCurrentAccessStaffReader(db)

	staff, err := reader.FindCurrentAccessStaff(context.Background(), 17)

	require.NoError(t, err)
	require.NotNil(t, staff)
	assert.Equal(t, uint64(17), staff.ID)
	assert.True(t, staff.IsActive)
	assert.False(t, staff.IsDeleted)

	missing, err := reader.FindCurrentAccessStaff(
		context.Background(),
		999,
	)
	require.Error(t, err)
	assert.Nil(t, missing)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestCurrentAccessResolverDB_RegularStaffUsesOnlyActiveClinicInventory(
	t *testing.T,
) {
	db, _ := setupPermissionAuditRollbackDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Account{},
		&model.StaffClinicAssignment{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			staff_clinic_assignments,
			staffs,
			accounts,
			clinics,
			companies
		CASCADE
	`).Error)

	company := &model.Company{Name: "current access inventory company"}
	require.NoError(t, db.Create(company).Error)
	inactiveClinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "inactive assigned clinic",
		IsActive:  true,
	}
	activeClinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      "active assigned clinic",
		IsActive:  true,
	}
	require.NoError(t, db.Create(inactiveClinic).Error)
	require.NoError(t, db.Create(activeClinic).Error)
	require.NoError(t, db.Model(inactiveClinic).
		Update("is_active", false).Error)

	account := &model.Account{
		Email:        "current-access-inventory@example.test",
		PasswordHash: "not-a-real-password-hash",
		IsActive:     true,
	}
	require.NoError(t, db.Create(account).Error)
	accountID := account.ID
	staffRow := &model.Staff{
		ClinicID:  inactiveClinic.ID,
		AccountID: &accountID,
		Name:      "current access inventory staff",
		IsActive:  true,
	}
	require.NoError(t, db.Create(staffRow).Error)
	require.NoError(t, db.Create([]model.StaffClinicAssignment{
		{
			StaffID:  staffRow.ID,
			ClinicID: inactiveClinic.ID,
			IsMain:   true,
		},
		{
			StaffID:  staffRow.ID,
			ClinicID: activeClinic.ID,
		},
	}).Error)

	resolver := authdomain.NewCurrentAccessResolverWithClinics(
		authdomain.NewCurrentAccessStaffReader(db),
		authdomain.NewAccountService(authdomain.NewAccountRepository(db)),
		staff.NewStaffClinicAssignmentService(
			staff.NewStaffClinicAssignmentRepository(db),
		),
		clinic.NewClinicService(
			clinic.NewClinicRepository(db),
			nil,
			nil,
		),
	)

	access, err := resolver.Resolve(context.Background(), staffRow.ID)

	require.NoError(t, err)
	assert.Equal(t, []uint64{activeClinic.ID}, access.ClinicIDs)
	assert.Equal(
		t,
		strconv.FormatUint(activeClinic.ID, 10),
		access.MainClinicID,
	)

	require.NoError(t, db.Model(activeClinic).
		Update("is_active", false).Error)
	access, err = resolver.Resolve(context.Background(), staffRow.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Nil(t, access)
}
