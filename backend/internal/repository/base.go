package repository

import "gorm.io/gorm"

// clinicScope は指定した clinic_id でレコードをフィルタする GORM スコープ。
// 主テーブルが直接 clinic_id カラムを持つ場合に使用する。
// JOIN 先テーブルを経由してテナント判定する repository では使用不可（JOIN 条件で明示する）。
func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("clinic_id = ?", clinicID)
	}
}
