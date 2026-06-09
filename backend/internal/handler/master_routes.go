package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterMasterRoutes はマスタ関連の全ルートを登録する（BUG-122: RBAC権限チェック適用）
func (h *Handler) RegisterMasterRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")

	// --- Permission middleware helpers (BUG-125: CRUD個別ガード) ---
	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.RequirePermission(string(resource), action)
	}

	// Animal Species
	masters.GET("/animal-species", perm(model.ResourceMasterAnimalSpecies, "view"), h.ListAnimalSpecies)
	masters.POST("/animal-species", perm(model.ResourceMasterAnimalSpecies, "create"), h.CreateAnimalSpecies)
	masters.PATCH("/animal-species/reorder", perm(model.ResourceMasterAnimalSpecies, "edit"), h.ReorderAnimalSpecies)
	masters.GET("/animal-species/:id", perm(model.ResourceMasterAnimalSpecies, "view"), h.GetAnimalSpecies)
	masters.PATCH("/animal-species/:id", perm(model.ResourceMasterAnimalSpecies, "edit"), h.UpdateAnimalSpecies)
	masters.DELETE("/animal-species/:id", perm(model.ResourceMasterAnimalSpecies, "delete"), h.DeleteAnimalSpecies)

	// Staffs
	masters.GET("/staffs", perm(model.ResourceMasterStaff, "view"), h.ListStaffs)
	masters.POST("/staffs", perm(model.ResourceMasterStaff, "create"), h.CreateStaff)
	masters.PATCH("/staffs/reorder", perm(model.ResourceMasterStaff, "edit"), h.ReorderStaffs)
	masters.GET("/staffs/:id", perm(model.ResourceMasterStaff, "view"), h.GetStaff)
	masters.PATCH("/staffs/:id", perm(model.ResourceMasterStaff, "edit"), h.UpdateStaff)
	masters.DELETE("/staffs/:id", perm(model.ResourceMasterStaff, "delete"), h.DeleteStaff)
	masters.GET("/staffs/:id/permission-groups", perm(model.ResourceMasterStaff, "view"), h.GetStaffPermissionGroups)
	masters.PUT("/staffs/:id/permission-groups", perm(model.ResourceMasterStaff, "edit"), h.SetStaffPermissionGroups)
	masters.GET("/staffs/:id/clinics", perm(model.ResourceMasterStaff, "view"), h.GetStaffClinicAssignments)
	masters.PUT("/staffs/:id/clinics", perm(model.ResourceMasterStaff, "edit"), h.SetStaffClinicAssignments)
	masters.GET("/staffs/:id/excluded-reservation-types", perm(model.ResourceMasterStaff, "view"), h.GetStaffExcludedReservationTypes)
	masters.PUT("/staffs/:id/excluded-reservation-types", perm(model.ResourceMasterStaff, "edit"), h.SetStaffExcludedReservationTypes)
	masters.GET("/staffs/:id/capable-reservation-types", perm(model.ResourceMasterStaff, "view"), h.GetStaffCapableReservationTypes)
	masters.PUT("/staffs/:id/capable-reservation-types", perm(model.ResourceMasterStaff, "edit"), h.SetStaffCapableReservationTypes)

	// Cages
	masters.GET("/cages", perm(model.ResourceMasterHospitalization, "view"), h.ListCages)
	masters.POST("/cages", perm(model.ResourceMasterHospitalization, "create"), h.CreateCage)
	masters.PATCH("/cages/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.ReorderCages)
	masters.GET("/cages/:id", perm(model.ResourceMasterHospitalization, "view"), h.GetCage)
	masters.PATCH("/cages/:id", perm(model.ResourceMasterHospitalization, "edit"), h.UpdateCage)
	masters.DELETE("/cages/:id", perm(model.ResourceMasterHospitalization, "delete"), h.DeleteCage)

	// Medicines
	masters.GET("/medicines", perm(model.ResourceMasterMedical, "view"), h.ListMedicines)
	masters.POST("/medicines", perm(model.ResourceMasterMedical, "create"), h.CreateMedicine)
	masters.PATCH("/medicines/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderMedicines)
	masters.GET("/medicines/:id", perm(model.ResourceMasterMedical, "view"), h.GetMedicine)
	masters.PATCH("/medicines/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateMedicine)
	masters.DELETE("/medicines/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteMedicine)

	// Vaccines
	masters.GET("/vaccines", perm(model.ResourceMasterMedical, "view"), h.ListVaccines)
	masters.POST("/vaccines", perm(model.ResourceMasterMedical, "create"), h.CreateVaccine)
	masters.PATCH("/vaccines/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderVaccines)
	masters.GET("/vaccines/:id", perm(model.ResourceMasterMedical, "view"), h.GetVaccine)
	masters.PATCH("/vaccines/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateVaccine)
	masters.DELETE("/vaccines/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteVaccine)

	// Insurances
	masters.GET("/insurances", perm(model.ResourceMasterInsurance, "view"), h.ListInsurances)
	masters.POST("/insurances", perm(model.ResourceMasterInsurance, "create"), h.CreateInsurance)
	masters.PATCH("/insurances/reorder", perm(model.ResourceMasterInsurance, "edit"), h.ReorderInsurances)
	masters.GET("/insurances/:id", perm(model.ResourceMasterInsurance, "view"), h.GetInsurance)
	masters.PATCH("/insurances/:id", perm(model.ResourceMasterInsurance, "edit"), h.UpdateInsurance)
	masters.DELETE("/insurances/:id", perm(model.ResourceMasterInsurance, "delete"), h.DeleteInsurance)

	// Reservation Category Groups
	masters.GET("/reservation-type-groups", perm(model.ResourceMasterReservationType, "view"), h.ListReservationTypeGroups)
	masters.POST("/reservation-type-groups", perm(model.ResourceMasterReservationType, "create"), h.CreateReservationTypeGroup)
	masters.PATCH("/reservation-type-groups/reorder", perm(model.ResourceMasterReservationType, "edit"), h.ReorderReservationTypeGroups)
	masters.GET("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "view"), h.GetReservationTypeGroup)
	masters.PATCH("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "edit"), h.UpdateReservationTypeGroup)
	masters.DELETE("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "delete"), h.DeleteReservationTypeGroup)

	// Reservation Types
	masters.GET("/reservation-types", perm(model.ResourceMasterReservationType, "view"), h.ListReservationTypes)
	masters.POST("/reservation-types", perm(model.ResourceMasterReservationType, "create"), h.CreateReservationType)
	masters.PATCH("/reservation-types/reorder", perm(model.ResourceMasterReservationType, "edit"), h.ReorderReservationTypes)
	masters.GET("/reservation-types/:id", perm(model.ResourceMasterReservationType, "view"), h.GetReservationType)
	masters.PATCH("/reservation-types/:id", perm(model.ResourceMasterReservationType, "edit"), h.UpdateReservationType)
	masters.DELETE("/reservation-types/:id", perm(model.ResourceMasterReservationType, "delete"), h.DeleteReservationType)
	// 予約不可時間
	masters.GET("/reservation-types/:id/unavailable-times", perm(model.ResourceMasterReservationType, "view"), h.ListUnavailableTimes)
	masters.POST("/reservation-types/:id/unavailable-times", perm(model.ResourceMasterReservationType, "edit"), h.CreateUnavailableTime)
	masters.DELETE("/reservation-types/:id/unavailable-times/:unavailable_time_id", perm(model.ResourceMasterReservationType, "delete"), h.DeleteUnavailableTime)
	// 予約可能開始時刻
	masters.GET("/reservation-types/:id/available-slots", perm(model.ResourceMasterReservationType, "view"), h.ListAvailableSlots)
	masters.POST("/reservation-types/:id/available-slots", perm(model.ResourceMasterReservationType, "edit"), h.CreateAvailableSlot)
	masters.DELETE("/reservation-types/:id/available-slots/:available_slot_id", perm(model.ResourceMasterReservationType, "delete"), h.DeleteAvailableSlot)
	// 職種紐付け
	masters.GET("/reservation-types/:id/occupations", perm(model.ResourceMasterReservationType, "view"), h.ListReservationTypeOccupations)
	masters.POST("/reservation-types/:id/occupations", perm(model.ResourceMasterReservationType, "edit"), h.LinkReservationTypeOccupation)
	masters.DELETE("/reservation-types/:id/occupations/:occupation_id", perm(model.ResourceMasterReservationType, "delete"), h.UnlinkReservationTypeOccupation)

	// Consultations
	masters.GET("/consultations", perm(model.ResourceMasterMedical, "view"), h.ListConsultations)
	masters.POST("/consultations", perm(model.ResourceMasterMedical, "create"), h.CreateConsultation)
	masters.PATCH("/consultations/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderConsultations)
	masters.GET("/consultations/:id", perm(model.ResourceMasterMedical, "view"), h.GetConsultation)
	masters.PATCH("/consultations/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateConsultation)
	masters.DELETE("/consultations/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteConsultation)

	// Procedures
	masters.GET("/procedures", perm(model.ResourceMasterMedical, "view"), h.ListProcedures)
	masters.POST("/procedures", perm(model.ResourceMasterMedical, "create"), h.CreateProcedure)
	masters.PATCH("/procedures/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderProcedures)
	masters.GET("/procedures/:id", perm(model.ResourceMasterMedical, "view"), h.GetProcedure)
	masters.PATCH("/procedures/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateProcedure)
	masters.DELETE("/procedures/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteProcedure)

	// Hospitalization Plans
	masters.GET("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "view"), h.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "create"), h.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.ReorderHospitalizationPlans)
	masters.GET("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "view"), h.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "edit"), h.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "delete"), h.DeleteHospitalizationPlan)

	// Trimming Courses
	masters.GET("/trimming-courses", perm(model.ResourceMasterTrimming, "view"), h.ListTrimmingCourses)
	masters.POST("/trimming-courses", perm(model.ResourceMasterTrimming, "create"), h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/reorder", perm(model.ResourceMasterTrimming, "edit"), h.ReorderTrimmingCourses)
	masters.GET("/trimming-courses/:id", perm(model.ResourceMasterTrimming, "view"), h.GetTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", perm(model.ResourceMasterTrimming, "edit"), h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", perm(model.ResourceMasterTrimming, "delete"), h.DeleteTrimmingCourse)

	// Trimming Options
	masters.GET("/trimming-options", perm(model.ResourceMasterTrimming, "view"), h.ListTrimmingOptions)
	masters.POST("/trimming-options", perm(model.ResourceMasterTrimming, "create"), h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/reorder", perm(model.ResourceMasterTrimming, "edit"), h.ReorderTrimmingOptions)
	masters.GET("/trimming-options/:id", perm(model.ResourceMasterTrimming, "view"), h.GetTrimmingOption)
	masters.PATCH("/trimming-options/:id", perm(model.ResourceMasterTrimming, "edit"), h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", perm(model.ResourceMasterTrimming, "delete"), h.DeleteTrimmingOption)

	// Trimming Course Types (#73)
	masters.GET("/trimming-course-types", perm(model.ResourceMasterTrimming, "view"), h.ListTrimmingCourseTypes)
	masters.POST("/trimming-course-types", perm(model.ResourceMasterTrimming, "create"), h.CreateTrimmingCourseType)
	masters.PATCH("/trimming-course-types/reorder", perm(model.ResourceMasterTrimming, "edit"), h.ReorderTrimmingCourseTypes)
	masters.GET("/trimming-course-types/:id", perm(model.ResourceMasterTrimming, "view"), h.GetTrimmingCourseType)
	masters.PATCH("/trimming-course-types/:id", perm(model.ResourceMasterTrimming, "edit"), h.UpdateTrimmingCourseType)
	masters.DELETE("/trimming-course-types/:id", perm(model.ResourceMasterTrimming, "delete"), h.DeleteTrimmingCourseType)

	// Campaigns (#81 割引キャンペーンマスタ。会計割引マスタのため ResourceAccounting 権限)
	masters.GET("/campaigns", perm(model.ResourceAccounting, "view"), h.ListCampaigns)
	masters.POST("/campaigns", perm(model.ResourceAccounting, "create"), h.CreateCampaign)
	masters.PATCH("/campaigns/reorder", perm(model.ResourceAccounting, "edit"), h.ReorderCampaigns)
	masters.GET("/campaigns/:id", perm(model.ResourceAccounting, "view"), h.GetCampaign)
	masters.PATCH("/campaigns/:id", perm(model.ResourceAccounting, "edit"), h.UpdateCampaign)
	masters.DELETE("/campaigns/:id", perm(model.ResourceAccounting, "delete"), h.DeleteCampaign)

	// Examination Types
	masters.GET("/examination-types", perm(model.ResourceMasterMedical, "view"), h.ListExaminationTypes)
	masters.POST("/examination-types", perm(model.ResourceMasterMedical, "create"), h.CreateExaminationType)
	masters.PATCH("/examination-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderExaminationTypes)
	masters.GET("/examination-types/:id", perm(model.ResourceMasterMedical, "view"), h.GetExaminationType)
	masters.PATCH("/examination-types/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteExaminationType)

	// Diagnosis Categories
	masters.GET("/diagnosis-types", perm(model.ResourceMasterMedical, "view"), h.ListDiagnosisTypes)
	masters.POST("/diagnosis-types", perm(model.ResourceMasterMedical, "create"), h.CreateDiagnosisType)
	masters.PATCH("/diagnosis-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderDiagnosisTypes)
	masters.GET("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "view"), h.GetDiagnosisType)
	masters.PATCH("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateDiagnosisType)
	masters.DELETE("/diagnosis-types/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteDiagnosisType)

	// Diagnosis Names
	masters.GET("/diagnosis-names", perm(model.ResourceMasterMedical, "view"), h.ListDiagnosisNames)
	masters.GET("/diagnosis-names/all", perm(model.ResourceMasterMedical, "view"), h.ListDiagnosisNamesAll)
	masters.POST("/diagnosis-names", perm(model.ResourceMasterMedical, "create"), h.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderDiagnosisNames)
	masters.GET("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "view"), h.GetDiagnosisName)
	masters.PATCH("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteDiagnosisName)

	// Checkup Types
	masters.GET("/checkup-types", perm(model.ResourceCheckups, "view"), h.ListCheckupTypes)
	masters.POST("/checkup-types", perm(model.ResourceCheckups, "create"), h.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", perm(model.ResourceCheckups, "edit"), h.ReorderCheckupTypes)
	masters.GET("/checkup-types/:id", perm(model.ResourceCheckups, "view"), h.GetCheckupType)
	masters.PATCH("/checkup-types/:id", perm(model.ResourceCheckups, "edit"), h.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", perm(model.ResourceCheckups, "delete"), h.DeleteCheckupType)

	// Occupations
	masters.GET("/occupations", perm(model.ResourceMasterStaff, "view"), h.ListOccupations)
	masters.POST("/occupations", perm(model.ResourceMasterStaff, "create"), h.CreateOccupation)
	masters.PATCH("/occupations/reorder", perm(model.ResourceMasterStaff, "edit"), h.ReorderOccupations)
	masters.GET("/occupations/:id", perm(model.ResourceMasterStaff, "view"), h.GetOccupation)
	masters.PATCH("/occupations/:id", perm(model.ResourceMasterStaff, "edit"), h.UpdateOccupation)
	masters.DELETE("/occupations/:id", perm(model.ResourceMasterStaff, "delete"), h.DeleteOccupation)

	h.RegisterPermissionGroupRoutes(masters)

	// Chief Complaint Categories
	masters.GET("/chief-complaint-types", perm(model.ResourceMasterMedical, "view"), h.ListChiefComplaints)
	masters.POST("/chief-complaint-types", perm(model.ResourceMasterMedical, "create"), h.CreateChiefComplaint)
	masters.PATCH("/chief-complaint-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderChiefComplaints)
	masters.GET("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "view"), h.GetChiefComplaint)
	masters.PATCH("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateChiefComplaint)
	masters.DELETE("/chief-complaint-types/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteChiefComplaint)

	// Inquiry Templates
	masters.GET("/inquiry-templates", perm(model.ResourceMasterMedical, "view"), h.ListInquiryTemplates)
	masters.POST("/inquiry-templates", perm(model.ResourceMasterMedical, "create"), h.CreateInquiryTemplate)
	masters.PATCH("/inquiry-templates/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderInquiryTemplates)
	masters.GET("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "view"), h.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteInquiryTemplate)

	// Merchandise Items
	masters.GET("/merchandise-items", perm(model.ResourceMasterMerchandise, "view"), h.ListMerchandiseItems)
	masters.POST("/merchandise-items", perm(model.ResourceMasterMerchandise, "create"), h.CreateMerchandiseItem)
	masters.PATCH("/merchandise-items/reorder", perm(model.ResourceMasterMerchandise, "edit"), h.ReorderMerchandiseItems)
	masters.GET("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "view"), h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "edit"), h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "delete"), h.DeleteMerchandiseItem)
}
