package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type TrimmingRepository interface {
	FindAll(ctx context.Context, petID *uuid.UUID, page, limit int) ([]model.TrimmingRecord, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingRecord, error)
	Create(ctx context.Context, trimming *model.TrimmingRecord) error
	Update(ctx context.Context, trimming *model.TrimmingRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type trimmingRepository struct {
	db *gorm.DB
}

func NewTrimmingRepository(db *gorm.DB) TrimmingRepository {
	return &trimmingRepository{db: db}
}

func (r *trimmingRepository) FindAll(ctx context.Context, petID *uuid.UUID, page, limit int) ([]model.TrimmingRecord, int64, error) {
	var trimmings []model.TrimmingRecord
	var total int64

	q := r.db.WithContext(ctx).Model(&model.TrimmingRecord{})
	if petID != nil {
		q = q.Where("pet_id = ?", petID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count trimming records")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("date DESC, created_at DESC").Find(&trimmings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find trimming records")
	}
	return trimmings, total, nil
}

func (r *trimmingRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.TrimmingRecord, error) {
	var trimming model.TrimmingRecord
	if err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Staff").
		Preload("Course").
		Preload("Options").
		First(&trimming, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_record", id.String())
		}
		return nil, apperrors.Wrap(err, "find trimming record by id")
	}
	return &trimming, nil
}

func (r *trimmingRepository) Create(ctx context.Context, trimming *model.TrimmingRecord) error {
	if err := r.db.WithContext(ctx).Create(trimming).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("trimming_record", trimming.Date.String())
		}
		return apperrors.Wrap(err, "create trimming record")
	}
	return nil
}

func (r *trimmingRepository) Update(ctx context.Context, trimming *model.TrimmingRecord) error {
	if err := r.db.WithContext(ctx).Save(trimming).Error; err != nil {
		return apperrors.Wrap(err, "update trimming record")
	}
	return nil
}

func (r *trimmingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingRecord{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete trimming record")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_record", id.String())
	}
	return nil
}
