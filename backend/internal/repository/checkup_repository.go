package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupFilters はクリニック横断一覧のフィルタ条件
type CheckupFilters struct {
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
}

type CheckupRepository interface {
	ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error)
	ListByClinic(ctx context.Context, clinicID uint64, filters CheckupFilters) ([]model.Checkup, error)
	FindByID(ctx context.Context, id uint64) (*model.Checkup, error)
	Create(ctx context.Context, checkup *model.Checkup) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
}

type checkupRepository struct {
	db *gorm.DB
}

func NewCheckupRepository(db *gorm.DB) CheckupRepository {
	return &checkupRepository{db: db}
}

func (r *checkupRepository) ListByClinic(ctx context.Context, clinicID uint64, filters CheckupFilters) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	q := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Preload("CheckupType").
		Preload("Doctor").
		Preload("MedicalRecord.Pet.Owner")
	if filters.StartDate != nil {
		q = q.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		q = q.Where("date <= ?", *filters.EndDate)
	}
	if filters.NextStartDate != nil {
		q = q.Where("next_date >= ?", *filters.NextStartDate)
	}
	if filters.NextEndDate != nil {
		q = q.Where("next_date <= ?", *filters.NextEndDate)
	}
	if err := q.Order("date DESC").Find(&checkups).Error; err != nil {
		return nil, apperrors.Wrap(err, "list checkups by clinic")
	}
	return checkups, nil
}

func (r *checkupRepository) ListByMedicalRecordID(ctx context.Context, medicalRecordID uint64) ([]model.Checkup, error) {
	checkups := make([]model.Checkup, 0)
	if err := r.db.WithContext(ctx).
		Where("medical_record_id = ?", medicalRecordID).
		Preload("CheckupType").
		Preload("Doctor").
		Order("date ASC").
		Find(&checkups).Error; err != nil {
		return nil, apperrors.Wrap(err, "list checkups by medical record id")
	}
	return checkups, nil
}

func (r *checkupRepository) FindByID(ctx context.Context, id uint64) (*model.Checkup, error) {
	var checkup model.Checkup
	if err := r.db.WithContext(ctx).
		Preload("CheckupType").
		Preload("Doctor").
		First(&checkup, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find checkup by id")
	}
	return &checkup, nil
}

func (r *checkupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	if err := r.db.WithContext(ctx).Create(checkup).Error; err != nil {
		return apperrors.Wrap(err, "create checkup")
	}
	return nil
}

func (r *checkupRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Checkup{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update checkup")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *checkupRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Checkup{}, id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete checkup")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", id))
	}
	return nil
}
