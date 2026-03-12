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
// secret には config.Config.JWTSecret を渡す。
func Auth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
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
