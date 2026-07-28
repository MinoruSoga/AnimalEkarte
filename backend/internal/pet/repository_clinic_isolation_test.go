package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupPetRepositoryIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Insurance{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE insurances CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE animal_species CASCADE").Error)
	return db
}

func TestPetRepository_Insurance_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	insA := testdb.MakeInsurance(t, db, clinicA, "医院Aの保険")
	insB := testdb.MakeInsurance(t, db, clinicB, "医院Bの保険")
	ownerA := makeTestOwner(t, db, clinicA, "保険飼主A")
	petLegit := testdb.MakePetWithInsurance(t, db, clinicA, ownerA.ID, &insA.ID, "正規保険ペット")
	petCross := testdb.MakePetWithInsurance(t, db, clinicA, ownerA.ID, &insB.ID, "越境保険ペット") // 別clinicの保険FKを植え付け

	// (legit) 単一 clinic の正規保険は Preload される
	gotLegit, err := repo.FindByID(ctx, clinicA, petLegit.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegit.Insurance)
	assert.Equal(t, insA.ID, gotLegit.Insurance.ID)

	// (i) 別テナント保険は FindByID(単一)で混入しない
	gotCross, err := repo.FindByID(ctx, clinicA, petCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCross.Insurance, "別クリニックの保険マスタが混入してはならない")

	// (ii) #86: 認可集合 [A,B] なら B の保険は見える
	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, petCross.ID)
	require.NoError(t, err)
	require.NotNil(t, gotForBoth.Insurance, "#86 で B も認可済みなら B の保険は見えるべき")
	assert.Equal(t, insB.ID, gotForBoth.Insurance.ID)

	// (iii) #86: 認可集合 [A] のみなら B の保険は隠れる
	gotForAOnly, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, petCross.ID)
	require.NoError(t, err)
	assert.Nil(t, gotForAOnly.Insurance, "認可外 clinic の保険マスタは #86 でも漏れてはならない")
}

func TestPetRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主（ペット隔離用）")
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

func TestPetRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "ペット削除テスト飼主")
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

func makeOwnerPetRelationshipTestPet(
	t *testing.T,
	db *gorm.DB,
	clinicID, ownerID, speciesID uint64,
	name string,
) *model.Pet {
	t.Helper()
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: speciesID,
		Name:            name,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func makeOwnerPetRelationshipTestSpecies(t *testing.T, db *gorm.DB) *model.AnimalSpecies {
	t.Helper()
	species := &model.AnimalSpecies{Name: "owner-pet preload isolation species"}
	require.NoError(t, db.WithContext(context.Background()).Create(species).Error)
	return species
}

func TestPetRepository_OwnerPreload_ClinicIsolation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	species := makeOwnerPetRelationshipTestSpecies(t, db)
	legitimatePet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerA.ID, species.ID, "正規関連ペット",
	)
	crossClinicPet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerB.ID, species.ID, "越境関連ペット",
	)

	gotLegitimate, err := repo.FindByID(ctx, clinicA, legitimatePet.ID)
	require.NoError(t, err)
	require.NotNil(t, gotLegitimate.Owner)
	assert.Equal(t, ownerA.ID, gotLegitimate.Owner.ID)

	gotCrossClinic, err := repo.FindByID(ctx, clinicA, crossClinicPet.ID)
	require.NoError(t, err)
	assert.Nil(t, gotCrossClinic.Owner, "認可外clinicのOwnerをPet.Ownerへpreloadしてはならない")

	gotForAOnly, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, crossClinicPet.ID)
	require.NoError(t, err)
	assert.Nil(t, gotForAOnly.Owner, "単一clinicの認可集合でも他clinicのOwnerを隠すべき")

	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, crossClinicPet.ID)
	require.NoError(t, err)
	require.NotNil(t, gotForBoth.Owner, "明示的に両clinicを認可した#86経路は既存semanticsを維持する")
	assert.Equal(t, ownerB.ID, gotForBoth.Owner.ID)
}

func TestPetRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "医院Aのポチ")

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, petA.ID, map[string]any{"name": "改ざん"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックの Update は NotFound: %v", err)
	})

	t.Run("ペットの値は変更されていない（データ改ざん防止）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.Equal(t, "医院Aのポチ", got.Name, "越境 Update で値が書き換わってはならない")
	})

	t.Run("同一クリニックIDからの Update は成功する（false-reject なし）", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, petA.ID, map[string]any{"name": "正規更新"})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.Equal(t, "正規更新", got.Name)
	})
}
