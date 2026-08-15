package medicalrecord

// checkup_field_repository.go — CheckupTypeFieldRepository / CheckupFieldResultRepository の実装。
// Moved from internal/repository/checkup_field_repository.go — BE9-2D roll-up. Type/constructor
// names are unchanged; the internal/repository facade re-exports them as aliases so no caller
// changes. Package-private clinicScope/dbOrTx are swapped for persistence.ClinicScope/DBOrTx
// (this package must not import internal/repository — that would be an import cycle via the
// facade). The dbOrTx→persistence.DBOrTx rename keeps the same ambient-tx participation
// (dbortx_inventory_lint matches all three call shapes); the audit_tx / dbortx inventory
// allowlist keys are updated to the medicalrecord/ path.

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CheckupTypeFieldRepository は健診パッケージのフィールド定義マスタ（checkup_type_fields）アクセス。
type CheckupTypeFieldRepository interface {
	// FindByCheckupTypeID は指定 checkup_type の生存フィールド定義を sort_order 昇順で返す。
	FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
}

// CheckupFieldResultRepository は健診結果値（checkup_field_results）アクセス。
type CheckupFieldResultRepository interface {
	// FindByCheckupID は指定 checkup の結果値を sort_order 昇順で返す。
	FindByCheckupID(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error)
	// FindByPetID は飼い主レポート用に pet 単位の健診結果を返す（生存 checkup 経由）。
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	// ReplaceForCheckup は checkup_field_results を一括置換する（既存全削除→一括挿入をトランザクション内で実行）。
	// 第 2 戻り値は実際に削除された行数（DELETE の RowsAffected）。サービス層が「削除が起きたか」を
	// 監査ゲートに使う（#211: スナップショットでなく実削除数で判定し、競合下の無監査 hard-delete を防ぐ）。
	ReplaceForCheckup(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error)
}

type checkupTypeFieldRepository struct {
	db *gorm.DB
}

func NewCheckupTypeFieldRepository(db *gorm.DB) CheckupTypeFieldRepository {
	return &checkupTypeFieldRepository{db: db}
}

func (r *checkupTypeFieldRepository) FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	fields := make([]model.CheckupTypeField, 0)
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("checkup_type_id = ?", checkupTypeID).
		Order("sort_order ASC, id ASC").
		Find(&fields).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type_field", fmt.Sprintf("checkup_type=%d", checkupTypeID))
	}
	return fields, nil
}

type checkupFieldResultRepository struct {
	db *gorm.DB
}

func NewCheckupFieldResultRepository(db *gorm.DB) CheckupFieldResultRepository {
	return &checkupFieldResultRepository{db: db}
}

