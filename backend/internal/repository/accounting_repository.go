package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type AccountingRepository interface {
	FindAll(ctx context.Context, status *string, page, limit int) ([]model.Billing, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Billing, error)
	Create(ctx context.Context, accounting *model.Billing) error
	Update(ctx context.Context, accounting *model.Billing) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, status *string, page, limit int) ([]model.Billing, int64, error) {
	var billings []model.Billing
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Billing{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count billings")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&billings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find billings")
	}
	return billings, total, nil
}

func (r *accountingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Billing, error) {
	var billing model.Billing
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Payments").
		Preload("Owner").
		Preload("Pet").
		First(&billing, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("billing", id.String())
		}
		return nil, apperrors.Wrap(err, "find billing by id")
	}
	return &billing, nil
}

func (r *accountingRepository) Create(ctx context.Context, accounting *model.Billing) error {
	if err := r.db.WithContext(ctx).Create(accounting).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("billing", accounting.ScheduledDate.String())
		}
		return apperrors.Wrap(err, "create billing")
	}
	return nil
}

func (r *accountingRepository) Update(ctx context.Context, accounting *model.Billing) error {
	if err := r.db.WithContext(ctx).Save(accounting).Error; err != nil {
		return apperrors.Wrap(err, "update billing")
	}
	return nil
}

func (r *accountingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Billing{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete billing")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("billing", id.String())
	}
	return nil
}
