package auth_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// countingGormLogger records every SQL statement issued through its session so
// tests can assert the exact DB query count of request-time auth resolution.
type countingGormLogger struct {
	logger.Interface
	statements []string
}

func (l *countingGormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	sql, rows := fc()
	l.statements = append(l.statements, sql)
	l.Interface.Trace(ctx, begin, func() (string, int64) { return sql, rows }, err)
}

func (l *countingGormLogger) reset() {
	l.statements = nil
}

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
	testdb.Truncate(
		t,
		db,
		"staff_clinic_assignments",
		"staffs",
		"accounts",
		"clinics",
		"companies",
	)

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

	statements := &countingGormLogger{Interface: db.Logger}
	countedDB := db.Session(&gorm.Session{Logger: statements})
	resolver := authdomain.NewCurrentAccessResolverWithClinics(
		authdomain.NewCurrentAccessStaffReader(countedDB),
		authdomain.NewAccountService(authdomain.NewAccountRepository(countedDB)),
		staff.NewStaffClinicAssignmentService(
			staff.NewStaffClinicAssignmentRepository(countedDB),
		),
		clinic.NewService(
			clinic.NewClinicRepository(countedDB),
			nil,
			nil,
		),
	)

	statements.reset()
	access, err := resolver.Resolve(context.Background(), staffRow.ID)

	require.NoError(t, err)
	assert.Equal(t, []uint64{activeClinic.ID}, access.ClinicIDs)
	assert.Equal(
		t,
		strconv.FormatUint(activeClinic.ID, 10),
		access.MainClinicID,
	)
	require.Len(t, statements.statements, 1,
		"non-admin Resolve must issue exactly one DB query")
	assert.Contains(t, statements.statements[0], "LEFT JOIN clinics",
		"the single graph query must fold in clinics.is_active")

	require.NoError(t, db.Model(activeClinic).
		Update("is_active", false).Error)
	statements.reset()
	access, err = resolver.Resolve(context.Background(), staffRow.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Nil(t, access)
	assert.Len(t, statements.statements, 1,
		"non-admin Resolve must stay a single query even when it fails closed")
}
