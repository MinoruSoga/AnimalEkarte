package pet

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

func setupPetOwnerRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupPetRepositoryTestDB(t)
	// テストDBのスキーマは SQL migration ではなくモデル単位の AutoMigrate で構築される。
	// PetOwner を登録しないと pet_owners が存在せず、以降の TRUNCATE で 42P01 になる。
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.PetOwner{}))
	require.NoError(t, db.Exec("TRUNCATE TABLE pet_owners RESTART IDENTITY CASCADE").Error)
	return db
}

func makePetOwnerLink(
	t *testing.T,
	db *gorm.DB,
	clinicID, petID, ownerID uint64,
	relationship string,
) model.PetOwner {
	t.Helper()
	link := model.PetOwner{
		ClinicID:     clinicID,
		PetID:        petID,
		OwnerID:      ownerID,
		Relationship: relationship,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(&link).Error)
	return link
}

func TestPetOwnerRepository_FindByPetID_ClinicIsolation(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "副飼主取得A")
	ownerB := makeTestOwner(t, db, clinicB, "副飼主取得B")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "副飼主ペットA")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "副飼主ペットB")
	linkA := makePetOwnerLink(t, db, clinicA, petA.ID, ownerA.ID, "家族A")
	makePetOwnerLink(t, db, clinicB, petB.ID, ownerB.ID, "家族B")

	got, err := repo.FindByPetID(ctx, clinicA, petA.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, linkA.ID, got[0].ID)

	crossClinic, err := repo.FindByPetID(ctx, clinicA, petB.ID)
	require.NoError(t, err)
	assert.Empty(t, crossClinic)
}

func TestPetOwnerRepository_ReplaceForPet_ClinicIsolation(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "副飼主置換A")
	ownerA2 := makeTestOwner(t, db, clinicA, "副飼主置換A2")
	ownerB := makeTestOwner(t, db, clinicB, "副飼主置換B")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "副飼主置換ペットA")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "副飼主置換ペットB")
	originalA := makePetOwnerLink(t, db, clinicA, petA.ID, ownerA.ID, "元リンクA")
	originalB := makePetOwnerLink(t, db, clinicB, petB.ID, ownerB.ID, "元リンクB")

	err := repo.ReplaceForPet(ctx, clinicA, petB.ID, nil, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var foreignLinks []model.PetOwner
	require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicB, petB.ID).Find(&foreignLinks).Error)
	require.Len(t, foreignLinks, 1)
	assert.Equal(t, originalB.ID, foreignLinks[0].ID)

	err = repo.ReplaceForPet(ctx, clinicA, petA.ID, []model.PetOwner{{
		OwnerID:      ownerB.ID,
		Relationship: "越境副飼主",
	}}, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var ownLinks []model.PetOwner
	require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicA, petA.ID).Find(&ownLinks).Error)
	require.Len(t, ownLinks, 1)
	assert.Equal(t, originalA.ID, ownLinks[0].ID)

	require.NoError(t, repo.ReplaceForPet(ctx, clinicA, petA.ID, []model.PetOwner{{
		ID:           originalB.ID,
		ClinicID:     clinicB,
		PetID:        petB.ID,
		OwnerID:      ownerA2.ID,
		Relationship: "正規化対象",
	}}, nil))
	ownLinks = nil
	require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicA, petA.ID).Find(&ownLinks).Error)
	require.Len(t, ownLinks, 1)
	assert.Equal(t, clinicA, ownLinks[0].ClinicID)
	assert.Equal(t, petA.ID, ownLinks[0].PetID)
	assert.Equal(t, ownerA2.ID, ownLinks[0].OwnerID)
}

