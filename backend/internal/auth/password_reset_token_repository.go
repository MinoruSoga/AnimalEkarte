package auth

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// PasswordResetTokenInvalidator is the narrow credential-change capability
// that revokes every outstanding reset token for one account.
type PasswordResetTokenInvalidator interface {
	DeleteByAccountID(ctx context.Context, accountID uint64) error
}

// PasswordResetTokenRepository provides global password-reset token persistence.
type PasswordResetTokenRepository interface {
	PasswordResetTokenInvalidator
	Create(ctx context.Context, token *model.PasswordResetToken) error
	FindLatestByAccountIDForUpdate(
		ctx context.Context,
		accountID uint64,
	) (*model.PasswordResetToken, error)
	FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	FindByTokenHashForUpdate(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	DeleteByID(ctx context.Context, id uint64) error
	DeleteIssued(ctx context.Context, id uint64, tokenHash string) error
	ConsumeByID(ctx context.Context, id uint64) error
}

type passwordResetTokenRepository struct{ db *gorm.DB }

// NewPasswordResetTokenRepository constructs global password-reset token persistence.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	db := silentPasswordResetTokenDB(ctx, r.db)
	if err := db.Create(token).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", "")
	}
	return nil
}

func (r *passwordResetTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	var token model.PasswordResetToken
	if err := silentPasswordResetTokenDB(ctx, r.db).
		Where("token_hash = ?", hash).
		First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"password_reset_token",
			"token_lookup",
		)
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) FindLatestByAccountIDForUpdate(
	ctx context.Context,
	accountID uint64,
) (*model.PasswordResetToken, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError(
			"password reset token lock requires an ambient transaction",
		)
	}
	var token model.PasswordResetToken
	if err := tx.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("created_at DESC").
		Order("id DESC").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"password_reset_token",
			fmt.Sprintf("account:%d", accountID),
		)
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) FindByTokenHashForUpdate(
	ctx context.Context,
	hash string,
) (*model.PasswordResetToken, error) {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return nil, apperrors.WrapInternalServerError("password reset token lock requires an ambient transaction")
	}
	var token model.PasswordResetToken
	if err := tx.WithContext(ctx).Session(&gorm.Session{
		Logger: tx.Logger.LogMode(logger.Silent),
	}).
		Where("token_hash = ?", hash).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&token).Error; err != nil {
		return nil, apperrors.FromGORM(
			err,
			"password_reset_token",
			"token_lookup",
		)
	}
	return &token, nil
}

func (r *passwordResetTokenRepository) DeleteByAccountID(ctx context.Context, accountID uint64) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError(
			"password reset token invalidation requires an ambient transaction",
		)
	}
	if err := tx.WithContext(ctx).
		Where("account_id = ?", accountID).
		Delete(&model.PasswordResetToken{}).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("account:%d", accountID))
	}
	return nil
}

func (r *passwordResetTokenRepository) DeleteByID(ctx context.Context, id uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Delete(&model.PasswordResetToken{}, id).Error; err != nil {
		return apperrors.FromGORM(err, "password_reset_token", fmt.Sprintf("%d", id))
	}
	return nil
}

// DeleteIssued removes only the exact token whose delivery failed. A no-op is
// successful because the token may already have been consumed or superseded.
func (r *passwordResetTokenRepository) DeleteIssued(
	ctx context.Context,
	id uint64,
	tokenHash string,
) error {
	db := silentPasswordResetTokenDB(ctx, r.db)
	if err := db.
		Where("id = ? AND token_hash = ?", id, tokenHash).
		Delete(&model.PasswordResetToken{}).Error; err != nil {
		return apperrors.FromGORM(
			err,
			"password_reset_token",
			fmt.Sprintf("%d", id),
		)
	}
	return nil
}

func silentPasswordResetTokenDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	db := persistence.DBOrTx(ctx, fallback)
	return db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	})
}

func (r *passwordResetTokenRepository) ConsumeByID(ctx context.Context, id uint64) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("password reset token consumption requires an ambient transaction")
	}
	result := tx.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.PasswordResetToken{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "password_reset_token", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapNotFound("password_reset_token", fmt.Sprintf("%d", id))
	}
	return nil
}
