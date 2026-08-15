package medicalrecord

// checkup_field_cascade_test.go — #211 follow-up: CASCADE DELETE 安全監査（FU1）
//
// 目的: migration 010 の ON DELETE 設計が「患者の健診結果値（checkup_field_results）を
//       マスタ/フィールド定義の削除で silent に失わない」ことを決定的に保証する。
//
// 注記（2026-07-04 統合）: migration 010 は 001_init.sql へ統合され独立ファイルとしては
// 存在しない（docs/architecture/erd.md §4.3）。本ファイルが「migration 010」と呼ぶのは 001_init.sql に
// 統合された当該 CREATE TABLE ブロックの設計意図であり、以下の (A) 静的テストは 001_init.sql
// から直接その CREATE TABLE ブロックを抽出する。同ファイル内には 010 の CREATE TABLE より
// 後段で、旧 012 由来の複合FK ALTER 文（checkup_type_field_id を単一列 FK から複合FKへ置換）
// が続くが、そちらは checkup_field_composite_fk_test.go が個別に検証する別レイヤーであり、
// 本ファイルが検証する「作成時点の単一列 SET NULL」設計そのものは変わらない。
//
// 設計（migration 010 の FK アクション・真実の源泉）:
//   checkup_type_fields.checkup_type_id      → checkup_types  ON DELETE CASCADE   (定義はパッケージ構成要素)
//   checkup_type_fields.clinic_id            → clinics        ON DELETE RESTRICT
//   checkup_field_results.checkup_id         → checkups       ON DELETE CASCADE   (結果は checkup の純粋従属子)
//   checkup_field_results.checkup_type_field_id → checkup_type_fields ON DELETE SET NULL ★患者データ保護の要
//   checkup_field_results.clinic_id          → clinics        ON DELETE RESTRICT
//
// 安全性の核心: 結果値→フィールド定義 の FK は SET NULL であり、field_name/field_type/unit/値 を
// 非正規化スナップショットとして結果行に保持する（exam_results.exam_type_field_id と同型）。
// よってマスタ/フィールド定義を削除しても結果値は残存し、自己記述的に飼主レポートへ表示できる。
// 一方 結果値→checkup の FK は CASCADE で、checkup（記録）が消えれば結果も意味を失うため共に消える
// （vital_records→medical_records / exam_results→exams と同じ「許容される従属 CASCADE」例外）。
//
// 2系統で検証する:
//   (A) 静的: migration 010 SQL の FK アクションを直接 assert（migration drift を CI で検出・DB 非依存）。
//   (B) 挙動: テスト DB の AutoMigrate は ON DELETE を再現しないため、2 テーブルを migration の実 DDL
//             （ファイルから抽出）で再作成し、hard-delete のカスケード挙動を実証する。B も migration を
//             読むため、FK アクションを退行させれば A だけでなく B も RED になる（drift 検出は二重）。

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

const checkupMigration010Path = "../../migrations/001_init.sql"

// readCheckupMigration010 は migration 010 の SQL テキストを返す（path はパッケージ相対）。
func readCheckupMigration010(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(checkupMigration010Path)
	require.NoError(t, err, "migration 010 を読めること")
	return string(raw)
}

// --- (A) 静的: migration 010 の FK ON DELETE アクションが安全設計と一致する ---
//
// このテストは DB を必要とせず、010 のテキストを直接検証する。誰かが
// checkup_type_field_id を SET NULL → CASCADE に変えれば（= マスタ削除で患者結果値が
// 連鎖消去される設計に退行）即 fail する。migrations/CLAUDE.md の CASCADE 規約の機械的ガード。
//
// 注意: assertFKOnDelete は「カラム名で始まる単一行に REFERENCES 句と ON DELETE 句が収まる」
//
//	現状の整形に依存する。将来 010.sql の FK 句を複数行に分割した場合は本ヘルパーの更新が必要
//	（その場合 CI は false-fail で気付ける = 安全側）。
func TestCheckupMigration_FKActions_PreservePatientResults(t *testing.T) {
	sql := readCheckupMigration010(t)

	fieldsDDL := extractCreateTableDDL(t, sql, "checkup_type_fields")
	resultsDDL := extractCreateTableDDL(t, sql, "checkup_field_results")

	// checkup_type_fields: 定義は checkup_type の構成要素 → CASCADE。clinic は RESTRICT。
	assertFKOnDelete(t, fieldsDDL, "checkup_type_id", "REFERENCES checkup_types(id)", "ON DELETE CASCADE")
	assertFKOnDelete(t, fieldsDDL, "clinic_id", "REFERENCES clinics(id)", "ON DELETE RESTRICT")

	// checkup_field_results: 結果値の保全が懸かる FK 群。
	//   checkup_id            → CASCADE   (checkup の純粋従属子・許容例外)
	//   checkup_type_field_id → SET NULL  ★患者結果値をマスタ削除から守る
	//   clinic_id             → RESTRICT
	assertFKOnDelete(t, resultsDDL, "checkup_id", "REFERENCES checkups(id)", "ON DELETE CASCADE")
	assertFKOnDelete(t, resultsDDL, "checkup_type_field_id", "REFERENCES checkup_type_fields(id)", "ON DELETE SET NULL")
	assertFKOnDelete(t, resultsDDL, "clinic_id", "REFERENCES clinics(id)", "ON DELETE RESTRICT")
}

