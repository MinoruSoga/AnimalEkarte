package repository

// owner_pet_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: OwnerRepository / PetRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B のレコードを
//   読み取れない、更新できない、削除できない。
//
// このテストは clinicScope を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupOwnerPetIsolationTestDB は owner/pet clinic_id 隔離テスト用に DB を整備する。
// setupTestDB が owners/billings をクリアし、ここで animal_species/pets/insurances を追加する。
func setupOwnerPetIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.AnimalSpecies{}, &model.Pet{}, &model.Insurance{}))
	// setupTestDB の TRUNCATE owners CASCADE が pets を清掃するが、
	// animal_species と insurances は個別に初期化する。
	db.Exec("TRUNCATE TABLE insurances CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// TestOwnerRepository_FindByID_ClinicIsolation は
// clinic A の飼主を clinic B の clinicID で取得できないことを検証する。
// clinicScope を owner.go から削除すると「別クリニックIDでは取得できない」が失敗する。
func TestOwnerRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A に飼主を1件作成（ペットなし → Pets preload は空リストを返す）。
	ownerA := makeOwner(t, db, clinicA, "医院Aの飼主")

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		// clinicScope が有効なら clinic B から clinic A の owner は見えない。
		got, err := repo.FindByID(ctx, clinicB, ownerA.ID)
		assert.Error(t, err, "clinic B から clinic A の owner を取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestPetRepository_FindByID_ClinicIsolation は
// clinic A のペットを clinic B の clinicID で取得できないことを検証する。
func TestPetRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeOwner(t, db, clinicA, "医院Aの飼主（ペット隔離用）")
	// makeSpeciesAndPet は animal_species + pet を作成する。insurance_id は nil。
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ポチ（隔離テスト）")

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, petA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, petA.ID)
		assert.Error(t, err, "clinic B から clinic A のペットを取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestOwnerRepository_Update_ClinicIsolation は
// 別クリニックIDからの Update が NotFound を返し、行が変更されないことを検証する。
// clinicScope を owner.go から削除すると「行が変更されていない」が失敗する。
func TestOwnerRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeOwner(t, db, clinicA, "更新テスト飼主")

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, ownerA.ID, map[string]any{"name": "不正書き換え"})
		require.Error(t, err, "clinic B から clinic A の owner を更新できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("行が変更されていない（データ改ざん防止）", func(t *testing.T) {
		// clinic A で読み取り、名前が変更されていないことを確認する。
		got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新テスト飼主", got.Name, "別クリニックからの Update で名前が変わってはならない")
	})

	t.Run("正しいクリニックIDからの Update は成功する", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, ownerA.ID, map[string]any{"name": "正常更新後の名前"})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		assert.Equal(t, "正常更新後の名前", got.Name)
	})
}

// TestOwnerRepository_Delete_ClinicIsolation は
// 別クリニックIDからの Delete が NotFound を返し、行が削除されないことを検証する。
// clinicScope を owner.go から削除すると「owner はまだ存在する」が失敗する。
func TestOwnerRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewOwnerRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeOwner(t, db, clinicA, "削除テスト飼主")

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, ownerA.ID)
		require.Error(t, err, "clinic B から clinic A の owner を削除できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("owner はまだ存在する（不正削除防止）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		assert.NotNil(t, got, "clinic A の owner は別クリニックの Delete で消えてはならない")
		assert.Equal(t, ownerA.ID, got.ID)
	})
}

// TestPetRepository_Delete_ClinicIsolation は
// 別クリニックIDからの Delete が NotFound を返し、ペットが削除されないことを検証する。
func TestPetRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeOwner(t, db, clinicA, "ペット削除テスト飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "削除テストペット")

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, petA.ID)
		require.Error(t, err, "clinic B から clinic A のペットを削除できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("ペットはまだ存在する（不正削除防止）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.NotNil(t, got, "clinic A のペットは別クリニックの Delete で消えてはならない")
		assert.Equal(t, petA.ID, got.ID)
	})
}
