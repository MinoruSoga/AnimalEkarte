package middleware

import (
	"context"
	"database/sql/driver"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	authdomain "github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/authjwt"
)

// JWTClaims は JWT ペイロードのエイリアス。
type JWTClaims = authjwt.Claims

// StaffValidationFailureNotifier は staff 有効性検証の一時的な取得エラーを通知する。
// 通知はベストエフォートであり、通知の有無にかかわらず一時障害は fail-closed で拒否する。
// 通知自体の失敗も認証判定を変えない。
type StaffValidationFailureNotifier func(ctx context.Context, staffID uint64, cause error) error

// Auth はJWTトークンを検証する認証ミドルウェア。
// httpOnly Cookie を優先して読み、なければ Authorization Bearer ヘッダにフォールバックする。
// tokenSvc は access JWT 検証に使用する（VerifyAccessToken）。
// isProduction には cfg.GinMode == "release" を渡す。
// auditSvc はクリニック切替の監査ログ記録に使用する（ベストエフォート: nil 許容）。
// accessResolver は現在の staff/account/clinic 所属を毎リクエスト再検証する。
func Auth(
	tokenSvc authdomain.TokenService,
	isProduction bool,
	auditSvc audit.Service,
	accessResolver authdomain.CurrentAccessResolver,
) gin.HandlerFunc {
	return auth(tokenSvc, isProduction, auditSvc, accessResolver, nil)
}

// AuthWithStaffValidationFailureNotifier は Auth に staff 有効性検証エラーの通知経路を追加する。
// 一時的な staff lookup 障害でも JWT の role/clinic/system-admin/epoch を保持して継続せず、
// 通知後に access validation unavailable として fail-closed する。
func AuthWithStaffValidationFailureNotifier(
	tokenSvc authdomain.TokenService,
	isProduction bool,
	auditSvc audit.Service,
	accessResolver authdomain.CurrentAccessResolver,
	notifier StaffValidationFailureNotifier,
) gin.HandlerFunc {
	return auth(tokenSvc, isProduction, auditSvc, accessResolver, notifier)
}

func auth(
	tokenSvc authdomain.TokenService,
	isProduction bool,
	auditSvc audit.Service,
	accessResolver authdomain.CurrentAccessResolver,
	notifier StaffValidationFailureNotifier,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			respondError(c, http.StatusUnauthorized, "authorization required")
			return
		}
		if tokenSvc == nil {
			slog.ErrorContext(
				c.Request.Context(),
				"access token validation unavailable: token service is not configured",
			)
			respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
			return
		}

		claims, err := tokenSvc.VerifyAccessToken(tokenStr)
		if err != nil || claims == nil {
			respondError(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		currentAccess, ok := resolveCurrentAccess(
			c,
			accessResolver,
			claims.UserID,
			notifier,
		)
		if !ok {
			return
		}

		// Live authority always wins on successful resolution. Temporary
		// lookup failures never reach this branch (fail-closed above).
		requestClaims := cloneJWTClaims(claims)
		if !authdomain.TokenMatchesAccountEpoch(
			claims,
			currentAccess.AccountEpoch,
		) {
			respondError(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		requestClaims.IsSystemAdmin = currentAccess.IsSystemAdmin
		requestClaims.ClinicIDs = append(
			[]uint64(nil),
			currentAccess.ClinicIDs...,
		)
		if requestClaims.IsSystemAdmin {
			currentMainID, parseErr := strconv.ParseUint(
				currentAccess.MainClinicID,
				10,
				64,
			)
			if parseErr != nil ||
				currentMainID == 0 ||
				!slices.Contains(currentAccess.ClinicIDs, currentMainID) {
				slog.ErrorContext(
					c.Request.Context(),
					"current system administrator clinic authority is invalid",
					"staff_id",
					currentAccess.StaffID,
				)
				respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
				return
			}
			requestClaims.ClinicID = currentAccess.MainClinicID
		}

		// クリニック切替: X-Clinic-ID ヘッダーが送信された場合、所属チェック後に上書き（BUG-128）
		clinicID, ok := resolveClinicID(
			c,
			requestClaims,
			isProduction,
			auditSvc,
		)
		if !ok {
			return
		}
		c.Set("user_id", requestClaims.UserID)
		c.Set("is_system_admin", requestClaims.IsSystemAdmin)
		c.Set("clinic_id", clinicID)
		c.Set(
			"clinic_ids",
			append([]uint64(nil), requestClaims.ClinicIDs...),
		)

		c.Next()
	}
}

func cloneJWTClaims(claims *JWTClaims) *JWTClaims {
	cloned := *claims
	cloned.ClinicIDs = append([]uint64(nil), claims.ClinicIDs...)
	return &cloned
}

// extractToken は access_token Cookie → 旧Cookie名 auth_token → Authorization Bearer
// ヘッダの順で JWT 文字列を抽出する（BE-refactor.md E-14）。見つからなければ空文字列を返す。
func extractToken(c *gin.Context) string {
	// access_token Cookie を優先して読む（XSS耐性あり）
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}
	// 後方互換: 旧Cookie名 auth_token にフォールバック
	if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
		return cookie
	}

	// Cookie がなければ Authorization Bearer ヘッダにフォールバック
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}
	return ""
}

