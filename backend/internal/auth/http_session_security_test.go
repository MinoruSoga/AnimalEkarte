package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const refreshSecurityJWTSecret = "test-secret-for-refresh-security"

type refreshHTTPAccountServiceStub struct {
	AccountService
	getByIDFn func(context.Context, uint64) (*model.Account, error)
}

func (s refreshHTTPAccountServiceStub) GetByID(
	ctx context.Context,
	id uint64,
) (*model.Account, error) {
	return s.getByIDFn(ctx, id)
}

type refreshHTTPStaffReaderStub struct {
	getByIDFn func(context.Context, uint64) (*model.Staff, error)
}

func (s refreshHTTPStaffReaderStub) GetByID(
	ctx context.Context,
	id uint64,
) (*model.Staff, error) {
	return s.getByIDFn(ctx, id)
}

func (refreshHTTPStaffReaderStub) FindByAccountID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return nil, apperrors.WrapNotFound("staff", "account")
}

type refreshHTTPAssignmentReaderStub struct {
	assignments []model.StaffClinicAssignment
	err         error
}

func (s refreshHTTPAssignmentReaderStub) FindAllByStaffID(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error) {
	return append([]model.StaffClinicAssignment(nil), s.assignments...), s.err
}

type refreshHTTPClinicListerStub struct {
	clinics []model.Clinic
	err     error
}

type loginHTTPAuthServiceStub struct {
	account *model.Account
	staff   *model.Staff
}

type authCookieIssueStub struct {
	TokenService
	access     *IssuedToken
	refreshErr error
}

func (s authCookieIssueStub) IssueAccessToken(
	uint64,
	string,
	bool,
	[]uint64,
	int64,
) (*IssuedToken, error) {
	return s.access, nil
}

func (s authCookieIssueStub) IssueRefreshToken(
	uint64,
	string,
	bool,
	[]uint64,
	int64,
) (*IssuedToken, error) {
	return nil, s.refreshErr
}

func (s loginHTTPAuthServiceStub) AuthenticateUser(
	context.Context,
	string,
	string,
) (*model.Account, *model.Staff, error) {
	return s.account, s.staff, nil
}

func (loginHTTPAuthServiceStub) ResolveClinicInfo(
	assignments []model.StaffClinicAssignment,
) (string, []uint64) {
	return ResolveClinicInfo(assignments)
}

func (loginHTTPAuthServiceStub) ResolveSystemAdminMainClinicID(
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) string {
	return ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
}

func (loginHTTPAuthServiceStub) CalculateEffectivePermissions(
	context.Context,
	bool,
	uint64,
	uint64,
) AuthEffectivePermissions {
	return make(AuthEffectivePermissions)
}

func (s refreshHTTPClinicListerStub) ListClinics(
	context.Context,
) ([]model.Clinic, error) {
	return append([]model.Clinic(nil), s.clinics...), s.err
}

func executeRefresh(
	t *testing.T,
	account *model.Account,
	staff *model.Staff,
	assignments []model.StaffClinicAssignment,
	tokenEpoch int64,
	tokenIsSystemAdmin bool,
) (*httptest.ResponseRecorder, TokenService) {
	t.Helper()
	blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{})
	tokens := NewTokenService(refreshSecurityJWTSecret, blacklist)
	issued, err := tokens.IssueRefreshToken(
		staff.ID,
		"99",
		tokenIsSystemAdmin,
		[]uint64{99},
		tokenEpoch,
	)
	require.NoError(t, err)

	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Accounts: refreshHTTPAccountServiceStub{
			getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
				if staff.AccountID == nil || id != *staff.AccountID {
					return nil, apperrors.WrapNotFound("account", "linked")
				}
				return account, nil
			},
		},
		Staff: refreshHTTPStaffReaderStub{
			getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				if id != staff.ID {
					return nil, apperrors.WrapNotFound("staff", "refresh")
				}
				return staff, nil
			},
		},
		StaffAssignments: refreshHTTPAssignmentReaderStub{assignments: assignments},
		Clinics:          refreshHTTPClinicListerStub{clinics: []model.Clinic{{ID: 2}, {ID: 3}}},
	}, CookieConfigForProduction(false))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", http.NoBody)
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: issued.Token,
	})
	c.Request = request
	handler.RefreshToken(c)
	return recorder, tokens
}

