package repository

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ShiftTemplateRepository はシフトテンプレートのデータアクセスインターフェース
type ShiftTemplateRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	Create(ctx context.Context, tpl *model.ShiftTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	ReplaceBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type shiftTemplateRepository struct{ db *gorm.DB }

// NewShiftTemplateRepository は ShiftTemplateRepository を初期化して返す
func NewShiftTemplateRepository(db *gorm.DB) ShiftTemplateRepository {
	return &shiftTemplateRepository{db: db}
}

func (r *shiftTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items := make([]model.ShiftTemplate, 0)
	err := r.db.WithContext(ctx).
		Preload("Breaks").
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", "")
	}
	return items, nil
}

func (r *shiftTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	var tpl model.ShiftTemplate
	err := r.db.WithContext(ctx).
		Preload("Breaks").
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		First(&tpl).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", strconv.FormatUint(id, 10))
	}
	return &tpl, nil
}

func (r *shiftTemplateRepository) Create(ctx context.Context, tpl *model.ShiftTemplate) error {
	if err := r.db.WithContext(ctx).Omit("Breaks").Create(tpl).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "shift_template", "")
	}
	return nil
}

func (r *shiftTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.ShiftTemplate{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		if isUniqueConstraintErr(result.Error) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftTemplate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftTemplateRepository) ReplaceBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("shift_template_id = ?", templateID).Delete(&model.ShiftTemplateBreak{}).Error; err != nil {
			return apperrors.FromGORM(err, "shift_template_break", fmt.Sprintf("%d", templateID))
		}
		if len(breaks) == 0 {
			return nil
		}
		for i := range breaks {
			breaks[i].ShiftTemplateID = templateID
		}
		if err := tx.Create(&breaks).Error; err != nil {
			return apperrors.FromGORM(err, "shift_template_break", "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to replace shift template breaks")
	}
	return nil
}

func (r *shiftTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.ShiftTemplate{}).
				Scopes(clinicScope(clinicID)).Where("id = ?", id).
				Update("sort_order", i)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("shift_template id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder shift templates")
	}
	return nil
}
