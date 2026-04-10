package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationCourseRepository は予約コース（reservation_categories の予約用ラッパー）のデータアクセスインターフェース
type ReservationCourseRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error)
	Create(ctx context.Context, st *model.ReservationCategory) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	// SwapSortOrder は隣接するレコードとの sort_order をスワップする。
	// direction は "up"（sort_order 小さい方）または "down"（sort_order 大きい方）。
	SwapSortOrder(ctx context.Context, clinicID, id uint64, direction string) error
}

type reservationCourseRepository struct{ db *gorm.DB }

func NewReservationCourseRepository(db *gorm.DB) ReservationCourseRepository {
	return &reservationCourseRepository{db: db}
}

func (r *reservationCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error) {
	items := make([]model.ReservationCategory, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_course", "")
	}
	return items, nil
}

func (r *reservationCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error) {
	var st model.ReservationCategory
	err := r.db.WithContext(ctx).First(&st, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationCourseRepository) Create(ctx context.Context, st *model.ReservationCategory) error {
	if err := r.db.WithContext(ctx).Create(st).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "reservation_course", "")
	}
	return nil
}

func (r *reservationCourseRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_course", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_course", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ReservationCategory{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		if isFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このコースは予約に使用されているため削除できません")
		}
		return apperrors.FromGORM(result.Error, "reservation_course", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_course", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCourseRepository) SwapSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target model.ReservationCategory
		if err := tx.First(&target, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", id))
		}

		var neighbor model.ReservationCategory
		q := tx.Where("clinic_id = ?", clinicID)
		if direction == "up" {
			q = q.Where("sort_order < ?", target.SortOrder).Order("sort_order DESC")
		} else {
			q = q.Where("sort_order > ?", target.SortOrder).Order("sort_order ASC")
		}
		if err := q.First(&neighbor).Error; err != nil {
			// 隣接なし → 変更なし
			return nil
		}

		targetOrder := target.SortOrder
		neighborOrder := neighbor.SortOrder

		if err := tx.Model(&model.ReservationCategory{}).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", target.ID))
		}
		if err := tx.Model(&model.ReservationCategory{}).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", neighbor.ID))
		}
		return nil
	})
}
