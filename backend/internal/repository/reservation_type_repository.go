// Package repository provides data access implementations for ReservationType entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ReservationType ----

type ReservationTypeRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, reservationType *model.ReservationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type reservationTypeRepository struct{ db *gorm.DB }

func NewReservationTypeRepository(db *gorm.DB) ReservationTypeRepository {
	return &reservationTypeRepository{db: db}
}

func (r *reservationTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	reservationTypes := make([]model.ReservationType, 0)
	if err := r.db.WithContext(ctx).
		Preload("Group", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&reservationTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return reservationTypes, nil
}

func (r *reservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	var st model.ReservationType
	err := r.db.WithContext(ctx).Preload("Group", "deleted_at IS NULL").Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&st).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationTypeRepository) Create(ctx context.Context, reservationType *model.ReservationType) error {
	if err := r.db.WithContext(ctx).Create(reservationType).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "reservation_type", "")
	}
	return nil
}

func (r *reservationTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *reservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByReservationTypeID は予約区分を参照している appointments の件数を返す。
// appointments テーブルは直接 clinic_id を持つためテナント分離を直接適用する。
func (r *reservationTypeRepository) CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Reservation{}).
		Scopes(clinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// Reorder はトランザクション内で予約区分の sort_order を ids の順序で更新する
func (r *reservationTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.ReservationType{}, "reservation_type", clinicID, ids)
}
