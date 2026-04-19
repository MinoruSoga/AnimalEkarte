package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
// secret には config.Config.JWTSecret を渡す。
func Auth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		var tokenStr string

		// access_token Cookie を優先して読む（XSS耐性あり）
		if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
			tokenStr = cookie
		}
		// 後方互換: 旧Cookie名 auth_token にフォールバック
		if tokenStr == "" {
			if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
				tokenStr = cookie
			}
		}

		// Cookie がなければ Authorization Bearer ヘッダにフォールバック
		if tokenStr == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
					tokenStr = parts[1]
				}
			}
		}

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

		// クリニック切替: X-Clinic-ID ヘッダーが送信された場合、所属チェック後に上書き（BUG-128）
		clinicID := claims.ClinicID
		if headerClinicID := c.GetHeader("X-Clinic-ID"); headerClinicID != "" {
			if claims.IsSystemAdmin {
				// system_admin はすべてのクリニックにアクセス可能
				clinicID = headerClinicID
			} else {
				hID, err := strconv.ParseUint(headerClinicID, 10, 64)
				if err != nil {
					respondError(c, http.StatusBadRequest, "invalid clinic id")
					return
				}
				if !slices.Contains(claims.ClinicIDs, hID) {
					respondError(c, http.StatusForbidden, "not assigned to this clinic")
					return
				}
				clinicID = headerClinicID
			}
		}
		c.Set("clinic_id", clinicID)

		c.Next()
	}
}
