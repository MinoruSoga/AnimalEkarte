package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the complete auth public/protected route surface.
// api must be the /api/v1 group. The returned group is the authenticated API
// group used by all other domain route registrars.
func (h *HTTPHandler) RegisterRoutes(
	ctx context.Context,
	api *gin.RouterGroup,
	middlewarePorts RouteMiddlewarePorts,
	rateLimits AuthRateLimitConfig,
) (*gin.RouterGroup, error) {
	if api == nil {
		return nil, fmt.Errorf("auth route group is required")
	}
	if middlewarePorts.CSRF == nil {
		return nil, fmt.Errorf("auth CSRF middleware is required")
	}
	if middlewarePorts.Authenticate == nil {
		return nil, fmt.Errorf("auth authentication middleware is required")
	}
	if middlewarePorts.RateLimitFactory == nil {
		return nil, fmt.Errorf("auth rate-limit factory is required")
	}
	if err := validateRateLimitConfig(rateLimits); err != nil {
		return nil, err
	}

	loginRateStore := middlewarePorts.RateLimitFactory.New(ctx)
	passwordRateStore := middlewarePorts.RateLimitFactory.New(ctx)
	refreshRateStore := middlewarePorts.RateLimitFactory.New(ctx)
	logoutRateStore := middlewarePorts.RateLimitFactory.New(ctx)
	logoutRedirectRateStore := middlewarePorts.RateLimitFactory.New(ctx)
	if loginRateStore == nil ||
		passwordRateStore == nil ||
		refreshRateStore == nil ||
		logoutRateStore == nil ||
		logoutRedirectRateStore == nil {
		return nil, fmt.Errorf("auth rate-limit store creation failed")
	}

	api.POST(
		"/login",
		middlewarePorts.CSRF,
		loginRateStore.Middleware(rateLimits.Login),
		h.Login,
	)
	api.POST(
		"/logout",
		middlewarePorts.CSRF,
		logoutRedirectRateStore.Middleware(rateLimits.LogoutRedirect),
		h.RedirectLogout,
	)
	api.POST(
		"/auth/refresh/logout",
		middlewarePorts.CSRF,
		logoutRateStore.Middleware(rateLimits.Logout),
		h.Logout,
	)
	api.POST(
		"/auth/refresh",
		middlewarePorts.CSRF,
		refreshRateStore.Middleware(rateLimits.Refresh),
		h.RefreshToken,
	)
	api.POST(
		"/auth/forgot-password",
		passwordRateStore.Middleware(rateLimits.PasswordReset),
		h.ForgotPassword,
	)
	api.POST(
		"/auth/reset-password",
		passwordRateStore.Middleware(rateLimits.PasswordReset),
		h.ResetPassword,
	)

	protected := api.Group("")
	protected.Use(middlewarePorts.Authenticate)
	protected.Use(middlewarePorts.CSRF)
	protected.GET("/me", h.GetMe)
	protected.PUT("/users/me/password", h.ChangeMyPassword)
	return protected, nil
}

func validateRateLimitConfig(config AuthRateLimitConfig) error {
	policies := []struct {
		name   string
		policy RateLimitPolicy
	}{
		{name: "login", policy: config.Login},
		{name: "password_reset", policy: config.PasswordReset},
		{name: "refresh", policy: config.Refresh},
		{name: "logout", policy: config.Logout},
		{name: "logout_redirect", policy: config.LogoutRedirect},
	}
	for _, item := range policies {
		if item.policy.Requests <= 0 ||
			item.policy.Window <= 0 ||
			item.policy.Burst <= 0 {
			return fmt.Errorf("invalid auth rate-limit policy: %s", item.name)
		}
	}
	return nil
}
