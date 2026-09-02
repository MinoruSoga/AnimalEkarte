package staff

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware is staff's consumer-side view of RBAC middleware.
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// PermissionChecker is staff's consumer-side view of an RBAC permission check.
type PermissionChecker func(c *gin.Context, resource, action string) bool

type handlerServices struct {
	Staff                 StaffService
	StaffClinicAssignment StaffClinicAssignmentService
	Occupation            OccupationService
	ShiftEntry            ShiftEntryService
	ShiftTemplate         ShiftTemplateService
}

// Handler owns staff, occupation, and shift HTTP behavior.
type Handler struct {
	svc               *handlerServices
	requirePermission PermissionMiddleware
	hasPermission     PermissionChecker
}

// NewHandler constructs the staff HTTP boundary.
func NewHandler(
	staffService StaffService,
	assignmentService StaffClinicAssignmentService,
	occupationService OccupationService,
	shiftEntryService ShiftEntryService,
	shiftTemplateService ShiftTemplateService,
	requirePermission PermissionMiddleware,
) *Handler {
	return NewHandlerWithPermissionChecker(
		staffService,
		assignmentService,
		occupationService,
		shiftEntryService,
		shiftTemplateService,
		requirePermission,
		nil,
	)
}

// NewHandlerWithPermissionChecker constructs the staff HTTP boundary with
// conditional permission checks.
func NewHandlerWithPermissionChecker(
	staffService StaffService,
	assignmentService StaffClinicAssignmentService,
	occupationService OccupationService,
	shiftEntryService ShiftEntryService,
	shiftTemplateService ShiftTemplateService,
	requirePermission PermissionMiddleware,
	hasPermission PermissionChecker,
) *Handler {
	return &Handler{
		svc: &handlerServices{
			Staff:                 staffService,
			StaffClinicAssignment: assignmentService,
			Occupation:            occupationService,
			ShiftEntry:            shiftEntryService,
			ShiftTemplate:         shiftTemplateService,
		},
		requirePermission: requirePermission,
		hasPermission:     hasPermission,
	}
}

func (h *Handler) RequirePermission(resource, action string) gin.HandlerFunc {
	if h.requirePermission != nil {
		return h.requirePermission(resource, action)
	}
	return func(c *gin.Context) {
		httpapi.RespondError(c, apperrors.WrapInternalServerError("permission middleware is not configured"))
		c.Abort()
	}
}

func extractClinicID(c *gin.Context) (uint64, bool) {
	return httpapi.ExtractClinicID(c)
}

func extractIsSystemAdmin(c *gin.Context) (bool, bool) {
	return httpapi.ExtractIsSystemAdmin(c)
}

// peekedSystemAdmin is fail-closed: a missing or invalid is_system_admin
// peek is treated as non-admin and never opens a membership bypass.
func peekedSystemAdmin(c *gin.Context) bool {
	isAdmin, ok := httpapi.PeekIsSystemAdmin(c)
	return ok && isAdmin
}

func optionalStaffID(c *gin.Context) *uint64 {
	return httpapi.OptionalStaffID(c)
}

func parseIDParam(c *gin.Context, key string) (uint64, bool) { //nolint:unparam // key is part of the httpapi helper signature; current callers all use "id"
	return httpapi.ParseIDParam(c, key)
}

func RespondError(c *gin.Context, err error) {
	httpapi.RespondError(c, err)
}

func mapSlice[M, R any](items []M, convert func(*M) R) []R {
	return httpapi.MapSlice(items, convert)
}

func localTime(value time.Time) time.Time {
	return httpapi.LocalTime(value)
}

func localTimeRFC3339(value time.Time) string {
	return httpapi.LocalTimeRFC3339(value)
}

type reorderRequest = httpapi.ReorderRequest

func parseOptionalUintQueryFilter(value, field string) (*uint64, error) {
	return httpapi.ParseOptionalUint64Field(value, field)
}

func (h *Handler) resolveStaffWithClinic(c *gin.Context) (clinicID, staffID uint64, ok bool) {
	clinicID, ok = extractClinicID(c)
	if !ok {
		return 0, 0, false
	}
	staffID, ok = parseIDParam(c, "id")
	if !ok {
		return 0, 0, false
	}
	if peekedSystemAdmin(c) {
		return clinicID, staffID, true
	}
	if err := h.svc.Staff.VerifyClinicMembership(c.Request.Context(), staffID, clinicID); err != nil {
		RespondError(c, err)
		return 0, 0, false
	}
	return clinicID, staffID, true
}

