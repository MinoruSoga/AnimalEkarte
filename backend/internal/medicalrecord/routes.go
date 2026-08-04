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
	examination             *ExaminationHandler
	checkupPackageImport    *CheckupPackageImportHandler
	requirePermission       PermissionMiddleware
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

// RegisterRoutes registers this slice's routes, matching the exact (method, path, permission)
// triples previously registered by internal/handler (master_routes.go's RegisterMasterRoutes,
// handler.go's registerVaccinationRoutesWithAuth/registerMedicalRecordRoutesWithAuth, and
// checkup/prescription/inquiry handler files) for these entities (BUG-122/BUG-125 RBAC parity —
// see this package's route-snapshot test for the before==after proof of path/method/handler
// name; the permission arguments are hand-transcribed here because the snapshot gate does not
// capture middleware, only the trailing handler name). rg is the authenticated `protected`
// group (cmd/api/main.go); the master routes nest under /masters, /vaccinations and /checkups
// are top-level, and the checkup/prescription/inquiry sub-resources nest under /medical-records
// (a second Group on the same prefix as internal/handler's own medical-records routes — gin
// merges by path, same coexistence as the /masters group already shared with RegisterMasterRoutes).
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
	masters.POST("/examination-types/:id/fields", perm(model.ResourceMasterMedical, "create"), h.examType.CreateExaminationTypeField)
	masters.PATCH("/examination-types/:id/fields/reorder", perm(model.ResourceMasterMedical, "edit"), h.examType.ReorderExaminationTypeFields)
	masters.PATCH("/examination-types/:id/fields/:fieldId", perm(model.ResourceMasterMedical, "edit"), h.examType.UpdateExaminationTypeField)
	masters.DELETE("/examination-types/:id/fields/:fieldId", perm(model.ResourceMasterMedical, "delete"), h.examType.DeleteExaminationTypeField)
	masters.PUT("/examination-types/:id/fields/:fieldId/reference-ranges", perm(model.ResourceMasterMedical, "edit"), h.examType.ReplaceExaminationTypeFieldReferenceRanges)

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

	// Vaccines (BE9-2D: moved from master_routes.go; ResourceMasterMedical parity)
	masters.GET("/vaccines", perm(model.ResourceMasterMedical, "view"), h.vaccine.ListVaccines)
	masters.POST("/vaccines", perm(model.ResourceMasterMedical, "create"), h.vaccine.CreateVaccine)
	masters.PATCH("/vaccines/reorder", perm(model.ResourceMasterMedical, "edit"), h.vaccine.ReorderVaccines)
	masters.GET("/vaccines/:id", perm(model.ResourceMasterMedical, "view"), h.vaccine.GetVaccine)
	masters.PATCH("/vaccines/:id", perm(model.ResourceMasterMedical, "edit"), h.vaccine.UpdateVaccine)
	masters.DELETE("/vaccines/:id", perm(model.ResourceMasterMedical, "delete"), h.vaccine.DeleteVaccine)

	// Checkup Types (BE9-2D: moved from master_routes.go; ResourceCheckups parity — NOT
	// ResourceMasterMedical, unlike the sibling medical masters above. :id/fields is view.)
	masters.GET("/checkup-types", perm(model.ResourceCheckups, "view"), h.checkupType.ListCheckupTypes)
	masters.POST("/checkup-types", perm(model.ResourceCheckups, "create"), h.checkupType.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", perm(model.ResourceCheckups, "edit"), h.checkupType.ReorderCheckupTypes)
	masters.GET("/checkup-types/:id", perm(model.ResourceCheckups, "view"), h.checkupType.GetCheckupType)
	masters.PATCH("/checkup-types/:id", perm(model.ResourceCheckups, "edit"), h.checkupType.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", perm(model.ResourceCheckups, "delete"), h.checkupType.DeleteCheckupType)
	masters.GET("/checkup-types/:id/fields", perm(model.ResourceCheckups, "view"), h.checkup.ListCheckupTypeFields)

	// Inquiry Templates (BE9-2D: moved from master_routes.go; ResourceMasterMedical parity)
	masters.GET("/inquiry-templates", perm(model.ResourceMasterMedical, "view"), h.inquiryTemplate.ListInquiryTemplates)
	masters.POST("/inquiry-templates", perm(model.ResourceMasterMedical, "create"), h.inquiryTemplate.CreateInquiryTemplate)
	masters.PATCH("/inquiry-templates/reorder", perm(model.ResourceMasterMedical, "edit"), h.inquiryTemplate.ReorderInquiryTemplates)
	masters.GET("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "view"), h.inquiryTemplate.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "edit"), h.inquiryTemplate.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "delete"), h.inquiryTemplate.DeleteInquiryTemplate)

	// Vaccinations (BE9-2D: moved from handler.go registerVaccinationRoutesWithAuth;
	// per-route ResourceVaccinations guards = BUG-125 parity, NOT a single group-level guard).
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", perm(model.ResourceVaccinations, "view"), h.vaccination.ListVaccinations)
	vaccinations.GET("/:id", perm(model.ResourceVaccinations, "view"), h.vaccination.GetVaccination)
	vaccinations.POST("", perm(model.ResourceVaccinations, "create"), h.vaccination.CreateVaccination)
	vaccinations.PATCH("/:id", perm(model.ResourceVaccinations, "edit"), h.vaccination.UpdateVaccination)
	vaccinations.DELETE("/:id", perm(model.ResourceVaccinations, "delete"), h.vaccination.DeleteVaccination)

	// Global checkups (BE9-2D: moved from handler.go checkup_handler.go RegisterGlobalCheckupRoutes;
	// ResourceCheckups parity).
	checkups := rg.Group("/checkups")
	checkups.GET("", perm(model.ResourceCheckups, "view"), h.checkup.ListGlobalCheckups)
	checkups.GET("/field-results", perm(model.ResourceCheckups, "view"), h.checkup.ListPetCheckupResults)

	// Medical-record sub-resources (BE9-2D: moved from handler.go RegisterCheckupRoutes/
	// RegisterPrescriptionRoutes/RegisterInquiryRoutes). Children follow the parent
	// medical-records permission (ResourceMedicalRecords) — BUG-133 parity, NOT the
	// ResourceCheckups permission the /checkups and /masters/checkup-types routes use.
	records := rg.Group("/medical-records")
	records.GET("/:id/checkups", perm(model.ResourceMedicalRecords, "view"), h.checkup.ListCheckups)
	records.POST("/:id/checkups", perm(model.ResourceMedicalRecords, "create"), h.checkup.CreateCheckup)
	records.PATCH("/:id/checkups/:checkupId", perm(model.ResourceMedicalRecords, "edit"), h.checkup.UpdateCheckup)
	records.DELETE("/:id/checkups/:checkupId", perm(model.ResourceMedicalRecords, "delete"), h.checkup.DeleteCheckup)
	records.GET("/:id/checkups/:checkupId/field-results", perm(model.ResourceMedicalRecords, "view"), h.checkup.ListCheckupFieldResults)
	records.PUT("/:id/checkups/:checkupId/field-results", perm(model.ResourceMedicalRecords, "edit"), h.checkup.ReplaceCheckupFieldResults)
	records.GET("/:id/prescriptions", perm(model.ResourceMedicalRecords, "view"), h.prescription.ListPrescriptions)
	records.POST("/:id/prescriptions", perm(model.ResourceMedicalRecords, "create"), h.prescription.CreatePrescription)
	records.PATCH("/:id/prescriptions/:prescriptionId", perm(model.ResourceMedicalRecords, "edit"), h.prescription.UpdatePrescription)
	records.DELETE("/:id/prescriptions/:prescriptionId", perm(model.ResourceMedicalRecords, "delete"), h.prescription.DeletePrescription)
	records.PATCH("/:id/inquiries", perm(model.ResourceMedicalRecords, "edit"), h.inquiry.UpdateInquiry)

	// Vitals / clinical-plan / images (BE9-2D sub-batch④a: moved from internal/handler
	// handler.go RegisterVitalRoutes/RegisterClinicalPlanRoutes/RegisterMedicalRecordImageRoutes).
	// All guard ResourceMedicalRecords (BUG-125 CRUD parity). NOTE the non-uniform verbs,
	// hand-transcribed verbatim from the pre-move Register funcs: vitals POST is "edit" (NOT
	// "create"), whereas images POST / POST upload are "create". treatment / treatment-plan stay
	// in internal/handler (sub-batch④b).
	records.GET("/:id/vitals", perm(model.ResourceMedicalRecords, "view"), h.vital.ListVitals)
	records.POST("/:id/vitals", perm(model.ResourceMedicalRecords, "edit"), h.vital.CreateVital)
	records.PATCH("/:id/vitals/:vitalId", perm(model.ResourceMedicalRecords, "edit"), h.vital.UpdateVital)
	records.DELETE("/:id/vitals/:vitalId", perm(model.ResourceMedicalRecords, "delete"), h.vital.DeleteVital)
	records.GET("/:id/clinical-plan", perm(model.ResourceMedicalRecords, "view"), h.clinicalPlan.GetClinicalPlan)
	records.PATCH("/:id/clinical-plan", perm(model.ResourceMedicalRecords, "edit"), h.clinicalPlan.UpdateClinicalPlan)
	records.DELETE("/:id/clinical-plan", perm(model.ResourceMedicalRecords, "delete"), h.clinicalPlan.DeleteClinicalPlan)
	records.GET("/:id/images", perm(model.ResourceMedicalRecords, "view"), h.medicalRecordImage.ListMedicalRecordImages)
	records.POST("/:id/images", perm(model.ResourceMedicalRecords, "create"), h.medicalRecordImage.CreateMedicalRecordImage)
	records.POST("/:id/images/upload", perm(model.ResourceMedicalRecords, "create"), h.medicalRecordImage.UploadMedicalRecordImage)
	records.DELETE("/:id/images/:imageId", perm(model.ResourceMedicalRecords, "delete"), h.medicalRecordImage.DeleteMedicalRecordImage)

	// Treatments (BE9-2D sub-batch④b: moved from treatment_handler.go RegisterTreatmentRoutes;
	// ResourceMedicalRecords parity on all five, PUT/PATCH=edit・POST=create・DELETE=delete 逐語転記).
	records.GET("/:id/treatments", perm(model.ResourceMedicalRecords, "view"), h.treatment.ListTreatments)
	records.POST("/:id/treatments", perm(model.ResourceMedicalRecords, "create"), h.treatment.CreateTreatment)
	records.PATCH("/:id/treatments/:treatmentId", perm(model.ResourceMedicalRecords, "edit"), h.treatment.UpdateTreatment)
	records.DELETE("/:id/treatments/:treatmentId", perm(model.ResourceMedicalRecords, "delete"), h.treatment.DeleteTreatment)
	records.PUT("/:id/treatments", perm(model.ResourceMedicalRecords, "edit"), h.treatment.BulkUpdateTreatments)

	// Pet treatment history (#158/#159、BE9-2D sub-batch④b: moved from pet_handler.go's
	// registration line — path/permission 逐語転記。/pets group は internal/handler の pet routes と
	// gin 上で path merge され共存する（/medical-records group の既存共存と同型・param 名 :id 一致）。
	pets := rg.Group("/pets")
	pets.GET("/:id/treatment-history", perm(model.ResourceMedicalRecords, "view"), h.treatment.ListPetTreatmentHistory)

	// Consultations / Procedures / Medicines(+dose-params) / Cages master (BE9-2D ⑥:
	// moved from master_routes.go; resource は旧値逐語転記 — cages のみ
	// ResourceMasterHospitalization、他は ResourceMasterMedical)。
	masters.GET("/consultations", perm(model.ResourceMasterMedical, "view"), h.consultation.ListConsultations)
	masters.POST("/consultations", perm(model.ResourceMasterMedical, "create"), h.consultation.CreateConsultation)
	masters.PATCH("/consultations/reorder", perm(model.ResourceMasterMedical, "edit"), h.consultation.ReorderConsultations)
	masters.GET("/consultations/:id", perm(model.ResourceMasterMedical, "view"), h.consultation.GetConsultation)
	masters.PATCH("/consultations/:id", perm(model.ResourceMasterMedical, "edit"), h.consultation.UpdateConsultation)
	masters.DELETE("/consultations/:id", perm(model.ResourceMasterMedical, "delete"), h.consultation.DeleteConsultation)
	masters.GET("/procedures", perm(model.ResourceMasterMedical, "view"), h.procedure.ListProcedures)
	masters.POST("/procedures", perm(model.ResourceMasterMedical, "create"), h.procedure.CreateProcedure)
	masters.PATCH("/procedures/reorder", perm(model.ResourceMasterMedical, "edit"), h.procedure.ReorderProcedures)
	masters.GET("/procedures/:id", perm(model.ResourceMasterMedical, "view"), h.procedure.GetProcedure)
	masters.PATCH("/procedures/:id", perm(model.ResourceMasterMedical, "edit"), h.procedure.UpdateProcedure)
	masters.DELETE("/procedures/:id", perm(model.ResourceMasterMedical, "delete"), h.procedure.DeleteProcedure)
	masters.GET("/medicines", perm(model.ResourceMasterMedical, "view"), h.medicine.ListMedicines)
	masters.POST("/medicines", perm(model.ResourceMasterMedical, "create"), h.medicine.CreateMedicine)
	masters.PATCH("/medicines/reorder", perm(model.ResourceMasterMedical, "edit"), h.medicine.ReorderMedicines)
	masters.GET("/medicines/:id", perm(model.ResourceMasterMedical, "view"), h.medicine.GetMedicine)
	masters.PATCH("/medicines/:id", perm(model.ResourceMasterMedical, "edit"), h.medicine.UpdateMedicine)
	masters.DELETE("/medicines/:id", perm(model.ResourceMasterMedical, "delete"), h.medicine.DeleteMedicine)
	masters.GET("/medicines/:id/dose-params", perm(model.ResourceMasterMedical, "view"), h.medicineDoseParam.ListMedicineDoseParams)
	masters.PUT("/medicines/:id/dose-params/:species", perm(model.ResourceMasterMedical, "edit"), h.medicineDoseParam.UpsertMedicineDoseParam)
	masters.DELETE("/medicines/:id/dose-params/:species", perm(model.ResourceMasterMedical, "delete"), h.medicineDoseParam.DeleteMedicineDoseParam)
	masters.GET("/cages", perm(model.ResourceMasterHospitalization, "view"), h.cage.ListCages)
	masters.POST("/cages", perm(model.ResourceMasterHospitalization, "create"), h.cage.CreateCage)
	masters.PATCH("/cages/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.cage.ReorderCages)
	masters.GET("/cages/:id", perm(model.ResourceMasterHospitalization, "view"), h.cage.GetCage)
	masters.PATCH("/cages/:id", perm(model.ResourceMasterHospitalization, "edit"), h.cage.UpdateCage)
	masters.DELETE("/cages/:id", perm(model.ResourceMasterHospitalization, "delete"), h.cage.DeleteCage)

	// Hospitalization Plans master (BE9-2D ⑤: moved from master_routes.go;
	// ResourceMasterHospitalization parity 逐語転記).
	masters.GET("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "view"), h.hospitalizationPlan.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "create"), h.hospitalizationPlan.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.hospitalizationPlan.ReorderHospitalizationPlans)
	masters.GET("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "view"), h.hospitalizationPlan.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "edit"), h.hospitalizationPlan.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "delete"), h.hospitalizationPlan.DeleteHospitalizationPlan)

	// Hospitalizations + nested daily-records / care-plan-items (BE9-2D ⑤: moved from
	// handler.go registerHospitalizationRoutesWithAuth / RegisterDailyRecordRoutes /
	// RegisterCarePlanItemRoutes; ResourceHospitalization per-route parity 逐語転記。
	// treatment-plan sub-resource は internal/handler 残置——/hospitalizations group は
	// gin path merge で共存・param 名 :id 一致)。
	hospitalizations := rg.Group("/hospitalizations")
	hospitalizations.GET("", perm(model.ResourceHospitalization, "view"), h.hospitalization.ListHospitalizations)
	hospitalizations.GET("/:id", perm(model.ResourceHospitalization, "view"), h.hospitalization.GetHospitalization)
	hospitalizations.POST("", perm(model.ResourceHospitalization, "create"), h.hospitalization.CreateHospitalization)
	hospitalizations.PATCH("/:id", perm(model.ResourceHospitalization, "edit"), h.hospitalization.UpdateHospitalization)
	hospitalizations.DELETE("/:id", perm(model.ResourceHospitalization, "delete"), h.hospitalization.DeleteHospitalization)
	hospitalizations.POST("/:id/discharge-with-billing", perm(model.ResourceHospitalization, "edit"), h.hospitalization.DischargeWithBilling)
	hospitalizations.GET("/:id/daily-records", perm(model.ResourceHospitalization, "view"), h.dailyRecord.ListDailyRecords)
	hospitalizations.POST("/:id/daily-records", perm(model.ResourceHospitalization, "create"), h.dailyRecord.CreateDailyRecord)
	hospitalizations.GET("/:id/daily-records/:date", perm(model.ResourceHospitalization, "view"), h.dailyRecord.GetDailyRecord)
	hospitalizations.POST("/:id/daily-records/:date/vitals", perm(model.ResourceHospitalization, "create"), h.dailyRecord.AddVitalRecord)
	hospitalizations.POST("/:id/daily-records/:date/care-logs", perm(model.ResourceHospitalization, "create"), h.dailyRecord.AddCareLog)
	hospitalizations.POST("/:id/daily-records/:date/staff-notes", perm(model.ResourceHospitalization, "create"), h.dailyRecord.AddStaffNote)
	hospitalizations.GET("/:id/care-plan-items", perm(model.ResourceHospitalization, "view"), h.carePlanItem.ListCarePlanItems)
	hospitalizations.POST("/:id/care-plan-items", perm(model.ResourceHospitalization, "create"), h.carePlanItem.CreateCarePlanItem)
	hospitalizations.PATCH("/:id/care-plan-items/:itemId", perm(model.ResourceHospitalization, "edit"), h.carePlanItem.UpdateCarePlanItem)
	hospitalizations.DELETE("/:id/care-plan-items/:itemId", perm(model.ResourceHospitalization, "delete"), h.carePlanItem.DeleteCarePlanItem)

	// Treatment plans (BE9-2D ⑥: moved from treatment_plan_handler.go の Register 2関数;
	// medical-records 配下=ResourceMedicalRecords / hospitalizations 配下=ResourceHospitalization 逐語転記)。
	records.GET("/:id/treatment-plans", perm(model.ResourceMedicalRecords, "view"), h.treatmentPlan.ListTreatmentPlansByMedicalRecord)
	records.POST("/:id/treatment-plans", perm(model.ResourceMedicalRecords, "create"), h.treatmentPlan.CreateTreatmentPlanForMedicalRecord)
	records.PATCH("/:id/treatment-plans/:planId", perm(model.ResourceMedicalRecords, "edit"), h.treatmentPlan.UpdateTreatmentPlanInMedicalRecord)
	records.DELETE("/:id/treatment-plans/:planId", perm(model.ResourceMedicalRecords, "delete"), h.treatmentPlan.DeleteTreatmentPlanInMedicalRecord)
	hospitalizations.GET("/:id/treatment-plans", perm(model.ResourceHospitalization, "view"), h.treatmentPlan.ListTreatmentPlansByHospitalization)
	hospitalizations.POST("/:id/treatment-plans", perm(model.ResourceHospitalization, "create"), h.treatmentPlan.CreateTreatmentPlanForHospitalization)
	hospitalizations.PATCH("/:id/treatment-plans/:planId", perm(model.ResourceHospitalization, "edit"), h.treatmentPlan.UpdateTreatmentPlanInHospitalization)
	hospitalizations.DELETE("/:id/treatment-plans/:planId", perm(model.ResourceHospitalization, "delete"), h.treatmentPlan.DeleteTreatmentPlanInHospitalization)

	// Medical record core + addenda + examinations (BE9-2D ⑦: moved from handler.go
	// registerMedicalRecordRoutesWithAuth / RegisterMedicalRecordAddendumRoutes /
	// registerExaminationRoutesWithAuth; RBAC 逐語転記。billing-confirmation は billing 残留
	// domain のため旧側 /medical-records group に残置・gin path merge 共存)。
	records.GET("", perm(model.ResourceMedicalRecords, "view"), h.medicalRecord.ListMedicalRecords)
	records.GET("/:id", perm(model.ResourceMedicalRecords, "view"), h.medicalRecord.GetMedicalRecord)
	records.POST("", perm(model.ResourceMedicalRecords, "create"), h.medicalRecord.CreateMedicalRecord)
	records.PATCH("/:id", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecord.UpdateMedicalRecord)
	records.DELETE("/:id", perm(model.ResourceMedicalRecords, "delete"), h.medicalRecord.DeleteMedicalRecord)
	records.PATCH("/:id/recommendation-reason", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecord.UpdateMedicalRecordRecommendationReason)
	records.GET("/:id/addenda", perm(model.ResourceMedicalRecords, "view"), h.medicalRecordAddendum.ListMedicalRecordAddenda)
	records.POST("/:id/addenda", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecordAddendum.CreateMedicalRecordAddendum)
	examinations := rg.Group("/examinations")
	examinations.GET("", perm(model.ResourceExaminations, "view"), h.examination.ListExaminations)
	examinations.GET("/:id", perm(model.ResourceExaminations, "view"), h.examination.GetExamination)
	examinations.GET("/:id/print-snapshot", perm(model.ResourceExaminations, "view"), h.examination.GetExaminationPrintSnapshot)
	examinations.POST("",
		perm(model.ResourceExaminations, "create"),
		perm(model.ResourceExaminations, "edit"),
		h.examination.CreateExamination,
	)
	examinations.PATCH("/:id", perm(model.ResourceExaminations, "edit"), h.examination.UpdateExamination)
	examinations.POST("/:id/unconfirm", perm(model.ResourceExaminationUnconfirm, "edit"), h.examination.UnconfirmExamination)
	examinations.DELETE("/:id", perm(model.ResourceExaminations, "delete"), h.examination.DeleteExamination)
	examinations.GET("/:id/items", perm(model.ResourceExaminations, "view"), h.examination.ListExaminationItems)
	examinations.PUT("/:id/items", perm(model.ResourceExaminations, "edit"), h.examination.ReplaceExaminationItems)

	// TASK-374: versioned clinic-scoped checkup package import (default-deny permission).
	// preview = dry-run zero domain write; apply = explicit create action.
	checkupPackageImports := rg.Group("/checkup-package-imports")
	checkupPackageImports.POST("/preview", perm(model.ResourceCheckupPackageImport, "create"), h.checkupPackageImport.PreviewCheckupPackageImport)
	checkupPackageImports.POST("", perm(model.ResourceCheckupPackageImport, "create"), h.checkupPackageImport.ApplyCheckupPackageImport)

	// Lab import saga (BE9-2D sub-batch③: moved from internal/handler lab_import_handler.go
	// RegisterLabImportRoutes). All routes guard ResourceLabImport (preview/commit=create,
	// job/events reads=view) — P5 parity.
	labImports := rg.Group("/lab-imports")
	labImports.POST("/preview", perm(model.ResourceLabImport, "create"), h.labImport.PreviewLabImport)
	labImports.POST("", perm(model.ResourceLabImport, "create"), h.labImport.CommitLabImport)
	labImports.GET("/:job_id", perm(model.ResourceLabImport, "view"), h.labImport.GetLabImportJob)
	labImports.GET("/:job_id/events", perm(model.ResourceLabImport, "view"), h.labImport.ListLabImportEvents)
	// TASK-032: compensating revert is a dedicated endpoint under lab-import:edit (not examination unconfirm).
	labImports.POST("/:job_id/revert", perm(model.ResourceLabImport, "edit"), h.labImport.RevertLabImport)

	// Lab report read-only queries (BE9-2D sub-batch③: moved from lab_report_handler.go
	// RegisterLabReportRoutes). Both reads guard ResourceLabImport "view" — NOT a separate
	// lab-report resource — per Phase 4B.1 決定3 (the report views are scoped to the same
	// import permission, mirroring the lab-import read routes above).
	labReports := rg.Group("/lab-reports")
	labReports.GET("/jobs/:job_id/summaries", perm(model.ResourceLabImport, "view"), h.labReport.GetLabJobReportSummaries)
	labReports.GET("/exams/:exam_id", perm(model.ResourceLabImport, "view"), h.labReport.GetLabExamReport)
}
