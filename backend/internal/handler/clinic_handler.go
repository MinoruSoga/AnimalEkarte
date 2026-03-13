package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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
// @Param id path integer true "クリニックID"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [get]
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
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

// UpdateClinic godoc
// @Summary クリニック情報更新
// @Description 指定IDのクリニック（院）情報を更新する
// @Tags Clinics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "クリニックID"
// @Param input body model.Clinic true "更新するクリニック情報"
// @Success 200 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [patch]
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input model.Clinic
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clinic, err := h.svc.Clinic.UpdateClinic(c.Request.Context(), id, &input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinic)
}

type createClinicRequest struct {
	Name               string `json:"name"                binding:"required"`
	PostalCode         string `json:"postal_code"`
	Address            string `json:"address"`
	PhoneNumber        string `json:"phone_number"`
	FaxNumber          string `json:"fax_number"`
	RegistrationNumber string `json:"registration_number"`
	DirectorName       string `json:"director_name"`
	Email              string `json:"email"`
	Website            string `json:"website"`
}

// CreateClinic godoc
// @Summary クリニック作成
// @Description 新規クリニックを登録する
// @Tags Clinics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body createClinicRequest true "クリニック登録情報"
// @Success 201 {object} model.Clinic
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics [post]
func (h *Handler) CreateClinic(c *gin.Context) {
	var req createClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clinic := &model.Clinic{
		Name:               req.Name,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		RegistrationNumber: req.RegistrationNumber,
		DirectorName:       req.DirectorName,
		Email:              req.Email,
		Website:            req.Website,
		IsActive:           true,
	}
	result, err := h.svc.Clinic.CreateClinic(c.Request.Context(), clinic)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// DeleteClinic godoc
// @Summary クリニック削除
// @Description 指定IDのクリニックを削除する
// @Tags Clinics
// @Produce json
// @Security BearerAuth
// @Param id path integer true "クリニックID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /clinics/{id} [delete]
func (h *Handler) DeleteClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.Clinic.DeleteClinic(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
