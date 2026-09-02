// Package shifttemplate owns shift_templates / shift_template_breaks data access
// (BE8-4 batch12 — leaf domain).
package staff

import (
	"context"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository はシフトテンプレートのデータアクセスインターフェース
type ShiftTemplateRepository interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
	FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	LockActiveByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error)
	Create(ctx context.Context, tpl *model.ShiftTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	UpdateBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type shiftTemplateRepository struct{ db *gorm.DB }

// New は Repository を初期化して返す
func NewShiftTemplateRepository(db *gorm.DB) ShiftTemplateRepository {
	return &shiftTemplateRepository{db: db}
}

func (r *shiftTemplateRepository) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "shift template transaction failed")
	}
	return nil
}

func (r *shiftTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ShiftTemplate, error) {
	items := make([]model.ShiftTemplate, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Breaks").
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", "")
	}
	return items, nil
}

func (r *shiftTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ShiftTemplate, error) {
	var tpl model.ShiftTemplate
	err := persistence.DBOrTx(ctx, r.db).
		Preload("Breaks").
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		First(&tpl).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "shift_template", strconv.FormatUint(id, 10))
	}
	return &tpl, nil
}

func (r *shiftTemplateRepository) LockActiveByIDForUpdate(
	ctx context.Context,
	clinicID, id uint64,
) (*model.ShiftTemplate, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError(
			"shift template update lock requires an active transaction",
		)
	}
	var template model.ShiftTemplate
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&template).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"shift_template",
			strconv.FormatUint(id, 10),
		)
	}
	return &template, nil
}

func (r *shiftTemplateRepository) Create(ctx context.Context, tpl *model.ShiftTemplate) error {
	db := persistence.DBOrTx(ctx, r.db)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantActive := tpl.IsActive
	if err := db.Omit("Breaks").Create(tpl).Error; err != nil {
		return apperrors.FromGORM(err, "shift_template", "")
	}
	if !wantActive {
		if err := db.Model(tpl).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "shift_template", strconv.FormatUint(tpl.ID, 10))
		}
		tpl.IsActive = false
	}
	return nil
}

func (r *shiftTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ShiftTemplate, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.ShiftTemplate{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *shiftTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.ShiftTemplate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "shift_template", strconv.FormatUint(id, 10))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("shift_template", strconv.FormatUint(id, 10))
	}
	return nil
}

func (r *shiftTemplateRepository) UpdateBreaks(ctx context.Context, templateID uint64, breaks []model.ShiftTemplateBreak) error {
	return persistence.ReplaceChildRowsByParentID(
		persistence.DBOrTx(ctx, r.db),
		templateID,
		breaks,
		&model.ShiftTemplateBreak{},
		"shift_template_id",
		"shift_template_break",
		func(row *model.ShiftTemplateBreak, id uint64) { row.ShiftTemplateID = id },
		"failed to replace shift template breaks",
	)
}

func (r *shiftTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.ShiftTemplate{}, "shift_template", clinicID, ids, "sort_order")
}
