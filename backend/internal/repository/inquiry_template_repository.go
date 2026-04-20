// Package repository provides data access implementations for InquiryTemplate entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- InquiryTemplate ----

type InquiryTemplateRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error)
	CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error)
	Create(ctx context.Context, template *model.InquiryTemplate) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type inquiryTemplateRepository struct{ db *gorm.DB }

func NewInquiryTemplateRepository(db *gorm.DB) InquiryTemplateRepository {
	return &inquiryTemplateRepository{db: db}
}

func (r *inquiryTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	templates := make([]model.InquiryTemplate, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("sort_order ASC, title ASC").
		Find(&templates).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", "")
	}
	return templates, nil
}

func (r *inquiryTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	var template model.InquiryTemplate
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", fmt.Sprintf("%d", id))
	}
	return &template, nil
}

func (r *inquiryTemplateRepository) Create(ctx context.Context, template *model.InquiryTemplate) error {
	err := r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "inquiry_template", "")
	}
	return nil
}

func (r *inquiryTemplateRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.InquiryTemplate, error) {
	result := r.db.WithContext(ctx).
		Model(&model.InquiryTemplate{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "inquiry_template", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("inquiry_template", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *inquiryTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.InquiryTemplate{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inquiry_template", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inquiry_template", fmt.Sprintf("%d", id))
	}
	return nil
}

// TODO: inquiry_answers テーブル追加時に以下の COUNT クエリを実装すること。
// 現スキーマには inquiry_template_id FK を持つテーブルが存在しないため常に 0 を返す。
func (r *inquiryTemplateRepository) CountUsageByInquiryTemplateID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (r *inquiryTemplateRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.InquiryTemplate{}, "inquiry_template", clinicID, ids)
}
