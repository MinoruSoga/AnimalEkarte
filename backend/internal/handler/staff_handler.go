// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
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

	// NOTE: pagination パラメータは無視（全件返却）
	// 将来的にページネーション対応が必要な場合は、別エンドポイント化を検討
	staffs, _, err := h.svc.Staff.List(c.Request.Context(), clinicID, 1, 1000)
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}

	staff, err := h.svc.Staff.Create(c.Request.Context(), &service.CreateStaffInput{
		ClinicID:      clinicID,
		Name:          req.Name,
		LicenseNumber: req.LicenseNumber,
		OccupationID:  req.OccupationID,
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
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	staff, err := h.svc.Staff.Update(c.Request.Context(), clinicID, id, &service.UpdateStaffInput{
		Name:          req.Name,
		LicenseNumber: req.LicenseNumber,
		OccupationID:  req.OccupationID,
		SortOrder:     req.SortOrder,
		IsActive:      req.IsActive,
	})
	if err != nil {
		RespondError(c, err)
		return
	}

	// パスワード変更（任意）: password フィールドが送信された場合のみ
	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) < 8 {
			RespondError(c, apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください"))
			return
		}
		if staff.AccountID != nil {
			account, accErr := h.repos.Account.GetByID(c.Request.Context(), *staff.AccountID)
			if accErr != nil {
				RespondError(c, accErr)
				return
			}
			hashed, hashErr := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if hashErr != nil {
				RespondError(c, apperrors.Wrap(hashErr, "failed to hash password"))
				return
			}
			account.PasswordHash = string(hashed)
			if updErr := h.repos.Account.Update(c.Request.Context(), account); updErr != nil {
				RespondError(c, updErr)
				return
			}
		}
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

// GetStaffPermissionGroups godoc
// GET /v1/masters/staffs/:id/permission-groups
func (h *Handler) GetStaffPermissionGroups(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	groupIDs, err := h.repos.PermissionGroup.GetGroupIDsByStaffID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"group_ids": groupIDs})
}

// SetStaffPermissionGroups godoc
// PUT /v1/masters/staffs/:id/permission-groups
func (h *Handler) SetStaffPermissionGroups(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req struct {
		GroupIDs []uint64 `json:"group_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.GroupIDs == nil {
		req.GroupIDs = []uint64{}
	}
	if err := h.repos.PermissionGroup.SetStaffGroups(c.Request.Context(), id, req.GroupIDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"group_ids": req.GroupIDs})
}

// GetStaffClinicAssignments godoc
// GET /v1/masters/staffs/:id/clinics
func (h *Handler) GetStaffClinicAssignments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	assignments, err := h.svc.StaffClinicAssignment.FindByStaffID(c.Request.Context(), id)
	if err != nil {
		RespondError(c, err)
		return
	}
	clinicIDs := make([]uint64, 0, len(assignments))
	for _, a := range assignments {
		clinicIDs = append(clinicIDs, a.ClinicID)
	}
	c.JSON(http.StatusOK, gin.H{"clinic_ids": clinicIDs})
}

// SetStaffClinicAssignments godoc
// PUT /v1/masters/staffs/:id/clinics
func (h *Handler) SetStaffClinicAssignments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, apperrors.WrapInvalidInput("invalid id"))
		return
	}
	var req struct {
		ClinicIDs []uint64 `json:"clinic_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if req.ClinicIDs == nil {
		req.ClinicIDs = []uint64{}
	}

	ctx := c.Request.Context()
	// 既存の割当を全削除
	if err := h.repos.StaffClinicAssignment.DeleteByStaffID(ctx, id); err != nil {
		RespondError(c, err)
		return
	}
	// 新しい割当を作成（最初の1件を is_main=true とする）
	for i, clinicID := range req.ClinicIDs {
		assignment := &model.StaffClinicAssignment{
			StaffID:  id,
			ClinicID: clinicID,
			IsMain:   i == 0,
		}
		if err := h.repos.StaffClinicAssignment.Create(ctx, assignment); err != nil {
			RespondError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"clinic_ids": req.ClinicIDs})
}

// ReorderStaffs godoc
func (h *Handler) ReorderStaffs(c *gin.Context) {
	clinicID, ok := extractClinicID(c)
	if !ok {
		return
	}
	var req reorderStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}
	if err := h.svc.Staff.Reorder(c.Request.Context(), clinicID, req.IDs); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reordered"})
}

