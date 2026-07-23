package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- mock AccountService (auth_session scope) ----

type mockAccountService struct {
	findByEmailFn        func(ctx context.Context, email string) (*model.Account, error)
	getByIDFn            func(ctx context.Context, id uint64) (*model.Account, error)
	updatePasswordHashFn func(ctx context.Context, accountID uint64, newHash string) error
}

func (m *mockAccountService) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockAccountService) GetByID(ctx context.Context, id uint64) (*model.Account, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAccountService) UpdatePasswordHash(ctx context.Context, accountID uint64, newHash string) error {
	if m.updatePasswordHashFn != nil {
		return m.updatePasswordHashFn(ctx, accountID, newHash)
	}
	return nil
}

// ---- mock StaffClinicAssignmentService (auth_session scope, configurable) ----

type mockAuthAssignmentService struct {
	findAllByStaffIDFn func(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
}

func (m *mockAuthAssignmentService) FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
	if m.findAllByStaffIDFn != nil {
		return m.findAllByStaffIDFn(ctx, staffID)
	}
	return nil, nil
}

func (m *mockAuthAssignmentService) Create(_ context.Context, _ *model.StaffClinicAssignment) error {
	return nil
}

func newHandlerForAuthSession(
	accountSvc service.AccountService,
	staffSvc service.StaffService,
	assignmentSvc service.StaffClinicAssignmentService,
	auditSvc service.AuditService,
) *Handler {
	return &Handler{
		cfg: &config.Config{JWTSecret: "test-secret-for-auth-session", GinMode: "debug"},
		svc: &service.Services{
			Account:               accountSvc,
			Staff:                 staffSvc,
			StaffClinicAssignment: assignmentSvc,
			Audit:                 auditSvc,
		},
	}
}

// ---- newJti ----

func TestNewJti(t *testing.T) {
	hexPattern := regexp.MustCompile(`^[0-9a-f]{32}$`)

	jti1 := newJti()
	jti2 := newJti()

	assert.Regexp(t, hexPattern, jti1, "newJti should return a 32-char lowercase hex string")
	assert.NotEmpty(t, jti1)
	assert.NotEqual(t, jti1, jti2, "successive calls should produce different jti values")
}

// ---- auditKnownAccountLoginFailure ----

func TestAuditKnownAccountLoginFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("skips audit when staff not resolved", func(t *testing.T) {
		audit := &mockAuditService{}
		h := newHandlerForAuthSession(
			&mockAccountService{},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return nil, apperrors.WrapNotFound("staff", "1")
				},
			},
			&mockAuthAssignmentService{},
			audit,
		)

		assert.NotPanics(t, func() {
			h.auditKnownAccountLoginFailure(context.Background(), 1, "127.0.0.1", "test-agent")
		})
		assert.Empty(t, audit.loggedActions, "audit log should not be written when staff lookup fails")
	})

	t.Run("skips audit when clinic assignments not resolved", func(t *testing.T) {
		audit := &mockAuditService{}
		h := newHandlerForAuthSession(
			&mockAccountService{},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10}, nil
				},
			},
			&mockAuthAssignmentService{
				findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			audit,
		)

		h.auditKnownAccountLoginFailure(context.Background(), 1, "127.0.0.1", "test-agent")
		assert.Empty(t, audit.loggedActions, "audit log should not be written when assignment lookup fails")
	})

	t.Run("skips audit when staff has no clinic assignments", func(t *testing.T) {
		audit := &mockAuditService{}
		h := newHandlerForAuthSession(
			&mockAccountService{},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10}, nil
				},
			},
			&mockAuthAssignmentService{
				findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{}, nil
				},
			},
			audit,
		)

		h.auditKnownAccountLoginFailure(context.Background(), 1, "127.0.0.1", "test-agent")
		assert.Empty(t, audit.loggedActions, "audit log should not be written when no assignments exist")
	})

	t.Run("logs failure with resolved clinic id", func(t *testing.T) {
		audit := &mockAuditService{}
		h := newHandlerForAuthSession(
			&mockAccountService{},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), accountID)
					return &model.Staff{ID: 10}, nil
				},
			},
			&mockAuthAssignmentService{
				findAllByStaffIDFn: func(_ context.Context, staffID uint64) ([]model.StaffClinicAssignment, error) {
					assert.Equal(t, uint64(10), staffID)
					return []model.StaffClinicAssignment{
						{ClinicID: 5, IsMain: false},
						{ClinicID: 7, IsMain: true},
					}, nil
				},
			},
			audit,
		)

		h.auditKnownAccountLoginFailure(context.Background(), 1, "127.0.0.1", "test-agent")
		require.Len(t, audit.loggedActions, 1)
		assert.Equal(t, model.AuditActionAuthLoginFailure, audit.loggedActions[0])
	})

	t.Run("swallows audit log error without panicking", func(t *testing.T) {
		audit := &mockAuditService{
			logAuthLoginFn: func(_ context.Context, _, _ *uint64, _, _, _ string) error {
				return fmt.Errorf("audit write failure")
			},
		}
		h := newHandlerForAuthSession(
			&mockAccountService{},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10}, nil
				},
			},
			&mockAuthAssignmentService{
				findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{{ClinicID: 5, IsMain: true}}, nil
				},
			},
			audit,
		)

		assert.NotPanics(t, func() {
			h.auditKnownAccountLoginFailure(context.Background(), 1, "127.0.0.1", "test-agent")
		})
	})
}

