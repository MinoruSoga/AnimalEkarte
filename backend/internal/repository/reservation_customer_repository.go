package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationCustomerRepository は予約顧客のデータアクセスインターフェース
type ReservationCustomerRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCustomer, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCustomer, error)
	UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
}

type reservationCustomerRepository struct{ db *gorm.DB }

func NewReservationCustomerRepository(db *gorm.DB) ReservationCustomerRepository {
	return &reservationCustomerRepository{db: db}
}

func (r *reservationCustomerRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ReservationCustomer, error) {
	items := make([]model.ReservationCustomer, 0)
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("clinic_id = ?", clinicID).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "find reservation customers")
	}
	return items, nil
}

func (r *reservationCustomerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationCustomer, error) {
	var c model.ReservationCustomer
	err := r.db.WithContext(ctx).
		Preload("Owner").
		First(&c, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_customer", fmt.Sprintf("%d", id))
	}
	return &c, nil
}

func (r *reservationCustomerRepository) UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationCustomer{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("owner_id", ownerID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update reservation customer owner link")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_customer", fmt.Sprintf("%d", id))
	}
	return nil
}
