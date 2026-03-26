package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListClinics godoc
func (h *Handler) ListClinics(c *gin.Context) {
	clinics, err := h.svc.Clinic.ListClinics(c.Request.Context())
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinics)
}

// GetClinic godoc
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
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
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	input := &service.UpdateClinicInput{
		Name:               req.Name,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		RegistrationNumber: req.RegistrationNumber,
		DirectorName:       req.DirectorName,
		Email:              req.Email,
		Website:            req.Website,
		LogoURL:            req.LogoURL,
		IsActive:           req.IsActive,
		StandardTaxRate:    req.StandardTaxRate,
		ReducedTaxRate:     req.ReducedTaxRate,
	}
	result, err := h.svc.Clinic.UpdateClinic(c.Request.Context(), id, input)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toClinicResponse(result))
}

// CreateClinic godoc
func (h *Handler) CreateClinic(c *gin.Context) {
	var req createClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
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
func (h *Handler) DeleteClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Clinic.DeleteClinic(c.Request.Context(), id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicRoutes はクリニック設定関連のルートを登録する
func (h *Handler) RegisterClinicRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics")
	clinics.GET("", h.ListClinics)
	clinics.POST("", h.CreateClinic)
	clinics.GET("/:id", h.GetClinic)
	clinics.PATCH("/:id", h.UpdateClinic)
	clinics.DELETE("/:id", h.DeleteClinic)
}
