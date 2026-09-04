package persistence

import (
	"context"
	"fmt"

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

// ReorderByClinicID updates orderColumn for IDs under clinic scope.
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
		for index, id := range ids {
			result := tx.Model(model).
				Scopes(ClinicScope(clinicID)).
				Where("id = ?", id).
				Update(orderColumn, index+1)
			if result.Error != nil {
				return apperrors.FromGORM(
					result.Error,
					resource,
					fmt.Sprintf("%d", id),
				)
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(
					fmt.Sprintf("%s id %d not found in this clinic", resource, id),
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
func ReorderGlobal(
	ctx context.Context,
	db *gorm.DB,
	model any,
	resource string,
	ids []uint64,
	orderColumn string,
) error {
	if err := DBOrTx(ctx, db).Transaction(func(tx *gorm.DB) error {
		for index, id := range ids {
			result := tx.Model(model).
				Where("id = ?", id).
				Update(orderColumn, index+1)
			if result.Error != nil {
				return apperrors.FromGORM(
					result.Error,
					resource,
					fmt.Sprintf("%d", id),
				)
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(
					fmt.Sprintf("%s id %d not found", resource, id),
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
