package auth

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
	LegacyTokenCookieName  = "auth_token"

	RefreshTokenCookiePath       = "/api/v1/auth"         //nolint:gosec // cookie route path, not a credential
	LegacyRefreshTokenCookiePath = "/api/v1/auth/refresh" //nolint:gosec // cookie route path, not a credential

	maxRefreshTokenCookieBytes = 4096
	loginFailureAuditWorkers   = 4
)

// HTTPStaffReader is the auth HTTP layer's minimal staff dependency.
type HTTPStaffReader interface {
	GetByID(ctx context.Context, id uint64) (*model.Staff, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
}

// StaffClinicAssignmentReader resolves the current clinic memberships for a staff member.
type StaffClinicAssignmentReader interface {
	FindAllByStaffID(ctx context.Context, staffID uint64) ([]model.StaffClinicAssignment, error)
}

// ClinicLister supplies clinics used by login and /me response composition.
type ClinicLister interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
}

// AuthAuditEntry is the auth-owned representation of a permission-group audit event.
type AuthAuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	IPAddress  string
	UserAgent  string
}

// AuthAuditLogger is the narrow audit surface consumed by auth HTTP handlers.
type AuthAuditLogger interface {
	LogAuthLogin(
		ctx context.Context,
		clinicID, staffID *uint64,
		action, ipAddress, userAgent string,
	) error
	LogEntry(ctx context.Context, entry AuthAuditEntry) error
}

// HTTPDependencies contains only capabilities directly consumed by auth HTTP behavior.
type HTTPDependencies struct {
	Auth                 Service
	Tokens               TokenService
	TokenBlacklist       TokenBlacklistService
	PasswordReset        PasswordResetService
	Accounts             AccountService
	Staff                HTTPStaffReader
	StaffAssignments     StaffClinicAssignmentReader
	Clinics              ClinicLister
	PermissionGroups     PermissionGroupService
	EffectivePermissions EffectivePermissionService
	Audit                AuthAuditLogger
}

// CookieConfig makes environment-specific cookie behavior explicit at composition time.
type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
}

// CookieConfigForProduction preserves the legacy release/debug cookie policy.
func CookieConfigForProduction(isProduction bool) CookieConfig {
	if isProduction {
		return CookieConfig{Secure: true, SameSite: http.SameSiteNoneMode}
	}
	return CookieConfig{Secure: false, SameSite: http.SameSiteLaxMode}
}

// PermissionRequirement identifies one resource/action pair for OR authorization.
type PermissionRequirement struct {
	Resource string
	Action   string
}

// RateLimitPolicy is an endpoint-specific token-bucket policy.
type RateLimitPolicy struct {
	Requests float64
	Window   time.Duration
	Burst    int
	// CountOnFailure は、status >= 400 の応答だけが予算を消費するモードを
	// 指定する。成功 request はバケットを消費しない。
	CountOnFailure bool
}

// RequestsPerSecond converts the typed policy to the middleware rate unit.
func (p RateLimitPolicy) RequestsPerSecond() float64 {
	if p.Window <= 0 {
		return 0
	}
	return p.Requests / p.Window.Seconds()
}

// AuthRateLimitConfig contains every auth endpoint's hardened rate policy.
type AuthRateLimitConfig struct {
	Login           RateLimitPolicy
	LoginPerAccount RateLimitPolicy
	PasswordReset   RateLimitPolicy
	Refresh         RateLimitPolicy
	Logout          RateLimitPolicy
	LogoutRedirect  RateLimitPolicy
}

// DefaultAuthRateLimitConfig matches the hardened route configuration.
// Login は IP 単位・失敗応答のみカウント（共有 NAT の正当利用で枯渇しない）。
// LoginPerAccount はアカウント email 単位の失敗カウンターで、IP を跨いだ
// 同一アカウントへの継続的な試行を止める。
func DefaultAuthRateLimitConfig() AuthRateLimitConfig {
	return AuthRateLimitConfig{
		Login:           RateLimitPolicy{Requests: 30, Window: time.Minute, Burst: 30, CountOnFailure: true},
		LoginPerAccount: RateLimitPolicy{Requests: 10, Window: time.Minute, Burst: 10, CountOnFailure: true},
		PasswordReset:   RateLimitPolicy{Requests: 3, Window: time.Minute, Burst: 3},
		Refresh:         RateLimitPolicy{Requests: 30, Window: time.Minute, Burst: 30},
		Logout:          RateLimitPolicy{Requests: 30, Window: time.Minute, Burst: 30},
		LogoutRedirect:  RateLimitPolicy{Requests: 30, Window: time.Minute, Burst: 30},
	}
}

// LoginFailureLimiter throttles repeated failed login attempts per account key.
// Admit reports whether an attempt may proceed now; RecordFailure consumes
// budget only after a failed attempt, so successful logins are free.
type LoginFailureLimiter interface {
	Admit(key string) (retryAfterSeconds int, ok bool)
	RecordFailure(key string)
}

// RateLimitStore is a route-local limiter bucket collection.
type RateLimitStore interface {
	Middleware(policy RateLimitPolicy) gin.HandlerFunc
	FailureLimiter(policy RateLimitPolicy) LoginFailureLimiter
}

// RateLimitStoreFactory creates independently isolated auth rate-limit stores.
type RateLimitStoreFactory interface {
	New(ctx context.Context) RateLimitStore
}

// RouteMiddlewarePorts are cross-cutting HTTP middleware supplied by composition.
type RouteMiddlewarePorts struct {
	CSRF             gin.HandlerFunc
	Authenticate     gin.HandlerFunc
	RateLimitFactory RateLimitStoreFactory
}

// backgroundWorkGate serializes worker registration with terminal shutdown.
// WaitGroup.Add must never race with Wait after the counter can reach zero, so
// callers register while mu is held and CloseAndWait permanently closes the
// registration side before it begins waiting.
type backgroundWorkGate struct {
	mu     sync.Mutex
	closed bool
	wg     sync.WaitGroup
}

func (g *backgroundWorkGate) Register() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return false
	}
	g.wg.Add(1)
	return true
}

func (g *backgroundWorkGate) Done() {
	g.wg.Done()
}

func (g *backgroundWorkGate) CloseAndWait() {
	g.mu.Lock()
	g.closed = true
	g.mu.Unlock()
	g.wg.Wait()
}

// HTTPHandler owns auth endpoints and global authorization behavior.
type HTTPHandler struct {
	deps               HTTPDependencies
	cookies            CookieConfig
	loginFailureTiming loginFailureResponseTiming
	loginFailures      LoginFailureLimiter
	loginAuditOnce     sync.Once
	loginAuditSlots    chan struct{}
	loginAuditGate     backgroundWorkGate
}

// NewHTTPHandler constructs the auth HTTP boundary.
func NewHTTPHandler(deps HTTPDependencies, cookies CookieConfig) *HTTPHandler {
	return &HTTPHandler{
		deps:               deps,
		cookies:            cookies,
		loginFailureTiming: loginFailureResponseTiming{}.withDefaults(),
		loginAuditSlots:    make(chan struct{}, loginFailureAuditWorkers),
	}
}
