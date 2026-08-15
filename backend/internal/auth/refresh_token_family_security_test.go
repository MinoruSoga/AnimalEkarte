package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
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

const refreshFamilyTestSecret = "test-secret-for-refresh-family-security"

type refreshFamilyMemoryBlacklist struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

func newRefreshFamilyMemoryBlacklist() *refreshFamilyMemoryBlacklist {
	return &refreshFamilyMemoryBlacklist{entries: make(map[string]time.Time)}
}

func (r *refreshFamilyMemoryBlacklist) Create(
	_ context.Context,
	entry *model.TokenBlacklist,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.entries[entry.JTI]; exists {
		return apperrors.WrapAlreadyExists("token_blacklist", entry.JTI)
	}
	r.entries[entry.JTI] = entry.ExpiresAt
	return nil
}

func (r *refreshFamilyMemoryBlacklist) ExistsByJTI(
	_ context.Context,
	jti string,
) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	expiresAt, exists := r.entries[jti]
	return exists && expiresAt.After(time.Now()), nil
}

func (*refreshFamilyMemoryBlacklist) DeleteExpired(context.Context) error {
	return nil
}

type refreshFamilyRaceRepository struct {
	rootJTI          string
	familyMarker     string
	familyMarkerErr  error
	familyMarkerSeen bool
}

func (r *refreshFamilyRaceRepository) Create(
	_ context.Context,
	entry *model.TokenBlacklist,
) error {
	switch entry.JTI {
	case r.rootJTI:
		return apperrors.WrapAlreadyExists("token_blacklist", entry.JTI)
	case r.familyMarker:
		r.familyMarkerSeen = true
		return r.familyMarkerErr
	default:
		return nil
	}
}

func (*refreshFamilyRaceRepository) ExistsByJTI(context.Context, string) (bool, error) {
	return false, nil
}

func (*refreshFamilyRaceRepository) DeleteExpired(context.Context) error {
	return nil
}

type blockingRefreshFamilyTokenService struct {
	TokenService
	issueStarted  chan struct{}
	continueIssue chan struct{}
}

func (s *blockingRefreshFamilyTokenService) IssueRefreshTokenInFamily(
	staffID uint64,
	mainClinicID string,
	isSystemAdmin bool,
	clinicIDs []uint64,
	accountEpoch int64,
	familyID string,
	familyExpiresAt time.Time,
) (*IssuedToken, error) {
	close(s.issueStarted)
	<-s.continueIssue
	return s.TokenService.IssueRefreshTokenInFamily(
		staffID,
		mainClinicID,
		isSystemAdmin,
		clinicIDs,
		accountEpoch,
		familyID,
		familyExpiresAt,
	)
}

func newRefreshFamilyHandler(
	t *testing.T,
	blacklist TokenBlacklistService,
	updatedAt time.Time,
) (*HTTPHandler, TokenService, *model.Account, *model.Staff) {
	t.Helper()
	accountID := uint64(41)
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
	tokens := NewTokenService(refreshFamilyTestSecret, blacklist)
	handler := NewHTTPHandler(HTTPDependencies{
		Tokens:         tokens,
		TokenBlacklist: blacklist,
		Accounts: refreshHTTPAccountServiceStub{
			getByIDFn: func(_ context.Context, id uint64) (*model.Account, error) {
				if id != account.ID {
					return nil, apperrors.WrapNotFound("account", "refresh-family")
				}
				return account, nil
			},
		},
		Staff: refreshHTTPStaffReaderStub{
			getByIDFn: func(_ context.Context, id uint64) (*model.Staff, error) {
				if id != staff.ID {
					return nil, apperrors.WrapNotFound("staff", "refresh-family")
				}
				return staff, nil
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
	return handler, tokens, account, staff
}

func postRefreshToken(
	t *testing.T,
	handler *HTTPHandler,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		http.NoBody,
	)
	c.Request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: token,
	})
	handler.RefreshToken(c)
	return recorder
}

func postRefreshLogout(
	t *testing.T,
	handler *HTTPHandler,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh/logout",
		http.NoBody,
	)
	c.Request.AddCookie(&http.Cookie{
		Name:  RefreshTokenCookieName,
		Value: token,
	})
	handler.Logout(c)
	return recorder
}

func currentRefreshCookie(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == RefreshTokenCookieName &&
			cookie.Path == RefreshTokenCookiePath &&
			cookie.MaxAge > 0 {
			return cookie.Value
		}
	}
	t.Fatal("response did not issue a current refresh cookie")
	return ""
}

