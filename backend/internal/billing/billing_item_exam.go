package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ValidateExamCreateReference は検査由来の請求明細の作成前検証だった。
// 検査の請求元（billing_source_bindings.examination_id）と固定値の writer が未接続（EMR-481 / EMR-505 Q7）のため、
// 検査由来の明細作成は service で書き込み前に拒否している。BillingItemRepository の契約と lintscan の inventory のために残し、
// 呼ばれても成功を装わず同じ固定エラー（422 finance_consumer_pending）を返す。held for EMR-481。
func (*billingItemRepository) ValidateExamCreateReference(_ context.Context, _, _, _ uint64) error {
	return apperrors.WrapFinanceConsumerPending(billingSourceItemPendingMessage)
}
