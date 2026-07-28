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
	// BUG-454: multi-clinic authorization must not restore a cross-clinic broken FK graph
	// (intentional legacy contract change — authorized set alone is insufficient for Owner).
	assert.Nil(t, gotForBoth.Owner, "認可集合に両clinicが含まれても pet.clinic_id と不一致の Owner は復元してはならない")
}

// TestPetRepository_FindAll_CorruptedOwnerGraph_NoCrossClinicRestore は BUG-454 の回帰:
// pet(clinicA) が owner(clinicB) を指す破損 FK について、viewer が {A,B} を認可していても
// FindAll が Owner を Preload せず、ownerB 名での search/order JOIN 汚染も起きないこと。
func TestPetRepository_FindAll_CorruptedOwnerGraph_NoCrossClinicRestore(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの正規飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの越境飼主")
	// name_kana を明示し、破損 JOIN が order に影響しないことを検証できるようにする。
	require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerA.ID).Update("name_kana", "あああ").Error)
	require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerB.ID).Update("name_kana", "んんん").Error)

	species := makeOwnerPetRelationshipTestSpecies(t, db)
	legitimatePet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerA.ID, species.ID, "正規関連ペット",
	)
	crossClinicPet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerB.ID, species.ID, "越境関連ペット",
	)
	// clinicB 側の正規ペット（search が ownerB 名で正しくヒットする対照群）。
	legitimatePetB := makeOwnerPetRelationshipTestPet(
		t, db, clinicB, ownerB.ID, species.ID, "医院Bの正規ペット",
	)

	clinicIDs := []uint64{clinicA, clinicB}

	t.Run("FindAll は破損 FK の Owner を Preload しない", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, clinicIDs, PetListFilters{IncludeDeceased: true}, 1, 100)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(3))

		byID := make(map[uint64]model.Pet, len(pets))
		for i := range pets {
			byID[pets[i].ID] = pets[i]
		}

		legit, ok := byID[legitimatePet.ID]
		require.True(t, ok, "正規関連ペットは一覧に含まれる")
		require.NotNil(t, legit.Owner, "正規関連の Owner は Preload される")
		assert.Equal(t, ownerA.ID, legit.Owner.ID)

		cross, ok := byID[crossClinicPet.ID]
		require.True(t, ok, "破損 FK の pet 自体は pet.clinic_id 認可で一覧に残る")
		assert.Nil(t, cross.Owner, "pet(A)->owner(B) の Owner を multi-clinic viewer でも復元してはならない")

		legitB, ok := byID[legitimatePetB.ID]
		require.True(t, ok)
		require.NotNil(t, legitB.Owner)
		assert.Equal(t, ownerB.ID, legitB.Owner.ID)
	})

	t.Run("ownerB 名での search は破損 pet を JOIN 経由で surface しない", func(t *testing.T) {
		pets, _, err := repo.FindAll(ctx, clinicIDs, PetListFilters{
			Search:          ownerB.Name,
			IncludeDeceased: true,
		}, 1, 100)
		require.NoError(t, err)

		ids := make([]uint64, len(pets))
		for i := range pets {
			ids[i] = pets[i].ID
		}
		assert.Contains(t, ids, legitimatePetB.ID, "同一 clinic の正規関連は owner 名 search でヒットする")
		assert.NotContains(t, ids, crossClinicPet.ID, "破損 FK は owners JOIN 相関で search に乗ってはならない")
		assert.NotContains(t, ids, legitimatePet.ID, "無関係な正規 pet は ownerB 名でヒットしない")
	})

	t.Run("order は破損 owner の name_kana に依存しない", func(t *testing.T) {
		// 破損 pet の owners JOIN は相関不一致で NULL になるため name_kana は NULL。
		// PostgreSQL の ASC 既定は NULLS LAST のため、NULL kana の破損 pet は
		// ownerB（name_kana=んんん）の正規 pet より後ろに並ぶ。相関が無いと ownerB と
		// 同じ kana キーで並び、相対位置が崩れる。
		pets, _, err := repo.FindAll(ctx, clinicIDs, PetListFilters{IncludeDeceased: true}, 1, 100)
		require.NoError(t, err)

		idxCross := -1
		idxLegitB := -1
		for i := range pets {
			switch pets[i].ID {
			case crossClinicPet.ID:
				idxCross = i
			case legitimatePetB.ID:
				idxLegitB = i
			}
		}
		require.True(t, idxCross >= 0 && idxLegitB >= 0)
		// 相関後: 破損 pet の order キーは NULL（NULLS LAST）→ ownerB 正規 pet より後ろ。
		// 破損 JOIN が ownerB を拾うと name_kana=んんん で legitimatePetB と同じキーになり、
		// pets.id タイブレークのみになる（相対位置が id 依存に退化する）。
		assert.Greater(t, idxCross, idxLegitB,
			"相関後は NULL kana の破損 pet が ownerB 正規 pet より後ろに並ぶべき")
		assert.Nil(t, pets[idxCross].Owner)
		require.NotNil(t, pets[idxLegitB].Owner)
		assert.Equal(t, ownerB.ID, pets[idxLegitB].Owner.ID)
	})

	t.Run("FindByID / FindByIDForClinics も破損 Owner を復元しない", func(t *testing.T) {
		gotA, err := repo.FindByID(ctx, clinicA, crossClinicPet.ID)
		require.NoError(t, err)
		assert.Nil(t, gotA.Owner)

		gotBoth, err := repo.FindByIDForClinics(ctx, clinicIDs, crossClinicPet.ID)
		require.NoError(t, err)
		assert.Nil(t, gotBoth.Owner, "multi-clinic 認可でも pet.clinic_id 不一致 Owner は nil")

		gotLegit, err := repo.FindByIDForClinics(ctx, clinicIDs, legitimatePet.ID)
		require.NoError(t, err)
		require.NotNil(t, gotLegit.Owner)
		assert.Equal(t, ownerA.ID, gotLegit.Owner.ID)
	})
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
