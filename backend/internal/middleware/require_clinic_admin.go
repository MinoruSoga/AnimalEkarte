package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireClinicAdmin は clinic_admin または system_admin のみアクセスを許可するミドルウェア
func RequireClinicAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userTypeVal, exists := c.Get("user_type")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			c.Abort()
			return
		}
		userType, ok := userTypeVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			c.Abort()
			return
		}

		if userType != "system_admin" && userType != "clinic_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "clinic admin or above required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
