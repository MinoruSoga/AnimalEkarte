package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// reorderByClinicID はマスタテーブルの並び順をトランザクション内で一括更新する汎用ヘルパー。
// clinicID スコープ付きの Reorder 実装で使用する。列名は "sort_order" 固定
// （"display_order" 系は paymentmethod 等のドメインパッケージ側ラッパを使う）。
func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64) error {
	return repohelpers.ReorderByClinicID(ctx, db, model, resource, clinicID, ids, "sort_order")
}

// reorderGlobal はクリニック横断マスタテーブルの並び順をトランザクション内で一括更新する汎用ヘルパー。
// clinicID スコープなしの Reorder 実装で使用する（animal_species 等）。orderColumn は通常 "sort_order"。
func reorderGlobal(ctx context.Context, db *gorm.DB, model any, resource string, ids []uint64, orderColumn string) error {
	return repohelpers.ReorderGlobal(ctx, db, model, resource, ids, orderColumn)
}

// findByIDScoped はマスタ/業務テーブルの clinic スコープ付き FindByID を実行する汎用ヘルパー。
// clinicScope必須+apperrors.FromGORMによるエラー変換、というテナント隔離契約を集約する。
// Preload が必要な呼び出し側は本ヘルパーの対象外（呼び出し元で個別実装する）。
func findByIDScoped[T any](ctx context.Context, db *gorm.DB, resource string, clinicID, id uint64) (*T, error) {
	return repohelpers.FindByIDScoped[T](ctx, db, resource, clinicID, id)
}

// updateScopedByID はマスタ/業務テーブルの clinic スコープ付き Update を実行する汎用ヘルパー。
// clinicScope必須+apperrors.FromGORMによるエラー変換+RowsAffected==0→WrapNotFound、というテナント隔離契約を集約する。
// Preload 付き refetch が必要な呼び出し側は本関数の成功後に自身で FindByID を呼ぶこと（挙動保存のため refetch はここでは行わない）。
func updateScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64, fields map[string]any) error {
	return repohelpers.UpdateScopedByID(ctx, db, m, resource, clinicID, id, fields)
}

// deleteScopedByID はマスタ/業務テーブルの clinic スコープ付き Delete（ソフトデリート）を実行する汎用ヘルパー。
// clinicScope必須+apperrors.FromGORMによるエラー変換+RowsAffected==0→WrapNotFound、というテナント隔離契約を集約する。
func deleteScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64) error {
	return repohelpers.DeleteScopedByID(ctx, db, m, resource, clinicID, id)
}