func refreshFamilyClaim(t *testing.T, tokenString string) string {
	t.Helper()
	claims := jwt.MapClaims{}
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, claims)
	require.NoError(t, err)
	familyID, ok := claims["refresh_family_id"].(string)
	require.True(t, ok, "refresh token must carry a signed family identifier")
	require.NotEmpty(t, familyID)
	return familyID
}

func TestRefreshTokenReuseRevokesAlreadyIssuedDescendant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newRefreshFamilyMemoryBlacklist()
	blacklist := NewTokenBlacklistService(repo)
	updatedAt := time.Now().Add(-time.Hour)
	handler, tokens, account, staff := newRefreshFamilyHandler(t, blacklist, updatedAt)
	root, err := tokens.IssueRefreshToken(
		staff.ID,
		"2",
		false,
		[]uint64{2},
		account.UpdatedAt.UnixNano(),
	)
	require.NoError(t, err)
	rootClaims, err := tokens.ParseRefreshTokenClaims(root.Token)
	require.NoError(t, err)
	assert.Equal(t, rootClaims.ID, refreshFamilyClaim(t, root.Token))

	attacker := postRefreshToken(t, handler, root.Token)
	require.Equal(t, http.StatusOK, attacker.Code, attacker.Body.String())
	firstDescendant := currentRefreshCookie(t, attacker)
	assert.Equal(t, rootClaims.ID, refreshFamilyClaim(t, firstDescendant))

	attackerSecondRotation := postRefreshToken(t, handler, firstDescendant)
	require.Equal(
		t,
		http.StatusOK,
		attackerSecondRotation.Code,
		attackerSecondRotation.Body.String(),
	)
	descendant := currentRefreshCookie(t, attackerSecondRotation)
	assert.Equal(t, rootClaims.ID, refreshFamilyClaim(t, descendant))

	victimReplay := postRefreshToken(t, handler, root.Token)
	require.Equal(t, http.StatusUnauthorized, victimReplay.Code, victimReplay.Body.String())

	descendantAttempt := postRefreshToken(t, handler, descendant)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		descendantAttempt.Code,
		descendantAttempt.Body.String(),
	)
	markerExpiry, exists := repo.entries[refreshFamilyBlacklistKey(rootClaims.ID)]
	require.True(t, exists)
	assert.WithinDuration(t, time.Now().Add(refreshTokenTTL), markerExpiry, 5*time.Second)
}

func TestRefreshTokenLogoutRevokesAlreadyIssuedDescendant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newRefreshFamilyMemoryBlacklist()
	blacklist := NewTokenBlacklistService(repo)
	updatedAt := time.Now().Add(-time.Hour)
	handler, tokens, account, staff := newRefreshFamilyHandler(t, blacklist, updatedAt)
	root, err := tokens.IssueRefreshToken(
		staff.ID,
		"2",
		false,
		[]uint64{2},
		account.UpdatedAt.UnixNano(),
	)
	require.NoError(t, err)

	attacker := postRefreshToken(t, handler, root.Token)
	require.Equal(t, http.StatusOK, attacker.Code, attacker.Body.String())
	descendant := currentRefreshCookie(t, attacker)

	victimLogout := postRefreshLogout(t, handler, root.Token)
	require.Equal(t, http.StatusOK, victimLogout.Code, victimLogout.Body.String())

	descendantAttempt := postRefreshToken(t, handler, descendant)
	assert.Equal(
		t,
		http.StatusUnauthorized,
		descendantAttempt.Code,
		descendantAttempt.Body.String(),
	)
}

func TestLegacyRefreshRotationCarriesRootJTIAsFamily(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newRefreshFamilyMemoryBlacklist()
	blacklist := NewTokenBlacklistService(repo)
	issuedAt := time.Now().Add(-time.Hour).Truncate(time.Second)
	handler, _, account, staff := newRefreshFamilyHandler(
		t,
		blacklist,
		issuedAt.Add(-time.Minute),
	)
	legacyRootClaims := &authjwt.Claims{
		UserID:   "17",
		ClinicID: "2",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "legacy-refresh-root-jti",
			Subject:   refreshTokenSubject,
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
		},
	}
	legacyRoot := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyRootClaims)
	legacyRootString, err := legacyRoot.SignedString([]byte(refreshFamilyTestSecret))
	require.NoError(t, err)
	require.Equal(t, staff.ID, uint64(17))
	require.Equal(t, account.ID, uint64(41))

	attacker := postRefreshToken(t, handler, legacyRootString)
	require.Equal(t, http.StatusOK, attacker.Code, attacker.Body.String())
	descendant := currentRefreshCookie(t, attacker)
	assert.Equal(t, legacyRootClaims.ID, refreshFamilyClaim(t, descendant))

	victimReplay := postRefreshToken(t, handler, legacyRootString)
	require.Equal(t, http.StatusUnauthorized, victimReplay.Code, victimReplay.Body.String())
	descendantAttempt := postRefreshToken(t, handler, descendant)
	assert.Equal(t, http.StatusUnauthorized, descendantAttempt.Code)
}

