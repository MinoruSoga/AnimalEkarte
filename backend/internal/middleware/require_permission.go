package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionChecker は権限チェックのインターフェース（循環インポート回避）
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uint64, resource model.Resource, action string) (bool, error)
}

// RequirePermission は指定リソース×アクションの権限を持つユーザーのみアクセスを許可するミドルウェア
// system_admin / clinic_admin は全権限を持つためバイパスする
func RequirePermission(resource model.Resource, action string, checker PermissionChecker) gin.HandlerFunc {
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

		// system_admin / clinic_admin は全権限バイパス
		ut := model.UserType(userType)
		if ut == model.UserTypeSystemAdmin || ut == model.UserTypeClinicAdmin {
			c.Next()
			return
		}

		// staff は DB から実効権限をチェック
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
			c.Abort()
			return
		}
		userIDStr, ok := userIDVal.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			c.Abort()
			return
		}
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
			c.Abort()
			return
		}

		hasPerm, err := checker.HasPermission(c.Request.Context(), userID, resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "permission check failed"})
			c.Abort()
			return
		}
		if !hasPerm {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}
		c.Next()
	}
}
