package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock PasswordResetService ----

type mockPasswordResetService struct {
	forgotPasswordFn func(ctx context.Context, email string) error
	resetPasswordFn  func(ctx context.Context, rawToken, newPassword string) error
}

func (m *mockPasswordResetService) ForgotPassword(ctx context.Context, email string) error {
	if m.forgotPasswordFn != nil {
		return m.forgotPasswordFn(ctx, email)
	}
	return nil
}

func (m *mockPasswordResetService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if m.resetPasswordFn != nil {
		return m.resetPasswordFn(ctx, rawToken, newPassword)
	}
	return nil
}

// ---- test helper ----

func newHandlerWithAuthPasswordSvc(staffSvc service.StaffService, accountSvc service.AccountService, pwResetSvc service.PasswordResetService) *Handler {
	return &Handler{
		svc: &service.Services{
			Staff:         staffSvc,
			Account:       accountSvc,
			PasswordReset: pwResetSvc,
		},
	}
}

// hashPasswordForTest はテスト用に bcrypt ハッシュを生成するヘルパー。
func hashPasswordForTest(t *testing.T, password string) string {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hashed)
}

// ---- ChangeMyPassword ----

func TestChangeMyPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentHash := hashPasswordForTest(t, "oldPassw0rd")
	accountID := uint64(9)

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		staffSvc   *mockStaffService
		accountSvc *mockAccountService
		wantStatus int
	}{
		{
			name: "changes password successfully",
			body: map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) {
				setStaffID(c)
			},
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), id)
					return &model.Staff{ID: id, AccountID: &accountID}, nil
				},
			},
			accountSvc: &mockAccountService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
					assert.Equal(t, accountID, id)
					return &model.Account{ID: id, PasswordHash: currentHash}, nil
				},
				updatePasswordHashFn: func(_ context.Context, id uint64, newHash string) error {
					assert.Equal(t, accountID, id)
					assert.NotEmpty(t, newHash)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when body binding fails",
			body:       map[string]any{"current_password": "oldPassw0rd"},
			setupCtx:   func(c *gin.Context) { setStaffID(c) },
			staffSvc:   &mockStaffService{},
			accountSvc: &mockAccountService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when new password is too weak",
			body:       map[string]any{"current_password": "oldPassw0rd", "new_password": "short"},
			setupCtx:   func(c *gin.Context) { setStaffID(c) },
			staffSvc:   &mockStaffService{},
			accountSvc: &mockAccountService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when user context is missing",
			body:       map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx:   func(_ *gin.Context) {},
			staffSvc:   &mockStaffService{},
			accountSvc: &mockAccountService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns error when Staff.GetByID fails",
			body:     map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) { setStaffID(c) },
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return nil, apperrors.WrapNotFound("staff", "1")
				},
			},
			accountSvc: &mockAccountService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 400 when staff has no linked account",
			body:     map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) { setStaffID(c) },
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, AccountID: nil}, nil
				},
			},
			accountSvc: &mockAccountService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns error when Account.GetByID fails",
			body:     map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) { setStaffID(c) },
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, AccountID: &accountID}, nil
				},
			},
			accountSvc: &mockAccountService{
				getByIDFn: func(_ context.Context, _ uint64) (*model.Account, error) {
					return nil, apperrors.WrapNotFound("account", "9")
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "returns 401 when current password is wrong",
			body:     map[string]any{"current_password": "wrongPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) { setStaffID(c) },
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, AccountID: &accountID}, nil
				},
			},
			accountSvc: &mockAccountService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
					return &model.Account{ID: id, PasswordHash: currentHash}, nil
				},
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:     "returns error when UpdatePasswordHash fails",
			body:     map[string]any{"current_password": "oldPassw0rd", "new_password": "newPassw0rd"},
			setupCtx: func(c *gin.Context) { setStaffID(c) },
			staffSvc: &mockStaffService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
					return &model.Staff{ID: id, AccountID: &accountID}, nil
				},
			},
			accountSvc: &mockAccountService{
				getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
					return &model.Account{ID: id, PasswordHash: currentHash}, nil
				},
				updatePasswordHashFn: func(_ context.Context, _ uint64, _ string) error {
					return fmt.Errorf("db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAuthPasswordSvc(tt.staffSvc, tt.accountSvc, &mockPasswordResetService{})

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.ChangeMyPassword(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ForgotPassword ----

func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		pwResetSvc *mockPasswordResetService
		wantStatus int
	}{
		{
			name: "returns 200 when email exists",
			body: map[string]any{"email": "owner@example.com"},
			pwResetSvc: &mockPasswordResetService{
				forgotPasswordFn: func(_ context.Context, email string) error {
					assert.Equal(t, "owner@example.com", email)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 200 even when email does not exist (no leak)",
			body: map[string]any{"email": "nobody@example.com"},
			pwResetSvc: &mockPasswordResetService{
				forgotPasswordFn: func(_ context.Context, _ string) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when email is missing",
			body:       map[string]any{},
			pwResetSvc: &mockPasswordResetService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when email format is invalid",
			body:       map[string]any{"email": "not-an-email"},
			pwResetSvc: &mockPasswordResetService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns 500 when service fails unexpectedly",
			body: map[string]any{"email": "owner@example.com"},
			pwResetSvc: &mockPasswordResetService{
				forgotPasswordFn: func(_ context.Context, _ string) error {
					return fmt.Errorf("smtp failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAuthPasswordSvc(&mockStaffService{}, &mockAccountService{}, tt.pwResetSvc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			h.ForgotPassword(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ---- ResetPassword ----

func TestResetPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       any
		pwResetSvc *mockPasswordResetService
		wantStatus int
	}{
		{
			name: "resets password successfully",
			body: map[string]any{"token": "valid-token", "password": "newPassw0rd"},
			pwResetSvc: &mockPasswordResetService{
				resetPasswordFn: func(_ context.Context, token, newPassword string) error {
					assert.Equal(t, "valid-token", token)
					assert.Equal(t, "newPassw0rd", newPassword)
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 when token is missing",
			body:       map[string]any{"password": "newPassw0rd"},
			pwResetSvc: &mockPasswordResetService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when password is too short for binding",
			body:       map[string]any{"token": "valid-token", "password": "short1"},
			pwResetSvc: &mockPasswordResetService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when password fails complexity validation",
			body:       map[string]any{"token": "valid-token", "password": "alllettersnodigit"},
			pwResetSvc: &mockPasswordResetService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "returns error when token is invalid or expired",
			body: map[string]any{"token": "expired-token", "password": "newPassw0rd"},
			pwResetSvc: &mockPasswordResetService{
				resetPasswordFn: func(_ context.Context, _, _ string) error {
					return apperrors.WrapInvalidInput("トークンが無効です")
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithAuthPasswordSvc(&mockStaffService{}, &mockAccountService{}, tt.pwResetSvc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			h.ResetPassword(c)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
