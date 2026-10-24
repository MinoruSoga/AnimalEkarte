package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// 飼主の売上・来院金額の集計（累計 LTV・1回来院の最大額・直近365日の売上上位）は、held for EMR-506。
// 基準は請求確定額（税込。受領・返金とは別建て）に決まった（EMR-549 comment 0b600d6a / PO-NO-LEGACY-COMPAT 8823f814）。
// 保存先（billings.claim_state='finalized' と current_calculation_id が指す billing_calculation_snapshots.total_amount、
// 日付は確定イベント finance_source_events.occurred_at）は 001_init.sql にあるが、それを書く確定 writer
// （POST /accountings/complete）が未接続で finalized の行が存在しない。この状態で読むと全飼主 0 円・0 件を返し、
// 実績のない値を真として配る（zero-fill）ため、固定エラー（422 finance_consumer_pending）で拒否する。
// 呼出し元（lstep のタグ同期）はエラーを伝播するため、0 に丸められず同期の失敗として可視化される。
// 受領額（payment_funds）や返金からの代替集計は作らない。順位づけ（FindOwnersByAnnualRevenue）は今期の対象外で、基準が揃っても 422 のまま。
// resume: 確定 writer が接続され、確定額の reader を実装したとき（順位づけを除く）。
const (
	ltvPendingMessage                   = "飼主の累計売上・1回来院の最大額は、請求確定額（確定時の算定固定と確定日）の読取が未接続のため利用できません"
	annualRevenueRankingPendingMessage = "飼主の直近365日売上の順位づけは、今期の対象外のため利用できません"
)

// SumPaidByOwner は飼い主の累計売上（LTV）を返す（Lステップタグ同期用）。メソッド名は lstep の依存面のため維持する。
func (*accountingRepository) SumPaidByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, apperrors.WrapFinanceConsumerPending(ltvPendingMessage)
}

// MaxSingleVisitAmountByOwner は飼い主の1回来院最大額を返す（CPMスポット判定用）。
func (*accountingRepository) MaxSingleVisitAmountByOwner(_ context.Context, _, _ uint64) (int64, error) {
	return 0, apperrors.WrapFinanceConsumerPending(ltvPendingMessage)
}

// OwnerAnnualRevenue は飼い主ごとの直近365日売上集計。
type OwnerAnnualRevenue struct {
	OwnerID uint64
	Revenue int64
}

// FindOwnersByAnnualRevenue は直近365日の売上上位の飼い主を返していた（LTV_上位20 タグ用）。順位づけは今期の対象外。
func (*accountingRepository) FindOwnersByAnnualRevenue(_ context.Context, _ uint64) ([]OwnerAnnualRevenue, error) {
	return nil, apperrors.WrapFinanceConsumerPending(annualRevenueRankingPendingMessage)
}
