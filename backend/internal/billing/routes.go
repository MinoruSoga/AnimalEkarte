package billing

import (
	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/model"
)

// PermissionMiddleware builds the gin.HandlerFunc that gates a route on (resource, action)
// （medicalrecord/reservation と同型・composition root が具象を注入する）。
type PermissionMiddleware func(resource, action string) gin.HandlerFunc

// Handler composes this slice's per-entity handlers and registers their routes under a
// single, package-unique RegisterRoutes entry point（openapi_route_drift_test.go 規約）。
type Handler struct {
	insurance           *InsuranceHandler
	campaign            *CampaignHandler
	paymentMethodMaster *PaymentMethodMasterHandler
	estimate            *EstimateHandler
	billingConfirmation *BillingConfirmationHandler
	billingItem         *BillingItemHandler
	refund              *RefundHandler
	accounting          *AccountingHandler
	cashRegister        *CashRegisterHandler
	accountingReport    *AccountingReportHandler
	requirePermission   PermissionMiddleware
}

// NewHandler は billing domain の routing composition を構築する。
func NewHandler(
	insurance *InsuranceHandler,
	campaign *CampaignHandler,
	paymentMethodMaster *PaymentMethodMasterHandler,
	estimate *EstimateHandler,
	billingConfirmation *BillingConfirmationHandler,
	billingItem *BillingItemHandler,
	refund *RefundHandler,
	accounting *AccountingHandler,
	cashRegister *CashRegisterHandler,
	accountingReport *AccountingReportHandler,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		insurance:           insurance,
		campaign:            campaign,
		paymentMethodMaster: paymentMethodMaster,
		estimate:            estimate,
		billingConfirmation: billingConfirmation,
		billingItem:         billingItem,
		refund:              refund,
		accounting:          accounting,
		cashRegister:        cashRegister,
		accountingReport:    accountingReport,
		requirePermission:   requirePermission,
	}
}

