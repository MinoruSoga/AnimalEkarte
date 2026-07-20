package medicalrecord

// medical_record_owner_pet_preload_clinic_isolation_test.go — AUD-008
// 汚染された Owner/Pet FK を持つカルテから、別 clinic の個人情報が Preload されないことを検証する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

func setupMedicalRecordOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupMedicalRecordListTestDB(t)
	db.Exec("TRUNCATE TABLE medical_records, pets, animal_species, owners CASCADE")
	return db
}

func TestMedicalRecordRepository_FindByID_FindAll_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := makeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")

	ownerBID, petBID := ownerB.ID, petB.ID
	contaminated := &model.MedicalRecord{
		ClinicID: clinicA,
		Date:     time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
		OwnerID:  &ownerBID,
		PetID:    &petBID,
		Status:   model.MedicalRecordStatusDraft,
		RecordNo: "AUD008-CONTAMINATED",
	}
	require.NoError(t, db.WithContext(ctx).Create(contaminated).Error)

	t.Run("FindByID does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})

	t.Run("FindAll does not preload foreign owner/pet", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, []uint64{clinicA}, MedicalRecordListFilters{}, 1, 50)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.MedicalRecord
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Nil(t, found.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, found.Pet, "foreign Pet must not be preloaded")
	})
}

func TestMedicalRecordRepository_Create_ParticipatesInAmbientTransaction(t *testing.T) {
	db := setupMedicalRecordOwnerPetPreloadDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "自院飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "自院ペット")
	ownerID, petID := owner.ID, pet.ID

	tx := testTransactor{db: db}
	sentinel := errors.New("force medical record create rollback")
	err := tx.WithTx(ctx, func(txCtx context.Context) error {
		rec := &model.MedicalRecord{
			ClinicID: clinicA,
			Date:     time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
			OwnerID:  &ownerID,
			PetID:    &petID,
			Status:   model.MedicalRecordStatusDraft,
			RecordNo: "AUD008-TX-CREATE",
		}
		if e := repo.Create(txCtx, rec); e != nil {
			return e
		}
		return sentinel
	})
	require.Error(t, err)

	var count int64
	require.NoError(t, db.WithContext(ctx).Model(&model.MedicalRecord{}).
		Where("record_no = ?", "AUD008-TX-CREATE").Count(&count).Error)
	assert.Zero(t, count, "medical record create must roll back with the ambient transaction")
}
