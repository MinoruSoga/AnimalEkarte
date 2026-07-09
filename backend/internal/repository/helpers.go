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
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			result := tx.Model(model).
				Scopes(clinicScope(clinicID)).Where("id = ?", id).
				Update("sort_order", i+1)
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

// reorderGlobal はクリニック横断マスタテーブルの並び順をトランザクション内で一括更新する汎用ヘルパー。
// clinicID スコープなしの Reorder 実装で使用する（animal_species 等）。
func reorderGlobal(ctx context.Context, db *gorm.DB, model any, resource string, ids []uint64) error {
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	}); err != nil {
		return apperrors.Wrap(err, "failed to reorder "+resource)
	}
	return nil
}

// findByIDScoped はマスタ/業務テーブルの clinic スコープ付き FindByID を実行する汎用ヘルパー。
// P4(clinicScope 必須)+P9(FromGORM) のテナント隔離契約を集約する。
// Preload が必要な呼び出し側は本ヘルパーの対象外（呼び出し元で個別実装する）。
func findByIDScoped[T any](ctx context.Context, db *gorm.DB, resource string, clinicID, id uint64) (*T, error) {
	var record T
	if err := db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).First(&record).Error; err != nil {
		return nil, apperrors.FromGORM(err, resource, fmt.Sprintf("%d", id))
	}
	return &record, nil
}

// updateScopedByID はマスタ/業務テーブルの clinic スコープ付き Update を実行する汎用ヘルパー。
// P4(clinicScope 必須)+P9(FromGORM)+RowsAffected==0→WrapNotFound のテナント隔離契約を集約する。
// Preload 付き refetch が必要な呼び出し側は本関数の成功後に自身で FindByID を呼ぶこと（挙動保存のため refetch はここでは行わない）。
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

// deleteScopedByID はマスタ/業務テーブルの clinic スコープ付き Delete（ソフトデリート）を実行する汎用ヘルパー。
// P4(clinicScope 必須)+P9(FromGORM)+RowsAffected==0→WrapNotFound のテナント隔離契約を集約する。
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
