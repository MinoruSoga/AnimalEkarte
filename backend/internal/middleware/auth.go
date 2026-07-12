package middleware

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/animal-ekarte/backend/internal/service"
)

// JWTClaims はJWTのペイロード
type JWTClaims struct {
	UserID        string   `json:"user_id"`
	ClinicID      string   `json:"clinic_id"`
	IsSystemAdmin bool     `json:"is_system_admin"`
	ClinicIDs     []uint64 `json:"clinic_ids,omitempty"`
	jwt.RegisteredClaims
}

// Auth はJWTトークンを検証する認証ミドルウェア。
// httpOnly Cookie を優先して読み、なければ Authorization Bearer ヘッダにフォールバックする。
// secret には config.Config.JWTSecret、isProduction には cfg.GinMode == "release" を渡す。
// auditSvc はクリニック切替の監査ログ記録に使用する（ベストエフォート: nil 許容）。
// staffSvc は staff の有効性チェック（is_active && deleted_at IS NULL）に使用する。
func Auth(secret string, isProduction bool, auditSvc service.AuditService, staffSvc service.StaffService) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			respondError(c, http.StatusUnauthorized, "authorization required")
			return
		}

		claims := &JWTClaims{}
		if _, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		}); err != nil {
			respondError(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// err == nil が確定しているためクレームを安全に格納
		c.Set("user_id", claims.UserID)
		c.Set("is_system_admin", claims.IsSystemAdmin)

		// P10: staff 有効性チェック（deactivation / soft-delete 防止）
		if !checkStaffActive(c, staffSvc, claims.UserID) {
			return
		}

		// クリニック切替: X-Clinic-ID ヘッダーが送信された場合、所属チェック後に上書き（BUG-128）
		clinicID, ok := resolveClinicID(c, claims, isProduction, auditSvc)
		if !ok {
			return
		}
		c.Set("clinic_id", clinicID)
		c.Set("clinic_ids", claims.ClinicIDs) // #84: 所属医院リスト(owner作成時の医院選択の所属検証に使用)

		c.Next()
	}
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

// checkStaffActive は staff の有効性（is_active && deleted_at IS NULL）を検証する
// （BE-refactor.md E-14: P10 deactivation / soft-delete 防止）。戻り値が false の場合、
// 既に 403 レスポンスを書き込み済みのため呼び出し元は即座に return すること。
// userID が数値でない、staffSvc が nil、または DB 一時障害の場合は検証をスキップし
// true を返す（既存挙動を維持）。
func checkStaffActive(c *gin.Context, staffSvc service.StaffService, userID string) bool {
	staffID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || staffSvc == nil {
		return true
	}
	staff, getErr := staffSvc.GetByID(c.Request.Context(), staffID)
	if getErr != nil {
		slog.WarnContext(c.Request.Context(), "failed to verify staff validity (non-fatal)", "staff_id", staffID, "error", getErr)
		// 取得失敗は許容（DB 一時障害）→ 続行
		return true
	}
	if staff == nil || !staff.IsActive {
		// 削除済み or 無効化済み
		respondError(c, http.StatusForbidden, "staff account is no longer active")
		return false
	}
	return true
}

// resolveClinicID は X-Clinic-ID ヘッダーによるクリニック切替（所属検証・監査ログ・
// prev_clinic_id cookie 更新）を行い、最終的な clinic_id を返す（BE-refactor.md E-14,
// BUG-128, FEAT-374 Phase 2）。ok=false の場合は呼び出し元で既に応答済みのため
// 即座に return すること。
func resolveClinicID(c *gin.Context, claims *JWTClaims, isProduction bool, auditSvc service.AuditService) (string, bool) {
	clinicID := claims.ClinicID
	headerClinicID := c.GetHeader("X-Clinic-ID")
	if headerClinicID == "" {
		return clinicID, true
	}

	if claims.IsSystemAdmin {
		// system_admin はすべてのクリニックにアクセス可能
		clinicID = headerClinicID
	} else {
		hID, err := strconv.ParseUint(headerClinicID, 10, 64)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid clinic id")
			return "", false
		}
		if !slices.Contains(claims.ClinicIDs, hID) {
			respondError(c, http.StatusForbidden, "not assigned to this clinic")
			return "", false
		}
		clinicID = headerClinicID
	}

	// FEAT-374 Phase 2: クリニック切替 audit log（差分検出 + cookie 更新）
	prevClinicCookie, _ := c.Cookie("prev_clinic_id")
	currentClinicID, parseErr := strconv.ParseUint(clinicID, 10, 64)
	if parseErr == nil {
		if prevClinicCookie != "" && prevClinicCookie != clinicID {
			// 差分検出 → audit log（ベストエフォート）
			if prevID, perr := strconv.ParseUint(prevClinicCookie, 10, 64); perr == nil {
				if auditSvc != nil {
					var actorID *uint64
					if uid, err := strconv.ParseUint(claims.UserID, 10, 64); err == nil {
						actorID = &uid
					}
					if err := auditSvc.LogClinicSwitch(c.Request.Context(),
						actorID, prevID, currentClinicID,
						c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
						slog.ErrorContext(c.Request.Context(),
							"failed to write clinic switch audit (best-effort)",
							"error", err)
					}
				}
			}
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
	}
	return clinicID, true
}
