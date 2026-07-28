package owner

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestOwnerRepository_PetsInsurance_CrossClinicPreloadIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	insB := testdb.MakeInsurance(t, db, clinicB, "医院Bの保険(owner)")
	ownerCross := makeTestOwner(t, db, clinicA, "越境保険飼主")
	testdb.MakePetWithInsurance(t, db, clinicA, ownerCross.ID, &insB.ID, "越境保険ペット(owner)")

	// (i) FindByID(単一)で nested Pets.Insurance に別テナント保険が混入しない
	got, err := repo.FindByID(ctx, clinicA, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, got.Pets, 1)
	assert.Nil(t, got.Pets[0].Insurance, "別クリニックの保険マスタが Pets.Insurance に混入してはならない")

	// (ii) #86 [A,B] なら B の保険は見える
	gotBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, gotBoth.Pets, 1)
	require.NotNil(t, gotBoth.Pets[0].Insurance, "#86 認可済みなら B の保険は見えるべき")
	assert.Equal(t, insB.ID, gotBoth.Pets[0].Insurance.ID)

	// (iii) #86 [A] のみなら隠れる
	gotA, err := repo.FindByIDForClinics(ctx, []uint64{clinicA}, ownerCross.ID)
	require.NoError(t, err)
	require.Len(t, gotA.Pets, 1)
	assert.Nil(t, gotA.Pets[0].Insurance, "認可外 clinic の保険は #86 でも漏れてはならない")
}

func TestOwnerRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	// clinic A に飼主を1件作成（ペットなし → Pets preload は空リストを返す）。
	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")

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

func TestOwnerRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "更新テスト飼主")

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		updated, err := repo.UpdateAndFind(
			ctx,
			clinicB,
			ownerA.ID,
			map[string]any{"name": "不正書き換え"},
		)
		require.Error(t, err, "clinic B から clinic A の owner を更新できてはならない")
		assert.Nil(t, updated)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("行が変更されていない（データ改ざん防止）", func(t *testing.T) {
		// clinic A で読み取り、名前が変更されていないことを確認する。
		got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新テスト飼主", got.Name, "別クリニックからの Update で名前が変わってはならない")
	})

	t.Run("正しいクリニックIDからの Update は成功する", func(t *testing.T) {
		updated, err := repo.UpdateAndFind(
			ctx,
			clinicA,
			ownerA.ID,
			map[string]any{"name": "正常更新後の名前"},
		)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "正常更新後の名前", updated.Name)
	})
}

func TestOwnerRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "削除テスト飼主")

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

func TestOwnerRepository_FindByID_UsesAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "owner ambient find",
		Email:    "owner-ambient-find@example.com",
	}
	rollbackOuter := errors.New("rollback outer transaction after ambient find")

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := persistence.TxFromContext(txCtx)
		require.NotNil(t, ambientTx)
		require.NoError(t, ambientTx.WithContext(txCtx).Create(owner).Error)

		loaded, findErr := newTestRepository(db).FindByID(txCtx, clinicID, owner.ID)
		require.NoError(t, findErr)
		require.Equal(t, owner.ID, loaded.ID)
		require.Equal(t, owner.Email, loaded.Email)
		return rollbackOuter
	})
	require.ErrorIs(t, err, rollbackOuter)

	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", owner.Email).
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount, "outer rollback must remove the owner observed through FindByID")
}

func TestOwnerRepository_PetsOwnerPreload_HasClinicScopeContract(t *testing.T) {
	source, err := os.ReadFile("repository.go")
	require.NoError(t, err)
	assert.Contains(
		t,
		string(source),
		`Preload("Pets.Owner", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs)`,
		"Pets.Ownerはouter Owner/Petsの写像に依存せずclinic scopeを明示すべき",
	)
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

func TestOwnerRepository_PetsPreload_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	species := makeOwnerPetRelationshipTestSpecies(t, db)
	legitimatePet := makeOwnerPetRelationshipTestPet(
		t, db, clinicA, ownerA.ID, species.ID, "正規clinicのペット",
	)
	crossClinicPet := makeOwnerPetRelationshipTestPet(
		t, db, clinicB, ownerA.ID, species.ID, "破損clinic関連のペット",
	)
	legitimatePetB := makeOwnerPetRelationshipTestPet(
		t, db, clinicB, ownerB.ID, species.ID, "医院Bの正規関連ペット",
	)

	owners, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "")
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, owners, 1)
	require.Len(t, owners[0].Pets, 1, "FindAllは認可外clinicのPetをpreloadしてはならない")
	assert.Equal(t, legitimatePet.ID, owners[0].Pets[0].ID)

	ownersForBoth, totalForBoth, err := repo.FindAll(ctx, []uint64{clinicA, clinicB}, 1, 100, "")
	require.NoError(t, err)
	require.Equal(t, int64(2), totalForBoth)
	require.Len(t, ownersForBoth, 2)
	petsByOwnerID := make(map[uint64][]model.Pet, len(ownersForBoth))
	for i := range ownersForBoth {
		petsByOwnerID[ownersForBoth[i].ID] = ownersForBoth[i].Pets
	}
	require.Len(t, petsByOwnerID[ownerA.ID], 2)
	assert.ElementsMatch(t, []uint64{legitimatePet.ID, crossClinicPet.ID}, []uint64{
		petsByOwnerID[ownerA.ID][0].ID,
		petsByOwnerID[ownerA.ID][1].ID,
	})
	require.Len(t, petsByOwnerID[ownerB.ID], 1)
	assert.Equal(t, legitimatePetB.ID, petsByOwnerID[ownerB.ID][0].ID)

	got, err := repo.FindByID(ctx, clinicA, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, got.Pets, 1, "FindByIDは認可外clinicのPetをpreloadしてはならない")
	assert.Equal(t, legitimatePet.ID, got.Pets[0].ID)
	require.NotNil(t, got.Pets[0].Owner)
	assert.Equal(t, clinicA, got.Pets[0].Owner.ClinicID)

	gotForBoth, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerA.ID)
	require.NoError(t, err)
	require.Len(t, gotForBoth.Pets, 2, "明示的に両clinicを認可した#86経路は既存semanticsを維持する")
	assert.ElementsMatch(t, []uint64{legitimatePet.ID, crossClinicPet.ID}, []uint64{
		gotForBoth.Pets[0].ID,
		gotForBoth.Pets[1].ID,
	})
	for i := range gotForBoth.Pets {
		require.NotNil(t, gotForBoth.Pets[i].Owner)
		assert.Equal(t, clinicA, gotForBoth.Pets[i].Owner.ClinicID)
	}

	gotOwnerB, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, ownerB.ID)
	require.NoError(t, err)
	require.Len(t, gotOwnerB.Pets, 1)
	assert.Equal(t, legitimatePetB.ID, gotOwnerB.Pets[0].ID)
	require.NotNil(t, gotOwnerB.Pets[0].Owner, "認可集合の2番目のclinicでもPets.Ownerをpreloadすべき")
	assert.Equal(t, clinicB, gotOwnerB.Pets[0].Owner.ClinicID)
}
