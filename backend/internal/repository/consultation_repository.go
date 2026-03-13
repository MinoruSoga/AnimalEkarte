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
	FindAll(ctx context.Context) ([]model.Consultation, error)
	FindByID(ctx context.Context, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uint64) error
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&consultations).Error; err != nil {
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

func (r *consultationRepository) Update(ctx context.Context, consultation *model.Consultation) error {
	result := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Where("id = ? AND clinic_id = ?", consultation.ID, consultation.ClinicID).
		Updates(consultation)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update consultation")
	}
	if result.RowsAffected == 0 {
		return apperrors.Wrap(apperrors.ErrNotFound, "update consultation")
	}
	return nil
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
