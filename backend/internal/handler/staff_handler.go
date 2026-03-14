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
func (h *Handler) ListStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var role *string
	if r := c.Query("role"); r != "" {
		role = &r
	}
	staffs, err := h.svc.Staff.List(c.Request.Context(), clinicID, role)
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

// RegisterMasterRoutes はマスタ関連の全ルートを登録する
func (h *Handler) RegisterMasterRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")

	masters.GET("/animal-species", h.ListAnimalSpecies)

	masters.GET("/staffs", h.ListStaffs)
	masters.POST("/staffs", h.CreateStaff)
	masters.PATCH("/staffs/:id", h.UpdateStaff)
	masters.DELETE("/staffs/:id", h.DeleteStaff)

	masters.GET("/cages", h.ListCages)
	masters.POST("/cages", h.CreateCage)
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
	masters.PATCH("/vaccines/:id", h.UpdateVaccine)
	masters.DELETE("/vaccines/:id", h.DeleteVaccine)

	masters.GET("/insurances", h.ListInsurances)
	masters.POST("/insurances", h.CreateInsurance)
	masters.PATCH("/insurances/:id", h.UpdateInsurance)
	masters.DELETE("/insurances/:id", h.DeleteInsurance)

	masters.GET("/service-types", h.ListServiceTypes)
	masters.POST("/service-types", h.CreateServiceType)
	masters.PATCH("/service-types/:id", h.UpdateServiceType)
	masters.DELETE("/service-types/:id", h.DeleteServiceType)

	masters.GET("/consultations", h.ListConsultations)
	masters.POST("/consultations", h.CreateConsultation)
	masters.PATCH("/consultations/:id", h.UpdateConsultation)
	masters.DELETE("/consultations/:id", h.DeleteConsultation)

	masters.GET("/procedures", h.ListProcedures)
	masters.POST("/procedures", h.CreateProcedure)
	masters.PATCH("/procedures/:id", h.UpdateProcedure)
	masters.DELETE("/procedures/:id", h.DeleteProcedure)

	masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", h.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", h.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", h.DeleteHospitalizationPlan)

	masters.GET("/trimming-courses", h.ListTrimmingCourses)
	masters.POST("/trimming-courses", h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", h.DeleteTrimmingCourse)

	masters.GET("/trimming-options", h.ListTrimmingOptions)
	masters.POST("/trimming-options", h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/:id", h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", h.DeleteTrimmingOption)

	masters.GET("/examination-types", h.ListExaminationTypes)
	masters.POST("/examination-types", h.CreateExaminationType)
	masters.PATCH("/examination-types/:id", h.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", h.DeleteExaminationType)

	masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
	masters.POST("/diagnosis-categories", h.CreateDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/reorder", h.ReorderDiagnosisCategories) // 静的パスを /:id より前に登録
	masters.PATCH("/diagnosis-categories/:id", h.UpdateDiagnosisCategory)
	masters.DELETE("/diagnosis-categories/:id", h.DeleteDiagnosisCategory)

	masters.GET("/diagnosis-names", h.ListDiagnosisNames)
	masters.POST("/diagnosis-names", h.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", h.ReorderDiagnosisNames) // 静的パスを /:id より前に登録
	masters.PATCH("/diagnosis-names/:id", h.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", h.DeleteDiagnosisName)

	masters.GET("/checkup-types", h.ListCheckupTypes)
	masters.POST("/checkup-types", h.CreateCheckupType)
	masters.PATCH("/checkup-types/:id", h.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", h.DeleteCheckupType)

	masters.GET("/job-titles", h.ListJobTitles)
	masters.POST("/job-titles", h.CreateJobTitle)
	masters.PATCH("/job-titles/:id", h.UpdateJobTitle)
	masters.DELETE("/job-titles/:id", h.DeleteJobTitle)

	masters.GET("/chief-complaints", h.ListChiefComplaints)
	masters.POST("/chief-complaints", h.CreateChiefComplaint)
	masters.PATCH("/chief-complaints/:id", h.UpdateChiefComplaint)
	masters.DELETE("/chief-complaints/:id", h.DeleteChiefComplaint)

	masters.GET("/inquiry-templates", h.ListInquiryTemplates)
	masters.POST("/inquiry-templates", h.CreateInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", h.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", h.DeleteInquiryTemplate)
}
