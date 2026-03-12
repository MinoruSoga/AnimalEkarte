package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListInventory godoc
// @Summary 在庫一覧取得
// @Description 在庫アイテムの一覧をページネーション付きで取得する
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param category query string false "カテゴリフィルター"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory [get]
func (h *Handler) ListInventory(c *gin.Context) {
	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var category *string
	if cat := c.Query("category"); cat != "" {
		category = &cat
	}

	items, total, err := h.svc.Inventory.List(c.Request.Context(), category, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: items, Total: total, Page: page, Limit: limit})
}

// GetInventory godoc
// @Summary 在庫詳細取得
// @Description 指定IDの在庫アイテムを取得する
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "在庫ID"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory/{id} [get]
func (h *Handler) GetInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.svc.Inventory.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateInventory godoc
// @Summary 在庫作成
// @Description 新しい在庫アイテムを作成する
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.InventoryItem true "在庫情報"
// @Success 201 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory [post]
func (h *Handler) CreateInventory(c *gin.Context) {
	var input model.InventoryItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Inventory.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateInventory godoc
// @Summary 在庫更新
// @Description 指定IDの在庫アイテムを更新する
// @Tags Inventory
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "在庫ID"
// @Param input body model.InventoryItem true "更新する在庫情報"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory/{id} [put]
func (h *Handler) UpdateInventory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.InventoryItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Inventory.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}
