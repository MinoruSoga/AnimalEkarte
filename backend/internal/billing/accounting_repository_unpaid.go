package billing

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MonthlyUnpaidOwnerPet は飼主+ペット単位の月次未納繰越集約結果。#114
type MonthlyUnpaidOwnerPet struct {
	OwnerID            uint64  `json:"owner_id"`
	OwnerName          string  `json:"owner_name"`
	PetID              *uint64 `json:"pet_id,omitempty"`
	PetName            string  `json:"pet_name"`
	PrevMonthCarryover int64   `json:"prev_month_carryover"`
	CurrentMonthUnpaid int64   `json:"current_month_unpaid"`
	NextMonthCarryover int64   `json:"next_month_carryover"`
}

// MonthlyUnpaidSummary は月次未納繰越のサマリー情報。#114
type MonthlyUnpaidSummary struct {
	PrevMonthCarryover int64 `json:"prev_month_carryover"`
	CurrentMonthUnpaid int64 `json:"current_month_unpaid"`
	NextMonthCarryover int64 `json:"next_month_carryover"`
}

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

// OwnerUnpaidBalance は飼主単位の未納残高（未入金 billing の合計と件数）。#182
type OwnerUnpaidBalance struct {
	TotalAmount int64 `json:"total_amount"`
	Count       int64 `json:"count"`
}

// validBillingOwnerPetScope excludes a billing whose optional owner or pet
// relation crosses a clinic boundary. NULL owner/pet values remain valid for
// direct billings.
// DEC-27: pets.owner_id is the current owner; billings.owner_id is a snapshot
// at billing time. Do not require equality after pet transfer.
func validBillingOwnerPetScope(db *gorm.DB) *gorm.DB {
	return db.Where(`
		(
			billings.owner_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM owners AS billing_owner
				WHERE billing_owner.id = billings.owner_id
				  AND billing_owner.clinic_id = billings.clinic_id
				  AND billing_owner.deleted_at IS NULL
			)
		)
		AND (
			billings.pet_id IS NULL
			OR EXISTS (
				SELECT 1
				FROM pets AS billing_pet
				WHERE billing_pet.id = billings.pet_id
				  AND billing_pet.clinic_id = billings.clinic_id
				  AND billing_pet.deleted_at IS NULL
			)
		)
	`)
}

// whereUnpaidBalancePositive は unpaidAmountSQL > 0 の行に絞る（BUG-007 含む未収定義）。
func whereUnpaidBalancePositive(db *gorm.DB) *gorm.DB {
	return db.Where(fmt.Sprintf("(%s) > 0", unpaidAmountSQL))
}

// attachOutstandingAmounts は Preload 済み Payments から OutstandingAmount を埋める。
func attachOutstandingAmounts(billings []model.Billing) {
	for i := range billings {
		billings[i].OutstandingAmount = OutstandingAmount(&billings[i])
	}
}

// FindUnpaidByBilling は未収残高 > 0 の billings を会計単位で返す。#120 / BUG-007
// waiting 全額に加え、クレジット訂正由来の completed 差額も含む。
func (r *accountingRepository) FindUnpaidByBilling(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]model.Billing, int64, error) {
	billings := make([]model.Billing, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{}).
		Joins(leftJoinPaymentsSQL).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Scopes(validBillingOwnerPetScope).
		Scopes(whereUnpaidBalancePositive).
		Where("billings.scheduled_date BETWEEN ? AND ?", startDate, endDate)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	if err := q.Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items", "deleted_at IS NULL").
		Preload("Payments", "deleted_at IS NULL").
		Scopes(persistence.Paginate(page, limit)).
		Order("billings.scheduled_date ASC, billings.created_at ASC").
		Find(&billings).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "billing", "")
	}
	attachOutstandingAmounts(billings)
	return billings, total, nil
}

