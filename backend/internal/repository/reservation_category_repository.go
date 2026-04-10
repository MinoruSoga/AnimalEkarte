// Package repository provides data access implementations for ReservationCategory entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ReservationCategory ----

type ReservationCategoryRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error)
	Create(ctx context.Context, reservationCategory *model.ReservationCategory) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationCategoryRepository struct{ db *gorm.DB }

func NewReservationCategoryRepository(db *gorm.DB) ReservationCategoryRepository {
	return &reservationCategoryRepository{db: db}
}

func (r *reservationCategoryRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategory, error) {
	reservationCategories := make([]model.ReservationCategory, 0)
	if err := r.db.WithContext(ctx).
		Preload("Group").
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&reservationCategories).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_category", "")
	}
	return reservationCategories, nil
}

func (r *reservationCategoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategory, error) {
	var st model.ReservationCategory
	err := r.db.WithContext(ctx).Preload("Group").First(&st, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_category", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationCategoryRepository) Create(ctx context.Context, reservationCategory *model.ReservationCategory) error {
	if err := r.db.WithContext(ctx).Create(reservationCategory).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "reservation_category", "")
	}
	return nil
}

func (r *reservationCategoryRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationCategory{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_category", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.ReservationCategory{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_category", fmt.Sprintf("%d", id))
		}
		if count == 0 {
			return apperrors.WrapNotFound("reservation_category", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *reservationCategoryRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ReservationCategory{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		// BUG-030: ON DELETE RESTRICT の FK 制約違反は 409 Conflict に変換する
		if isFKConstraintErr(result.Error) {
			return apperrors.WrapConflict("このサービス種別は予約に使用されているため削除できません")
		}
		return apperrors.FromGORM(result.Error, "reservation_category", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_category", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder はトランザクション内で予約区分の sort_order を ids の順序で更新する
func (r *reservationCategoryRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.ReservationCategory{}, "reservation_category", clinicID, ids)
}
