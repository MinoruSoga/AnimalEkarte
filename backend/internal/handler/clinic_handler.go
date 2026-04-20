package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// ListClinics godoc
// scope=all: 全クリニック一覧を返す（hospital-settings の can_view 権限が必要）
// scope なし: staff_clinic_assignments に紐づくクリニック一覧を返す
func (h *Handler) ListClinics(c *gin.Context) {
	scope := c.Query("scope")

	if scope == "all" {
		if !h.hasPermission(c, string(model.ResourceHospitalSettings), "view") {
			RespondError(c, apperrors.WrapForbidden("医院マスタの閲覧権限が必要です"))
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
	rules, err := h.svc.EffectivePermission.GetEffectivePermissions(c.Request.Context(), staffID)
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
	id, ok := parseIDParam(c, "id")
	if !ok {
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
	id, ok := parseIDParam(c, "id")
	if !ok {
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
// hospital-settings.can_create 権限が必要（RequirePermission ミドルウェアで事前検査済み）
func (h *Handler) CreateClinic(c *gin.Context) {
	var req createClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	result, err := h.svc.Clinic.CreateClinic(c.Request.Context(), &service.CreateClinicInput{
		Name:               req.Name,
		PostalCode:         req.PostalCode,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		FaxNumber:          req.FaxNumber,
		RegistrationNumber: req.RegistrationNumber,
		DirectorName:       req.DirectorName,
		Email:              req.Email,
		Website:            req.Website,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// DeleteClinic godoc
// hospital-settings.can_delete 権限が必要（RequirePermission ミドルウェアで事前検査済み）
func (h *Handler) DeleteClinic(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
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
	clinics.POST("", h.RequirePermission(string(model.ResourceHospitalSettings), "create"), h.CreateClinic)
	clinics.PATCH("/:id", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateClinic)
	clinics.DELETE("/:id", h.RequirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteClinic)
}
