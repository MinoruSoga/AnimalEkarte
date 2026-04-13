package repository

import (
	"context"

	"gorm.io/gorm"
)

// clinicScope は指定した clinic_id でレコードをフィルタする GORM スコープ。
// 主テーブルが直接 clinic_id カラムを持つ場合に使用する。
// JOIN 先テーブルを経由してテナント判定する repository では使用不可（JOIN 条件で明示する）。
func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinic_id = ?", clinicID)
	}
}

// dbOrTx はコンテキストにトランザクションが存在する場合それを返し、
// ない場合はデフォルトの db に WithContext を付けて返す。
// Transactor.WithTx で開始したトランザクション内から呼ばれた場合、自動的に同一 tx を使用する。
func dbOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := txFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}