func TestHTTPHandler_IssueAuthCookies_WritesNothingUntilBothTokensExist(
	t *testing.T,
) {
	issueFailure := errors.New("refresh entropy unavailable")
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens: authCookieIssueStub{
			access: &IssuedToken{
				Token:     "valid-access-token",
				ExpiresAt: time.Now().Add(accessTokenTTL),
			},
			refreshErr: issueFailure,
		},
	}, CookieConfigForProduction(false))
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	err := handler.IssueAuthCookies(c, 17, "23", false, []uint64{23}, 1)

	require.ErrorIs(t, err, issueFailure)
	assert.Empty(t, recorder.Header().Values("Set-Cookie"))
}

func TestHTTPHandler_Login_IssuesCurrentAccountEpoch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	updatedAt := time.Unix(1_727_123_456, 789_012_345)
	account := &model.Account{
		ID:        accountID,
		Email:     "login-epoch@example.test",
		IsActive:  true,
		UpdatedAt: updatedAt,
	}
	staff := &model.Staff{
		ID:        17,
		AccountID: &accountID,
		Name:      "Epoch Staff",
		IsActive:  true,
	}
	blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{})
	tokens := NewTokenService(refreshSecurityJWTSecret, blacklist)
	handler := NewHTTPHandler(HTTPDependencies{
		Auth:           loginHTTPAuthServiceStub{account: account, staff: staff},
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		StaffAssignments: refreshHTTPAssignmentReaderStub{
			assignments: []model.StaffClinicAssignment{
				{StaffID: staff.ID, ClinicID: 2, IsMain: true},
			},
		},
		Clinics: refreshHTTPClinicListerStub{clinics: []model.Clinic{{ID: 2}}},
	}, CookieConfigForProduction(false))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		strings.NewReader(fmt.Sprintf(
			`{"email":%q,"password":"irrelevant-password"}`,
			account.Email,
		)),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	handler.Login(c)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name != RefreshTokenCookieName || cookie.MaxAge <= 0 {
			continue
		}
		claims, err := tokens.ParseRefreshTokenClaims(cookie.Value)
		require.NoError(t, err)
		assert.Equal(t, updatedAt.UnixNano(), claims.AccountEpoch)
		return
	}
	t.Fatal("login did not issue a refresh token")
}

func TestHTTPHandler_Login_FailsBeforeCookiesWithoutResolvableClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	databaseUnavailable := errors.New("database unavailable")

	tests := []struct {
		name         string
		clinicError  error
		expectedCode int
	}{
		{
			name:         "system administrator fallback dependency failure",
			clinicError:  databaseUnavailable,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "system administrator fallback returns no clinic",
			expectedCode: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHTTPHandler(HTTPDependencies{
				Auth: loginHTTPAuthServiceStub{
					account: &model.Account{
						ID:            accountID,
						Email:         "system-admin@example.test",
						IsActive:      true,
						IsSystemAdmin: true,
						UpdatedAt:     time.Now().Add(-time.Hour),
					},
					staff: &model.Staff{
						ID:        17,
						AccountID: &accountID,
						IsActive:  true,
					},
				},
				Tokens:           NewTokenService(refreshSecurityJWTSecret, nil),
				StaffAssignments: refreshHTTPAssignmentReaderStub{},
				Clinics: refreshHTTPClinicListerStub{
					err: test.clinicError,
				},
			}, CookieConfigForProduction(false))

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/v1/login",
				strings.NewReader(
					`{"email":"system-admin@example.test","password":"irrelevant-password"}`,
				),
			)
			c.Request.Header.Set("Content-Type", "application/json")

			handler.Login(c)

			require.Equal(t, test.expectedCode, recorder.Code, recorder.Body.String())
			assert.Empty(
				t,
				recorder.Header().Values("Set-Cookie"),
				"failed login must not issue authentication cookies",
			)
		})
	}
}

