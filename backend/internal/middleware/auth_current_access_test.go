package middleware

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/authjwt"
	"github.com/animal-ekarte/backend/internal/model"
)

type middlewareCurrentAccessResolver struct {
	access    *authdomain.CurrentAccess
	err       error
	staffIDs  []uint64
	resolveFn func(context.Context, uint64) (*authdomain.CurrentAccess, error)
}

type nilClaimsTokenService struct {
	authdomain.TokenService
}

func (nilClaimsTokenService) VerifyAccessToken(string) (*authjwt.Claims, error) {
	return nil, nil
}

type currentAccessAccountReaderFunc func(
	context.Context,
	uint64,
) (*model.Account, error)

func (f currentAccessAccountReaderFunc) GetByID(
	ctx context.Context,
	accountID uint64,
) (*model.Account, error) {
	return f(ctx, accountID)
}

type currentAccessAssignmentReaderFunc func(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error)

func (f currentAccessAssignmentReaderFunc) FindAllByStaffID(
	ctx context.Context,
	staffID uint64,
) ([]model.StaffClinicAssignment, error) {
	return f(ctx, staffID)
}

type currentAccessClinicReaderFunc func(
	context.Context,
) ([]model.Clinic, error)

func (f currentAccessClinicReaderFunc) ListClinics(
	ctx context.Context,
) ([]model.Clinic, error) {
	return f(ctx)
}

func (r *middlewareCurrentAccessResolver) Resolve(
	ctx context.Context,
	staffID uint64,
) (*authdomain.CurrentAccess, error) {
	r.staffIDs = append(r.staffIDs, staffID)
	if r.resolveFn != nil {
		return r.resolveFn(ctx, staffID)
	}
	return r.access, r.err
}

type currentAccessRequestResult struct {
	response         *httptest.ResponseRecorder
	downstreamCalled bool
	clinicID         string
	clinicIDs        []uint64
	isSystemAdmin    bool
}

func runCurrentAccessRequest(
	t *testing.T,
	claims jwt.MapClaims,
	headerClinicID string,
	resolver authdomain.CurrentAccessResolver,
) currentAccessRequestResult {
	t.Helper()
	token := makeToken(t, jwt.SigningMethodHS256, claims)
	response := httptest.NewRecorder()
	result := currentAccessRequestResult{response: response}
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, nil, resolver))
	router.GET("/test", func(c *gin.Context) {
		result.downstreamCalled = true
		result.clinicID = c.GetString("clinic_id")
		result.isSystemAdmin = c.GetBool("is_system_admin")
		if currentClinicIDs, ok := c.Get("clinic_ids"); ok {
			result.clinicIDs = append(
				[]uint64(nil),
				currentClinicIDs.([]uint64)...,
			)
		}
		c.Status(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	request.Header.Set("Authorization", "Bearer "+token)
	if headerClinicID != "" {
		request.Header.Set("X-Clinic-ID", headerClinicID)
	}
	router.ServeHTTP(response, request)
	return result
}

func currentAccessClaims(
	defaultClinicID string,
	clinicIDs []uint64,
	isSystemAdmin bool,
	accountEpoch int64,
) jwt.MapClaims {
	now := time.Now()
	return jwt.MapClaims{
		"user_id":         "123",
		"clinic_id":       defaultClinicID,
		"clinic_ids":      clinicIDs,
		"is_system_admin": isSystemAdmin,
		"account_epoch":   accountEpoch,
		"jti":             "middleware-current-access-token",
		"iat":             now.Unix(),
		"exp":             now.Add(15 * time.Minute).Unix(),
	}
}

func legacyCurrentAccessClaims(issuedAt time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":         "123",
		"clinic_id":       "1",
		"clinic_ids":      []uint64{1},
		"is_system_admin": false,
		"iat":             issuedAt.Unix(),
		"jti":             "middleware-legacy-current-access-token",
		"exp":             issuedAt.Add(15 * time.Minute).Unix(),
	}
}

func TestAuth_HealthyRequestUsesOnlyCurrentAccessState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resolver := &middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
		StaffID:       123,
		AccountEpoch:  700,
		IsSystemAdmin: false,
		ClinicIDs:     []uint64{2, 3},
	}}

	result := runCurrentAccessRequest(
		t,
		currentAccessClaims("3", []uint64{1, 3}, true, 700),
		"",
		resolver,
	)

	assert.Equal(t, http.StatusOK, result.response.Code)
	assert.True(t, result.downstreamCalled)
	assert.Equal(t, "3", result.clinicID)
	assert.False(t, result.isSystemAdmin)
	assert.Equal(t, []uint64{2, 3}, result.clinicIDs)
	assert.Equal(t, []uint64{123}, resolver.staffIDs)
}

