package repository

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SumPaidByOwner は飼い主の支払済み請求合計（LTV）を返す（Lステップタグ同期用）。
func (r *accountingRepository) SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
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
		Scopes(clinicScope(clinicID)).
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

// FindOwnersByAnnualRevenue は直近365日の完了済み請求額合計を飼い主ごとに集計し、降順で返す（LTV上位％判定用）。
func (r *accountingRepository) FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]OwnerAnnualRevenue, error) {
	cutoff := time.Now().AddDate(0, 0, -365)
	var results []OwnerAnnualRevenue
	err := r.db.WithContext(ctx).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("status = ? AND deleted_at IS NULL AND completed_at >= ?", model.BillingStatusCompleted, cutoff).
		Select("owner_id, COALESCE(SUM(total_amount), 0) AS revenue").
		Group("owner_id").
		Order("revenue DESC").
		Scan(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing", fmt.Sprintf("clinic=%d", clinicID))
	}
	return results, nil
}