// extractCreateTableDDL は migration SQL から `CREATE TABLE <table> ( ... );` ブロックを取り出す
// （終端の ");" を含むため、そのまま Exec できる）。
func extractCreateTableDDL(t *testing.T, sql, table string) string {
	t.Helper()
	marker := "CREATE TABLE " + table + " ("
	start := strings.Index(sql, marker)
	require.GreaterOrEqual(t, start, 0, "CREATE TABLE %s がマイグレーションに見つからない", table)
	rest := sql[start:]
	// CREATE TABLE 本体は最初の行頭 ");" で閉じる（10桁整形の本マイグレーションで安定）。
	end := strings.Index(rest, "\n);")
	require.GreaterOrEqual(t, end, 0, "CREATE TABLE %s の終端 ');' が見つからない", table)
	return rest[:end+len("\n);")]
}

// assertFKOnDelete は DDL 内の指定カラム行が指定 REFERENCES と ON DELETE アクションを持つことを検証する。
func assertFKOnDelete(t *testing.T, ddl, col, refs, action string) {
	t.Helper()
	for line := range strings.SplitSeq(ddl, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, col+" ") {
			continue
		}
		require.Contains(t, line, refs, "%s の REFERENCES 句が想定と異なる", col)
		assert.Contains(t, line, action,
			"FK %s は %s でなければならない（患者結果値の保全 / migrations CASCADE 規約）", col, action)
		return
	}
	t.Fatalf("カラム %s の定義行が DDL に見つからない", col)
}

// --- (B) 挙動: migration の実 DDL で再作成し hard-delete のカスケードを実証 ---

// clinicsFKPattern は clinics への FK 句を表す（テスト DB に clinics テーブルが無いため除去する）。
var clinicsFKPattern = regexp.MustCompile(`REFERENCES clinics\(id\)\s+ON DELETE RESTRICT`)

// setupCheckupCascadeTestDB は AutoMigrate 後に checkup_type_fields / checkup_field_results を
// DROP し、migration 010 から抽出した実 DDL で再作成する。AutoMigrate は GORM の relation タグに
// OnDelete 指定が無いため ON DELETE 句を再現しない。本 setup は migration を読むため、
// checkup_id(CASCADE) / checkup_type_field_id(SET NULL) / checkup_type_id(CASCADE) の実アクションを
// そのまま検証対象にできる（FK 退行は B でも RED になる）。
// clinic_id FK のみ除去する（clinics テーブルがテスト DB に不在のため。clinic RESTRICT は静的テスト A が担保）。
func setupCheckupCascadeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupCheckupFieldTestDB(t)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS checkup_field_results CASCADE").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS checkup_type_fields CASCADE").Error)

	sql := readCheckupMigration010(t)
	fieldsDDL := clinicsFKPattern.ReplaceAllString(extractCreateTableDDL(t, sql, "checkup_type_fields"), "")
	resultsDDL := clinicsFKPattern.ReplaceAllString(extractCreateTableDDL(t, sql, "checkup_field_results"), "")

	require.NoError(t, db.Exec(fieldsDDL).Error, "checkup_type_fields を実 DDL で再作成できること")
	require.NoError(t, db.Exec(resultsDDL).Error, "checkup_field_results を実 DDL で再作成できること")
	// Later 001_init ALTERs add import provenance columns used by the GORM model.
	require.NoError(t, db.Exec(`
		ALTER TABLE checkup_type_fields
			ADD COLUMN IF NOT EXISTS import_namespace text,
			ADD COLUMN IF NOT EXISTS import_key text;
		ALTER TABLE checkup_types
			ADD COLUMN IF NOT EXISTS import_namespace text,
			ADD COLUMN IF NOT EXISTS import_key text;
	`).Error)
	return db
}

// seedCheckupResultTree は clinic A に owner→pet→MR→checkup_type→field→checkup→result を 1 本作る。
func seedCheckupResultTree(t *testing.T, db *gorm.DB) (fieldID, checkupID uint64) {
	t.Helper()
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "カスケード飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "カスケードポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CASCADE", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石除去必要の有無",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	checkupID = makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	boolTrue := true
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean,
		ValueBool: &boolTrue, SortOrder: 1,
	}).Error)
	return field.ID, checkupID
}

