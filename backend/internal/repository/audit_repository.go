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
	// CreateTx は ambient transaction（dbOrTx）に参加して監査ログを永続化する（#211）。
	// ctx に Transactor.WithTx のトランザクションがあれば同一 tx で書き込み、無ければ Create と
	// 同様に base db へ書く（後方互換）。tx 内監査（fail-closed）の利用者向けの書込経路。
	CreateTx(ctx context.Context, log *model.AuditLog) error
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

// CreateTx は dbOrTx 経由で監査ログを永続化する。Transactor.WithTx 内から ctx 付きで呼ぶと
// 同一トランザクションに参加し、caller が tx を rollback すれば監査書込も巻き戻る（#211）。
// ambient tx が無い場合は base db に書き込む（Create と等価＝後方互換）。
// audit_logs は P4 clinicScope 例外テーブルのため Scopes は付与しない（既存 Create と同方針）。
func (r *auditRepository) CreateTx(ctx context.Context, log *model.AuditLog) error {
	if err := dbOrTx(ctx, r.db).Create(log).Error; err != nil {
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
