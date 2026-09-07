package reservation

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationTypeLiffRepository は予約コース（reservation_types の予約用ラッパー）のデータアクセスインターフェース
type ReservationTypeLiffRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	Create(ctx context.Context, st *model.ReservationType) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	// DeleteWithDependencyChecks locks the master row, re-evaluates children/appointment
	// usage, then soft-deletes in one transaction (RSV-07).
	DeleteWithDependencyChecks(ctx context.Context, clinicID, id uint64, usage reservationTypeUsageChecker) error
	// UpdateSortOrder は隣接するレコードとの sort_order をスワップする。
	// direction は "up"（sort_order 小さい方）または "down"（sort_order 大きい方）。
	UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
}

type reservationTypeLiffRepository struct{ db *gorm.DB }

func NewReservationTypeLiffRepository(db *gorm.DB) ReservationTypeLiffRepository {
	return &reservationTypeLiffRepository{db: db}
}

func (r *reservationTypeLiffRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	items := make([]model.ReservationType, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_liff", "")
	}
	return items, nil
}

func (r *reservationTypeLiffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return persistence.FindByIDScoped[model.ReservationType](ctx, persistence.DBOrTx(ctx, r.db), "reservation_type_liff", clinicID, id)
}

func (r *reservationTypeLiffRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

func (r *reservationTypeLiffRepository) Create(ctx context.Context, st *model.ReservationType) error {
	db := r.db.WithContext(ctx)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	// LIFF create hardcodes IsActive=true; only ReservationVisible needs compensation.
	wantVisible := st.ReservationVisible
	if err := db.Create(st).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_liff", "")
	}
	if !wantVisible {
		if err := db.Model(st).Update("reservation_visible", false).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", st.ID))
		}
		st.ReservationVisible = false
	}
	return nil
}

func (r *reservationTypeLiffRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateReservationTypeLiffInput) (*model.ReservationType, error) {
	// RSV-03: write + reload in one transaction.
	var loaded *model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := r.update(txCtx, clinicID, id, buildReservationTypeLiffUpdate(&cmd)); err != nil {
			return err
		}
		var findErr error
		loaded, findErr = r.FindByID(txCtx, clinicID, id)
		return findErr
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *reservationTypeLiffRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.ReservationType{}, "reservation_type_liff", clinicID, id, fields)
}

func (r *reservationTypeLiffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM reservation_types AS child_reservation_types
			WHERE child_reservation_types.parent_id = reservation_types.id
			  AND child_reservation_types.clinic_id = ?
			  AND child_reservation_types.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM appointments
			WHERE appointments.reservation_type_id = reservation_types.id
			  AND appointments.clinic_id = ?
			  AND appointments.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.ReservationType{})
	if result.Error != nil {
		// race condition 時のフォールバックとして FK 制約エラーを Conflict に変換する。
		if persistence.IsFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このコースは予約に使用されているため削除できません")
		}
		return apperrors.FromGORM(result.Error, "reservation_type_liff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeDeleteIfUnusedMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *reservationTypeLiffRepository) normalizeDeleteIfUnusedMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この予約コースには子予約区分が登録されているため削除できません")
	}
	var usage int64
	if err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", id).
		Count(&usage).Error; err != nil {
		return apperrors.FromGORM(err, "reservation", "")
	}
	if usage > 0 {
		return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
	}
	return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
}

// DeleteWithDependencyChecks serializes master delete with appointment create
// (which FOR SHARE locks reservation_types). RSV-07.
func (r *reservationTypeLiffRepository) DeleteWithDependencyChecks(
	ctx context.Context,
	clinicID, id uint64,
	usage reservationTypeUsageChecker,
) error {
	return persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		var st model.ReservationType
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&st).Error
		if err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", id))
		}
		var childCount int64
		if err := tx.Model(&model.ReservationType{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("parent_id = ? AND deleted_at IS NULL", id).
			Count(&childCount).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", id))
		}
		if childCount > 0 {
			return apperrors.WrapConflict("この予約コースには子予約区分が登録されているため削除できません")
		}
		exists, err := usage.ExistsByReservationTypeID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to check reservation dependency")
		}
		if exists {
			return apperrors.WrapConflict("この予約コースは予約データで使用中のため削除できません")
		}
		if err := r.Delete(txCtx, clinicID, id); err != nil {
			return err
		}
		return nil
	})
}

func (r *reservationTypeLiffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var target model.ReservationType
		if err := tx.Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&target).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", id))
		}

		var neighbor model.ReservationType
		q := tx.Scopes(persistence.ClinicScope(clinicID))
		if direction == "up" {
			q = q.Where("sort_order < ?", target.SortOrder).Order("sort_order DESC")
		} else {
			q = q.Where("sort_order > ?", target.SortOrder).Order("sort_order ASC")
		}
		if err := q.First(&neighbor).Error; err != nil {
			wrapped := apperrors.FromGORM(err, "reservation_type_liff", "neighbor")
			if errors.Is(wrapped, apperrors.ErrNotFound) {
				// 隣接なし → 変更なし
				return nil
			}
			return wrapped
		}

		targetOrder := target.SortOrder
		neighborOrder := neighbor.SortOrder

		if err := tx.Model(&model.ReservationType{}).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", target.ID))
		}
		if err := tx.Model(&model.ReservationType{}).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", neighbor.ID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to swap sort order")
	}
	return nil
}
