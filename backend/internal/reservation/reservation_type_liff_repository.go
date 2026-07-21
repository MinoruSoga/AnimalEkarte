package reservation

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// ReservationTypeLiffRepository は予約コース（reservation_types の予約用ラッパー）のデータアクセスインターフェース
type ReservationTypeLiffRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	Create(ctx context.Context, st *model.ReservationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
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
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_liff", "")
	}
	return items, nil
}

func (r *reservationTypeLiffRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	return repohelpers.FindByIDScoped[model.ReservationType](ctx, r.db, "reservation_type_liff", clinicID, id)
}

func (r *reservationTypeLiffRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

func (r *reservationTypeLiffRepository) Create(ctx context.Context, st *model.ReservationType) error {
	if err := r.db.WithContext(ctx).Create(st).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_liff", "")
	}
	return nil
}

func (r *reservationTypeLiffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.ReservationType{}, "reservation_type_liff", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationTypeLiffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
	if result.Error != nil {
		// service 層で ExistsByReservationTypeID による先行チェック済みのため通常は発生しない。
		// race condition 時のフォールバックとして FK 制約エラーを Conflict に変換する。
		if repohelpers.IsFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このコースは予約に使用されているため削除できません")
		}
		return apperrors.FromGORM(result.Error, "reservation_type_liff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_liff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationTypeLiffRepository) UpdateSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		var target model.ReservationType
		if err := tx.Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&target).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", id))
		}

		var neighbor model.ReservationType
		q := tx.Scopes(repohelpers.ClinicScope(clinicID))
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

		if err := tx.Model(&model.ReservationType{}).Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", target.ID))
		}
		if err := tx.Model(&model.ReservationType{}).Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type_liff", fmt.Sprintf("%d", neighbor.ID))
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to swap sort order")
	}
	return nil
}
