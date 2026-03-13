package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testSecret = "test-secret-key"

func makeToken(t *testing.T, method jwt.SigningMethod, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	s, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":   "user-uuid-123",
		"clinic_id": "clinic-uuid-456",
		"user_type": "admin",
		"exp":       time.Now().Add(time.Hour).Unix(),
	}
}

// runAuthMiddleware sets up a test router with the Auth middleware and a
// downstream handler that records the context values, then fires the request.
func runAuthMiddleware(t *testing.T, authHeader string) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var captured *gin.Context
	w := httptest.NewRecorder()
	router := gin.New()
	router.Use(Auth(testSecret))
	router.GET("/test", func(c *gin.Context) {
		captured = c
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	router.ServeHTTP(w, req)

	return w, captured
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing Authorization header returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authorization header required")
	})

	t.Run("malformed header without Bearer scheme returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "Token some-token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token string returns 401", func(t *testing.T) {
		w, _ := runAuthMiddleware(t, "Bearer not-a-valid-token")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid or expired token")
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		expiredClaims := jwt.MapClaims{
			"user_id":   "user-uuid-123",
			"clinic_id": "clinic-uuid-456",
			"user_type": "admin",
			"exp":       time.Now().Add(-time.Hour).Unix(),
		}
		token := makeToken(t, jwt.SigningMethodHS256, expiredClaims)
		w, _ := runAuthMiddleware(t, "Bearer "+token)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token signed with HS384 (also HMAC) passes through", func(t *testing.T) {
		// The middleware checks t.Method.(*jwt.SigningMethodHMAC), which matches all
		// HMAC variants (HS256, HS384, HS512). An HS384 token with the correct secret
		// is therefore accepted. Algorithm confusion attacks via non-HMAC methods
		// (e.g. RS256, none) are blocked, but HS384 is not.
		token := makeToken(t, jwt.SigningMethodHS384, validClaims())
		w, _ := runAuthMiddleware(t, "Bearer "+token)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("token signed with wrong secret returns 401", func(t *testing.T) {
		// Manually sign with a different key
		claims := validClaims()
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := rawToken.SignedString([]byte("wrong-secret"))
		assert.NoError(t, err)

		w, _ := runAuthMiddleware(t, "Bearer "+tokenStr)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid HS256 token passes through with correct context values", func(t *testing.T) {
		token := makeToken(t, jwt.SigningMethodHS256, validClaims())
		w, captured := runAuthMiddleware(t, "Bearer "+token)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotNil(t, captured)
		assert.Equal(t, "user-uuid-123", captured.GetString("user_id"))
		assert.Equal(t, "clinic-uuid-456", captured.GetString("clinic_id"))
		assert.Equal(t, "admin", captured.GetString("user_type"))
	})

	t.Run("Bearer scheme is case-insensitive", func(t *testing.T) {
		token := makeToken(t, jwt.SigningMethodHS256, validClaims())
		w, _ := runAuthMiddleware(t, "bearer "+token)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
