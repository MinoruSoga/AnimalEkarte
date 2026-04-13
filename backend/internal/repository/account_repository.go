package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountRepository はアカウント管理のインターフェース
type AccountRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.Account, error)
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	Create(ctx context.Context, account *model.Account) error
	Update(ctx context.Context, id uint64, fields map[string]any) error
}

// accountRepository は AccountRepository の実装
type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository はAccountRepositoryを初期化して返す
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// FindByID はIDでアカウントを取得する
func (r *accountRepository) FindByID(ctx context.Context, id uint64) (*model.Account, error) {
	var account model.Account
	if err := dbOrTx(ctx, r.db).First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", id))
	}
	return &account, nil
}

// FindByEmail はメールアドレスでアカウントを検索する
func (r *accountRepository) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	if err := dbOrTx(ctx, r.db).First(&account, "email = ? AND deleted_at IS NULL", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", email)
		}
		return nil, apperrors.FromGORM(err, "account", email)
	}
	return &account, nil
}

// Create はアカウントを新規作成する
func (r *accountRepository) Create(ctx context.Context, account *model.Account) error {
	if err := dbOrTx(ctx, r.db).Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "create")
	}
	return nil
}

// Update はアカウントの指定フィールドのみを更新する（Save() による全フィールド上書き防止）。
func (r *accountRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	result := dbOrTx(ctx, r.db).
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
