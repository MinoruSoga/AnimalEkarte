package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// PasswordResetTokenRepository provides global password-reset token persistence.
type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	DeleteByAccountID(ctx context.Context, accountID uint64) error
	DeleteByID(ctx context.Context, id uint64) error
}

type passwordResetTokenRepository struct{ db *gorm.DB }

// NewPasswordResetTokenRepository constructs global password-reset token persistence.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(token).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", "")
	}
	return nil
}

func (r *passwordResetTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	if err := repohelpers.DBOrTx(ctx, r.db).Where("token_hash = ?", hash).First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(err, "password_reset_token", hash)
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) DeleteByAccountID(ctx context.Context, accountID uint64) error {
	if err := repohelpers.DBOrTx(ctx, r.db).
		Where("account_id = ?", accountID).
		Delete(&model.PasswordResetToken{}).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("account:%d", accountID))
	}
	return nil
}

func (r *passwordResetTokenRepository) DeleteByID(ctx context.Context, id uint64) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Delete(&model.PasswordResetToken{}, id).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("%d", id))
	}
	return nil
}
