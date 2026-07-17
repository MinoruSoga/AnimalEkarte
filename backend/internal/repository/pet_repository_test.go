package repository

// pet_repository_test.go — PetRepository の統合テスト。
//
// FindByID / Delete の clinic_id 隔離は owner_pet_clinic_isolation_test.go、
// Insurance Preload の #86 隔離は cross_clinic_preload_isolation_test.go、
// Update の clinic_id 隔離は pet_write_medimage_clinic_isolation_test.go で別途カバー済みのため、
// 本ファイルは重複しない観点（検索・ページング・生存フィルタ・カウント系・誕生日検索など）に絞る。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupPetRepositoryTestDB は pet_repository のテスト用に DB を整備する。
// owners CASCADE (setupTestDB) が pets/pet_chronic_conditions を連鎖クリアするが、
// animal_species/insurances は個別に初期化する。
func setupPetRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Insurance{}))
	db.Exec("TRUNCATE TABLE insurances CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makePetDetailed はテスト用ペットを詳細フィールド込みで作成する。
func makePetDetailed(t *testing.T, db *gorm.DB, clinicID, ownerID, speciesID uint64, name string, birthDate, deceasedAt *time.Time) *model.Pet {
	t.Helper()
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		AnimalSpeciesID: speciesID,
		Name:            name,
		BirthDate:       birthDate,
		DeceasedAt:      deceasedAt,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(pet).Error)
	return pet
}

