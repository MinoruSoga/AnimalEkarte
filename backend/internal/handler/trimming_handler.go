package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListTrimmings godoc
// @Summary トリミング一覧取得
// @Description トリミング記録の一覧をページネーション付きで取得する
// @Tags Trimmings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param pet_id query integer false "ペットIDフィルター"
// @Success 200 {object} handler.TrimmingListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings [get]
func (h *Handler) ListTrimmings(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var petID *uint64
	if petIDStr := c.Query("pet_id"); petIDStr != "" {
		id, err := strconv.ParseUint(petIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pet_id"})
			return
		}
		petID = &id
	}

	var ownerID *uint64
	if s := c.Query("owner_id"); s != "" {
		id, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner_id"})
			return
		}
		ownerID = &id
	}

	trimmings, total, err := h.svc.Trimming.List(c.Request.Context(), clinicID, petID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: trimmings, Total: total, Page: page, Limit: limit})
}

// GetTrimming godoc
// @Summary トリミング詳細取得
// @Description 指定IDのトリミング記録を取得する
// @Tags Trimmings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングID"
// @Success 200 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/{id} [get]
func (h *Handler) GetTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	trimming, err := h.svc.Trimming.GetByID(c.Request.Context(), clinicID, id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, trimming)
}

// CreateTrimming godoc
// @Summary トリミング作成
// @Description 新しいトリミング記録を作成する
// @Tags Trimmings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.TrimmingRecord true "トリミング情報"
// @Success 201 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings [post]
func (h *Handler) CreateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	var input model.TrimmingRecord
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Trimming.Create(c.Request.Context(), clinicID, &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateTrimming godoc
// @Summary トリミング更新
// @Description 指定IDのトリミング記録を更新する
// @Tags Trimmings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "トリミングID"
// @Param input body model.TrimmingRecord true "更新するトリミング情報"
// @Success 200 {object} model.TrimmingRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /trimmings/{id} [put]
func (h *Handler) UpdateTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.TrimmingRecord
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Trimming.Update(c.Request.Context(), clinicID, &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *Handler) DeleteTrimming(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Trimming.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
