package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TokenBlacklistRepository は refresh_token JTI ブラックリストのデータアクセスを定義する。
type TokenBlacklistRepository interface {
	// Create は JTI をブラックリストに追加する。
	Create(ctx context.Context, entry *model.TokenBlacklist) error
	// ExistsByJTI は指定 JTI がブラックリストに存在するかを返す。
	// 期限切れエントリは存在しないものとして扱う。
	ExistsByJTI(ctx context.Context, jti string) (bool, error)
	// DeleteExpired は expires_at が現在時刻より古いエントリを物理削除する（定期メンテ用）。
	DeleteExpired(ctx context.Context) error
}

type tokenBlacklistRepository struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository は TokenBlacklistRepository の実装を返す。
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Create(ctx context.Context, entry *model.TokenBlacklist) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", entry.JTI)
	}
	return nil
}

func (r *tokenBlacklistRepository) ExistsByJTI(ctx context.Context, jti string) (bool, error) {
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

func (r *tokenBlacklistRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&model.TokenBlacklist{}).Error; err != nil {
		return apperrors.FromGORM(err, "token_blacklist", "expired")
	}
	return nil
}