// RegisterRoutes は billing domain の全 route を登録する（旧登録箇所からの RBAC 逐語転記）。
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	perm := func(resource model.Resource, action string) gin.HandlerFunc {
		return h.requirePermission(string(resource), action)
	}

	masters := rg.Group("/masters")

	// Insurances（旧 master_routes.go 逐語）
	masters.GET("/insurances", perm(model.ResourceMasterInsurance, "view"), h.insurance.ListInsurances)
	masters.POST("/insurances", perm(model.ResourceMasterInsurance, "create"), h.insurance.CreateInsurance)
	masters.PATCH("/insurances/reorder", perm(model.ResourceMasterInsurance, "edit"), h.insurance.ReorderInsurances)
	masters.GET("/insurances/:id", perm(model.ResourceMasterInsurance, "view"), h.insurance.GetInsurance)
	masters.PATCH("/insurances/:id", perm(model.ResourceMasterInsurance, "edit"), h.insurance.UpdateInsurance)
	masters.DELETE("/insurances/:id", perm(model.ResourceMasterInsurance, "delete"), h.insurance.DeleteInsurance)

	// Campaigns（旧 master_routes.go 逐語）
	masters.GET("/campaigns", perm(model.ResourceAccounting, "view"), h.campaign.ListCampaigns)
	masters.POST("/campaigns", perm(model.ResourceAccounting, "create"), h.campaign.CreateCampaign)
	masters.PATCH("/campaigns/reorder", perm(model.ResourceAccounting, "edit"), h.campaign.ReorderCampaigns)
	masters.GET("/campaigns/:id", perm(model.ResourceAccounting, "view"), h.campaign.GetCampaign)
	masters.PATCH("/campaigns/:id", perm(model.ResourceAccounting, "edit"), h.campaign.UpdateCampaign)
	masters.DELETE("/campaigns/:id", perm(model.ResourceAccounting, "delete"), h.campaign.DeleteCampaign)

	// 支払方法マスタ（旧 payment_method_master_handler.go RegisterPaymentMethodMasterRoutes 逐語）
	pm := rg.Group("/payment-methods")
	pm.GET("", h.requirePermission(string(model.ResourcePaymentMethod), "view"), h.paymentMethodMaster.ListPaymentMethods)
	pm.POST("", h.requirePermission(string(model.ResourcePaymentMethod), "create"), h.paymentMethodMaster.CreatePaymentMethod)
	pm.PATCH("/reorder", h.requirePermission(string(model.ResourcePaymentMethod), "edit"), h.paymentMethodMaster.ReorderPaymentMethods)
	pm.GET("/:id", h.requirePermission(string(model.ResourcePaymentMethod), "view"), h.paymentMethodMaster.GetPaymentMethod)
	pm.PATCH("/:id", h.requirePermission(string(model.ResourcePaymentMethod), "edit"), h.paymentMethodMaster.UpdatePaymentMethod)
	pm.DELETE("/:id", h.requirePermission(string(model.ResourcePaymentMethod), "delete"), h.paymentMethodMaster.DeletePaymentMethod)

	// 見積（旧 handler.go 逐語 + TASK-012 successors）
	estimates := rg.Group("/estimates")
	estimates.GET("", h.requirePermission(string(model.ResourceEstimates), "view"), h.estimate.ListEstimates)
	estimates.GET("/:id", h.requirePermission(string(model.ResourceEstimates), "view"), h.estimate.GetEstimate)
	estimates.POST("", h.requirePermission(string(model.ResourceEstimates), "create"), h.estimate.CreateEstimate)
	estimates.POST("/:id/successors", h.requirePermission(string(model.ResourceEstimates), "create"), h.estimate.CreateEstimateSuccessor)
	estimates.PATCH("/:id", h.requirePermission(string(model.ResourceEstimates), "edit"), h.estimate.UpdateEstimate)
	estimates.DELETE("/:id", h.requirePermission(string(model.ResourceEstimates), "delete"), h.estimate.DeleteEstimate)

	// 会計医師確認（旧 billing_confirmation_handler.go RegisterBillingConfirmationRoutes 逐語・
	// /medical-records group は gin path merge で medicalrecord 側と共存）
	records := rg.Group("/medical-records")
	permEdit := h.requirePermission(string(model.ResourceAccounting), "edit")
	records.GET("/:id/billing-confirmation", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingConfirmation.GetBillingConfirmation)
	records.POST("/:id/billing-confirmation/confirm", permEdit, h.billingConfirmation.ConfirmBillingConfirmation)
	records.POST("/:id/billing-confirmation/return", permEdit, h.billingConfirmation.ReturnBillingConfirmation)

	// 明細（旧 billing_item_handler.go RegisterBillingItemRoutes 逐語）
	items := rg.Group("/billing-items")
	items.GET("/unbilled", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingItem.GetUnbilledItems)
	// BUG-013: additive details envelope (items + typed blocking warnings). Same view permission.
	items.GET("/unbilled-details", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingItem.GetUnbilledItemDetails)
	items.GET("/ungrouped-same-day", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingItem.GetUngroupedSameDay)
	items.POST("", h.requirePermission(string(model.ResourceAccounting), "create"), h.billingItem.CreateBillingItem)
	items.PATCH("/:id", h.requirePermission(string(model.ResourceAccounting), "edit"), h.billingItem.UpdateBillingItem)
	items.DELETE("/:id", h.requirePermission(string(model.ResourceAccounting), "delete"), h.billingItem.DeleteBillingItem)
	items.GET("/:id/discount-suggestions", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingItem.GetBillingItemDiscountSuggestions)

	// 会計（旧 handler.go registerAccountingRoutesWithAuth 逐語）+ 返金
	accountings := rg.Group("/accountings")
	accountings.GET("", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.ListAccountings)
	// BUG-370: 月末未納者一覧
	accountings.GET("/unpaid", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.ListUnpaidBillings)
	// #182: 会計画面表示用 飼主未納残高
	accountings.GET("/unpaid-balance", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.GetOwnerUnpaidBalance)
	// #114: 月次未納繰越集計
	accountings.GET("/unpaid-monthly", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.GetUnpaidMonthlySummary)
	// BUG-368: レジ締め日次集計
	accountings.GET("/daily-summary", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.GetDailySummary)
	accountings.GET("/:id", h.requirePermission(string(model.ResourceAccounting), "view"), h.accounting.GetAccounting)
	accountings.POST("", h.requirePermission(string(model.ResourceAccounting), "create"), h.accounting.CreateAccounting)
	accountings.PATCH("/:id", h.requirePermission(string(model.ResourceAccounting), "edit"), h.accounting.UpdateAccounting)
	// #189: 確定済みカード金額の確定後訂正（専用フロー）。確定/締め後訂正の既存権限 accounting-post-close-edit:edit を再利用。
	accountings.POST("/:id/credit-correction", h.requirePermission(string(model.ResourceAccountingPostCloseEdit), "edit"), h.accounting.CorrectCreditPayment)
	// BUG-371 / #118: DELETE は廃止し論理削除 (POST /:id/cancel) に統合。専用権限 accounting-cancel を使用する。
	accountings.POST("/:id/cancel", h.requirePermission(string(model.ResourceAccountingCancel), "edit"), h.accounting.CancelAccounting)
	// 返金 routes は internal/billing.RegisterRoutes へ移動（BE9-2C B③・/accountings group は gin path merge で共存）
	accountings.GET("/:id/refunds", h.requirePermission(string(model.ResourceAccounting), "view"), h.refund.ListRefunds)
	accountings.POST("/:id/refunds", h.requirePermission(string(model.ResourceAccounting), "create"), h.refund.CreateRefund)

	// レジ締め・月次レポート（旧 cash_register_handler.go / accounting_report_handler.go 逐語）
	cr := rg.Group("/cash-register")
	cr.GET("/preview", h.requirePermission(string(model.ResourceCashRegisterClose), "view"), h.cashRegister.GetCashRegisterPreview)
	cr.POST("/closes", h.requirePermission(string(model.ResourceCashRegisterClose), "create"), h.cashRegister.CloseCashRegister)
	cr.GET("/closes", h.requirePermission(string(model.ResourceCashRegisterClose), "view"), h.cashRegister.ListCashRegisterCloses)
	cr.GET("/closes/:id", h.requirePermission(string(model.ResourceCashRegisterClose), "view"), h.cashRegister.GetCashRegisterClose)

	reports := rg.Group("/reports")
	reports.GET("/monthly", h.requirePermission(string(model.ResourceAccountingReports), "view"), h.accountingReport.GetMonthlyReport)
	reports.GET("/monthly/csv", h.requirePermission(string(model.ResourceAccountingReports), "view"), h.accountingReport.ExportMonthlyCSV)
}
