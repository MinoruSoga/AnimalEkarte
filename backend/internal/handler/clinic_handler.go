package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
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

// buildClinicUpdateFields は updateClinicRequest から非 nil フィールドのみを map に変換する。
// GORM のゼロ値問題を回避し、PATCH セマンティクス（未送信フィールドは既存値を保持）を実現する。
func buildClinicUpdateFields(req *updateClinicRequest) map[string]any {
	fields := make(map[string]any)
	if req.Name != nil {
		fields["name"] = *req.Name
	}
	if req.PostalCode != nil {
		fields["postal_code"] = *req.PostalCode
	}
	if req.Address != nil {
		fields["address"] = *req.Address
	}
	if req.PhoneNumber != nil {
		fields["phone_number"] = *req.PhoneNumber
	}
	if req.FaxNumber != nil {
		fields["fax_number"] = *req.FaxNumber
	}
	if req.RegistrationNumber != nil {
		fields["registration_number"] = *req.RegistrationNumber
	}
	if req.DirectorName != nil {
		fields["director_name"] = *req.DirectorName
	}
	if req.Email != nil {
		fields["email"] = *req.Email
	}
	if req.Website != nil {
		fields["website"] = *req.Website
	}
	if req.LogoURL != nil {
		fields["logo_url"] = *req.LogoURL
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}
	if req.StandardTaxRate != nil {
		r := *req.StandardTaxRate
		if r > 0 && r <= 1 {
			fields["standard_tax_rate"] = r
		}
	}
	if req.ReducedTaxRate != nil {
		r := *req.ReducedTaxRate
		if r > 0 && r <= 1 {
			fields["reduced_tax_rate"] = r
		}
	}
	return fields
}

// UpdateClinic godoc
func (h *Handler) UpdateClinic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fields := buildClinicUpdateFields(&req)
	result, err := h.svc.Clinic.UpdateClinic(c.Request.Context(), id, fields)
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

// RegisterClinicRoutes はクリニック設定関連のルートを登録する
func (h *Handler) RegisterClinicRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics")
	clinics.GET("", h.ListClinics)
	clinics.POST("", h.CreateClinic)
	clinics.GET("/:id", h.GetClinic)
	clinics.PATCH("/:id", h.UpdateClinic)
	clinics.DELETE("/:id", h.DeleteClinic)
}
