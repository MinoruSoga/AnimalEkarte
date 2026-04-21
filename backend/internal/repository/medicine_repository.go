// Package repository provides data access implementations for Medicine entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Medicine ----

type MedicineRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
	CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	medicines := make([]model.Medicine, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.Medicine{}).Scopes(clinicScope(clinicID))
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medicine", "")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&medicines).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "medicine", "")
	}
	return medicines, total, nil
}

func (r *medicineRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error) {
	var medicine model.Medicine
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&medicine).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", id))
	}
	return &medicine, nil
}

// CountUsageByMedicineID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-108）
// clinic_id フィルタを JOIN で適用しテナント分離を保証する（BUG-377）
func (r *medicineRepository) CountUsageByMedicineID(ctx context.Context, clinicID, medicineID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
		Where("treatments.medicine_id = ? AND treatments.deleted_at IS NULL", medicineID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.clinic_id = ? AND hospitalizations.deleted_at IS NULL", clinicID).
		Where("care_plan_items.medicine_id = ? AND care_plan_items.deleted_at IS NULL", medicineID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *medicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Scopes(clinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "medicine", "")
	}
	return count, nil
}

func (r *medicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {
	err := r.db.WithContext(ctx).Create(medicine).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "medicine", "")
	}
	return nil
}

func (r *medicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Medicine, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *medicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Medicine{}, "medicine", clinicID, ids)
}

func (r *medicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Medicine{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
	}
	return nil
}
