package repohelpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// ValidateClinicScopedMasterIDs delegates to persistence.
func ValidateClinicScopedMasterIDs(
	ctx context.Context,
	db *gorm.DB,
	clinicID uint64,
	ids []uint64,
	masterModel any,
	resource string,
	invalidInputMsg string,
) error {
	return persistence.ValidateClinicScopedMasterIDs(
		ctx,
		db,
		clinicID,
		ids,
		masterModel,
		resource,
		invalidInputMsg,
	)
}

// DeleteJunctionViaMasterClinicScope delegates to persistence.
func DeleteJunctionViaMasterClinicScope(
	tx *gorm.DB,
	clinicID, staffID uint64,
	junctionModel any,
	masterModel any,
	junctionFKColumn string,
	junctionResource string,
	errDetail string,
) error {
	return persistence.DeleteJunctionViaMasterClinicScope(
		tx,
		clinicID,
		staffID,
		junctionModel,
		masterModel,
		junctionFKColumn,
		junctionResource,
		errDetail,
	)
}

// DeleteJunctionByClinicAndStaff delegates to persistence.
func DeleteJunctionByClinicAndStaff(
	tx *gorm.DB,
	clinicID, staffID uint64,
	junctionModel any,
	junctionResource string,
	errDetail string,
) error {
	return persistence.DeleteJunctionByClinicAndStaff(
		tx,
		clinicID,
		staffID,
		junctionModel,
		junctionResource,
		errDetail,
	)
}

// InsertJunctionRowsInBatches delegates to persistence.
func InsertJunctionRowsInBatches[T any](
	tx *gorm.DB,
	rows []T,
	junctionResource string,
	errDetail string,
) error {
	return persistence.InsertJunctionRowsInBatches(
		tx,
		rows,
		junctionResource,
		errDetail,
	)
}

// ReplaceJunctionInTransaction delegates to persistence.
func ReplaceJunctionInTransaction(
	db *gorm.DB,
	fn func(tx *gorm.DB) error,
	wrapMessage string,
) error {
	return persistence.ReplaceJunctionInTransaction(db, fn, wrapMessage)
}
