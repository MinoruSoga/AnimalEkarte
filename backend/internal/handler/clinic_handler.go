// Package handler keeps the legacy clinic HTTP surface during BE9 migration.
// Route implementations live in internal/clinic; hasPermission remains here
// temporarily because global legacy permission middleware still consumes it.
package handler

import (
	"github.com/gin-gonic/gin"

	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
)

func (h *Handler) clinicDomainHandler() *clinicdomain.Handler {
	if h.svc == nil {
		return clinicdomain.NewHandler(nil, nil, nil, nil, h.RequirePermission)
	}
	return clinicdomain.NewHandler(
		h.svc.Clinic,
		h.svc.ClinicHoliday,
		h.svc.ClosingSettings,
		h.svc.Company,
		h.RequirePermission,
	)
}

func (h *Handler) ListClinics(c *gin.Context)  { h.clinicDomainHandler().ListClinics(c) }
func (h *Handler) GetClinic(c *gin.Context)    { h.clinicDomainHandler().GetClinic(c) }
func (h *Handler) CreateClinic(c *gin.Context) { h.clinicDomainHandler().CreateClinic(c) }
func (h *Handler) UpdateClinic(c *gin.Context) { h.clinicDomainHandler().UpdateClinic(c) }
func (h *Handler) DeleteClinic(c *gin.Context) { h.clinicDomainHandler().DeleteClinic(c) }
func (h *Handler) RegisterClinicRoutes(rg *gin.RouterGroup) {
	h.clinicDomainHandler().RegisterClinicRoutes(rg)
}

// hasPermission is the compatibility authorization hook used by legacy global
// permission middleware and discount guards. New domain handlers consume the
// auth-owned RequirePermission method value instead.
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
	clinicID, ok := extractClinicID(c)
	if !ok {
		return false
	}
	rules, err := h.svc.EffectivePermission.GetEffectivePermissions(c.Request.Context(), staffID, clinicID)
	if err != nil {
		return false
	}
	for i := range rules {
		rule := &rules[i]
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
