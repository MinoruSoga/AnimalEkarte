package medicalrecord

// diagnosis_name_repository_test.go — DiagnosisNameRepository の統合テスト。
//
// 移動元: internal/repository/diagnosisname/repository_test.go（BE8-4 batch28）→ BE9-2C roll-up。
//
// makeDiagnosisTypeMaster/makeDiagnosisNameRec は diagnosis_type_repository_test.go で定義
// 済み（旧 diagnosistype/diagnosisname 両サブパッケージが import cycle 回避のため独立に複製
// していたものを、BE9-2C で同一 medicalrecord パッケージへ統合したため一本化 — ここでの
// 再定義は削除している）。makeSpeciesAndPet/makeHistoryMedicalRecord は consultation/checkup
// 等と同様の局所複製として、このファイルにのみ残す。
//
// 保護する不変条件:
//   - FindAll/FindAllByCategoryID/FindByID/Update/Delete/Reorder は clinic_id で正しく分離される。
//   - FindAllByFilter は is_active=true のみ返す（CODE-QUALITY-232）。
//   - CountUsageByDiagnosisNameID は medical_records を JOIN してクリニック分離しつつ
//     diagnosis_name_id/diagnosis_2_name_id 両方をカウントし、ソフトデリート済み clinical_plan を除外する（P2）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupDiagnosisNameTestDB は diagnosis_types / diagnosis_names / clinical_plans 用に DB を整備する。
// CountUsageByDiagnosisNameID が medical_records を JOIN するため、medical_records.pet_id の FK を
// 満たせるよう pets/animal_species も migrate する（treatment_repository_test.go と同じ理由）。
func setupDiagnosisNameTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.DiagnosisType{}, &model.DiagnosisName{}, &model.ClinicalPlan{},
	))
	db.Exec("TRUNCATE TABLE clinical_plans CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_names CASCADE")
	db.Exec("TRUNCATE TABLE diagnosis_types CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeSpeciesAndPet はテスト用の AnimalSpecies と Pet を作成して返す。
func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
	t.Helper()
	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: species.ID,
		Name:            petName,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
	t.Helper()
	pet := petID
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     date,
		PetID:    &pet,
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func TestDiagnosisNameRepository_FindAll(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	n1 := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "診断名1")
	n2 := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "診断名2")
	typeB := makeDiagnosisTypeMaster(t, db, clinicB, "分類B")
	_ = makeDiagnosisNameRec(t, db, clinicB, typeB.ID, "医院Bの診断名")

	got, total, err := repo.FindAll(ctx, clinicA, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	ids := []uint64{got[0].ID, got[1].ID}
	assert.ElementsMatch(t, []uint64{n1.ID, n2.ID}, ids)
}

func TestDiagnosisNameRepository_FindAllByCategoryID(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	typeB := makeDiagnosisTypeMaster(t, db, clinicA, "分類B")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "分類A所属の診断名")
	_ = makeDiagnosisNameRec(t, db, clinicA, typeB.ID, "分類B所属の診断名")

	got, total, err := repo.FindAllByCategoryID(ctx, clinicA, typeA.ID, 1, 100)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, nameA.ID, got[0].ID)
}

func TestDiagnosisNameRepository_FindAllByFilter(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	typeB := makeDiagnosisTypeMaster(t, db, clinicA, "分類B")
	activeA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "有効な診断名A")
	activeB := makeDiagnosisNameRec(t, db, clinicA, typeB.ID, "有効な診断名B")

	inactive := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "無効な診断名")
	// bool フィールドに GORM の default タグ(true)がある場合、struct リテラルで false を
	// 指定しても zero-value と区別できず DB 側 DEFAULT(true) が採用されてしまう
	// （gorm@v1.30.0 callbacks/create.go の isZero→DefaultValueInterface 上書き挙動）。
	// そのため生成後に単一カラム Update で明示的に false へ変更する
	// （shift_entry_repository_test.go の Staff.IsActive 無効化と同じ手法）。
	require.NoError(t, db.WithContext(ctx).Model(&model.DiagnosisName{}).Where("id = ?", inactive.ID).Update("is_active", false).Error)

	t.Run("typeID指定なしは有効な全件を返す", func(t *testing.T) {
		got, err := repo.FindAllByFilter(ctx, clinicA, nil)
		require.NoError(t, err)
		ids := make([]uint64, len(got))
		for i, n := range got {
			ids[i] = n.ID
		}
		assert.ElementsMatch(t, []uint64{activeA.ID, activeB.ID}, ids)
	})

	t.Run("typeID指定で該当分類のみ返す", func(t *testing.T) {
		got, err := repo.FindAllByFilter(ctx, clinicA, &typeA.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, activeA.ID, got[0].ID)
	})

	t.Run("is_active=falseの診断名は除外される", func(t *testing.T) {
		got, err := repo.FindAllByFilter(ctx, clinicA, nil)
		require.NoError(t, err)
		for _, n := range got {
			assert.NotEqual(t, inactive.ID, n.ID)
		}
	})
}

