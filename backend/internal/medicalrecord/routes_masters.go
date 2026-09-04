package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

func (h *Handler) registerEarlyMasterRoutes(masters *gin.RouterGroup, perm medicalRoutePerm) {
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

	masters.GET("/vaccines", perm(model.ResourceMasterMedical, "view"), h.vaccine.ListVaccines)
	masters.POST("/vaccines", perm(model.ResourceMasterMedical, "create"), h.vaccine.CreateVaccine)
	masters.PATCH("/vaccines/reorder", perm(model.ResourceMasterMedical, "edit"), h.vaccine.ReorderVaccines)
	masters.GET("/vaccines/:id", perm(model.ResourceMasterMedical, "view"), h.vaccine.GetVaccine)
	masters.PATCH("/vaccines/:id", perm(model.ResourceMasterMedical, "edit"), h.vaccine.UpdateVaccine)
	masters.DELETE("/vaccines/:id", perm(model.ResourceMasterMedical, "delete"), h.vaccine.DeleteVaccine)

	// Checkup Types use ResourceCheckups (not ResourceMasterMedical). :id/fields is view.
	masters.GET("/checkup-types", perm(model.ResourceCheckups, "view"), h.checkupType.ListCheckupTypes)
	masters.POST("/checkup-types", perm(model.ResourceCheckups, "create"), h.checkupType.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", perm(model.ResourceCheckups, "edit"), h.checkupType.ReorderCheckupTypes)
	masters.GET("/checkup-types/:id", perm(model.ResourceCheckups, "view"), h.checkupType.GetCheckupType)
	masters.PATCH("/checkup-types/:id", perm(model.ResourceCheckups, "edit"), h.checkupType.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", perm(model.ResourceCheckups, "delete"), h.checkupType.DeleteCheckupType)
	masters.GET("/checkup-types/:id/fields", perm(model.ResourceCheckups, "view"), h.checkup.ListCheckupTypeFields)

	masters.GET("/inquiry-templates", perm(model.ResourceMasterMedical, "view"), h.inquiryTemplate.ListInquiryTemplates)
	masters.POST("/inquiry-templates", perm(model.ResourceMasterMedical, "create"), h.inquiryTemplate.CreateInquiryTemplate)
	masters.PATCH("/inquiry-templates/reorder", perm(model.ResourceMasterMedical, "edit"), h.inquiryTemplate.ReorderInquiryTemplates)
	masters.GET("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "view"), h.inquiryTemplate.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "edit"), h.inquiryTemplate.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "delete"), h.inquiryTemplate.DeleteInquiryTemplate)
}

func (h *Handler) registerLateMasterRoutes(masters *gin.RouterGroup, perm medicalRoutePerm) {
	// cages uses ResourceMasterHospitalization; other masters here use ResourceMasterMedical.
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

	masters.GET("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "view"), h.hospitalizationPlan.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "create"), h.hospitalizationPlan.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.hospitalizationPlan.ReorderHospitalizationPlans)
	masters.GET("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "view"), h.hospitalizationPlan.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "edit"), h.hospitalizationPlan.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "delete"), h.hospitalizationPlan.DeleteHospitalizationPlan)
}
