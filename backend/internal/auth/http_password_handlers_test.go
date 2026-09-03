package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type passwordHandlerResetService struct {
	resetPasswordFn func(context.Context, string, string) error
}

func (passwordHandlerResetService) ForgotPassword(context.Context, string) error {
	return nil
}

func (s passwordHandlerResetService) ResetPassword(
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

func (s passwordHandlerResetService) ResetPasswordWithResult(
	ctx context.Context,
	rawToken, newPassword string,
	_ CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	if s.resetPasswordFn != nil {
		if err := s.resetPasswordFn(ctx, rawToken, newPassword); err != nil {
			return nil, err
		}
	}
	return &PasswordResetCompletion{AccountID: 41}, nil
}

func (passwordHandlerResetService) Wait() {}

type passwordHTTPStaffReader struct {
	staff *model.Staff
	err   error
}

type passwordHTTPAssignmentReader struct{}

func (passwordHTTPAssignmentReader) FindAllByStaffID(
	_ context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return []model.StaffClinicAssignment{
		{StaffID: staffID, ClinicID: 23, IsMain: true},
	}, nil
}

func (s passwordHTTPStaffReader) GetByID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return s.staff, s.err
}

func (s passwordHTTPStaffReader) FindByAccountID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return nil, apperrors.WrapNotFound("staff", "account")
}

type passwordHTTPAccountService struct {
	account         *model.Account
	getError        error
	updateError     error
	changeError     error
	updatedHash     string
	updatedID       uint64
	getCalls        int
	updateCalls     int
	changeCalls     int
	currentPassword string
	newPassword     string
	credentialAudit CredentialMutationAudit
}

func (s *passwordHTTPAccountService) FindByEmail(
	context.Context,
	string,
) (*model.Account, error) {
	return s.account, nil
}

func (s *passwordHTTPAccountService) GetByID(
	context.Context,
	uint64,
) (*model.Account, error) {
	s.getCalls++
	return s.account, s.getError
}

func (s *passwordHTTPAccountService) UpdatePasswordHash(
	_ context.Context,
	accountID uint64,
	newHash string,
) error {
	s.updateCalls++
	s.updatedID = accountID
	s.updatedHash = newHash
	return s.updateError
}

func (s *passwordHTTPAccountService) ChangePassword(
	_ context.Context,
	accountID uint64,
	currentPassword, newPassword string,
	audit CredentialMutationAudit,
) error {
	s.changeCalls++
	s.updatedID = accountID
	s.currentPassword = currentPassword
	s.newPassword = newPassword
	s.credentialAudit = audit
	return s.changeError
}

func executePasswordHandler(
	t *testing.T,
	method, target string,
	body any,
	configure func(*gin.Context),
	handle gin.HandlerFunc,
) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(method, target, bytes.NewReader(encoded))
	c.Request.Header.Set("Content-Type", "application/json")
	if configure != nil {
		configure(c)
	}
	handle(c)
	return response
}

func TestValidatePassword_BoundsAndComplexity(t *testing.T) {
	assert.ErrorIs(t, ValidatePassword("short1"), apperrors.ErrInvalidInput)
	assert.ErrorIs(
		t,
		ValidatePassword(strings.Repeat("あ", 3)+"1"),
		apperrors.ErrInvalidInput,
	)
	assert.ErrorIs(t, ValidatePassword("onlyletters"), apperrors.ErrInvalidInput)
	assert.ErrorIs(t, ValidatePassword("12345678"), apperrors.ErrInvalidInput)
	assert.ErrorIs(
		t,
		ValidatePassword(strings.Repeat("界", 24)+"a1"),
		apperrors.ErrInvalidInput,
	)
	assert.NoError(t, ValidatePassword(strings.Repeat("あ", 7)+"1"))
	assert.NoError(t, ValidatePassword("ValidPassw0rd"))
}

