package pet

// pet_repository_test.go — PetRepository の統合テスト。
//
// FindByID / Delete の clinic_id 隔離は owner_pet_clinic_isolation_test.go、
// Insurance Preload の #86 隔離は cross_clinic_preload_isolation_test.go、
// Update の clinic_id 隔離は pet_write_medimage_clinic_isolation_test.go で別途カバー済みのため、
// 本ファイルは重複しない観点（検索・ページング・生存フィルタ・カウント系・誕生日検索など）に絞る。

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupPetRepositoryTestDB は pet_repository のテスト用に DB を整備する。
// owners CASCADE (testdb.SetupTestDB) が pets/pet_chronic_conditions を連鎖クリアするが、
// animal_species/insurances は個別に初期化する。
func setupPetRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Insurance{}))
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
	repo := NewRepository(db)
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
	for i := range pets {
		if pets[i].ID == id {
			return i
		}
	}
	return -1
}

// TestPetRepository_FindAll_Search は BUG-001 の検索契約回帰。
// 空白非依存の飼主フルネーム、飼主No（owners.id の文字列一致）、ペット番号、
// 空白のみ/未知番号の fail-closed、既存 name/kana/phone、count 一貫、clinic 隔離を固定する。
// フィクスチャは意図的に合成ラベルのみを使う（実在人物・事故由来識別子を置かない）。
func TestPetRepository_FindAll_Search(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	// Unmistakably synthetic fixtures (local constants reduce accidental person-like literals).
	const (
		ownerFullNameHalf = "検索対象姓 検索対象名"
		ownerFullNameFWQ  = "検索対象姓　検索対象名"
		ownerFullNameRep  = "検索対象姓  検索対象名"
		ownerFullNamePad  = "  検索対象姓 検索対象名  "
		ownerFullNameNone = "検索対象姓検索対象名"
		ownerNameKana     = "けんさくたいしょうせい けんさくたいしょうめい"
		ownerKanaKataQ    = "ケンサクタイショウセイ"
		ownerPhone        = "000-0000-0000"
		surnameOwnerName  = "部分一致姓 部分一致名"
		surnameQuery      = "部分一致姓"
		storedFWOwnerName = "全角保存姓　全角保存名"
		storedFWQueryHalf = "全角保存姓 全角保存名"
		storedFWQueryNone = "全角保存姓全角保存名"
		petNumberShared   = "PN-BUG001-SHARED"
		petNumberUnknown  = "PN-BUG001-UNKNOWN"
		petNameA          = "検索ペットA"
		petNameB          = "検索ペットB"
		petNameSurname    = "検索ペット姓"
		petNameFW         = "全角空白ペット"
	)

	ownerA := makeTestOwner(t, db, clinicA, ownerFullNameHalf)
	require.NoError(t, db.WithContext(ctx).Model(&model.Owner{}).Where("id = ?", ownerA.ID).Updates(map[string]any{
		"name_kana": ownerNameKana,
		"phone":     ownerPhone,
	}).Error)
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, petNameA)
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).Where("id = ?", petA.ID).Update("pet_number", petNumberShared).Error)

	// 別 clinic に同名・同 pet_number。owner ID は別採番（グローバル PK の重複は不可）。
	ownerB := makeTestOwner(t, db, clinicB, ownerFullNameHalf)
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, petNameB)
	require.NoError(t, db.WithContext(ctx).Model(&model.Pet{}).Where("id = ?", petB.ID).Update("pet_number", petNumberShared).Error)

	ownerSurname := makeTestOwner(t, db, clinicA, surnameOwnerName)
	petSurname := makeSpeciesAndPet(t, db, clinicA, ownerSurname.ID, petNameSurname)

	assertHitA := func(t *testing.T, search string) {
		t.Helper()
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: search}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "search=%q count", search)
		require.Len(t, pets, 1, "search=%q rows", search)
		assert.Equal(t, petA.ID, pets[0].ID, "search=%q pet id", search)
		assert.Equal(t, clinicA, pets[0].ClinicID)
	}

	assertZero := func(t *testing.T, clinicIDs []uint64, search string) {
		t.Helper()
		pets, total, err := repo.FindAll(ctx, clinicIDs, PetListFilters{Search: search}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total, "search=%q clinic=%v count", search, clinicIDs)
		assert.Empty(t, pets, "search=%q clinic=%v rows", search, clinicIDs)
	}

	t.Run("owner full name half-width space", func(t *testing.T) {
		assertHitA(t, ownerFullNameHalf)
	})
	t.Run("owner full name full-width space", func(t *testing.T) {
		assertHitA(t, ownerFullNameFWQ)
	})
	t.Run("owner full name repeated spaces", func(t *testing.T) {
		assertHitA(t, ownerFullNameRep)
	})
	t.Run("owner full name leading trailing whitespace", func(t *testing.T) {
		assertHitA(t, ownerFullNamePad)
	})
	t.Run("owner full name without spaces", func(t *testing.T) {
		assertHitA(t, ownerFullNameNone)
	})
	t.Run("owner name stored with full-width space matches half-width query", func(t *testing.T) {
		ownerFW := makeTestOwner(t, db, clinicA, storedFWOwnerName)
		petFW := makeSpeciesAndPet(t, db, clinicA, ownerFW.ID, petNameFW)
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: storedFWQueryHalf}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petFW.ID, pets[0].ID)
		pets2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: storedFWQueryNone}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total2)
		require.Len(t, pets2, 1)
		assert.Equal(t, petFW.ID, pets2[0].ID)
	})
	t.Run("owner number is owners.id text", func(t *testing.T) {
		assertHitA(t, fmt.Sprintf("%d", ownerA.ID))
	})
	t.Run("pet number exact", func(t *testing.T) {
		assertHitA(t, petNumberShared)
	})
	t.Run("pet number partial ILIKE", func(t *testing.T) {
		// AC: pets.pet_number is substring match (string, not numeric parse).
		assertHitA(t, "BUG001-SHARED")
		assertHitA(t, "PN-BUG001")
	})
	t.Run("whitespace only returns zero without listing clinic", func(t *testing.T) {
		assertZero(t, []uint64{clinicA}, "   ")
		assertZero(t, []uint64{clinicA}, "　　")
		assertZero(t, []uint64{clinicA}, " \t\n ")
	})
	t.Run("unknown number returns zero", func(t *testing.T) {
		assertZero(t, []uint64{clinicA}, "999999999999")
		assertZero(t, []uint64{clinicA}, petNumberUnknown)
	})
	t.Run("existing surname partial still matches", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: surnameQuery}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petSurname.ID, pets[0].ID)
	})
	t.Run("existing phone search still matches", func(t *testing.T) {
		assertHitA(t, ownerPhone)
	})
	t.Run("existing pet name search still matches", func(t *testing.T) {
		assertHitA(t, petNameA)
	})
	t.Run("kana-normalized owner name still matches", func(t *testing.T) {
		// stored name_kana is hiragana; katakana query hits via NormalizeKana path
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: ownerKanaKataQ}, 1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		ids := make([]uint64, len(pets))
		for i, p := range pets {
			ids[i] = p.ID
		}
		assert.Contains(t, ids, petA.ID)
	})
	t.Run("count and pagination stay consistent", func(t *testing.T) {
		_, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: ownerFullNameHalf}, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		page1, total1, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: ownerFullNameHalf}, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, total, total1)
		require.Len(t, page1, 1)
		assert.Equal(t, petA.ID, page1[0].ID)
		page2, total2, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: ownerFullNameHalf}, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, total, total2)
		assert.Empty(t, page2)
	})
	t.Run("clinic isolation same name and pet number", func(t *testing.T) {
		pets, total, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: ownerFullNameHalf}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, pets, 1)
		assert.Equal(t, petA.ID, pets[0].ID)
		assert.NotEqual(t, petB.ID, pets[0].ID)

		petsPN, totalPN, err := repo.FindAll(ctx, []uint64{clinicA}, PetListFilters{Search: petNumberShared}, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), totalPN)
		require.Len(t, petsPN, 1)
		assert.Equal(t, petA.ID, petsPN[0].ID)
	})
	t.Run("clinic isolation foreign owner number", func(t *testing.T) {
		// clinic B の owner ID を clinic A スコープで検索しても 0 件（存在リークなし）
		assertZero(t, []uint64{clinicA}, fmt.Sprintf("%d", ownerB.ID))
	})
}

