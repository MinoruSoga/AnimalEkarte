// Package passwordreset owns password_reset_tokens data access.
package passwordreset

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Repository is the data access interface for password reset tokens.
type Repository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	DeleteByAccountID(ctx context.Context, accountID uint64) error
	DeleteByID(ctx context.Context, id uint64) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", "")
	}
	return nil
}

func (r *repository) FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(err, "password_reset_token", hash)
	}
	return &token, nil
}

func (r *repository) DeleteByAccountID(ctx context.Context, accountID uint64) error {
	if err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Delete(&model.PasswordResetToken{}).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("account:%d", accountID))
	}
	return nil
}

func (r *repository) DeleteByID(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.PasswordResetToken{}, id).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("%d", id))
	}
	return nil
}
