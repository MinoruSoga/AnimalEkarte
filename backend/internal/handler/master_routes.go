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

	// Medicines
	// #201 B-2c: 薬剤 × 種の投与量計算パラメータ（dog/cat の mg/kg）。species を自然キーに upsert。

	// Vaccines: moved to internal/medicalrecord.Handler.RegisterRoutes (BE9-2D — composed
	// directly in cmd/api/main.go, ADR-006 aggregator 非経由). See internal/medicalrecord/routes.go.

	// Insurances

	// Reservation Category Groups

	// Reservation Types
	// 予約不可時間
	// 予約可能開始時刻
	// 職種紐付け

	// Consultations

	// Procedures

	// cages/medicines(+dose-params)/consultations/procedures: BE9-2D ⑥ で
	// internal/medicalrecord の RegisterRoutes へ移動。
	// Hospitalization Plans
	// hospitalization-plans: BE9-2D ⑤ で internal/medicalrecord の RegisterRoutes へ移動。

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

	// Examination Types / Diagnosis Categories / Diagnosis Names: moved to
	// internal/medicalrecord.Handler.RegisterRoutes (BE9-2C master-CRUD slice — composed
	// directly in cmd/api/main.go, ADR-006 "aggregator 非経由"). See
	// internal/medicalrecord/routes.go.

	// Checkup Types (+ :id/fields): moved to internal/medicalrecord.Handler.RegisterRoutes
	// (BE9-2D — composed in cmd/api/main.go). See internal/medicalrecord/routes.go.

	// Occupations
	masters.GET("/occupations", perm(model.ResourceMasterStaff, "view"), h.ListOccupations)
	masters.POST("/occupations", perm(model.ResourceMasterStaff, "create"), h.CreateOccupation)
	masters.PATCH("/occupations/reorder", perm(model.ResourceMasterStaff, "edit"), h.ReorderOccupations)
	masters.GET("/occupations/:id", perm(model.ResourceMasterStaff, "view"), h.GetOccupation)
	masters.PATCH("/occupations/:id", perm(model.ResourceMasterStaff, "edit"), h.UpdateOccupation)
	masters.DELETE("/occupations/:id", perm(model.ResourceMasterStaff, "delete"), h.DeleteOccupation)

	h.RegisterPermissionGroupRoutes(masters)

	// Chief Complaint Categories: moved to internal/medicalrecord.Handler.RegisterRoutes
	// (BE9-2C master-CRUD slice). See internal/medicalrecord/routes.go.

	// Inquiry Templates: moved to internal/medicalrecord.Handler.RegisterRoutes (BE9-2D —
	// composed in cmd/api/main.go). See internal/medicalrecord/routes.go.

	// Merchandise Items
	masters.GET("/merchandise-items", perm(model.ResourceMasterMerchandise, "view"), h.ListMerchandiseItems)
	masters.POST("/merchandise-items", perm(model.ResourceMasterMerchandise, "create"), h.CreateMerchandiseItem)
	masters.PATCH("/merchandise-items/reorder", perm(model.ResourceMasterMerchandise, "edit"), h.ReorderMerchandiseItems)
	masters.GET("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "view"), h.GetMerchandiseItem)
	masters.PATCH("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "edit"), h.UpdateMerchandiseItem)
	masters.DELETE("/merchandise-items/:id", perm(model.ResourceMasterMerchandise, "delete"), h.DeleteMerchandiseItem)
}
