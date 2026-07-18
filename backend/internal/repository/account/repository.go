// Package account is a thin domain split of the flat repository package.
// It owns accounts (login account) data access. accounts テーブルは clinic_id を
// 持たないグローバルなログインアカウントのため、clinicScope は使用しない
// （P4 例外対象、repository/CLAUDE.md 参照）。
package account

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository はアカウント管理のインターフェース
type Repository interface {
	FindByID(ctx context.Context, id uint64) (*model.Account, error)
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	Create(ctx context.Context, account *model.Account) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
}

// repository は Repository の実装
type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// FindByID はIDでアカウントを取得する
func (r *repository) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	var account model.Account
	if err := repohelpers.DBOrTx(ctx, r.db).First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", id))
	}
	return &account, nil
}

// FindByEmail はメールアドレスでアカウントを検索する
func (r *repository) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	if err := repohelpers.DBOrTx(ctx, r.db).First(&account, "email = ? AND deleted_at IS NULL", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", email)
		}
		return nil, apperrors.FromGORM(err, "account", email)
	}
	return &account, nil
}

// Create はアカウントを新規作成する
func (r *repository) Create(ctx context.Context, account *model.Account) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "create")
	}
	return nil
}

// Update はアカウントの指定フィールドのみを更新する（Save() による全フィールド上書き防止）。
func (r *repository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.Account{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "account", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
	}
	return nil
}
