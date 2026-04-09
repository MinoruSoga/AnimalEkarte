package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationCourseRepository は予約コース（service_types の予約用ラッパー）のデータアクセスインターフェース
type ReservationCourseRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ServiceType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error)
	Create(ctx context.Context, st *model.ServiceType) error
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

func (r *reservationCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ServiceType, error) {
	items := make([]model.ServiceType, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "find reservation courses")
	}
	return items, nil
}

func (r *reservationCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ServiceType, error) {
	var st model.ServiceType
	err := r.db.WithContext(ctx).First(&st, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationCourseRepository) Create(ctx context.Context, st *model.ServiceType) error {
	if err := r.db.WithContext(ctx).Create(st).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.Wrap(err, "create reservation course")
	}
	return nil
}

func (r *reservationCourseRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ServiceType{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update reservation course")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_course", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ServiceType{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		if isFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このコースは予約に使用されているため削除できません")
		}
		return apperrors.Wrap(result.Error, "delete reservation course")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_course", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCourseRepository) SwapSortOrder(ctx context.Context, clinicID, id uint64, direction string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var target model.ServiceType
		if err := tx.First(&target, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_course", fmt.Sprintf("%d", id))
		}

		var neighbor model.ServiceType
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

		if err := tx.Model(&model.ServiceType{}).Where("id = ?", target.ID).Update("sort_order", neighborOrder).Error; err != nil {
			return apperrors.Wrap(err, "swap sort_order target")
		}
		if err := tx.Model(&model.ServiceType{}).Where("id = ?", neighbor.ID).Update("sort_order", targetOrder).Error; err != nil {
			return apperrors.Wrap(err, "swap sort_order neighbor")
		}
		return nil
	})
}
