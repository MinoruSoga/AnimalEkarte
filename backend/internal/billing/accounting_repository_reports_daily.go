package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// GetDailySummary は指定日（JST）の会計完了分を集計していた。BUG-368 / held for EMR-509。
// 旧実装は billings.status='completed'・completed_at と payment_splits を正として集計していたが、いずれも廃止された
// （001_init.sql:10889,10919）。日次の売上は実際の受領日（payment_funds.received_at）と確定イベント、返金は
// fund_consumption_events を基準にする再設計で、その reader が未接続のため固定エラー（422）で拒否する。
// resume: finance_source_events の日次 reader が実装されたとき。
func (*accountingRepository) GetDailySummary(_ context.Context, _ uint64, _ time.Time) (*DailySummaryResult, error) {
	return nil, apperrors.WrapFinanceConsumerPending("日次の売上集計は、受領・返金・確定の記録（finance_source_events）の読取が未接続のため利用できません")
}