// resolveCurrentAccess revalidates mutable staff/account/clinic authority.
// Temporary staff-lookup failures may alert via notifier but always fail closed;
// JWT role/clinic/system-admin/epoch claims are never preserved as continuity.
func resolveCurrentAccess(
	c *gin.Context,
	resolver authdomain.CurrentAccessResolver,
	userID string,
	notifier StaffValidationFailureNotifier,
) (*authdomain.CurrentAccess, bool) {
	staffID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || staffID == 0 {
		respondError(c, http.StatusUnauthorized, "invalid staff identity")
		return nil, false
	}
	if resolver == nil {
		slog.ErrorContext(
			c.Request.Context(),
			"current access validation unavailable: resolver is not configured",
			"staff_id",
			staffID,
		)
		respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
		return nil, false
	}

	ctx := c.Request.Context()
	access, resolveErr := resolver.Resolve(ctx, staffID)
	if resolveErr != nil {
		return nil, rejectResolveCurrentAccessError(c, ctx, notifier, staffID, resolveErr)
	}
	if access == nil || access.StaffID != staffID ||
		access.AccountEpoch <= 0 {
		slog.ErrorContext(
			ctx,
			"current access validation returned invalid identity",
			"staff_id",
			staffID,
		)
		respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
		return nil, false
	}
	return access, true
}

func rejectResolveCurrentAccessError(
	c *gin.Context,
	ctx context.Context,
	notifier StaffValidationFailureNotifier,
	staffID uint64,
	resolveErr error,
) bool {
	var staffLookupErr *authdomain.StaffLookupError
	if errors.As(resolveErr, &staffLookupErr) {
		cause := errors.Unwrap(staffLookupErr)
		if apperrors.IsNotFound(cause) {
			respondError(c, http.StatusForbidden, "staff account is no longer active")
			return false
		}
		if isTemporaryStaffValidationError(cause) {
			slog.WarnContext(
				ctx,
				"failed to verify staff validity; denying request (fail-closed)",
				"staff_id",
				staffID,
				"error",
				cause,
			)
			logStaffValidationNotifyFailure(notifier, ctx, staffID, cause)
			respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
			return false
		}
	}

	if errors.Is(resolveErr, apperrors.ErrForbidden) {
		respondError(c, http.StatusForbidden, "current access is no longer available")
		return false
	}
	if errors.Is(resolveErr, apperrors.ErrUnauthorized) {
		respondError(c, http.StatusUnauthorized, "invalid or expired token")
		return false
	}

	slog.ErrorContext(
		ctx,
		"current access validation failed closed",
		"staff_id",
		staffID,
		"error",
		resolveErr,
	)
	respondError(c, http.StatusServiceUnavailable, "access validation unavailable")
	return false
}

// isTemporaryStaffValidationError recognizes only typed database availability
// failures eligible for operational notification before fail-closed denial.
// Generic errors, PostgreSQL programming/constraint errors, and identity
// errors must never be treated as temporary continuity exceptions.
func isTemporaryStaffValidationError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, pgconn.ErrConnClosed) {
		return true
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}
	switch postgresError.Code {
	case "08000", // connection_exception
		"08001", // sqlclient_unable_to_establish_sqlconnection
		"08003", // connection_does_not_exist
		"08006", // connection_failure
		"08007", // transaction_resolution_unknown
		"53300", // too_many_connections
		"57P01", // admin_shutdown
		"57P02", // crash_shutdown
		"57P03": // cannot_connect_now
		return true
	default:
		return false
	}
}

