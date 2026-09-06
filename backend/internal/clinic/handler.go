package clinic

import "github.com/gin-gonic/gin"

// PermissionMiddleware is clinic's consumer-side view of auth-owned RBAC.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler owns the clinic, company, holiday, and closing-settings HTTP surface.
type Handler struct {
	clinicSvc          Service
	holidaySvc         ClinicHolidayService
	closingSettingsSvc ClosingSettingsService
	companySvc         CompanyService
	requirePermission  PermissionMiddleware
}

// NewHandler constructs the clinic HTTP adapter from consumer-side interfaces.
func NewHandler(
	clinicSvc Service,
	holidaySvc ClinicHolidayService,
	closingSettingsSvc ClosingSettingsService,
	companySvc CompanyService,
	requirePermission PermissionMiddleware,
) *Handler {
	RegisterContactBindingValidators()
	return &Handler{
		clinicSvc:          clinicSvc,
		holidaySvc:         holidaySvc,
		closingSettingsSvc: closingSettingsSvc,
		companySvc:         companySvc,
		requirePermission:  requirePermission,
	}
}
