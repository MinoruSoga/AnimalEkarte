package owner

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterRoutes registers the complete 16-route authenticated owner surface.
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	owners := protected.Group("/owners")
	owners.GET("", h.requirePermission(string(model.ResourceOwners), "view"), h.ListOwners)
	owners.GET("/:id", h.requirePermission(string(model.ResourceOwners), "view"), h.GetOwner)
	owners.POST("", h.requirePermission(string(model.ResourceOwners), "create"), h.CreateOwner)
	owners.PATCH("/:id", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwner)
	owners.DELETE("/:id", h.requirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwner)
	owners.PATCH("/:id/line-user-id", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineUserID)
	owners.PATCH("/:id/line", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineUserID)
	owners.PATCH("/:id/delivery-exclusion", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryExclusion)
	owners.PATCH("/:id/delivery-caution", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerDeliveryCaution)
	owners.PATCH("/:id/transfer-status", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerTransferStatus)
	owners.PATCH("/:id/line-id-confirm", h.requirePermission(string(model.ResourceOwners), "edit"), h.UpdateOwnerLineIDConfirm)
	// POC-08 / SOLO-33: removed decorative /clinics/:clinic_id/owners/* PATCH aliases that
	// ignored path clinic_id and were absent from OpenAPI. Canonical /owners/:id/* remains.
}
