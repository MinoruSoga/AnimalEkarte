package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CarePlanItemRepository はケアプランアイテムのデータアクセスインターフェース
type CarePlanItemRepository interface {
	FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CarePlanItem, error)
	Create(ctx context.Context, item *model.CarePlanItem) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type carePlanItemRepository struct {
	db *gorm.DB
}

// NewCarePlanItemRepository は CarePlanItemRepository を初期化して返す
func NewCarePlanItemRepository(db *gorm.DB) CarePlanItemRepository {
	return &carePlanItemRepository{db: db}
}

func (r *carePlanItemRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.CarePlanItem, error) {
	items := make([]model.CarePlanItem, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.hospitalization_id = ?", clinicID, hospitalizationID).
		Order("care_plan_items.sort_order ASC").
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "care_plan_item", "")
	}
	return items, nil
}

func (r *carePlanItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CarePlanItem, error) {
	var item model.CarePlanItem
	err := r.db.WithContext(ctx).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.id = ?", clinicID, id).
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		First(&item).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "care_plan_item", fmt.Sprintf("%d", id))
	}
	return &item, nil
}

func (r *carePlanItemRepository) Create(ctx context.Context, item *model.CarePlanItem) error {
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		return apperrors.FromGORM(err, "care_plan_item", "")
	}
	return nil
}

func (r *carePlanItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.CarePlanItem{}).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.id = ?", clinicID, id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "care_plan_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *carePlanItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Joins("JOIN hospitalizations ON hospitalizations.id = care_plan_items.hospitalization_id AND hospitalizations.deleted_at IS NULL").
		Where("hospitalizations.clinic_id = ? AND care_plan_items.id = ?", clinicID, id).
		Unscoped().
		Delete(&model.CarePlanItem{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "care_plan_item", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("care_plan_item", fmt.Sprintf("%d", id))
	}
	return nil
}
