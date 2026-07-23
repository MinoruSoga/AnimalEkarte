package handler

import (
	"github.com/gin-gonic/gin"

	inventorydomain "github.com/animal-ekarte/backend/internal/inventory"
)

// inventoryDomainHandler is a BE9-2F compatibility bridge for the legacy
// Handler route aggregator. Delete it and the methods below after
// handler.RegisterRoutes composes internal/inventory directly.
func (h *Handler) inventoryDomainHandler() *inventorydomain.Handler {
	return inventorydomain.NewHandler(h.svc.Inventory, h.svc.MerchandiseItem, h.RequirePermission)
}

func (h *Handler) ListInventory(c *gin.Context) {
	h.inventoryDomainHandler().ListInventory(c)
}

func (h *Handler) GetInventory(c *gin.Context) {
	h.inventoryDomainHandler().GetInventory(c)
}

func (h *Handler) CreateInventory(c *gin.Context) {
	h.inventoryDomainHandler().CreateInventory(c)
}

func (h *Handler) UpdateInventory(c *gin.Context) {
	h.inventoryDomainHandler().UpdateInventory(c)
}

func (h *Handler) DeleteInventory(c *gin.Context) {
	h.inventoryDomainHandler().DeleteInventory(c)
}
