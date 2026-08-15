package reservation

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationTypeRepository is the data access interface for reservation types.
type ReservationTypeRepository interface {
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

type reservationTypeRepository struct{ db *gorm.DB }

// New constructs a ReservationTypeRepository.
func NewReservationTypeRepository(db *gorm.DB) ReservationTypeRepository {
	return &reservationTypeRepository{db: db}
}

func (r *reservationTypeRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	reservationTypes := make([]model.ReservationType, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&reservationTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return reservationTypes, nil
}

func (r *reservationTypeRepository) FindAllWithChildren(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	reservationTypes := make([]model.ReservationType, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID).Order("sort_order ASC, name ASC")
		}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id IS NULL").
		Order("sort_order ASC, name ASC").
		Find(&reservationTypes).Error; err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", "")
	}
	return reservationTypes, nil
}

func (r *reservationTypeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	var st model.ReservationType
	db := persistence.DBOrTx(ctx, r.db)
	if persistence.TxFromContext(ctx) != nil {
		db = db.Clauses(clause.Locking{Strength: "SHARE"})
	}
	err := db.
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Parent", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&st).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationTypeRepository) FindByIDWithChildren(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error) {
	var st model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Group", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Parent", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("clinic_id = ? AND deleted_at IS NULL", clinicID).Order("sort_order ASC, name ASC")
		}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&st).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", id))
	}
	return &st, nil
}

func (r *reservationTypeRepository) Create(ctx context.Context, reservationType *model.ReservationType) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := reservationType.IsActive
	wantVisible := reservationType.ReservationVisible
	if err := db.Create(reservationType).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type", "")
	}
	if !wantActive {
		if err := db.Model(reservationType).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", reservationType.ID))
		}
		reservationType.IsActive = false
	}
	if !wantVisible {
		if err := db.Model(reservationType).Update("reservation_visible", false).Error; err != nil {
			return apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", reservationType.ID))
		}
		reservationType.ReservationVisible = false
	}
	return nil
}

func (r *reservationTypeRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ReservationType, error) {
	// RSV-03: write + reload in one transaction.
	var loaded *model.ReservationType
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		result := tx.
			Model(&model.ReservationType{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			Updates(fields)
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
		}
		if result.RowsAffected == 0 {
			return apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
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

func (r *reservationTypeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).Delete(&model.ReservationType{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountUsageByReservationTypeID returns appointment references.
func (r *reservationTypeRepository) CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Reservation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND deleted_at IS NULL", id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation", "")
	}
	return count, nil
}

// CountChildrenByParentID returns child reservation-type count.
func (r *reservationTypeRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.ReservationType{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "reservation_type", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

// Reorder updates sort_order for ids (ambient WithTx via persistence.ReorderByClinicID → DBOrTx).
func (r *reservationTypeRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.ReservationType{}, "reservation_type", clinicID, ids, "sort_order")
}
