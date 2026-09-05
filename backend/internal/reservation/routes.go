package reservation

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action).
// reservation の consumer-side view — permission_middleware.go は target:auth（未移行）のため
// reservation は auth を import せず、composition root (cmd/api/main.go) が具象を注入する。
// internal/medicalrecord の同名型と同型（BE9-2B pilot 以来の規約）。
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point（openapi_route_drift_test.go の
// buildFuncsFromDir が bare 名で func map を構築するため、per-entity 複数 RegisterRoutes は禁止）。
type Handler struct {
	reservationType        *ReservationTypeHandler
	reservationTypeGroup   *ReservationTypeGroupHandler
	reservationTypeLiff    *ReservationTypeLiffHandler
	reservationStaff       *ReservationStaffHandler
	reservationSchedule    *ReservationScheduleHandler
	reservationCRUD        *ReservationHandler
	reservationAdmin       *ReservationAdminHandler
	lineReservationSetting *LineReservationSettingHandler
	liff                   *LiffHandler
	// LIFF 認証/レートリミット middleware（middleware package は service を import しており
	// reservation からの import は循環——PermissionMiddleware と同じく合成済み closure を
	// composition root (cmd/api/main.go) が注入する）
	liffAuth      gin.HandlerFunc
	liffRateLimit func(limit int) gin.HandlerFunc
	// linkLiffAccount は LINE 紐付け handler（line_link=lstep 系・未移行）の注入 closure。
	linkLiffAccount   gin.HandlerFunc
	requirePermission PermissionMiddleware
}

// NewHandler は reservation domain の routing composition を構築する。
func NewHandler(
	reservationType *ReservationTypeHandler,
	reservationTypeGroup *ReservationTypeGroupHandler,
	reservationTypeLiff *ReservationTypeLiffHandler,
	reservationStaff *ReservationStaffHandler,
	reservationSchedule *ReservationScheduleHandler,
	reservationCRUD *ReservationHandler,
	reservationAdmin *ReservationAdminHandler,
	lineReservationSetting *LineReservationSettingHandler,
	liff *LiffHandler,
	liffAuth gin.HandlerFunc,
	liffRateLimit func(limit int) gin.HandlerFunc,
	linkLiffAccount gin.HandlerFunc,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		reservationType:        reservationType,
		reservationTypeGroup:   reservationTypeGroup,
		reservationTypeLiff:    reservationTypeLiff,
		reservationStaff:       reservationStaff,
		reservationSchedule:    reservationSchedule,
		reservationCRUD:        reservationCRUD,
		reservationAdmin:       reservationAdmin,
		lineReservationSetting: lineReservationSetting,
		liff:                   liff,
		liffAuth:               liffAuth,
		liffRateLimit:          liffRateLimit,
		linkLiffAccount:        linkLiffAccount,
		requirePermission:      requirePermission,
	}
}

