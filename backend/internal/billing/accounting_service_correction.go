package billing

import (
	"context"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// isCardPaymentMethod は訂正フロー（#189）の対象となるカード系支払い手段か判定する。
// クレジットカードと電子マネーはレジ実機（カード端末）での金額入力を伴い、同種の打ち間違いリスクを持つ。
// 現金は #188（お釣り直接上書き）、銀行振込は端末入力を伴わないため本フロー対象外。
func isCardPaymentMethod(m model.PaymentMethod) bool {
	return m == model.PaymentMethodCreditCard || m == model.PaymentMethodElectronicMoney
}

// CorrectCreditPayment は確定済み会計のクレジット（カード）金額を確定後に訂正する（#189）。
// held for EMR-506: 旧実装は payments / payment_splits の金額を書き換えていたが、両表は廃止された。
// 置換先は受領の記録訂正（fund_record_corrections）で、方法・日付の訂正は金銭を動かさない追記になり、
// カード金額そのものの訂正は新しい family として未定義のため、書き込み前に固定エラー（422）で拒否する。
// 入力の形（理由・手段・金額）の不正だけは従来どおり先に 400 を返す。
func (*accountingService) CorrectCreditPayment(_ context.Context, input *CorrectCreditPaymentInput) (*model.Billing, error) {
	// 入力検証（トランザクション外・低コスト）
	if strings.TrimSpace(input.Reason) == "" {
		return nil, apperrors.WrapInvalidInput("訂正理由は必須です")
	}
	if !isCardPaymentMethod(input.Method) {
		return nil, apperrors.WrapInvalidInput("クレジットカードまたは電子マネーの訂正のみ対応しています")
	}
	if input.Amount <= 0 {
		return nil, apperrors.WrapInvalidInput("訂正金額は1円以上でなければなりません")
	}
	return nil, apperrors.WrapFinanceConsumerPending("確定済みカード金額の訂正は、受領記録の訂正の保存先が未接続のため利用できません")
}
