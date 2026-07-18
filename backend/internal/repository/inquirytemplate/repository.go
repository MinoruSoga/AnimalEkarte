// Package inquirytemplate owns inquiry_templates data access (BE8-4 batch4 — leaf domain).
package inquirytemplate

import (
	"context"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for inquiry templates.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, template *model.InquiryTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	templates := make([]model.InquiryTemplate, 0)
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("sort_order ASC, title ASC").
		Find(&templates).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", "")
	}
	return templates, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	return repohelpers.FindByIDScoped[model.InquiryTemplate](ctx, r.db, "inquiry_template", clinicID, id)
}

func (r *repository) Create(ctx context.Context, template *model.InquiryTemplate) error {
	err := r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		return apperrors.FromGORM(err, "inquiry_template", "")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, id, fields); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, id)
}

// 現スキーマに inquiry_template_id を参照する FK テーブルが存在しないため常に 0 を返す。
// PO判断（2026-05-25）: inquiry_answers は当面実装しない。
// 将来 inquiry_answers 等を追加する場合は、その実装 PR 内でこの関数を COUNT クエリに書き換えること。
func (r *repository) CountUsageByInquiryTemplateID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (r *repository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, ids, "sort_order")
}
