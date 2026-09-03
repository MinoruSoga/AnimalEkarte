package billing

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// Shared clinic helpers live in repohelpers so domain subpackages do not fork
// clinicScope / reorder copies (and so Reorder joins ambient WithTx via DBOrTx).

func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return persistence.ClinicScope(clinicID)
}

func updateScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, db, m, resource, clinicID, id, fields)
}

func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64, orderColumn string) error {
	return persistence.ReorderByClinicID(ctx, db, model, resource, clinicID, ids, orderColumn)
}
