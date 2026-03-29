package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AuthRepository はトークン管理のデータアクセスインターフェース
type AuthRepository interface {
	// リフレッシュトークン
	CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserTokens(ctx context.Context, userID uint64) error

	// パスワードリセットトークン
	CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error
	FindPasswordResetToken(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error
}

type authRepository struct {
	db *gorm.DB
}

// NewAuthRepository はAuthRepositoryを初期化して返す
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("refresh_token", tokenHash)
		}
		return nil, apperrors.Wrap(err, "find refresh token")
	}
	return &token, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", now).Error
}

func (r *authRepository) RevokeAllUserTokens(ctx context.Context, userID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *authRepository) CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) FindPasswordResetToken(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("password_reset_token", tokenHash)
		}
		return nil, apperrors.Wrap(err, "find password reset token")
	}
	return &token, nil
}

func (r *authRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.PasswordResetToken{}).
		Where("token_hash = ? AND used_at IS NULL", tokenHash).
		Update("used_at", now).Error
}
