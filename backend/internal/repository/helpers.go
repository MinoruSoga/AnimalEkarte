package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// reorderByClinicID はマスタテーブルの並び順をトランザクション内で一括更新する汎用ヘルパー。
// clinicID スコープ付きの Reorder 実装で使用する。
func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(model).
				Where("id = ? AND clinic_id = ?", id, clinicID).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, resource, fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("%s id %d not found in this clinic", resource, id))
			}
		}
		return nil
	})
}

// reorderGlobal はクリニック横断マスタテーブルの並び順をトランザクション内で一括更新する汎用ヘルパー。
// clinicID スコープなしの Reorder 実装で使用する（animal_species 等）。
func reorderGlobal(ctx context.Context, db *gorm.DB, model any, resource string, ids []uint64) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(model).
				Where("id = ?", id).
				Update("sort_order", i+1)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, resource, fmt.Sprintf("%d", id))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapInvalidInput(fmt.Sprintf("%s id %d not found", resource, id))
			}
		}
		return nil
	})
}