// ---- authenticateUser ----

func TestAuthenticateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)
	passwordHash := string(hash)

	t.Run("returns unauthorized when account not found", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return nil, apperrors.WrapNotFound("account", "unknown@test.com")
				},
			},
			&mockStaffService{},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "unknown@test.com", "x", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})

	t.Run("passes through generic FindByEmail error", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return nil, fmt.Errorf("db failure")
				},
			},
			&mockStaffService{},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "x", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find account by email")
		assert.Contains(t, err.Error(), "db failure")
		assert.ErrorContains(t, err, "db failure")
	})

	t.Run("returns unauthorized when account inactive", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: false, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffService{},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "correct-password", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})

	t.Run("returns unauthorized and audits on wrong password", func(t *testing.T) {
		audit := &mockAuditService{}
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10}, nil
				},
			},
			&mockAuthAssignmentService{
				findAllByStaffIDFn: func(_ context.Context, _ uint64) ([]model.StaffClinicAssignment, error) {
					return []model.StaffClinicAssignment{{ClinicID: 5, IsMain: true}}, nil
				},
			},
			audit,
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "wrong-password", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
		assert.Len(t, audit.loggedActions, 1, "wrong password should trigger a best-effort audit log")
	})

	t.Run("wraps error when staff lookup fails after password match", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return nil, fmt.Errorf("staff db failure")
				},
			},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "correct-password", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
	})

	t.Run("returns unauthorized when staff inactive", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
					return &model.Account{ID: 1, IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, _ uint64) (*model.Staff, error) {
					return &model.Staff{ID: 10, IsActive: false}, nil
				},
			},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "correct-password", "127.0.0.1", "ua")

		assert.Nil(t, account)
		assert.Nil(t, staff)
		require.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrUnauthorized))
	})

	t.Run("returns account and staff on success", func(t *testing.T) {
		h := newHandlerForAuthSession(
			&mockAccountService{
				findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
					assert.Equal(t, "e@test.com", email)
					return &model.Account{ID: 1, Email: "e@test.com", IsActive: true, PasswordHash: passwordHash}, nil
				},
			},
			&mockStaffService{
				findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
					assert.Equal(t, uint64(1), accountID)
					return &model.Staff{ID: 10, IsActive: true}, nil
				},
			},
			&mockAuthAssignmentService{},
			&mockAuditService{},
		)

		account, staff, err := h.authenticateUser(context.Background(), "e@test.com", "correct-password", "127.0.0.1", "ua")

		require.NoError(t, err)
		require.NotNil(t, account)
		require.NotNil(t, staff)
		assert.Equal(t, uint64(1), account.ID)
		assert.Equal(t, uint64(10), staff.ID)
	})
}

// ---- resolveClinicInfo ----