func (r *checkupFieldResultRepository) FindByCheckupID(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	results := make([]model.CheckupFieldResult, 0)
	// dbOrTx: ambient tx 内から呼ばれた場合は同一 tx で読む（#211 置換後の read-your-writes /
	// 置換前スナップショットを削除と同一 tx で一貫取得）。tx 外では base db（従来挙動）。
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where(`EXISTS (
				SELECT 1
				FROM checkups
				WHERE checkups.id = checkup_field_results.checkup_id
				  AND checkups.clinic_id = checkup_field_results.clinic_id
			)`).
		// P3.1: clinic-scoped マスタ Preload は clinic_id 述語必須。
		Preload("CheckupTypeField", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("checkup_field_results.checkup_id = ?", checkupID).
		Order("sort_order ASC, id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
	}
	return results, nil
}

func (r *checkupFieldResultRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	results := make([]model.CheckupFieldResult, 0)
	// 結果は生存 checkup 経由で pet 単位に解決する。checkup_field_results.clinic_id と
	// checkups.clinic_id の双方を clinic でスコープし、クロステナント混入を防ぐ。
	err := r.db.WithContext(ctx).
		Joins("JOIN checkups ON checkups.id = checkup_field_results.checkup_id"+
			" AND checkups.clinic_id = ?"+
			" AND checkups.deleted_at IS NULL", clinicID).
		Where("checkup_field_results.clinic_id = ? AND checkups.pet_id = ?", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = checkups.pet_id AND p.clinic_id = checkups.clinic_id)").
		// P3.1: clinic-scoped マスタ Preload は clinic_id 述語必須。
		Preload("CheckupTypeField", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		// 親 checkup と種別（飼い主レポートの日付・パッケージ名表示用）。
		Preload("Checkup", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Checkup.CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("checkups.date DESC, checkup_field_results.sort_order ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("pet=%d", petID))
	}
	return results, nil
}

// ReplaceForCheckup は checkup_field_results を一括置換する（PUT セマンティクス）。
// 親 checkup の clinic 隔離はサービス層の FindByID で保証されている前提だが、
// トランザクション内でも clinic スコープ付きで再確認する（並行削除/clinic 越境防止）。
// results の CheckupID / ClinicID は本メソッド内で強制上書きする。
func (r *checkupFieldResultRepository) ReplaceForCheckup(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, int64, error) {
	// dbOrTx: ambient tx（Transactor.WithTx）内から呼ばれた場合は同一 tx に join し、
	// .Transaction は savepoint（ネスト）として実行される。これにより削除+挿入が caller の tx に入り、
	// 後続の監査書込が失敗して caller が tx を rollback すると削除も巻き戻る（#211 fail-closed 原子性）。
	// ambient tx が無い場合は base db で独立トランザクションを開く（従来挙動＝後方互換）。
	var deletedCount int64
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		// 本クロージャ引数 tx が savepoint レベルの tx ハンドル。キャプチャされる ctx は外側（ambient）tx を
		// txKey に持つため、内部処理は必ず引数 tx を使い dbOrTx(ctx,…) で取り直さないこと
		// （同一 *sql.Tx なので実害はないが、将来の抽出リファクタで別ハンドルにすり替わる罠を避ける）。
		var count int64
		if err := tx.Model(&model.Checkup{}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", checkupID, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", checkupID))
		}
		if count == 0 {
			return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
		}

		// 既存結果を clinic スコープで全削除（CASCADE では消えないため明示削除）。
		// RowsAffected を呼び出し元へ返し、サービス層が「実際に削除が起きたか」を監査ゲートに使う
		// （#211 security MEDIUM-1: READ COMMITTED 下でスナップショットが 0 件でも、並行 INSERT が
		// commit した行を本 DELETE が消す競合がある。スナップショット件数でなく実削除数でゲートすることで
		// 監査なしの hard-delete が残らないようにする）。
		del := tx.Where("checkup_id = ? AND clinic_id = ?", checkupID, clinicID).
			Delete(&model.CheckupFieldResult{})
		if del.Error != nil {
			return apperrors.FromGORM(del.Error, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
		}
		deletedCount = del.RowsAffected

		if len(results) == 0 {
			return nil
		}

		for i := range results {
			results[i].CheckupID = checkupID
			results[i].ClinicID = clinicID
			results[i].ID = 0 // 新規挿入のため auto increment に任せる
			// value_list は NOT NULL DEFAULT '{}'。nil の pq.StringArray は driver が
			// NULL を返すため、明示的に空配列へ正規化して GORM の default 省略挙動への依存を断つ。
			if results[i].ValueList == nil {
				results[i].ValueList = pq.StringArray{}
			}
		}
		if err := tx.Create(&results).Error; err != nil {
			return apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
		}
		return nil
	})
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to replace checkup field results")
	}
	// 置換後の最終状態を読み直す。ctx は ambient tx を保持するため dbOrTx で同一 tx から読み、
	// 削除+挿入後の結果を read-your-writes で返す（tx 外呼び出し時は base db＝従来挙動）。
	saved, err := r.FindByCheckupID(ctx, clinicID, checkupID)
	if err != nil {
		return nil, 0, err
	}
	return saved, deletedCount, nil
}
