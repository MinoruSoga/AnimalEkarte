package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/authjwt"
	"github.com/animal-ekarte/backend/internal/model"
)

const sessionTestJWTSecret = "test-secret-for-auth-session-handler"

type sessionBlacklistService struct {
	revocations []string
	revokeError error
}

func (s *sessionBlacklistService) RevokeToken(
	_ context.Context,
	jti string,
	_ time.Time,
) error {
	s.revocations = append(s.revocations, jti)
	return s.revokeError
}

func (*sessionBlacklistService) IsRevoked(context.Context, string) (bool, error) {
	return false, nil
}

func (*sessionBlacklistService) DeleteExpired(context.Context) error {
	return nil
}

type countingSessionTokenService struct {
	TokenService
	parseRefreshTokenClaimsCalls int
}

func (s *countingSessionTokenService) ParseRefreshTokenClaims(
	token string,
) (*authjwt.Claims, error) {
	s.parseRefreshTokenClaimsCalls++
	return s.TokenService.ParseRefreshTokenClaims(token)
}

type sessionAuditLogger struct {
	actions   []string
	staffIDs  []uint64
	clinicIDs []uint64
	err       error
}

func (a *sessionAuditLogger) LogAuthLogin(
	_ context.Context,
	clinicID, staffID *uint64,
	action, _, _ string,
) error {
	a.actions = append(a.actions, action)
	if staffID != nil {
		a.staffIDs = append(a.staffIDs, *staffID)
	}
	if clinicID != nil {
		a.clinicIDs = append(a.clinicIDs, *clinicID)
	}
	return a.err
}

func (*sessionAuditLogger) LogEntry(context.Context, AuthAuditEntry) error {
	return nil
}

type sessionStaffReader struct {
	getByIDFn         func(context.Context, uint64) (*model.Staff, error)
	findByAccountIDFn func(context.Context, uint64) (*model.Staff, error)
}

func (s sessionStaffReader) GetByID(
	ctx context.Context,
	id uint64,
) (*model.Staff, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return &model.Staff{ID: id, IsActive: true}, nil
}

func (s sessionStaffReader) FindByAccountID(
	ctx context.Context,
	accountID uint64,
) (*model.Staff, error) {
	if s.findByAccountIDFn != nil {
		return s.findByAccountIDFn(ctx, accountID)
	}
	return &model.Staff{ID: accountID, IsActive: true}, nil
}

type sessionAccountService struct {
	getByIDFn func(context.Context, uint64) (*model.Account, error)
}

func (sessionAccountService) FindByEmail(
	context.Context,
	string,
) (*model.Account, error) {
	return nil, apperrors.WrapNotFound("account", "lookup")
}

func (s sessionAccountService) GetByID(
	ctx context.Context,
	id uint64,
) (*model.Account, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return &model.Account{ID: id}, nil
}

func (sessionAccountService) UpdatePasswordHash(context.Context, uint64, string) error {
	return nil
}

type sessionAssignmentReader struct {
	assignments []model.StaffClinicAssignment
	err         error
}

func (s sessionAssignmentReader) FindAllByStaffID(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error) {
	return append([]model.StaffClinicAssignment(nil), s.assignments...), s.err
}

type sessionClinicLister struct {
	clinics []model.Clinic
	err     error
}

func (s sessionClinicLister) ListClinics(context.Context) ([]model.Clinic, error) {
	return append([]model.Clinic(nil), s.clinics...), s.err
}

type sessionLoginAuthService struct {
	account *model.Account
	staff   *model.Staff
}

func (s sessionLoginAuthService) AuthenticateUser(
	context.Context,
	string,
	string,
) (*model.Account, *model.Staff, error) {
	return s.account, s.staff, nil
}

func (sessionLoginAuthService) ResolveClinicInfo(
	assignments []model.StaffClinicAssignment,
) (string, []uint64) {
	return ResolveClinicInfo(assignments)
}