func TestHTTPHandler_Login_SystemAdminTokenUsesOnlyActiveExistingClinics(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	updatedAt := time.Now().Add(-time.Hour)
	tokens := NewTokenService(refreshSecurityJWTSecret, nil)
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: loginHTTPAuthServiceStub{
			account: &model.Account{
				ID:            accountID,
				Email:         "system-admin@example.test",
				IsActive:      true,
				IsSystemAdmin: true,
				UpdatedAt:     updatedAt,
			},
			staff: &model.Staff{
				ID:        17,
				AccountID: &accountID,
				IsActive:  true,
			},
		},
		Tokens: tokens,
		StaffAssignments: refreshHTTPAssignmentReaderStub{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 22, IsMain: true},
			},
		},
		Clinics: refreshHTTPClinicListerStub{clinics: []model.Clinic{
			{ID: 0, IsActive: true},
			{ID: 22, IsActive: false},
			{ID: 23, IsActive: true},
			{ID: 24, IsActive: true},
		}},
	}, CookieConfigForProduction(false))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		strings.NewReader(
			`{"email":"system-admin@example.test","password":"irrelevant-password"}`,
		),
	)
	c.Request.Header.Set("Content-Type", "application/json")
	handler.Login(c)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var response LoginResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.NotNil(t, response.User)
	assert.Equal(t, "23", response.User.MainClinicID)
	assert.Equal(t, []MeClinicMembership{
		{ClinicID: "23", IsMain: true},
		{ClinicID: "24"},
	}, response.User.Clinics)
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name != AccessTokenCookieName || cookie.MaxAge <= 0 {
			continue
		}
		claims, err := tokens.VerifyAccessToken(cookie.Value)
		require.NoError(t, err)
		assert.Equal(t, "23", claims.ClinicID)
		assert.Equal(t, []uint64{23, 24}, claims.ClinicIDs)
		return
	}
	t.Fatal("login did not issue an access token")
}

func TestHTTPHandler_RefreshToken_UsesCurrentAccountAndAssignments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	updatedAt := time.Unix(1_727_123_456, 789_012_345)
	account := &model.Account{
		ID:            accountID,
		IsActive:      true,
		IsSystemAdmin: false,
		UpdatedAt:     updatedAt,
	}
	staff := &model.Staff{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}

	recorder, tokens := executeRefresh(
		t,
		account,
		staff,
		[]model.StaffClinicAssignment{
			{StaffID: staff.ID, ClinicID: 2, IsMain: true},
			{StaffID: staff.ID, ClinicID: 3},
		},
		updatedAt.UnixNano(),
		true,
	)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var rotated *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == RefreshTokenCookieName &&
			cookie.Path == RefreshTokenCookiePath &&
			cookie.MaxAge > 0 {
			rotated = cookie
		}
	}
	require.NotNil(t, rotated)
	claims, err := tokens.ParseRefreshTokenClaims(rotated.Value)
	require.NoError(t, err)
	assert.False(t, claims.IsSystemAdmin, "stale privileged claim must not be copied")
	assert.Equal(t, "2", claims.ClinicID)
	assert.Equal(t, []uint64{2, 3}, claims.ClinicIDs)
	assert.Equal(t, updatedAt.UnixNano(), claims.AccountEpoch)
}

func TestHTTPHandler_RefreshToken_SystemAdminRevalidatesActiveClinicScope(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	updatedAt := time.Now().Add(-time.Hour)
	account := &model.Account{
		ID:            accountID,
		IsActive:      true,
		IsSystemAdmin: true,
		UpdatedAt:     updatedAt,
	}
	staff := &model.Staff{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}
	blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{})
	tokens := NewTokenService(refreshSecurityJWTSecret, blacklist)
	issued, err := tokens.IssueRefreshToken(
		staff.ID,
		"22",
		true,
		[]uint64{22},
		updatedAt.UnixNano(),
	)
	require.NoError(t, err)
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Accounts: refreshHTTPAccountServiceStub{
			getByIDFn: func(context.Context, uint64) (*model.Account, error) {
				return account, nil
			},
		},
		Staff: refreshHTTPStaffReaderStub{
			getByIDFn: func(context.Context, uint64) (*model.Staff, error) {
				return staff, nil
			},
		},
		StaffAssignments: refreshHTTPAssignmentReaderStub{
			assignments: []model.StaffClinicAssignment{
				{StaffID: staff.ID, ClinicID: 22, IsMain: true},
			},
		},
		Clinics: refreshHTTPClinicListerStub{clinics: []model.Clinic{
			{ID: 22, IsActive: false},
			{ID: 23, IsActive: true},
			{ID: 24, IsActive: true},
		}},
	}, CookieConfigForProduction(false))

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		http.NoBody,
	)
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: issued.Token,
	})
	c.Request = request
	handler.RefreshToken(c)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name != AccessTokenCookieName || cookie.MaxAge <= 0 {
			continue
		}
		claims, verifyErr := tokens.VerifyAccessToken(cookie.Value)
		require.NoError(t, verifyErr)
		assert.Equal(t, "23", claims.ClinicID)
		assert.Equal(t, []uint64{23, 24}, claims.ClinicIDs)
		return
	}
	t.Fatal("refresh did not issue an access token")
}

