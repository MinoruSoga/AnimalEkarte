package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims はJWTのペイロード
type JWTClaims struct {
	UserID   string `json:"user_id"`
	ClinicID string `json:"clinic_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

// Auth はJWTトークンを検証する認証ミドルウェア。
// httpOnly Cookie を優先して読み、なければ Authorization Bearer ヘッダにフォールバックする。
// secret には config.Config.JWTSecret を渡す。
func Auth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		var tokenStr string

		// httpOnly Cookie を優先して読む（XSS耐性あり）
		if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
			tokenStr = cookie
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// クレームをコンテキストに格納
		c.Set("user_id", claims.UserID)
		c.Set("clinic_id", claims.ClinicID)
		c.Set("user_type", claims.UserType)
		c.Next()
	}
}