// resolveClinicID は X-Clinic-ID ヘッダーによるクリニック切替（所属検証・監査ログ・
// prev_clinic_id cookie 更新）を行い、最終的な clinic_id を返す（BE-refactor.md E-14,
// BUG-128, FEAT-374 Phase 2）。ok=false の場合は呼び出し元で既に応答済みのため
// 即座に return すること。
func resolveClinicID(
	c *gin.Context,
	claims *JWTClaims,
	isProduction bool,
	auditSvc audit.Service,
) (string, bool) {
	clinicID := claims.ClinicID
	headerClinicID := c.GetHeader("X-Clinic-ID")
	if headerClinicID == "" {
		defaultID, err := strconv.ParseUint(clinicID, 10, 64)
		if err != nil || defaultID == 0 {
			respondError(c, http.StatusUnauthorized, "invalid clinic identity")
			return "", false
		}
		if !slices.Contains(claims.ClinicIDs, defaultID) {
			respondError(c, http.StatusForbidden, "not assigned to this clinic")
			return "", false
		}
		return clinicID, true
	}

	headerID, err := strconv.ParseUint(headerClinicID, 10, 64)
	if err != nil || headerID == 0 {
		respondError(c, http.StatusBadRequest, "invalid clinic id")
		return "", false
	}
	if !slices.Contains(claims.ClinicIDs, headerID) {
		respondError(c, http.StatusForbidden, "not assigned to this clinic")
		return "", false
	}
	clinicID = headerClinicID

	// FEAT-374 Phase 2: クリニック切替 audit log（差分検出 + cookie 更新）
	prevClinicCookie, _ := c.Cookie("prev_clinic_id")
	auditFromClinicID := prevClinicCookie
	if auditFromClinicID == "" &&
		claims.IsSystemAdmin &&
		claims.ClinicID != clinicID {
		auditFromClinicID = claims.ClinicID
	}
	currentClinicID, parseErr := strconv.ParseUint(clinicID, 10, 64)
	if parseErr != nil {
		return clinicID, true
	}
	if auditFromClinicID != "" && auditFromClinicID != clinicID {
		logClinicSwitchBestEffort(c, auditSvc, claims.UserID, auditFromClinicID, currentClinicID)
	}
	// cookie 更新（初回も差分も同じく書込）
	if prevClinicCookie != clinicID {
		sameSite := http.SameSiteLaxMode
		secure := false
		if isProduction {
			sameSite = http.SameSiteNoneMode
			secure = true
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "prev_clinic_id",
			Value:    clinicID,
			Path:     "/",
			MaxAge:   15 * 60, // 15分（access_token と同寿命）
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})
	}
	return clinicID, true
}

func notifyStaffValidationFailure(
	notifier StaffValidationFailureNotifier,
	ctx context.Context,
	staffID uint64,
	cause error,
) error {
	if notifier == nil {
		return nil
	}
	return notifier(ctx, staffID, cause)
}

func logStaffValidationNotifyFailure(
	notifier StaffValidationFailureNotifier,
	ctx context.Context,
	staffID uint64,
	cause error,
) {
	notifyErr := notifyStaffValidationFailure(notifier, ctx, staffID, cause)
	if notifyErr == nil {
		return
	}
	slog.ErrorContext(
		ctx,
		"failed to notify staff validation failure (non-fatal)",
		"staff_id",
		staffID,
		"error",
		notifyErr,
	)
}

func logClinicSwitchBestEffort(
	c *gin.Context,
	auditSvc audit.Service,
	userID string,
	fromClinicID string,
	currentClinicID uint64,
) {
	if auditSvc == nil {
		return
	}
	prevID, perr := strconv.ParseUint(fromClinicID, 10, 64)
	if perr != nil {
		return
	}
	var actorID *uint64
	if uid, err := strconv.ParseUint(userID, 10, 64); err == nil {
		actorID = &uid
	}
	if err := auditSvc.LogClinicSwitch(
		c.Request.Context(),
		actorID, prevID, currentClinicID,
		c.ClientIP(), c.GetHeader("User-Agent"),
	); err != nil {
		slog.ErrorContext(c.Request.Context(),
			"failed to write clinic switch audit (best-effort)",
			"error", err)
	}
}
