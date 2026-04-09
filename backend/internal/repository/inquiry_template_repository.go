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
	Create(ctx context.Context, template *model.InquiryTemplate) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type inquiryTemplateRepository struct{ db *gorm.DB }

func NewInquiryTemplateRepository(db *gorm.DB) InquiryTemplateRepository {
	return &inquiryTemplateRepository{db: db}
}

func (r *inquiryTemplateRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.InquiryTemplate, error) {
	templates := make([]model.InquiryTemplate, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, title ASC").
		Find(&templates).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", "")
	}
	return templates, nil
}

func (r *inquiryTemplateRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InquiryTemplate, error) {
	var template model.InquiryTemplate
	err := r.db.WithContext(ctx).First(&template, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "inquiry_template", fmt.Sprintf("%d", id))
	}
	return &template, nil
}

func (r *inquiryTemplateRepository) Create(ctx context.Context, template *model.InquiryTemplate) error {
	err := r.db.WithContext(ctx).Create(template).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("inquiry_template", template.Title)
		}
		return apperrors.FromGORM(err, "inquiry_template", "")
	}
	return nil
}

func (r *inquiryTemplateRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.InquiryTemplate{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inquiry_template", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inquiry_template", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *inquiryTemplateRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.InquiryTemplate{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "inquiry_template", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("inquiry_template", fmt.Sprintf("%d", id))
	}
	return nil
}
