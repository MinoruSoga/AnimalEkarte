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
	GetByID(ctx context.Context, id uint64) (*model.Account, error)
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	Create(ctx context.Context, account *model.Account) error
	Update(ctx context.Context, account *model.Account) error
	Delete(ctx context.Context, id uint64) error
}

// accountRepository は AccountRepository の実装
type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository はAccountRepositoryを初期化して返す
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// GetByID はIDでアカウントを取得する
func (r *accountRepository) GetByID(ctx context.Context, id uint64) (*model.Account, error) {
	var account model.Account
	if err := r.db.WithContext(ctx).First(&account, "id = ? AND deleted_at IS NULL", id).Error; err != nil {
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
	if err := r.db.WithContext(ctx).First(&account, "email = ? AND deleted_at IS NULL", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("account", email)
		}
		return nil, apperrors.FromGORM(err, "account", email)
	}
	return &account, nil
}

// Create はアカウントを新規作成する
func (r *accountRepository) Create(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "create")
	}
	return nil
}

// Update はアカウント情報を更新する
func (r *accountRepository) Update(ctx context.Context, account *model.Account) error {
	if err := r.db.WithContext(ctx).Save(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", fmt.Sprintf("%d", account.ID))
	}
	return nil
}

// Delete はアカウントを論理削除する
func (r *accountRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Model(&model.Account{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("now()")).Error; err != nil {
		return apperrors.FromGORM(err, "account", fmt.Sprintf("%d", id))
	}
	return nil
}
