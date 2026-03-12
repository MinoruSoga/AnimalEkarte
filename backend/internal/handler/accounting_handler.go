package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListAccountings godoc
// @Summary 会計一覧取得
// @Description 会計（請求）の一覧をページネーション付きで取得する
// @Tags Accountings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param status query string false "ステータスフィルター"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings [get]
func (h *Handler) ListAccountings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uuid.UUID
	if s := c.Query("pet_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}
	var ownerID *uuid.UUID
	if s := c.Query("owner_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	accountings, total, err := h.svc.Accounting.List(c.Request.Context(), clinicID, petID, ownerID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: accountings, Total: total, Page: page, Limit: limit})
}

// GetAccounting godoc
// @Summary 会計詳細取得
// @Description 指定IDの会計（請求）を取得する
// @Tags Accountings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "会計ID"
// @Success 200 {object} model.Billing
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/{id} [get]
func (h *Handler) GetAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	accounting, err := h.svc.Accounting.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, accounting)
}

// CreateAccounting godoc
// @Summary 会計作成
// @Description 新しい会計（請求）を作成する
// @Tags Accountings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.Billing true "会計情報"
// @Success 201 {object} model.Billing
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings [post]
func (h *Handler) CreateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input model.Billing
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Accounting.Create(c.Request.Context(), clinicID, &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateAccounting godoc
// @Summary 会計更新
// @Description 指定IDの会計（請求）を更新する
// @Tags Accountings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "会計ID"
// @Param input body model.Billing true "更新する会計情報"
// @Success 200 {object} model.Billing
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accountings/{id} [put]
func (h *Handler) UpdateAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Billing
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Accounting.Update(c.Request.Context(), clinicID, &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *Handler) DeleteAccounting(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Accounting.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
