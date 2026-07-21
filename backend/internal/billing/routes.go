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
	requirePermission   PermissionMiddleware
}

// NewHandler は billing domain の routing composition を構築する。
func NewHandler(
	insurance *InsuranceHandler,
	campaign *CampaignHandler,
	paymentMethodMaster *PaymentMethodMasterHandler,
	estimate *EstimateHandler,
	billingConfirmation *BillingConfirmationHandler,
	requirePermission PermissionMiddleware,
) *Handler {
	return &Handler{
		insurance:           insurance,
		campaign:            campaign,
		paymentMethodMaster: paymentMethodMaster,
		estimate:            estimate,
		billingConfirmation: billingConfirmation,
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

	// 見積（旧 handler.go 逐語）
	estimates := rg.Group("/estimates")
	estimates.GET("", h.requirePermission(string(model.ResourceEstimates), "view"), h.estimate.ListEstimates)
	estimates.GET("/:id", h.requirePermission(string(model.ResourceEstimates), "view"), h.estimate.GetEstimate)
	estimates.POST("", h.requirePermission(string(model.ResourceEstimates), "create"), h.estimate.CreateEstimate)
	estimates.PATCH("/:id", h.requirePermission(string(model.ResourceEstimates), "edit"), h.estimate.UpdateEstimate)
	estimates.DELETE("/:id", h.requirePermission(string(model.ResourceEstimates), "delete"), h.estimate.DeleteEstimate)

	// 会計医師確認（旧 billing_confirmation_handler.go RegisterBillingConfirmationRoutes 逐語・
	// /medical-records group は gin path merge で medicalrecord 側と共存）
	records := rg.Group("/medical-records")
	permEdit := h.requirePermission(string(model.ResourceAccounting), "edit")
	records.GET("/:id/billing-confirmation", h.requirePermission(string(model.ResourceAccounting), "view"), h.billingConfirmation.GetBillingConfirmation)
	records.POST("/:id/billing-confirmation/confirm", permEdit, h.billingConfirmation.ConfirmBillingConfirmation)
	records.POST("/:id/billing-confirmation/return", permEdit, h.billingConfirmation.ReturnBillingConfirmation)
}
