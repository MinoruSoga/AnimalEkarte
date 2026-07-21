package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// BE9-2C R②: 実装は repohelpers へ昇格。以下は既存呼び出し面互換の delegate。

func validateClinicScopedMasterIDs(ctx context.Context, db *gorm.DB, clinicID uint64, ids []uint64, m any, resource, invalidMsg string) error {
	return repohelpers.ValidateClinicScopedMasterIDs(ctx, db, clinicID, ids, m, resource, invalidMsg)
}

func deleteJunctionViaMasterClinicScope(tx *gorm.DB, clinicID, staffID uint64, junction, master any, fkColumn, resource, resourceID string) error {
	return repohelpers.DeleteJunctionViaMasterClinicScope(tx, clinicID, staffID, junction, master, fkColumn, resource, resourceID)
}

func deleteJunctionByClinicAndStaff(tx *gorm.DB, clinicID, staffID uint64, junction any, resource, resourceID string) error {
	return repohelpers.DeleteJunctionByClinicAndStaff(tx, clinicID, staffID, junction, resource, resourceID)
}

func insertJunctionRowsInBatches[T any](tx *gorm.DB, rows []T, resource, resourceID string) error {
	return repohelpers.InsertJunctionRowsInBatches(tx, rows, resource, resourceID)
}

func replaceJunctionInTransaction(db *gorm.DB, fn func(tx *gorm.DB) error, wrapMsg string) error {
	return repohelpers.ReplaceJunctionInTransaction(db, fn, wrapMsg)
}
