package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ReservationCategoryGroupRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategoryGroup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategoryGroup, error)
	CountCategories(ctx context.Context, groupID uint64) (int64, error)
	Create(ctx context.Context, g *model.ReservationCategoryGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationCategoryGroupRepository struct{ db *gorm.DB }

func NewReservationCategoryGroupRepository(db *gorm.DB) ReservationCategoryGroupRepository {
	return &reservationCategoryGroupRepository{db: db}
}

func (r *reservationCategoryGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCategoryGroup, error) {
	var list []model.ReservationCategoryGroup
	if err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&list).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_category_group", "")
	}
	return list, nil
}

func (r *reservationCategoryGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCategoryGroup, error) {
	var g model.ReservationCategoryGroup
	if err := r.db.WithContext(ctx).First(&g, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_category_group", fmt.Sprintf("%d", id))
	}
	return &g, nil
}

func (r *reservationCategoryGroupRepository) CountCategories(ctx context.Context, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReservationCategory{}).
		Where("group_id = ?", groupID).Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_category", "")
	}
	return count, nil
}

func (r *reservationCategoryGroupRepository) Create(ctx context.Context, g *model.ReservationCategoryGroup) error {
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称のグループが既に登録されています")
		}
		return apperrors.FromGORM(err, "reservation_category_group", "")
	}
	return nil
}

func (r *reservationCategoryGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationCategoryGroup{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_category_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_category_group", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCategoryGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.ReservationCategoryGroup{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_category_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_category_group", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCategoryGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(r.db, ctx, &model.ReservationCategoryGroup{}, "reservation_category_group", clinicID, ids)
}
