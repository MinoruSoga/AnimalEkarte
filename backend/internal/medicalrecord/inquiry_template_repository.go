package medicalrecord

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// InquiryTemplateRepository is the data access interface for inquiry templates.
// Moved from internal/repository/inquirytemplate (BE8-4 batch4) — BE9-2D roll-up. Renamed from
// that subpackage's generic "Repository" to this entity-specific name only because medicalrecord
// holds multiple repository interfaces in one package; every external caller only ever saw
// this name via the internal/repository facade, so no call site changes.
type InquiryTemplateRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, template *model.InquiryTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type inquiryTemplateRepository struct{ db *gorm.DB }

// NewInquiryTemplateRepository constructs an InquiryTemplateRepository.
func NewInquiryTemplateRepository(db *gorm.DB) InquiryTemplateRepository {
	return &inquiryTemplateRepository{db: db}
}

func (r *inquiryTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	templates := make([]model.InquiryTemplate, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, title ASC").
		Find(&templates).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", "")
	}
	return templates, nil
}

func (r *inquiryTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	return persistence.FindByIDScoped[model.InquiryTemplate](ctx, r.db, "inquiry_template", clinicID, id)
}

func (r *inquiryTemplateRepository) Create(ctx context.Context, template *model.InquiryTemplate) error {
	db := r.db.WithContext(ctx)
	wantActive := template.IsActive
	if err := db.Create(template).Error; err != nil {
		return apperrors.FromGORM(err, "inquiry_template", "")
	}
	if !wantActive {
		if err := db.Model(template).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "inquiry_template", fmt.Sprintf("%d", template.ID))
		}
		template.IsActive = false
	}
	return nil
}

func (r *inquiryTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error) {
	if err := persistence.UpdateScopedByID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *inquiryTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.InquiryTemplate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inquiry_template", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inquiry_template", fmt.Sprintf("%d", id))
	}
	return nil
}

// 現スキーマに inquiry_template_id を参照する FK テーブルが存在しないため常に 0 を返す。
// PO判断（2026-05-25）: inquiry_answers は当面実装しない。
// 将来 inquiry_answers 等を追加する場合は、その実装 PR 内でこの関数を COUNT クエリに書き換えること。
func (r *inquiryTemplateRepository) CountUsageByInquiryTemplateID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (r *inquiryTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, ids, "sort_order")
}
