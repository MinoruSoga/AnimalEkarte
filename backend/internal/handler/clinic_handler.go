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
// system_admin は全クリニック一覧を返す。それ以外は所属クリニックのみ返す。
func (h *Handler) ListClinics(c *gin.Context) {
	userType, ok := extractUserType(c)
	if !ok {
		return
	}

	if userType == model.UserTypeSystemAdmin {
		clinics, err := h.svc.Clinic.ListClinics(c.Request.Context())
		if err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, clinics)
		return
	}

	// system_admin 以外: JWT の clinic_id に対応する 1 件のみ返す
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	clinic, err := h.svc.Clinic.GetClinicByID(c.Request.Context(), clinicID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, []model.Clinic{*clinic})
}

// GetClinic godoc
// system_admin は任意クリニックを取得可能。それ以外は所属クリニックのみ。
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		clinicID, ok := extractClinicID(c)
		if !ok {
			return
		}
		if id != clinicID {
			RespondError(c, apperrors.WrapForbidden("cannot access other clinics"))
			return
		}
	}
	clinic, err := h.svc.Clinic.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinic)
}

// UpdateClinic godoc
// system_admin は任意クリニックを更新可能。clinic_admin は所属クリニックのみ。
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		clinicID, ok := extractClinicID(c)
		if !ok {
			return
		}
		if id != clinicID {
			RespondError(c, apperrors.WrapForbidden("cannot update other clinics"))
			return
		}
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
// system_admin のみ実行可能
func (h *Handler) CreateClinic(c *gin.Context) {
	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		RespondError(c, apperrors.WrapForbidden("clinic creation requires system_admin"))
		return
	}

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
// system_admin のみ実行可能
func (h *Handler) DeleteClinic(c *gin.Context) {
	userType, ok := extractUserType(c)
	if !ok {
		return
	}
	if userType != model.UserTypeSystemAdmin {
		RespondError(c, apperrors.WrapForbidden("clinic deletion requires system_admin"))
		return
	}

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
