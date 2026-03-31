package repository

import (
	"context"
	"encoding/json"
	"log/slog"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// AuditRepository は監査ログのデータアクセス層
type AuditRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
}

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository はAuditRepositoryを初期化して返す
func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		slog.WarnContext(ctx, "audit log write failed",
			slog.String("action", log.Action),
			slog.String("error", err.Error()))
		return err
	}
	return nil
}

// MarshalAuditJSON は監査ログ用に値をJSONバイト列にシリアライズするヘルパー。
// nil の場合は nil を返す。エラー時は nil を返す（監査ログはベストエフォート）。
func MarshalAuditJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
