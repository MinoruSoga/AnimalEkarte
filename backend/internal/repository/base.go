package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// maxMasterListRows はマスタ一覧取得の安全上限件数（C-10）。
// procedure_repository.go 等の Limit を集約。正本は repohelpers.MaxMasterListRows。
const maxMasterListRows = repohelpers.MaxMasterListRows

// clinicScope は指定した clinic_id でレコードをフィルタする GORM スコープ。
// 主テーブルが直接 clinic_id カラムを持つ場合に使用する。
// JOIN 先テーブルを経由してテナント判定する repository では使用不可（JOIN 条件で明示する）。
func clinicScope(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return repohelpers.ClinicScope(clinicID)
}

// clinicScopeIn は複数の clinic_id でレコードをフィルタする GORM スコープ (#86 拠点横断一覧)。
// clinicIDs はハンドラ層で所属医院検証済みであること（検証なしの呼び出しはクロステナント漏洩）。
// 空スライスは IN (NULL) となり 0 件を返す。呼び出し元で非空を保証すること。
func clinicScopeIn(clinicIDs []uint64) func(*gorm.DB) *gorm.DB {
	return repohelpers.ClinicScopeIn(clinicIDs)
}

// paginate は 1-origin の page と limit を Offset/Limit に変換する共有 Scope。
// page/limit の値正規化（下限・上限）は service 層 normalizePagination の責務であり、ここでは行わない。
func paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

// medicalRecordTenantScope は clinic_id を持たない medical_record 子テーブル
// (inquiries/treatments/clinical_plans/medical_record_images 等) のテナント隔離 JOIN。
// childTable は呼び出し側のコンパイル時リテラルのみ許可する（実行時変数の SQL 組み立て禁止）。
func medicalRecordTenantScope(childTable string, clinicID uint64) func(*gorm.DB) *gorm.DB {
	return repohelpers.MedicalRecordTenantScope(childTable, clinicID)
}

// dbOrTx はコンテキストにトランザクションが存在する場合それを返し、
// ない場合はデフォルトの db に WithContext を付けて返す。
// Transactor.WithTx で開始したトランザクション内から呼ばれた場合、自動的に同一 tx を使用する。
func dbOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	return repohelpers.DBOrTx(ctx, db)
}