func TestAuth_FailsClosedWhenTokenVerifierReturnsNilClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(
		nilClaimsTokenService{},
		false,
		nil,
		&middlewareCurrentAccessResolver{},
	))
	downstreamCalled := false
	router.GET("/test", func(c *gin.Context) {
		downstreamCalled = true
		c.Status(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	request.Header.Set("Authorization", "Bearer verifier-returned-nil")

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.False(t, downstreamCalled)
}

func TestAuth_FailsClosedWhenTokenVerifierIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(
		nil,
		false,
		nil,
		&middlewareCurrentAccessResolver{},
	))
	downstreamCalled := false
	router.GET("/test", func(c *gin.Context) {
		downstreamCalled = true
		c.Status(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	request.Header.Set("Authorization", "Bearer verifier-is-missing")

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.False(t, downstreamCalled)
}

func TestAuth_RejectsRemovedDefaultAndHeaderClinicAssignments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		defaultClinic  string
		headerClinic   string
		signedClinics  []uint64
		currentClinics []uint64
	}{
		{
			name:           "removed signed default clinic",
			defaultClinic:  "1",
			signedClinics:  []uint64{1, 2},
			currentClinics: []uint64{2},
		},
		{
			name:           "removed header clinic",
			defaultClinic:  "1",
			headerClinic:   "2",
			signedClinics:  []uint64{1, 2},
			currentClinics: []uint64{1},
		},
		{
			name:           "regular staff has no active assignments",
			defaultClinic:  "1",
			signedClinics:  []uint64{1},
			currentClinics: []uint64{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := runCurrentAccessRequest(
				t,
				currentAccessClaims(
					test.defaultClinic,
					test.signedClinics,
					false,
					700,
				),
				test.headerClinic,
				&middlewareCurrentAccessResolver{
					access: &authdomain.CurrentAccess{
						StaffID:       123,
						AccountEpoch:  700,
						IsSystemAdmin: false,
						ClinicIDs:     test.currentClinics,
					},
				},
			)

			assert.Equal(t, http.StatusForbidden, result.response.Code)
			assert.False(t, result.downstreamCalled)
			for _, cookie := range result.response.Result().Cookies() {
				assert.NotEqual(t, "prev_clinic_id", cookie.Name)
			}
		})
	}
}

func TestAuth_RegularStaffRejectsInactiveCurrentClinic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resolver := authdomain.NewCurrentAccessResolverWithClinics(
		&mockCurrentAccessStaffReader{},
		currentAccessAccountReaderFunc(func(
			context.Context,
			uint64,
		) (*model.Account, error) {
			return &model.Account{
				ID:        41,
				IsActive:  true,
				UpdatedAt: time.Unix(0, 700),
			}, nil
		}),
		currentAccessAssignmentReaderFunc(func(
			context.Context,
			uint64,
		) ([]model.StaffClinicAssignment, error) {
			return []model.StaffClinicAssignment{
				{StaffID: 123, ClinicID: 1, IsMain: true},
				{StaffID: 123, ClinicID: 2},
			}, nil
		}),
		currentAccessClinicReaderFunc(func(
			context.Context,
		) ([]model.Clinic, error) {
			return []model.Clinic{
				{ID: 1, IsActive: false},
				{ID: 2, IsActive: true},
			}, nil
		}),
	)

	for _, test := range []struct {
		name           string
		headerClinicID string
		wantStatus     int
		wantDownstream bool
	}{
		{
			name:       "inactive signed default is rejected",
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "inactive header clinic is rejected",
			headerClinicID: "1",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "active header clinic is accepted",
			headerClinicID: "2",
			wantStatus:     http.StatusOK,
			wantDownstream: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := runCurrentAccessRequest(
				t,
				currentAccessClaims("1", []uint64{1, 2}, false, 700),
				test.headerClinicID,
				resolver,
			)

			assert.Equal(t, test.wantStatus, result.response.Code)
			assert.Equal(t, test.wantDownstream, result.downstreamCalled)
		})
	}
}

