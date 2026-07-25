package medicalrecord

// checkup_field_repository_test.go — #211 健診パッケージ
//   SC2: 型付きフィールド機構の round-trip（number/boolean/multi_select）。
//   SC4: pet 単位 read のマスタ Preload clinic 隔離（CheckupTypeField）。
//
// FindByPetID の Preload から "clinic_id = ?" を外すと cross-clinic ケースが失敗する（temp-revert RED）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupCheckupFieldTestDB は setupIsolatedTestDB（呼び出し毎に使い捨ての新規 DB 接続）を使う。
// checkup_field_results/checkup_type_fields は本ファイル・checkup_field_cascade_test.go・
// checkup_field_composite_fk_test.go の 3 ヘルパーが同じテーブルを異なるスキーマ（AutoMigrate 由来 /
// migration 実 DDL 由来）で意図的に毎回 DROP+CREATE し合う（migration drift 検出のため必須の挙動）。
// setupTestDB の共有コネクションプールでこれを行うと、別テストが保持する古いテーブル/型 OID 参照の
// キャッシュ済み prepared statement が壊れるため、この cluster だけ専用の使い捨て接続を使う
// （詳細は setupIsolatedTestDB のコメント参照）。
func setupCheckupFieldTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)
	// checkup_field_type ENUM は migration 010 で新設。setupTestDB のハードコード一覧に
	// 含まれないため、本テストで明示的に作成する（exam_result_status は setupTestDB が作成済み）。
	db.Exec("DROP TABLE IF EXISTS checkup_field_results CASCADE")
	db.Exec("DROP TABLE IF EXISTS checkup_type_fields CASCADE")
	db.Exec("DROP TYPE IF EXISTS checkup_field_type CASCADE")
	require.NoError(t, db.Exec("CREATE TYPE checkup_field_type AS ENUM ('number','single_select','multi_select','boolean','checklist','text')").Error)
	require.NoError(t, db.AutoMigrate(
		&model.AnimalSpecies{}, &model.Pet{}, &model.CheckupType{},
		&model.CheckupTypeField{}, &model.Checkup{}, &model.CheckupFieldResult{},
	))
	db.Exec("TRUNCATE TABLE checkup_field_results CASCADE")
	db.Exec("TRUNCATE TABLE checkup_type_fields CASCADE")
	db.Exec("TRUNCATE TABLE checkups CASCADE")
	db.Exec("TRUNCATE TABLE checkup_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeCheckupTypeField(t *testing.T, db *gorm.DB, f *model.CheckupTypeField) *model.CheckupTypeField {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(f).Error)
	return f
}

// SC2: 各 field_type の値が round-trip で保持される（boolean / number+status / multi_select）。
func TestCheckupFieldResultRepository_ReplaceForCheckup_RoundTripsFieldTypes(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	fieldRepo := NewCheckupTypeFieldRepository(db)
	resultRepo := NewCheckupFieldResultRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "健診飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "健診ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-CHKFIELD", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")

	maxScore := 4.0
	fBool := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石除去必要の有無", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1})
	fNum := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "歯石付着度", FieldType: model.CheckupFieldTypeNumber, MaxValue: &maxScore, SortOrder: 2})
	fMulti := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "ケアアドバイス", FieldType: model.CheckupFieldTypeMultiSelect, Options: datatypes.JSON([]byte(`[{"value":"brush","label":"歯磨き"}]`)), SortOrder: 3})

	// FindByCheckupTypeID は sort_order 昇順で全フィールドを返す。
	fields, err := fieldRepo.FindByCheckupTypeID(ctx, clinicA, ct.ID)
	require.NoError(t, err)
	require.Len(t, fields, 3)
	assert.Equal(t, model.CheckupFieldTypeBoolean, fields[0].FieldType)

	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	boolTrue := true
	score := 5.0 // max(4) 超過 → high
	saved, _, err := resultRepo.ReplaceForCheckup(ctx, clinicA, checkupID, []model.CheckupFieldResult{
		{CheckupTypeFieldID: &fBool.ID, FieldName: fBool.Name, FieldType: model.CheckupFieldTypeBoolean, ValueBool: &boolTrue, SortOrder: 1},
		{CheckupTypeFieldID: &fNum.ID, FieldName: fNum.Name, FieldType: model.CheckupFieldTypeNumber, ValueNumber: &score, RefMax: &maxScore, IsAbnormal: true, Status: model.ExaminationResultStatusHigh, SortOrder: 2},
		{CheckupTypeFieldID: &fMulti.ID, FieldName: fMulti.Name, FieldType: model.CheckupFieldTypeMultiSelect, ValueList: []string{"brush"}, SortOrder: 3},
	})
	require.NoError(t, err)
	require.Len(t, saved, 3)

	got, err := resultRepo.FindByCheckupID(ctx, clinicA, checkupID)
	require.NoError(t, err)
	require.Len(t, got, 3)

	// boolean
	require.NotNil(t, got[0].ValueBool)
	assert.True(t, *got[0].ValueBool)
	require.NotNil(t, got[0].CheckupTypeField, "同一 clinic の field 定義は Preload される")
	// number + abnormal status
	require.NotNil(t, got[1].ValueNumber)
	assert.InDelta(t, 5.0, *got[1].ValueNumber, 0.001)
	assert.Equal(t, model.ExaminationResultStatusHigh, got[1].Status)
	assert.True(t, got[1].IsAbnormal)
	// multi_select
	assert.Equal(t, []string{"brush"}, []string(got[2].ValueList))

	// 置換セマンティクス: 再 Replace で件数が入れ替わる（全削除→挿入）。
	saved2, _, err := resultRepo.ReplaceForCheckup(ctx, clinicA, checkupID, []model.CheckupFieldResult{
		{CheckupTypeFieldID: &fBool.ID, FieldName: fBool.Name, FieldType: model.CheckupFieldTypeBoolean, ValueBool: &boolTrue, SortOrder: 1},
	})
	require.NoError(t, err)
	assert.Len(t, saved2, 1, "PUT 置換: 既存3件が1件に入れ替わる")
}

