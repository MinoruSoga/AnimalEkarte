package medicalrecord

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

func (h *Handler) registerHospitalizationRoutes(hospitalizations *gin.RouterGroup, perm medicalRoutePerm) {
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
}

func (h *Handler) registerTreatmentPlanRoutes(records, hospitalizations *gin.RouterGroup, perm medicalRoutePerm) {
	records.GET("/:id/treatment-plans", perm(model.ResourceMedicalRecords, "view"), h.treatmentPlan.ListTreatmentPlansByMedicalRecord)
	records.POST("/:id/treatment-plans", perm(model.ResourceMedicalRecords, "create"), h.treatmentPlan.CreateTreatmentPlanForMedicalRecord)
	records.PATCH("/:id/treatment-plans/:planId", perm(model.ResourceMedicalRecords, "edit"), h.treatmentPlan.UpdateTreatmentPlanInMedicalRecord)
	records.DELETE("/:id/treatment-plans/:planId", perm(model.ResourceMedicalRecords, "delete"), h.treatmentPlan.DeleteTreatmentPlanInMedicalRecord)
	hospitalizations.GET("/:id/treatment-plans", perm(model.ResourceHospitalization, "view"), h.treatmentPlan.ListTreatmentPlansByHospitalization)
	hospitalizations.POST("/:id/treatment-plans", perm(model.ResourceHospitalization, "create"), h.treatmentPlan.CreateTreatmentPlanForHospitalization)
	hospitalizations.PATCH("/:id/treatment-plans/:planId", perm(model.ResourceHospitalization, "edit"), h.treatmentPlan.UpdateTreatmentPlanInHospitalization)
	hospitalizations.DELETE("/:id/treatment-plans/:planId", perm(model.ResourceHospitalization, "delete"), h.treatmentPlan.DeleteTreatmentPlanInHospitalization)
}
