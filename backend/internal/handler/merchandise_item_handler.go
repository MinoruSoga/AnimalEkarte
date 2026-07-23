package handler

import "github.com/gin-gonic/gin"

// The methods in this file are BE9-2F compatibility delegates for
// master_routes.go. Delete them after the inventory domain registers its
// merchandise-item routes directly.

func (h *Handler) ListMerchandiseItems(c *gin.Context) {
	h.inventoryDomainHandler().ListMerchandiseItems(c)
}

func (h *Handler) GetMerchandiseItem(c *gin.Context) {
	h.inventoryDomainHandler().GetMerchandiseItem(c)
}

func (h *Handler) CreateMerchandiseItem(c *gin.Context) {
	h.inventoryDomainHandler().CreateMerchandiseItem(c)
}

func (h *Handler) UpdateMerchandiseItem(c *gin.Context) {
	h.inventoryDomainHandler().UpdateMerchandiseItem(c)
}

func (h *Handler) ReorderMerchandiseItems(c *gin.Context) {
	h.inventoryDomainHandler().ReorderMerchandiseItems(c)
}

func (h *Handler) DeleteMerchandiseItem(c *gin.Context) {
	h.inventoryDomainHandler().DeleteMerchandiseItem(c)
}
