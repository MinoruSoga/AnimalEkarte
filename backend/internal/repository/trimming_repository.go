package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type TrimmingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
	Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type trimmingRepository struct {
	db *gorm.DB
}

func NewTrimmingRepository(db *gorm.DB) TrimmingRepository {
	return &trimmingRepository{db: db}
}

func (r *trimmingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.TrimmingRecord, int64, error) {
	trimmings := make([]model.TrimmingRecord, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.TrimmingRecord{}).Where("trimming_records.clinic_id = ?", clinicID)
	if petID != nil {
		q = q.Where("trimming_records.pet_id = ?", petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN pets ON pets.id = trimming_records.pet_id").Where("pets.owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count trimming records")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("date DESC, created_at DESC").Find(&trimmings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find trimming records")
	}
	return trimmings, total, nil
}

func (r *trimmingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	var trimming model.TrimmingRecord
	if err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Staff").
		Preload("Course").
		Preload("Options").
		First(&trimming, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming_record", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find trimming record by id")
	}
	return &trimming, nil
}

func (r *trimmingRepository) Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	trimming.ClinicID = clinicID
	if err := r.db.WithContext(ctx).Create(trimming).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("trimming_record", trimming.Date.String())
		}
		return apperrors.Wrap(err, "create trimming record")
	}
	return nil
}

func (r *trimmingRepository) Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error {
	trimming.ClinicID = clinicID
	result := r.db.WithContext(ctx).
		Model(&model.TrimmingRecord{}).
		Where("id = ? AND clinic_id = ?", trimming.ID, clinicID).
		Updates(trimming)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update trimming record")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_record", fmt.Sprintf("%d", trimming.ID))
	}
	return nil
}

func (r *trimmingRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingRecord{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete trimming record")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("trimming_record", fmt.Sprintf("%d", id))
	}
	return nil
}