func (sessionLoginAuthService) ResolveSystemAdminMainClinicID(
	mainClinicID string,
	isSystemAdmin bool,
	allClinics []model.Clinic,
) string {
	return ResolveSystemAdminMainClinicID(mainClinicID, isSystemAdmin, allClinics)
}

func (sessionLoginAuthService) CalculateEffectivePermissions(
	context.Context,
	bool,
	uint64,
	uint64,
) AuthEffectivePermissions {
	return AuthEffectivePermissions{}
}

func TestHTTPHandler_Login_AuditsResolvedMainClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: sessionLoginAuthService{
			account: &model.Account{
				ID:        accountID,
				Email:     "staff@example.test",
				IsActive:  true,
				UpdatedAt: time.Unix(1_721_000_000, 0),
			},
			staff: &model.Staff{
				ID:        17,
				AccountID: &accountID,
				Name:      "Main Clinic Staff",
				IsActive:  true,
			},
		},
		Tokens: NewTokenService("login-main-audit-secret", nil),
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 10},
				{StaffID: 17, ClinicID: 20, IsMain: true},
			},
		},
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{
				{ID: 10, Name: "First Clinic"},
				{ID: 20, Name: "Main Clinic"},
			},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		strings.NewReader(
			`{"email":"staff@example.test","password":"password1"}`,
		),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, "20", c.GetString("clinic_id"))
	assert.Equal(t, []string{model.AuditActionAuthLoginSuccess}, audit.actions)
	assert.Equal(t, []uint64{17}, audit.staffIDs)
	assert.Equal(t, []uint64{20}, audit.clinicIDs)
}

func TestHTTPHandler_LoginAuditsFallbackClinicForUnassignedSystemAdmin(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Auth: sessionLoginAuthService{
			account: &model.Account{
				ID:            accountID,
				IsActive:      true,
				IsSystemAdmin: true,
				UpdatedAt:     time.Unix(1_721_000_000, 0),
			},
			staff: &model.Staff{
				ID:        17,
				AccountID: &accountID,
				Name:      "System Administrator",
				IsActive:  true,
			},
		},
		Tokens:           NewTokenService("login-system-admin-audit-secret", nil),
		StaffAssignments: sessionAssignmentReader{},
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{{ID: 23, Name: "Fallback Clinic", IsActive: true}},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/login",
		strings.NewReader(
			`{"email":"system-admin@example.test","password":"password1"}`,
		),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Login(c)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, "23", c.GetString("clinic_id"))
	assert.Equal(t, []string{model.AuditActionAuthLoginSuccess}, audit.actions)
	assert.Equal(t, []uint64{17}, audit.staffIDs)
	assert.Equal(t, []uint64{23}, audit.clinicIDs)
}

