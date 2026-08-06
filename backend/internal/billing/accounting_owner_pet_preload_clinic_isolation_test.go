package billing

// accounting_owner_pet_preload_clinic_isolation_test.go — AUD-002
// 汚染された Owner/Pet FK を持つ billing から、別 clinic の個人情報が Preload されないことを検証する。
// Preload から clinic_id 述語を外すと本テストは RED になる。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupAccountingOwnerPetPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupAccountingIsolationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}))
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func TestAccountingRepository_FindByID_Update_DoesNotPreloadForeignOwnerPet(t *testing.T) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院Bの飼主")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "医院Bのペット")
	ownerBID, petBID := ownerB.ID, petB.ID

	contaminated := makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicA,
		OwnerID:       &ownerBID,
		PetID:         &petBID,
		TotalAmount:   1500,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
	})

	t.Run("FindByID does not preload foreign owner/pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded")
	})

	t.Run("Update refetch does not preload foreign owner/pet", func(t *testing.T) {
		memo := "aud-002-preload"
		got, err := repo.Update(ctx, clinicA, contaminated.ID, map[string]any{"memo": memo})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Equal(t, petBID, *got.PetID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded on Update refetch")
		assert.Nil(t, got.Pet, "foreign Pet must not be preloaded on Update refetch")
	})
}

func TestAccountingRepository_FindAll_RejectsCorruptCrossClinicPetRelation(
	t *testing.T,
) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "cross-clinic billing owner")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "cross-clinic billing pet")
	petBID := petB.ID
	makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicA,
		OwnerID:       &ownerA.ID,
		PetID:         &petBID,
		TotalAmount:   1500,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
	})

	items, total, err := repo.FindAll(
		ctx,
		clinicA,
		nil,
		nil,
		nil,
		nil,
		nil,
		"", 1,
		50,
	)

	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items, "clinic A billing must not resolve a clinic B pet")
}

func TestAccountingRepository_FindAllForClinics_RejectsCorruptCrossClinicPetRelation(
	t *testing.T,
) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "cross-clinic multi billing owner")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerA.ID, "cross-clinic multi billing pet")
	petBID := petB.ID
	makeBillingWith(t, db, billingFixtureOpts{
		ClinicID:      clinicA,
		OwnerID:       &ownerA.ID,
		PetID:         &petBID,
		TotalAmount:   1600,
		Status:        model.BillingStatusWaiting,
		ScheduledDate: time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
	})

	items, total, err := repo.FindAllForClinics(
		ctx,
		[]uint64{clinicA},
		nil,
		nil,
		nil,
		nil,
		nil,
		"", 1,
		50,
	)

	require.NoError(t, err)
	assert.Zero(t, total)
	assert.Empty(t, items, "authorized clinics must not resolve a pet outside their scope")
}

func TestAccountingRepository_FindAll_PreservesBillingWithoutPet(t *testing.T) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	billing := makeBillingRet(t, db, clinicA)

	items, total, err := repo.FindAll(
		ctx,
		clinicA,
		nil,
		nil,
		nil,
		nil,
		nil,
		"", 1,
		50,
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	assert.Equal(t, billing.ID, items[0].ID)
	assert.Nil(t, items[0].PetID)
}

func TestAccountingRepository_CoreReadsEnforceExactOwnerPetCorrelation(
	t *testing.T,
) {
	db := setupAccountingOwnerPetPreloadDB(t)
	repo := NewAccountingRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := testdb.MakeTestOwner(t, db, clinicA, "医院Aの飼主")
	otherOwnerA := testdb.MakeTestOwner(t, db, clinicA, "医院Aの別飼主")
	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院Bの飼主")
	wrongOwnerPet := makeSpeciesAndPet(
		t,
		db,
		clinicA,
		otherOwnerA.ID,
		"同一医院の別飼主ペット",
	)
	foreignPet := makeSpeciesAndPet(
		t,
		db,
		clinicB,
		ownerB.ID,
		"別医院のペット",
	)
	localPetWithForeignOwner := makeSpeciesAndPet(
		t,
		db,
		clinicA,
		ownerB.ID,
		"別医院飼主に紐づく医院Aペット",
	)

	sameClinicMismatch := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerA.ID,
		&wrongOwnerPet.ID,
		model.BillingStatusWaiting,
		1000,
		nil,
	)
	crossClinicMismatch := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerB.ID,
		&foreignPet.ID,
		model.BillingStatusWaiting,
		2000,
		nil,
	)
	foreignOwnerLocalPetMismatch := makeTenantRelationBilling(
		t,
		db,
		clinicA,
		&ownerB.ID,
		&localPetWithForeignOwner.ID,
		model.BillingStatusWaiting,
		3000,
		nil,
	)

	t.Run("detail keeps billing but removes same-clinic wrong-owner pet", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, sameClinicMismatch.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Owner)
		assert.Equal(t, ownerA.ID, got.Owner.ID)
		assert.Nil(t, got.Pet)
	})

	t.Run("list and update refetch remove same-clinic wrong-owner pet", func(t *testing.T) {
		items, _, err := repo.FindAll(
			ctx,
			clinicA,
			nil,
			nil,
			nil,
			nil,
			nil,
			"", 1,
			100,
		)
		require.NoError(t, err)
		var found *model.Billing
		for i := range items {
			if items[i].ID == sameClinicMismatch.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		require.NotNil(t, found.Owner)
		assert.Nil(t, found.Pet)

		updated, err := repo.Update(
			ctx,
			clinicA,
			sameClinicMismatch.ID,
			map[string]any{"memo": "exact owner pet correlation"},
		)
		require.NoError(t, err)
		require.NotNil(t, updated.Owner)
		assert.Nil(t, updated.Pet)
	})

	t.Run("multi-clinic detail sanitizes while list excludes cross-clinic pet relation", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			crossClinicMismatch.ID,
		)
		require.NoError(t, err)
		assert.Nil(t, got.Owner)
		assert.Nil(t, got.Pet)

		items, _, err := repo.FindAllForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			nil,
			nil,
			nil,
			nil,
			nil,
			"", 1,
			100,
		)
		require.NoError(t, err)
		var found *model.Billing
		for i := range items {
			if items[i].ID == crossClinicMismatch.ID {
				found = &items[i]
				break
			}
		}
		assert.Nil(t, found, "list must exclude a billing whose pet belongs to another clinic")
	})

	t.Run("pet is removed when its billing owner relation is invalid", func(t *testing.T) {
		got, err := repo.FindByIDForClinics(
			ctx,
			[]uint64{clinicA, clinicB},
			foreignOwnerLocalPetMismatch.ID,
		)
		require.NoError(t, err)
		assert.Nil(t, got.Owner)
		assert.Nil(t, got.Pet)
	})
}
