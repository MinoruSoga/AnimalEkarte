package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
)

func TestNewAuthCompositionBuildsTypedGraph(t *testing.T) {
	repositories := newAuthRepositories(&gorm.DB{})
	composition := newAuthComposition(repositories, authCompositionDependencies{
		Transactor:          persistence.NewTransactor(&gorm.DB{}),
		JWTSecret:           "test-secret",
		PasswordResetConfig: auth.PasswordResetConfig{},
		CookieConfig:        auth.CookieConfigForProduction(false),
	})

	assert.NotNil(t, composition.Handler)
	assert.NotNil(t, composition.Accounts)
	assert.NotNil(t, composition.PermissionGroups)
	assert.NotNil(t, composition.Tokens)
	assert.NotNil(t, composition.PasswordReset)
	assert.NotNil(t, composition.DrainPasswordReset)
}

func TestAuthCompositionRegistersHardenedRouteBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := &gorm.DB{}
	composition := newAuthComposition(
		newAuthRepositories(db),
		authCompositionDependencies{
			Transactor: persistence.NewTransactor(db),
			JWTSecret:  "test-secret",
		},
	)
	router := gin.New()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	protected, err := composition.registerRoutes(
		ctx,
		router.Group("/api/v1"),
		authRouteDependencies{RateLimits: auth.DefaultAuthRateLimitConfig()},
	)

	require.NoError(t, err)
	require.NotNil(t, protected)
	require.Len(t, router.Routes(), 8)
}

func TestNewStaffCompositionBuildsServicesBeforeRBACHandler(t *testing.T) {
	repositories := newStaffRepositories(&gorm.DB{})
	composition := newStaffComposition(repositories, staffCompositionDependencies{
		Transactor: persistence.NewTransactor(&gorm.DB{}),
	})

	assert.NotNil(t, composition.Staff)
	assert.NotNil(t, composition.Assignments)
	assert.NotNil(t, composition.Occupations)
	assert.NotNil(t, composition.ShiftEntries)
	assert.NotNil(t, composition.ShiftTemplates)
	assert.NotNil(t, composition.newHandler(nil))
}

func TestRegisterBaseRoutesOwnsNonDomainHTTPRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	require.NoError(t, registerBaseRoutes(router, nil))

	routes := make(map[string]struct{})
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	assert.Contains(t, routes, http.MethodGet+" /health")
	assert.Contains(t, routes, http.MethodGet+" /api/v1/health")
	assert.Contains(t, routes, http.MethodGet+" /uploads/*filepath")
	assert.Contains(t, routes, http.MethodPost+" /_internal/scheduled-jobs/:jobAction")
}

func TestSMTPConfigAdaptersPreserveDomainFields(t *testing.T) {
	authConfig, err := smtpConfigFromAuth(&auth.PasswordResetConfig{
		SMTPHost: "smtp.auth.test",
		SMTPPort: "587",
		SMTPUser: "auth-user",
		SMTPPass: "auth-pass",
	})
	require.NoError(t, err)
	assert.Equal(t, "smtp.auth.test", authConfig.Host)
	assert.Equal(t, "587", authConfig.Port)
	assert.Equal(t, "auth-user", authConfig.User)
	assert.Equal(t, "auth-pass", authConfig.Pass)

	reservationConfig := smtpConfigFromReservation(reservation.SMTPConfig{
		Host: "smtp.reservation.test",
		Port: "465",
		User: "reservation-user",
		Pass: "reservation-pass",
	})
	assert.Equal(t, "smtp.reservation.test", reservationConfig.Host)
	assert.Equal(t, "465", reservationConfig.Port)
	assert.Equal(t, "reservation-user", reservationConfig.User)
	assert.Equal(t, "reservation-pass", reservationConfig.Pass)
}

func TestSMTPConfigFromAuthRejectsNilConfig(t *testing.T) {
	_, err := smtpConfigFromAuth(nil)
	require.Error(t, err)
}

func TestDrainFunctionIsNilSafe(t *testing.T) {
	drain := nilSafeDrain(nil)
	require.NotNil(t, drain)
	drain()

	called := false
	drain = nilSafeDrain(func() { called = true })
	drain()
	assert.True(t, called)
}
