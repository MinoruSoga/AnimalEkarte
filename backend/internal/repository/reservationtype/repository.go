// Package reservationtype owns reservation_types master data access (Wave D thin slice).
// Child tables (slots/unavailable/occupation/liff) and reservation core stay in the parent package.
package reservationtype

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for reservation types.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	FindAllWithChildren(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	FindByIDWithChildren(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
	CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	Create(ctx context.Context, reservationType *model.ReservationType) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	reservationTypes := make([]model.ReservationType, 0)
	if err := repohelpers.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&reservationTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return reservationTypes, nil
}

func (r *repository) FindAllWithChildren(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	reservationTypes := make([]model.ReservationType, 0)
	if err := repohelpers.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID).Order("sort_order ASC, name ASC")
		}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&reservationTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return reservationTypes, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	var st model.ReservationType
	err := repohelpers.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Parent", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).First(&st).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *repository) FindByIDWithChildren(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	var st model.ReservationType
	err := repohelpers.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Parent", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID).Order("sort_order ASC, name ASC")
		}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&st).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *repository) Create(ctx context.Context, reservationType *model.ReservationType) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(reservationType).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByReservationTypeID returns appointment references.
func (r *repository) CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child reservation-type count.
func (r *repository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.ReservationType{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

// Reorder updates sort_order for ids (ambient WithTx via repohelpers.ReorderByClinicID → DBOrTx).
func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.ReservationType{}, "reservation_type", clinicID, ids, "sort_order")
}