// FindUnpaidByOwner は未収を飼主単位で集約する。#120 / BUG-007
// GROUP BY owner_id で 1 クエリで取得（N+1 回避）。金額は unpaidAmountSQL 合計。
func (r *accountingRepository) FindUnpaidByOwner(ctx context.Context, clinicID uint64, startDate, endDate string, page, limit int) ([]UnpaidOwnerAggregate, int64, UnpaidSummary, error) {
	aggregates := make([]UnpaidOwnerAggregate, 0)
	var totalOwners int64
	var summary UnpaidSummary

	base := r.db.WithContext(ctx).
		Table("billings").
		Joins(
			"JOIN owners ON owners.id = billings.owner_id"+
				" AND owners.clinic_id = billings.clinic_id"+
				" AND owners.deleted_at IS NULL",
		).
		Joins(leftJoinPaymentsSQL).
		Scopes(validBillingOwnerPetScope).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Scopes(whereUnpaidBalancePositive).
		Where("billings.scheduled_date BETWEEN ? AND ?", startDate, endDate)

	// サマリー取得（売掛金総額・件数・飼主数）— 金額は未収残高定義
	if err := base.Session(&gorm.Session{}).
		Select(fmt.Sprintf(
			"COALESCE(SUM(%s), 0) AS total_amount, COUNT(billings.id) AS billing_count, COUNT(DISTINCT billings.owner_id) AS owner_count",
			unpaidAmountSQL,
		)).
		Scan(&summary).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	totalOwners = summary.OwnerCount

	// 飼主単位集約
	if err := base.Session(&gorm.Session{}).
		Select(fmt.Sprintf(
			"billings.owner_id AS owner_id, owners.name AS owner_name, COUNT(billings.id) AS count, COALESCE(SUM(%s), 0) AS total_amount, MIN(billings.scheduled_date)::text AS oldest_scheduled, MAX(billings.scheduled_date)::text AS latest_scheduled",
			unpaidAmountSQL,
		)).
		Group("billings.owner_id, owners.name").
		Order("oldest_scheduled ASC").
		Scopes(persistence.Paginate(page, limit)).
		Scan(&aggregates).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}
	return aggregates, totalOwners, summary, nil
}

// SumUnpaidByOwner は指定飼主の未納残高（未収残高合計と件数）を返す。#182 / BUG-007
// 既存の未納集計（#120/#114）と同一定義で算出し、
// 会計画面の残高表示と未納者一覧/繰越集計の値が乖離しないようにする。
func (r *accountingRepository) SumUnpaidByOwner(ctx context.Context, clinicID, ownerID uint64) (OwnerUnpaidBalance, error) {
	var result OwnerUnpaidBalance
	if err := r.db.WithContext(ctx).
		Table("billings").
		Joins(leftJoinPaymentsSQL).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Scopes(validBillingOwnerPetScope).
		Where("billings.owner_id = ?", ownerID).
		Scopes(whereUnpaidBalancePositive).
		Select(fmt.Sprintf(
			"COALESCE(SUM(%s), 0) AS total_amount, COUNT(billings.id) AS count",
			unpaidAmountSQL,
		)).
		Scan(&result).Error; err != nil {
		return OwnerUnpaidBalance{}, apperrors.FromGORM(err, "billing", "")
	}
	return result, nil
}