// RegisterMasterRoutes はマスタ関連の全ルートを登録する（BUG-122: RBAC権限チェック適用）
func (h *Handler) RegisterMasterRoutes(rg *gin.RouterGroup) {
	masters := rg.Group("/masters")

	// --- Permission middleware definitions ---
	permAnimalSpecies := h.RequirePermission(string(model.ResourceMasterAnimalSpecies), "edit")
	permStaff := h.RequirePermission(string(model.ResourceMasterStaff), "edit")
	permHosp := h.RequirePermission(string(model.ResourceMasterHospitalization), "edit")
	permMedical := h.RequirePermission(string(model.ResourceMasterMedical), "edit")
	permInsurance := h.RequirePermission(string(model.ResourceMasterInsurance), "edit")
	permServiceType := h.RequirePermission(string(model.ResourceMasterServiceType), "edit")
	permTrimming := h.RequirePermission(string(model.ResourceMasterTrimming), "edit")
	permCheckups := h.RequirePermission(string(model.ResourceCheckups), "edit")
	permMerchandise := h.RequirePermission(string(model.ResourceMasterMerchandise), "edit")

	// Animal Species
	masters.GET("/animal-species", h.ListAnimalSpecies)
	masters.POST("/animal-species", permAnimalSpecies, h.CreateAnimalSpecies)
	masters.PATCH("/animal-species/reorder", permAnimalSpecies, h.ReorderAnimalSpecies)
	masters.GET("/animal-species/:id", h.GetAnimalSpecies)
	masters.PATCH("/animal-species/:id", permAnimalSpecies, h.UpdateAnimalSpecies)
	masters.DELETE("/animal-species/:id", permAnimalSpecies, h.DeleteAnimalSpecies)

	// Staffs
	masters.GET("/staffs", h.ListStaffs)
	masters.POST("/staffs", permStaff, h.CreateStaff)
	masters.PATCH("/staffs/reorder", permStaff, h.ReorderStaffs)
	masters.GET("/staffs/:id", h.GetStaff)
	masters.PATCH("/staffs/:id", permStaff, h.UpdateStaff)
	masters.DELETE("/staffs/:id", permStaff, h.DeleteStaff)
	masters.GET("/staffs/:id/permission-groups", h.GetStaffPermissionGroups)
	masters.PUT("/staffs/:id/permission-groups", permStaff, h.SetStaffPermissionGroups)
	masters.GET("/staffs/:id/clinics", h.GetStaffClinicAssignments)
	masters.PUT("/staffs/:id/clinics", permStaff, h.SetStaffClinicAssignments)

	// Cages
	masters.GET("/cages", h.ListCages)
	masters.POST("/cages", permHosp, h.CreateCage)
	masters.PATCH("/cages/reorder", permHosp, h.ReorderCages)
	masters.GET("/cages/:id", h.GetCage)
	masters.PATCH("/cages/:id", permHosp, h.UpdateCage)
	masters.DELETE("/cages/:id", permHosp, h.DeleteCage)

	// Medicines
	masters.GET("/medicines", h.ListMedicines)
	masters.POST("/medicines", permMedical, h.CreateMedicine)
	masters.PATCH("/medicines/reorder", permMedical, h.ReorderMedicines)
	masters.GET("/medicines/:id", h.GetMedicine)
	masters.PATCH("/medicines/:id", permMedical, h.UpdateMedicine)
	masters.DELETE("/medicines/:id", permMedical, h.DeleteMedicine)

	// Vaccines
	masters.GET("/vaccines", h.ListVaccines)
	masters.POST("/vaccines", permMedical, h.CreateVaccine)
	masters.PATCH("/vaccines/reorder", permMedical, h.ReorderVaccines)
	masters.GET("/vaccines/:id", h.GetVaccine)
	masters.PATCH("/vaccines/:id", permMedical, h.UpdateVaccine)
	masters.DELETE("/vaccines/:id", permMedical, h.DeleteVaccine)

	// Insurances
	masters.GET("/insurances", h.ListInsurances)
	masters.POST("/insurances", permInsurance, h.CreateInsurance)
	masters.PATCH("/insurances/reorder", permInsurance, h.ReorderInsurances)
	masters.GET("/insurances/:id", h.GetInsurance)
	masters.PATCH("/insurances/:id", permInsurance, h.UpdateInsurance)
	masters.DELETE("/insurances/:id", permInsurance, h.DeleteInsurance)

	// Service Types
	masters.GET("/service-types", h.ListServiceTypes)
	masters.POST("/service-types", permServiceType, h.CreateServiceType)
	masters.PATCH("/service-types/reorder", permServiceType, h.ReorderServiceTypes)
	masters.GET("/service-types/:id", h.GetServiceType)
	masters.PATCH("/service-types/:id", permServiceType, h.UpdateServiceType)
	masters.DELETE("/service-types/:id", permServiceType, h.DeleteServiceType)

	// Consultations
	masters.GET("/consultations", h.ListConsultations)
	masters.POST("/consultations", permMedical, h.CreateConsultation)
	masters.PATCH("/consultations/reorder", permMedical, h.ReorderConsultations)
	masters.GET("/consultations/:id", h.GetConsultation)
	masters.PATCH("/consultations/:id", permMedical, h.UpdateConsultation)
	masters.DELETE("/consultations/:id", permMedical, h.DeleteConsultation)

	// Procedures
	masters.GET("/procedures", h.ListProcedures)
	masters.POST("/procedures", permMedical, h.CreateProcedure)
	masters.PATCH("/procedures/reorder", permMedical, h.ReorderProcedures)
	masters.GET("/procedures/:id", h.GetProcedure)
	masters.PATCH("/procedures/:id", permMedical, h.UpdateProcedure)
	masters.DELETE("/procedures/:id", permMedical, h.DeleteProcedure)

	// Hospitalization Plans
	masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", permHosp, h.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", permHosp, h.ReorderHospitalizationPlans)
	masters.GET("/hospitalization-plans/:id", h.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", permHosp, h.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", permHosp, h.DeleteHospitalizationPlan)

	// Trimming Courses
	masters.GET("/trimming-courses", h.ListTrimmingCourses)
	masters.POST("/trimming-courses", permTrimming, h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/reorder", permTrimming, h.ReorderTrimmingCourses)
	masters.GET("/trimming-courses/:id", h.GetTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", permTrimming, h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", permTrimming, h.DeleteTrimmingCourse)

	// Trimming Options
	masters.GET("/trimming-options", h.ListTrimmingOptions)
	masters.POST("/trimming-options", permTrimming, h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/reorder", permTrimming, h.ReorderTrimmingOptions)
	masters.GET("/trimming-options/:id", h.GetTrimmingOption)
	masters.PATCH("/trimming-options/:id", permTrimming, h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", permTrimming, h.DeleteTrimmingOption)

	// Examination Types
	masters.GET("/examination-types", h.ListExaminationTypes)
	masters.POST("/examination-types", permMedical, h.CreateExaminationType)
	masters.PATCH("/examination-types/reorder", permMedical, h.ReorderExaminationTypes)
	masters.GET("/examination-types/:id", h.GetExaminationType)
	masters.PATCH("/examination-types/:id", permMedical, h.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", permMedical, h.DeleteExaminationType)

	// Diagnosis Categories
	masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
	masters.POST("/diagnosis-categories", permMedical, h.CreateDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/reorder", permMedical, h.ReorderDiagnosisCategories)
	masters.GET("/diagnosis-categories/:id", h.GetDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/:id", permMedical, h.UpdateDiagnosisCategory)
	masters.DELETE("/diagnosis-categories/:id", permMedical, h.DeleteDiagnosisCategory)

	// Diagnosis Names
	masters.GET("/diagnosis-names", h.ListDiagnosisNames)
	masters.POST("/diagnosis-names", permMedical, h.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", permMedical, h.ReorderDiagnosisNames)
	masters.GET("/diagnosis-names/:id", h.GetDiagnosisName)
	masters.PATCH("/diagnosis-names/:id", permMedical, h.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", permMedical, h.DeleteDiagnosisName)

	// Checkup Types
	masters.GET("/checkup-types", h.ListCheckupTypes)
	masters.POST("/checkup-types", permCheckups, h.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", permCheckups, h.ReorderCheckupTypes)
	masters.GET("/checkup-types/:id", h.GetCheckupType)
	masters.PATCH("/checkup-types/:id", permCheckups, h.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", permCheckups, h.DeleteCheckupType)

	// Occupations
	masters.GET("/occupations", h.ListOccupations)
	masters.POST("/occupations", permStaff, h.CreateOccupation)
	masters.PATCH("/occupations/reorder", permStaff, h.ReorderOccupations)
	masters.GET("/occupations/:id", h.GetOccupation)
	masters.PATCH("/occupations/:id", permStaff, h.UpdateOccupation)
	masters.DELETE("/occupations/:id", permStaff, h.DeleteOccupation)

	// Permission Groups (既に権限チェック済み — 唯一の正しい実装)
	masters.GET("/permission-groups", h.ListPermissionGroups)
	masters.GET("/permission-groups/:id", h.GetPermissionGroup)
	pgWrite := masters.Group("/permission-groups")
	pgWrite.Use(h.RequirePermission(string(model.ResourceMasterPermission), "edit"))
	pgWrite.POST("", h.CreatePermissionGroup)
	pgWrite.PATCH("", h.ReorderPermissionGroups)
	pgWrite.PATCH("/:id", h.UpdatePermissionGroup)
	pgWrite.DELETE("/:id", h.DeletePermissionGroup)
	pgWrite.PUT("/:id/rules", h.SetPermissionGroupRules)

	// Chief Complaint Categories
	masters.GET("/chief-complaint-categories", h.ListChiefComplaints)
	masters.POST("/chief-complaint-categories", permMedical, h.CreateChiefComplaint)
	masters.GET("/chief-complaint-categories/:id", h.GetChiefComplaint)
	masters.PATCH("/chief-complaint-categories/:id", permMedical, h.UpdateChiefComplaint)
	masters.DELETE("/chief-complaint-categories/:id", permMedical, h.DeleteChiefComplaint)

	// Inquiry Templates
	masters.GET("/inquiry-templates", h.ListInquiryTemplates)
	masters.POST("/inquiry-templates", permMedical, h.CreateInquiryTemplate)
	masters.GET("/inquiry-templates/:id", h.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", permMedical, h.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", permMedical, h.DeleteInquiryTemplate)

	// Merchandise Items
	masters.GET("/merchandise-items", h.ListMerchandiseItems)
	masters.POST("/merchandise-items", permMerchandise, h.CreateMerchandiseItem)
	masters.POST("/merchandise-items/reorder", permMerchandise, h.ReorderMerchandiseItems)
	masters.GET("/merchandise-items/:id", h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", permMerchandise, h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", permMerchandise, h.DeleteMerchandiseItem)
}
