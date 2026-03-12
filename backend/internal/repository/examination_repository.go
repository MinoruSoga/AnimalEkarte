package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ExaminationRepository interface {
	FindAll(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, status *string, page, limit int) ([]model.Exam, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Exam, error)
	Create(ctx context.Context, exam *model.Exam) error
	Update(ctx context.Context, exam *model.Exam) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type examinationRepository struct {
	db *gorm.DB
}

func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

func (r *examinationRepository) FindAll(ctx context.Context, clinicID uuid.UUID, petID *uuid.UUID, ownerID *uuid.UUID, status *string, page, limit int) ([]model.Exam, int64, error) {
	var exams []model.Exam
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Exam{})
	if petID != nil {
		q = q.Where("exams.pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Joins("JOIN pets ON pets.id = exams.pet_id").Where("pets.owner_id = ?", *ownerID)
	}
	if status != nil {
		q = q.Where("exams.status = ?", *status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count exams")
	}
	if err := q.Preload("ExamType").Preload("Pet").Preload("Doctor").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find exams")
	}
	return exams, total, nil
}

func (r *examinationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	var exam model.Exam
	if err := r.db.WithContext(ctx).
		Preload("ExamType").
		Preload("Pet").
		Preload("Doctor").
		Preload("Items").
		First(&exam, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("exam", id.String())
		}
		return nil, apperrors.Wrap(err, "find exam by id")
	}
	return &exam, nil
}

func (r *examinationRepository) Create(ctx context.Context, exam *model.Exam) error {
	if err := r.db.WithContext(ctx).Create(exam).Error; err != nil {
		return apperrors.Wrap(err, "create exam")
	}
	return nil
}

func (r *examinationRepository) Update(ctx context.Context, exam *model.Exam) error {
	result := r.db.WithContext(ctx).Where("id = ?", exam.ID).Save(exam)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update exam")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", exam.ID.String())
	}
	return nil
}

func (r *examinationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Exam{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete exam")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", id.String())
	}
	return nil
}
