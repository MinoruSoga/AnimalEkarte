package main

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/middleware"
)

type authRouteDependencies struct {
	Audit        audit.Service
	IsProduction bool
	RateLimits   auth.AuthRateLimitConfig
}

func (c authComposition) registerRoutes(
	ctx context.Context,
	api *gin.RouterGroup,
	dependencies authRouteDependencies,
) (*gin.RouterGroup, error) {
	return c.Handler.RegisterRoutes(
		ctx,
		api,
		auth.RouteMiddlewarePorts{
			CSRF: middleware.RequireXRequestedWith(),
			Authenticate: middleware.AuthWithStaffValidationFailureNotifier(
				c.Tokens,
				dependencies.IsProduction,
				dependencies.Audit,
				c.CurrentAccess,
				notifyAuthStaffValidationFailure,
			),
			RateLimitFactory: authRateLimitStoreFactory{},
		},
		dependencies.RateLimits,
	)
}

type authRateLimitStore struct {
	inner *middleware.RateLimitStore
}

func (s authRateLimitStore) Middleware(
	policy auth.RateLimitPolicy,
) gin.HandlerFunc {
	return middleware.RateLimit(
		s.inner,
		rate.Limit(policy.RequestsPerSecond()),
		policy.Burst,
	)
}

type authRateLimitStoreFactory struct{}

func (authRateLimitStoreFactory) New(
	ctx context.Context,
) auth.RateLimitStore {
	return authRateLimitStore{inner: middleware.NewRateLimitStore(ctx)}
}

func notifyAuthStaffValidationFailure(
	ctx context.Context,
	staffID uint64,
	cause error,
) error {
	slog.ErrorContext(
		ctx,
		"auth staff validation failed open",
		"event", "auth_staff_validation_fail_open",
		"staff_id", staffID,
		"error", cause,
	)
	return nil
}
