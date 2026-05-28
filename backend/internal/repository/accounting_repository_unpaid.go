package repository

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UnpaidOwnerAggregate は飼主単位の未納集約結果
// BUG-370
type UnpaidOwnerAggregate struct {
	OwnerID         uint64 `json:"owner_id"`
	OwnerName       string `json:"owner_name"`
	Count           int64  `json:"count"`
	TotalAmount     int64  `json:"total_amount"`
	OldestScheduled string `json:"oldest_scheduled"`
	LatestScheduled string `json:"latest_scheduled"`
}

// UnpaidSummary は未納者一覧のサマリー情報（売掛金総額）
// BUG-370
type UnpaidSummary struct {
	TotalAmount  int64 `json:"total_amount"`
	BillingCount int64 `json:"billing_count"`
	OwnerCount   int64 `json:"owner_count"`
}

// FindUnpaidByBilling は未納 (status=waiting かつ scheduled_date < baseDate) の billings を
// 会計単位で返す。BUG-370 AC-5
func (r *accountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).
		Where("status = ?", model.BillingStatusWaiting).
		Where("scheduled_date < ?", baseDate)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner", "deleted_at IS NULL").Preload("Pet", "deleted_at IS NULL").Preload("Items", "deleted_at IS NULL").
		Offset((page - 1) * limit).Limit(limit).
		Order("scheduled_date ASC, created_at ASC").
		Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	return billings, total, nil
}

// FindUnpaidByOwner は未納を飼主単位で集約する。BUG-370 AC-4
// GROUP BY owner_id で 1 クエリで取得（N+1 回避）。
func (r *accountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, baseDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	aggregates := make([]UnpaidOwnerAggregate, 0)
	var totalOwners int64
	var summary UnpaidSummary

	base := r.db.WithContext(ctx).
		Table("billings").
		Joins("JOIN owners ON owners.id = billings.owner_id AND owners.deleted_at IS NULL").
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billings.status = ?", model.BillingStatusWaiting).
		Where("billings.scheduled_date < ?", baseDate)

	// サマリー取得（売掛金総額・件数・飼主数）
	if err := base.Session(&gorm.Session{}).
		Select("COALESCE(SUM(billings.total_amount), 0) AS total_amount, COUNT(billings.id) AS billing_count, COUNT(DISTINCT billings.owner_id) AS owner_count").
		Scan(&summary).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	totalOwners = summary.OwnerCount

	// 飼主単位集約
	if err := base.Session(&gorm.Session{}).
		Select("billings.owner_id AS owner_id, owners.name AS owner_name, COUNT(billings.id) AS count, COALESCE(SUM(billings.total_amount), 0) AS total_amount, MIN(billings.scheduled_date)::text AS oldest_scheduled, MAX(billings.scheduled_date)::text AS latest_scheduled").
		Group("billings.owner_id, owners.name").
		Order("oldest_scheduled ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Scan(&aggregates).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	return aggregates, totalOwners, summary, nil
}
