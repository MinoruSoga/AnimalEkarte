package medicalrecord

// medicine_dose_param_clinic_isolation_test.go — #201 / #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: MedicineDoseParamRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B の投与量パラメータを
//   読み取れない（計算時のクロステナント参照拒否）、更新できない、削除できない。
//
// このテストは clinicScope を repository から削除すると必ず失敗するよう設計されている。
// 子テーブルが clinic_id を非正規化保持し clinicScope を直適用していることの担保。

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var (
	medicineDoseParamSchemaOnce sync.Once
	medicineDoseParamSchemaErr  error
)

// setupMedicineDoseParamIsolationTestDB は dose param 隔離テスト用に新 ENUM を作成し AutoMigrate する。
// setupTestDB は 001 の ENUM のみ作成するため、#201 で追加した ENUM はここで作成する。
func setupMedicineDoseParamIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// DROP TABLE/DROP TYPE を含む破壊的リセットはプロセス全体で一度だけ実行する（sync.Once）。
	// setupTestDB が全テストで DB 接続プールを共有するため、毎テスト DROP TYPE すると別テストが
	// 保持する古い型 OID 参照のキャッシュ済み prepared statement が "cache lookup failed"
	// (SQLSTATE XX000) で壊れる。
	medicineDoseParamSchemaOnce.Do(func() {
		medicineDoseParamSchemaErr = setupMedicineDoseParamSchema(db)
	})
	if medicineDoseParamSchemaErr != nil {
		t.Fatalf("failed to set up medicine dose param schema: %v", medicineDoseParamSchemaErr)
	}
	db.Exec("TRUNCATE TABLE medicine_dose_params CASCADE")
	db.Exec("TRUNCATE TABLE medicines CASCADE")
	return db
}

// setupMedicineDoseParamSchema は dose param 用テーブルを準備する。
// medicineDoseParamSchemaOnce 経由でプロセス全体につき一度だけ呼ばれる。
//
// medicine_calculation_type/medicine_dose_basis/medicine_rounding_mode/medicine_dose_species の
// 4 ENUM は setupTestDB（ltv_repository_test.go の共有 enumTypes）が既に idempotent に作成済みのため、
// ここで DROP+CREATE しない（以前はここでも作成していたが #201 の対応で共有 setupTestDB 側に統合済み・
// 残っていた重複コード。共有 ENUM を再作成すると、medicines.calculation_type 等を通じて既に
// この型を参照した他テストのキャッシュ済み prepared statement が壊れる）。
func setupMedicineDoseParamSchema(db *gorm.DB) error {
	// 永続テスト DB の残存行対策（既知 gotcha）: medicine_dose_params.species は NOT NULL（default なし）
	// のため、AutoMigrate 前に残存行があると ADD COLUMN で 23502 になる。DROP TABLE で fresh 作成する。
	if err := db.Exec("DROP TABLE IF EXISTS medicine_dose_params CASCADE").Error; err != nil {
		return fmt.Errorf("failed to drop medicine_dose_params: %w", err)
	}
	// medicines は calculation_type に DEFAULT 'none' があるため AutoMigrate の ADD NOT NULL は残存行でも成功する。
	// DROP 直後なので ensureAutoMigrated は使わない（キャッシュ済み型だと DROP 後の再 CREATE が抑止される）。
	if err := db.AutoMigrate(&model.Medicine{}, &model.MedicineDoseParam{}); err != nil {
		return fmt.Errorf("failed to migrate medicine dose param schema: %w", err)
	}
	testdb.MarkAutoMigrated(&model.Medicine{}, &model.MedicineDoseParam{})
	return nil
}

