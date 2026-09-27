package medicalrecord

// care_plan_item_manual_check_test.go — EMR-179: ケアプラン持ち物（type=item）の
// 手入力「その他」明細を支える chk_care_plan_item_ref 緩和の検証。
//
// 目的: migration 011 が (a) other_reason 列を追加し、(b) type=item で
//       hospitalization_plan_id NULL を「category='other' AND btrim(other_reason)<>''」
//       を満たす手入力行に限り許容することを保証する。
//
// 設計（migration 011 の真実の源泉）:
//   care_plan_items.other_reason      → text NOT NULL DEFAULT ''（新規追加）
//   chk_care_plan_item_ref            → type=item の NULL 参照を手入力契約時のみ許容
//     （medicine/treatment は従来通り参照必須、food/instruction は従来通り自由）
//
// 2系統で検証する（checkup_field_cascade_test.go と同型のパターン）:
//   (A) 静的: migration 011 の SQL テキストを直接 assert（migration drift 検出・DB 非依存）。
//   (B) 挙動: AutoMigrate は CHECK 制約を生成しないため、011 から抽出した実 CHECK 式を
//             検証専用のプローブテーブルへ適用し、拒否/受理を実証する。本物の
//             care_plan_items に CHECK を永続適用すると、参照なし item 行を作る
//             既存テスト（AutoMigrate 環境では CHECK 不在を前提）が壊れるため、
//             プローブテーブル方式で副作用を隔離する。

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/testdb"
)

const carePlanItemManualMigrationPath = "../../migrations/011_care_plan_items_manual_other.sql"

// readCarePlanItemManualMigration は migration 011 の SQL テキストを返す（path はパッケージ相対）。
func readCarePlanItemManualMigration(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(carePlanItemManualMigrationPath)
	require.NoError(t, err, "migration 011 を読めること")
	return string(raw)
}

// --- (A) 静的: migration 011 が列追加と緩和済み CHECK を含む ---

func TestCarePlanItemManualMigration_StaticContract(t *testing.T) {
	sql := readCarePlanItemManualMigration(t)

	assert.Contains(t, sql, "other_reason", "011 は other_reason 列を追加すること")
	assert.Contains(t, sql, "ADD COLUMN", "011 は ADD COLUMN で other_reason を追加すること")
	assert.Contains(t, sql, "chk_care_plan_item_ref", "011 は既存 CHECK を同名で再作成すること")
	assert.Contains(t, sql, "DROP CONSTRAINT", "011 は旧 CHECK を先に DROP すること（冪等）")
	assert.Contains(t, sql, "'other'", "緩和条件は category='other' を要求すること")
	assert.Contains(t, sql, "btrim(other_reason)", "緩和条件は trim 後の非空理由を要求すること")

	// medicine/treatment の参照必須は維持される（緩和は type=item のみ）。
	assert.Contains(t, sql, "medicine_id IS NOT NULL", "medicine 参照必須の維持")
	assert.Contains(t, sql, "procedure_id IS NOT NULL", "treatment 参照必須の維持")
	assert.Contains(t, sql, "hospitalization_plan_id IS NOT NULL", "マスタ参照型 item は従来通り有効")
}

// --- (B) 挙動: migration 011 の実 CHECK 式をプローブテーブルへ適用し受理/拒否を実証 ---

// extractCheckExpr は migration SQL から `ADD CONSTRAINT <name> CHECK ( ... )` の
// CHECK 式本体（先頭 "CHECK (" から対応する閉じ括弧まで）を取り出す。
// 011 の CHECK は複数行でネストした括弧を含むため、単純な行検索ではなく括弧深度で終端を決める。
func extractCheckExpr(t *testing.T, sql, constraint string) string {
	t.Helper()
	marker := "ADD CONSTRAINT " + constraint + " CHECK ("
	start := strings.Index(sql, marker)
	require.GreaterOrEqual(t, start, 0, "%s の ADD CONSTRAINT が見つからない", constraint)
	rest := sql[start+len(marker)-1:] // "(" から開始
	depth := 0
	for i, r := range rest {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return "CHECK " + rest[:i+1]
			}
		}
	}
	t.Fatalf("%s の CHECK 式が閉じていない", constraint)
	return ""
}