func TestHTTPHandler_RefreshToken_MapsOnlyIdentityFailuresToUnauthorized(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	updatedAt := time.Now().Add(-time.Hour)
	account := &model.Account{
		ID:        accountID,
		IsActive:  true,
		UpdatedAt: updatedAt,
	}
	staff := &model.Staff{
		ID:        17,
		AccountID: &accountID,
		IsActive:  true,
	}
	databaseUnavailable := errors.New("database unavailable")

	tests := []struct {
		name         string
		staffError   error
		accountError error
		expectedCode int
	}{
		{
			name:         "missing staff is an authentication failure",
			staffError:   apperrors.WrapNotFound("staff", "17"),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "staff dependency error is a server failure",
			staffError:   databaseUnavailable,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "missing account is an authentication failure",
			accountError: apperrors.WrapNotFound("account", "41"),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "account dependency error is a server failure",
			accountError: databaseUnavailable,
			expectedCode: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{})
			tokens := NewTokenService(refreshSecurityJWTSecret, blacklist)
			issued, err := tokens.IssueRefreshToken(
				staff.ID,
				"2",
				false,
				[]uint64{2},
				updatedAt.UnixNano(),
			)
			require.NoError(t, err)
			handler := NewHTTPHandler(HTTPDependencies{
				Tokens:         tokens,
				TokenBlacklist: blacklist,
				Staff: refreshHTTPStaffReaderStub{
					getByIDFn: func(context.Context, uint64) (*model.Staff, error) {
						if test.staffError != nil {
							return nil, test.staffError
						}
						return staff, nil
					},
				},
				Accounts: refreshHTTPAccountServiceStub{
					getByIDFn: func(context.Context, uint64) (*model.Account, error) {
						if test.accountError != nil {
							return nil, test.accountError
						}
						return account, nil
					},
				},
				StaffAssignments: refreshHTTPAssignmentReaderStub{
					assignments: []model.StaffClinicAssignment{{
						StaffID:  staff.ID,
						ClinicID: 2,
						IsMain:   true,
					}},
				},
				Clinics: refreshHTTPClinicListerStub{},
			}, CookieConfigForProduction(false))

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/refresh",
				http.NoBody,
			)
			c.Request.AddCookie(&http.Cookie{
				Name:  RefreshTokenCookieName,
				Value: issued.Token,
			})
			handler.RefreshToken(c)

			assert.Equal(t, test.expectedCode, recorder.Code, recorder.Body.String())
			for _, cookie := range recorder.Result().Cookies() {
				assert.False(
					t,
					cookie.Name == AccessTokenCookieName && cookie.MaxAge > 0,
					"failed refresh must not issue an access token",
				)
			}
		})
	}
}