func makeDoseTestMedicine(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Medicine {
	t.Helper()
	strength := 10.0
	unit := model.MedicineUnitPerTablet
	m := &model.Medicine{
		ClinicID:        clinicID,
		Name:            name,
		TaxType:         model.TaxTypeExcluded, // not null 列を明示（GORM default 依存を避ける）
		CalculationType: model.MedicineCalculationTypePerWeight,
		MedicineUnit:    &unit,
		Strength:        &strength,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(m).Error)
	return m
}

func makeDoseParam(t *testing.T, db *gorm.DB, clinicID, medicineID uint64, species model.MedicineDoseSpecies) *model.MedicineDoseParam {
	t.Helper()
	maxRate := 10.0
	p := &model.MedicineDoseParam{
		ClinicID:   clinicID,
		MedicineID: medicineID,
		Species:    species,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: &maxRate, // per_weight には上限必須（患者安全）
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

// TestMedicineDoseParamRepository_FindByMedicineAndSpecies_ClinicIsolation は
// 計算時ルックアップが別クリニックのパラメータを取得できないことを検証する（クロステナント計算拒否）。
func TestMedicineDoseParamRepository_FindByMedicineAndSpecies_ClinicIsolation(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	medA := makeDoseTestMedicine(t, db, clinicA, "医院Aの薬剤")
	paramA := makeDoseParam(t, db, clinicA, medA.ID, model.MedicineDoseSpeciesDog)

	t.Run("同一クリニックでは取得できる", func(t *testing.T) {
		got, err := repo.FindByMedicineAndSpecies(ctx, clinicA, medA.ID, model.MedicineDoseSpeciesDog)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, paramA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックからは取得できない（クロステナント計算拒否＝fail-closed）", func(t *testing.T) {
		got, err := repo.FindByMedicineAndSpecies(ctx, clinicB, medA.ID, model.MedicineDoseSpeciesDog)
		assert.Error(t, err, "clinic B から clinic A の dose param を取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestMedicineDoseParamRepository_Create_ForcesClinicID は repo.Create が
// 呼び出し側がセットした param.ClinicID を信頼せず、引数の clinicID で上書きすることを検証する
// （security review HIGH-1: クロステナント書込の防止）。
func TestMedicineDoseParamRepository_Create_ForcesClinicID(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	medA := makeDoseTestMedicine(t, db, clinicA, "医院Aの薬剤")
	maxRate := 10.0

	// 悪意ある（または bug の）呼び出し: param.ClinicID に別クリニック B を仕込む。
	param := &model.MedicineDoseParam{
		ClinicID:   clinicB, // ← 上書きされるべき
		MedicineID: medA.ID,
		Species:    model.MedicineDoseSpeciesDog,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: &maxRate,
	}
	require.NoError(t, repo.Create(ctx, clinicA, param))

	t.Run("clinic_id は引数の clinicA で上書きされる", func(t *testing.T) {
		assert.Equal(t, clinicA, param.ClinicID)
	})
	t.Run("clinicA スコープで取得できる", func(t *testing.T) {
		got, err := repo.FindByMedicineAndSpecies(ctx, clinicA, medA.ID, model.MedicineDoseSpeciesDog)
		require.NoError(t, err)
		assert.Equal(t, clinicA, got.ClinicID)
	})
	t.Run("clinicB スコープでは見えない", func(t *testing.T) {
		_, err := repo.FindByMedicineAndSpecies(ctx, clinicB, medA.ID, model.MedicineDoseSpeciesDog)
		assert.True(t, apperrors.IsNotFound(err), "clinic B から見えてはならない: %v", err)
	})
}

// TestMedicineDoseParamRepository_FindByMedicineID_ClinicIsolation は一覧が別クリニックで空になることを検証する。
func TestMedicineDoseParamRepository_FindByMedicineID_ClinicIsolation(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	medA := makeDoseTestMedicine(t, db, clinicA, "医院Aの薬剤")
	makeDoseParam(t, db, clinicA, medA.ID, model.MedicineDoseSpeciesDog)
	makeDoseParam(t, db, clinicA, medA.ID, model.MedicineDoseSpeciesCat)

	// clinic B も自院の薬剤 + パラメータを持つ（混入なしを実データで検証）。
	medB := makeDoseTestMedicine(t, db, clinicB, "医院Bの薬剤")
	makeDoseParam(t, db, clinicB, medB.ID, model.MedicineDoseSpeciesDog)

	t.Run("clinicA は自院の2件のみ", func(t *testing.T) {
		got, err := repo.FindByMedicineID(ctx, clinicA, medA.ID)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("clinicB は自院の1件のみ", func(t *testing.T) {
		got, err := repo.FindByMedicineID(ctx, clinicB, medB.ID)
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("clinicB から clinicA の薬剤は0件（混入なし）", func(t *testing.T) {
		got, err := repo.FindByMedicineID(ctx, clinicB, medA.ID)
		require.NoError(t, err)
		assert.Empty(t, got, "clinic B から clinic A の dose param が見えてはならない")
	})
}

// TestMedicineDoseParamRepository_Update_ClinicIsolation は別クリニックからの更新が拒否され行が不変なことを検証する。
func TestMedicineDoseParamRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	medA := makeDoseTestMedicine(t, db, clinicA, "医院Aの薬剤")
	paramA := makeDoseParam(t, db, clinicA, medA.ID, model.MedicineDoseSpeciesDog)

	t.Run("別クリニックからの Update は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, paramA.ID, map[string]any{"dose_per_kg": 999})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("dose_per_kg が改ざんされていない", func(t *testing.T) {
		got, err := repo.FindByMedicineAndSpecies(ctx, clinicA, medA.ID, model.MedicineDoseSpeciesDog)
		require.NoError(t, err)
		assert.Equal(t, 5.0, got.DosePerKg, "別クリニックからの Update で dose_per_kg が変わってはならない")
	})
}

// TestMedicineDoseParamRepository_Delete_ClinicIsolation は別クリニックからの削除が拒否され行が残ることを検証する。
func TestMedicineDoseParamRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	medA := makeDoseTestMedicine(t, db, clinicA, "医院Aの薬剤")
	paramA := makeDoseParam(t, db, clinicA, medA.ID, model.MedicineDoseSpeciesDog)

	t.Run("別クリニックからの Delete は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, paramA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("パラメータはまだ存在する", func(t *testing.T) {
		got, err := repo.FindByMedicineAndSpecies(ctx, clinicA, medA.ID, model.MedicineDoseSpeciesDog)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, paramA.ID, got.ID)
	})
}