func TestAuth_CurrentAdminStateControlsGlobalBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin demotion removes stale global bypass", func(t *testing.T) {
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, true, 700),
			"99",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  700,
				IsSystemAdmin: false,
				ClinicIDs:     []uint64{1},
			}},
		)

		assert.Equal(t, http.StatusForbidden, result.response.Code)
		assert.False(t, result.downstreamCalled)
	})

	t.Run("current admin cannot select a clinic outside current active authority", func(t *testing.T) {
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, false, 700),
			"99",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  700,
				IsSystemAdmin: true,
				MainClinicID:  "1",
				ClinicIDs:     []uint64{1},
			}},
		)

		assert.Equal(t, http.StatusForbidden, result.response.Code)
		assert.False(t, result.downstreamCalled)
	})

	t.Run("current admin can select a clinic in current active authority", func(t *testing.T) {
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, false, 700),
			"99",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  700,
				IsSystemAdmin: true,
				MainClinicID:  "1",
				ClinicIDs:     []uint64{1, 99},
			}},
		)

		assert.Equal(t, http.StatusOK, result.response.Code)
		assert.True(t, result.downstreamCalled)
		assert.True(t, result.isSystemAdmin)
		assert.Equal(t, "99", result.clinicID)
	})

	t.Run("current admin cannot select zero clinic", func(t *testing.T) {
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, false, 700),
			"0",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  700,
				IsSystemAdmin: true,
				MainClinicID:  "1",
				ClinicIDs:     []uint64{1},
			}},
		)

		assert.Equal(t, http.StatusBadRequest, result.response.Code)
		assert.False(t, result.downstreamCalled)
	})
}

func TestAuth_CurrentSystemAdminClinicIdentityReplacesStaleTokenDefault(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)
	result := runCurrentAccessRequest(
		t,
		currentAccessClaims("99", []uint64{99}, true, 700),
		"",
		&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
			StaffID:       123,
			AccountEpoch:  700,
			IsSystemAdmin: true,
			MainClinicID:  "1",
			ClinicIDs:     []uint64{1, 2},
		}},
	)

	assert.Equal(t, http.StatusOK, result.response.Code)
	assert.True(t, result.downstreamCalled)
	assert.Equal(t, "1", result.clinicID)
	assert.Equal(t, []uint64{1, 2}, result.clinicIDs)
}

func TestAuth_FirstSystemAdminClinicOverrideIsAudited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &mockMiddlewareAuditService{}
	resolver := &middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
		StaffID:       123,
		AccountEpoch:  700,
		IsSystemAdmin: true,
		MainClinicID:  "1",
		ClinicIDs:     []uint64{1, 2},
	}}
	token := makeToken(
		t,
		jwt.SigningMethodHS256,
		currentAccessClaims("1", []uint64{1}, false, 700),
	)
	response := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testTokenSvc(), false, spy, resolver))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Clinic-ID", "2")

	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	if assert.NotNil(t, spy.logClinicSwitchCall) {
		assert.Equal(t, uint64(1), spy.logClinicSwitchCall.fromClinicID)
		assert.Equal(t, uint64(2), spy.logClinicSwitchCall.toClinicID)
	}
}

func TestAuth_PO005SystemAdminFallbackFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resolver := authdomain.NewCurrentAccessResolver(
		&mockCurrentAccessStaffReader{
			findFn: func(
				context.Context,
				uint64,
			) (*authdomain.CurrentAccessStaffIdentity, error) {
				return nil, driver.ErrBadConn
			},
		},
		middlewareCurrentAccessAccountReader{},
		middlewareCurrentAccessAssignmentReader{},
	)

	for _, test := range []struct {
		name           string
		headerClinicID string
	}{
		{
			name:           "denies signed active clinic scope on temporary staff lookup",
			headerClinicID: "1",
		},
		{
			name:           "denies arbitrary global clinic on temporary staff lookup",
			headerClinicID: "2",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			notifierCalls := 0
			token := makeToken(
				t,
				jwt.SigningMethodHS256,
				currentAccessClaims("1", []uint64{1}, true, 700),
			)
			response := httptest.NewRecorder()
			router := gin.New()
			router.Use(AuthWithStaffValidationFailureNotifier(
				testTokenSvc(),
				false,
				nil,
				resolver,
				func(context.Context, uint64, error) error {
					notifierCalls++
					return nil
				},
			))
			downstreamCalled := false
			router.GET("/test", func(c *gin.Context) {
				downstreamCalled = true
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("X-Clinic-ID", test.headerClinicID)

			router.ServeHTTP(response, request)

			assert.Equal(t, http.StatusServiceUnavailable, response.Code)
			assert.False(t, downstreamCalled)
			assert.Equal(t, 1, notifierCalls)
			require.Contains(t, response.Body.String(), "access validation unavailable")
		})
	}
}

