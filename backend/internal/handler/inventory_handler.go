package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
// @Param status query string false "ステータスフィルター (sufficient, low, out_of_stock)"
// @Success 200 {object} handler.InventoryListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory [get]
func (h *Handler) ListInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var category *string
	if cat := c.Query("category"); cat != "" {
		category = &cat
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	items, total, err := h.svc.Inventory.List(c.Request.Context(), clinicID, category, status, page, limit)
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
// @Param id path integer true "在庫ID"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory/{id} [get]
func (h *Handler) GetInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.svc.Inventory.GetByID(c.Request.Context(), clinicID, id)
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input model.InventoryItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Inventory.Create(c.Request.Context(), clinicID, &input); err != nil {
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
// @Param id path integer true "在庫ID"
// @Param input body model.InventoryItem true "更新する在庫情報"
// @Success 200 {object} model.InventoryItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /inventory/{id} [put]
func (h *Handler) UpdateInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
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
	if err := h.svc.Inventory.Update(c.Request.Context(), clinicID, &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *Handler) DeleteInventory(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Inventory.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