func TestPetOwnerRepository_ReplaceForPet_RejectsSoftDeletedOwner(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	primaryOwner := makeTestOwner(t, db, clinicID, "削除済み副飼主拒否・主飼主")
	currentSecondaryOwner := makeTestOwner(t, db, clinicID, "削除済み副飼主拒否・現副飼主")
	deletedSecondaryOwner := makeTestOwner(t, db, clinicID, "削除済み副飼主拒否・削除済み")
	pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "削除済み副飼主拒否ペット")
	original := makePetOwnerLink(t, db, clinicID, pet.ID, currentSecondaryOwner.ID, "現副飼主")
	originalVersion := pet.Version

	require.NoError(t, db.Delete(deletedSecondaryOwner).Error)

	err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
		OwnerID:      deletedSecondaryOwner.ID,
		Relationship: "削除済み副飼主",
	}}, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)

	var links []model.PetOwner
	require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicID, pet.ID).Find(&links).Error)
	require.Len(t, links, 1)
	assert.Equal(t, original.ID, links[0].ID)
	assert.Equal(t, currentSecondaryOwner.ID, links[0].OwnerID)

	var rejectedLinkCount int64
	require.NoError(t, db.Model(&model.PetOwner{}).
		Where("clinic_id = ? AND pet_id = ? AND owner_id = ?", clinicID, pet.ID, deletedSecondaryOwner.ID).
		Count(&rejectedLinkCount).Error)
	assert.Zero(t, rejectedLinkCount)

	var reloadedPet model.Pet
	require.NoError(t, db.Unscoped().First(&reloadedPet, pet.ID).Error)
	assert.Equal(t, originalVersion, reloadedPet.Version)
}

func TestPetOwnerRepository_CountByOwnerID_ClinicIsolation(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "副飼主件数A")
	ownerB := makeTestOwner(t, db, clinicB, "副飼主件数B")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "副飼主件数ペットA")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "副飼主件数ペットB")
	makePetOwnerLink(t, db, clinicA, petA.ID, ownerA.ID, "件数A")
	makePetOwnerLink(t, db, clinicB, petB.ID, ownerB.ID, "件数B")

	count, err := repo.CountByOwnerID(ctx, clinicA, ownerA.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	crossClinicCount, err := repo.CountByOwnerID(ctx, clinicA, ownerB.ID)
	require.NoError(t, err)
	assert.Zero(t, crossClinicCount)
}

func TestPetOwnerRepository_FindSharedPetsByOwnerID(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	sharedOwnerA := makeTestOwner(t, db, clinicA, "共同飼育ペット取得A")
	primaryOwnerA := makeTestOwner(t, db, clinicA, "共同飼育ペット主飼主A")
	ownerB := makeTestOwner(t, db, clinicB, "共同飼育ペット取得B")

	primaryPet := makeSpeciesAndPet(t, db, clinicA, sharedOwnerA.ID, "主飼主ペット")
	sharedPet := makeSpeciesAndPet(t, db, clinicA, primaryOwnerA.ID, "共同飼育ペット")
	deceasedPet := makeSpeciesAndPet(t, db, clinicA, primaryOwnerA.ID, "死亡共同飼育ペット")
	deletedPet := makeSpeciesAndPet(t, db, clinicA, primaryOwnerA.ID, "削除済み共同飼育ペット")
	foreignPet := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "他院共同飼育ペット")

	birthDate := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.Local)
	weight := 4.2
	require.NoError(t, db.Model(&model.Pet{}).
		Where("clinic_id = ? AND id = ?", clinicA, sharedPet.ID).
		Updates(map[string]any{
			"pet_number":  "P-0001",
			"gender":      model.PetGenderFemale,
			"birth_date":  birthDate,
			"color":       "茶",
			"weight":      weight,
			"environment": "室内",
			"remarks":     "共同飼育",
		}).Error)
	deceasedAt := time.Now()
	require.NoError(t, db.Model(&model.Pet{}).
		Where("clinic_id = ? AND id = ?", clinicA, deceasedPet.ID).
		Updates(map[string]any{
			"status":      model.PetStatusDeceased,
			"deceased_at": deceasedAt,
		}).Error)
	require.NoError(t, db.Delete(&deletedPet).Error)

	makePetOwnerLink(t, db, clinicA, primaryPet.ID, sharedOwnerA.ID, "主飼主混入")
	makePetOwnerLink(t, db, clinicA, sharedPet.ID, sharedOwnerA.ID, "家族")
	makePetOwnerLink(t, db, clinicA, deceasedPet.ID, sharedOwnerA.ID, "親族")
	makePetOwnerLink(t, db, clinicA, deletedPet.ID, sharedOwnerA.ID, "削除済み")
	makePetOwnerLink(t, db, clinicB, foreignPet.ID, ownerB.ID, "他院")

	got, err := repo.FindSharedPetsByOwnerID(ctx, clinicA, sharedOwnerA.ID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, sharedPet.ID, got[0].ID)
	assert.Equal(t, "P-0001", got[0].PetNumber)
	assert.Equal(t, "共同飼育ペット", got[0].Name)
	assert.Equal(t, "家族", got[0].Relationship)
	assert.Equal(t, model.PetStatusAlive, got[0].Status)
	assert.NotEmpty(t, got[0].AnimalSpeciesName)
	assert.Equal(t, model.PetGenderFemale, got[0].Gender)
	require.NotNil(t, got[0].BirthDate)
	assert.Equal(t, "2020-01-02", got[0].BirthDate.Format(time.DateOnly))
	assert.Equal(t, "茶", got[0].Color)
	require.NotNil(t, got[0].Weight)
	assert.InDelta(t, 4.2, *got[0].Weight, 0.001)
	assert.Equal(t, "室内", got[0].Environment)
	assert.Equal(t, "共同飼育", got[0].Remarks)
	assert.Equal(t, deceasedPet.ID, got[1].ID)
	assert.Equal(t, "親族", got[1].Relationship)
	assert.Equal(t, model.PetStatusDeceased, got[1].Status)
	assert.Nil(t, got[1].BirthDate)
	assert.Nil(t, got[1].Weight)
	crossClinic, err := repo.FindSharedPetsByOwnerID(ctx, clinicA, ownerB.ID)
	require.NoError(t, err)
	assert.Empty(t, crossClinic)
}

