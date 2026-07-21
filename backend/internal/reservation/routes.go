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
	reservationType      *ReservationTypeHandler
	reservationTypeGroup *ReservationTypeGroupHandler
	reservationTypeLiff  *ReservationTypeLiffHandler
	reservationStaff     *ReservationStaffHandler
	reservationSchedule  *ReservationScheduleHandler
	reservationCRUD      *ReservationHandler
	reservationAdmin     *ReservationAdminHandler
	requirePermission    PermissionMiddleware
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
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		reservationType:      reservationType,
		reservationTypeGroup: reservationTypeGroup,
		reservationTypeLiff:  reservationTypeLiff,
		reservationStaff:     reservationStaff,
		reservationSchedule:  reservationSchedule,
		reservationCRUD:      reservationCRUD,
		reservationAdmin:     reservationAdmin,
		requirePermission:    requirePermission,
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
	reservations.PATCH("/:id", h.requirePermission(string(model.ResourceReservations), "edit"), h.reservationCRUD.UpdateReservation)
	reservations.PATCH("/:id/reservation-route", h.requirePermission(string(model.ResourceReservations), "edit"), h.reservationCRUD.UpdateReservationReservationRoute)
	reservations.DELETE("/:id", h.requirePermission(string(model.ResourceReservations), "delete"), h.reservationCRUD.DeleteReservation)
}
