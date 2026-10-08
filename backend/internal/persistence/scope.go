package persistence

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// MaxMasterListRows is the safety cap for master list queries.
const MaxMasterListRows = 10000

// ClinicScope filters rows by clinic_id on the primary table.
func ClinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinic_id = ?", clinicID)
	}
}

// MedicalRecordTenantScope joins clinic-less children through medical_records.
// childTable must be a compile-time literal.
func MedicalRecordTenantScope(
	childTable string,
	clinicID uint64,
) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(
			"JOIN medical_records ON medical_records.id = "+
				childTable+
				".medical_record_id AND medical_records.clinic_id = ? "+
				"AND medical_records.deleted_at IS NULL",
			clinicID,
		)
	}
}

// BillingTenantScope joins clinic-less billing children through billings.
// Some billing child tables keep clinic_id only for provenance rows
// (e.g. billing_items.clinic_id is set solely for vaccination/exam provenance),
// so tenant correlation must be enforced via the parent billings row.
// childTable must be a compile-time literal.
func BillingTenantScope(
	childTable string,
	clinicIDs []uint64,
) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins(
			"JOIN billings ON billings.id = "+
				childTable+
				".billing_id AND billings.clinic_id IN ? "+
				"AND billings.deleted_at IS NULL",
			clinicIDs,
		)
	}
}

// ClinicScopeIn filters rows by clinic_id IN (...). Empty input yields no rows.
func ClinicScopeIn(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinic_id IN ?", clinicIDs)
	}
}

// FindByIDScoped loads a row by ID under clinic scope.
func FindByIDScoped[T any](
	ctx context.Context,
	db *gorm.DB,
	resource string,
	clinicID uint64,
	id uint64,
) (*T, error) {
	var record T
	if err := db.WithContext(ctx).
		Scopes(ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&record).Error; err != nil {
		return nil, apperrors.FromGORM(err, resource, fmt.Sprintf("%d", id))
	}
	return &record, nil
}

// UpdateScopedByID updates fields for an ID under clinic scope.
func UpdateScopedByID(
	ctx context.Context,
	db *gorm.DB,
	model any,
	resource string,
	clinicID uint64,
	id uint64,
	fields map[string]any,
) error {
	result := db.WithContext(ctx).
		Model(model).
		Scopes(ClinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(
			result.Error,
			resource,
			fmt.Sprintf("%d", id),
		)
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound(resource, fmt.Sprintf("%d", id))
	}
	return nil
}

// DeleteScopedByID soft-deletes a row by ID under clinic scope.
func DeleteScopedByID(
	ctx context.Context,
	db *gorm.DB,
	model any,
	resource string,
	clinicID uint64,
	id uint64,
) error {
	result := db.WithContext(ctx).
		Scopes(ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(model)
	if result.Error != nil {
		return apperrors.FromGORM(
			result.Error,
			resource,
			fmt.Sprintf("%d", id),
		)
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound(resource, fmt.Sprintf("%d", id))
	}
	return nil
}

// reorderTarget pairs an id with its requested 1-based order value.
type reorderTarget struct {
	id    uint64
	order int
}

// sortedReorderTargets pairs each id with its requested order value, then sorts
// the pairs by id ascending. Row locks are still taken one UPDATE at a time, so
// applying them in id order matches the ascending FOR UPDATE order used by
// cage merge/unmerge (cage_repository.go) and keeps lock acquisition
// deterministic across writers. The written order values still follow the
// client-supplied sequence.
func sortedReorderTargets(ids []uint64) []reorderTarget {
	targets := make([]reorderTarget, len(ids))
	for index, id := range ids {
		targets[index] = reorderTarget{id: id, order: index + 1}
	}
	// Stable: duplicate ids keep input order so the last write still wins,
	// identical to the previous sequential behavior.
	slices.SortStableFunc(targets, func(a, b reorderTarget) int {
		return cmp.Compare(a.id, b.id)
	})
	return targets
}

// ReorderByClinicID updates orderColumn for IDs under clinic scope.
// UPDATEs are applied in ascending id order (not the client's display order) so
// row-lock acquisition order matches other writers — the assigned order values
// still follow the client-supplied sequence.
func ReorderByClinicID(
	ctx context.Context,
	db *gorm.DB,
	model any,
	resource string,
	clinicID uint64,
	ids []uint64,
	orderColumn string,
) error {
	if err := DBOrTx(ctx, db).Transaction(func(tx *gorm.DB) error {
		for _, target := range sortedReorderTargets(ids) {
			result := tx.Model(model).
				Scopes(ClinicScope(clinicID)).
				Where("id = ?", target.id).
				Update(orderColumn, target.order)
			if result.Error != nil {
				return apperrors.FromGORM(
					result.Error,
					resource,
					fmt.Sprintf("%d", target.id),
				)
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(
					fmt.Sprintf("%s id %d not found in this clinic", resource, target.id),
				)
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder "+resource)
	}
	return nil
}

// ReorderGlobal updates orderColumn for IDs without clinic scope.
// Like ReorderByClinicID, UPDATEs run in ascending id order so concurrent
// reorders lock rows deterministically; order values follow the input sequence.
func ReorderGlobal(
	ctx context.Context,
	db *gorm.DB,
	model any,
	resource string,
	ids []uint64,
	orderColumn string,
) error {
	if err := DBOrTx(ctx, db).Transaction(func(tx *gorm.DB) error {
		for _, target := range sortedReorderTargets(ids) {
			result := tx.Model(model).
				Where("id = ?", target.id).
				Update(orderColumn, target.order)
			if result.Error != nil {
				return apperrors.FromGORM(
					result.Error,
					resource,
					fmt.Sprintf("%d", target.id),
				)
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(
					fmt.Sprintf("%s id %d not found", resource, target.id),
				)
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder "+resource)
	}
	return nil
}

// Paginate converts one-origin page/limit values into a GORM scope.
func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}
