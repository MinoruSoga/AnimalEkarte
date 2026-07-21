package handler

import (
	"context"

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

	// TASK-RES-011: 予約区分（LINE管理用）routes は internal/reservation.RegisterRoutes へ移動（BE9-2C R①）

	// TASK-RES-012: スタッフ
	// 予約スタッフ routes は internal/reservation.RegisterRoutes へ移動（BE9-2C R②）

	// TASK-RES-013: スタッフ個人スケジュール
	// 予約スケジュール routes は internal/reservation.RegisterRoutes へ移動（BE9-2C R②）

	// 予約管理（LINE管理用）routes は internal/reservation.RegisterRoutes へ移動（BE9-2C R④）

	// TASK-RES-015: 顧客管理
	customers := clinics.Group("/line-customers")
	customers.GET("", h.RequirePermission(string(model.ResourceOwners), "view"), h.ListLineCustomers)
	customers.PATCH("/:customerId/link-owner", h.RequirePermission(string(model.ResourceOwners), "edit"), h.LinkOwnerToLineCustomer)
}

// RegisterLiffRoutes はLIFF公開APIのルートを登録する（LINE IDトークン認証）。
// ctx はレートリミッタークリーンアップゴルーチンのライフタイム管理に使用する。
func (h *Handler) RegisterLiffRoutes(ctx context.Context, r *gin.Engine) {
	liffAuth := middleware.LiffAuth(h.liffCustomerLookup, h.liffSettingLookup)

	// LIFF公開エンドポイントへのレートリミット（IPアドレスベース）
	// /link: 10回/分（アカウント紐付けは高頻度アクセス不要）
	// /settings, /my-reservations: 30回/分（読み取り系は余裕を持たせる）
	liffRateLimitStore := middleware.NewRateLimitStore(ctx)

	liff := r.Group("/api/liff/:clinicId")

	// 設定は認証不要（トップページ表示用）— 30回/分
	liff.GET("/settings", middleware.LiffRateLimit(liffRateLimitStore, 30), h.GetLiffSettings)
	// BE-021: LINE User ID 紐付け（link_token + line_id_token で自己認証）— 10回/分
	liff.POST("/link", middleware.LiffRateLimit(liffRateLimitStore, 10), h.LinkLiffAccount)

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
	authed.GET("/my-reservations", middleware.LiffRateLimit(liffRateLimitStore, 30), h.GetLiffMyReservations)
	authed.DELETE("/my-reservations/:id", h.CancelLiffReservation)
	authed.GET("/health-card", h.GetLiffHealthCard)
}
