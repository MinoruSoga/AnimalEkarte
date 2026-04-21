package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type HospitalizationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	ExistsByCageID(ctx context.Context, cageID uint64) (bool, error)
	CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

type hospitalizationRepository struct {
	db *gorm.DB
}

func NewHospitalizationRepository(db *gorm.DB) HospitalizationRepository {
	return &hospitalizationRepository{db: db}
}

func (r *hospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	hospitalizations := make([]model.Hospitalization, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Hospitalization{}).Scopes(clinicScope(clinicID))
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startDate != nil {
		q = q.Where("hospitalizations.start_date >= ?", *startDate)
	}
	if endDate != nil {
		q = q.Where("hospitalizations.start_date <= ?", *endDate)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	if err := q.Preload("Pet", "deleted_at IS NULL").Preload("Pet.AnimalSpecies", "deleted_at IS NULL").Preload("Owner", "deleted_at IS NULL").Preload("Cage", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").
		Offset((page - 1) * limit).Limit(limit).Order("start_date DESC, created_at DESC").
		Find(&hospitalizations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	return hospitalizations, total, nil
}

func (r *hospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	var hospitalization model.Hospitalization
	err := r.db.WithContext(ctx).
		Preload("Pet", "deleted_at IS NULL").
		Preload("Pet.AnimalSpecies", "deleted_at IS NULL").
		Preload("Owner", "deleted_at IS NULL").
		Preload("Cage", "deleted_at IS NULL").
		Preload("Doctor", "deleted_at IS NULL").
		Preload("CarePlanItems", "deleted_at IS NULL").
		Preload("DailyRecords", "deleted_at IS NULL").
		Preload("TreatmentPlans", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&hospitalization).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", fmt.Sprintf("%d", id))
	}
	return &hospitalization, nil
}

func (r *hospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	err := r.db.WithContext(ctx).Create(hospitalization).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("hospitalization", hospitalization.StartDate.String())
		}
		return apperrors.FromGORM(err, "hospitalization", "")
	}
	return nil
}

func (r *hospitalizationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Hospitalization{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "hospitalization", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *hospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Hospitalization{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "hospitalization", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("hospitalization", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *hospitalizationRepository) ExistsByCageID(ctx context.Context, cageID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Hospitalization{}).
		Where("cage_id = ?", cageID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "hospitalization", "")
	}
	return count > 0, nil
}

func (r *hospitalizationRepository) CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON care_plan_items.hospitalization_id = hospitalizations.id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.hospitalization_id = ?", clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}

func (r *hospitalizationRepository) CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.DailyRecord{}).
		Joins("JOIN hospitalizations ON daily_records.hospitalization_id = hospitalizations.id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND daily_records.hospitalization_id = ?", clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "daily_record", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}

func (r *hospitalizationRepository) CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TreatmentPlan{}).
		Joins("JOIN hospitalizations ON treatment_plans.hospitalization_id = hospitalizations.id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND treatment_plans.hospitalization_id = ?", clinicID, hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "treatment_plan", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}
