package medicalrecord

// medicine_dose_param_repository_test.go — MedicineDoseParamRepository の統合テスト（実 Postgres テスト DB）。
//
// medicine_dose_param_clinic_isolation_test.go が clinic_id 隔離（FindByMedicineAndSpecies /
// Create の ClinicID 強制上書き / FindByMedicineID / Update・Delete の別クリニック拒否）を既にカバーしているため、
// 本ファイルはそこで未カバーの happy path（同一クリニックでの Update/Delete 成功）・not-found ケース・
// 空結果・ソート順を対象とする。setupMedicineDoseParamIsolationTestDB / makeDoseTestMedicine /
// makeDoseParam は medicine_dose_param_clinic_isolation_test.go で定義済みのため再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicineDoseParamRepository_Update_HappyPathReturnsUpdatedFields(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	med := makeDoseTestMedicine(t, db, clinicID, "更新対象薬剤")
	param := makeDoseParam(t, db, clinicID, med.ID, model.MedicineDoseSpeciesDog)

	got, err := repo.Update(ctx, clinicID, param.ID, map[string]any{"dose_per_kg": 12.5, "notes": "更新済み"})
	require.NoError(t, err)
	assert.Equal(t, 12.5, got.DosePerKg)
	assert.Equal(t, "更新済み", got.Notes)
	assert.Equal(t, param.ID, got.ID)
}

func TestMedicineDoseParamRepository_Update_NotFoundForNonexistentID(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	_, err := repo.Update(ctx, clinicID, 999999, map[string]any{"dose_per_kg": 1.0})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
}

func TestMedicineDoseParamRepository_Delete_HappyPathRemovesRow(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	med := makeDoseTestMedicine(t, db, clinicID, "削除対象薬剤")
	param := makeDoseParam(t, db, clinicID, med.ID, model.MedicineDoseSpeciesCat)

	require.NoError(t, repo.Delete(ctx, clinicID, param.ID))

	_, err := repo.FindByMedicineAndSpecies(ctx, clinicID, med.ID, model.MedicineDoseSpeciesCat)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "削除後は NotFound であるべき: %v", err)
}

func TestMedicineDoseParamRepository_Delete_NotFoundForNonexistentID(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	err := repo.Delete(ctx, clinicID, 999999)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestMedicineDoseParamRepository_FindByMedicineAndSpecies_NotFoundWhenSpeciesMissing(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	med := makeDoseTestMedicine(t, db, clinicID, "犬のみ設定薬剤")
	makeDoseParam(t, db, clinicID, med.ID, model.MedicineDoseSpeciesDog)

	// 猫用パラメータは未設定
	_, err := repo.FindByMedicineAndSpecies(ctx, clinicID, med.ID, model.MedicineDoseSpeciesCat)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "未設定の種は NotFound であるべき: %v", err)
}

func TestMedicineDoseParamRepository_FindByMedicineID_EmptyWhenNoParams(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	med := makeDoseTestMedicine(t, db, clinicID, "未設定薬剤")

	got, err := repo.FindByMedicineID(ctx, clinicID, med.ID)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestMedicineDoseParamRepository_FindByMedicineID_OrderedBySpeciesAscending(t *testing.T) {
	db := setupMedicineDoseParamIsolationTestDB(t)
	repo := NewMedicineDoseParamRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	med := makeDoseTestMedicine(t, db, clinicID, "順序検証薬剤")
	// medicine_dose_species は Postgres ENUM 型（migrations/001_init.sql:
	// CREATE TYPE medicine_dose_species AS ENUM ('dog', 'cat')。旧 009_add_medicine_dose_params.sql
	// は 2026-07-04 に統合済み・docs/architecture/erd.md §4.3）であり、ENUM は文字列の
	// アルファベット順ではなく「型定義での宣言順」でソートされる。dog が cat より先に
	// 宣言されているため、cat を先に作成しても species ASC では dog が先頭に来る。
	makeDoseParam(t, db, clinicID, med.ID, model.MedicineDoseSpeciesCat)
	makeDoseParam(t, db, clinicID, med.ID, model.MedicineDoseSpeciesDog)

	got, err := repo.FindByMedicineID(ctx, clinicID, med.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, model.MedicineDoseSpeciesDog, got[0].Species, "species ASC は ENUM 宣言順（dog, cat）で dog が先頭")
	assert.Equal(t, model.MedicineDoseSpeciesCat, got[1].Species)
}