func TestHTTPHandler_ChangeMyPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(9)

	tests := []struct {
		name          string
		body          any
		staff         *model.Staff
		staffError    error
		changeError   error
		setUser       bool
		expectedCode  int
		expectChanged bool
	}{
		{
			name:          "changes password",
			body:          ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			staff:         &model.Staff{ID: 1, AccountID: &accountID},
			setUser:       true,
			expectedCode:  http.StatusOK,
			expectChanged: true,
		},
		{
			name: "changes to an eight-rune Unicode password",
			body: ChangeMyPasswordRequest{
				CurrentPassword: "oldPassw0rd",
				NewPassword:     strings.Repeat("あ", 7) + "1",
			},
			staff:         &model.Staff{ID: 1, AccountID: &accountID},
			setUser:       true,
			expectedCode:  http.StatusOK,
			expectChanged: true,
		},
		{
			name: "rejects a four-rune Unicode password",
			body: ChangeMyPasswordRequest{
				CurrentPassword: "oldPassw0rd",
				NewPassword:     strings.Repeat("あ", 3) + "1",
			},
			setUser:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "binding failure",
			body:         map[string]any{"current_password": "oldPassw0rd"},
			setUser:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "weak new password",
			body:         ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "allletters"},
			setUser:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "missing user context",
			body:         ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "staff lookup failure",
			body:         ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			staffError:   apperrors.WrapNotFound("staff", "1"),
			setUser:      true,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "staff has no account",
			body:         ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			staff:        &model.Staff{ID: 1},
			setUser:      true,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:          "account lookup failure",
			body:          ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			staff:         &model.Staff{ID: 1, AccountID: &accountID},
			changeError:   apperrors.WrapNotFound("account", "9"),
			setUser:       true,
			expectedCode:  http.StatusNotFound,
			expectChanged: true,
		},
		{
			name:          "wrong current password",
			body:          ChangeMyPasswordRequest{CurrentPassword: "wrongPassw0rd", NewPassword: "newPassw0rd"},
			staff:         &model.Staff{ID: 1, AccountID: &accountID},
			changeError:   apperrors.WrapUnauthorized("現在のパスワードが正しくありません"),
			setUser:       true,
			expectedCode:  http.StatusUnauthorized,
			expectChanged: true,
		},
		{
			name:          "password update failure",
			body:          ChangeMyPasswordRequest{CurrentPassword: "oldPassw0rd", NewPassword: "newPassw0rd"},
			staff:         &model.Staff{ID: 1, AccountID: &accountID},
			changeError:   errors.New("database unavailable"),
			setUser:       true,
			expectedCode:  http.StatusInternalServerError,
			expectChanged: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accounts := &passwordHTTPAccountService{
				changeError: test.changeError,
			}
			handler := NewHTTPHandler(HTTPDependencies{
				Staff: passwordHTTPStaffReader{
					staff: test.staff,
					err:   test.staffError,
				},
				StaffAssignments: passwordHTTPAssignmentReader{},
				Accounts:         accounts,
				Audit:            &credentialAuditCapture{},
			}, CookieConfigForProduction(false))

			response := executePasswordHandler(
				t,
				http.MethodPut,
				"/api/v1/users/me/password",
				test.body,
				func(c *gin.Context) {
					if test.setUser {
						c.Set("clinic_id", "23")
						c.Set("user_id", "1")
					}
				},
				handler.ChangeMyPassword,
			)

			assert.Equal(t, test.expectedCode, response.Code, response.Body.String())
			if test.expectChanged {
				assert.Equal(t, accountID, accounts.updatedID)
				assert.Equal(t, test.body.(ChangeMyPasswordRequest).CurrentPassword, accounts.currentPassword)
				assert.Equal(t, test.body.(ChangeMyPasswordRequest).NewPassword, accounts.newPassword)
			} else {
				assert.Zero(t, accounts.changeCalls)
			}
		})
	}
}

