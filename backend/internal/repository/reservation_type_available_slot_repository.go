package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationTypeAvailableSlotRepository は予約可能開始時刻の永続化インターフェース
type ReservationTypeAvailableSlotRepository interface {
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationTypeAvailableSlotRepository struct {
	db *gorm.DB
}

// NewReservationTypeAvailableSlotRepository はリポジトリを初期化して返す
func NewReservationTypeAvailableSlotRepository(db *gorm.DB) ReservationTypeAvailableSlotRepository {
	return &reservationTypeAvailableSlotRepository{db: db}
}

func (r *reservationTypeAvailableSlotRepository) FindAll(
	ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeAvailableSlot, error) {
	var results []model.ReservationTypeAvailableSlot
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("reservation_type_id = ?", reservationTypeID).
		Order("available_type ASC, day_of_week ASC, specific_date ASC, start_time ASC, id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_available_slots", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *reservationTypeAvailableSlotRepository) FindByID(
	ctx context.Context, clinicID, id uint64,
) (*model.ReservationTypeAvailableSlot, error) {
	var result model.ReservationTypeAvailableSlot
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	return &result, nil
}

func (r *reservationTypeAvailableSlotRepository) Create(
	ctx context.Context, slot *model.ReservationTypeAvailableSlot,
) error {
	if err := r.db.WithContext(ctx).Create(slot).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_available_slot", "")
	}
	return nil
}

func (r *reservationTypeAvailableSlotRepository) Delete(
	ctx context.Context, clinicID, id uint64,
) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.ReservationTypeAvailableSlot{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.FromGORM(gorm.ErrRecordNotFound, "reservation_type_available_slot", fmt.Sprintf("%d", id))
	}
	return nil
}
