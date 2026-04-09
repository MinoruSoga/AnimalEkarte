// Package repository provides data access implementations for Consultation entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Consultation ----

type ConsultationRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByConsultationID(ctx context.Context, consultationID uint64) (int64, error)
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).Order("sort_order ASC, name ASC").Find(&consultations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", "")
	}
	return consultations, nil
}

func (r *consultationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	var consultation model.Consultation
	err := r.db.WithContext(ctx).First(&consultation, "id = ? AND clinic_id = ?", id, clinicID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", id))
	}
	return &consultation, nil
}

func (r *consultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	err := r.db.WithContext(ctx).Create(consultation).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapConflict("同じ名称が既に登録されています")
		}
		return apperrors.FromGORM(err, "consultation", "")
	}
	return nil
}

func (r *consultationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *consultationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Consultation{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *consultationRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Consultation{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("consultation id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// CountUsageByConsultationID は診察マスタを参照している treatments の件数を返す（BUG-107）
func (r *consultationRepository) CountUsageByConsultationID(ctx context.Context, consultationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Where("consultation_id = ?", consultationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	return count, nil
}
