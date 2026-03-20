// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/service"
)

// ---- Staff ----

// ListStaffs godoc
// FE互換: 直接配列を返す（ページネーション不要）
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}

	// NOTE: pagination パラメータは無視（全件返却）
	// 将来的にページネーション対応が必要な場合は、別エンドポイント化を検討
	staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, role, 1, 1000)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponseList(staffs))
}

// CreateStaff godoc
func (h *Handler) CreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staff, err := h.svc.Staff.CreateWithAccount(c.Request.Context(), &service.CreateStaffInput{
		ClinicID:      clinicID,
		Name:          req.Name,
		StaffRole:     req.StaffRole,
		Email:         req.Email,
		Password:      req.Password,
		LicenseNumber: req.LicenseNumber,
		JobTitleID:    req.JobTitleID,
		SortOrder:     req.SortOrder,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toStaffResponse(staff))
}

// UpdateStaff godoc
func (h *Handler) UpdateStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req updateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}

	staff, err := h.svc.Staff.Update(c.Request.Context(), clinicID, id, &service.UpdateStaffInput{
		Name:          req.Name,
		StaffRole:     req.StaffRole,
		LicenseNumber: req.LicenseNumber,
		JobTitleID:    req.JobTitleID,
		SortOrder:     req.SortOrder,
		IsActive:      req.IsActive,
	})
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponse(staff))
}

// GetStaff godoc
func (h *Handler) GetStaff(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	staff, err := h.svc.Staff.GetByID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStaffResponse(staff))
}

// DeleteStaff godoc
func (h *Handler) DeleteStaff(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	if err := h.svc.Staff.Delete(c.Request.Context(), clinicID, id); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReorderStaffs godoc
func (h *Handler) ReorderStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
		return
	}
	if err := h.svc.Staff.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// RegisterMasterRoutes はマスタ関連の全ルートを登録する
