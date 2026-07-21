package lstep

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LineLinkTokenRepository は LINE 紐付けトークンのデータアクセスインターフェース（BE-021）。
type LineLinkTokenRepository interface {
	Create(ctx context.Context, t *model.LineLinkToken) error
	// FindByToken は未使用のトークンを検索する。
	FindByToken(ctx context.Context, token string) (*model.LineLinkToken, error)
	// MarkUsed はトークンを使用済みにする。
	MarkUsed(ctx context.Context, id uint64, usedAt time.Time) error
}

type lineLinkTokenRepository struct{ db *gorm.DB }

func NewLineLinkTokenRepository(db *gorm.DB) LineLinkTokenRepository {
	return &lineLinkTokenRepository{db: db}
}

func (r *lineLinkTokenRepository) Create(ctx context.Context, t *model.LineLinkToken) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return apperrors.FromGORM(err, "line_link_token", "")
	}
	return nil
}

func (r *lineLinkTokenRepository) FindByToken(ctx context.Context, token string) (*model.LineLinkToken, error) {
	var t model.LineLinkToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND used_at IS NULL AND expires_at > ?", token, time.Now()).
		First(&t).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_link_token", token)
	}
	return &t, nil
}

func (r *lineLinkTokenRepository) MarkUsed(ctx context.Context, id uint64, usedAt time.Time) error {
	err := r.db.WithContext(ctx).
		Model(&model.LineLinkToken{}).
		Where("id = ?", id).
		Update("used_at", usedAt).Error
	if err != nil {
		return apperrors.FromGORM(err, "line_link_token", "")
	}
	return nil
}