func attachPermissionAssignmentAudit(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		c.Abort()
		return
	}
	targetStaffID, ok := parseIDParam(c, "id")
	if !ok {
		c.Abort()
		return
	}
	actorStaffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		c.Abort()
		return
	}
	ctx := withPermissionAssignmentAudit(c.Request.Context(), PermissionAssignmentAudit{
		ClinicID:      clinicID,
		ActorStaffID:  actorStaffID,
		TargetStaffID: targetStaffID,
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.Header.Get("User-Agent"),
	})
	c.Request = c.Request.WithContext(ctx)
	c.Next()
}

// RegisterRoutes registers every staff-owned route on the protected API group.
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	h.registerMasterRoutes(protected)
	h.RegisterShiftRoutes(protected)
	h.RegisterShiftTemplateRoutes(protected)
}

func (h *Handler) registerMasterRoutes(protected *gin.RouterGroup) {
	masters := protected.Group("/masters")
	perm := h.RequirePermission

	masters.GET("/staffs", perm(string(model.ResourceMasterStaff), "view"), h.ListStaffs)
	masters.POST("/staffs", perm(string(model.ResourceMasterStaff), "create"), h.CreateStaff)
	masters.PATCH("/staffs/reorder", perm(string(model.ResourceMasterStaff), "edit"), h.ReorderStaffs)
	masters.GET("/staffs/:id", perm(string(model.ResourceMasterStaff), "view"), h.GetStaff)
	masters.PATCH("/staffs/:id", perm(string(model.ResourceMasterStaff), "edit"), h.UpdateStaff)
	masters.DELETE("/staffs/:id", perm(string(model.ResourceMasterStaff), "delete"), h.DeleteStaff)
	masters.GET("/staffs/:id/permission-groups", perm(string(model.ResourceMasterStaff), "view"), h.GetStaffPermissionGroups)
	masters.PUT(
		"/staffs/:id/permission-groups",
		perm(string(model.ResourceMasterStaff), "edit"),
		perm(string(model.ResourceMasterPermission), "edit"),
		attachPermissionAssignmentAudit,
		h.SetStaffPermissionGroups,
	)
	masters.GET("/staffs/:id/clinics", perm(string(model.ResourceMasterStaff), "view"), h.GetStaffClinicAssignments)
	masters.PUT(
		"/staffs/:id/clinics",
		perm(string(model.ResourceMasterStaff), "edit"),
		attachPermissionAssignmentAudit, // AUS-01: actor/target metadata for fail-closed clinic replace audit
		h.SetStaffClinicAssignments,
	)
	masters.GET("/staffs/:id/excluded-reservation-types", perm(string(model.ResourceMasterStaff), "view"), h.GetStaffExcludedReservationTypes)
	masters.PUT("/staffs/:id/excluded-reservation-types", perm(string(model.ResourceMasterStaff), "edit"), h.SetStaffExcludedReservationTypes)
	masters.GET("/staffs/:id/capable-reservation-types", perm(string(model.ResourceMasterStaff), "view"), h.GetStaffCapableReservationTypes)
	masters.PUT("/staffs/:id/capable-reservation-types", perm(string(model.ResourceMasterStaff), "edit"), h.SetStaffCapableReservationTypes)

	masters.GET("/occupations", perm(string(model.ResourceMasterStaff), "view"), h.ListOccupations)
	masters.POST("/occupations", perm(string(model.ResourceMasterStaff), "create"), h.CreateOccupation)
	masters.PATCH("/occupations/reorder", perm(string(model.ResourceMasterStaff), "edit"), h.ReorderOccupations)
	masters.GET("/occupations/:id", perm(string(model.ResourceMasterStaff), "view"), h.GetOccupation)
	masters.PATCH("/occupations/:id", perm(string(model.ResourceMasterStaff), "edit"), h.UpdateOccupation)
	masters.DELETE("/occupations/:id", perm(string(model.ResourceMasterStaff), "delete"), h.DeleteOccupation)
}
