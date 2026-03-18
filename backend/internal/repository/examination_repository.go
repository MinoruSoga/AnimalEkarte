package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ExaminationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	Create(ctx context.Context, exam *model.Examination) error
	Update(ctx context.Context, clinicID uint64, exam *model.Examination) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type examinationRepository struct {
	db *gorm.DB
}

func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

func (r *examinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Examination{}).
			Joins("JOIN medical_records ON medical_records.id = exams.medical_record_id").
			Where("medical_records.clinic_id = ?", clinicID)
		if petID != nil {
			q = q.Where("exams.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = exams.pet_id").Where("pets.owner_id = ?", *ownerID)
		}
		if status != nil {
			q = q.Where("exams.status = ?", *status)
		}
		if startDate != nil {
			q = q.Where("exams.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("exams.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count exams")
	}

	exams := make([]model.Examination, 0)
	if err := buildBase().Preload("ExaminationType").Preload("Pet").Preload("Doctor").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find exams")
	}
	return exams, total, nil
}

func (r *examinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	var exam model.Examination
	if err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = exams.medical_record_id").
		Where("exams.id = ? AND medical_records.clinic_id = ?", id, clinicID).
		Preload("ExaminationType").Preload("Pet").Preload("Doctor").Preload("Items").
		First(&exam).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find exam by id")
	}
	return &exam, nil
}

func (r *examinationRepository) Create(ctx context.Context, exam *model.Examination) error {
	if err := r.db.WithContext(ctx).Create(exam).Error; err != nil {
		return apperrors.Wrap(err, "create exam")
	}
	return nil
}

func (r *examinationRepository) Update(ctx context.Context, clinicID uint64, exam *model.Examination) error {
	result := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Where("id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", exam.ID, clinicID).
		Updates(exam)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update exam")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", exam.ID))
	}
	return nil
}

func (r *examinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", id, clinicID).
		Delete(&model.Examination{})
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete exam")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return nil
}
