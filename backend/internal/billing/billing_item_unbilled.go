package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	UnbilledWarningSourceExam                   = "exam"
	UnbilledWarningCodeExamTypeMasterUnbillable = "exam_type_master_unbillable"
)

// 未請求候補の集約（診療・予防接種・検査・トリミング）、集約の版（revision）の照合、同日の未会計対象化の件数は、
// 請求元の固定値 projection と billing_source_bindings が未接続（EMR-481）のため、推測も書き込みもせず、
// 固定エラー（422 finance_consumer_pending）で拒否する（EMR-505 / Q7）。旧実装は現在の master の税率や当日の
// campaign を読む projection で、未知の税情報を既定値で補う経路（D11 が禁じる）だった。
// resume: EMR-481 の treatment subset（treatments.billing_source_projection の DDL を含む）が merge されたとき。
const unbilledPendingMessage = "未請求の診療・予防接種・検査・トリミングの候補は、請求元の紐付け（billing_source_bindings）と固定値の projection が未接続のため取得できません"

func unbilledPendingError() error {
	return apperrors.WrapFinanceConsumerPending(unbilledPendingMessage)
}

// GetUnbilledItems は GET /billing-items/unbilled（legacy raw-array 契約）。held for EMR-481。
func (*billingItemService) GetUnbilledItems(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, unbilledPendingError()
}

// GetUnbilledItemDetails は GET /billing-items/unbilled-details。held for EMR-481。
func (*billingItemService) GetUnbilledItemDetails(_ context.Context, _, _ uint64) (*UnbilledDetails, error) {
	return nil, unbilledPendingError()
}

// AssertNoBlockingUnbilled は blocking な未請求警告（BUG-013）の有無を集約から判定していた。集約が未接続のため判定できず、
// 通過させる（成功を装う）代わりに固定エラーを返す。accounting_service_core.go:67 の請求作成ガードも同じ結果になる
// （Lead 判断事項。U4-506 がその呼出しを外すか置き換えるまで、pet に紐付く請求ヘッダは作成できない）。
func (*billingItemService) AssertNoBlockingUnbilled(_ context.Context, _, _ uint64) error {
	return unbilledPendingError()
}

// AssertUnbilledForComplete は complete 確定時の未請求集約の版（expected_unbilled_revision）の照合（EMR-196②）。
// 集約が未接続のため照合できず、固定エラーを返す。complete 自体も 462/506 が保留している。
func (*billingItemService) AssertUnbilledForComplete(_ context.Context, _, _ uint64, _ string) error {
	return unbilledPendingError()
}

// GetUngroupedSameDaySummary は同日同ペットの未会計対象化項目の件数（#77）。診察側の件数が billing_confirmations と
// billings.status を読む SQL（medicalrecord/treatment_repository.go）で廃止列・表に依存するため、サマリー全体を保留する。
func (*billingItemService) GetUngroupedSameDaySummary(_ context.Context, _, _ uint64, _ time.Time) (UngroupedSameDaySummary, error) {
	return UngroupedSameDaySummary{}, unbilledPendingError()
}
