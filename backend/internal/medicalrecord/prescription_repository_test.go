package medicalrecord

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - FindByMedicalRecordID / FindByID / Update / Delete は clinic_id でテナント隔離される。
//   - FindByMedicalRecordID は prescribed_at DESC で返す。
//   - FindActiveByOwner は clinic_id + pet の現在飼主スコープでソフトデリート済みを除外する
//     （pet_id が無い処方は除外し、deleted_at IS NULL を明示条件に持つ、LSTEP-BE-009）。
//   - Update / Delete は対象なしで NotFound を返す。
//   - Delete はソフトデリートであり、以後 FindByID / FindActiveByOwner から除外される。
//
// makeTestOwner は testdb.MakeTestOwner に直接委譲する。withTx は repository.Transactor.WithTx
// を import cycle なしで再現する repohelpers 直結ヘルパー（BE8-4 方針）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func makeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
	t.Helper()
	return testdb.MakeTestOwner(t, db, clinicID, name)
}

// withTx mirrors repository.Transactor.WithTx (repohelpers-based ambient tx) without
// importing the flat repository package, which would create an import cycle
// (repository imports this subpackage via its facade).
func withTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	})
}

// setupPrescriptionTestDB は prescriptions と、FK 先の pets/animal_species を用意する。
// owners / medical_records は setupTestDB がすでに用意する。
func setupPrescriptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AnimalSpecies{}, &model.Pet{}, &model.Prescription{}))
	db.Exec("TRUNCATE TABLE prescriptions CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makePrescriptionMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{ClinicID: clinicID, RecordNo: recordNo, Date: time.Now(), Status: model.MedicalRecordStatusFinalized}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func makePrescription(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, medicalRecordID *uint64, prescribedAt time.Time) *model.Prescription {
	t.Helper()
	p := &model.Prescription{
		ClinicID:        clinicID,
		OwnerID:         ownerID,
		MedicalRecordID: medicalRecordID,
		PrescribedAt:    prescribedAt,
		DurationDays:    7,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(p).Error)
	return p
}

func TestPrescriptionRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	mrA := makePrescriptionMedicalRecord(t, db, clinicA, "PR-MR-A")
	mrB := makePrescriptionMedicalRecord(t, db, clinicB, "PR-MR-B")

	earlier := makePrescription(t, db, clinicA, ownerA.ID, &mrA.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	later := makePrescription(t, db, clinicA, ownerA.ID, &mrA.ID, time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC))
	makePrescription(t, db, clinicB, ownerA.ID, &mrB.ID, time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC))

	t.Run("returns prescriptions ordered by prescribed_at DESC", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, later.ID, got[0].ID)
		assert.Equal(t, earlier.ID, got[1].ID)
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicB, mrA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestPrescriptionRepository_FindByID(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	prescription := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, prescription.ID)
		require.NoError(t, err)
		assert.Equal(t, prescription.ID, got.ID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, prescription.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPrescriptionRepository_FindActiveByOwner(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	otherOwner := makeTestOwner(t, db, clinicA, "飼主B")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "処方ポチA")
	otherPet := makeSpeciesAndPet(t, db, clinicA, otherOwner.ID, "処方ポチB")

	active := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())
	toBeDeleted := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())
	other := makePrescription(t, db, clinicA, otherOwner.ID, nil, time.Now())
	require.NoError(t, db.Model(active).Update("pet_id", petA.ID).Error)
	require.NoError(t, db.Model(toBeDeleted).Update("pet_id", petA.ID).Error)
	require.NoError(t, db.Model(other).Update("pet_id", otherPet.ID).Error)
	// NOTE: clinicB の prescription は作らない。FindActiveByOwner は対象 clinic 内の pet と
	// その現在飼主を相関するため、下の clinicB 問い合わせは pet 経由でも空になることを確認する。

	require.NoError(t, repo.Delete(ctx, clinicA, toBeDeleted.ID))

	t.Run("returns only active (non-deleted) prescriptions for the owner within the clinic", func(t *testing.T) {
		got, err := repo.FindActiveByOwner(ctx, clinicA, ownerA.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, active.ID, got[0].ID)
	})

	t.Run("clinic isolation: different clinic scope returns empty", func(t *testing.T) {
		got, err := repo.FindActiveByOwner(ctx, clinicB, ownerA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("empty for owner with no prescriptions", func(t *testing.T) {
		strangerOwner := makeTestOwner(t, db, clinicA, "処方なし飼主")
		got, err := repo.FindActiveByOwner(ctx, clinicA, strangerOwner.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestPrescriptionRepository_FindActiveByOwner_CurrentOwnerAfterTransfer(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()
	const clinicID = uint64(70104)

	fixture := makeCurrentOwnerTransferFixture(
		t,
		db,
		clinicID,
		"MR-PRESCRIPTION-CURRENT-OWNER",
		time.Now(),
	)
	prescription := &model.Prescription{
		ClinicID:        clinicID,
		OwnerID:         fixture.PreviousOwner.ID,
		PetID:           &fixture.Pet.ID,
		MedicalRecordID: &fixture.Record.ID,
		PrescribedAt:    time.Now(),
		DurationDays:    7,
	}
	require.NoError(t, db.WithContext(ctx).Create(prescription).Error)
	nullPet := &model.Prescription{
		ClinicID:     clinicID,
		OwnerID:      fixture.CurrentOwner.ID,
		PrescribedAt: time.Now(),
		DurationDays: 7,
	}
	require.NoError(t, db.WithContext(ctx).Create(nullPet).Error)

	got, err := repo.FindActiveByOwner(ctx, clinicID, fixture.CurrentOwner.ID)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, prescription.ID, got[0].ID)
	assert.Equal(t, fixture.PreviousOwner.ID, got[0].OwnerID, "returned owner_id remains the historical snapshot")
	require.NotNil(t, got[0].PetID)

	previous, err := repo.FindActiveByOwner(ctx, clinicID, fixture.PreviousOwner.ID)
	require.NoError(t, err)
	assert.Empty(t, previous)
}

func TestPrescriptionRepository_Create(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")

	p := &model.Prescription{ClinicID: clinicA, OwnerID: ownerA.ID, PrescribedAt: time.Now(), DurationDays: 3}
	require.NoError(t, repo.Create(ctx, p))
	assert.NotZero(t, p.ID)
}

func TestPrescriptionRepository_Update(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	p := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())

	t.Run("updates successfully", func(t *testing.T) {
		days := 14
		require.NoError(t, repo.Update(ctx, clinicA, p.ID, UpdatePrescriptionInput{DurationDays: &days}))
		got, err := repo.FindByID(ctx, clinicA, p.ID)
		require.NoError(t, err)
		assert.Equal(t, 14, got.DurationDays)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		days := 1
		err := repo.Update(ctx, clinicA, uint64(999999), UpdatePrescriptionInput{DurationDays: &days})
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation: wrong clinic returns NotFound", func(t *testing.T) {
		days := 99
		err := repo.Update(ctx, clinicB, p.ID, UpdatePrescriptionInput{DurationDays: &days})
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestPrescriptionRepository_Delete(t *testing.T) {
	db := setupPrescriptionTestDB(t)
	repo := NewPrescriptionRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	p := makePrescription(t, db, clinicA, ownerA.ID, nil, time.Now())

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, p.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deletes successfully", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, p.ID))

		var unscoped int64
		require.NoError(t, db.Unscoped().Model(&model.Prescription{}).Where("id = ?", p.ID).Count(&unscoped).Error)
		assert.Equal(t, int64(1), unscoped, "物理的には行が残る（ソフトデリート）")

		var scoped int64
		require.NoError(t, db.Model(&model.Prescription{}).Where("id = ?", p.ID).Count(&scoped).Error)
		assert.Equal(t, int64(0), scoped)
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, p.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
