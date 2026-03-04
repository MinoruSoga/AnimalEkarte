package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllInventoryItems godoc
// @Summary 在庫アイテム一覧取得
// @Description 登録されている在庫アイテムの一覧を取得します
// @Tags inventory
// @Accept json
// @Produce json
// @Success 200 {array} model.InventoryItem
// @Failure 500 {object} map[string]string
// @Router /inventory-items [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllInventoryItems(c *gin.Context) {
	ctx := c.Request.Context()

	items, err := h.svc.GetAllInventoryItems(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get inventory items", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetInventoryItemByID godoc
// @Summary 在庫アイテム詳細取得
// @Description 指定されたIDの在庫アイテム情報を取得します
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "在庫アイテムID (UUID)"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items/{id} [get]
func (h *Handler) GetInventoryItemByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	item, err := h.svc.GetInventoryItemByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "inventory item", id)
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetInventoryItemsByCategory godoc
// @Summary 在庫アイテムカテゴリ別取得
// @Description 指定されたカテゴリの在庫アイテムを取得します
// @Tags inventory
// @Accept json
// @Produce json
// @Param category path string true "カテゴリ (medicine, consumable, food, other)"
// @Success 200 {array} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items/category/{category} [get]
func (h *Handler) GetInventoryItemsByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")

	items, err := h.svc.GetInventoryItemsByCategory(ctx, category)
	if err != nil {
		h.handleError(c, err, "inventory item", category)
		return
	}

	c.JSON(http.StatusOK, items)
}

// GetInventoryItemsByStatus godoc
// @Summary 在庫アイテムステータス別取得
// @Description 指定されたステータスの在庫アイテムを取得します
// @Tags inventory
// @Accept json
// @Produce json
// @Param status path string true "ステータス (sufficient, low, out_of_stock)"
// @Success 200 {array} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items/status/{status} [get]
func (h *Handler) GetInventoryItemsByStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Param("status")

	items, err := h.svc.GetInventoryItemsByStatus(ctx, status)
	if err != nil {
		h.handleError(c, err, "inventory item", status)
		return
	}

	c.JSON(http.StatusOK, items)
}

// CreateInventoryItem godoc
// @Summary 在庫アイテム作成
// @Description 新しい在庫アイテムを作成します
// @Tags inventory
// @Accept json
// @Produce json
// @Param item body model.CreateInventoryItemRequest true "在庫アイテム情報"
// @Success 201 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items [post]
func (h *Handler) CreateInventoryItem(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.svc.CreateInventoryItem(ctx, &req)
	if err != nil {
		h.handleError(c, err, "inventory item", "")
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateInventoryItem godoc
// @Summary 在庫アイテム更新
// @Description 既存の在庫アイテムを更新します
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "在庫アイテムID (UUID)"
// @Param item body model.UpdateInventoryItemRequest true "在庫アイテム情報"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items/{id} [put]
func (h *Handler) UpdateInventoryItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.svc.UpdateInventoryItem(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "inventory item", id)
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteInventoryItem godoc
// @Summary 在庫アイテム削除
// @Description 指定された在庫アイテムを削除します
// @Tags inventory
// @Accept json
// @Produce json
// @Param id path string true "在庫アイテムID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory-items/{id} [delete]
func (h *Handler) DeleteInventoryItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteInventoryItem(ctx, id); err != nil {
		h.handleError(c, err, "inventory item", id)
		return
	}

	c.Status(http.StatusNoContent)
}
