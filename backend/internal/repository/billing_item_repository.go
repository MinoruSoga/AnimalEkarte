// Package repository provides data access implementations for BillingItem entity.
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// BillingItemRepository は billing_items テーブルの CRUD を担うインターフェース
type BillingItemRepository interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	Create(ctx context.Context, item *model.BillingItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	// HasItemByOwnerSince は指定飼い主の請求アイテムに names いずれかが存在するか返す（FEAT-379）。
	HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// HasFoodPurchaseByOwnerSince は names 指定時は名前で、未指定時は category=food で判定する（FEAT-379）。
	HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	// FindUnbilledTrimmingItemsByPetID は指定ペットの未請求トリミングコース/オプションを返す(#77)。
	FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)
	// CountNonAccountingTrimmingByPetAndDate は同日同ペットの「未会計対象化」トリミング appointment 件数を返す(#77)。
	CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

type billingItemRepository struct{ db *gorm.DB }

// NewBillingItemRepository は BillingItemRepository を初期化して返す
func NewBillingItemRepository(db *gorm.DB) BillingItemRepository {
	return &billingItemRepository{db: db}
}

// BE-refactor.md R1-1 follow-up (go-reviewer指摘・D2と同型): billing_item_service の
// CreateItem/UpdateItem/DeleteItem は Create/Update/Delete + recalculateTotals
// （FindByBillingID + UpdateBillingTotals）を WithTx 内で txCtx 付きで呼ぶ。
// dbOrTx 未参加のままだと SavePaymentSplits と同じ部分コミットが起こりうる
// （明細書込は独立 tx で即コミット、直後の合計再計算のみ失敗すると ambient tx の rollback で
// 明細だけ残り billing.subtotal/tax_total/total_amount と不整合になる）。
// FindByID も Update 後の再読込（txCtx 付き）で呼ばれるため対象に含める。
func (r *billingItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error) {
	var item model.BillingItem
	err := dbOrTx(ctx, r.db).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.id = ?", id).
		First(&item).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *billingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
	items := make([]model.BillingItem, 0)
	if err := dbOrTx(ctx, r.db).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.billing_id = ?", billingID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error; err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", "")
	}
	return items, nil
}

func (r *billingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	if err := dbOrTx(ctx, r.db).Create(item).Error; err != nil {
		return apperrors.FromGORM(err, "billing_item", "")
	}
	return nil
}

