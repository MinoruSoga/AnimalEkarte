package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// GetCompany godoc
// @Summary 企業情報取得
// @Description 動物病院グループの企業情報を取得する
// @Tags Company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.Company
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /company [get]
func (h *Handler) GetCompany(c *gin.Context) {
	company, err := h.svc.Clinic.GetCompany(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, company)
}

// UpdateCompany godoc
// @Summary 企業情報更新
// @Description 動物病院グループの企業情報を更新する
// @Tags Company
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body model.Company true "更新する企業情報"
// @Success 200 {object} model.Company
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /company [put]
func (h *Handler) UpdateCompany(c *gin.Context) {
	var input model.Company
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Clinic.UpdateCompany(c.Request.Context(), &input); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, input)
}

// ListClinics godoc
// @Summary クリニック一覧取得
// @Description 全クリニック（院）の一覧を取得する
// @Tags Clinics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Clinic
// @Failure 500 {object} map[string]string
// @Router /clinics [get]
func (h *Handler) ListClinics(c *gin.Context) {
	clinics, err := h.svc.Clinic.ListClinics(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinics)
}

// GetClinic godoc
// @Summary クリニック詳細取得
// @Description 指定IDのクリニック（院）を取得する
// @Tags Clinics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "クリニックID"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [get]
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	clinic, err := h.svc.Clinic.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinic)
}