// FindMonthlyUnpaidCarryover は対象月の未納繰越（前月繰越・当月未払い・次月繰越）を
// 飼主+ペット単位で返す。#114 / BUG-007
// firstDay: YYYY-MM-01, lastDay: YYYY-MM-DD（月末）
func (r *accountingRepository) FindMonthlyUnpaidCarryover(ctx context.Context, clinicID uint64, firstDay, lastDay string, page, limit int) ([]MonthlyUnpaidOwnerPet, int64, MonthlyUnpaidSummary, error) {
	items := make([]MonthlyUnpaidOwnerPet, 0)
	var summary MonthlyUnpaidSummary
	var total int64

	// 未収残高 > 0 かつ scheduled_date <= lastDay が集計対象。
	// 前月繰越(< firstDay) + 当月未払い(firstDay〜lastDay) = 次月繰越(<= lastDay)。
	base := r.db.WithContext(ctx).
		Table("billings").
		Joins(
			"JOIN owners ON owners.id = billings.owner_id"+
				" AND owners.clinic_id = billings.clinic_id"+
				" AND owners.deleted_at IS NULL",
		).
		Joins(leftJoinPaymentsSQL).
		Scopes(validBillingOwnerPetScope).
		Where("billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Scopes(whereUnpaidBalancePositive).
		Where("billings.scheduled_date <= ?", lastDay)

	amt := unpaidAmountSQL

	// サマリー取得（3列一括 CASE WHEN）
	if err := base.Session(&gorm.Session{}).
		Select(fmt.Sprintf(`
			COALESCE(SUM(CASE WHEN billings.scheduled_date < ? THEN (%s) ELSE 0 END), 0) AS prev_month_carryover,
			COALESCE(SUM(CASE WHEN billings.scheduled_date >= ? AND billings.scheduled_date <= ? THEN (%s) ELSE 0 END), 0) AS current_month_unpaid,
			COALESCE(SUM(%s), 0) AS next_month_carryover
		`, amt, amt, amt), firstDay, firstDay, lastDay).
		Scan(&summary).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}

	// ページネーション用件数（飼主+ペットの組み合わせ数）
	if err := r.db.WithContext(ctx).Raw(fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM billings
			JOIN owners
			  ON owners.id = billings.owner_id
			 AND owners.clinic_id = billings.clinic_id
			 AND owners.deleted_at IS NULL
			LEFT JOIN payments ON payments.billing_id = billings.id AND payments.deleted_at IS NULL
			WHERE billings.clinic_id = ? AND billings.deleted_at IS NULL
			  AND billings.scheduled_date <= ?
			  AND (%s) > 0
			  AND (
				billings.owner_id IS NULL
				OR EXISTS (
					SELECT 1 FROM owners AS billing_owner
					WHERE billing_owner.id = billings.owner_id
					  AND billing_owner.clinic_id = billings.clinic_id
					  AND billing_owner.deleted_at IS NULL
				)
			  )
			  AND (
				billings.pet_id IS NULL
				OR EXISTS (
					SELECT 1 FROM pets AS billing_pet
					WHERE billing_pet.id = billings.pet_id
					  AND billing_pet.clinic_id = billings.clinic_id
					  AND billing_pet.deleted_at IS NULL
				)
			  )
			GROUP BY billings.owner_id, billings.pet_id
		) sub
	`, unpaidAmountSQL), clinicID, lastDay).Scan(&total).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}

	// 飼主+ペット単位集約（CASE WHEN で3列同時集計）。
	// DEC-27: join pets by identity + clinic only; do not require pets.owner_id
	// = billings.owner_id (snapshot vs current owner after transfer).
	if err := base.Session(&gorm.Session{}).
		Joins(
			"LEFT JOIN pets ON pets.id = billings.pet_id"+
				" AND pets.clinic_id = billings.clinic_id"+
				" AND pets.deleted_at IS NULL",
		).
		Select(fmt.Sprintf(`
			billings.owner_id AS owner_id,
			owners.name AS owner_name,
			pets.id AS pet_id,
			COALESCE(pets.name, '') AS pet_name,
			COALESCE(SUM(CASE WHEN billings.scheduled_date < ? THEN (%s) ELSE 0 END), 0) AS prev_month_carryover,
			COALESCE(SUM(CASE WHEN billings.scheduled_date >= ? AND billings.scheduled_date <= ? THEN (%s) ELSE 0 END), 0) AS current_month_unpaid,
			COALESCE(SUM(%s), 0) AS next_month_carryover
		`, amt, amt, amt), firstDay, firstDay, lastDay).
		Group("billings.owner_id, owners.name, pets.id, COALESCE(pets.name, '')").
		Order("owners.name ASC, COALESCE(pets.name, '') ASC").
		Scopes(persistence.Paginate(page, limit)).
		Scan(&items).Error; err != nil {
		return nil, 0, summary, apperrors.FromGORM(err, "billing", "")
	}

	return items, total, summary, nil
}