func TestPasswordAcceptedByChangeBindingAlsoPassesLoginBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	password := strings.Repeat("あ", 7) + "1"

	changeBody, err := json.Marshal(ChangeMyPasswordRequest{
		CurrentPassword: "oldPassw0rd",
		NewPassword:     password,
	})
	require.NoError(t, err)
	changeContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	changeContext.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/v1/users/me/password",
		bytes.NewReader(changeBody),
	)
	changeContext.Request.Header.Set("Content-Type", "application/json")
	var changeRequest ChangeMyPasswordRequest
	require.NoError(t, bindAuthJSON(changeContext, &changeRequest))
	require.NoError(t, ValidatePassword(changeRequest.NewPassword))

	loginBody, err := json.Marshal(LoginInput{
		Email:    "unicode-password@example.test",
		Password: password,
	})
	require.NoError(t, err)
	loginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	loginContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		bytes.NewReader(loginBody),
	)
	loginContext.Request.Header.Set("Content-Type", "application/json")
	var loginRequest LoginInput
	require.NoError(t, bindAuthJSON(loginContext, &loginRequest))
	assert.Equal(t, changeRequest.NewPassword, loginRequest.Password)
}

func TestHTTPHandler_ChangeMyPassword_DelegatesAtomicAccountUseCase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(9)
	accounts := &passwordHTTPAccountService{}
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: passwordHTTPStaffReader{
			staff: &model.Staff{ID: 1, AccountID: &accountID},
		},
		StaffAssignments: passwordHTTPAssignmentReader{},
		Accounts:         accounts,
		Audit:            &credentialAuditCapture{},
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
			c.Set("user_id", "1")
		},
		handler.ChangeMyPassword,
	)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, 1, accounts.changeCalls)
	assert.Zero(t, accounts.getCalls, "HTTP must not open a read-then-update race")
	assert.Zero(t, accounts.updateCalls, "HTTP must not bypass the atomic password use case")
	assert.Equal(t, accountID, accounts.updatedID)
	assert.Equal(t, "oldPassw0rd", accounts.currentPassword)
	assert.Equal(t, "newPassw0rd", accounts.newPassword)
}

func TestHTTPHandler_ResetPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		body         any
		service      passwordHandlerResetService
		expectedCode int
		expectedBody string
	}{
		{
			name: "success",
			body: ResetPasswordRequest{Token: "valid-token", Password: "newPassw0rd"},
			service: passwordHandlerResetService{
				resetPasswordFn: func(_ context.Context, token, password string) error {
					assert.Equal(t, "valid-token", token)
					assert.Equal(t, "newPassw0rd", password)
					return nil
				},
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "binding failure",
			body:         map[string]any{"password": "newPassw0rd"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "complexity failure",
			body:         ResetPasswordRequest{Token: "valid-token", Password: "onlyletters"},
			expectedCode: http.StatusBadRequest,
			expectedBody: `"error":"パスワードは英字と数字の両方を含めてください"`,
		},
		{
			name: "service failure",
			body: ResetPasswordRequest{Token: "expired-token", Password: "newPassw0rd"},
			service: passwordHandlerResetService{
				resetPasswordFn: func(context.Context, string, string) error {
					return apperrors.WrapInvalidInput("invalid token")
				},
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accountID := uint64(41)
			handler := NewHTTPHandler(
				HTTPDependencies{
					PasswordReset: test.service,
					Staff: passwordResetAuditStaffReader{
						staff: &model.Staff{
							ID:        17,
							AccountID: &accountID,
						},
					},
					StaffAssignments: passwordHTTPAssignmentReader{},
					Audit:            &credentialAuditCapture{},
				},
				CookieConfigForProduction(false),
			)
			response := executePasswordHandler(
				t,
				http.MethodPost,
				"/api/v1/auth/reset-password",
				test.body,
				nil,
				handler.ResetPassword,
			)
			assert.Equal(t, test.expectedCode, response.Code, response.Body.String())
			if test.expectedBody != "" {
				assert.Contains(t, response.Body.String(), test.expectedBody)
				assert.NotContains(t, response.Body.String(), ": invalid input")
			}
		})
	}
}
