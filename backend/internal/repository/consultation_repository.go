// Package repository provides data access implementations for Consultation entity.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation ----

type ConsultationRepository interface {
	FindAll(ctx context.Context) ([]model.Consultation, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, consultation *model.Consultation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context) ([]model.Consultation, error) {
	var consultations []model.Consultation
	if err := r.db.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&consultations).Error; err != nil {
		return nil, apperrors.Wrap(err, "find consultations")
	}
	return consultations, nil
}

func (r *consultationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Consultation, error) {
	var consultation model.Consultation
	if err := r.db.WithContext(ctx).First(&consultation, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("consultation", id.String())
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
	if err := r.db.WithContext(ctx).Save(consultation).Error; err != nil {
		return apperrors.Wrap(err, "update consultation")
	}
	return nil
}

func (r *consultationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Consultation{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete consultation")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("consultation", id.String())
	}
	return nil
}