func TestUniqueRefreshTokenValues_BoundsAndDeduplicates(t *testing.T) {
	tokens, err := UniqueRefreshTokenValues([]*http.Cookie{
		{Name: "other", Value: "ignored"},
		{Name: RefreshTokenCookieName},
		{Name: RefreshTokenCookieName, Value: "one"},
		{Name: RefreshTokenCookieName, Value: "one"},
		{Name: RefreshTokenCookieName, Value: "two"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two"}, tokens)

	tokens, err = UniqueRefreshTokenValues([]*http.Cookie{
		{Name: RefreshTokenCookieName, Value: "one"},
		{Name: RefreshTokenCookieName, Value: "two"},
		{Name: RefreshTokenCookieName, Value: "three"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, tokens)

	_, err = UniqueRefreshTokenValues([]*http.Cookie{{
		Name:  RefreshTokenCookieName,
		Value: strings.Repeat("x", 32<<10),
	}})
	assert.ErrorIs(t, err, apperrors.ErrInvalidInput)
}

func TestHTTPHandler_LogoutRevokesCookiesAndAuditsSignedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := NewTokenService("logout-test-secret", nil)
	first, err := tokens.IssueRefreshToken(17, "23", false, []uint64{23}, 1)
	require.NoError(t, err)
	second, err := tokens.IssueRefreshToken(17, "23", false, []uint64{23}, 1)
	require.NoError(t, err)
	blacklist := &sessionBlacklistService{}
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Audit:          audit,
	}, CookieConfigForProduction(false))

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh/logout", http.NoBody)
	c.Request.AddCookie(&http.Cookie{Name: RefreshTokenCookieName, Value: first.Token})
	c.Request.AddCookie(&http.Cookie{Name: RefreshTokenCookieName, Value: second.Token})
	handler.Logout(c)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Len(t, blacklist.revocations, 2)
	assert.Equal(t, []string{model.AuditActionAuthLogout}, audit.actions)
	assert.Equal(t, []uint64{17}, audit.staffIDs)
	assert.Equal(t, []uint64{23}, audit.clinicIDs)
	expiredCookies := response.Result().Cookies()
	assert.GreaterOrEqual(t, len(expiredCookies), 5)
}

func TestHTTPHandler_LogoutRejectsInvalidCookieSetsAndRevocationFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := NewTokenService("logout-test-secret", nil)
	first, err := tokens.IssueRefreshToken(17, "23", false, []uint64{23}, 1)
	require.NoError(t, err)

	tests := []struct {
		name      string
		cookies   []*http.Cookie
		blacklist TokenBlacklistService
		wantCode  int
	}{
		{
			name: "blacklist is unavailable",
			cookies: []*http.Cookie{
				{Name: RefreshTokenCookieName, Value: first.Token},
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "revocation fails",
			cookies: []*http.Cookie{
				{Name: RefreshTokenCookieName, Value: first.Token},
			},
			blacklist: &sessionBlacklistService{revokeError: errors.New("revoke failed")},
			wantCode:  http.StatusInternalServerError,
		},
		{
			name: "already revoked is idempotent",
			cookies: []*http.Cookie{
				{Name: RefreshTokenCookieName, Value: first.Token},
			},
			blacklist: &sessionBlacklistService{
				revokeError: apperrors.WrapAlreadyExists("token", "jti"),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "invalid and oversized tokens are ignored",
			cookies: []*http.Cookie{
				{Name: RefreshTokenCookieName, Value: "not-a-token"},
				{
					Name:  RefreshTokenCookieName,
					Value: strings.Repeat("x", maxRefreshTokenCookieBytes+1),
				},
			},
			blacklist: &sessionBlacklistService{},
			wantCode:  http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHTTPHandler(HTTPDependencies{
				Tokens:         tokens,
				TokenBlacklist: test.blacklist,
			}, CookieConfigForProduction(false))
			response := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(response)
			c.Request = httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/refresh/logout",
				http.NoBody,
			)
			for _, cookie := range test.cookies {
				c.Request.AddCookie(cookie)
			}
			handler.Logout(c)
			assert.Equal(t, test.wantCode, response.Code, response.Body.String())
		})
	}
}

func TestLogout_MixedCookieIdentitiesRevokeEveryValidFamilyAndSkipAudit(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	tokens := NewTokenService("logout-mixed-cookie-secret", nil)
	first, err := tokens.IssueRefreshToken(17, "23", false, []uint64{23}, 1)
	require.NoError(t, err)
	second, err := tokens.IssueRefreshToken(17, "23", false, []uint64{23}, 1)
	require.NoError(t, err)
	conflicting, err := tokens.IssueRefreshToken(18, "24", false, []uint64{24}, 1)
	require.NoError(t, err)
	blacklist := &sessionBlacklistService{}
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Audit:          audit,
	}, CookieConfigForProduction(false))
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh/logout",
		http.NoBody,
	)
	for _, value := range []string{
		"attacker-malformed-cookie",
		first.Token,
		second.Token,
		conflicting.Token,
	} {
		c.Request.AddCookie(&http.Cookie{
			Name:  RefreshTokenCookieName,
			Value: value,
		})
	}

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.ElementsMatch(t, []string{
		issuedRefreshFamilyMarker(t, tokens, first),
		issuedRefreshFamilyMarker(t, tokens, second),
		issuedRefreshFamilyMarker(t, tokens, conflicting),
	}, blacklist.revocations)
	assert.Empty(t, audit.actions)
	assertLogoutKnownPathsCleared(t, response)
}

