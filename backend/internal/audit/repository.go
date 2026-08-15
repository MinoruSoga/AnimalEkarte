// Package audit owns audit-log validation, construction, and persistence.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the persistence boundary used by the audit service.
type Repository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	CreateTx(ctx context.Context, log *model.AuditLog) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository constructs the audit-log repository.
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create writes through the base database connection.
func (r *repository) Create(ctx context.Context, log *model.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", "")
	}
	return nil
}

// CreateTx participates in the caller's ambient persistence transaction.
// Missing transaction context fails closed so audit writes cannot escape the
// business transaction through the base database connection.
func (r *repository) CreateTx(ctx context.Context, log *model.AuditLog) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("audit log write requires an ambient transaction")
	}
	if err := tx.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", "")
	}
	return nil
}

// MarshalJSON serializes audit values best-effort. Invalid values are omitted
// so metadata serialization cannot suppress the core audit event.
// INF-06: marshal failures are logged (type only; never the value) so omission
// is observable without failing the audit write.
func MarshalJSON(value any) []byte {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		slog.Warn("audit value marshal failed; field omitted",
			"value_type", fmt.Sprintf("%T", value),
			"error", err,
		)
		return nil
	}
	return encoded
}
