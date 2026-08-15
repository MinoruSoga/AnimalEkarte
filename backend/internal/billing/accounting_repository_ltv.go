package billing

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// validBillingOwnerMedicalRecordScope keeps direct billings, but excludes a
// billing whose optional medical record is cross-clinic or soft-deleted.
// DEC-27: medical_records.owner_id and billings.owner_id are independent
// snapshots; do not require equality (pet transfer must not drop LTV rows).
func validBillingOwnerMedicalRecordScope(db *gorm.DB) *gorm.DB {
	return db.Where(`billings.medical_record_id IS NULL OR EXISTS (
		SELECT 1
		FROM medical_records AS mr
		WHERE mr.id = billings.medical_record_id
		  AND mr.clinic_id = billings.clinic_id
		  AND mr.deleted_at IS NULL
	)`)
}

// SumPaidByOwner は飼い主の支払済み請求合計（LTV）を返す（Lステップタグ同期用）。
func (r *accountingRepository) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Scopes(validBillingOwnerMedicalRecordScope).
		Where("owner_id = ? AND status = ? AND deleted_at IS NULL", ownerID, model.BillingStatusCompleted).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "billing", fmt.Sprintf("owner=%d", ownerID))
	}
	return total, nil
}

// MaxSingleVisitAmountByOwner は飼い主の1回来院最大支払額を返す（CPMスポット判定用）。
func (r *accountingRepository) MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var maxAmount int64
	err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Scopes(validBillingOwnerMedicalRecordScope).
		Where("owner_id = ? AND status = ? AND deleted_at IS NULL", ownerID, model.BillingStatusCompleted).
		Select("COALESCE(MAX(total_amount), 0)").
		Scan(&maxAmount).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "billing", fmt.Sprintf("owner=%d", ownerID))
	}
	return maxAmount, nil
}

// OwnerAnnualRevenue は飼い主ごとの直近365日売上集計。
type OwnerAnnualRevenue struct {
	OwnerID uint64
	Revenue int64
}

// ltvTopPercent は Lステップ LTV_上位20 タグの対象割合（％）。
// topN = ceil(N * ltvTopPercent / 100) = (N * ltvTopPercent + 99) / 100（正の整数除算）。
const ltvTopPercent = 20

// ownerAnnualRevenueTopPercentSQL は直近365日完了会計を現在の clinic owner に集計し、
// 売上降順・owner_id 昇順の確定タイブレークで exact top-percent だけを返す（G2F-03）。
//
// Bound contract:
//   - DB が total_owners から topN を算出し、rn <= topN の行だけを返す
//   - 呼び出し側 Go は返却集合を再スライスせず top 集合として扱う
//   - 全 clinic owner 売上を Go ヒープへ materialize しない
//
// Query plan (EXPLAIN evidence pinned by TestAccountingRepository_FindOwnersByAnnualRevenue_ExplainPlanIsWindowBounded):
// PostgreSQL evaluates the window phase (WindowAgg over the grouped owner_revenue CTE)
// and filters to rn <= topN before the final result is scanned into Go. The returned
// row count is O(ceil(N*20/100)), not O(N).
const ownerAnnualRevenueTopPercentSQL = `
WITH owner_revenue AS (
	SELECT
		billings.owner_id AS owner_id,
		COALESCE(SUM(billings.total_amount), 0) AS revenue
	FROM billings
	INNER JOIN owners
		ON owners.id = billings.owner_id
		AND owners.clinic_id = billings.clinic_id
		AND owners.deleted_at IS NULL
	WHERE billings.clinic_id = ?
		AND billings.status = ?
		AND billings.deleted_at IS NULL
		AND billings.completed_at >= ?
		AND billings.owner_id IS NOT NULL
		AND (
			billings.medical_record_id IS NULL OR EXISTS (
				SELECT 1
				FROM medical_records AS mr
				WHERE mr.id = billings.medical_record_id
					AND mr.clinic_id = billings.clinic_id
					AND mr.deleted_at IS NULL
			)
		)
	GROUP BY billings.owner_id
),
ranked AS (
	SELECT
		owner_id,
		revenue,
		COUNT(*) OVER () AS total_owners,
		ROW_NUMBER() OVER (ORDER BY revenue DESC, owner_id ASC) AS rn
	FROM owner_revenue
)
SELECT owner_id, revenue
FROM ranked
WHERE rn <= ((total_owners * ? + 99) / 100)
ORDER BY revenue DESC, owner_id ASC
`

// FindOwnersByAnnualRevenue は直近365日の完了済み請求額合計を飼い主ごとに集計し、
// LTV 上位 ltvTopPercent% の飼い主だけを降順で返す（bounded top-N contract / LTV上位％判定用）。
//
// セマンティクス:
//   - clinic scope + current non-deleted owners only
//   - completed billings with completed_at within 365 days
//   - deterministic ties: revenue DESC, owner_id ASC
//   - exact topN = (count * 20 + 99) / 100; empty set when no qualifying owners
func (r *accountingRepository) FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]OwnerAnnualRevenue, error) {
	cutoff := time.Now().AddDate(0, 0, -365)
	var results []OwnerAnnualRevenue
	err := r.db.WithContext(ctx).
		Raw(
			ownerAnnualRevenueTopPercentSQL,
			clinicID,
			model.BillingStatusCompleted,
			cutoff,
			ltvTopPercent,
		).
		Scan(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("clinic=%d", clinicID))
	}
	return results, nil
}
