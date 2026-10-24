package billing

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type closeCategoryRow struct {
	Category string
	Amount   int64
}

func validateCloseAggregateCategory(category string) error {
	for _, allowed := range model.AllItemCategories() {
		if category == string(allowed) {
			return nil
		}
	}

	return fmt.Errorf("unknown item category in close aggregate: %q", category)
}

func toCategoryAggregateRows(rows []closeCategoryRow) ([]CategoryAggregateRow, error) {
	for _, row := range rows {
		if err := validateCloseAggregateCategory(row.Category); err != nil {
			return nil, err
		}
	}

	result := make([]CategoryAggregateRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, CategoryAggregateRow(row))
	}
	return result, nil
}

// completedBillingsPetClinicCTE selects completed billings in the close period.
// Parent pets use clinic-correlated LEFT JOIN for lint parent correlation only;
// rows are never filtered out by pet_id parent clinic (SEC-SWEEP-02-BILL-B1a-FIX).
// Foreign pet/owner name blanking is done in the details query outer joins.
// No pets.deleted_at / deceased_at on correlation (historical close totals must stay).
// cols must be billings-qualified (JOIN pets would make bare id/clinic_id ambiguous).
func completedBillingsPetClinicCTE(cols string) string {
	return `WITH completed_billings AS (
		SELECT ` + cols + `
		FROM billings
		LEFT JOIN pets ON pets.id = billings.pet_id AND pets.clinic_id = billings.clinic_id
		WHERE billings.clinic_id = ? AND billings.deleted_at IS NULL AND billings.status = ?
		  AND billings.completed_at >= ?
		  AND billings.completed_at < ?
	)`
}

// GetCloseAggregate は指定期間内の会計を集計していた。FEAT-368 / held for EMR-509。
// 旧実装は completed_at 基準の billings と payment_splits・billing_refunds を読んでいたが、いずれも廃止された。
// 締めの理論現金は、締め境界の版・開始残高・現金の補充回収・受領と返金の記録（finance_source_events）を束ねる
// 締め文脈（D12）から導く再設計で、その reader が未接続のため、空の集計を返さず固定エラー（422）で拒否する。
// scanClose* 系の旧 SQL は、この reader の置換時に 509 が削除または再実装する。
func (*accountingRepository) GetCloseAggregate(_ context.Context, _ GetCloseAggregateInput) (*CloseAggregateResult, error) {
	return nil, apperrors.WrapFinanceConsumerPending("レジ締めの集計は、受領・返金・確定の記録（finance_source_events）と締め境界の版の読取が未接続のため利用できません")
}