// setupCarePlanItemManualCheckDB は migration 011 の CHECK 式を載せた検証専用
// プローブテーブルを作る。本物の care_plan_items へは永続変更しない
// （参照なし item 行を前提にする既存 AutoMigrate テスト群を壊さないため）。
func setupCarePlanItemManualCheckDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)
	checkExpr := extractCheckExpr(t, readCarePlanItemManualMigration(t), "chk_care_plan_item_ref")
	probeDDL := `
		CREATE TABLE IF NOT EXISTS care_plan_items_manual_probe (
			id                      BIGSERIAL        PRIMARY KEY,
			hospitalization_id      bigint           NOT NULL,
			type                    care_plan_type   NOT NULL,
			name                    text             NOT NULL DEFAULT '',
			category                text             NOT NULL DEFAULT '',
			other_reason            text             NOT NULL DEFAULT '',
			medicine_id             bigint,
			procedure_id            bigint,
			hospitalization_plan_id bigint,
			CONSTRAINT chk_care_plan_item_ref ` + checkExpr + `
		)`
	require.NoError(t, db.Exec(probeDDL).Error, "011 の CHECK をプローブテーブルへ適用できること")
	t.Cleanup(func() {
		_ = db.Exec("DROP TABLE IF EXISTS care_plan_items_manual_probe").Error
	})
	return db
}

func insertProbeRow(db *gorm.DB, typ string, medicineID, procedureID, hospPlanID any, category, otherReason string) error {
	return db.Exec(`
		INSERT INTO care_plan_items_manual_probe (hospitalization_id, type, name, category, other_reason, medicine_id, procedure_id, hospitalization_plan_id)
		VALUES (1, ?, 't', ?, ?, ?, ?, ?)`,
		typ, category, otherReason, medicineID, procedureID, hospPlanID).Error
}

func isCheckViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "chk_care_plan_item_ref")
}

// 手入力 item（ref NULL）は category=other かつ btrim(other_reason) が非空のときのみ受理される。
func TestCarePlanItemManualCheck_ManualItemAcceptedOnlyWithReason(t *testing.T) {
	db := setupCarePlanItemManualCheckDB(t)

	t.Run("manual item with category=other and reason is accepted", func(t *testing.T) {
		err := insertProbeRow(db, "item", nil, nil, nil, "other", "持ち込み療養食")
		assert.NoError(t, err)
	})

	t.Run("manual item with blank reason is rejected", func(t *testing.T) {
		err := insertProbeRow(db, "item", nil, nil, nil, "other", "")
		assert.True(t, isCheckViolation(err), "理由なし手入力は CHECK 違反であるべき: %v", err)
	})

	t.Run("manual item with whitespace-only reason is rejected", func(t *testing.T) {
		err := insertProbeRow(db, "item", nil, nil, nil, "other", "   ")
		assert.True(t, isCheckViolation(err), "空白理由は CHECK 違反であるべき: %v", err)
	})

	t.Run("manual item without category=other is rejected", func(t *testing.T) {
		err := insertProbeRow(db, "item", nil, nil, nil, "", "理由あり")
		assert.True(t, isCheckViolation(err), "category=other でない ref NULL item は CHECK 違反であるべき: %v", err)
	})

	t.Run("referenced item still accepted", func(t *testing.T) {
		err := insertProbeRow(db, "item", nil, nil, uint64(42), "", "")
		assert.NoError(t, err)
	})

	t.Run("food row stays free-form", func(t *testing.T) {
		err := insertProbeRow(db, "food", nil, nil, nil, "", "")
		assert.NoError(t, err)
	})
}
