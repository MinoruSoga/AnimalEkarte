package repository

// checkup_field_composite_fk_test.go — BE-refactor.md R3-7 (D13): 臨床結果テーブルの DB レベル
// 複合 FK（clinic_id 込み）による越境 INSERT/UPDATE の物理拒否を検証する。
//
// migration 012 は checkup_field_results(checkup_type_field_id, clinic_id) →
// checkup_type_fields(id, clinic_id) の複合 FK を張り、別クリニックのフィールド定義を参照する
// 結果行の永続化を DB レベルで拒否する（アプリ層 FindByID ガード #124 f4e7b7a7 の defense-in-depth）。
// ON DELETE SET NULL (checkup_type_field_id) で migration 010 の患者結果値保護（フィールド定義削除時に
// 結果値を残す）を挙動保存する。
//
// 検証方式は checkup_field_cascade_test.go と同じ「migration の実 DDL を抽出して適用」パターン。
// setupCheckupCascadeTestDB が 010 の実 DDL でテーブルを再作成し、本ファイルが 012 の実 ALTER を
// 適用する。よって 012 の FK を退行させれば本テストが RED になる（drift 検出）。
//
// 注: exam_results は clinic_id 列を持たず（参照先 exam_type_fields も持たない）複合 FK を張れないため
// 対象外。同等防御は clinic_id 列追加 + backfill という非 additive スキーマ拡張を要する別タスク。

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

const checkupMigration012Path = "../../migrations/012_add_clinical_result_composite_fk.sql"

// applyMigration012CompositeFK は migration 012 の実 ALTER 文をテスト DB へ適用する。
// 012 は clinics を参照しないため（UNIQUE(id,clinic_id) と複合 FK のみ）そのまま適用できる。
// setupCheckupCascadeTestDB が 010 の実 DDL でテーブルを再作成した後に呼ぶこと
// （010 のインライン FK checkup_field_results_checkup_type_field_id_fkey を 012 が DROP するため）。
func applyMigration012CompositeFK(t *testing.T, db *gorm.DB) {
	t.Helper()
	raw, err := os.ReadFile(checkupMigration012Path)
	require.NoError(t, err, "migration 012 を読めること")

	for _, stmt := range splitSQLStatements(string(raw)) {
		require.NoError(t, db.Exec(stmt).Error, "migration 012 の文を適用できること: %s", stmt)
	}
}

// splitSQLStatements は SQL テキストから行コメント（-- …）を除去し、';' 区切りの文に分割する。
func splitSQLStatements(sql string) []string {
	var b strings.Builder
	for line := range strings.SplitSeq(sql, "\n") {
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	var out []string
	for s := range strings.SplitSeq(b.String(), ";") {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

// setupCheckupCompositeFKTestDB は 010 実 DDL 再作成（setupCheckupCascadeTestDB）に 012 の複合 FK を
// 重ねたテスト DB を返す。
func setupCheckupCompositeFKTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupCheckupCascadeTestDB(t)
	applyMigration012CompositeFK(t, db)
	return db
}

// TestCheckupFieldResult_CompositeFK_RejectsCrossClinicField は BE-refactor.md R3-7 の核心:
// clinic B の結果行が clinic A のフィールド定義を参照しようとすると DB レベルで拒否される。
func TestCheckupFieldResult_CompositeFK_RejectsCrossClinicField(t *testing.T) {
	db := setupCheckupCompositeFKTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A のフィールド定義を作る。
	ctA := makeCheckupTypeMaster(t, db, clinicA, "歯科検診A")
	fieldA := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ctA.ID, Name: "歯石除去必要の有無",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	// clinic B の checkup を作る（越境 INSERT の器）。
	ownerB := makeOwner(t, db, clinicB, "越境飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "越境ポチ")
	mrB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-XCLINIC", time.Now())
	ctB := makeCheckupTypeMaster(t, db, clinicB, "歯科検診B")
	checkupB := makeCheckupRec(t, db, clinicB, mrB.ID, petB.ID, ctB.ID)

	boolTrue := true
	// clinic B の結果行が clinic A のフィールド定義(fieldA)を参照 → 複合 FK (field.clinic=A, result.clinic=B) 不一致で拒否。
	err := db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicB, CheckupID: checkupB, CheckupTypeFieldID: &fieldA.ID,
		FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean,
		ValueBool: &boolTrue, SortOrder: 1,
	}).Error

	require.Error(t, err, "別クリニックのフィールド定義を参照する結果行は複合 FK で拒否される")
	assert.True(t, isFKConstraintErr(err), "FK 違反(23503)であること: %v", err)
}

// TestCheckupFieldResult_CompositeFK_AcceptsSameClinicField は挙動保存の確認:
// 同一クリニックのフィールド定義を参照する正当な結果行は従来どおり永続化できる（false-reject なし）。
func TestCheckupFieldResult_CompositeFK_AcceptsSameClinicField(t *testing.T) {
	db := setupCheckupCompositeFKTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "正当飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "正当ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-OK", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石除去必要の有無",
		FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	boolTrue := true
	err := db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean,
		ValueBool: &boolTrue, SortOrder: 1,
	}).Error

	require.NoError(t, err, "同一クリニックのフィールド定義参照は許可される（false-reject なし）")
}

// TestCheckupFieldResult_CompositeFK_PreservesSetNullBehavior は挙動保存の要:
// 複合 FK の列指定 SET NULL でも migration 010 の患者結果値保護が維持される。
// フィールド定義を hard-delete しても結果行は残り、checkup_type_field_id のみ NULL 化され、
// NOT NULL の clinic_id と非正規化スナップショットは保持される。
func TestCheckupFieldResult_CompositeFK_PreservesSetNullBehavior(t *testing.T) {
	db := setupCheckupCompositeFKTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeOwner(t, db, clinicA, "SETNULL飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "SETNULLポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-SETNULL", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")
	field := makeCheckupTypeField(t, db, &model.CheckupTypeField{
		ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石付着度",
		FieldType: model.CheckupFieldTypeNumber, SortOrder: 1,
	})
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	score := 3.0
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &field.ID,
		FieldName: "歯石付着度", FieldType: model.CheckupFieldTypeNumber,
		ValueNumber: &score, SortOrder: 1,
	}).Error)

	// フィールド定義を hard-delete → 列指定 SET NULL で checkup_type_field_id のみ NULL 化される。
	require.NoError(t, db.Exec("DELETE FROM checkup_type_fields WHERE id = ?", field.ID).Error)

	var results []model.CheckupFieldResult
	require.NoError(t, db.WithContext(ctx).Where("checkup_id = ?", checkupID).Find(&results).Error)
	require.Len(t, results, 1, "フィールド定義削除後も結果値は残存する（010 の SET NULL 保護を複合 FK でも維持）")

	r := results[0]
	assert.Nil(t, r.CheckupTypeFieldID, "checkup_type_field_id のみ NULL 化される")
	assert.Equal(t, clinicA, r.ClinicID, "NOT NULL の clinic_id は保持される（列指定 SET NULL）")
	assert.Equal(t, "歯石付着度", r.FieldName, "非正規化スナップショットは保持される")
	require.NotNil(t, r.ValueNumber)
	assert.InDelta(t, 3.0, *r.ValueNumber, 0.001)
}