func TestResolveClinicInfo(t *testing.T) {
	tests := []struct {
		name          string
		assignments   []model.StaffClinicAssignment
		wantMainID    string
		wantClinicIDs []uint64
	}{
		{
			name:          "empty assignments",
			assignments:   nil,
			wantMainID:    "",
			wantClinicIDs: []uint64{},
		},
		{
			name: "uses IsMain assignment",
			assignments: []model.StaffClinicAssignment{
				{ClinicID: 1, IsMain: false},
				{ClinicID: 2, IsMain: true},
				{ClinicID: 3, IsMain: false},
			},
			wantMainID:    "2",
			wantClinicIDs: []uint64{1, 2, 3},
		},
		{
			name: "falls back to first assignment when no IsMain",
			assignments: []model.StaffClinicAssignment{
				{ClinicID: 8, IsMain: false},
				{ClinicID: 9, IsMain: false},
			},
			wantMainID:    "8",
			wantClinicIDs: []uint64{8, 9},
		},
		{
			name: "single assignment without IsMain",
			assignments: []model.StaffClinicAssignment{
				{ClinicID: 42, IsMain: false},
			},
			wantMainID:    "42",
			wantClinicIDs: []uint64{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mainClinicID, clinicIDs := resolveClinicInfo(tt.assignments)
			assert.Equal(t, tt.wantMainID, mainClinicID)
			assert.Equal(t, tt.wantClinicIDs, clinicIDs)
		})
	}
}

// ---- issueAuthCookies ----

func TestIssueAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("dev mode uses Lax + non-secure cookies", func(t *testing.T) {
		h := newHandlerForAuthSession(&mockAccountService{}, &mockStaffService{}, &mockAuthAssignmentService{}, &mockAuditService{})
		h.cfg.GinMode = "debug"

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := h.issueAuthCookies(c, 10, "1", false, []uint64{1, 2})
		require.NoError(t, err)

		cookies := w.Result().Cookies()
		var access, refresh, legacyRefreshClear *http.Cookie
		for _, ck := range cookies {
			switch ck.Name {
			case accessTokenCookieName:
				access = ck
			case refreshTokenCookieName:
				if ck.Path == "/api/v1/auth" {
					refresh = ck
				}
				if ck.Path == "/api/v1/auth/refresh" && ck.MaxAge < 0 {
					legacyRefreshClear = ck
				}
			}
		}
		require.NotNil(t, access)
		require.NotNil(t, refresh)
		require.NotNil(t, legacyRefreshClear)
		assert.False(t, access.Secure)
		assert.Equal(t, http.SameSiteLaxMode, access.SameSite)
		assert.True(t, access.HttpOnly)
		assert.Equal(t, "/", access.Path)
		assert.Equal(t, "/api/v1/auth", refresh.Path)

		claims := &middleware.JWTClaims{}
		_, err = jwt.ParseWithClaims(access.Value, claims, func(_ *jwt.Token) (any, error) {
			return []byte(h.cfg.JWTSecret), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "10", claims.UserID)
		assert.Equal(t, "1", claims.ClinicID)
		assert.False(t, claims.IsSystemAdmin)
		assert.Equal(t, []uint64{1, 2}, claims.ClinicIDs)

		refreshClaims := &middleware.JWTClaims{}
		_, err = jwt.ParseWithClaims(refresh.Value, refreshClaims, func(_ *jwt.Token) (any, error) {
			return []byte(h.cfg.JWTSecret), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "refresh", refreshClaims.Subject)
	})

	t.Run("release mode uses None + secure cookies", func(t *testing.T) {
		h := newHandlerForAuthSession(&mockAccountService{}, &mockStaffService{}, &mockAuthAssignmentService{}, &mockAuditService{})
		h.cfg.GinMode = "release"

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := h.issueAuthCookies(c, 20, "3", true, []uint64{3})
		require.NoError(t, err)

		cookies := w.Result().Cookies()
		var access *http.Cookie
		for _, ck := range cookies {
			if ck.Name == accessTokenCookieName {
				access = ck
			}
		}
		require.NotNil(t, access)
		assert.True(t, access.Secure)
		assert.Equal(t, http.SameSiteNoneMode, access.SameSite)

		claims := &middleware.JWTClaims{}
		_, err = jwt.ParseWithClaims(access.Value, claims, func(_ *jwt.Token) (any, error) {
			return []byte(h.cfg.JWTSecret), nil
		})
		require.NoError(t, err)
		assert.True(t, claims.IsSystemAdmin)
	})
}
