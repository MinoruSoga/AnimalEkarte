// Package audit owns audit-log validation, construction, and persistence.
package audit

import (
	"context"
	"encoding/json"

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

// CreateTx participates in an ambient persistence transaction when one exists.
func (r *repository) CreateTx(ctx context.Context, log *model.AuditLog) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(log).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", "")
	}
	return nil
}

// MarshalJSON serializes audit values best-effort. Invalid values are omitted
// so metadata serialization cannot suppress the core audit event.
func MarshalJSON(value any) []byte {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}
