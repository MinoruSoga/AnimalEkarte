// Package shifttemplate owns shift_templates / shift_template_breaks data access
// (BE8-4 batch12 — leaf domain).
package shifttemplate

import (
	"context"
	"strconv"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository はシフトテンプレートのデータアクセスインターフェース
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	Create(ctx context.Context, tpl *model.ShiftTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByShiftTemplateID(ctx context.Context, clinicID, id uint64) (int64, error)
}

type repository struct{ db *gorm.DB }

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items := make([]model.ShiftTemplate, 0)
	err := r.db.WithContext(ctx).
		Preload("Breaks").
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", "")
	}
	return items, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	var tpl model.ShiftTemplate
	err := r.db.WithContext(ctx).
		Preload("Breaks").
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		First(&tpl).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", strconv.FormatUint(id, 10))
	}
	return &tpl, nil
}

func (r *repository) Create(ctx context.Context, tpl *model.ShiftTemplate) error {
	if err := r.db.WithContext(ctx).Omit("Breaks").Create(tpl).Error; err != nil {
		return apperrors.FromGORM(err, "shift_template", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ShiftTemplate{}).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftTemplate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *repository) UpdateBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
	return repohelpers.ReplaceChildRowsByParentID(
		repohelpers.DBOrTx(ctx, r.db),
		templateID,
		breaks,
		&model.ShiftTemplateBreak{},
		"shift_template_id",
		"shift_template_break",
		func(row *model.ShiftTemplateBreak, id uint64) { row.ShiftTemplateID = id },
		"failed to replace shift template breaks",
	)
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.ShiftTemplate{}, "shift_template", clinicID, ids, "sort_order")
}

// CountUsageByShiftTemplateID はシフトテンプレートを参照している shift_template_breaks の件数を返す。
func (r *repository) CountUsageByShiftTemplateID(ctx context.Context, clinicID, id uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.ShiftTemplateBreak{}).
		Joins("JOIN shift_templates ON shift_templates.id = shift_template_breaks.shift_template_id AND shift_templates.deleted_at IS NULL").
		Where("shift_templates.clinic_id = ? AND shift_template_breaks.shift_template_id = ?", clinicID, id).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "shift_template_break", "")
	}
	return count, nil
}
