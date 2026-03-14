package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UserAccountRepository はユーザーアカウントのデータアクセス層
type UserAccountRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.UserAccount, error)
	FindByIDWithMemberships(ctx context.Context, id uint64) (*UserAccountWithMemberships, error)
	FindByClinicID(ctx context.Context, clinicID uint64) ([]model.UserAccount, error)
	Create(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
	Delete(ctx context.Context, id uint64) error
	FindPermissions(ctx context.Context, userID, clinicID uint64) ([]model.UserPermission, error)
	SetPermissions(ctx context.Context, userID, clinicID uint64, permissions []string) error
}

// UserAccountWithMemberships はユーザー・所属クリニック・権限をまとめたクエリ結果
type UserAccountWithMemberships struct {
	model.UserAccount
	Memberships []model.UserClinicMembership
	Permissions []model.UserPermission
}

type userAccountRepository struct {
	db *gorm.DB
}

// NewUserAccountRepository はUserAccountRepositoryを初期化して返す
func NewUserAccountRepository(db *gorm.DB) UserAccountRepository {
	return &userAccountRepository{db: db}
}

// FindByEmail はメールアドレスでユーザーアカウントを取得する。
// アカウントが存在しない場合は ErrNotFound を返す。
func (r *userAccountRepository) FindByEmail(ctx context.Context, email string) (*model.UserAccount, error) {
	var account model.UserAccount
	if err := r.db.WithContext(ctx).First(&account, "email = ? AND deleted_at IS NULL", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("user_account", email)
		}
		return nil, apperrors.Wrap(err, "find user account by email")
	}
	return &account, nil
}

// FindByClinicID はクリニックIDに紐づくユーザーアカウント一覧を返す。
func (r *userAccountRepository) FindByClinicID(ctx context.Context, clinicID uint64) ([]model.UserAccount, error) {
	var accounts []model.UserAccount
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_clinic_memberships ON user_clinic_memberships.user_id = user_accounts.id").
		Where("user_clinic_memberships.clinic_id = ? AND user_accounts.deleted_at IS NULL", clinicID).
		Find(&accounts).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user accounts by clinic id")
	}
	return accounts, nil
}

// Create はユーザーアカウントとクリニック所属を作成する。
func (r *userAccountRepository) Create(ctx context.Context, account *model.UserAccount, clinicID uint64, staffID *uint64, isMain bool) error {
	account.StaffID = staffID
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(account).Error; err != nil {
			return apperrors.Wrap(err, "create user account")
		}
		membership := &model.UserClinicMembership{
			UserID:   account.ID,
			ClinicID: clinicID,
			IsMain:   isMain,
		}
		if err := tx.Create(membership).Error; err != nil {
			return apperrors.Wrap(err, "create user clinic membership")
		}
		return nil
	})
}

// Update はユーザーアカウントを部分更新する。
func (r *userAccountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	if err := r.db.WithContext(ctx).Model(&model.UserAccount{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(fields).Error; err != nil {
		return apperrors.Wrap(err, "update user account")
	}
	return nil
}

// Delete はユーザーアカウントを論理削除する。
func (r *userAccountRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		Delete(&model.UserAccount{}).Error; err != nil {
		return apperrors.Wrap(err, "delete user account")
	}
	return nil
}

// FindPermissions はユーザーの権限一覧を返す。
func (r *userAccountRepository) FindPermissions(ctx context.Context, userID, clinicID uint64) ([]model.UserPermission, error) {
	var perms []model.UserPermission
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND clinic_id = ?", userID, clinicID).
		Find(&perms).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user permissions")
	}
	return perms, nil
}

// SetPermissions はユーザーの権限を全置換する（削除→追加のトランザクション）。
func (r *userAccountRepository) SetPermissions(ctx context.Context, userID, clinicID uint64, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND clinic_id = ?", userID, clinicID).
			Delete(&model.UserPermission{}).Error; err != nil {
			return apperrors.Wrap(err, "delete existing permissions")
		}
		for _, p := range permissions {
			perm := &model.UserPermission{
				UserID:     userID,
				ClinicID:   clinicID,
				Permission: model.PermissionType(p),
			}
			if err := tx.Create(perm).Error; err != nil {
				return apperrors.Wrap(err, "create permission")
			}
		}
		return nil
	})
}

// FindByIDWithMemberships はIDでユーザーを取得し、所属クリニック・権限も合わせて返す。
func (r *userAccountRepository) FindByIDWithMemberships(ctx context.Context, id uint64) (*UserAccountWithMemberships, error) {
	var account model.UserAccount
	if err := r.db.WithContext(ctx).
		Preload("Staff").
		Preload("JobTitle").
		First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("user_account", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find user account by id")
	}

	memberships := make([]model.UserClinicMembership, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Find(&memberships).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user clinic memberships")
	}

	permissions := make([]model.UserPermission, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Find(&permissions).Error; err != nil {
		return nil, apperrors.Wrap(err, "find user permissions")
	}

	return &UserAccountWithMemberships{
		UserAccount: account,
		Memberships: memberships,
		Permissions: permissions,
	}, nil
}