func issuedRefreshFamilyMarker(
	t *testing.T,
	tokens TokenService,
	issued *IssuedToken,
) string {
	t.Helper()
	claims, err := tokens.ParseRefreshTokenClaims(issued.Token)
	require.NoError(t, err)
	familyID, valid := refreshTokenFamilyID(claims)
	require.True(t, valid)
	return refreshFamilyBlacklistKey(familyID)
}

func assertLogoutKnownPathsCleared(
	t *testing.T,
	response *httptest.ResponseRecorder,
) {
	t.Helper()
	cleared := make(map[string]bool)
	for _, cookie := range response.Result().Cookies() {
		if cookie.MaxAge < 0 {
			cleared[cookie.Name+"|"+cookie.Path] = true
		}
	}
	for _, key := range []string{
		AccessTokenCookieName + "|/",
		LegacyTokenCookieName + "|/",
		RefreshTokenCookieName + "|" + RefreshTokenCookiePath,
		RefreshTokenCookieName + "|" + LegacyRefreshTokenCookiePath,
		"prev_clinic_id|/",
	} {
		assert.True(t, cleared[key], "logout must clear %s", key)
	}
}

func buildSessionTestRefreshToken(t *testing.T, jti string) string {
	t.Helper()
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &authjwt.Claims{
		UserID:       "10",
		ClinicID:     "1",
		AccountEpoch: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   refreshTokenSubject,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	signed, err := token.SignedString([]byte(sessionTestJWTSecret))
	require.NoError(t, err)
	return signed
}

func TestLogout_CookieFanoutRevokesValidFamilyAndClearsKnownPaths(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	blacklist := &sessionBlacklistService{}
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         NewTokenService(sessionTestJWTSecret, nil),
		TokenBlacklist: blacklist,
		Audit:          audit,
	}, CookieConfigForProduction(false))

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh/logout",
		http.NoBody,
	)
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: strings.Repeat("a", maxRefreshTokenCookieBytes+1),
	})
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: "attacker-malformed-cookie",
	})
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: buildSessionTestRefreshToken(t, "legitimate-jti"),
	})
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = request

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(
		t,
		[]string{refreshFamilyBlacklistKey("legitimate-jti")},
		blacklist.revocations,
	)
	assert.Equal(t, []string{model.AuditActionAuthLogout}, audit.actions)
	assertLogoutKnownPathsCleared(t, response)
}

func TestLogout_CookieTossingRevokesLegitimateFamiliesAndClearsKnownPaths(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	blacklist := &sessionBlacklistService{}
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         NewTokenService(sessionTestJWTSecret, nil),
		TokenBlacklist: blacklist,
		Audit:          &sessionAuditLogger{},
	}, CookieConfigForProduction(false))
	router := gin.New()
	router.POST("/api/v1/auth/refresh/logout", handler.Logout)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	logoutURL, err := url.Parse(server.URL + "/api/v1/auth/refresh/logout")
	require.NoError(t, err)
	currentToken := buildSessionTestRefreshToken(t, "current-jti")
	legacyToken := buildSessionTestRefreshToken(t, "legacy-jti")
	jar.SetCookies(logoutURL, []*http.Cookie{
		{
			Name:  RefreshTokenCookieName,
			Value: currentToken,
			Path:  RefreshTokenCookiePath,
		},
		{
			Name:  RefreshTokenCookieName,
			Value: legacyToken,
			Path:  LegacyRefreshTokenCookiePath,
		},
		{
			Name:  RefreshTokenCookieName,
			Value: "attacker-malformed-cookie",
			Path:  "/api/v1/auth/refresh/logout",
		},
	})
	client := &http.Client{Jar: jar}
	request, err := http.NewRequest(
		http.MethodPost,
		logoutURL.String(),
		http.NoBody,
	)
	require.NoError(t, err)
	response, err := client.Do(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.ElementsMatch(
		t,
		[]string{
			refreshFamilyBlacklistKey("current-jti"),
			refreshFamilyBlacklistKey("legacy-jti"),
		},
		blacklist.revocations,
	)
	for _, cookie := range jar.Cookies(logoutURL) {
		assert.NotEqual(t, currentToken, cookie.Value)
		assert.NotEqual(t, legacyToken, cookie.Value)
	}
}