// FU-SC1: フィールド定義（マスタ）を hard-delete しても結果値は残存し、FK は NULL 化され、
// 非正規化スナップショットが保持される（ON DELETE SET NULL）。
// これがマスタ削除で患者結果値が silent 消去されないことの決定的証拠。
func TestCheckupFieldResult_FieldDefinitionHardDelete_PreservesResults(t *testing.T) {
	db := setupCheckupCascadeTestDB(t)
	ctx := context.Background()
	fieldID, checkupID := seedCheckupResultTree(t, db)

	// フィールド定義をハード削除（マスタ管理 UI でのフィールド除去 / データ purge を模倣）。
	require.NoError(t, db.Exec("DELETE FROM checkup_type_fields WHERE id = ?", fieldID).Error)

	var results []model.CheckupFieldResult
	require.NoError(t, db.WithContext(ctx).Where("checkup_id = ?", checkupID).Find(&results).Error)
	require.Len(t, results, 1, "フィールド定義削除後も結果値は残存しなければならない（SET NULL）")

	r := results[0]
	assert.Nil(t, r.CheckupTypeFieldID, "削除されたフィールド定義への FK は NULL 化される")
	assert.Equal(t, "歯石除去必要の有無", r.FieldName, "非正規化スナップショット（field_name）は保持される")
	assert.Equal(t, model.CheckupFieldTypeBoolean, r.FieldType, "field_type スナップショットは保持される")
	require.NotNil(t, r.ValueBool, "結果値（value_bool）は保持される")
	assert.True(t, *r.ValueBool)
}

// FU-SC1（security M1）: checkup_type（パッケージ＝マスタ）を hard-delete すると、フィールド定義は
// CASCADE で消えるが、結果値は checkup_type_field_id の SET NULL で残存する。3連鎖
// （checkup_types → CASCADE checkup_type_fields → SET NULL checkup_field_results）の end-to-end 実証。
//
// checkups.checkup_type_id は RESTRICT のため、結果の checkup が同じ checkup_type を指すと削除が阻止される。
// 連鎖そのものを観測するため、結果の field は削除対象 ctA に属させ、結果の checkup は別 ctB に属させる
// （FK 挙動の検証が目的。実運用では service が field と checkup の type 一致を保証する）。
func TestCheckupFieldResult_CheckupTypeHardDelete_PreservesResultsViaCascadeSetNull(t *testing.T) {
	db := setupCheckupCascadeTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "連鎖飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "連鎖ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CHAIN", time.Now())

	ctA := makeCheckupTypeMaster(t, db, clinicA, "削除対象パッケージ")
	ctB := makeCheckupTypeMaster(t, db, clinicA, "結果の所属パッケージ")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ctA.ID, Name: "歯石付着度",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ctB.ID)

	score := 3.0
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: "歯石付着度", FieldType: model.CheckupFieldTypeNumber,
		ValueNumber: &score, SortOrder: 1,
	}).Error)

	// checkup_type（マスタ）をハード削除 → フィールド定義は CASCADE 消去。
	require.NoError(t, db.Exec("DELETE FROM checkup_types WHERE id = ?", ctA.ID).Error)

	var fieldCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.CheckupTypeField{}).
		Where("id = ?", field.ID).Count(&fieldCount).Error)
	assert.Equal(t, int64(0), fieldCount, "checkup_type 削除でフィールド定義は CASCADE 削除される")

	var results []model.CheckupFieldResult
	require.NoError(t, db.WithContext(ctx).Where("checkup_id = ?", checkupID).Find(&results).Error)
	require.Len(t, results, 1, "checkup_type 削除後も結果値は残存しなければならない（フィールド CASCADE→結果 SET NULL）")
	assert.Nil(t, results[0].CheckupTypeFieldID, "CASCADE 消去されたフィールドへの FK は NULL 化される")
	assert.Equal(t, "歯石付着度", results[0].FieldName, "スナップショットは保持される")
	require.NotNil(t, results[0].ValueNumber)
	assert.InDelta(t, 3.0, *results[0].ValueNumber, 0.001, "結果値は保持される")
}

// FU-SC1: checkup（健診記録）を hard-delete すると結果値は CASCADE で消える。
// 結果値は checkup の純粋従属子であり、記録が消えれば意味を失うため共に消えるのが正しい
// （vital_records→medical_records / exam_results→exams と同じ許容例外）。意図した挙動をロックする。
func TestCheckupFieldResult_CheckupHardDelete_CascadesResults(t *testing.T) {
	db := setupCheckupCascadeTestDB(t)
	ctx := context.Background()
	_, checkupID := seedCheckupResultTree(t, db)

	var before int64
	require.NoError(t, db.WithContext(ctx).Model(&model.CheckupFieldResult{}).
		Where("checkup_id = ?", checkupID).Count(&before).Error)
	require.Equal(t, int64(1), before, "前提: 結果値が 1 件存在する")

	// checkup をハード削除（DB レベル CASCADE を発火させる。通常運用はソフトデリートだが
	// データ purge / 物理削除時に結果値が孤児化しないことを保証する）。
	require.NoError(t, db.Exec("DELETE FROM checkups WHERE id = ?", checkupID).Error)

	var after int64
	require.NoError(t, db.WithContext(ctx).Model(&model.CheckupFieldResult{}).
		Where("checkup_id = ?", checkupID).Count(&after).Error)
	assert.Equal(t, int64(0), after, "checkup 削除で結果値は CASCADE 削除される（純粋従属子・孤児を残さない）")
}
