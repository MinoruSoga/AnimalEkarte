package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// PasswordResetTokenRepository はパスワードリセットトークンのデータアクセスを定義する。
type PasswordResetTokenRepository interface {
	// Create はトークンを新規作成する。
	Create(ctx context.Context, token *model.PasswordResetToken) error
	// FindByTokenHash は SHA256 ハッシュでトークンを検索する。
	FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	// DeleteByAccountID はアカウントの既存トークンをすべて削除する（トークン発行前のクリーンアップ用）。
	DeleteByAccountID(ctx context.Context, accountID uint64) error
	// DeleteByID はトークンを物理削除する（使用済み後）。
	DeleteByID(ctx context.Context, id uint64) error
}

type passwordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository は PasswordResetTokenRepository の実装を返す。
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", "")
	}
	return nil
}

func (r *passwordResetTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(err, "password_reset_token", hash)
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) DeleteByAccountID(ctx context.Context, accountID uint64) error {
	if err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Delete(&model.PasswordResetToken{}).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("account:%d", accountID))
	}
	return nil
}

func (r *passwordResetTokenRepository) DeleteByID(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.PasswordResetToken{}, id).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("%d", id))
	}
	return nil
}
