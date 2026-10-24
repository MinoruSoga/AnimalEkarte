package staff

import (
	"fmt"

	"gorm.io/gorm"
)

// financeStaffReference は staffs(id) を ON DELETE RESTRICT で参照する財務系の表と列（001_init.sql の最終形）。
// 旧 payments.paid_by・billing_refunds.refunded_by は表ごと廃止され、受領・配賦・返金・会計の変更履歴・医師確認・
// 見積・レジ関連の記録へ置き換わった（EMR-549 unit 4 / inventory §1.4）。スタッフの論理削除は RESTRICT を迂回するため、
// 削除ガード（Delete）と依存件数（CountBlockingReferencesByStaffID）が同じ一覧を使い、1 件でも参照があれば削除を拒否する。
// いずれの表も clinic_id を持ち、deleted_at は持たない（追記型）。
// 範囲外: 財務以外で staffs を RESTRICT 参照する表（在庫の数量基準・移動イベント、ワクチン証明、入院の移動など）は、
// それぞれの所有者（510 / 499 など）の一覧で扱う。
type financeStaffReference struct {
	table  string
	column string
	label  string
}

var financeStaffReferences = []financeStaffReference{
	{table: "payment_funds", column: "recorded_by", label: "入金記録"},
	{table: "fund_consumption_events", column: "recorded_by", label: "入金の配賦・返金記録"},
	{table: "fund_record_corrections", column: "recorded_by", label: "入金記録の訂正"},
	{table: "owner_deposit_operations", column: "performed_by", label: "預り金の操作"},
	{table: "billing_change_events", column: "recorded_by", label: "会計の変更履歴"},
	{table: "billing_price_acknowledgements", column: "acknowledged_by", label: "会計の金額確認"},
	{table: "billing_fulfillment_events", column: "recorded_by", label: "実施・準備の記録"},
	{table: "billing_source_difference_resolutions", column: "resolved_by", label: "請求元の差分の解消"},
	{table: "doctor_review_sets", column: "reviewed_by", label: "会計の医師確認"},
	{table: "estimate_presentations", column: "presented_by", label: "見積の提示"},
	{table: "estimate_replies", column: "recorded_by", label: "見積の回答記録"},
	{table: "estimate_calculation_revisions", column: "recorded_by", label: "見積の算定記録"},
	{table: "financial_document_editions", column: "issued_by", label: "請求書・領収書の発行"},
	{table: "cash_adjustment_events", column: "recorded_by", label: "レジ現金の補充・回収"},
	{table: "cash_opening_balances", column: "confirmed_by", label: "レジの開始残高"},
	{table: "cash_actual_corrections", column: "corrected_by", label: "レジ実査の訂正"},
	{table: "finance_source_events", column: "recorded_by", label: "会計イベント"},
}

// notReferencedByFinanceStaffColumns は staffs を更新する文に、財務系の参照が 1 件もない条件を足す Scope。
// 表名・列名は上の固定一覧のみで、外部入力は含まない。
func notReferencedByFinanceStaffColumns(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, ref := range financeStaffReferences {
			db = db.Where(fmt.Sprintf(`NOT EXISTS (
				SELECT 1 FROM %[1]s
				WHERE %[1]s.%[2]s = staffs.id
				  AND %[1]s.clinic_id = ?
			)`, ref.table, ref.column), clinicID)
		}
		return db
	}
}
