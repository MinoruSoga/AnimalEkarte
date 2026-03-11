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
	FindAll(ctx context.Context, status *string, page, limit int) ([]model.Accounting, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Accounting, error)
	Create(ctx context.Context, accounting *model.Accounting) error
	Update(ctx context.Context, accounting *model.Accounting) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type accountingRepository struct {
	db *gorm.DB
}

func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

func (r *accountingRepository) FindAll(ctx context.Context, status *string, page, limit int) ([]model.Accounting, int64, error) {
	var accountings []model.Accounting
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Accounting{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count accountings")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("scheduled_date DESC, created_at DESC").Find(&accountings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find accountings")
	}
	return accountings, total, nil
}

func (r *accountingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Accounting, error) {
	var accounting model.Accounting
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		Preload("Owner").
		Preload("Pet").
		First(&accounting, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("accounting", id.String())
		}
		return nil, apperrors.Wrap(err, "find accounting by id")
	}
	return &accounting, nil
}

func (r *accountingRepository) Create(ctx context.Context, accounting *model.Accounting) error {
	if err := r.db.WithContext(ctx).Create(accounting).Error; err != nil {
		return apperrors.Wrap(err, "create accounting")
	}
	return nil
}

func (r *accountingRepository) Update(ctx context.Context, accounting *model.Accounting) error {
	if err := r.db.WithContext(ctx).Save(accounting).Error; err != nil {
		return apperrors.Wrap(err, "update accounting")
	}
	return nil
}

func (r *accountingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.Accounting{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete accounting")
	}
	return nil
}
