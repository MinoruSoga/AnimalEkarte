package repository

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
		return apperrors.FromGORM(err, "audit_log", "")
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