func TestDiagnosisNameRepository_FindByID(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "診断名A")

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, nameA.ID)
		require.NoError(t, err)
		assert.Equal(t, nameA.ID, got.ID)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, 99999999)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからは取得できない", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, nameA.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisNameRepository_Create(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	n := &model.DiagnosisName{ClinicID: clinicA, DiagnosisTypeID: typeA.ID, Name: "新規診断名"}
	require.NoError(t, repo.Create(ctx, n))
	assert.NotZero(t, n.ID)

	got, err := repo.FindByID(ctx, clinicA, n.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規診断名", got.Name)
}

func TestDiagnosisNameRepository_Update(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "更新前")

	t.Run("同一クリニックからの更新は成功する", func(t *testing.T) {
		name := "更新後"
		got, err := repo.Update(ctx, clinicA, nameA.ID, UpdateDiagnosisNameInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "更新後", got.Name)
	})

	t.Run("別クリニックからの更新はNotFound", func(t *testing.T) {
		name := "不正"
		got, err := repo.Update(ctx, clinicB, nameA.ID, UpdateDiagnosisNameInput{Name: &name})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		name := "x"
		got, err := repo.Update(ctx, clinicA, 99999999, UpdateDiagnosisNameInput{Name: &name})
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisNameRepository_Delete(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "削除対象")

	t.Run("別クリニックからの削除はNotFoundで実際には削除されない", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, nameA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, nameA.ID)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("同一クリニックからの削除は成功しソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, nameA.ID))
		got, err := repo.FindByID(ctx, clinicA, nameA.ID)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.DiagnosisName
		require.NoError(t, db.Unscoped().Where("id = ?", nameA.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid)
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 99999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestDiagnosisNameRepository_Reorder(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	n1 := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "名前1")
	n2 := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "名前2")

	require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{n2.ID, n1.ID}))

	got, _, err := repo.FindAll(ctx, clinicA, 1, 100)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, n2.ID, got[0].ID)
	assert.Equal(t, n1.ID, got[1].ID)
}

func TestDiagnosisNameRepository_CountUsageByDiagnosisNameID(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "診断使用状況飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "診断使用状況ペット")
	typeA := makeDiagnosisTypeMaster(t, db, clinicA, "分類A")
	nameA := makeDiagnosisNameRec(t, db, clinicA, typeA.ID, "使用状況診断名")

	t.Run("未使用は0件", func(t *testing.T) {
		count, err := repo.CountUsageByDiagnosisNameID(ctx, clinicA, nameA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	mr1 := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-DIAG-1", time.Now())
	plan1 := &model.ClinicalPlan{MedicalRecordID: mr1.ID, DiagnosisNameID: &nameA.ID}
	require.NoError(t, db.WithContext(ctx).Create(plan1).Error)

	mr2 := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-DIAG-2", time.Now())
	plan2 := &model.ClinicalPlan{MedicalRecordID: mr2.ID, Diagnosis2NameID: &nameA.ID}
	require.NoError(t, db.WithContext(ctx).Create(plan2).Error)

	t.Run("diagnosis_name_idとdiagnosis_2_name_id両方をカウントする", func(t *testing.T) {
		count, err := repo.CountUsageByDiagnosisNameID(ctx, clinicA, nameA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ソフトデリートされたclinical_planは除外される", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(&model.ClinicalPlan{}, plan1.ID).Error)
		count, err := repo.CountUsageByDiagnosisNameID(ctx, clinicA, nameA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックからのカウントは0件（medical_recordsのclinic_idで隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByDiagnosisNameID(ctx, clinicB, nameA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
