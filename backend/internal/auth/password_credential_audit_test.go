package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// credentialAuditCapture remains a general HTTP audit stub for auth handler
// tests unrelated to transaction-bound credential mutation.
type credentialAuditCapture struct {
	entries []AuthAuditEntry
	err     error
}

func (*credentialAuditCapture) LogAuthLogin(
	context.Context,
	*uint64,
	*uint64,
	string,
	string,
	string,
) error {
	return nil
}

func (a *credentialAuditCapture) LogEntry(
	_ context.Context,
	entry AuthAuditEntry,
) error {
	a.entries = append(a.entries, entry)
	return a.err
}

type passwordResetCompletionStub struct {
	result    *PasswordResetCompletion
	err       error
	calls     *int
	lastAudit CredentialMutationAudit
}

func (passwordResetCompletionStub) ForgotPassword(context.Context, string) error {
	return nil
}

func (s passwordResetCompletionStub) ResetPassword(
	ctx context.Context,
	rawToken, newPassword string,
) error {
	_, err := s.ResetPasswordWithResult(
		ctx,
		rawToken,
		newPassword,
		testPasswordResetAudit(),
	)
	return err
}

func (s passwordResetCompletionStub) ResetPasswordWithResult(
	_ context.Context,
	_, _ string,
	audit CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	if s.calls != nil {
		*s.calls++
	}
	s.lastAudit = audit
	return s.result, s.err
}

func (passwordResetCompletionStub) Wait() {}

type passwordResetAuditStaffReader struct {
	staff *model.Staff
	err   error
}

func (s passwordResetAuditStaffReader) GetByID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return s.staff, s.err
}

func (s passwordResetAuditStaffReader) FindByAccountID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return s.staff, s.err
}

type passwordResetAuditAssignments struct {
	assignments []model.StaffClinicAssignment
	err         error
}

func (s passwordResetAuditAssignments) FindAllByStaffID(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error) {
	return append([]model.StaffClinicAssignment(nil), s.assignments...), s.err
}

func TestHTTPHandler_ChangeMyPasswordPassesExplicitCredentialAuditMetadata(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	accounts := &passwordHTTPAccountService{}
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: passwordHTTPStaffReader{
			staff: &model.Staff{
				ID:        17,
				ClinicID:  23,
				AccountID: &accountID,
			},
		},
		StaffAssignments: passwordResetAuditAssignments{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 23, IsMain: true},
			},
		},
		Accounts: accounts,
	}, CookieConfigForProduction(false))

	response := executePasswordHandler(
		t,
		http.MethodPut,
		"/api/v1/users/me/password",
		ChangeMyPasswordRequest{
			CurrentPassword: "oldPassw0rd",
			NewPassword:     "newPassw0rd",
		},
		func(c *gin.Context) {
			c.Set("clinic_id", "23")
			c.Set("user_id", "17")
			c.Request.RemoteAddr = "192.0.2.1:1234"
			c.Request.Header.Set("User-Agent", "credential-audit-test")
		},
		handler.ChangeMyPassword,
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	audit := accounts.credentialAudit
	require.NotNil(t, audit.ActorID)
	assert.Equal(t, uint64(23), audit.ClinicID)
	assert.Equal(t, uint64(17), *audit.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, audit.ActorType)
	assert.Equal(t, model.AuditActionAuthPasswordChange, audit.Action)
	assert.Equal(t, uint64(17), audit.TargetStaffID)
	assert.Equal(t, "192.0.2.1", audit.IPAddress)
	assert.Equal(t, "credential-audit-test", audit.UserAgent)

	encoded, err := json.Marshal(audit)
	require.NoError(t, err)
	for _, secret := range []string{
		"oldPassw0rd",
		"newPassw0rd",
		"token",
		"email",
	} {
		assert.NotContains(t, string(encoded), secret)
	}
}

