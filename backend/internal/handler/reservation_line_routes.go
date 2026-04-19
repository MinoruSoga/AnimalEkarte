package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/middleware"
)

// RegisterLineReservationRoutes はLINE予約管理APIのルートを登録する
//
// BUG-LINE-005: 旧実装では `/clinics/:id/.../:id` のように `:id` が重複しており、
// Gin の c.Param("id") が最初にマッチした clinic_id を返すため、個別リソース操作が
// 他レコードへ誤って適用される CRITICAL SECURITY バグがあった。
// ネストした子リソースのパスパラメータは全て固有名（typeId/staffId/reservationId/customerId）に変更する。
// 親の `/clinics/:id` は clinic_handler.go 側の CRUD ルート（/clinics/:id）と整合させるため `:id` のまま保持し、
// 実際の clinic_id 判定は JWT の `extractClinicID(c)` を使うため URL 側の `:id` は識別子として参照しない。
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
	staffs.PUT("/:staffId", h.UpdateReservationStaff)
	staffs.DELETE("/:staffId", h.DeleteReservationStaff)
	staffs.PATCH("/:staffId/status", h.PatchReservationStaffStatus)
	staffs.PATCH("/:staffId/sort-order", h.PatchReservationStaffSortOrder)
	staffs.POST("/:staffId/image", h.UploadReservationStaffImage)

	// TASK-RES-013: スタッフ個人スケジュール
	schedules := clinics.Group("/reservation-staffs/:staffId/schedules")
	schedules.GET("", h.ListReservationSchedules)
	schedules.PUT("/:date", h.UpsertReservationSchedule)
	schedules.DELETE("/:date", h.DeleteReservationSchedule)

	// TASK-RES-014: 予約管理
	reservations := clinics.Group("/reservations")
	reservations.GET("", h.ListReservationsAdmin)
	reservations.POST("", h.CreateReservationAdmin)
	reservations.DELETE("/:reservationId", h.DeleteReservationAdmin)

	// TASK-RES-015: 顧客管理
	customers := clinics.Group("/line-customers")
	customers.GET("", h.ListLineCustomers)
	customers.PATCH("/:customerId/link-owner", h.LinkOwnerToLineCustomer)
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
