package clinic

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ListClinics godoc
// scope=all: 全クリニック一覧を返す（ルートの hospital-settings.view で認可。system_admin 不要）
// scope なし: staff_clinic_assignments に紐づくクリニック一覧を返す
func (h *Handler) ListClinics(c *gin.Context) {
	query := NewListClinicQuery(c.Request.URL.Query())

	if query.Scope == "all" {
		// hospital-settings 画面の医院マスタは割当外拠点も含む全件が必要（docs/spec/screens/19-clinic-settings.md）。
		// 認可は RegisterClinicRoutes の requirePermission(view) に委ねる。
		clinics, err := h.clinicSvc.ListClinics(c.Request.Context())
		if err != nil {
			httpapi.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, httpapi.MapSlice(clinics, ToClinicResponse))
		return
	}

	// デフォルト: staff_clinic_assignments から割当済みクリニック一覧を返す
	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	clinics, err := h.clinicSvc.ListClinicsByStaffID(c.Request.Context(), staffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, httpapi.MapSlice(clinics, ToClinicResponse))
}

func requireSystemAdmin(c *gin.Context) bool {
	isSystemAdmin, ok := httpapi.ExtractIsSystemAdmin(c)
	if !ok {
		return false
	}
	if !isSystemAdmin {
		httpapi.RespondError(c, apperrors.WrapForbidden("system administrator access required"))
		return false
	}
	return true
}

// GetClinic godoc
// system_admin は任意クリニックを取得可能。それ以外は所属クリニックのみ。
func (h *Handler) GetClinic(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "clinic_id")
	if !ok {
		return
	}
	isSystemAdmin, ok := httpapi.ExtractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
		clinicID, ok := httpapi.ExtractClinicID(c)
		if !ok {
			return
		}
		if id != clinicID {
			httpapi.RespondError(c, apperrors.WrapForbidden("cannot access other clinics"))
			return
		}
	}
	clinic, err := h.clinicSvc.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToClinicResponse(clinic))
}

// UpdateClinic godoc
// system_admin は任意クリニックを更新可能。それ以外は所属クリニックのみ。
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, ok := httpapi.ParseIDParam(c, "clinic_id")
	if !ok {
		return
	}
	isSystemAdmin, ok := httpapi.ExtractIsSystemAdmin(c)
	if !ok {
		return
	}
	if !isSystemAdmin {
		clinicID, ok := httpapi.ExtractClinicID(c)
		if !ok {
			return
		}
		if id != clinicID {
			httpapi.RespondError(c, apperrors.WrapForbidden("cannot update other clinics"))
			return
		}
	}
	var req UpdateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}

	result, err := h.clinicSvc.UpdateClinic(c.Request.Context(), id, req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ToClinicResponse(result))
}

// CreateClinic godoc
// system_admin のみ作成可能。
func (h *Handler) CreateClinic(c *gin.Context) {
	if !requireSystemAdmin(c) {
		return
	}
	var req CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(httpapi.ParseBindError(err)))
		return
	}
	result, err := h.clinicSvc.CreateClinic(c.Request.Context(), req.ToServiceInput())
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/api/v1/clinics/%d", result.ID))
	c.JSON(http.StatusCreated, ToClinicResponse(result))
}

// DeleteClinic godoc
// system_admin のみ削除可能。
func (h *Handler) DeleteClinic(c *gin.Context) {
	if !requireSystemAdmin(c) {
		return
	}
	id, ok := httpapi.ParseIDParam(c, "clinic_id")
	if !ok {
		return
	}
	if err := h.clinicSvc.DeleteClinic(c.Request.Context(), id); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RegisterClinicRoutes はクリニック設定関連のルートを登録する
func (h *Handler) RegisterClinicRoutes(rg *gin.RouterGroup) {
	RegisterContactBindingValidators()
	clinics := rg.Group("/clinics")
	clinics.GET("", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.ListClinics)
	clinics.GET("/:clinic_id", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.GetClinic)
	clinics.POST("", h.requirePermission(string(model.ResourceHospitalSettings), "create"), h.CreateClinic)
	clinics.PATCH("/:clinic_id", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.UpdateClinic)
	clinics.DELETE("/:clinic_id", h.requirePermission(string(model.ResourceHospitalSettings), "delete"), h.DeleteClinic)
}
