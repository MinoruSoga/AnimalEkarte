// Package reservationtypeavailableslot owns reservation_type_available_slots data access
// (BE8-4 Wave D — "reservation type children (slot/unavailable)" per repository/CLAUDE.md's
// next cut order).
package reservationtypeavailableslot

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository は予約可能開始時刻の永続化インターフェース
type Repository interface {
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type repository struct {
	db *gorm.DB
}

// New はリポジトリを初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(
	ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeAvailableSlot, error) {
	var results []model.ReservationTypeAvailableSlot
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("reservation_type_id = ?", reservationTypeID).
		Order("available_type ASC, day_of_week ASC, specific_date ASC, start_time ASC, id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_available_slots", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *repository) FindByID(
	ctx context.Context, clinicID, id uint64,
) (*model.ReservationTypeAvailableSlot, error) {
	var result model.ReservationTypeAvailableSlot
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	return &result, nil
}

func (r *repository) Create(
	ctx context.Context, slot *model.ReservationTypeAvailableSlot,
) error {
	if err := r.db.WithContext(ctx).Create(slot).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_available_slot", "")
	}
	return nil
}

func (r *repository) Delete(
	ctx context.Context, clinicID, id uint64,
) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.ReservationTypeAvailableSlot{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	return nil
}
