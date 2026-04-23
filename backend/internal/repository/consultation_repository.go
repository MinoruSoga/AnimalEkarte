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
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type consultationRepository struct{ db *gorm.DB }

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Order("sort_order ASC, name ASC").Find(&consultations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", "")
	}
	return consultations, nil
}

func (r *consultationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	var consultation model.Consultation
	err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&consultation).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", id))
	}
	return &consultation, nil
}

func (r *consultationRepository) Create(ctx context.Context, consultation *model.Consultation) error {
	err := r.db.WithContext(ctx).Create(consultation).Error
	if err != nil {
		return apperrors.FromGORM(err, "consultation", "")
	}
	return nil
}

func (r *consultationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Consultation, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
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
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Consultation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("consultation", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *consultationRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return reorderByClinicID(ctx, r.db, &model.Consultation{}, "consultation", clinicID, ids)
}

// CountChildrenByParentID は指定した親 ID を持つ子診察項目の件数を返す。
// 親を削除する前に孤立子が残らないことを確認するために使用する。
func (r *consultationRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Consultation{}).
		Scopes(clinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

// CountUsageByConsultationID は診察マスタを参照している treatments の件数を返す（BUG-107）
// treatments テーブルに直接 clinic_id がないため medical_records を JOIN してテナント分離する
func (r *consultationRepository) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Treatment{}).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).
		Where("treatments.consultation_id = ? AND treatments.deleted_at IS NULL", consultationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	return count, nil
}
