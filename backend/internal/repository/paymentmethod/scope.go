package paymentmethod

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// Local copies of parent-package clinic CRUD helpers. Kept private so this
// subpackage does not import repository (which would create an import cycle
// via repositories.go). A later repohelpers extraction will dedupe.

func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinic_id = ?", clinicID)
	}
}

func updateScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64, fields map[string]any) error {
	result := db.WithContext(ctx).
		Model(m).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, resource, fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound(resource, fmt.Sprintf("%d", id))
	}
	return nil
}

func deleteScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64) error {
	result := db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(m)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, resource, fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound(resource, fmt.Sprintf("%d", id))
	}
	return nil
}

func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64, orderColumn string) error {
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(model).
				Scopes(clinicScope(clinicID)).Where("id = ?", id).
				Update(orderColumn, i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, resource, fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("%s id %d not found in this clinic", resource, id))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder "+resource)
	}
	return nil
}
