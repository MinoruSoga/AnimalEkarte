package billing

import (
	"time"
)

// accounting_reports_dto.go — B④: repository/accounting_repository_reports.go（B⑤残留）から
// 会計サービスが参照する DTO を先行移動（repository 側は alias）。

// DailySummaryResult はレジ締め日次集計の結果。BUG-368
type DailySummaryResult struct {
	PaymentTotals  []PaymentMethodTotal `json:"payment_totals"`
	CategoryTotals []CategoryTotal      `json:"category_totals"`
	BillingCount   int64                `json:"billing_count"`
	GrandTotal     int64                `json:"grand_total"`
}

// PaymentMethodTotal は支払方法別の売上合計。BUG-368
type PaymentMethodTotal struct {
	Method string `json:"method"`
	Total  int64  `json:"total"`
}

// CategoryTotal は診療区分別の売上合計。BUG-368
type CategoryTotal struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

// GetCloseAggregateInput は締めプレビュー集計の入力パラメータ
type GetCloseAggregateInput struct {
	ClinicID    uint64
	PeriodStart time.Time // JST timestamptz
	PeriodEnd   time.Time // JST timestamptz
}

// CloseAggregateResult は締めプレビュー集計の結果
type CloseAggregateResult struct {
	PaymentRows    []PaymentAggregateRow
	CategoryRows   []CategoryAggregateRow
	TotalRefund    int64
	BillingDetails []CloseBillingDetailRow
	TaxBreakdown   []TaxBreakdownRow `json:"-"`
	// UnclassifiedOtherCount は category=other 明細を1件以上持つ会計の distinct 件数（DEC-40）。
	// billing_items 行数や MIN(category) 明細行数とは独立に集計する。
	UnclassifiedOtherCount int64
}

// MonthlyReportResult は月次売上レポートの結果
type MonthlyReportResult struct {
	PaymentRows       []MonthlyPaymentRow
	CategoryRows      []MonthlyCategoryRow
	DailyBillingCount map[string]int64
	ClosedAM          map[string]bool
	ClosedPM          map[string]bool
	GrandTotal        int64
	TotalRefund       int64
	BillingCount      int64
	TaxBreakdown      []TaxBreakdownRow `json:"-"`
}

// PaymentAggregateRow は支払方法別の集計結果1行
type PaymentAggregateRow struct {
	PaymentMethodID *uint64
	Amount          int64
}

// CategoryAggregateRow はカテゴリ別の集計結果1行
type CategoryAggregateRow struct {
	Category string
	Amount   int64
}

// CloseBillingDetailRow は個別会計一覧の1レコード
type CloseBillingDetailRow struct {
	BillingID         uint64    `json:"billing_id"`
	PaidAt            time.Time `json:"paid_at"`
	OwnerName         string    `json:"owner_name"`
	PetName           string    `json:"pet_name"`
	IsHospitalization bool      `json:"is_hospitalization"`
	Category          string    `json:"category"`
	PaymentMethodID   *uint64   `json:"payment_method_id,omitempty"`
	BillingAmount     int64     `json:"billing_amount"`
	RefundAmount      int64     `json:"refund_amount"`
	NetAmount         int64     `json:"net_amount"`
}

// TaxBreakdownRow は税率別の集計結果1行
type TaxBreakdownRow struct {
	TaxRate       int64 // 整数パーセント: 10 (標準) or 8 (軽減)
	TaxableAmount int64
	TaxAmount     int64
}

// MonthlyPaymentRow は月次レポートの日×支払方法別集計行
type MonthlyPaymentRow struct {
	Date            string
	PaymentMethodID *uint64
	Amount          int64
}

// MonthlyCategoryRow は月次レポートの日×カテゴリ別集計行
type MonthlyCategoryRow struct {
	Date     string
	Category string
	Amount   int64
}
