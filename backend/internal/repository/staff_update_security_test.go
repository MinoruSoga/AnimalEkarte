package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

func setupStaffUpdateSecurityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
		&model.Account{},
		&model.Occupation{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE staff_clinic_assignments, staffs, occupations, accounts, clinics, companies CASCADE",
	).Error)
	return db
}

func makeStaffUpdateClinic(t *testing.T, db *gorm.DB, companyID uint64, name string) *model.Clinic {
	t.Helper()
	clinic := &model.Clinic{CompanyID: companyID, Name: name, IsActive: true}
	require.NoError(t, db.Create(clinic).Error)
	return clinic
}

func makeAccountStaffForUpdate(
	t *testing.T,
	db *gorm.DB,
	clinic *model.Clinic,
	email string,
	passwordHash string,
) (*model.Account, *model.Staff) {
	t.Helper()
	account := &model.Account{Email: email, PasswordHash: passwordHash, IsActive: true}
	require.NoError(t, db.Create(account).Error)
	staff := &model.Staff{
		ClinicID:  clinic.ID,
		AccountID: &account.ID,
		Name:      "更新前スタッフ",
		IsActive:  true,
		StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinic.ID,
		IsMain:   true,
	}).Error)
	return account, staff
}

func newStaffUpdateServiceForDB(
	db *gorm.DB,
	accounts staffdomain.StaffAccountStore,
) staffdomain.StaffService {
	return staffdomain.NewStaffServiceWithCredentialAudit(
		NewStaffRepository(db),
		accounts,
		NewStaffClinicAssignmentRepository(db),
		nil,
		nil,
		nil,
		nil,
		NewOccupationRepository(db),
		nil,
		NewTransactor(db),
		noopStaffUpdateCredentialAuditTxLogger{},
	)
}

type noopStaffUpdateCredentialAuditTxLogger struct{}

func (noopStaffUpdateCredentialAuditTxLogger) LogEntryTx(
	context.Context,
	staffdomain.CredentialAuditEntry,
) error {
	return nil
}

func staffUpdateCredentialAudit(
	clinicID, targetStaffID uint64,
) *staffdomain.CredentialMutationAudit {
	return &staffdomain.CredentialMutationAudit{
		ClinicID:      clinicID,
		ActorStaffID:  999,
		TargetStaffID: targetStaffID,
	}
}

type failAfterStaffAccountUpdate struct {
	staffdomain.StaffAccountStore
	err error
}

func (s failAfterStaffAccountUpdate) UpdatePasswordHash(
	ctx context.Context,
	id uint64,
	newHash string,
	updatedAt time.Time,
) error {
	if err := s.StaffAccountStore.UpdatePasswordHash(
		ctx,
		id,
		newHash,
		updatedAt,
	); err != nil {
		return err
	}
	return s.err
}

func TestStaffServiceUpdatePasswordOnlyRejectsCrossClinicTargetDatabase(t *testing.T) {
	db := setupStaffUpdateSecurityDB(t)
	company := &model.Company{Name: "staff update tenant company"}
	require.NoError(t, db.Create(company).Error)
	clinicA := makeStaffUpdateClinic(t, db, company.ID, "医院A")
	clinicB := makeStaffUpdateClinic(t, db, company.ID, "医院B")
	account, staff := makeAccountStaffForUpdate(
		t,
		db,
		clinicB,
		"staff-cross-clinic@example.com",
		"unchanged-hash",
	)
	service := newStaffUpdateServiceForDB(db, NewAccountRepository(db))
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinicA.ID,
		staff.ID,
		&staffdomain.UpdateStaffInput{
			Password:            &password,
			AuthorizedClinicIDs: []uint64{clinicA.ID},
			CredentialAudit:     staffUpdateCredentialAudit(clinicA.ID, staff.ID),
		},
	)

	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "unexpected error: %v", err)
	assert.Nil(t, updated)
	var reloaded model.Account
	require.NoError(t, db.First(&reloaded, account.ID).Error)
	assert.Equal(t, "unchanged-hash", reloaded.PasswordHash)
}

func TestStaffServiceUpdateRollsBackProfileAndPasswordTogetherDatabase(t *testing.T) {
	db := setupStaffUpdateSecurityDB(t)
	company := &model.Company{Name: "staff update rollback company"}
	require.NoError(t, db.Create(company).Error)
	clinic := makeStaffUpdateClinic(t, db, company.ID, "医院")
	account, staff := makeAccountStaffForUpdate(
		t,
		db,
		clinic,
		"staff-rollback@example.com",
		"old-hash",
	)
	sentinel := errors.New("fail after account update")
	accountStore := failAfterStaffAccountUpdate{
		StaffAccountStore: NewAccountRepository(db),
		err:               sentinel,
	}
	service := newStaffUpdateServiceForDB(db, accountStore)
	name := "更新後スタッフ"
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinic.ID,
		staff.ID,
		&staffdomain.UpdateStaffInput{
			Name:                &name,
			Password:            &password,
			AuthorizedClinicIDs: []uint64{clinic.ID},
			CredentialAudit:     staffUpdateCredentialAudit(clinic.ID, staff.ID),
		},
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Nil(t, updated)
	var reloadedStaff model.Staff
	require.NoError(t, db.First(&reloadedStaff, staff.ID).Error)
	assert.Equal(t, "更新前スタッフ", reloadedStaff.Name)
	var reloadedAccount model.Account
	require.NoError(t, db.First(&reloadedAccount, account.ID).Error)
	assert.Equal(t, "old-hash", reloadedAccount.PasswordHash)
}

func TestStaffServiceUpdateCommitsProfileAndPasswordTogetherDatabase(t *testing.T) {
	db := setupStaffUpdateSecurityDB(t)
	company := &model.Company{Name: "staff update commit company"}
	require.NoError(t, db.Create(company).Error)
	clinic := makeStaffUpdateClinic(t, db, company.ID, "医院")
	account, staff := makeAccountStaffForUpdate(
		t,
		db,
		clinic,
		"staff-commit@example.com",
		"old-hash",
	)
	service := newStaffUpdateServiceForDB(db, NewAccountRepository(db))
	name := "更新後スタッフ"
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinic.ID,
		staff.ID,
		&staffdomain.UpdateStaffInput{
			Name:                &name,
			Password:            &password,
			AuthorizedClinicIDs: []uint64{clinic.ID},
			CredentialAudit:     staffUpdateCredentialAudit(clinic.ID, staff.ID),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, name, updated.Name)
	var reloadedAccount model.Account
	require.NoError(t, db.First(&reloadedAccount, account.ID).Error)
	require.NoError(t, bcrypt.CompareHashAndPassword(
		[]byte(reloadedAccount.PasswordHash),
		[]byte(password),
	))
}
