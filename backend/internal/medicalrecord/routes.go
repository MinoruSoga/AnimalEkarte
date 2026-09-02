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
// implementation. medicalrecord never imports internal/auth. Mirrors
// internal/manualarticle's identically named type (BE9-2B pilot).
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point. Exactly one exported RegisterRoutes
// per migrated dir is required by internal/apicontract/openapi_route_drift_test.go's
// buildFuncsFromDir walker, which keys discovered funcs by bare name — same-named
// RegisterRoutes methods on separate structs would collide in that map and silently
// drop route sets from drift coverage. The per-entity handlers stay separate structs so each
// only holds the service(s) it actually needs (Go/Gin guideline: consumer declares minimal
// dependencies) — Handler is purely a routing composition, not a new aggregate.
type Handler struct {
	diagnosis             *DiagnosisHandler
	examType              *ExamTypeHandler
	chiefComplaint        *ChiefComplaintHandler
	checkup               *CheckupHandler
	checkupType           *CheckupTypeHandler
	vaccine               *VaccineHandler
	vaccination           *VaccinationHandler
	prescription          *PrescriptionHandler
	inquiry               *InquiryHandler
	inquiryTemplate       *InquiryTemplateHandler
	labImport             *LabImportHandler
	labReport             *LabReportHandler
	vital                 *VitalHandler
	clinicalPlan          *ClinicalPlanHandler
	medicalRecordImage    *MedicalRecordImageHandler
	treatment             *TreatmentHandler
	hospitalization       *HospitalizationHandler
	hospitalizationPlan   *HospitalizationPlanHandler
	dailyRecord           *DailyRecordHandler
	carePlanItem          *CarePlanItemHandler
	consultation          *ConsultationHandler
	procedure             *ProcedureHandler
	medicine              *MedicineHandler
	medicineDoseParam     *MedicineDoseParamHandler
	cage                  *CageHandler
	treatmentPlan         *TreatmentPlanHandler
	medicalRecord         *MedicalRecordHandler
	medicalRecordAddendum *MedicalRecordAddendumHandler
	examination           *ExaminationHandler
	checkupPackageImport  *CheckupPackageImportHandler
	requirePermission     PermissionMiddleware
}

// NewHandler initializes a Handler.
func NewHandler(
	diagnosis *DiagnosisHandler,
	examType *ExamTypeHandler,
	chiefComplaint *ChiefComplaintHandler,
	checkup *CheckupHandler,
	checkupType *CheckupTypeHandler,
	vaccine *VaccineHandler,
	vaccination *VaccinationHandler,
	prescription *PrescriptionHandler,
	inquiry *InquiryHandler,
	inquiryTemplate *InquiryTemplateHandler,
	labImport *LabImportHandler,
	labReport *LabReportHandler,
	vital *VitalHandler,
	clinicalPlan *ClinicalPlanHandler,
	medicalRecordImage *MedicalRecordImageHandler,
	treatment *TreatmentHandler,
	hospitalization *HospitalizationHandler,
	hospitalizationPlan *HospitalizationPlanHandler,
	dailyRecord *DailyRecordHandler,
	carePlanItem *CarePlanItemHandler,
	consultation *ConsultationHandler,
	procedure *ProcedureHandler,
	medicine *MedicineHandler,
	medicineDoseParam *MedicineDoseParamHandler,
	cage *CageHandler,
	treatmentPlan *TreatmentPlanHandler,
	medicalRecord *MedicalRecordHandler,
	medicalRecordAddendum *MedicalRecordAddendumHandler,
	examination *ExaminationHandler,
	checkupPackageImport *CheckupPackageImportHandler,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		diagnosis:             diagnosis,
		examType:              examType,
		chiefComplaint:        chiefComplaint,
		checkup:               checkup,
		checkupType:           checkupType,
		vaccine:               vaccine,
		vaccination:           vaccination,
		prescription:          prescription,
		inquiry:               inquiry,
		inquiryTemplate:       inquiryTemplate,
		labImport:             labImport,
		labReport:             labReport,
		vital:                 vital,
		clinicalPlan:          clinicalPlan,
		medicalRecordImage:    medicalRecordImage,
		treatment:             treatment,
		hospitalization:       hospitalization,
		hospitalizationPlan:   hospitalizationPlan,
		dailyRecord:           dailyRecord,
		carePlanItem:          carePlanItem,
		consultation:          consultation,
		procedure:             procedure,
		medicine:              medicine,
		medicineDoseParam:     medicineDoseParam,
		cage:                  cage,
		treatmentPlan:         treatmentPlan,
		medicalRecord:         medicalRecord,
		medicalRecordAddendum: medicalRecordAddendum,
		examination:           examination,
		checkupPackageImport:  checkupPackageImport,
		requirePermission:     requirePermission,
	}
}

type medicalRoutePerm func(resource model.Resource, action string) gin.HandlerFunc

// RegisterRoutes registers this package's HTTP routes on the authenticated
// protected group. Path, method, and RBAC triples match the pre-BE9 handler
// registration (BUG-122/BUG-125). Snapshot tests cover path/method/handler name;
// permission arguments are transcribed here because the snapshot gate does not
// capture middleware.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.requirePermission(string(resource), action)
	}
	masters := rg.Group("/masters")
	h.registerEarlyMasterRoutes(masters, perm)
	h.registerVaccinationAndCheckupRoutes(rg, perm)
	records := rg.Group("/medical-records")
	h.registerMedicalRecordChildRoutes(records, perm)
	h.registerPetTreatmentHistoryRoute(rg, perm)
	h.registerLateMasterRoutes(masters, perm)
	hospitalizations := rg.Group("/hospitalizations")
	h.registerHospitalizationRoutes(hospitalizations, perm)
	h.registerTreatmentPlanRoutes(records, hospitalizations, perm)
	h.registerMedicalRecordCoreRoutes(records, perm)
	h.registerExaminationAndPackageRoutes(rg, perm)
	h.registerLabRoutes(rg, perm)
}
