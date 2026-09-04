package billing

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// lastHandlerSegment は internal/medicalrecord の同名ヘルパーと同一実装。
func lastHandlerSegment(fullName string) string {
	name := strings.TrimSuffix(fullName, "-fm")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}

// TestRegisterRoutes_Snapshot は billing 側の BE9-2C route-snapshot 回帰チェック
// （manualarticle/medicalrecord/reservation の先例を踏襲）。B①〜B⑤ で
// internal/handler/testdata/route_snapshot.golden から 50 route（insurances 6 +
// campaigns 6 + payment-methods 6+見積5+会計医師確認3+明細6+返金2+会計10）を drop し、本 package の RegisterRoutes が登録する。
func TestRegisterRoutes_Snapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	noopPermission := func(_, _ string) gin.HandlerFunc {
		return func(c *gin.Context) {}
	}
	h := NewHandler(
		NewInsuranceHandler(nil),
		NewCampaignHandler(nil),
		NewPaymentMethodMasterHandler(nil),
		NewEstimateHandler(nil, func(_ *gin.Context, _, _ string) bool { return true }),
		NewBillingConfirmationHandler(nil, noopPermission),
		NewBillingItemHandler(nil, nil, nil, noopPermission),
		NewRefundHandler(nil, noopPermission),
		NewAccountingHandler(nil, nil, func(_ *gin.Context, _, _ string) bool { return true }),
		NewCashRegisterHandler(nil, noopPermission),
		NewAccountingReportHandler(nil, noopPermission),
		noopPermission,
	)

	r := gin.New()
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	lines := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		lines = append(lines, fmt.Sprintf("%s %s %s", route.Method, route.Path, lastHandlerSegment(route.Handler)))
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"

	want := "" +
		"DELETE /api/v1/billing-items/:id DeleteBillingItem\n" +
		"DELETE /api/v1/estimates/:id DeleteEstimate\n" +
		"DELETE /api/v1/masters/campaigns/:id DeleteCampaign\n" +
		"DELETE /api/v1/masters/insurances/:id DeleteInsurance\n" +
		"DELETE /api/v1/payment-methods/:id DeletePaymentMethod\n" +
		"GET /api/v1/accountings ListAccountings\n" +
		"GET /api/v1/accountings/:id GetAccounting\n" +
		"GET /api/v1/accountings/:id/refunds ListRefunds\n" +
		"GET /api/v1/accountings/daily-summary GetDailySummary\n" +
		"GET /api/v1/accountings/unpaid ListUnpaidBillings\n" +
		"GET /api/v1/accountings/unpaid-balance GetOwnerUnpaidBalance\n" +
		"GET /api/v1/accountings/unpaid-monthly GetUnpaidMonthlySummary\n" +
		"GET /api/v1/billing-items/:id/discount-suggestions GetBillingItemDiscountSuggestions\n" +
		"GET /api/v1/billing-items/unbilled GetUnbilledItems\n" +
		"GET /api/v1/billing-items/unbilled-details GetUnbilledItemDetails\n" +
		"GET /api/v1/billing-items/ungrouped-same-day GetUngroupedSameDay\n" +
		"GET /api/v1/cash-register/closes ListCashRegisterCloses\n" +
		"GET /api/v1/cash-register/closes/:id GetCashRegisterClose\n" +
		"GET /api/v1/cash-register/preview GetCashRegisterPreview\n" +
		"GET /api/v1/estimates ListEstimates\n" +
		"GET /api/v1/estimates/:id GetEstimate\n" +
		"GET /api/v1/masters/campaigns ListCampaigns\n" +
		"GET /api/v1/masters/campaigns/:id GetCampaign\n" +
		"GET /api/v1/masters/insurances ListInsurances\n" +
		"GET /api/v1/masters/insurances/:id GetInsurance\n" +
		"GET /api/v1/medical-records/:id/billing-confirmation GetBillingConfirmation\n" +
		"GET /api/v1/payment-methods ListPaymentMethods\n" +
		"GET /api/v1/payment-methods/:id GetPaymentMethod\n" +
		"GET /api/v1/reports/monthly GetMonthlyReport\n" +
		"GET /api/v1/reports/monthly/csv ExportMonthlyCSV\n" +
		"PATCH /api/v1/accountings/:id UpdateAccounting\n" +
		"PATCH /api/v1/billing-items/:id UpdateBillingItem\n" +
		"PATCH /api/v1/estimates/:id UpdateEstimate\n" +
		"PATCH /api/v1/masters/campaigns/:id UpdateCampaign\n" +
		"PATCH /api/v1/masters/campaigns/reorder ReorderCampaigns\n" +
		"PATCH /api/v1/masters/insurances/:id UpdateInsurance\n" +
		"PATCH /api/v1/masters/insurances/reorder ReorderInsurances\n" +
		"PATCH /api/v1/payment-methods/:id UpdatePaymentMethod\n" +
		"PATCH /api/v1/payment-methods/reorder ReorderPaymentMethods\n" +
		"POST /api/v1/accountings CreateAccounting\n" +
		"POST /api/v1/accountings/:id/cancel CancelAccounting\n" +
		"POST /api/v1/accountings/:id/credit-correction CorrectCreditPayment\n" +
		"POST /api/v1/accountings/:id/refunds CreateRefund\n" +
		"POST /api/v1/accountings/complete CompleteAccounting\n" +
		"POST /api/v1/billing-items CreateBillingItem\n" +
		"POST /api/v1/cash-register/closes CloseCashRegister\n" +

		"POST /api/v1/estimates CreateEstimate\n" +
		"POST /api/v1/estimates/:id/successors CreateEstimateSuccessor\n" +
		"POST /api/v1/masters/campaigns CreateCampaign\n" +
		"POST /api/v1/masters/insurances CreateInsurance\n" +
		"POST /api/v1/medical-records/:id/billing-confirmation/confirm ConfirmBillingConfirmation\n" +
		"POST /api/v1/medical-records/:id/billing-confirmation/return ReturnBillingConfirmation\n" +
		"POST /api/v1/payment-methods CreatePaymentMethod\n"

	assert.Equal(t, want, got)
}