func TestHTTPHandler_ChangeMyPasswordSystemAdminUsesRequestedActiveClinic(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	accounts := &passwordHTTPAccountService{
		account: &model.Account{
			ID:            accountID,
			IsActive:      true,
			IsSystemAdmin: true,
		},
	}
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: passwordHTTPStaffReader{
			staff: &model.Staff{ID: 17, AccountID: &accountID},
		},
		StaffAssignments: passwordResetAuditAssignments{},
		Accounts:         accounts,
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{
				{ID: 22, Name: "Another Active Clinic", IsActive: true},
				{ID: 23, Name: "Requested Active Clinic", IsActive: true},
			},
		},
	}, CookieConfigForProduction(false))

	response := executePasswordHandler(
		t,
		http.MethodPut,
		"/api/v1/users/me/password",
		ChangeMyPasswordRequest{
			CurrentPassword: "oldPassw0rd",
			NewPassword:     "newPassw0rd",
		},
		func(c *gin.Context) {
			c.Set("clinic_id", "23")
			c.Set("user_id", "17")
		},
		handler.ChangeMyPassword,
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, uint64(23), accounts.credentialAudit.ClinicID)
}

func TestHTTPHandler_ChangeMyPasswordRejectsInactiveSystemAdminRequestedClinic(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	accounts := &passwordHTTPAccountService{
		account: &model.Account{
			ID:            accountID,
			IsActive:      true,
			IsSystemAdmin: true,
		},
	}
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: passwordHTTPStaffReader{
			staff: &model.Staff{ID: 17, AccountID: &accountID},
		},
		StaffAssignments: passwordResetAuditAssignments{},
		Accounts:         accounts,
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{
				{ID: 23, Name: "Inactive Clinic", IsActive: false},
				{ID: 24, Name: "Other Active Clinic", IsActive: true},
			},
		},
	}, CookieConfigForProduction(false))

	response := executePasswordHandler(
		t,
		http.MethodPut,
		"/api/v1/users/me/password",
		ChangeMyPasswordRequest{
			CurrentPassword: "oldPassw0rd",
			NewPassword:     "newPassw0rd",
		},
		func(c *gin.Context) {
			c.Set("clinic_id", "23")
			c.Set("user_id", "17")
		},
		handler.ChangeMyPassword,
	)

	assert.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
	assert.Zero(t, accounts.changeCalls)
}

func TestHTTPHandler_ResetPasswordPassesExplicitCredentialAuditMetadata(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	capturedAudit := CredentialMutationAudit{}
	service := passwordResetAuditInputCapture{
		onReset: func(audit CredentialMutationAudit) {
			capturedAudit = audit
		},
	}
	handler := NewHTTPHandler(
		HTTPDependencies{PasswordReset: service},
		CookieConfigForProduction(false),
	)

	response := executePasswordHandler(
		t,
		http.MethodPost,
		"/api/v1/auth/reset-password",
		ResetPasswordRequest{
			Token:    "private-reset-token",
			Password: "newPassw0rd",
		},
		func(c *gin.Context) {
			c.Request.RemoteAddr = "192.0.2.41:1234"
			c.Request.Header.Set("User-Agent", "reset-credential-audit-test")
		},
		handler.ResetPassword,
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Nil(t, capturedAudit.ActorID)
	assert.Equal(t, model.AuditActorTypeSystem, capturedAudit.ActorType)
	assert.Equal(t, model.AuditActionAuthPasswordReset, capturedAudit.Action)
	assert.Zero(t, capturedAudit.ClinicID)
	assert.Zero(t, capturedAudit.TargetStaffID)
	assert.Equal(t, "192.0.2.41", capturedAudit.IPAddress)
	assert.Equal(t, "reset-credential-audit-test", capturedAudit.UserAgent)

	encoded, err := json.Marshal(capturedAudit)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "private-reset-token")
	assert.NotContains(t, string(encoded), "newPassw0rd")
}

type passwordResetAuditInputCapture struct {
	onReset func(CredentialMutationAudit)
}

func (passwordResetAuditInputCapture) ForgotPassword(context.Context, string) error {
	return nil
}

func (s passwordResetAuditInputCapture) ResetPassword(
	ctx context.Context,
	rawToken, newPassword string,
) error {
	_, err := s.ResetPasswordWithResult(
		ctx,
		rawToken,
		newPassword,
		testPasswordResetAudit(),
	)
	return err
}

func (s passwordResetAuditInputCapture) ResetPasswordWithResult(
	_ context.Context,
	_, _ string,
	audit CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	if s.onReset != nil {
		s.onReset(audit)
	}
	return &PasswordResetCompletion{AccountID: 41}, nil
}

func (passwordResetAuditInputCapture) Wait() {}
