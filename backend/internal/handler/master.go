package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllMasterItems godoc
// @Summary マスタアイテム一覧取得
// @Description 登録されているマスタアイテムの一覧を取得します
// @Tags master-items
// @Accept json
// @Produce json
// @Param category query string false "カテゴリでフィルタリング"
// @Success 200 {array} model.MasterItem
// @Failure 500 {object} map[string]string
// @Router /master-items [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllMasterItems(c *gin.Context) {
	ctx := c.Request.Context()

	items, err := h.svc.GetAllMasterItems(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get master items", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetMasterItemByID godoc
// @Summary マスタアイテム詳細取得
// @Description 指定されたIDのマスタアイテム情報を取得します
// @Tags master-items
// @Accept json
// @Produce json
// @Param id path string true "マスタアイテムID (UUID)"
// @Success 200 {object} model.MasterItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /master-items/{id} [get]
func (h *Handler) GetMasterItemByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	item, err := h.svc.GetMasterItemByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "master_item", id)
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetMasterItemsByCategory godoc
// @Summary カテゴリ別マスタアイテム取得
// @Description 指定されたカテゴリのマスタアイテム一覧を取得します
// @Tags master-items
// @Accept json
// @Produce json
// @Param category path string true "カテゴリ (examination, vaccine, medicine, staff, insurance, cage, serviceType, trimming_course, trimming_option)"
// @Success 200 {array} model.MasterItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /master-items/category/{category} [get]
func (h *Handler) GetMasterItemsByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")

	items, err := h.svc.GetMasterItemsByCategory(ctx, category)
	if err != nil {
		h.handleError(c, err, "master_item", category)
		return
	}

	c.JSON(http.StatusOK, items)
}

// CreateMasterItem godoc
// @Summary マスタアイテム作成
// @Description 新しいマスタアイテムを作成します
// @Tags master-items
// @Accept json
// @Produce json
// @Param item body model.CreateMasterItemRequest true "マスタアイテム情報"
// @Success 201 {object} model.MasterItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /master-items [post]
func (h *Handler) CreateMasterItem(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateMasterItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.svc.CreateMasterItem(ctx, &req)
	if err != nil {
		h.handleError(c, err, "master_item", "")
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateMasterItem godoc
// @Summary マスタアイテム更新
// @Description 既存のマスタアイテムを更新します
// @Tags master-items
// @Accept json
// @Produce json
// @Param id path string true "マスタアイテムID (UUID)"
// @Param item body model.UpdateMasterItemRequest true "マスタアイテム情報"
// @Success 200 {object} model.MasterItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /master-items/{id} [put]
func (h *Handler) UpdateMasterItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateMasterItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	item, err := h.svc.UpdateMasterItem(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "master_item", id)
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteMasterItem godoc
// @Summary マスタアイテム削除
// @Description 指定されたマスタアイテムを削除します
// @Tags master-items
// @Accept json
// @Produce json
// @Param id path string true "マスタアイテムID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /master-items/{id} [delete]
func (h *Handler) DeleteMasterItem(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteMasterItem(ctx, id); err != nil {
		h.handleError(c, err, "master_item", id)
		return
	}

	c.Status(http.StatusNoContent)
}
