package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ExaminationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	Create(ctx context.Context, exam *model.Examination) error
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error)
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
			Where("exams.clinic_id = ?", clinicID)
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
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}

	exams := make([]model.Examination, 0)
	if err := buildBase().Preload("ExaminationType").Preload("Pet.Owner").Preload("Doctor").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}
	return exams, total, nil
}

func (r *examinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	var exam model.Examination
	err := r.db.WithContext(ctx).
		Where("exams.id = ? AND exams.clinic_id = ?", id, clinicID).
		Preload("ExaminationType").Preload("Pet.Owner").Preload("Doctor").Preload("Items").
		First(&exam).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", id))
	}
	return &exam, nil
}

func (r *examinationRepository) Create(ctx context.Context, exam *model.Examination) error {
	err := r.db.WithContext(ctx).Create(exam).Error
	if err != nil {
		return apperrors.FromGORM(err, "exam", "")
	}
	return nil
}

func (r *examinationRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *examinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Delete(&model.Examination{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *examinationRepository) CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exam_results.exam_id = exams.id").
		Where("exams.clinic_id = ? AND exam_results.exam_id = ?", clinicID, examID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
	}
	return count, nil
}
