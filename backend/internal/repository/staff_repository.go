// Package repository provides data access implementations for Staff entity.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, role *string) ([]model.Staff, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Staff, error)
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, role *string) ([]model.Staff, error) {
	var staffs []model.Staff
	q := r.db.WithContext(ctx).Model(&model.Staff{})
	if role != nil {
		q = q.Where("staff_role = ?", *role)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&staffs).Error; err != nil {
		return nil, apperrors.Wrap(err, "find staffs")
	}
	return staffs, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Staff, error) {
	var staff model.Staff
	if err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("staff", id.String())
		}
		return nil, apperrors.Wrap(err, "find staff by id")
	}
	return &staff, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if err := r.db.WithContext(ctx).Create(staff).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("staff", staff.Name)
		}
		return apperrors.Wrap(err, "create staff")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, staff *model.Staff) error {
	if err := r.db.WithContext(ctx).Save(staff).Error; err != nil {
		return apperrors.Wrap(err, "update staff")
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Staff{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete staff")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", id.String())
	}
	return nil
}
