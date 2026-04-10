// Package repository provides data access implementations for Staff entity.
package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	// Create はスタッフを作成する。
	Create(ctx context.Context, staff *model.Staff) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		// staffs テーブルに clinic_id は存在しない。
		// staff_clinic_assignments を経由して clinic_id でフィルタ
		q := r.db.WithContext(ctx).Model(&model.Staff{}).
			Joins("INNER JOIN staff_clinic_assignments ON staff_clinic_assignments.staff_id = staffs.id").
			Where("staff_clinic_assignments.clinic_id = ?", clinicID).
			Where("staffs.deleted_at IS NULL")
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	if err := buildBase().
		Preload("Account").
		Preload("Occupation").
		Offset((page - 1) * limit).Limit(limit).
		Order("staffs.sort_order ASC, staffs.name ASC").
		Distinct("staffs.*").
		Find(&staffs).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "staff", "")
	}
	return staffs, total, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).Preload("Account").Preload("Occupation").First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

func (r *staffRepository) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).First(&staff, "account_id = ?", accountID).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("account_id=%d", accountID))
	}
	return &staff, nil
}

func (r *staffRepository) Create(ctx context.Context, staff *model.Staff) error {
	if err := r.db.WithContext(ctx).Create(staff).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("staff", staff.Name)
		}
		return apperrors.FromGORM(err, "staff", "")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// staffs テーブルに clinic_id は存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタ
	result := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("staffs.id = ?", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ?)", clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// staffs テーブルに clinic_id は存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタ
	result := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("id = ?", id).
		Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ?)", clinicID).
		Update("deleted_at", gorm.Expr("now()"))
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	// staffs テーブルに clinic_id カラムは存在しない。
	// staff_clinic_assignments を経由して clinic_id でフィルタする。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Staff{}).
				Where("staffs.id = ?", id).
				Where("EXISTS (SELECT 1 FROM staff_clinic_assignments WHERE staff_clinic_assignments.staff_id = staffs.id AND staff_clinic_assignments.clinic_id = ?)", clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("staff id %d not found in this clinic", id))
			}
		}
		return nil
	})
}