func TestPetRepository_FindByIDForClinics(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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
	repo := NewRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "更新テスト飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "更新前ペット名")

	t.Run("成功", func(t *testing.T) {
		name := "更新後ペット名"
		err := repo.Update(ctx, clinicA, pet.ID, UpdatePetInput{Name: &name})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, pet.ID)
		require.NoError(t, err)
		assert.Equal(t, "更新後ペット名", got.Name)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		name := "x"
		err := repo.Update(ctx, clinicA, 999888004, UpdatePetInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPetRepository_Delete(t *testing.T) {
	db := setupPetRepositoryTestDB(t)
	repo := NewRepository(db)
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
	repo := NewRepository(db)
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

	ownerMismatched := makeTestOwner(t, db, clinicB, "不整合誕生日飼主")
	makePetDetailed(t, db, clinicA, ownerMismatched.ID, speciesID, "医院Aから医院B飼主を誤参照", &matchBirthday, nil)

	got, err := repo.FindOwnersByPetBirthday(ctx, clinicA, int(time.March), 15)
	require.NoError(t, err)
	assert.Contains(t, got, ownerMatch.ID)
	assert.NotContains(t, got, ownerMismatched.ID)
	assert.Len(t, got, 1, "重複排除・死亡除外・別クリニック除外・誕生日不一致除外")
}

// makeSyncSpeciesID は AnimalSpecies を1件作成しIDを返す簡易ヘルパー。
func makeSyncSpeciesID(t *testing.T, db *gorm.DB) uint64 {
	t.Helper()
	sp := &model.AnimalSpecies{Name: "汎用種別"}
	require.NoError(t, db.WithContext(context.Background()).Create(sp).Error)
	return sp.ID
}

func TestUpdateAndFind_RejectsMissingOwnerOrInsuranceInWriteTx(t *testing.T) {
	// POC-03: owner_id / insurance_id must be revalidated under FOR SHARE in the write tx.
	db := setupPetRepositoryTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(91)
	owner := makeTestOwner(t, db, clinicID, "fk revalidate owner")
	petModel := makeSpeciesAndPet(t, db, clinicID, owner.ID, "fk revalidate pet")
	repo := NewRepository(db)

	t.Run("missing owner", func(t *testing.T) {
		_, err := repo.UpdateAndFind(ctx, clinicID, petModel.ID, PetUpdate{
			fields: map[string]any{colPetOwnerID: uint64(9_999_991)},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "missing owner must be invalid input: %v", err)
		var loaded model.Pet
		require.NoError(t, db.First(&loaded, petModel.ID).Error)
		assert.Equal(t, owner.ID, loaded.OwnerID)
	})

	t.Run("missing insurance", func(t *testing.T) {
		missing := uint64(9_999_992)
		_, err := repo.UpdateAndFind(ctx, clinicID, petModel.ID, PetUpdate{
			fields: map[string]any{colPetInsuranceID: &missing},
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "missing insurance must be invalid input: %v", err)
	})

	t.Run("clear insurance still allowed", func(t *testing.T) {
		var nilID *uint64
		updated, err := repo.UpdateAndFind(ctx, clinicID, petModel.ID, PetUpdate{
			fields: map[string]any{colPetInsuranceID: nilID},
		})
		require.NoError(t, err)
		assert.Nil(t, updated.InsuranceID)
	})
}