func TestPetRepository_FindAll(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA1 := makeTestOwner(t, db, clinicA, "検索飼主1")
	ownerA2 := makeTestOwner(t, db, clinicA, "検索飼主2")
	makeSpeciesAndPet(t, db, clinicA, ownerA1.ID, "ペットA1")
	petA2 := makeSpeciesAndPet(t, db, clinicA, ownerA1.ID, "ペットA2")
	petA3 := makeSpeciesAndPet(t, db, clinicA, ownerA2.ID, "ペットA3")
	makeSpeciesAndPet(t, db, clinicB, (makeTestOwner(t, db, clinicB, "検索飼主B")).ID, "ペットB1")

	t.Run("clinic スコープと総数", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, pets, 3)
		for _, p := range pets {
			assert.Equal(t, clinicA, p.ClinicID)
		}
	})

	t.Run("ページング", func(t *testing.T) {
		page1, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{}, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, page1, 2)

		page2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{}, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total2)
		assert.Len(t, page2, 1)
	})

	t.Run("ownerID フィルタ", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{OwnerID: &ownerA2.ID}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petA3.ID, pets[0].ID)
	})

	t.Run("ペット名の部分一致検索", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: "A2"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petA2.ID, pets[0].ID)
	})

	t.Run("飼主名での検索もヒットする", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: "検索飼主2"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petA3.ID, pets[0].ID)
	})

	t.Run("かな正規化検索(カタカナ検索でひらがな登録データにヒット)", func(t *testing.T) {
		ownerKana := makeTestOwner(t, db, clinicA, "ひらがな飼主")
		petKana := makeSpeciesAndPet(t, db, clinicA, ownerKana.ID, "かな検索用ペット")
		require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).Where("id = ?", petKana.ID).Update("name_kana", "ぽち").Error)

		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: "ポチ"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petKana.ID, pets[0].ID)
	})

	t.Run("飼主電話番号での検索もヒットする", func(t *testing.T) {
		ownerPhone := makeTestOwner(t, db, clinicA, "電話検索飼主")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerPhone.ID).Update("phone", "090-1111-2222").Error)
		petPhone := makeSpeciesAndPet(t, db, clinicA, ownerPhone.ID, "電話検索用ペット")

		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: "090-1111-2222"}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petPhone.ID, pets[0].ID)
	})

	t.Run("species フィルタ", func(t *testing.T) {
		speciesX := &model.AnimalSpecies{Name: "species フィルタ用種"}
		require.NoError(t, db.WithContext(ctx).Create(speciesX).Error)
		ownerSpecies := makeTestOwner(t, db, clinicA, "species フィルタ飼主")
		petSpeciesMatch := makePetDetailed(t, db, clinicA, ownerSpecies.ID, speciesX.ID, "species一致ペット", nil, nil)

		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{AnimalSpeciesID: &speciesX.ID}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petSpeciesMatch.ID, pets[0].ID)
	})

	t.Run("include_deceased 既定 false は死亡ペットを除外する", func(t *testing.T) {
		ownerDeceased := makeTestOwner(t, db, clinicA, "生死フィルタ飼主")
		deceasedAt := time.Now().Add(-24 * time.Hour)
		speciesID := makeSyncSpeciesID(t, db)
		alivePet := makePetDetailed(t, db, clinicA, ownerDeceased.ID, speciesID, "生死フィルタ生存ペット", nil, nil)
		deadPet := makePetDetailed(t, db, clinicA, ownerDeceased.ID, speciesID, "生死フィルタ死亡ペット", nil, &deceasedAt)

		pets, _, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{OwnerID: &ownerDeceased.ID}, 1, 10)
		require.NoError(t, err)
		ids := make([]uint64, len(pets))
		for i, p := range pets {
			ids[i] = p.ID
		}
		assert.Contains(t, ids, alivePet.ID)
		assert.NotContains(t, ids, deadPet.ID, "include_deceased 未指定(false)は死亡ペットを除外する")

		petsIncl, _, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{OwnerID: &ownerDeceased.ID, IncludeDeceased: true}, 1, 10)
		require.NoError(t, err)
		idsIncl := make([]uint64, len(petsIncl))
		for i, p := range petsIncl {
			idsIncl[i] = p.ID
		}
		assert.Contains(t, idsIncl, alivePet.ID)
		assert.Contains(t, idsIncl, deadPet.ID, "include_deceased=true は死亡ペットを含める")
	})

	t.Run("順序は owners.name_kana ASC, pets.id ASC で安定する", func(t *testing.T) {
		ownerZ := makeTestOwner(t, db, clinicA, "順序飼主Z")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerZ.ID).Update("name_kana", "われ").Error)
		ownerA := makeTestOwner(t, db, clinicA, "順序飼主A")
		require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerA.ID).Update("name_kana", "あお").Error)
		speciesID := makeSyncSpeciesID(t, db)
		petZ := makePetDetailed(t, db, clinicA, ownerZ.ID, speciesID, "順序ペットZ", nil, nil)
		petA1 := makePetDetailed(t, db, clinicA, ownerA.ID, speciesID, "順序ペットA1", nil, nil)
		petA2 := makePetDetailed(t, db, clinicA, ownerA.ID, speciesID, "順序ペットA2", nil, nil)

		pets, _, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{}, 1, 100)
		require.NoError(t, err)

		idxZ := indexOfPetID(pets, petZ.ID)
		idxA1 := indexOfPetID(pets, petA1.ID)
		idxA2 := indexOfPetID(pets, petA2.ID)
		require.True(t, idxA1 >= 0 && idxA2 >= 0 && idxZ >= 0)
		assert.Less(t, idxA1, idxZ, "kana 昇順: あお(A) は われ(Z) より前")
		assert.Less(t, idxA1, idxA2, "同一 owner 内は pets.id ASC でタイブレーク")
	})

	t.Run("owner サマリが埋め込まれる", func(t *testing.T) {
		pets, _, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{OwnerID: &ownerA1.ID}, 1, 10)
		require.NoError(t, err)
		require.NotEmpty(t, pets)
		require.NotNil(t, pets[0].Owner)
		assert.Equal(t, ownerA1.ID, pets[0].Owner.ID)
		assert.Equal(t, ownerA1.Name, pets[0].Owner.Name)
	})

	t.Run("#86 拠点横断: clinic_ids 複数指定で両医院のペットを返す", func(t *testing.T) {
		// 総数は先行 subtest 群が同一 DB にペットを追加し続けるため固定値では検証しない
		// （このファイルの既存慣習どおり subtest 間で DB をリセットしない）。
		// ここでは clinicA/clinicB 双方のペットが含まれることのみを確認する。
		pets, _, err := repo.FindAll(ctx, []uint64{clinicA, clinicB}, PetListFilters{}, 1, 100)
		require.NoError(t, err)
		clinics := make(map[uint64]bool)
		for _, p := range pets {
			clinics[p.ClinicID] = true
		}
		assert.True(t, clinics[clinicA])
		assert.True(t, clinics[clinicB])
	})

	t.Run("クロステナント隔離: 空スライスは0件", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{}, PetListFilters{}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, pets)
	})
}

