package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LineCustomerRepository は予約顧客のデータアクセスインターフェース
type LineCustomerRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
	UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error
	FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error)
	UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error
}

type reservationCustomerRepository struct{ db *gorm.DB }

func NewLineCustomerRepository(db *gorm.DB) LineCustomerRepository {
	return &reservationCustomerRepository{db: db}
}

func (r *reservationCustomerRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.LineCustomer, error) {
	items := make([]model.LineCustomer, 0)
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Where("clinic_id = ?", clinicID).
		Order("created_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_customer", "")
	}
	return items, nil
}

func (r *reservationCustomerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
	var c model.LineCustomer
	err := r.db.WithContext(ctx).
		Preload("Owner").
		Preload("Owner.Pets").
		Preload("Owner.Pets.AnimalSpecies").
		First(&c, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_customer", fmt.Sprintf("%d", id))
	}
	return &c, nil
}

func (r *reservationCustomerRepository) FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error) {
	var c model.LineCustomer
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND line_user_id = ?", clinicID, lineUserID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c = model.LineCustomer{
			ClinicID:    clinicID,
			LineUserID:  lineUserID,
			DisplayName: displayName,
		}
		if err2 := r.db.WithContext(ctx).Create(&c).Error; err2 != nil {
			return nil, apperrors.FromGORM(err2, "reservation_customer", "")
		}
		return &c, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_customer", lineUserID)
	}
	return &c, nil
}

func (r *reservationCustomerRepository) UpdateAdditionalFields(ctx context.Context, clinicID, id uint64, fields []byte) error {
	result := r.db.WithContext(ctx).
		Model(&model.LineCustomer{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("additional_fields", fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_customer", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *reservationCustomerRepository) UpdateOwnerLink(ctx context.Context, clinicID, id uint64, ownerID *uint64) error {
	result := r.db.WithContext(ctx).
		Model(&model.LineCustomer{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Update("owner_id", ownerID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_customer", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_customer", fmt.Sprintf("%d", id))
	}
	return nil
}
