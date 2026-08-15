package pet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func beginAmbientPetOwnerTransaction(t *testing.T, db *gorm.DB) (*gorm.DB, context.Context) {
	t.Helper()
	tx := db.Begin()
	require.NoError(t, tx.Error)
	return tx, persistence.WithTxValue(context.Background(), tx)
}

func TestPetOwnerRepository_FindByPetID_AmbientTransaction(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "ambient取得飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ambient取得ペット")
	tx, txCtx := beginAmbientPetOwnerTransaction(t, db)
	link := makePetOwnerLink(t, tx, clinicID, pet.ID, owner.ID, "未commit取得")

	got, err := repo.FindByPetID(txCtx, clinicID, pet.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, link.ID, got[0].ID)
	require.NoError(t, tx.Rollback().Error)

	var count int64
	require.NoError(t, db.Model(&model.PetOwner{}).Where("id = ?", link.ID).Count(&count).Error)
	assert.Zero(t, count)
}

func TestPetOwnerRepository_ReplaceForPet_AmbientTransaction(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	primaryOwner := makeTestOwner(t, db, clinicID, "ambient置換主飼主")
	oldOwner := makeTestOwner(t, db, clinicID, "ambient置換旧副飼主")
	newOwner := makeTestOwner(t, db, clinicID, "ambient置換新副飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "ambient置換ペット")
	oldPrimary := makePetOwnerLink(t, db, clinicID, pet.ID, primaryOwner.ID, "主")
	oldSecondary := makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "旧")

	tx, txCtx := beginAmbientPetOwnerTransaction(t, db)
	require.NoError(t, repo.ReplaceForPet(txCtx, clinicID, pet.ID, []model.PetOwner{{
		OwnerID:      newOwner.ID,
		Relationship: "新",
	}}, nil))
	require.NoError(t, tx.Rollback().Error)

	var links []model.PetOwner
	require.NoError(t, db.WithContext(ctx).
		Where("clinic_id = ? AND pet_id = ?", clinicID, pet.ID).
		Order("id ASC").
		Find(&links).Error)
	require.Len(t, links, 2)
	assert.Equal(t, oldPrimary.ID, links[0].ID)
	assert.Equal(t, oldSecondary.ID, links[1].ID)
	assert.Equal(t, primaryOwner.ID, links[0].OwnerID)
	assert.Equal(t, oldOwner.ID, links[1].OwnerID)
}

func TestPetOwnerRepository_CountByOwnerID_AmbientTransaction(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	repo := NewPetOwnerRepository(db)
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "ambient件数飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "ambient件数ペット")
	tx, txCtx := beginAmbientPetOwnerTransaction(t, db)
	link := makePetOwnerLink(t, tx, clinicID, pet.ID, owner.ID, "未commit件数")

	count, err := repo.CountByOwnerID(txCtx, clinicID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
	require.NoError(t, tx.Rollback().Error)

	var persisted int64
	require.NoError(t, db.Model(&model.PetOwner{}).Where("id = ?", link.ID).Count(&persisted).Error)
	assert.Zero(t, persisted)
}
