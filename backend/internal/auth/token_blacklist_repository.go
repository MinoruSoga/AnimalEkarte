package auth

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// TokenBlacklistRepository provides global refresh-token JTI denylist persistence.
type TokenBlacklistRepository interface {
	Create(ctx context.Context, entry *model.TokenBlacklist) error
	ExistsByJTI(ctx context.Context, jti string) (bool, error)
	DeleteExpired(ctx context.Context) error
}

type tokenBlacklistRepository struct{ db *gorm.DB }

// NewTokenBlacklistRepository constructs global refresh-token JTI denylist persistence.
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Create(ctx context.Context, entry *model.TokenBlacklist) error {
	if err := repohelpers.DBOrTx(ctx, r.db).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", entry.JTI)
	}
	return nil
}

func (r *tokenBlacklistRepository) ExistsByJTI(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := repohelpers.DBOrTx(ctx, r.db).
		Model(&model.TokenBlacklist{}).
		Where("jti = ? AND expires_at > ?", jti, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "token_blacklist", jti)
	}
	return count > 0, nil
}

func (r *tokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	if err := repohelpers.DBOrTx(ctx, r.db).
		Where("expires_at <= ?", time.Now()).
		Delete(&model.TokenBlacklist{}).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", "expired")
	}
	return nil
}
