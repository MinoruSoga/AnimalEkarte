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
// scope=all: 全クリニック一覧を返す（master-staff の can_view 権限が必要）
// scope なし: staff_clinic_assignments に紐づくクリニック一覧を返す
func (h *Handler) ListClinics(c *gin.Context) {
	scope := c.Query("scope")

	if scope == "all" {
		if !h.hasPermission(c, string(model.ResourceMasterStaff), "view") {
			RespondError(c, apperrors.WrapForbidden("master-staff の閲覧権限が必要です"))
			return
		}
		clinics, err := h.svc.Clinic.ListClinics(c.Request.Context())
		if err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, clinics)
		return
	}

	// デフォルト: staff_clinic_assignments から割当済みクリニック一覧を返す
	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	clinics, err := h.svc.Clinic.ListClinicsByStaffID(c.Request.Context(), staffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, clinics)
}

// hasPermission はユーザーの実効権限を確認する。
// is_system_admin=true は全権限バイパス。
// それ以外は permission_group_rules から判定する。
func (h *Handler) hasPermission(c *gin.Context, resource, action string) bool {
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return false
	}
	if isSystemAdmin {
		return true
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return false
	}
	rules, err := h.repos.PermissionGroup.GetEffectivePermissionsByStaffID(c.Request.Context(), staffID)
	if err != nil {
		return false
	}
	for _, rule := range rules {
		if rule.Resource != resource {
			continue
		}
		switch action {
		case "view":
			return rule.CanView
		case "create":
			return rule.CanCreate
		case "edit":
			return rule.CanEdit
		case "delete":
			return rule.CanDelete
		}
	}
	return false
}

// GetClinic godoc
// system_admin は任意クリニックを取得可能。それ以外は所属クリニックのみ。
func (h *Handler) GetClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
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
// system_admin は任意クリニックを更新可能。それ以外は所属クリニックのみ。
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
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
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
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
	isSystemAdmin, ok := extractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
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
	clinics.GET("/:id", h.GetClinic)

	clinicWrite := clinics.Group("")
	clinicWrite.Use(h.RequirePermission(string(model.ResourceHospitalSettings), "edit"))
	clinicWrite.POST("", h.CreateClinic)
	clinicWrite.PATCH("/:id", h.UpdateClinic)
	clinicWrite.DELETE("/:id", h.DeleteClinic)
}
