package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// ListVaccinations godoc
// @Summary 予防接種一覧取得
// @Description 予防接種記録の一覧をページネーション付きで取得する
// @Tags Vaccinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "ページ番号 (default: 1)"
// @Param limit query int false "件数 (1-100, default: 20)"
// @Param pet_id query string false "ペットIDフィルター"
// @Param owner_id query string false "飼主IDフィルター"
// @Success 200 {object} PaginatedResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations [get]
func (h *Handler) ListVaccinations(c *gin.Context) {
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

	vaccinations, total, err := h.svc.Vaccination.List(c.Request.Context(), petID, ownerID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PaginatedResponse{Data: vaccinations, Total: total, Page: page, Limit: limit})
}

// GetVaccination godoc
// @Summary 予防接種詳細取得
// @Description 指定IDの予防接種記録を取得する
// @Tags Vaccinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "予防接種ID"
// @Success 200 {object} model.Vaccination
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [get]
func (h *Handler) GetVaccination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	vaccination, err := h.svc.Vaccination.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, vaccination)
}

// CreateVaccination godoc
// @Summary 予防接種作成
// @Description 新しい予防接種記録を作成する
// @Tags Vaccinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.Vaccination true "予防接種情報"
// @Success 201 {object} model.Vaccination
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations [post]
func (h *Handler) CreateVaccination(c *gin.Context) {
	var input model.Vaccination
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = uuid.New()
	if err := h.svc.Vaccination.Create(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateVaccination godoc
// @Summary 予防接種更新
// @Description 指定IDの予防接種記録を更新する
// @Tags Vaccinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "予防接種ID"
// @Param input body model.Vaccination true "更新する予防接種情報"
// @Success 200 {object} model.Vaccination
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [patch]
func (h *Handler) UpdateVaccination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Vaccination
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.ID = id
	if err := h.svc.Vaccination.Update(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// DeleteVaccination godoc
// @Summary 予防接種削除
// @Description 指定IDの予防接種記録を削除する
// @Tags Vaccinations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "予防接種ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [delete]
func (h *Handler) DeleteVaccination(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Vaccination.Delete(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
