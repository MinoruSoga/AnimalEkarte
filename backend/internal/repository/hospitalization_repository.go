package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type HospitalizationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	// LockByIDForUpdate は FOR UPDATE で入院行をロック取得する（Discharge Q2-C）。
	// Repositories.Transaction（repo-swap）内の txRepos 経由でのみ呼ぶこと。
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	Create(ctx context.Context, hospitalization *model.Hospitalization) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	UpdateIfNotDischarged(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error)
	Delete(ctx context.Context, clinicID, id uint64) error
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
	if err := q.Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Pet.AnimalSpecies").Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Cage", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Doctor", "deleted_at IS NULL").
		Scopes(paginate(page, limit)).Order("start_date DESC, created_at DESC").
		Find(&hospitalizations).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "hospitalization", "")
	}
	return hospitalizations, total, nil
}

func (r *hospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	var hospitalization model.Hospitalization
	err := r.db.WithContext(ctx).
		Preload("Pet", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Pet.AnimalSpecies").
		Preload("Owner", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Cage", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Doctor", "deleted_at IS NULL").
		Preload("CarePlanItems").
		Preload("DailyRecords").
		Preload("TreatmentPlans", "deleted_at IS NULL").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&hospitalization).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "hospitalization", fmt.Sprintf("%d", id))
	}
	return &hospitalization, nil
}

// LockByIDForUpdate は FOR UPDATE で入院を行ロック取得する（Discharge Q2-C）。
// OwnerID/PetID など Q2-A 再検証に必要なスカラーはロック取得時の行スナップショットに含まれる。
// DischargeWithBilling は Repositories.Transaction（repo-swap）内の txRepos 経由で呼ぶこと。
// r.db が tx にバインドされていないとロックは SELECT 終了と同時に解放され直列化できない。
func (r *hospitalizationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	var hospitalization model.Hospitalization
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).First(&hospitalization).Error
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

func (r *hospitalizationRepository) UpdateIfNotDischarged(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Hospitalization, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Hospitalization{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ? AND status != ?", id, model.HospitalizationStatusDischarged).
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
		Scopes(clinicScope(clinicID)).
		Where("treatment_plans.hospitalization_id = ? AND treatment_plans.deleted_at IS NULL", hospitalizationID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "treatment_plan", fmt.Sprintf("hospitalization_id=%d", hospitalizationID))
	}
	return count, nil
}