func TestPetOwnerRepository_FindSharedPetsByOwnerID_AmbientTransaction(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	const clinicID = uint64(1)

	sharedOwner := makeTestOwner(t, db, clinicID, "ambient共同飼育飼主")
	primaryOwner := makeTestOwner(t, db, clinicID, "ambient共同飼育主飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "ambient共同飼育ペット")
	tx, txCtx := beginAmbientPetOwnerTransaction(t, db)
	link := makePetOwnerLink(t, tx, clinicID, pet.ID, sharedOwner.ID, "未commit共同飼育")

	got, err := repo.FindSharedPetsByOwnerID(txCtx, clinicID, sharedOwner.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, pet.ID, got[0].ID)
	assert.Equal(t, "未commit共同飼育", got[0].Relationship)
	require.NoError(t, tx.Rollback().Error)

	var count int64
	require.NoError(t, db.Model(&model.PetOwner{}).Where("id = ?", link.ID).Count(&count).Error)
	assert.Zero(t, count)
}

func TestPetOwnerRepository_ReplaceForPet(t *testing.T) {
	t.Run("replaces two links with one normalized link", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		primaryOwner := makeTestOwner(t, db, clinicID, "全置換主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "全置換旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "全置換新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "全置換ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, primaryOwner.ID, "主")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "旧")

		require.NoError(t, repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{{
			OwnerID:      newOwner.ID,
			Relationship: "新",
		}}, nil))

		var links []model.PetOwner
		require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicID, pet.ID).Find(&links).Error)
		require.Len(t, links, 1)
		assert.Equal(t, newOwner.ID, links[0].OwnerID)
		assert.Equal(t, "新", links[0].Relationship)
	})

	t.Run("empty slice removes every link", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		owner := makeTestOwner(t, db, clinicID, "全解除飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "全解除ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, owner.ID, "解除対象")

		require.NoError(t, repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{}, nil))

		var count int64
		require.NoError(t, db.Model(&model.PetOwner{}).
			Where("clinic_id = ? AND pet_id = ?", clinicID, pet.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("insert failure rolls back the preceding delete", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		repo := NewPetOwnerRepository(db)
		ctx := context.Background()
		const clinicID = uint64(1)

		owner := makeTestOwner(t, db, clinicID, "原子性飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "原子性ペット")
		original := makePetOwnerLink(t, db, clinicID, pet.ID, owner.ID, "元リンク")

		err := repo.ReplaceForPet(ctx, clinicID, pet.ID, []model.PetOwner{
			{OwnerID: owner.ID, Relationship: "重複1"},
			{OwnerID: owner.ID, Relationship: "重複2"},
		}, nil)
		require.Error(t, err)

		var links []model.PetOwner
		require.NoError(t, db.Where("clinic_id = ? AND pet_id = ?", clinicID, pet.ID).Find(&links).Error)
		require.Len(t, links, 1)
		assert.Equal(t, original.ID, links[0].ID)
		assert.Equal(t, "元リンク", links[0].Relationship)
	})
}
