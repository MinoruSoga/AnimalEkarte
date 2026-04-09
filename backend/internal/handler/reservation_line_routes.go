package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/middleware"
)

// RegisterLineReservationRoutes はLINE予約管理APIのルートを登録する
func (h *Handler) RegisterLineReservationRoutes(rg *gin.RouterGroup) {
	clinics := rg.Group("/clinics/:clinicId")

	// TASK-RES-010: 基本設定
	clinics.GET("/reservation-settings", h.GetReservationSetting)
	clinics.PUT("/reservation-settings", h.UpsertReservationSetting)

	// TASK-RES-011: コース
	courses := clinics.Group("/reservation-courses")
	courses.GET("", h.ListReservationCourses)
	courses.POST("", h.CreateReservationCourse)
	courses.PUT("/:id", h.UpdateReservationCourse)
	courses.DELETE("/:id", h.DeleteReservationCourse)
	courses.PATCH("/:id/status", h.PatchReservationCourseStatus)
	courses.PATCH("/:id/sort-order", h.PatchReservationCourseSortOrder)
	courses.POST("/:id/image", h.UploadReservationCourseImage)

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
	schedules := clinics.Group("/reservation-staffs/:staffId/schedules")
	schedules.GET("", h.ListReservationSchedules)
	schedules.PUT("/:date", h.UpsertReservationSchedule)
	schedules.DELETE("/:date", h.DeleteReservationSchedule)

	// TASK-RES-014: 予約管理
	reservations := clinics.Group("/reservations")
	reservations.GET("", h.ListReservationsAdmin)
	reservations.POST("", h.CreateReservationAdmin)
	reservations.DELETE("/:id", h.DeleteReservationAdmin)

	// TASK-RES-015: 顧客管理
	customers := clinics.Group("/reservation-customers")
	customers.GET("", h.ListReservationCustomers)
	customers.PATCH("/:id/link-owner", h.LinkOwnerToReservationCustomer)
}

// RegisterLiffRoutes はLIFF公開APIのルートを登録する（LINE IDトークン認証）。
func (h *Handler) RegisterLiffRoutes(r *gin.Engine) {
	liffAuth := middleware.LiffAuth(h.repos.ReservationCustomerMgr)

	liff := r.Group("/api/liff/:clinicId")

	// 設定は認証不要（トップページ表示用）
	liff.GET("/settings", h.GetLiffSettings)

	// 以下は LINE IDトークン認証が必要
	authed := liff.Group("")
	authed.Use(liffAuth)

	authed.GET("/profile", h.GetLiffProfile)
	authed.GET("/courses", h.GetLiffCourses)
	authed.GET("/staffs", h.GetLiffStaffs)
	authed.GET("/available-dates", h.GetLiffAvailableDates)
	authed.GET("/available-times", h.GetLiffAvailableTimes)
	authed.POST("/reservations", h.CreateLiffReservation)
	authed.GET("/my-reservations", h.GetLiffMyReservations)
	authed.DELETE("/my-reservations/:id", h.CancelLiffReservation)
}
