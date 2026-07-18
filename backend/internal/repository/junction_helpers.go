package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

const junctionInsertBatchSize = 100

// validateClinicScopedMasterIDs は junction 置換前に、参照先マスタ ID が同一 clinic に属し
// ソフトデリートされていないことを件数一致で検証する。
// db には呼び出し元が dbOrTx(ctx, r.db) を渡すこと（ambient tx 参加）。
func validateClinicScopedMasterIDs(
	ctx context.Context,
	db *gorm.DB,
	clinicID uint64,
	ids []uint64,
	masterModel any,
	resource string,
	invalidInputMsg string,
) error {
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := db.WithContext(ctx).
		Model(masterModel).
		Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, ids).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, resource, "")
	}
	if count != int64(len(ids)) {
		return apperrors.WrapInvalidInput(invalidInputMsg)
	}
	return nil
}

// deleteJunctionViaMasterClinicScope は junction が自前 clinic_id を持たない場合、
// マスタ table の clinic_id サブクエリで削除対象を clinic にスコープする。
// junctionFKColumn は呼び出し側のコンパイル時リテラルのみ許可する。
func deleteJunctionViaMasterClinicScope(
	tx *gorm.DB,
	clinicID, staffID uint64,
	junctionModel any,
	masterModel any,
	junctionFKColumn string,
	junctionResource string,
	errDetail string,
) error {
	subQuery := tx.Model(masterModel).Select("id").Scopes(clinicScope(clinicID))
	if err := tx.Where(
		fmt.Sprintf("staff_id = ? AND %s IN (?)", junctionFKColumn),
		staffID,
		subQuery,
	).Delete(junctionModel).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errDetail)
	}
	return nil
}

// deleteJunctionByClinicAndStaff は junction 自体に clinic_id 列を持つ場合の削除。
func deleteJunctionByClinicAndStaff(
	tx *gorm.DB,
	clinicID, staffID uint64,
	junctionModel any,
	junctionResource string,
	errDetail string,
) error {
	if err := tx.Scopes(clinicScope(clinicID)).Where("staff_id = ?", staffID).
		Delete(junctionModel).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errDetail)
	}
	return nil
}

// insertJunctionRowsInBatches は junction 行を CreateInBatches で一括挿入する。
func insertJunctionRowsInBatches[T any](
	tx *gorm.DB,
	rows []T,
	junctionResource string,
	errDetail string,
) error {
	if len(rows) == 0 {
		return nil
	}
	if err := tx.CreateInBatches(rows, junctionInsertBatchSize).Error; err != nil {
		return apperrors.FromGORM(err, junctionResource, errDetail)
	}
	return nil
}

// replaceJunctionInTransaction は dbOrTx 済みの db で junction 置換 tx を実行し、
// 失敗時に wrapMessage でラップする。
func replaceJunctionInTransaction(
	db *gorm.DB,
	fn func(tx *gorm.DB) error,
	wrapMessage string,
) error {
	if err := db.Transaction(fn); err != nil {
		return apperrors.Wrap(err, wrapMessage)
	}
	return nil
}