// SC4: pet 単位 read のマスタ Preload は clinic_id でスコープされ、別 clinic の field 定義が混入しない。
func TestCheckupFieldResultRepository_FindByPetID_FieldPreloadClinicIsolation(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	resultRepo := NewCheckupFieldResultRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "隔離飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "隔離ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-ISO", time.Now())
	ctA := makeCheckupTypeMaster(t, db, clinicA, "医院Aの歯科検診")

	fieldA := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ctA.ID, Name: "医院Aの項目", FieldType: model.CheckupFieldTypeBoolean})
	// clinic B の field 定義（クロステナント汚染の植え付け対象）。
	fieldB := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicB, CheckupTypeID: ctA.ID, Name: "医院Bの項目", FieldType: model.CheckupFieldTypeBoolean})

	checkupID := makeCheckupRec(t, db, clinicA, mrA.ID, petA.ID, ctA.ID)

	// 直接 DB に汚染行を挿入（write 側ガードをバイパスした過去データ/書込漏れを模倣）。
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &fieldB.ID,
		FieldName: "snapshot", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID: clinicA, CheckupID: checkupID, CheckupTypeFieldID: &fieldA.ID,
		FieldName: "snapshot", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 2,
	}).Error)

	got, err := resultRepo.FindByPetID(ctx, clinicA, petA.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)

	byFieldID := map[uint64]model.CheckupFieldResult{}
	for _, r := range got {
		require.NotNil(t, r.CheckupTypeFieldID)
		byFieldID[*r.CheckupTypeFieldID] = r
	}
	assert.Nil(t, byFieldID[fieldB.ID].CheckupTypeField, "別クリニックの field 定義マスタが Preload で混入してはならない")
	require.NotNil(t, byFieldID[fieldA.ID].CheckupTypeField, "同一クリニックの field 定義は Preload されるべき")
	assert.Equal(t, "医院Aの項目", byFieldID[fieldA.ID].CheckupTypeField.Name)
}

func TestCheckupFieldResultRepository_FindByPetID_RejectsCorruptCrossClinicPetRelation(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	resultRepo := NewCheckupFieldResultRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "健診結果取得対象飼主")
	crossClinicPet := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "破損した別医院ペット")
	mr := makeHistoryMedicalRecord(t, db, clinicA, crossClinicPet.ID, "MR-CORRUPT-CHECKUP", time.Now())
	checkupType := makeCheckupTypeMaster(t, db, clinicA, "医院Aの健診")
	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, crossClinicPet.ID, checkupType.ID)
	require.NoError(t, db.WithContext(ctx).Create(&model.CheckupFieldResult{
		ClinicID:  clinicA,
		CheckupID: checkupID,
		FieldName: "別医院ペット由来結果",
		FieldType: model.CheckupFieldTypeBoolean,
		SortOrder: 1,
	}).Error)

	got, err := resultRepo.FindByPetID(ctx, clinicA, crossClinicPet.ID)

	require.NoError(t, err)
	assert.Empty(t, got)
}