func TestLogout_RejectsOversizedRefreshCookieBeforeJWTVerification(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	blacklist := &sessionBlacklistService{}
	tokens := &countingSessionTokenService{
		TokenService: NewTokenService(sessionTestJWTSecret, nil),
	}
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Audit:          &sessionAuditLogger{},
	}, CookieConfigForProduction(false))

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh/logout",
		http.NoBody,
	)
	request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: strings.Repeat("a", maxRefreshTokenCookieBytes+1),
	})
	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = request

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Empty(t, blacklist.revocations)
	assert.Zero(t, tokens.parseRefreshTokenClaimsCalls)
}

func TestHTTPHandler_GetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	accountID := uint64(41)
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: sessionStaffReader{
			getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				return &model.Staff{ID: id, Name: "Session Staff", AccountID: &accountID}, nil
			},
		},
		Accounts: sessionAccountService{
			getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
				return &model.Account{ID: id, Email: "staff@example.test"}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{{
				StaffID:  17,
				ClinicID: 23,
				IsMain:   true,
			}},
		},
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{{ID: 23, Name: "Main Clinic"}},
		},
		EffectivePermissions: permissionHTTPEffectiveService{
			rules: []model.PermissionGroupRule{{
				Resource: "owners",
				CanView:  true,
			}},
		},
	}, CookieConfigForProduction(false))

	response := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(response)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", http.NoBody)
	c.Set("user_id", "17")
	c.Set("clinic_id", "23")
	handler.GetMe(c)

	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Contains(t, response.Body.String(), `"display_name":"Session Staff"`)
	assert.Contains(t, response.Body.String(), `"owners":{"view":true`)

	missingResponse := httptest.NewRecorder()
	missing, _ := gin.CreateTestContext(missingResponse)
	missing.Request = httptest.NewRequest(http.MethodGet, "/api/v1/me", http.NoBody)
	handler.GetMe(missing)
	assert.Equal(t, http.StatusUnauthorized, missingResponse.Code)
}

func TestAuthPermissionMappingAndAuditHelpers(t *testing.T) {
	effective := permissionHTTPEffectiveService{rules: []model.PermissionGroupRule{{
		Resource:  "owners",
		CanView:   true,
		CanCreate: true,
	}}}
	service := NewAuthService(nil, nil, effective)
	admin := service.CalculateEffectivePermissions(context.Background(), true, 17, 0)
	assert.Len(t, admin, len(model.AllResources))
	assert.True(t, admin[string(model.ResourceOwners)].Delete)

	regular := service.CalculateEffectivePermissions(context.Background(), false, 17, 23)
	assert.True(t, regular["owners"].View)
	assert.True(t, regular["owners"].Create)
	assert.False(t, regular["owners"].Delete)

	failed := NewAuthService(
		nil,
		nil,
		permissionHTTPEffectiveService{err: errors.New("unavailable")},
	).CalculateEffectivePermissions(context.Background(), false, 17, 23)
	assert.Empty(t, failed)

	httpPermissions := BuildAllPermissions()
	assert.Len(t, httpPermissions, len(model.AllResources))
	assert.Empty(t, ToHTTPEffectivePermissions(nil))

	assignments := []model.StaffClinicAssignment{
		{ClinicID: 23},
		{ClinicID: 24, IsMain: true},
	}
	clinicID, ok := AuditClinicIDFromAssignments(assignments)
	assert.True(t, ok)
	assert.Equal(t, uint64(24), clinicID)
	clinicID, ok = AuditClinicIDFromAssignments(assignments[:1])
	assert.True(t, ok)
	assert.Equal(t, uint64(23), clinicID)
	_, ok = AuditClinicIDFromAssignments(nil)
	assert.False(t, ok)
}