func TestRefreshTokenConcurrentRotationDuplicateRevokesFamily(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updatedAt := time.Now().Add(-time.Hour)
	bootstrap := NewTokenService(refreshFamilyTestSecret, nil)
	root, err := bootstrap.IssueRefreshToken(17, "2", false, []uint64{2}, updatedAt.UnixNano())
	require.NoError(t, err)
	rootClaims, err := bootstrap.ParseRefreshTokenClaims(root.Token)
	require.NoError(t, err)
	raceRepo := &refreshFamilyRaceRepository{
		rootJTI:      rootClaims.ID,
		familyMarker: refreshFamilyBlacklistKey(rootClaims.ID),
	}
	handler, _, _, _ := newRefreshFamilyHandler(
		t,
		NewTokenBlacklistService(raceRepo),
		updatedAt,
	)

	response := postRefreshToken(t, handler, root.Token)

	assert.Equal(t, http.StatusUnauthorized, response.Code, response.Body.String())
	assert.True(t, raceRepo.familyMarkerSeen)
	assert.Empty(t, response.Header().Values("Set-Cookie"))
}

func TestRefreshTokenConcurrentReuseMarkerOutlivesDelayedWinner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newRefreshFamilyMemoryBlacklist()
	blacklist := NewTokenBlacklistService(repo)
	updatedAt := time.Now().Add(-time.Hour)
	handler, tokens, account, staff := newRefreshFamilyHandler(t, blacklist, updatedAt)
	root, err := tokens.IssueRefreshToken(
		staff.ID,
		"2",
		false,
		[]uint64{2},
		account.UpdatedAt.UnixNano(),
	)
	require.NoError(t, err)
	rootClaims, err := tokens.ParseRefreshTokenClaims(root.Token)
	require.NoError(t, err)

	blockingTokens := &blockingRefreshFamilyTokenService{
		TokenService:  tokens,
		issueStarted:  make(chan struct{}),
		continueIssue: make(chan struct{}),
	}
	handler.deps.Tokens = blockingTokens
	winnerResponse := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		winnerResponse <- postRefreshToken(t, handler, root.Token)
	}()

	<-blockingTokens.issueStarted
	reuse := postRefreshToken(t, handler, root.Token)
	require.Equal(t, http.StatusUnauthorized, reuse.Code, reuse.Body.String())
	// Cross a JWT NumericDate second boundary so the old sliding-expiry
	// implementation cannot accidentally mask the race through truncation.
	time.Sleep(1100 * time.Millisecond)
	close(blockingTokens.continueIssue)
	winner := <-winnerResponse
	require.Equal(t, http.StatusOK, winner.Code, winner.Body.String())

	descendant := currentRefreshCookie(t, winner)
	descendantClaims, err := tokens.ParseRefreshTokenClaims(descendant)
	require.NoError(t, err)
	markerExpiry, exists := repo.entries[refreshFamilyBlacklistKey(rootClaims.ID)]
	require.True(t, exists)
	assert.False(
		t,
		descendantClaims.ExpiresAt.Time.After(markerExpiry),
		"delayed winner must not outlive the concurrent family revocation marker",
	)
}

func TestRefreshTokenFamilyMarkerWriteFailureFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updatedAt := time.Now().Add(-time.Hour)
	bootstrap := NewTokenService(refreshFamilyTestSecret, nil)
	root, err := bootstrap.IssueRefreshToken(17, "2", false, []uint64{2}, updatedAt.UnixNano())
	require.NoError(t, err)
	rootClaims, err := bootstrap.ParseRefreshTokenClaims(root.Token)
	require.NoError(t, err)
	markerFailure := errors.New("family marker unavailable")
	raceRepo := &refreshFamilyRaceRepository{
		rootJTI:         rootClaims.ID,
		familyMarker:    refreshFamilyBlacklistKey(rootClaims.ID),
		familyMarkerErr: markerFailure,
	}
	handler, _, _, _ := newRefreshFamilyHandler(
		t,
		NewTokenBlacklistService(raceRepo),
		updatedAt,
	)

	response := postRefreshToken(t, handler, root.Token)

	assert.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
	assert.True(t, raceRepo.familyMarkerSeen)
	assert.Empty(t, response.Header().Values("Set-Cookie"))
}