func (h *Handler) RegisterMasterRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")

	masters.GET("/animal-species", h.ListAnimalSpecies)
	masters.POST("/animal-species", h.CreateAnimalSpecies)
	masters.PATCH("/animal-species/reorder", h.ReorderAnimalSpecies) // 静的パスを /:id より前に登録
	masters.GET("/animal-species/:id", h.GetAnimalSpecies)
	masters.PATCH("/animal-species/:id", h.UpdateAnimalSpecies)
	masters.DELETE("/animal-species/:id", h.DeleteAnimalSpecies)

	masters.GET("/staffs", h.ListStaffs)
	masters.POST("/staffs", h.CreateStaff)
	masters.PATCH("/staffs/reorder", h.ReorderStaffs) // 静的パスを /:id より前に登録
	masters.GET("/staffs/:id", h.GetStaff)
	masters.PATCH("/staffs/:id", h.UpdateStaff)
	masters.DELETE("/staffs/:id", h.DeleteStaff)

	masters.GET("/cages", h.ListCages)
	masters.POST("/cages", h.CreateCage)
	masters.PATCH("/cages/reorder", h.ReorderCages) // 静的パスを /:id より前に登録
	masters.GET("/cages/:id", h.GetCage)
	masters.PATCH("/cages/:id", h.UpdateCage)
	masters.DELETE("/cages/:id", h.DeleteCage)

	masters.GET("/medicines", h.ListMedicines)
	masters.POST("/medicines", h.CreateMedicine)
	masters.PATCH("/medicines/reorder", h.ReorderMedicines) // 静的パスを /:id より前に登録
	masters.GET("/medicines/:id", h.GetMedicine)
	masters.PATCH("/medicines/:id", h.UpdateMedicine)
	masters.DELETE("/medicines/:id", h.DeleteMedicine)

	masters.GET("/vaccines", h.ListVaccines)
	masters.POST("/vaccines", h.CreateVaccine)
	masters.PATCH("/vaccines/reorder", h.ReorderVaccines) // 静的パスを /:id より前に登録
	masters.GET("/vaccines/:id", h.GetVaccine)
	masters.PATCH("/vaccines/:id", h.UpdateVaccine)
	masters.DELETE("/vaccines/:id", h.DeleteVaccine)

	masters.GET("/insurances", h.ListInsurances)
	masters.POST("/insurances", h.CreateInsurance)
	masters.PATCH("/insurances/reorder", h.ReorderInsurances) // 静的パスを /:id より前に登録
	masters.GET("/insurances/:id", h.GetInsurance)
	masters.PATCH("/insurances/:id", h.UpdateInsurance)
	masters.DELETE("/insurances/:id", h.DeleteInsurance)

	masters.GET("/service-types", h.ListServiceTypes)
	masters.POST("/service-types", h.CreateServiceType)
	masters.PATCH("/service-types/reorder", h.ReorderServiceTypes) // 静的パスを /:id より前に登録
	masters.GET("/service-types/:id", h.GetServiceType)
	masters.PATCH("/service-types/:id", h.UpdateServiceType)
	masters.DELETE("/service-types/:id", h.DeleteServiceType)

	masters.GET("/consultations", h.ListConsultations)
	masters.POST("/consultations", h.CreateConsultation)
	masters.PATCH("/consultations/reorder", h.ReorderConsultations) // 静的パスを /:id より前に登録
	masters.GET("/consultations/:id", h.GetConsultation)
	masters.PATCH("/consultations/:id", h.UpdateConsultation)
	masters.DELETE("/consultations/:id", h.DeleteConsultation)

	masters.GET("/procedures", h.ListProcedures)
	masters.POST("/procedures", h.CreateProcedure)
	masters.PATCH("/procedures/reorder", h.ReorderProcedures) // 静的パスを /:id より前に登録
	masters.GET("/procedures/:id", h.GetProcedure)
	masters.PATCH("/procedures/:id", h.UpdateProcedure)
	masters.DELETE("/procedures/:id", h.DeleteProcedure)

	masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", h.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", h.ReorderHospitalizationPlans) // 静的パスを /:id より前に登録
	masters.GET("/hospitalization-plans/:id", h.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", h.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", h.DeleteHospitalizationPlan)

	masters.GET("/trimming-courses", h.ListTrimmingCourses)
	masters.POST("/trimming-courses", h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/reorder", h.ReorderTrimmingCourses) // 静的パスを /:id より前に登録
	masters.GET("/trimming-courses/:id", h.GetTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", h.DeleteTrimmingCourse)

	masters.GET("/trimming-options", h.ListTrimmingOptions)
	masters.POST("/trimming-options", h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/reorder", h.ReorderTrimmingOptions) // 静的パスを /:id より前に登録
	masters.GET("/trimming-options/:id", h.GetTrimmingOption)
	masters.PATCH("/trimming-options/:id", h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", h.DeleteTrimmingOption)

	masters.GET("/examination-types", h.ListExaminationTypes)
	masters.POST("/examination-types", h.CreateExaminationType)
	masters.PATCH("/examination-types/reorder", h.ReorderExaminationTypes) // 静的パスを /:id より前に登録
	masters.GET("/examination-types/:id", h.GetExaminationType)
	masters.PATCH("/examination-types/:id", h.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", h.DeleteExaminationType)

	masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
	masters.POST("/diagnosis-categories", h.CreateDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/reorder", h.ReorderDiagnosisCategories) // 静的パスを /:id より前に登録
	masters.GET("/diagnosis-categories/:id", h.GetDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/:id", h.UpdateDiagnosisCategory)
	masters.DELETE("/diagnosis-categories/:id", h.DeleteDiagnosisCategory)

	masters.GET("/diagnosis-names", h.ListDiagnosisNames)
	masters.POST("/diagnosis-names", h.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", h.ReorderDiagnosisNames) // 静的パスを /:id より前に登録
	masters.GET("/diagnosis-names/:id", h.GetDiagnosisName)
	masters.PATCH("/diagnosis-names/:id", h.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", h.DeleteDiagnosisName)

	masters.GET("/checkup-types", h.ListCheckupTypes)
	masters.POST("/checkup-types", h.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", h.ReorderCheckupTypes) // 静的パスを /:id より前に登録
	masters.GET("/checkup-types/:id", h.GetCheckupType)
	masters.PATCH("/checkup-types/:id", h.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", h.DeleteCheckupType)

	masters.GET("/job-titles", h.ListJobTitles)
	masters.POST("/job-titles", h.CreateJobTitle)
	masters.PATCH("/job-titles/reorder", h.ReorderJobTitles) // 静的パスを /:id より前に登録
	masters.GET("/job-titles/:id", h.GetJobTitle)
	masters.PATCH("/job-titles/:id", h.UpdateJobTitle)
	masters.DELETE("/job-titles/:id", h.DeleteJobTitle)

	masters.GET("/chief-complaint-categories", h.ListChiefComplaints)
	masters.POST("/chief-complaint-categories", h.CreateChiefComplaint)
	masters.GET("/chief-complaint-categories/:id", h.GetChiefComplaint)
	masters.PATCH("/chief-complaint-categories/:id", h.UpdateChiefComplaint)
	masters.DELETE("/chief-complaint-categories/:id", h.DeleteChiefComplaint)

	masters.GET("/inquiry-templates", h.ListInquiryTemplates)
	masters.POST("/inquiry-templates", h.CreateInquiryTemplate)
	masters.GET("/inquiry-templates/:id", h.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", h.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", h.DeleteInquiryTemplate)

	masters.GET("/merchandise-items", h.ListMerchandiseItems)
	masters.POST("/merchandise-items", h.CreateMerchandiseItem)
	masters.POST("/merchandise-items/reorder", h.ReorderMerchandiseItems) // 静的パスを /:id より前に登録
	masters.GET("/merchandise-items/:id", h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", h.DeleteMerchandiseItem)
}