func TestHTTPHandler_AuditKnownAccountLoginFailure(t *testing.T) {
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Staff: sessionStaffReader{
			findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
				assert.Equal(t, uint64(41), accountID)
				return &model.Staff{ID: 17}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{
			assignments: []model.StaffClinicAssignment{{
				StaffID: 17, ClinicID: 23, IsMain: true,
			}},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	handler.AuditKnownAccountLoginFailure(
		context.Background(),
		41,
		"127.0.0.1",
		"test-agent",
	)
	assert.Equal(t, []string{model.AuditActionAuthLoginFailure}, audit.actions)

	noDependencies := NewHTTPHandler(HTTPDependencies{}, CookieConfigForProduction(false))
	noDependencies.AuditKnownAccountLoginFailure(context.Background(), 41, "", "")

	staffError := NewHTTPHandler(HTTPDependencies{
		Staff: sessionStaffReader{
			findByAccountIDFn: func(context.Context, uint64) (*model.Staff, error) {
				return nil, errors.New("staff unavailable")
			},
		},
		StaffAssignments: sessionAssignmentReader{},
		Audit:            audit,
	}, CookieConfigForProduction(false))
	staffError.AuditKnownAccountLoginFailure(context.Background(), 41, "", "")

	assignmentError := NewHTTPHandler(HTTPDependencies{
		Staff: sessionStaffReader{},
		StaffAssignments: sessionAssignmentReader{
			err: errors.New("assignments unavailable"),
		},
		Audit: audit,
	}, CookieConfigForProduction(false))
	assignmentError.AuditKnownAccountLoginFailure(context.Background(), 41, "", "")

	noAssignments := NewHTTPHandler(HTTPDependencies{
		Staff:            sessionStaffReader{},
		StaffAssignments: sessionAssignmentReader{},
		Audit:            audit,
	}, CookieConfigForProduction(false))
	noAssignments.AuditKnownAccountLoginFailure(context.Background(), 41, "", "")
}

func TestHTTPHandler_AuditKnownSystemAdminLoginFailureUsesFirstActiveClinic(
	t *testing.T,
) {
	audit := &sessionAuditLogger{}
	handler := NewHTTPHandler(HTTPDependencies{
		Accounts: sessionAccountService{
			getByIDFn: func(_ context.Context, accountID uint64) (*model.Account, error) {
				return &model.Account{
					ID:            accountID,
					IsActive:      true,
					IsSystemAdmin: true,
				}, nil
			},
		},
		Staff: sessionStaffReader{
			findByAccountIDFn: func(_ context.Context, accountID uint64) (*model.Staff, error) {
				assert.Equal(t, uint64(41), accountID)
				return &model.Staff{ID: 17, AccountID: &accountID}, nil
			},
		},
		StaffAssignments: sessionAssignmentReader{},
		Clinics: sessionClinicLister{
			clinics: []model.Clinic{
				{ID: 22, Name: "Inactive Clinic", IsActive: false},
				{ID: 23, Name: "Active Clinic", IsActive: true},
			},
		},
		Audit: audit,
	}, CookieConfigForProduction(false))

	handler.AuditKnownAccountLoginFailure(
		context.Background(),
		41,
		"192.0.2.1",
		"system-admin-failure-test",
	)

	assert.Equal(t, []string{model.AuditActionAuthLoginFailure}, audit.actions)
	assert.Equal(t, []uint64{17}, audit.staffIDs)
	assert.Equal(t, []uint64{23}, audit.clinicIDs)
}