// RegisterRoutes は reservation domain の全 route を登録する。
// RBAC は旧 internal/handler 登録箇所（master_routes.go / reservation_line_routes.go）からの逐語転記。
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.requirePermission(string(resource), action)
	}

	masters := rg.Group("/masters")

	// Reservation Type Groups（旧 master_routes.go 逐語）
	masters.GET("/reservation-type-groups", perm(model.ResourceMasterReservationType, "view"), h.reservationTypeGroup.ListReservationTypeGroups)
	masters.POST("/reservation-type-groups", perm(model.ResourceMasterReservationType, "create"), h.reservationTypeGroup.CreateReservationTypeGroup)
	masters.PATCH("/reservation-type-groups/reorder", perm(model.ResourceMasterReservationType, "edit"), h.reservationTypeGroup.ReorderReservationTypeGroups)
	masters.GET("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "view"), h.reservationTypeGroup.GetReservationTypeGroup)
	masters.PATCH("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "edit"), h.reservationTypeGroup.UpdateReservationTypeGroup)
	masters.DELETE("/reservation-type-groups/:id", perm(model.ResourceMasterReservationType, "delete"), h.reservationTypeGroup.DeleteReservationTypeGroup)

	// Reservation Types（旧 master_routes.go 逐語）
	masters.GET("/reservation-types", perm(model.ResourceMasterReservationType, "view"), h.reservationType.ListReservationTypes)
	masters.POST("/reservation-types", perm(model.ResourceMasterReservationType, "create"), h.reservationType.CreateReservationType)
	masters.PATCH("/reservation-types/reorder", perm(model.ResourceMasterReservationType, "edit"), h.reservationType.ReorderReservationTypes)
	masters.GET("/reservation-types/:id", perm(model.ResourceMasterReservationType, "view"), h.reservationType.GetReservationType)
	masters.PATCH("/reservation-types/:id", perm(model.ResourceMasterReservationType, "edit"), h.reservationType.UpdateReservationType)
	masters.DELETE("/reservation-types/:id", perm(model.ResourceMasterReservationType, "delete"), h.reservationType.DeleteReservationType)

	masters.GET("/reservation-types/:id/unavailable-times", perm(model.ResourceMasterReservationType, "view"), h.reservationType.ListUnavailableTimes)
	masters.POST("/reservation-types/:id/unavailable-times", perm(model.ResourceMasterReservationType, "edit"), h.reservationType.CreateUnavailableTime)
	masters.DELETE("/reservation-types/:id/unavailable-times/:unavailable_time_id", perm(model.ResourceMasterReservationType, "delete"), h.reservationType.DeleteUnavailableTime)

	masters.GET("/reservation-types/:id/available-slots", perm(model.ResourceMasterReservationType, "view"), h.reservationType.ListAvailableSlots)
	masters.POST("/reservation-types/:id/available-slots", perm(model.ResourceMasterReservationType, "edit"), h.reservationType.CreateAvailableSlot)
	masters.DELETE("/reservation-types/:id/available-slots/:available_slot_id", perm(model.ResourceMasterReservationType, "delete"), h.reservationType.DeleteAvailableSlot)

	masters.GET("/reservation-types/:id/occupations", perm(model.ResourceMasterReservationType, "view"), h.reservationType.ListReservationTypeOccupations)
	masters.POST("/reservation-types/:id/occupations", perm(model.ResourceMasterReservationType, "edit"), h.reservationType.LinkReservationTypeOccupation)
	masters.DELETE("/reservation-types/:id/occupations/:occupation_id", perm(model.ResourceMasterReservationType, "delete"), h.reservationType.UnlinkReservationTypeOccupation)

	// 予約区分（LINE管理用・旧 reservation_line_routes.go 逐語。clinic_id の実判定は JWT 側）
	clinics := rg.Group("/clinics/:clinic_id")
	types := clinics.Group("/reservation-types")
	types.GET("", h.requirePermission(string(model.ResourceMasterReservationType), "view"), h.reservationTypeLiff.ListReservationTypeLiffs)
	types.POST("", h.requirePermission(string(model.ResourceMasterReservationType), "create"), h.reservationTypeLiff.CreateReservationTypeLiff)
	types.PUT("/:id", h.requirePermission(string(model.ResourceMasterReservationType), "edit"), h.reservationTypeLiff.UpdateReservationTypeLiff)
	types.DELETE("/:id", h.requirePermission(string(model.ResourceMasterReservationType), "delete"), h.reservationTypeLiff.DeleteReservationTypeLiff)
	types.PATCH("/:id/status", h.requirePermission(string(model.ResourceMasterReservationType), "edit"), h.reservationTypeLiff.UpdateReservationTypeLiffStatus)
	types.PATCH("/:id/sort-order", h.requirePermission(string(model.ResourceMasterReservationType), "edit"), h.reservationTypeLiff.UpdateReservationTypeLiffSortOrder)
	types.POST("/:id/image", h.requirePermission(string(model.ResourceMasterReservationType), "create"), h.reservationTypeLiff.UploadReservationTypeLiffImage)

	// 予約スタッフ（LINE管理用・旧 reservation_line_routes.go 逐語）
	staffs := clinics.Group("/reservation-staffs")
	staffs.GET("", h.requirePermission(string(model.ResourceMasterStaff), "view"), h.reservationStaff.ListReservationStaffs)
	staffs.POST("", h.requirePermission(string(model.ResourceMasterStaff), "create"), h.reservationStaff.CreateReservationStaff)
	staffs.PUT("/:staffId", h.requirePermission(string(model.ResourceMasterStaff), "edit"), h.reservationStaff.UpdateReservationStaff)
	staffs.DELETE("/:staffId", h.requirePermission(string(model.ResourceMasterStaff), "delete"), h.reservationStaff.DeleteReservationStaff)
	staffs.PATCH("/:staffId/status", h.requirePermission(string(model.ResourceMasterStaff), "edit"), h.reservationStaff.UpdateReservationStaffStatus)
	staffs.PATCH("/:staffId/sort-order", h.requirePermission(string(model.ResourceMasterStaff), "edit"), h.reservationStaff.UpdateReservationStaffSortOrder)
	staffs.POST("/:staffId/image", h.requirePermission(string(model.ResourceMasterStaff), "create"), h.reservationStaff.UploadReservationStaffImage)

	// 予約スケジュール（LINE管理用・旧 reservation_line_routes.go 逐語）
	schedules := clinics.Group("/reservation-staffs/:staffId/schedules")
	schedules.GET("", h.requirePermission(string(model.ResourceMasterStaff), "view"), h.reservationSchedule.ListReservationSchedules)
	schedules.PUT("/:date", h.requirePermission(string(model.ResourceMasterStaff), "edit"), h.reservationSchedule.UpsertReservationSchedule)
	schedules.DELETE("/:date", h.requirePermission(string(model.ResourceMasterStaff), "delete"), h.reservationSchedule.DeleteReservationSchedule)

	// LINE 予約基本設定。対象医院は URL :clinic_id を所属集合で認可する（JWT 選択医院ではない）。
	clinics.GET("/line-reservation-settings", h.requirePermission(string(model.ResourceHospitalSettings), "view"), h.lineReservationSetting.GetLineReservationSetting)
	clinics.PUT("/line-reservation-settings", h.requirePermission(string(model.ResourceHospitalSettings), "edit"), h.lineReservationSetting.SaveLineReservationSetting)

	// 予約管理（LINE管理用・旧 reservation_line_routes.go 逐語）
	adminReservations := clinics.Group("/reservations")
	adminReservations.GET("", h.requirePermission(string(model.ResourceReservations), "view"), h.reservationAdmin.ListReservationsAdmin)
	adminReservations.POST("", h.requirePermission(string(model.ResourceReservations), "create"), h.reservationAdmin.CreateReservationAdmin)
	adminReservations.DELETE("/:reservationId", h.requirePermission(string(model.ResourceReservations), "delete"), h.reservationAdmin.DeleteReservationAdmin)

	// 予約 CRUD（旧 handler.go RegisterReservationRoutes 逐語）
	reservations := rg.Group("/reservations")
	reservations.GET("", h.requirePermission(string(model.ResourceReservations), "view"), h.reservationCRUD.ListReservations)
	reservations.GET("/available-times", h.requirePermission(string(model.ResourceReservations), "view"), h.reservationCRUD.GetReservationAvailableTimes)
	reservations.GET("/:id", h.requirePermission(string(model.ResourceReservations), "view"), h.reservationCRUD.GetReservation)
	reservations.POST("", h.requirePermission(string(model.ResourceReservations), "create"), h.reservationCRUD.CreateReservation)
	reservations.POST("/batch", h.requirePermission(string(model.ResourceReservations), "create"), h.reservationCRUD.CreateReservationBatch)
	reservations.PATCH("/:id", h.requirePermission(string(model.ResourceReservations), "edit"), h.reservationCRUD.UpdateReservation)
	reservations.PATCH("/:id/reservation-route", h.requirePermission(string(model.ResourceReservations), "edit"), h.reservationCRUD.UpdateReservationReservationRoute)
	reservations.DELETE("/:id", h.requirePermission(string(model.ResourceReservations), "delete"), h.reservationCRUD.DeleteReservation)
}

// RegisterLiffRoutes はLIFF公開APIのルートを登録する（LINE IDトークン認証）。
// ctx はレートリミッタークリーンアップゴルーチンのライフタイム管理に使用する。
func (h *Handler) RegisterLiffRoutes(r *gin.Engine) {
	liffAuth := h.liffAuth

	// LIFF公開エンドポイントへのレートリミット（IPアドレスベース）
	// /link: 10回/分（アカウント紐付けは高頻度アクセス不要）
	// /settings, /my-reservations: 30回/分（読み取り系は余裕を持たせる）
	// store の構築（ctx 付き cleanup goroutine）は composition root 側（liffRateLimit closure に内包）。

	liff := r.Group("/api/liff/:clinicId")

	// 設定は認証不要（トップページ表示用）— 30回/分
	liff.GET("/settings", h.liffRateLimit(30), h.liff.GetLiffSettings)
	// BE-021: LINE User ID 紐付け（link_token + line_id_token で自己認証）— 10回/分
	liff.POST("/link", h.liffRateLimit(10), h.linkLiffAccount)

	// 以下は LINE IDトークン認証が必要
	authed := liff.Group("")
	authed.Use(h.liffRateLimit(60), liffAuth)

	authed.GET("/profile", h.liff.GetLiffProfile)
	authed.GET("/courses", h.liff.GetLiffTypes)
	authed.GET("/trimming-courses", h.liff.GetLiffTrimmingCourses) // BE-120
	authed.GET("/trimming-options", h.liff.GetLiffTrimmingOptions) // BE-120
	authed.GET("/staffs", h.liff.GetLiffStaffs)
	authed.GET("/available-dates", h.liff.GetLiffAvailableDates)
	authed.GET("/available-times", h.liff.GetLiffAvailableTimes)
	authed.POST("/reservations", h.liff.CreateLiffReservation)
	authed.GET("/my-reservations", h.liffRateLimit(30), h.liff.GetLiffMyReservations)
	authed.DELETE("/my-reservations/:id", h.liff.CancelLiffReservation)
	authed.GET("/health-card", h.liff.GetLiffHealthCard)
}
