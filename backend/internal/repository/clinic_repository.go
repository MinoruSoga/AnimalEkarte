package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicRepository interface {
	FindAll(ctx context.Context) ([]model.Clinic, error)
	FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
	GetCompany(ctx context.Context) (*model.Company, error)
	Create(ctx context.Context, clinic *model.Clinic) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error)
	CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error)
}

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := r.db.WithContext(ctx).Order("name ASC").Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", "")
	}
	return clinics, nil
}

func (r *clinicRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	clinics := make([]model.Clinic, 0)
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.clinic_id = clinics.id").
		Where("staff_clinic_assignments.staff_id = ?", staffID).
		Order("clinics.name ASC").
		Find(&clinics).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("staff_id=%d", staffID))
	}
	return clinics, nil
}

func (r *clinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	var clinic model.Clinic
	err := r.db.WithContext(ctx).First(&clinic, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", id))
	}
	return &clinic, nil
}

func (r *clinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	var company model.Company
	err := r.db.WithContext(ctx).First(&company).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "company", "singleton")
	}
	return &company, nil
}

func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	err := r.db.WithContext(ctx).Create(clinic).Error
	if err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("clinic", clinic.Name)
		}
		return apperrors.FromGORM(err, "clinic", "")
	}
	return nil
}

func (r *clinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).Model(&model.Clinic{}).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Clinic{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *clinicRepository) CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Owner{}).
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "owner", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}

func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("clinic_id = ? AND deleted_at IS NULL", clinicID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "staff", fmt.Sprintf("clinic_id=%d", clinicID))
	}
	return count, nil
}
