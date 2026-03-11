package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetAllVaccinations godoc
// @Summary ワクチン接種一覧取得
// @Description 登録されているワクチン接種の一覧を取得します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Success 200 {array} model.VaccinationRecord
// @Failure 500 {object} map[string]string
// @Router /vaccinations [get]
// @Security ApiKeyAuth
func (h *Handler) GetAllVaccinations(c *gin.Context) {
	ctx := c.Request.Context()

	vacs, err := h.svc.GetAllVaccinations(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get vaccinations", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, vacs)
}

// GetVaccinationByID godoc
// @Summary ワクチン接種詳細取得
// @Description 指定されたIDのワクチン接種情報を取得します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "ワクチン接種ID (UUID)"
// @Success 200 {object} model.VaccinationRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [get]
func (h *Handler) GetVaccinationByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	vac, err := h.svc.GetVaccinationByID(ctx, id)
	if err != nil {
		h.handleError(c, err, "vaccination", id)
		return
	}

	c.JSON(http.StatusOK, vac)
}

// GetVaccinationsByPetID godoc
// @Summary ペットのワクチン接種一覧取得
// @Description 指定されたペットIDのワクチン接種一覧を取得します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param petId path string true "ペットID (UUID)"
// @Success 200 {array} model.VaccinationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pets/{petId}/vaccinations [get]
func (h *Handler) GetVaccinationsByPetID(c *gin.Context) {
	ctx := c.Request.Context()
	petID := c.Param("petId")

	vacs, err := h.svc.GetVaccinationsByPetID(ctx, petID)
	if err != nil {
		h.handleError(c, err, "vaccination", petID)
		return
	}

	c.JSON(http.StatusOK, vacs)
}

// GetVaccinationsByOwnerID godoc
// @Summary 飼い主のワクチン接種一覧取得
// @Description 指定された飼い主IDのワクチン接種一覧を取得します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param ownerId path string true "飼い主ID (UUID)"
// @Success 200 {array} model.VaccinationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /owners/{ownerId}/vaccinations [get]
func (h *Handler) GetVaccinationsByOwnerID(c *gin.Context) {
	ctx := c.Request.Context()
	ownerID := c.Param("ownerId")

	vacs, err := h.svc.GetVaccinationsByOwnerID(ctx, ownerID)
	if err != nil {
		h.handleError(c, err, "vaccination", ownerID)
		return
	}

	c.JSON(http.StatusOK, vacs)
}

// CreateVaccination godoc
// @Summary ワクチン接種作成
// @Description 新しいワクチン接種を作成します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param vac body model.CreateVaccinationRecordRequest true "ワクチン接種情報"
// @Success 201 {object} model.VaccinationRecord
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations [post]
func (h *Handler) CreateVaccination(c *gin.Context) {
	ctx := c.Request.Context()
	var req model.CreateVaccinationRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	vac, err := h.svc.CreateVaccination(ctx, &req)
	if err != nil {
		h.handleError(c, err, "vaccination", "")
		return
	}

	c.JSON(http.StatusCreated, vac)
}

// UpdateVaccination godoc
// @Summary ワクチン接種更新
// @Description 既存のワクチン接種を更新します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "ワクチン接種ID (UUID)"
// @Param vac body model.UpdateVaccinationRecordRequest true "ワクチン接種情報"
// @Success 200 {object} model.VaccinationRecord
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [put]
func (h *Handler) UpdateVaccination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req model.UpdateVaccinationRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.ErrorContext(ctx, "invalid request body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	vac, err := h.svc.UpdateVaccination(ctx, id, &req)
	if err != nil {
		h.handleError(c, err, "vaccination", id)
		return
	}

	c.JSON(http.StatusOK, vac)
}

// DeleteVaccination godoc
// @Summary ワクチン接種削除
// @Description 指定されたワクチン接種を削除します
// @Tags vaccinations
// @Accept json
// @Produce json
// @Param id path string true "ワクチン接種ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vaccinations/{id} [delete]
func (h *Handler) DeleteVaccination(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.svc.DeleteVaccination(ctx, id); err != nil {
		h.handleError(c, err, "vaccination", id)
		return
	}

	c.Status(http.StatusNoContent)
}
