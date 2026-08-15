package staff

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupStaffUpdateSecurityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
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

func newStaffUpdateAccountStore(db *gorm.DB) StaffAccountStore {
	return staffResetInvalidationAccountStore{
		AccountRepository: auth.NewAccountRepository(db),
		resetTokens:       auth.NewPasswordResetTokenRepository(db),
	}
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
	accounts StaffAccountStore,
) StaffService {
	return NewStaffServiceWithCredentialAudit(
		NewStaffRepository(db),
		accounts,
		NewStaffClinicAssignmentRepository(db),
		nil,
		nil,
		nil,
		nil,
		NewOccupationRepository(db),
		nil,
		persistence.NewTransactor(db),
		noopStaffUpdateCredentialAuditTxLogger{},
	)
}

type noopStaffUpdateCredentialAuditTxLogger struct{}

func (noopStaffUpdateCredentialAuditTxLogger) LogEntryTx(
	context.Context,
	CredentialAuditEntry,
) error {
	return nil
}

func staffUpdateCredentialAudit(
	clinicID, targetStaffID uint64,
) *CredentialMutationAudit {
	return &CredentialMutationAudit{
		ClinicID:      clinicID,
		ActorStaffID:  999,
		TargetStaffID: targetStaffID,
	}
}

type failAfterStaffAccountUpdate struct {
	StaffAccountStore
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
	service := newStaffUpdateServiceForDB(db, newStaffUpdateAccountStore(db))
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinicA.ID,
		staff.ID,
		&UpdateStaffInput{
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
		StaffAccountStore: newStaffUpdateAccountStore(db),
		err:               sentinel,
	}
	service := newStaffUpdateServiceForDB(db, accountStore)
	name := "更新後スタッフ"
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinic.ID,
		staff.ID,
		&UpdateStaffInput{
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
	service := newStaffUpdateServiceForDB(db, newStaffUpdateAccountStore(db))
	name := "更新後スタッフ"
	password := "newpassword1"

	updated, err := service.Update(
		context.Background(),
		clinic.ID,
		staff.ID,
		&UpdateStaffInput{
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

func TestHandlerUpdateStaffPasswordReplacementPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		body              string
		checkerAllows     bool
		useLegacyHandler  bool
		wantStatus        int
		wantUpdateCalls   int
		wantCheckerCalls  int
		wantPasswordInput bool
	}{
		{
			name:             "denies password and profile update without permission",
			body:             `{"name":"更新後スタッフ","password":"newPassw0rd"}`,
			wantStatus:       http.StatusForbidden,
			wantCheckerCalls: 1,
		},
		{
			name:              "allows password update with permission",
			body:              `{"password":"newPassw0rd"}`,
			checkerAllows:     true,
			wantStatus:        http.StatusOK,
			wantUpdateCalls:   1,
			wantCheckerCalls:  1,
			wantPasswordInput: true,
		},
		{
			name:             "allows profile update without password and checker",
			body:             `{"name":"更新後スタッフ"}`,
			useLegacyHandler: true,
			wantStatus:       http.StatusOK,
			wantUpdateCalls:  1,
		},
		{
			name:             "rejects password update when checker is nil",
			body:             `{"password":"newPassw0rd"}`,
			useLegacyHandler: true,
			wantStatus:       http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalls := 0
			checkerCalls := 0
			service := &credentialAuditStaffService{
				result: &model.Staff{ID: 29, ClinicID: 23, Name: "Updated Staff"},
				calls:  &updateCalls,
			}
			checker := func(_ *gin.Context, resource, action string) bool {
				checkerCalls++
				assert.Equal(t, string(model.ResourceMasterPermission), resource)
				assert.Equal(t, "edit", action)
				return tt.checkerAllows
			}

			var handler *Handler
			if tt.useLegacyHandler {
				handler = NewHandler(service, nil, nil, nil, nil, nil)
			} else {
				handler = NewHandlerWithPermissionChecker(
					service,
					nil,
					nil,
					nil,
					nil,
					nil,
					checker,
				)
			}

			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				http.MethodPatch,
				"/api/v1/masters/staffs/29",
				bytes.NewBufferString(tt.body),
			)
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: "29"}}
			c.Set("clinic_id", "23")
			c.Set("clinic_ids", []uint64{23})
			c.Set("is_system_admin", false)
			c.Set("user_id", "17")
			handler.UpdateStaff(c)

			assert.Equal(t, tt.wantStatus, response.Code, response.Body.String())
			assert.Equal(t, tt.wantUpdateCalls, updateCalls)
			assert.Equal(t, tt.wantCheckerCalls, checkerCalls)
			if tt.wantPasswordInput {
				require.NotNil(t, service.lastInput)
				require.NotNil(t, service.lastInput.Password)
				assert.Equal(t, "newPassw0rd", *service.lastInput.Password)
			}
		})
	}
}
