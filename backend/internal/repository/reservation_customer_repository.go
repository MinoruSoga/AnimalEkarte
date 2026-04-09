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
	FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.ReservationCustomer, error)
	UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error
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

func (r *reservationCustomerRepository) FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.ReservationCustomer, error) {
	var c model.ReservationCustomer
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND line_user_id = ?", clinicID, lineUserID).
		First(&c).Error
	if err == gorm.ErrRecordNotFound {
		c = model.ReservationCustomer{
			ClinicID:    clinicID,
			LineUserID:  lineUserID,
			DisplayName: displayName,
		}
		if err2 := r.db.WithContext(ctx).Create(&c).Error; err2 != nil {
			return nil, apperrors.Wrap(err2, "create reservation customer")
		}
		return &c, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "find reservation customer by line user id")
	}
	return &c, nil
}

func (r *reservationCustomerRepository) UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error {
	result := r.db.WithContext(ctx).
		Model(&model.ReservationCustomer{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("additional_fields", fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update reservation customer additional fields")
	}
	return nil
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
