package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type TrimmingRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
	Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
	SetOptions(ctx context.Context, recordID uint64, optionIDs []uint64) error
}

type trimmingRepository struct {
	db *gorm.DB
}

func NewTrimmingRepository(db *gorm.DB) TrimmingRepository {
	return &trimmingRepository{db: db}
}

func (r *trimmingRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error) {
	trimmings := make([]model.TrimmingRecord, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.TrimmingRecord{}).Where("trimming_records.clinic_id = ?", clinicID)
	if petID != nil {
		q = q.Where("trimming_records.pet_id = ?", petID)
	}
	if ownerID != nil {
		q = q.Select("trimming_records.*").
			Joins("JOIN pets ON pets.id = trimming_records.pet_id").Where("pets.owner_id = ?", *ownerID)
	}
	if startDate != nil {
		q = q.Where("trimming_records.date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("trimming_records.date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count trimming records")
	}
	if err := q.
		Preload("Pet").
		Preload("Pet.Owner").
		Preload("Pet.AnimalSpecies").
		Preload("Staff").
		Preload("Course").
		Preload("Options").
		Offset((page - 1) * limit).Limit(limit).Order("date DESC, created_at DESC").
		Find(&trimmings).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find trimming records")
	}
	return trimmings, total, nil
}

func (r *trimmingRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error) {
	var trimming model.TrimmingRecord
	err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Staff").
		Preload("Course").
		Preload("Options").
		First(&trimming, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "trimming_record", fmt.Sprintf("%d", id))
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

func (r *trimmingRepository) SetOptions(ctx context.Context, recordID uint64, optionIDs []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := &model.TrimmingRecord{ID: recordID}
		if err := tx.Model(record).Association("Options").Unscoped().Clear(); err != nil {
			return apperrors.Wrap(err, "failed to clear trimming options")
		}
		if len(optionIDs) == 0 {
			return nil
		}
		options := make([]model.TrimmingOption, 0, len(optionIDs))
		for _, id := range optionIDs {
			options = append(options, model.TrimmingOption{ID: id})
		}
		if err := tx.Model(record).Association("Options").Replace(options); err != nil {
			return apperrors.Wrap(err, "failed to set trimming options")
		}
		return nil
	}); err != nil {
		return fmt.Errorf("set options trimming record: %w", err)
	}
	return nil
}
