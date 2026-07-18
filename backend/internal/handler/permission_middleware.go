package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// RequirePermission は指定リソース・アクションの権限を持つユーザーのみ通過させる gin ミドルウェアを返す。
// system_admin は全権限バイパス。それ以外は permission_group_rules から判定する。
func (h *Handler) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.hasPermission(c, resource, action) {
			RespondError(c, apperrors.WrapForbidden("forbidden"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermissionAny は指定された複数の(リソース, アクション)ペアの内、いずれかの権限を持つユーザーのみ通過させる。
// system_admin は全権限バイパス。複数の権限オプション(OR)をサポート。
func (h *Handler) RequirePermissionAny(permissions ...struct{ Resource, Action string }) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, perm := range permissions {
			if h.hasPermission(c, perm.Resource, perm.Action) {
				c.Next()
				return
			}
		}
		RespondError(c, apperrors.WrapForbidden("forbidden"))
		c.Abort()
	}
}
