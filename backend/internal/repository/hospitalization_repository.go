package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type HospitalizationRepository interface {
	FindAll(ctx context.Context, clinicID uuid.UUID, status *string, page, limit int) ([]model.Hospitalization, int64, error)
	FindByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
	Update(ctx context.Context, hospitalization *model.Hospitalization) error
	Delete(ctx context.Context, clinicID, id uuid.UUID) error
}

type hospitalizationRepository struct {
	db *gorm.DB
}

func NewHospitalizationRepository(db *gorm.DB) HospitalizationRepository {
	return &hospitalizationRepository{db: db}
}

func (r *hospitalizationRepository) FindAll(ctx context.Context, clinicID uuid.UUID, status *string, page, limit int) ([]model.Hospitalization, int64, error) {
	var hospitalizations []model.Hospitalization
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Hospitalization{}).Where("clinic_id = ?", clinicID)
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count hospitalizations")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("start_date DESC, created_at DESC").Find(&hospitalizations).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find hospitalizations")
	}
	return hospitalizations, total, nil
}

func (r *hospitalizationRepository) FindByID(ctx context.Context, clinicID, id uuid.UUID) (*model.Hospitalization, error) {
	var hospitalization model.Hospitalization
	if err := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("Cage").
		Preload("Doctor").
		Preload("CarePlanItems").
		Preload("DailyRecords").
		Preload("TreatmentPlans").
		First(&hospitalization, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("hospitalization", id.String())
		}
		return nil, apperrors.Wrap(err, "find hospitalization by id")
	}
	return &hospitalization, nil
}

func (r *hospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	if err := r.db.WithContext(ctx).Create(hospitalization).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("hospitalization", hospitalization.StartDate.String())
		}
		return apperrors.Wrap(err, "create hospitalization")
	}
	return nil
}

func (r *hospitalizationRepository) Update(ctx context.Context, hospitalization *model.Hospitalization) error {
	if err := r.db.WithContext(ctx).Save(hospitalization).Error; err != nil {
		return apperrors.Wrap(err, "update hospitalization")
	}
	return nil
}

func (r *hospitalizationRepository) Delete(ctx context.Context, clinicID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Hospitalization{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete hospitalization")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization", id.String())
	}
	return nil
}
