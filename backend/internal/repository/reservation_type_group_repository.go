package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ReservationTypeGroupRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error)
	CountCategories(ctx context.Context, groupID uint64) (int64, error)
	Create(ctx context.Context, g *model.ReservationTypeGroup) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeGroupRepository struct{ db *gorm.DB }

func NewReservationTypeGroupRepository(db *gorm.DB) ReservationTypeGroupRepository {
	return &reservationTypeGroupRepository{db: db}
}

func (r *reservationTypeGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationTypeGroup, error) {
	var list []model.ReservationTypeGroup
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&list).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_group", "")
	}
	return list, nil
}

func (r *reservationTypeGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeGroup, error) {
	var g model.ReservationTypeGroup
	if err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&g).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_group", fmt.Sprintf("%d", id))
	}
	return &g, nil
}

func (r *reservationTypeGroupRepository) CountCategories(ctx context.Context, groupID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).
		Where("group_id = ?", groupID).Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type", "")
	}
	return count, nil
}

func (r *reservationTypeGroupRepository) Create(ctx context.Context, g *model.ReservationTypeGroup) error {
	if err := r.db.WithContext(ctx).Create(g).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称のグループが既に登録されています")
		}
		return apperrors.FromGORM(err, "reservation_type_group", "")
	}
	return nil
}

func (r *reservationTypeGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationTypeGroup{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_group", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationTypeGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationTypeGroup{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_group", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_group", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationTypeGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.ReservationTypeGroup{}, "reservation_type_group", clinicID, ids)
}
