// Package repository provides data access implementations for Staff entity.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Staff ----

type StaffRepository interface {
	FindAll(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error)
	FindByID(ctx context.Context, id uint64) (*model.Staff, error)
	// CreateWithAccount はスタッフ・ユーザーアカウント・クリニック所属を単一トランザクションで作成する。
	CreateWithAccount(ctx context.Context, staff *model.Staff, account *model.UserAccount, membership *model.UserClinicMembership) error
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type staffRepository struct{ db *gorm.DB }

func NewStaffRepository(db *gorm.DB) StaffRepository { return &staffRepository{db: db} }

func (r *staffRepository) FindAll(ctx context.Context, clinicID uint64, role *string) ([]model.Staff, error) {
	staffs := make([]model.Staff, 0)
	q := r.db.WithContext(ctx).Model(&model.Staff{}).Where("clinic_id = ?", clinicID)
	if role != nil {
		q = q.Where("staff_role = ?", *role)
	}
	if err := q.Order("sort_order ASC, name ASC").Find(&staffs).Error; err != nil {
		return nil, apperrors.Wrap(err, "find staffs")
	}
	return staffs, nil
}

func (r *staffRepository) FindByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var staff model.Staff
	if err := r.db.WithContext(ctx).First(&staff, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find staff by id")
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

func (r *staffRepository) Update(ctx context.Context, staff *model.Staff) error {
	result := r.db.WithContext(ctx).
		Model(&model.Staff{}).
		Where("id = ? AND clinic_id = ?", staff.ID, staff.ClinicID).
		Updates(staff)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update staff")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staff.ID))
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
