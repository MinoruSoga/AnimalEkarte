package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/middleware"
)

// RegisterLineReservationRoutes はLINE予約管理APIのルートを登録する
func (h *Handler) RegisterLineReservationRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:id")

	// TASK-RES-010: 基本設定
	clinics.GET("/line-reservation-settings", h.GetLineReservationSetting)
	clinics.PUT("/line-reservation-settings", h.UpsertLineReservationSetting)

	// TASK-RES-011: 予約区分（LINE管理用）
	types := clinics.Group("/reservation-types")
	types.GET("", h.ListReservationTypeLiffs)
	types.POST("", h.CreateReservationTypeLiff)
	types.PUT("/:id", h.UpdateReservationTypeLiff)
	types.DELETE("/:id", h.DeleteReservationTypeLiff)
	types.PATCH("/:id/status", h.PatchReservationTypeLiffStatus)
	types.PATCH("/:id/sort-order", h.PatchReservationTypeLiffSortOrder)
	types.POST("/:id/image", h.UploadReservationTypeLiffImage)

	// TASK-RES-012: スタッフ
	staffs := clinics.Group("/reservation-staffs")
	staffs.GET("", h.ListReservationStaffs)
	staffs.POST("", h.CreateReservationStaff)
	staffs.PUT("/:id", h.UpdateReservationStaff)
	staffs.DELETE("/:id", h.DeleteReservationStaff)
	staffs.PATCH("/:id/status", h.PatchReservationStaffStatus)
	staffs.PATCH("/:id/sort-order", h.PatchReservationStaffSortOrder)
	staffs.POST("/:id/image", h.UploadReservationStaffImage)

	// TASK-RES-013: スタッフ個人スケジュール
	schedules := clinics.Group("/reservation-staffs/:id/schedules")
	schedules.GET("", h.ListReservationSchedules)
	schedules.PUT("/:date", h.UpsertReservationSchedule)
	schedules.DELETE("/:date", h.DeleteReservationSchedule)

	// TASK-RES-014: 予約管理
	reservations := clinics.Group("/reservations")
	reservations.GET("", h.ListReservationsAdmin)
	reservations.POST("", h.CreateReservationAdmin)
	reservations.DELETE("/:id", h.DeleteReservationAdmin)

	// TASK-RES-015: 顧客管理
	customers := clinics.Group("/line-customers")
	customers.GET("", h.ListLineCustomers)
	customers.PATCH("/:id/link-owner", h.LinkOwnerToLineCustomer)
}

// RegisterLiffRoutes はLIFF公開APIのルートを登録する（LINE IDトークン認証）。
func (h *Handler) RegisterLiffRoutes(r *gin.Engine) {
	liffAuth := middleware.LiffAuth(h.repos.LineCustomerMgr, h.repos.LineReservationSetting)

	liff := r.Group("/api/liff/:clinicId")

	// 設定は認証不要（トップページ表示用）
	liff.GET("/settings", h.GetLiffSettings)

	// 以下は LINE IDトークン認証が必要
	authed := liff.Group("")
	authed.Use(liffAuth)

	authed.GET("/profile", h.GetLiffProfile)
	authed.GET("/types", h.GetLiffTypes)
	authed.GET("/staffs", h.GetLiffStaffs)
	authed.GET("/available-dates", h.GetLiffAvailableDates)
	authed.GET("/available-times", h.GetLiffAvailableTimes)
	authed.POST("/reservations", h.CreateLiffReservation)
	authed.GET("/my-reservations", h.GetLiffMyReservations)
	authed.DELETE("/my-reservations/:id", h.CancelLiffReservation)
}
