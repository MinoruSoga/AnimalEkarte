package owner

import (
	"context"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware is owner's consumer-side view of auth-owned RBAC.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// PermissionChecker is owner's consumer-side view of a single RBAC decision.
type PermissionChecker func(c *gin.Context, resource, action string) bool

// DeletionLifecycle is the LSTEP cleanup capability invoked before owner
// deletion. Its existing best-effort contract is preserved.
type DeletionLifecycle interface {
	HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error
}

// Handler owns the complete authenticated owner HTTP surface.
type Handler struct {
	service           Service
	deletionLifecycle DeletionLifecycle
	requirePermission PermissionMiddleware
	hasPermission     PermissionChecker
}

// NewHandler constructs the owner HTTP boundary.
func NewHandler(
	service Service,
	deletionLifecycle DeletionLifecycle,
	requirePermission PermissionMiddleware,
	hasPermission PermissionChecker,
) *Handler {
	return &Handler{
		service:           service,
		deletionLifecycle: deletionLifecycle,
		requirePermission: requirePermission,
		hasPermission:     hasPermission,
	}
}
