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