func indexOfPetID(pets []model.Pet, id uint64) int {
	for i, p := range pets {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func TestPetRepository_FindByIDForClinics(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "拠点横断飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "拠点横断ペット")

	t.Run("認可済みclinic集合に含まれれば取得できる", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicA, clinicB}, petA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, petA.ID, got.ID)
	})

	t.Run("認可外clinic集合では取得できない", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{clinicB}, petA.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("空のclinic集合は即NotFound", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(ctx, []uint64{}, petA.ID)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPetRepository_FindLivingByOwner(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "生存フィルタ飼主")
	deceasedAt := time.Now().Add(-24 * time.Hour)
	alive1 := makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "生存ペット1", nil, nil)
	alive2 := makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "生存ペット2", nil, nil)
	dead := makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "死亡ペット", nil, &deceasedAt)
	softDeleted := makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "削除済みペット", nil, nil)
	require.NoError(t, db.WithContext(ctx).Delete(softDeleted).Error)

	got, err := repo.FindLivingByOwner(ctx, clinicA, owner.ID)
	require.NoError(t, err)

	ids := make([]uint64, 0, len(got))
	for _, p := range got {
		ids = append(ids, p.ID)
	}
	assert.Contains(t, ids, alive1.ID)
	assert.Contains(t, ids, alive2.ID)
	assert.NotContains(t, ids, dead.ID, "死亡ペットは除外される")
	assert.NotContains(t, ids, softDeleted.ID, "ソフト削除済みペットは除外される")
}

func TestPetRepository_CountByOwner(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "頭数カウント飼主")
	makeSpeciesAndPet(t, db, clinicA, owner.ID, "カウント対象1")
	makeSpeciesAndPet(t, db, clinicA, owner.ID, "カウント対象2")
	deleted := makeSpeciesAndPet(t, db, clinicA, owner.ID, "カウント除外(削除済み)")
	require.NoError(t, db.WithContext(ctx).Delete(deleted).Error)

	got, err := repo.CountByOwner(ctx, clinicA, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got)
}

func TestPetRepository_CountLivingByOwner(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "生存頭数飼主")
	deceasedAt := time.Now().Add(-24 * time.Hour)
	makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "生存中", nil, nil)
	makePetDetailed(t, db, clinicA, owner.ID, makeSyncSpeciesID(t, db), "死亡済み", nil, &deceasedAt)

	got, err := repo.CountLivingByOwner(ctx, clinicA, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got, "死亡ペットは除外して生存1頭のみ")
}

func TestPetRepository_CountLivingByOwnerIDs(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	t.Run("空スライスは空マップを即返す", func(t *testing.T) {
		got, err := repo.CountLivingByOwnerIDs(ctx, clinicA, []uint64{})
		require.NoError(t, err)
		assert.Equal(t, map[uint64]int64{}, got)
	})

	owner1 := makeTestOwner(t, db, clinicA, "一括カウント飼主1")
	owner2 := makeTestOwner(t, db, clinicA, "一括カウント飼主2")
	ownerNoPets := makeTestOwner(t, db, clinicA, "ペット無し飼主")

	makeSpeciesAndPet(t, db, clinicA, owner1.ID, "一括1-1")
	makeSpeciesAndPet(t, db, clinicA, owner1.ID, "一括1-2")
	deceasedAt := time.Now().Add(-24 * time.Hour)
	makePetDetailed(t, db, clinicA, owner1.ID, makeSyncSpeciesID(t, db), "一括1-死亡", nil, &deceasedAt)
	makeSpeciesAndPet(t, db, clinicA, owner2.ID, "一括2-1")

	t.Run("複数飼主を一括集計", func(t *testing.T) {
		got, err := repo.CountLivingByOwnerIDs(ctx, clinicA, []uint64{owner1.ID, owner2.ID, ownerNoPets.ID})
		require.NoError(t, err)
		assert.Equal(t, int64(2), got[owner1.ID], "死亡ペットを除いた生存数")
		assert.Equal(t, int64(1), got[owner2.ID])
		_, exists := got[ownerNoPets.ID]
		assert.False(t, exists, "ペット無し飼主はマップに現れない(呼び出し側で0扱い)")
	})
}

