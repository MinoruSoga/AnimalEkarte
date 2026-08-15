package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

const junctionInsertBatchSize = 100

// ReplaceChildRowsByParentID atomically replaces child rows for one parent.
// parentFKColumn must be a compile-time literal.
func ReplaceChildRowsByParentID[T any](
	db *gorm.DB,
	parentID uint64,
	rows []T,
	childModel any,
	parentFKColumn string,
	resource string,
	assignParentID func(*T, uint64),
	wrapMessage string,
) error {
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where(fmt.Sprintf("%s = ?", parentFKColumn), parentID).
			Delete(childModel).Error; err != nil {
			return apperrors.FromGORM(
				err,
				resource,
				fmt.Sprintf("%d", parentID),
			)
		}
		if len(rows) == 0 {
			return nil
		}
		for index := range rows {
			assignParentID(&rows[index], parentID)
		}
		if err := tx.CreateInBatches(
			rows,
			junctionInsertBatchSize,
		).Error; err != nil {
			return apperrors.FromGORM(err, resource, "")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, wrapMessage)
	}
	return nil
}

// ValidateClinicScopedMasterIDs validates active master IDs under clinic scope.
func ValidateClinicScopedMasterIDs(
	ctx context.Context,
	db *gorm.DB,
	clinicID uint64,
	ids []uint64,
	masterModel any,
	resource string,
	invalidInputMessage string,
) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := db.WithContext(ctx).
		Model(masterModel).
		Where(
			"clinic_id = ? AND id IN ? AND deleted_at IS NULL",
			clinicID,
			ids,
		).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, resource, "")
	}
	if count != int64(len(ids)) {
		return apperrors.WrapInvalidInput(invalidInputMessage)
	}
	return nil
}

// DeleteJunctionViaMasterClinicScope scopes a junction delete through its master.
// junctionFKColumn must be a compile-time literal.
func DeleteJunctionViaMasterClinicScope(
	tx *gorm.DB,
	clinicID uint64,
	staffID uint64,
	junctionModel any,
	masterModel any,
	junctionFKColumn string,
	junctionResource string,
	errorDetail string,
) error {
	subquery := tx.
		Model(masterModel).
		Select("id").
		Scopes(ClinicScope(clinicID))
	if err := tx.
		Where(
			fmt.Sprintf("staff_id = ? AND %s IN (?)", junctionFKColumn),
			staffID,
			subquery,
		).
		Delete(junctionModel).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errorDetail)
	}
	return nil
}

// DeleteJunctionByClinicAndStaff deletes a clinic-owned staff junction.
func DeleteJunctionByClinicAndStaff(
	tx *gorm.DB,
	clinicID uint64,
	staffID uint64,
	junctionModel any,
	junctionResource string,
	errorDetail string,
) error {
	if err := tx.
		Scopes(ClinicScope(clinicID)).
		Where("staff_id = ?", staffID).
		Delete(junctionModel).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errorDetail)
	}
	return nil
}

// InsertJunctionRowsInBatches inserts junction rows in bounded batches.
func InsertJunctionRowsInBatches[T any](
	tx *gorm.DB,
	rows []T,
	junctionResource string,
	errorDetail string,
) error {
	if len(rows) == 0 {
		return nil
	}
	if err := tx.CreateInBatches(
		rows,
		junctionInsertBatchSize,
	).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errorDetail)
	}
	return nil
}

// ReplaceJunctionInTransaction runs a junction replacement transaction.
func ReplaceJunctionInTransaction(
	db *gorm.DB,
	fn func(tx *gorm.DB) error,
	wrapMessage string,
) error {
	if err := db.Transaction(fn); err != nil {
		return apperrors.Wrap(err, wrapMessage)
	}
	return nil
}
