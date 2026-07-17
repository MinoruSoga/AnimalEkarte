// Package tokenblacklist owns token_blacklist data access (refresh JTI denylist).
package tokenblacklist

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// Repository is the data access interface for token blacklist entries.
type Repository interface {
	Create(ctx context.Context, entry *model.TokenBlacklist) error
	ExistsByJTI(ctx context.Context, jti string) (bool, error)
	DeleteExpired(ctx context.Context) error
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, entry *model.TokenBlacklist) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", entry.JTI)
	}
	return nil
}

func (r *repository) ExistsByJTI(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TokenBlacklist{}).
		Where("jti = ? AND expires_at > ?", jti, time.Now()).
		Count(&count).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "token_blacklist", jti)
	}
	return count > 0, nil
}

func (r *repository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&model.TokenBlacklist{}).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", "expired")
	}
	return nil
}