func TestPetRepository_CountUsageByAnimalSpeciesID(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	species := &model.AnimalSpecies{Name: "使用数テスト種"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	unusedSpecies := &model.AnimalSpecies{Name: "未使用種"}
	require.NoError(t, db.WithContext(ctx).Create(unusedSpecies).Error)

	owner := makeTestOwner(t, db, clinicA, "種別使用数飼主")
	makePetDetailed(t, db, clinicA, owner.ID, species.ID, "使用中ペット1", nil, nil)
	deletedPet := makePetDetailed(t, db, clinicA, owner.ID, species.ID, "使用中ペット2(削除済み)", nil, nil)
	require.NoError(t, db.WithContext(ctx).Delete(deletedPet).Error)

	t.Run("使用中は件数を返す(ソフト削除除外)", func(t *testing.T) {
		got, err := repo.CountUsageByAnimalSpeciesID(ctx, species.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), got)
	})

	t.Run("未使用は0", func(t *testing.T) {
		got, err := repo.CountUsageByAnimalSpeciesID(ctx, unusedSpecies.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})
}

func TestPetRepository_Create(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "新規作成飼主")
	species := &model.AnimalSpecies{Name: "新規作成種"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	pet := &model.Pet{ClinicID: clinicA, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "新規作成ペット"}
	err := repo.Create(ctx, pet)
	require.NoError(t, err)
	assert.NotZero(t, pet.ID)
	require.NotNil(t, pet.Owner, "Create 後は Owner が Preload されている")
	assert.Equal(t, owner.ID, pet.Owner.ID)
	require.NotNil(t, pet.AnimalSpecies)
}

func TestPetRepository_Update(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "更新テスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新前ペット名")

	t.Run("成功", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, pet.ID, map[string]any{"name": "更新後ペット名"})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後ペット名", got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 999888004, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPetRepository_Delete(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "削除テスト飼主(pet_repository)")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "削除対象ペット(pet_repository)")

	t.Run("成功", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, pet.ID)
		require.NoError(t, err)

		_, err = repo.FindByID(ctx, clinicA, pet.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999888005)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPetRepository_FindOwnersByPetBirthday(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	speciesID := makeSyncSpeciesID(t, db)

	matchBirthday := time.Date(1990, time.March, 15, 0, 0, 0, 0, time.UTC)
	otherBirthday := time.Date(1990, time.April, 1, 0, 0, 0, 0, time.UTC)
	deceasedAt := time.Now().Add(-24 * time.Hour)

	ownerMatch := makeTestOwner(t, db, clinicA, "誕生日一致飼主")
	makePetDetailed(t, db, clinicA, ownerMatch.ID, speciesID, "誕生日一致ペット", &matchBirthday, nil)

	// 同一飼主が複数ペット・同一誕生日を持っても owner_id は重複排除される
	makePetDetailed(t, db, clinicA, ownerMatch.ID, speciesID, "誕生日一致ペット2", &matchBirthday, nil)

	ownerOther := makeTestOwner(t, db, clinicA, "誕生日不一致飼主")
	makePetDetailed(t, db, clinicA, ownerOther.ID, speciesID, "誕生日不一致ペット", &otherBirthday, nil)

	ownerDeceasedOnly := makeTestOwner(t, db, clinicA, "死亡ペットのみ飼主")
	makePetDetailed(t, db, clinicA, ownerDeceasedOnly.ID, speciesID, "誕生日一致だが死亡", &matchBirthday, &deceasedAt)

	ownerOtherClinic := makeTestOwner(t, db, clinicB, "別クリニック誕生日一致飼主")
	makePetDetailed(t, db, clinicB, ownerOtherClinic.ID, speciesID, "別クリニックの誕生日一致ペット", &matchBirthday, nil)

	got, err := repo.FindOwnersByPetBirthday(ctx, clinicA, int(time.March), 15)
	require.NoError(t, err)
	assert.Contains(t, got, ownerMatch.ID)
	assert.Len(t, got, 1, "重複排除・死亡除外・別クリニック除外・誕生日不一致除外")
}

// makeSyncSpeciesID は AnimalSpecies を1件作成しIDを返す簡易ヘルパー。
func makeSyncSpeciesID(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	sp := &model.AnimalSpecies{Name: "汎用種別"}
	require.NoError(t, db.WithContext(context.Background()).Create(sp).Error)
	return sp.ID
}
