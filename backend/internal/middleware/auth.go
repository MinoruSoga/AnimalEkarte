package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/animal-ekarte/backend/internal/model"
)

// JWTClaims はJWTのペイロード
type JWTClaims struct {
	UserID   string `json:"user_id"`
	ClinicID string `json:"clinic_id"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

// ActiveUserFinder はアクティブなユーザーをIDで取得するインターフェース（循環インポート回避）
type ActiveUserFinder interface {
	FindActiveByID(ctx context.Context, id uint64) (*model.UserAccount, error)
}

// Auth はJWTトークンを検証する認証ミドルウェア。
// httpOnly Cookie を優先して読み、なければ Authorization Bearer ヘッダにフォールバックする。
// secret には config.Config.JWTSecret を渡す。
// userRepo を指定するとJWT検証後にDBのアカウント状態もチェックする（BUG-061対応）。
func Auth(secret string, userRepo ...ActiveUserFinder) gin.HandlerFunc {
	key := []byte(secret)
	var repo ActiveUserFinder
	if len(userRepo) > 0 {
		repo = userRepo[0]
	}
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		claims := &JWTClaims{}
		if _, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		}); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// DBのアカウント状態をチェックする（BUG-061: アカウント停止・削除ユーザーの遮断）
		if repo != nil {
			userIDUint, parseErr := strconv.ParseUint(claims.UserID, 10, 64)
			if parseErr != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
				c.Abort()
				return
			}
			if _, err := repo.FindActiveByID(c.Request.Context(), userIDUint); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "account is disabled or deleted"})
				c.Abort()
				return
			}
		}

		// err == nil が確定しているためクレームを安全に格納
		c.Set("user_id", claims.UserID)
		c.Set("clinic_id", claims.ClinicID)
		c.Set("user_type", claims.UserType)
		c.Next()
	}
}
