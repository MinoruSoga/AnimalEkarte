package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action).
// This is medicalrecord's consumer-side view of the permission-checking middleware — ADR-006
// classifies permission_middleware.go as target:auth, and medicalrecord (topologically before
// auth in ADR-006's permitted dependency graph) must not depend on the not-yet-migrated auth
// domain package. The composition root (cmd/api/main.go) supplies the concrete
// implementation (today, the transitional *handler.Handler.RequirePermission method value;
// once BE9-2C/2D migrates auth, an *auth.Handler method value instead) — medicalrecord never
// imports internal/handler or internal/auth. Mirrors internal/manualarticle's identically
// named type (BE9-2B pilot).
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point. Exactly one exported RegisterRoutes
// per migrated dir is required by internal/apicontract/openapi_route_drift_test.go's
// buildFuncsFromDir walker, which keys discovered funcs by bare name — three same-named
// RegisterRoutes methods on three separate structs would collide in that map and silently
// drop two of the three route sets from drift coverage. The per-entity handlers
// (DiagnosisHandler/ExamTypeHandler/ChiefComplaintHandler) stay separate structs so each
// only holds the service(s) it actually needs (Go/Gin guideline: consumer declares minimal
// dependencies) — Handler is purely a routing composition, not a new aggregate.
type Handler struct {
	diagnosis         *DiagnosisHandler
	examType          *ExamTypeHandler
	chiefComplaint    *ChiefComplaintHandler
	requirePermission PermissionMiddleware
}

// NewHandler initializes a Handler.
func NewHandler(diagnosis *DiagnosisHandler, examType *ExamTypeHandler, chiefComplaint *ChiefComplaintHandler, requirePermission PermissionMiddleware) *Handler {
	return &Handler{
		diagnosis:         diagnosis,
		examType:          examType,
		chiefComplaint:    chiefComplaint,
		requirePermission: requirePermission,
	}
}

// RegisterRoutes registers this slice's master routes under rg's /masters sub-group,
// matching the exact (method, path, permission) triples previously registered by
// internal/handler/master_routes.go's RegisterMasterRoutes for these three entities (BUG-122/
// BUG-125 RBAC parity — see this package's tests for the before==after proof).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")

	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.requirePermission(string(resource), action)
	}

	// Examination Types
	masters.GET("/examination-types", perm(model.ResourceMasterMedical, "view"), h.examType.ListExaminationTypes)
	masters.POST("/examination-types", perm(model.ResourceMasterMedical, "create"), h.examType.CreateExaminationType)
	masters.PATCH("/examination-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.examType.ReorderExaminationTypes)
	masters.GET("/examination-types/:id", perm(model.ResourceMasterMedical, "view"), h.examType.GetExaminationType)
	masters.PATCH("/examination-types/:id", perm(model.ResourceMasterMedical, "edit"), h.examType.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", perm(model.ResourceMasterMedical, "delete"), h.examType.DeleteExaminationType)

	// Diagnosis Categories
	masters.GET("/diagnosis-types", perm(model.ResourceMasterMedical, "view"), h.diagnosis.ListDiagnosisTypes)
	masters.POST("/diagnosis-types", perm(model.ResourceMasterMedical, "create"), h.diagnosis.CreateDiagnosisType)
	masters.PATCH("/diagnosis-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.diagnosis.ReorderDiagnosisTypes)
	masters.GET("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "view"), h.diagnosis.GetDiagnosisType)
	masters.PATCH("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "edit"), h.diagnosis.UpdateDiagnosisType)
	masters.DELETE("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "delete"), h.diagnosis.DeleteDiagnosisType)

	// Diagnosis Names
	masters.GET("/diagnosis-names", perm(model.ResourceMasterMedical, "view"), h.diagnosis.ListDiagnosisNames)
	masters.GET("/diagnosis-names/all", perm(model.ResourceMasterMedical, "view"), h.diagnosis.ListDiagnosisNamesAll)
	masters.POST("/diagnosis-names", perm(model.ResourceMasterMedical, "create"), h.diagnosis.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", perm(model.ResourceMasterMedical, "edit"), h.diagnosis.ReorderDiagnosisNames)
	masters.GET("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "view"), h.diagnosis.GetDiagnosisName)
	masters.PATCH("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "edit"), h.diagnosis.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "delete"), h.diagnosis.DeleteDiagnosisName)

	// Chief Complaint Categories
	masters.GET("/chief-complaint-types", perm(model.ResourceMasterMedical, "view"), h.chiefComplaint.ListChiefComplaints)
	masters.POST("/chief-complaint-types", perm(model.ResourceMasterMedical, "create"), h.chiefComplaint.CreateChiefComplaint)
	masters.PATCH("/chief-complaint-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.chiefComplaint.ReorderChiefComplaints)
	masters.GET("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "view"), h.chiefComplaint.GetChiefComplaint)
	masters.PATCH("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "edit"), h.chiefComplaint.UpdateChiefComplaint)
	masters.DELETE("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "delete"), h.chiefComplaint.DeleteChiefComplaint)
}
