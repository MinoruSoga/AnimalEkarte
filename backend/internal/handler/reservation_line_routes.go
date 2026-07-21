package handler

import (
	"github.com/gin-gonic/gin"

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

// LIFF 公開 API routes は internal/reservation.RegisterLiffRoutes へ移動（BE9-2C R⑤）
