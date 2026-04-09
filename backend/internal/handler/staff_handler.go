// Package handler provides HTTP handler implementations for Staff entity.
package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

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

	ctx := c.Request.Context()

	// BUG-145: email が指定されている場合は重複チェックを行い、Account を作成してスタッフに紐づける。
	// Account 作成・bcrypt ハッシュ化はすべて StaffService.CreateWithAccount に委譲する。
	email := strings.TrimSpace(req.Email)
	var staff *model.Staff
	var err error

	if email != "" {
		if req.Password == "" {
			RespondError(c, apperrors.WrapInvalidInput("password is required when email is provided"))
			return
		}
		if err := validatePassword(req.Password); err != nil {
			RespondError(c, err)
			return
		}
		reservationVisible := true
		if req.ReservationVisible != nil {
			reservationVisible = *req.ReservationVisible
		}
		staff, err = h.svc.Staff.CreateWithAccount(ctx, &service.CreateStaffWithAccountInput{
			ClinicID:           clinicID,
			Name:               req.Name,
			LicenseNumber:      req.LicenseNumber,
			OccupationID:       req.OccupationID,
			SortOrder:          req.SortOrder,
			Email:              email,
			Password:           req.Password,
			StaffType:          req.StaffType,
			ReservationVisible: reservationVisible,
			ReservationComment: req.ReservationComment,
		})
	} else {
		reservationVisible := true
		if req.ReservationVisible != nil {
			reservationVisible = *req.ReservationVisible
		}
		staff, err = h.svc.Staff.Create(ctx, &service.CreateStaffInput{
			ClinicID:           clinicID,
			Name:               req.Name,
			LicenseNumber:      req.LicenseNumber,
			OccupationID:       req.OccupationID,
			SortOrder:          req.SortOrder,
			StaffType:          req.StaffType,
			ReservationVisible: reservationVisible,
			ReservationComment: req.ReservationComment,
		})
	}
	if err != nil {
		RespondError(c, err)
		return
	}

	// BUG-153: 現在のクリニックに自動割り当て
	if asgErr := h.svc.StaffClinicAssignment.Create(ctx, &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   true,
	}); asgErr != nil {
		RespondError(c, apperrors.Wrap(asgErr, "failed to assign staff to clinic"))
		return
	}

	// Account を Preload して email を返すため再取得する
	if reloaded, reloadErr := h.repos.Staff.FindByID(ctx, staff.ID); reloadErr == nil {
		staff = reloaded
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

	// パスワードのみの更新かどうかを判定（BUG-131）
	hasProfileUpdate := req.Name != nil || req.LicenseNumber != nil || req.OccupationID != nil || req.SortOrder != nil || req.IsActive != nil ||
		req.StaffType != nil || req.ReservationVisible != nil || req.ReservationComment != nil
	hasPasswordUpdate := req.Password != nil && *req.Password != ""

	if !hasProfileUpdate && !hasPasswordUpdate {
		RespondError(c, apperrors.WrapInvalidInput("at least one field must be provided"))
		return
	}

	var staff *model.Staff
	if hasProfileUpdate {
		var svcErr error
		staff, svcErr = h.svc.Staff.Update(c.Request.Context(), clinicID, id, &service.UpdateStaffInput{
			Name:               req.Name,
			LicenseNumber:      req.LicenseNumber,
			OccupationID:       req.OccupationID,
			SortOrder:          req.SortOrder,
			IsActive:           req.IsActive,
			StaffType:          req.StaffType,
			ReservationVisible: req.ReservationVisible,
			ReservationComment: req.ReservationComment,
		})
		if svcErr != nil {
			RespondError(c, svcErr)
			return
		}
	} else {
		// パスワードのみ更新の場合、既存スタッフを取得
		var findErr error
		staff, findErr = h.svc.Staff.GetByID(c.Request.Context(), id)
		if findErr != nil {
			RespondError(c, findErr)
			return
		}
	}

	// パスワード変更（任意）: password フィールドが送信された場合のみ
	// bcrypt ハッシュ化・Account 更新は StaffService.UpdatePassword に委譲する。
	if hasPasswordUpdate {
		if err := validatePassword(*req.Password); err != nil {
			RespondError(c, err)
			return
		}
		if staff.AccountID != nil {
			if updErr := h.svc.Staff.UpdatePassword(c.Request.Context(), *staff.AccountID, *req.Password); updErr != nil {
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

	// 削除→作成をサービス層のトランザクションで実行する
	if err := h.svc.Staff.SetClinicAssignments(c.Request.Context(), id, req.ClinicIDs); err != nil {
		RespondError(c, err)
		return
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

	// --- Permission middleware helpers (BUG-125: CRUD個別ガード) ---
	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.RequirePermission(string(resource), action)
	}

	// Animal Species
	masters.GET("/animal-species", h.ListAnimalSpecies)
	masters.POST("/animal-species", perm(model.ResourceMasterAnimalSpecies, "create"), h.CreateAnimalSpecies)
	masters.PATCH("/animal-species/reorder", perm(model.ResourceMasterAnimalSpecies, "edit"), h.ReorderAnimalSpecies)
	masters.GET("/animal-species/:id", h.GetAnimalSpecies)
	masters.PATCH("/animal-species/:id", perm(model.ResourceMasterAnimalSpecies, "edit"), h.UpdateAnimalSpecies)
	masters.DELETE("/animal-species/:id", perm(model.ResourceMasterAnimalSpecies, "delete"), h.DeleteAnimalSpecies)

	// Staffs
	masters.GET("/staffs", h.ListStaffs)
	masters.POST("/staffs", perm(model.ResourceMasterStaff, "create"), h.CreateStaff)
	masters.PATCH("/staffs/reorder", perm(model.ResourceMasterStaff, "edit"), h.ReorderStaffs)
	masters.GET("/staffs/:id", h.GetStaff)
	masters.PATCH("/staffs/:id", perm(model.ResourceMasterStaff, "edit"), h.UpdateStaff)
	masters.DELETE("/staffs/:id", perm(model.ResourceMasterStaff, "delete"), h.DeleteStaff)
	masters.GET("/staffs/:id/permission-groups", h.GetStaffPermissionGroups)
	masters.PUT("/staffs/:id/permission-groups", perm(model.ResourceMasterStaff, "edit"), h.SetStaffPermissionGroups)
	masters.GET("/staffs/:id/clinics", h.GetStaffClinicAssignments)
	masters.PUT("/staffs/:id/clinics", perm(model.ResourceMasterStaff, "edit"), h.SetStaffClinicAssignments)

	// Cages
	masters.GET("/cages", h.ListCages)
	masters.POST("/cages", perm(model.ResourceMasterHospitalization, "create"), h.CreateCage)
	masters.PATCH("/cages/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.ReorderCages)
	masters.GET("/cages/:id", h.GetCage)
	masters.PATCH("/cages/:id", perm(model.ResourceMasterHospitalization, "edit"), h.UpdateCage)
	masters.DELETE("/cages/:id", perm(model.ResourceMasterHospitalization, "delete"), h.DeleteCage)

	// Medicines
	masters.GET("/medicines", h.ListMedicines)
	masters.POST("/medicines", perm(model.ResourceMasterMedical, "create"), h.CreateMedicine)
	masters.PATCH("/medicines/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderMedicines)
	masters.GET("/medicines/:id", h.GetMedicine)
	masters.PATCH("/medicines/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateMedicine)
	masters.DELETE("/medicines/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteMedicine)

	// Vaccines
	masters.GET("/vaccines", h.ListVaccines)
	masters.POST("/vaccines", perm(model.ResourceMasterMedical, "create"), h.CreateVaccine)
	masters.PATCH("/vaccines/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderVaccines)
	masters.GET("/vaccines/:id", h.GetVaccine)
	masters.PATCH("/vaccines/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateVaccine)
	masters.DELETE("/vaccines/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteVaccine)

	// Insurances
	masters.GET("/insurances", h.ListInsurances)
	masters.POST("/insurances", perm(model.ResourceMasterInsurance, "create"), h.CreateInsurance)
	masters.PATCH("/insurances/reorder", perm(model.ResourceMasterInsurance, "edit"), h.ReorderInsurances)
	masters.GET("/insurances/:id", h.GetInsurance)
	masters.PATCH("/insurances/:id", perm(model.ResourceMasterInsurance, "edit"), h.UpdateInsurance)
	masters.DELETE("/insurances/:id", perm(model.ResourceMasterInsurance, "delete"), h.DeleteInsurance)

	// Service Types
	masters.GET("/service-types", h.ListServiceTypes)
	masters.POST("/service-types", perm(model.ResourceMasterServiceType, "create"), h.CreateServiceType)
	masters.PATCH("/service-types/reorder", perm(model.ResourceMasterServiceType, "edit"), h.ReorderServiceTypes)
	masters.GET("/service-types/:id", h.GetServiceType)
	masters.PATCH("/service-types/:id", perm(model.ResourceMasterServiceType, "edit"), h.UpdateServiceType)
	masters.DELETE("/service-types/:id", perm(model.ResourceMasterServiceType, "delete"), h.DeleteServiceType)

	// Consultations
	masters.GET("/consultations", h.ListConsultations)
	masters.POST("/consultations", perm(model.ResourceMasterMedical, "create"), h.CreateConsultation)
	masters.PATCH("/consultations/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderConsultations)
	masters.GET("/consultations/:id", h.GetConsultation)
	masters.PATCH("/consultations/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateConsultation)
	masters.DELETE("/consultations/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteConsultation)

	// Procedures
	masters.GET("/procedures", h.ListProcedures)
	masters.POST("/procedures", perm(model.ResourceMasterMedical, "create"), h.CreateProcedure)
	masters.PATCH("/procedures/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderProcedures)
	masters.GET("/procedures/:id", h.GetProcedure)
	masters.PATCH("/procedures/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateProcedure)
	masters.DELETE("/procedures/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteProcedure)

	// Hospitalization Plans
	masters.GET("/hospitalization-plans", h.ListHospitalizationPlans)
	masters.POST("/hospitalization-plans", perm(model.ResourceMasterHospitalization, "create"), h.CreateHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/reorder", perm(model.ResourceMasterHospitalization, "edit"), h.ReorderHospitalizationPlans)
	masters.GET("/hospitalization-plans/:id", h.GetHospitalizationPlan)
	masters.PATCH("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "edit"), h.UpdateHospitalizationPlan)
	masters.DELETE("/hospitalization-plans/:id", perm(model.ResourceMasterHospitalization, "delete"), h.DeleteHospitalizationPlan)

	// Trimming Courses
	masters.GET("/trimming-courses", h.ListTrimmingCourses)
	masters.POST("/trimming-courses", perm(model.ResourceMasterTrimming, "create"), h.CreateTrimmingCourse)
	masters.PATCH("/trimming-courses/reorder", perm(model.ResourceMasterTrimming, "edit"), h.ReorderTrimmingCourses)
	masters.GET("/trimming-courses/:id", h.GetTrimmingCourse)
	masters.PATCH("/trimming-courses/:id", perm(model.ResourceMasterTrimming, "edit"), h.UpdateTrimmingCourse)
	masters.DELETE("/trimming-courses/:id", perm(model.ResourceMasterTrimming, "delete"), h.DeleteTrimmingCourse)

	// Trimming Options
	masters.GET("/trimming-options", h.ListTrimmingOptions)
	masters.POST("/trimming-options", perm(model.ResourceMasterTrimming, "create"), h.CreateTrimmingOption)
	masters.PATCH("/trimming-options/reorder", perm(model.ResourceMasterTrimming, "edit"), h.ReorderTrimmingOptions)
	masters.GET("/trimming-options/:id", h.GetTrimmingOption)
	masters.PATCH("/trimming-options/:id", perm(model.ResourceMasterTrimming, "edit"), h.UpdateTrimmingOption)
	masters.DELETE("/trimming-options/:id", perm(model.ResourceMasterTrimming, "delete"), h.DeleteTrimmingOption)

	// Examination Types
	masters.GET("/examination-types", h.ListExaminationTypes)
	masters.POST("/examination-types", perm(model.ResourceMasterMedical, "create"), h.CreateExaminationType)
	masters.PATCH("/examination-types/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderExaminationTypes)
	masters.GET("/examination-types/:id", h.GetExaminationType)
	masters.PATCH("/examination-types/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateExaminationType)
	masters.DELETE("/examination-types/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteExaminationType)

	// Diagnosis Categories
	masters.GET("/diagnosis-categories", h.ListDiagnosisCategories)
	masters.POST("/diagnosis-categories", perm(model.ResourceMasterMedical, "create"), h.CreateDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderDiagnosisCategories)
	masters.GET("/diagnosis-categories/:id", h.GetDiagnosisCategory)
	masters.PATCH("/diagnosis-categories/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateDiagnosisCategory)
	masters.DELETE("/diagnosis-categories/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteDiagnosisCategory)

	// Diagnosis Names
	masters.GET("/diagnosis-names", h.ListDiagnosisNames)
	masters.POST("/diagnosis-names", perm(model.ResourceMasterMedical, "create"), h.CreateDiagnosisName)
	masters.PATCH("/diagnosis-names/reorder", perm(model.ResourceMasterMedical, "edit"), h.ReorderDiagnosisNames)
	masters.GET("/diagnosis-names/:id", h.GetDiagnosisName)
	masters.PATCH("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateDiagnosisName)
	masters.DELETE("/diagnosis-names/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteDiagnosisName)

	// Checkup Types
	masters.GET("/checkup-types", h.ListCheckupTypes)
	masters.POST("/checkup-types", perm(model.ResourceCheckups, "create"), h.CreateCheckupType)
	masters.PATCH("/checkup-types/reorder", perm(model.ResourceCheckups, "edit"), h.ReorderCheckupTypes)
	masters.GET("/checkup-types/:id", h.GetCheckupType)
	masters.PATCH("/checkup-types/:id", perm(model.ResourceCheckups, "edit"), h.UpdateCheckupType)
	masters.DELETE("/checkup-types/:id", perm(model.ResourceCheckups, "delete"), h.DeleteCheckupType)

	// Occupations
	masters.GET("/occupations", h.ListOccupations)
	masters.POST("/occupations", perm(model.ResourceMasterStaff, "create"), h.CreateOccupation)
	masters.PATCH("/occupations/reorder", perm(model.ResourceMasterStaff, "edit"), h.ReorderOccupations)
	masters.GET("/occupations/:id", h.GetOccupation)
	masters.PATCH("/occupations/:id", perm(model.ResourceMasterStaff, "edit"), h.UpdateOccupation)
	masters.DELETE("/occupations/:id", perm(model.ResourceMasterStaff, "delete"), h.DeleteOccupation)

	// Permission Groups — CRUD個別ガード
	masters.GET("/permission-groups", h.ListPermissionGroups)
	masters.GET("/permission-groups/:id", h.GetPermissionGroup)
	masters.POST("/permission-groups", perm(model.ResourceMasterPermission, "create"), h.CreatePermissionGroup)
	masters.PATCH("/permission-groups", perm(model.ResourceMasterPermission, "edit"), h.ReorderPermissionGroups)
	masters.PATCH("/permission-groups/:id", perm(model.ResourceMasterPermission, "edit"), h.UpdatePermissionGroup)
	masters.DELETE("/permission-groups/:id", perm(model.ResourceMasterPermission, "delete"), h.DeletePermissionGroup)
	masters.PUT("/permission-groups/:id/rules", perm(model.ResourceMasterPermission, "edit"), h.SetPermissionGroupRules)

	// Chief Complaint Categories
	masters.GET("/chief-complaint-categories", h.ListChiefComplaints)
	masters.POST("/chief-complaint-categories", perm(model.ResourceMasterMedical, "create"), h.CreateChiefComplaint)
	masters.GET("/chief-complaint-categories/:id", h.GetChiefComplaint)
	masters.PATCH("/chief-complaint-categories/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateChiefComplaint)
	masters.DELETE("/chief-complaint-categories/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteChiefComplaint)

	// Inquiry Templates
	masters.GET("/inquiry-templates", h.ListInquiryTemplates)
	masters.POST("/inquiry-templates", perm(model.ResourceMasterMedical, "create"), h.CreateInquiryTemplate)
	masters.GET("/inquiry-templates/:id", h.GetInquiryTemplate)
	masters.PATCH("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "edit"), h.UpdateInquiryTemplate)
	masters.DELETE("/inquiry-templates/:id", perm(model.ResourceMasterMedical, "delete"), h.DeleteInquiryTemplate)

	// Merchandise Items
	masters.GET("/merchandise-items", h.ListMerchandiseItems)
	masters.POST("/merchandise-items", perm(model.ResourceMasterMerchandise, "create"), h.CreateMerchandiseItem)
	masters.POST("/merchandise-items/reorder", perm(model.ResourceMasterMerchandise, "edit"), h.ReorderMerchandiseItems)
	masters.GET("/merchandise-items/:id", h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "edit"), h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "delete"), h.DeleteMerchandiseItem)
}
