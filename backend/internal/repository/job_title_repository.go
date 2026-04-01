// Package repository provides data access implementations for JobTitle entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- JobTitle ----

type JobTitleRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.JobTitle, error)
	FindByID(ctx context.Context, id uint64) (*model.JobTitle, error)
	Create(ctx context.Context, jobTitle *model.JobTitle) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountStaffsByJobTitleID(ctx context.Context, jobTitleID uint64) (int64, error)
}

type jobTitleRepository struct{ db *gorm.DB }

func NewJobTitleRepository(db *gorm.DB) JobTitleRepository {
	return &jobTitleRepository{db: db}
}

func (r *jobTitleRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.JobTitle, error) {
	jobTitles := make([]model.JobTitle, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("sort_order ASC, name ASC").
		Find(&jobTitles).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "job_title", "")
	}
	return jobTitles, nil
}

func (r *jobTitleRepository) FindByID(ctx context.Context, id uint64) (*model.JobTitle, error) {
	var jobTitle model.JobTitle
	err := r.db.WithContext(ctx).First(&jobTitle, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "job_title", fmt.Sprintf("%d", id))
	}
	return &jobTitle, nil
}

func (r *jobTitleRepository) Create(ctx context.Context, jobTitle *model.JobTitle) error {
	err := r.db.WithContext(ctx).Create(jobTitle).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("job_title", jobTitle.Name)
		}
		return apperrors.FromGORM(err, "job_title", "")
	}
	return nil
}

func (r *jobTitleRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.JobTitle{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "job_title", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("job_title", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *jobTitleRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.JobTitle{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "job_title", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("job_title", fmt.Sprintf("%d", id))
	}
	return nil
}

// CountStaffsByJobTitleID は指定役職を参照しているスタッフ数を返す（BUG-112）
func (r *jobTitleRepository) CountStaffsByJobTitleID(ctx context.Context, jobTitleID uint64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("job_title_id = ?", jobTitleID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "staff", "")
	}
	return count, nil
}

func (r *jobTitleRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.JobTitle{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "job_title", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("job_title id %d not found in this clinic", id))
			}
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "reorder job title")
	}
	return nil
}