func TestHTTPHandler_RefreshToken_RotatesBoundedLegacySession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	issuedAt := time.Now().Truncate(time.Second)
	accountID := uint64(41)
	staff := &model.Staff{ID: 42, AccountID: &accountID, IsActive: true}

	tests := []struct {
		name             string
		accountUpdatedAt time.Time
		expectedCode     int
	}{
		{
			name:             "unchanged account rotates legacy refresh to epoch token",
			accountUpdatedAt: issuedAt.Add(-time.Minute),
			expectedCode:     http.StatusOK,
		},
		{
			name:             "account mutation after issuance rejects legacy refresh",
			accountUpdatedAt: issuedAt.Add(time.Second),
			expectedCode:     http.StatusUnauthorized,
		},
		{
			name:             "same-second account mutation rejects legacy refresh",
			accountUpdatedAt: issuedAt.Add(500 * time.Millisecond),
			expectedCode:     http.StatusUnauthorized,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blacklist := NewTokenBlacklistService(&mockTokenBlacklistRepository{})
			tokens := NewTokenService(testTokenJWTSecret, blacklist)
			legacyToken := signTestJWT(
				t,
				jwt.SigningMethodHS256,
				legacyClaimsAt(refreshTokenSubject, refreshTokenTTL, issuedAt),
			)
			account := &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: test.accountUpdatedAt,
			}
			handler := NewHTTPHandler(HTTPDependencies{
				Tokens:         tokens,
				TokenBlacklist: blacklist,
				Staff: refreshHTTPStaffReaderStub{
					getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
						assert.Equal(t, staff.ID, id)
						return staff, nil
					},
				},
				Accounts: refreshHTTPAccountServiceStub{
					getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
						assert.Equal(t, accountID, id)
						return account, nil
					},
				},
				StaffAssignments: refreshHTTPAssignmentReaderStub{
					assignments: []model.StaffClinicAssignment{{
						StaffID:  staff.ID,
						ClinicID: 7,
						IsMain:   true,
					}},
				},
				Clinics: refreshHTTPClinicListerStub{},
			}, CookieConfigForProduction(false))

			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/refresh",
				http.NoBody,
			)
			c.Request.AddCookie(&http.Cookie{
				Name:  RefreshTokenCookieName,
				Value: legacyToken,
			})
			handler.RefreshToken(c)

			require.Equal(t, test.expectedCode, recorder.Code, recorder.Body.String())
			if test.expectedCode != http.StatusOK {
				return
			}
			for _, cookie := range recorder.Result().Cookies() {
				if cookie.Name != RefreshTokenCookieName ||
					cookie.Path != RefreshTokenCookiePath ||
					cookie.MaxAge <= 0 {
					continue
				}
				claims, err := tokens.ParseRefreshTokenClaims(cookie.Value)
				require.NoError(t, err)
				assert.Equal(t, test.accountUpdatedAt.UnixNano(), claims.AccountEpoch)
				return
			}
			t.Fatal("legacy refresh did not rotate to an epoch-bearing token")
		})
	}
}

func TestHTTPHandler_RefreshToken_RejectsStaleOrInactiveIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	currentUpdatedAt := time.Unix(1_727_123_456, 789_012_345)

	tests := []struct {
		name       string
		tokenEpoch int64
		account    *model.Account
		staff      *model.Staff
	}{
		{
			name:       "password change invalidates old refresh token",
			tokenEpoch: currentUpdatedAt.Add(-time.Second).UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: true},
		},
		{
			name:       "password reset invalidates old refresh token",
			tokenEpoch: currentUpdatedAt.Add(-time.Minute).UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: true},
		},
		{
			name:       "admin demotion invalidates old privileged refresh token",
			tokenEpoch: currentUpdatedAt.Add(-time.Hour).UnixNano(),
			account: &model.Account{
				ID:            accountID,
				IsActive:      true,
				IsSystemAdmin: false,
				UpdatedAt:     currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: true},
		},
		{
			name:       "inactive account is rejected",
			tokenEpoch: currentUpdatedAt.UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  false,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: true},
		},
		{
			name:       "deleted account is rejected",
			tokenEpoch: currentUpdatedAt.UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
				DeletedAt: gorm.DeletedAt{Time: currentUpdatedAt, Valid: true},
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: true},
		},
		{
			name:       "inactive staff is rejected",
			tokenEpoch: currentUpdatedAt.UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, AccountID: &accountID, IsActive: false},
		},
		{
			name:       "deleted staff is rejected",
			tokenEpoch: currentUpdatedAt.UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{
				ID:        17,
				AccountID: &accountID,
				IsActive:  true,
				DeletedAt: gorm.DeletedAt{Time: currentUpdatedAt, Valid: true},
			},
		},
		{
			name:       "staff without linked account is rejected",
			tokenEpoch: currentUpdatedAt.UnixNano(),
			account: &model.Account{
				ID:        accountID,
				IsActive:  true,
				UpdatedAt: currentUpdatedAt,
			},
			staff: &model.Staff{ID: 17, IsActive: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder, _ := executeRefresh(
				t,
				test.account,
				test.staff,
				[]model.StaffClinicAssignment{{StaffID: 17, ClinicID: 2, IsMain: true}},
				test.tokenEpoch,
				true,
			)

			assert.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
			for _, cookie := range recorder.Result().Cookies() {
				assert.False(
					t,
					cookie.Name == AccessTokenCookieName && cookie.MaxAge > 0,
					"rejected refresh must not issue an access token",
				)
			}
		})
	}
}
