package inventory

import "github.com/gin-gonic/gin"

// PermissionMiddleware is inventory's consumer-side view of RBAC middleware.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler owns the inventory HTTP boundary without depending on the legacy
// handler/service aggregators.
type Handler struct {
	inventory         InventoryService
	merchandiseItem   MerchandiseItemService
	requirePermission PermissionMiddleware
}

// NewHandler constructs the inventory HTTP boundary.
func NewHandler(
	inventoryService InventoryService,
	merchandiseItemService MerchandiseItemService,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		inventory:         inventoryService,
		merchandiseItem:   merchandiseItemService,
		requirePermission: requirePermission,
	}
}