// SC3 正本ランタイム保証（security review M2）: FindByCheckupTypeID は clinic スコープで、
// 別 clinic の field 定義（同一 checkup_type_id を指すクロステナント汚染行）を返さない。
// service の #124 同型ガード（fieldByID メンバシップ検証）はこの clinic スコープに依拠するため、
// モックでなく実 SQL で保証する。FindByCheckupTypeID の clinicScope を temp-revert すると汚染行が漏れ RED。
func TestCheckupTypeFieldRepository_FindByCheckupTypeID_ClinicScoped(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	fieldRepo := NewCheckupTypeFieldRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ctA := makeCheckupTypeMaster(t, db, clinicA, "医院Aの歯科検診")
	fieldA := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ctA.ID, Name: "医院Aの項目", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 1})
	// clinic B が clinic A の checkup_type を指す汚染フィールド（write 漏れ/過去データを模倣）。
	fieldB := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicB, CheckupTypeID: ctA.ID, Name: "医院Bの項目", FieldType: model.CheckupFieldTypeBoolean, SortOrder: 2})

	got, err := fieldRepo.FindByCheckupTypeID(ctx, clinicA, ctA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1, "clinic A の field のみ返るべき（別 clinic の汚染行は除外）")
	assert.Equal(t, fieldA.ID, got[0].ID)
	for _, f := range got {
		assert.NotEqual(t, fieldB.ID, f.ID, "別クリニックの field 定義が混入してはならない")
		assert.Equal(t, clinicA, f.ClinicID)
	}
}

// FU3（#211 follow-up）: 残り 3 field_type（single_select / checklist / text）の round-trip。
// 既存テストが number / boolean / multi_select を網羅済みのため、機構が主張する 6 型を全て検証する。
// 値カラム規約（checkup_field_result_service.go）: single_select / text → value_text、checklist → value_list。
func TestCheckupFieldResultRepository_ReplaceForCheckup_RoundTripsRemainingFieldTypes(t *testing.T) {
	db := setupCheckupFieldTestDB(t)
	resultRepo := NewCheckupFieldResultRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "残型飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "残型ポチ")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-FIELDTYPES", time.Now())
	ct := makeCheckupTypeMaster(t, db, clinicA, "歯科検診")

	fSingle := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "総合評価", FieldType: model.CheckupFieldTypeSingleSelect, SortOrder: 1})
	fCheck := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "処置チェック", FieldType: model.CheckupFieldTypeChecklist, SortOrder: 2})
	fText := makeCheckupTypeField(t, db, &model.CheckupTypeField{ClinicID: clinicA, CheckupTypeID: ct.ID, Name: "所見", FieldType: model.CheckupFieldTypeText, SortOrder: 3})

	checkupID := makeCheckupRec(t, db, clinicA, mr.ID, pet.ID, ct.ID)

	saved, _, err := resultRepo.ReplaceForCheckup(ctx, clinicA, checkupID, []model.CheckupFieldResult{
		{CheckupTypeFieldID: &fSingle.ID, FieldName: fSingle.Name, FieldType: model.CheckupFieldTypeSingleSelect, ValueText: "要経過観察", SortOrder: 1},
		{CheckupTypeFieldID: &fCheck.ID, FieldName: fCheck.Name, FieldType: model.CheckupFieldTypeChecklist, ValueList: []string{"スケーリング", "ポリッシング"}, SortOrder: 2},
		{CheckupTypeFieldID: &fText.ID, FieldName: fText.Name, FieldType: model.CheckupFieldTypeText, ValueText: "下顎臼歯に軽度の歯肉炎あり。次回再評価。", SortOrder: 3},
	})
	require.NoError(t, err)
	require.Len(t, saved, 3)

	got, err := resultRepo.FindByCheckupID(ctx, clinicA, checkupID)
	require.NoError(t, err)
	require.Len(t, got, 3)

	// single_select → value_text に単一選択値が保持される。
	assert.Equal(t, model.CheckupFieldTypeSingleSelect, got[0].FieldType)
	assert.Equal(t, "要経過観察", got[0].ValueText)
	// checklist → value_list に複数チェック値が保持される。
	assert.Equal(t, model.CheckupFieldTypeChecklist, got[1].FieldType)
	assert.Equal(t, []string{"スケーリング", "ポリッシング"}, []string(got[1].ValueList))
	// text → value_text に自由記述が保持される。
	assert.Equal(t, model.CheckupFieldTypeText, got[2].FieldType)
	assert.Equal(t, "下顎臼歯に軽度の歯肉炎あり。次回再評価。", got[2].ValueText)
}