func TestAuth_RejectsAccountEpochMismatchAndResolverFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("epoch mismatch invalidates signed access token", func(t *testing.T) {
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, false, 699),
			"",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  700,
				IsSystemAdmin: false,
				ClinicIDs:     []uint64{1},
			}},
		)

		assert.Equal(t, http.StatusUnauthorized, result.response.Code)
		assert.False(t, result.downstreamCalled)
	})

	t.Run("missing epoch without bounded legacy timestamps is rejected before resolution", func(t *testing.T) {
		resolver := &middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
			StaffID:       123,
			AccountEpoch:  700,
			IsSystemAdmin: false,
			ClinicIDs:     []uint64{1},
		}}
		claims := currentAccessClaims("1", []uint64{1}, false, 0)
		delete(claims, "iat")
		result := runCurrentAccessRequest(
			t,
			claims,
			"",
			resolver,
		)

		assert.Equal(t, http.StatusUnauthorized, result.response.Code)
		assert.False(t, result.downstreamCalled)
		assert.Empty(t, resolver.staffIDs)
	})

	t.Run("bounded legacy token continues when account has not changed", func(t *testing.T) {
		issuedAt := time.Now().Truncate(time.Second)
		resolver := &middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
			StaffID:       123,
			AccountEpoch:  issuedAt.Add(-time.Minute).UnixNano(),
			IsSystemAdmin: false,
			ClinicIDs:     []uint64{1},
		}}
		result := runCurrentAccessRequest(
			t,
			legacyCurrentAccessClaims(issuedAt),
			"",
			resolver,
		)

		assert.Equal(t, http.StatusOK, result.response.Code)
		assert.True(t, result.downstreamCalled)
		assert.Equal(t, []uint64{123}, resolver.staffIDs)
	})

	t.Run("bounded legacy token is rejected after account mutation", func(t *testing.T) {
		issuedAt := time.Now().Truncate(time.Second)
		result := runCurrentAccessRequest(
			t,
			legacyCurrentAccessClaims(issuedAt),
			"",
			&middlewareCurrentAccessResolver{access: &authdomain.CurrentAccess{
				StaffID:       123,
				AccountEpoch:  issuedAt.Add(time.Second).UnixNano(),
				IsSystemAdmin: false,
				ClinicIDs:     []uint64{1},
			}},
		)

		assert.Equal(t, http.StatusUnauthorized, result.response.Code)
		assert.False(t, result.downstreamCalled)
	})

	t.Run("account or assignment resolution error fails closed", func(t *testing.T) {
		resolverFailure := errors.New("current access unavailable")
		result := runCurrentAccessRequest(
			t,
			currentAccessClaims("1", []uint64{1}, false, 700),
			"",
			&middlewareCurrentAccessResolver{err: resolverFailure},
		)

		assert.Equal(t, http.StatusServiceUnavailable, result.response.Code)
		assert.False(t, result.downstreamCalled)
		require.Contains(t, result.response.Body.String(), "access validation unavailable")
	})
}

func TestAuth_AccountAndAssignmentFailuresNeverUsePO005(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name          string
		accountErr    error
		assignmentErr error
	}{
		{
			name:       "temporary account read error",
			accountErr: driver.ErrBadConn,
		},
		{
			name:          "temporary assignment read error",
			assignmentErr: driver.ErrBadConn,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := authdomain.NewCurrentAccessResolver(
				&mockCurrentAccessStaffReader{},
				currentAccessAccountReaderFunc(func(
					context.Context,
					uint64,
				) (*model.Account, error) {
					if test.accountErr != nil {
						return nil, test.accountErr
					}
					return &model.Account{
						ID:        41,
						IsActive:  true,
						UpdatedAt: time.Unix(0, 700),
					}, nil
				}),
				currentAccessAssignmentReaderFunc(func(
					context.Context,
					uint64,
				) ([]model.StaffClinicAssignment, error) {
					if test.assignmentErr != nil {
						return nil, test.assignmentErr
					}
					return []model.StaffClinicAssignment{{
						StaffID:  123,
						ClinicID: 1,
						IsMain:   true,
					}}, nil
				}),
			)
			notifierCalls := 0
			token := makeToken(
				t,
				jwt.SigningMethodHS256,
				currentAccessClaims("1", []uint64{1}, false, 700),
			)
			response := httptest.NewRecorder()
			router := gin.New()
			router.Use(AuthWithStaffValidationFailureNotifier(
				testTokenSvc(),
				false,
				nil,
				resolver,
				func(context.Context, uint64, error) error {
					notifierCalls++
					return nil
				},
			))
			downstreamCalled := false
			router.GET("/test", func(c *gin.Context) {
				downstreamCalled = true
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(
				http.MethodGet,
				"/test",
				http.NoBody,
			)
			request.Header.Set("Authorization", "Bearer "+token)

			router.ServeHTTP(response, request)

			assert.Equal(t, http.StatusServiceUnavailable, response.Code)
			assert.False(t, downstreamCalled)
			assert.Zero(t, notifierCalls)
		})
	}
}