func (r *billingItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := dbOrTx(ctx, r.db).
		Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Where("billing_items.id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := dbOrTx(ctx, r.db).
		Joins("JOIN billings ON billings.id = billing_items.billing_id AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
		Delete(&model.BillingItem{}, "billing_items.id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *billingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	result := dbOrTx(ctx, r.db).
		Model(&model.Billing{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", billingID).
		Updates(map[string]any{
			"subtotal":     subtotal,
			"tax_total":    taxTotal,
			"total_amount": totalAmount,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "billing", fmt.Sprintf("%d", billingID))
	}
	// P2: RowsAffected == 0 は clinic scope 外 or soft-delete の可能性（NOT FOUND）
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing", fmt.Sprintf("%d", billingID))
	}
	return nil
}

func (r *billingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if len(names) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.completed_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.name IN ? AND billing_items.deleted_at IS NULL", names).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}

func (r *billingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.BillingItem{}).
		Joins("JOIN billings ON billings.id = billing_items.billing_id").
		Where("billings.clinic_id = ? AND billings.owner_id = ? AND billings.completed_at >= ? AND billings.deleted_at IS NULL", clinicID, ownerID, since).
		Where("billing_items.deleted_at IS NULL")
	if len(names) > 0 {
		q = q.Where("billing_items.name IN ?", names)
	} else {
		q = q.Where("billing_items.category = ?", string(model.ItemCategoryFood))
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic:%d owner:%d", clinicID, ownerID))
	}
	return count > 0, nil
}

func (r *billingItemRepository) FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error) {
	type row struct {
		AppointmentID    uint64
		OriginID         uint64
		Name             string
		UnitPrice        int64
		SortOrder        int
		TrimmingCourseID *uint64
		TrimmingOptionID *uint64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			a.id AS appointment_id,
			tc.id AS origin_id,
			tc.name AS name,
			COALESCE(tc.price, 0)::bigint AS unit_price,
			0 AS sort_order,
			tc.id AS trimming_course_id,
			NULL::bigint AS trimming_option_id
		FROM appointment_trimming_details atd
		JOIN appointments a ON a.id = atd.appointment_id AND a.deleted_at IS NULL
		JOIN reservation_types rt ON rt.id = a.reservation_type_id AND rt.deleted_at IS NULL
		JOIN trimming_courses tc ON tc.id = atd.course_id AND tc.deleted_at IS NULL
		WHERE a.clinic_id = ?
		  AND a.pet_id = ?
		  AND a.status = ?
		  AND rt.category = ?
		  AND COALESCE(tc.price, 0) > 0
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items bi
		      JOIN billings b ON b.id = bi.billing_id AND b.deleted_at IS NULL
		      WHERE bi.appointment_id = a.id
		        AND bi.trimming_course_id = tc.id
		        AND bi.deleted_at IS NULL
		        AND b.status != ?
		  )
		UNION ALL
		SELECT
			a.id AS appointment_id,
			topt.id AS origin_id,
			topt.name AS name,
			COALESCE(topt.price, 0)::bigint AS unit_price,
			100 + COALESCE(ato.sort_order, 0) AS sort_order,
			NULL::bigint AS trimming_course_id,
			topt.id AS trimming_option_id
		FROM appointment_trimming_details atd
		JOIN appointments a ON a.id = atd.appointment_id AND a.deleted_at IS NULL
		JOIN reservation_types rt ON rt.id = a.reservation_type_id AND rt.deleted_at IS NULL
		JOIN appointment_trimming_options ato ON ato.appointment_id = a.id
		JOIN trimming_options topt ON topt.id = ato.option_id AND topt.deleted_at IS NULL
		WHERE a.clinic_id = ?
		  AND a.pet_id = ?
		  AND a.status = ?
		  AND rt.category = ?
		  AND COALESCE(topt.price, 0) > 0
		  AND NOT EXISTS (
		      SELECT 1
		      FROM billing_items bi
		      JOIN billings b ON b.id = bi.billing_id AND b.deleted_at IS NULL
		      WHERE bi.appointment_id = a.id
		        AND bi.trimming_option_id = topt.id
		        AND bi.deleted_at IS NULL
		        AND b.status != ?
		  )
		ORDER BY appointment_id ASC, sort_order ASC, origin_id ASC
	`,
		clinicID, petID, model.ReservationStatusAccounting, model.ReservationTypeCategoryTrimming, model.BillingStatusCancelled,
		clinicID, petID, model.ReservationStatusAccounting, model.ReservationTypeCategoryTrimming, model.BillingStatusCancelled,
	).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "billing_item", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, petID))
	}

	items := make([]model.BillingItem, 0, len(rows))
	for i, row := range rows {
		appointmentID := row.AppointmentID
		items = append(items, model.BillingItem{
			ID:                    uint64(i + 1),
			BillingID:             0,
			Category:              model.ItemCategoryTrimming,
			Name:                  row.Name,
			UnitPrice:             row.UnitPrice,
			Quantity:              1,
			TaxType:               model.TaxTypeExcluded,
			TaxRate:               0.10,
			IsInsuranceApplicable: false,
			Source:                model.ItemSourceTrimming,
			AppointmentID:         &appointmentID,
			TrimmingCourseID:      row.TrimmingCourseID,
			TrimmingOptionID:      row.TrimmingOptionID,
			SortOrder:             row.SortOrder,
		})
	}
	return items, nil
}

// CountNonAccountingTrimmingByPetAndDate は同日同ペットの「未会計対象化」トリミング appointment 件数を返す(#77)。
// トリミング予約区分で status が accounting/completed/cancelled でない = まだ会計待ちに進んでいない取り残し候補。
func (r *billingItemRepository) CountNonAccountingTrimmingByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Joins("JOIN reservation_types rt ON rt.id = appointments.reservation_type_id AND rt.deleted_at IS NULL").
		Where("appointments.clinic_id = ? AND appointments.pet_id = ? AND appointments.deleted_at IS NULL", clinicID, petID).
		Where("rt.category = ?", model.ReservationTypeCategoryTrimming).
		Where("appointments.status NOT IN ?", []model.ReservationStatus{
			model.ReservationStatusAccounting,
			model.ReservationStatusCompleted,
			model.ReservationStatusCancelled,
		}).
		Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", date).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "appointment", fmt.Sprintf("clinic=%d pet=%d trimming", clinicID, petID))
	}
	return count, nil
}
