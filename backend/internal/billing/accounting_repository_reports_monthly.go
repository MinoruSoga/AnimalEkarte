package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// GetMonthlyReport は指定年月の月次売上レポートを集計する。FEAT-368
// 集計は GetMonthlyReportByPeriod に委ね、その reader は未接続のため固定エラー（422）を返す。held for EMR-509。
func (r *accountingRepository) GetMonthlyReport(ctx context.Context, clinicID uint64, year, month int) (*MonthlyReportResult, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	return r.GetMonthlyReportByPeriod(ctx, clinicID, start, end)
}

// GetMonthlyReportByPeriod は指定期間（start 以上 end 未満）の売上レポートを集計していた。held for EMR-509。
// 旧実装は completed_at 基準の billings・payment_splits・billing_refunds と、会計ごとの税の丸めに基づく税率別集計を
// 読んでいたが、基準が実際の受領日（payment_funds.received_at）と確定イベントに変わり、税の集計も請求書単位の
// 税率ごとの切り捨て（D5）と合わない。reader が未接続のため固定エラー（422）で拒否する。
func (*accountingRepository) GetMonthlyReportByPeriod(_ context.Context, _ uint64, _, _ time.Time) (*MonthlyReportResult, error) {
	return nil, apperrors.WrapFinanceConsumerPending("月次の売上レポートは、受領・返金・確定の記録（finance_source_events）の読取が未接続のため利用できません")
}
