package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/middleware"
	"github.com/animal-ekarte/backend/internal/model"
)

// RegisterLineReservationRoutes はLINE予約管理APIのルートを登録する
//
// clinic_id の実際の判定は JWT の extractClinicID(c) を使うため、
// URL の :clinic_id パラメータは識別子として参照しない。
// 子リソースのパスパラメータは固有名（typeId/staffId/reservationId/customerId）を使用する。
func (h *Handler) RegisterLineReservationRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinic_id")

	// TASK-RES-010: 基本設定
	clinics.GET("/line-reservation-settings", h.RequirePermission(string(model.ResourceHospitalSettings), "view"), h.GetLineReservationSetting)
	clinics.PUT("/line-reservation-settings", h.RequirePermission(string(model.ResourceHospitalSettings), "edit"), h.SaveLineReservationSetting)

	// TASK-RES-011: 予約区分（LINE管理用）
	types := clinics.Group("/reservation-types")
	types.GET("", h.RequirePermission(string(model.ResourceMasterReservationType), "view"), h.ListReservationTypeLiffs)
	types.POST("", h.RequirePermission(string(model.ResourceMasterReservationType), "create"), h.CreateReservationTypeLiff)
	types.PUT("/:id", h.RequirePermission(string(model.ResourceMasterReservationType), "edit"), h.UpdateReservationTypeLiff)
	types.DELETE("/:id", h.RequirePermission(string(model.ResourceMasterReservationType), "delete"), h.DeleteReservationTypeLiff)
	types.PATCH("/:id/status", h.RequirePermission(string(model.ResourceMasterReservationType), "edit"), h.PatchReservationTypeLiffStatus)
	types.PATCH("/:id/sort-order", h.RequirePermission(string(model.ResourceMasterReservationType), "edit"), h.PatchReservationTypeLiffSortOrder)
	types.POST("/:id/image", h.RequirePermission(string(model.ResourceMasterReservationType), "create"), h.UploadReservationTypeLiffImage)

	// TASK-RES-012: スタッフ
	staffs := clinics.Group("/reservation-staffs")
	staffs.GET("", h.RequirePermission(string(model.ResourceMasterStaff), "view"), h.ListReservationStaffs)
	staffs.POST("", h.RequirePermission(string(model.ResourceMasterStaff), "create"), h.CreateReservationStaff)
	staffs.PUT("/:staffId", h.RequirePermission(string(model.ResourceMasterStaff), "edit"), h.UpdateReservationStaff)
	staffs.DELETE("/:staffId", h.RequirePermission(string(model.ResourceMasterStaff), "delete"), h.DeleteReservationStaff)
	staffs.PATCH("/:staffId/status", h.RequirePermission(string(model.ResourceMasterStaff), "edit"), h.PatchReservationStaffStatus)
	staffs.PATCH("/:staffId/sort-order", h.RequirePermission(string(model.ResourceMasterStaff), "edit"), h.PatchReservationStaffSortOrder)
	staffs.POST("/:staffId/image", h.RequirePermission(string(model.ResourceMasterStaff), "create"), h.UploadReservationStaffImage)

	// TASK-RES-013: スタッフ個人スケジュール
	schedules := clinics.Group("/reservation-staffs/:staffId/schedules")
	schedules.GET("", h.RequirePermission(string(model.ResourceMasterStaff), "view"), h.ListReservationSchedules)
	schedules.PUT("/:date", h.RequirePermission(string(model.ResourceMasterStaff), "edit"), h.UpsertReservationSchedule)
	schedules.DELETE("/:date", h.RequirePermission(string(model.ResourceMasterStaff), "delete"), h.DeleteReservationSchedule)

	// TASK-RES-014: 予約管理
	reservations := clinics.Group("/reservations")
	reservations.GET("", h.RequirePermission(string(model.ResourceReservations), "view"), h.ListReservationsAdmin)
	reservations.POST("", h.RequirePermission(string(model.ResourceReservations), "create"), h.CreateReservationAdmin)
	reservations.DELETE("/:reservationId", h.RequirePermission(string(model.ResourceReservations), "delete"), h.DeleteReservationAdmin)

	// TASK-RES-015: 顧客管理
	customers := clinics.Group("/line-customers")
	customers.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListLineCustomers)
	customers.PATCH("/:customerId/link-owner", h.RequirePermission(string(model.ResourceOwners), "edit"), h.LinkOwnerToLineCustomer)
}

// RegisterLiffRoutes はLIFF公開APIのルートを登録する（LINE IDトークン認証）。
func (h *Handler) RegisterLiffRoutes(r *gin.Engine) {
	liffAuth := middleware.LiffAuth(h.repos.LineCustomerMgr, h.repos.LineReservationSetting)

	liff := r.Group("/api/liff/:clinicId")

	// 設定は認証不要（トップページ表示用）
	liff.GET("/settings", h.GetLiffSettings)
	// BE-021: LINE User ID 紐付け（link_token + line_id_token で自己認証）
	liff.POST("/link", h.LinkLiffAccount)

	// 以下は LINE IDトークン認証が必要
	authed := liff.Group("")
	authed.Use(liffAuth)

	authed.GET("/profile", h.GetLiffProfile)
	authed.GET("/courses", h.GetLiffTypes)
	authed.GET("/trimming-courses", h.GetLiffTrimmingCourses) // BE-120
	authed.GET("/trimming-options", h.GetLiffTrimmingOptions) // BE-120
	authed.GET("/staffs", h.GetLiffStaffs)
	authed.GET("/available-dates", h.GetLiffAvailableDates)
	authed.GET("/available-times", h.GetLiffAvailableTimes)
	authed.POST("/reservations", h.CreateLiffReservation)
	authed.GET("/my-reservations", h.GetLiffMyReservations)
	authed.DELETE("/my-reservations/:id", h.CancelLiffReservation)
}
