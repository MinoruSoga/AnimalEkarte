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
	FindAll(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	// CreateWithAccount はスタッフ・ユーザーアカウント・クリニック所属を単一トランザクションで作成する。
	CreateWithAccount(ctx context.Context, staff *model.Staff, account *model.UserAccount, membership *model.UserClinicMembership) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error) {
	staffs := make([]model.Staff, 0)
	var total int64

	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Staff{}).Where("clinic_id = ?", clinicID)
		if role != nil {
			q = q.Where("staff_role = ?", *role)
		}
		return q
	}

	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count staffs")
	}
	if err := buildBase().
		Offset((page - 1) * limit).Limit(limit).
		Order("sort_order ASC, name ASC").
		Find(&staffs).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find staffs")
	}
	return staffs, total, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &staff, nil
}

func (r *staffRepository) CreateWithAccount(ctx context.Context, staff *model.Staff, account *model.UserAccount, membership *model.UserClinicMembership) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(staff).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return apperrors.WrapAlreadyExists("staff", staff.Name)
			}
			return apperrors.Wrap(err, "create staff")
		}

		account.StaffID = &staff.ID
		if err := tx.Create(account).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return apperrors.WrapAlreadyExists("user_account", account.Email)
			}
			return apperrors.Wrap(err, "create user account")
		}

		membership.UserID = account.ID
		if err := tx.Create(membership).Error; err != nil {
			return apperrors.Wrap(err, "create user clinic membership")
		}

		return nil
	}); err != nil {
		return apperrors.Wrap(err, "create staff with account transaction")
	}
	return nil
}

func (r *staffRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	result := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("id = ? AND clinic_id = ?", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update staff")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.Staff{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete staff")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *staffRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(&model.Staff{}).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.Wrap(result.Error, "reorder staff")
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("staff id %d not found in this clinic", id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "reorder staff")
	}
	return nil
}
