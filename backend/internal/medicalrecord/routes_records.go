package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

func (h *Handler) registerVaccinationAndCheckupRoutes(rg *gin.RouterGroup, perm medicalRoutePerm) {
	vaccinations := rg.Group("/vaccinations")
	vaccinations.GET("", perm(model.ResourceVaccinations, "view"), h.vaccination.ListVaccinations)
	vaccinations.GET("/:id", perm(model.ResourceVaccinations, "view"), h.vaccination.GetVaccination)
	vaccinations.POST("", perm(model.ResourceVaccinations, "create"), h.vaccination.CreateVaccination)
	vaccinations.PATCH("/:id", perm(model.ResourceVaccinations, "edit"), h.vaccination.UpdateVaccination)
	vaccinations.DELETE("/:id", perm(model.ResourceVaccinations, "delete"), h.vaccination.DeleteVaccination)

	checkups := rg.Group("/checkups")
	checkups.GET("", perm(model.ResourceCheckups, "view"), h.checkup.ListGlobalCheckups)
	checkups.GET("/field-results", perm(model.ResourceCheckups, "view"), h.checkup.ListPetCheckupResults)
}

func (h *Handler) registerMedicalRecordChildRoutes(records *gin.RouterGroup, perm medicalRoutePerm) {
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

	// vitals POST is "edit" (not "create"); images POST / upload are "create".
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

	records.GET("/:id/treatments", perm(model.ResourceMedicalRecords, "view"), h.treatment.ListTreatments)
	records.POST("/:id/treatments", perm(model.ResourceMedicalRecords, "create"), h.treatment.CreateTreatment)
	records.PATCH("/:id/treatments/:treatmentId", perm(model.ResourceMedicalRecords, "edit"), h.treatment.UpdateTreatment)
	records.DELETE("/:id/treatments/:treatmentId", perm(model.ResourceMedicalRecords, "delete"), h.treatment.DeleteTreatment)
	records.PUT("/:id/treatments", perm(model.ResourceMedicalRecords, "edit"), h.treatment.BulkUpdateTreatments)
}

func (h *Handler) registerPetTreatmentHistoryRoute(rg *gin.RouterGroup, perm medicalRoutePerm) {
	pets := rg.Group("/pets")
	pets.GET("/:id/treatment-history", perm(model.ResourceMedicalRecords, "view"), h.treatment.ListPetTreatmentHistory)
}

func (h *Handler) registerMedicalRecordCoreRoutes(records *gin.RouterGroup, perm medicalRoutePerm) {
	records.GET("", perm(model.ResourceMedicalRecords, "view"), h.medicalRecord.ListMedicalRecords)
	records.GET("/:id", perm(model.ResourceMedicalRecords, "view"), h.medicalRecord.GetMedicalRecord)
	records.POST("", perm(model.ResourceMedicalRecords, "create"), h.medicalRecord.CreateMedicalRecord)
	records.PATCH("/:id", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecord.UpdateMedicalRecord)
	records.DELETE("/:id", perm(model.ResourceMedicalRecords, "delete"), h.medicalRecord.DeleteMedicalRecord)
	records.PATCH("/:id/recommendation-reason", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecord.UpdateMedicalRecordRecommendationReason)
	records.GET("/:id/addenda", perm(model.ResourceMedicalRecords, "view"), h.medicalRecordAddendum.ListMedicalRecordAddenda)
	records.POST("/:id/addenda", perm(model.ResourceMedicalRecords, "edit"), h.medicalRecordAddendum.CreateMedicalRecordAddendum)
}

func (h *Handler) registerExaminationAndPackageRoutes(rg *gin.RouterGroup, perm medicalRoutePerm) {
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

	checkupPackageImports := rg.Group("/checkup-package-imports")
	checkupPackageImports.POST("/preview", perm(model.ResourceCheckupPackageImport, "create"), h.checkupPackageImport.PreviewCheckupPackageImport)
	checkupPackageImports.POST("", perm(model.ResourceCheckupPackageImport, "create"), h.checkupPackageImport.ApplyCheckupPackageImport)
}
