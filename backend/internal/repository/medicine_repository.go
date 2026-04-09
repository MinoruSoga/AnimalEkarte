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
	CountChildren(ctx context.Context, clinicID, parentID uint64) (int64, error)
	CountUsageByMedicineID(ctx context.Context, medicineID uint64) (int64, error)
	Create(ctx context.Context, medicine *model.Medicine) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type medicineRepository struct{ db *gorm.DB }

func NewMedicineRepository(db *gorm.DB) MedicineRepository { return &medicineRepository{db: db} }

func (r *medicineRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Medicine, int64, error) {
	medicines := make([]model.Medicine, 0)
	var total int64

	buildBase := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&model.Medicine{}).Where("clinic_id = ?", clinicID)
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
		First(&medicine, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", id))
	}
	return &medicine, nil
}

// CountUsageByMedicineID は treatments と care_plan_items で参照されている件数の合計を返す（BUG-108）
func (r *medicineRepository) CountUsageByMedicineID(ctx context.Context, medicineID uint64) (int64, error) {
	var treatmentCount, carePlanCount int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Where("medicine_id = ?", medicineID).
		Count(&treatmentCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	if err := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Where("medicine_id = ?", medicineID).
		Count(&carePlanCount).Error; err != nil {
		return 0, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return treatmentCount + carePlanCount, nil
}

func (r *medicineRepository) CountChildren(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Where("clinic_id = ? AND parent_id = ?", clinicID, parentID).
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

func (r *medicineRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Medicine{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := r.db.WithContext(ctx).Model(&model.Medicine{}).
			Where("id = ? AND clinic_id = ?", id, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "medicine", fmt.Sprintf("%d", id))
		}
		if count == 0 {
			return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
		}
	}
	return nil
}

func (r *medicineRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Medicine{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("medicine id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "reorder medicine")
	}
	return nil
}

func (r *medicineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Medicine{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medicine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
	}
	return nil
}
