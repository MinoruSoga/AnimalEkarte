package inventory

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterRoutes registers the inventory domain's 11 legacy-compatible routes.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	inventory := rg.Group("/inventory")
	inventory.GET("", h.requirePermission(string(model.ResourceInventory), "view"), h.ListInventory)
	inventory.GET("/:id", h.requirePermission(string(model.ResourceInventory), "view"), h.GetInventory)
	inventory.POST("", h.requirePermission(string(model.ResourceInventory), "create"), h.CreateInventory)
	inventory.PATCH("/:id", h.requirePermission(string(model.ResourceInventory), "edit"), h.UpdateInventory)
	inventory.DELETE("/:id", h.requirePermission(string(model.ResourceInventory), "delete"), h.DeleteInventory)

	masters := rg.Group("/masters")
	masters.GET("/merchandise-items", h.requirePermission(string(model.ResourceMasterMerchandise), "view"), h.ListMerchandiseItems)
	masters.POST("/merchandise-items", h.requirePermission(string(model.ResourceMasterMerchandise), "create"), h.CreateMerchandiseItem)
	masters.PATCH("/merchandise-items/reorder", h.requirePermission(string(model.ResourceMasterMerchandise), "edit"), h.ReorderMerchandiseItems)
	masters.GET("/merchandise-items/:id", h.requirePermission(string(model.ResourceMasterMerchandise), "view"), h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", h.requirePermission(string(model.ResourceMasterMerchandise), "edit"), h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", h.requirePermission(string(model.ResourceMasterMerchandise), "delete"), h.DeleteMerchandiseItem)
}
