// Package repository provides data access implementations for Consultation entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation ----

type ConsultationRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	FindByID(ctx context.Context, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	if err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&consultations).Error; err != nil {
		return nil, apperrors.Wrap(err, "find consultations")
	}
	return consultations, nil
}

func (r *consultationRepository) FindByID(ctx context.Context, id uint64) (*model.Consultation, error) {
	var consultation model.Consultation
	if err := r.db.WithContext(ctx).First(&consultation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find consultation by id")
	}
	return &consultation, nil
}

func (r *consultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	if err := r.db.WithContext(ctx).Create(consultation).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("consultation", consultation.Name)
		}
		return apperrors.Wrap(err, "create consultation")
	}
	return nil
}

func (r *consultationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "update consultation")
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, id)
}

func (r *consultationRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Consultation{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete consultation")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *consultationRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Consultation{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder consultation")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("consultation id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("reorder consultation: %w", err)
	}
	return nil
}
