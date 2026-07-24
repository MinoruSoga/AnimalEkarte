package repohelpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// MaxMasterListRows delegates to persistence.
const MaxMasterListRows = persistence.MaxMasterListRows

// ClinicScope delegates to persistence.
func ClinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return persistence.ClinicScope(clinicID)
}

// MedicalRecordTenantScope delegates to persistence.
func MedicalRecordTenantScope(childTable string, clinicID uint64) func(*gorm.DB) *gorm.DB {
	return persistence.MedicalRecordTenantScope(childTable, clinicID)
}

// ClinicScopeIn delegates to persistence.
func ClinicScopeIn(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return persistence.ClinicScopeIn(clinicIDs)
}

// FindByIDScoped delegates to persistence.
func FindByIDScoped[T any](ctx context.Context, db *gorm.DB, resource string, clinicID, id uint64) (*T, error) {
	return persistence.FindByIDScoped[T](ctx, db, resource, clinicID, id)
}

// UpdateScopedByID delegates to persistence.
func UpdateScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, db, m, resource, clinicID, id, fields)
}

// DeleteScopedByID delegates to persistence.
func DeleteScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64) error {
	return persistence.DeleteScopedByID(ctx, db, m, resource, clinicID, id)
}

// ReorderByClinicID delegates to persistence.
func ReorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64, orderColumn string) error {
	return persistence.ReorderByClinicID(ctx, db, model, resource, clinicID, ids, orderColumn)
}

// ReorderGlobal delegates to persistence.
func ReorderGlobal(ctx context.Context, db *gorm.DB, model any, resource string, ids []uint64, orderColumn string) error {
	return persistence.ReorderGlobal(ctx, db, model, resource, ids, orderColumn)
}

// Paginate delegates to persistence.
func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return persistence.Paginate(page, limit)
}
