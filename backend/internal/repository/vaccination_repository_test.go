package repository

// vaccination_repository_test.go
// vaccination_repository.go の実 DB 結合テスト。
// makeTestOwner / makeSpeciesAndPet（db_setup_test.go / accounting_repository_unpaid_test.go）、
// makeVaccineMaster / makeVaccinationRecord（vaccination_master_preload_clinic_isolation_test.go）、
// makeDoctor（isolation_test_helpers_test.go）を再利用する。

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

// setupVaccinationRepoTestDB は vaccination_repository のテストに必要なテーブルを整備する。
func setupVaccinationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Owner{}, &model.Pet{}, &model.Vaccine{}, &model.Staff{}, &model.Vaccination{},
	))
	db.Exec("TRUNCATE TABLE vaccinations CASCADE")
	db.Exec("TRUNCATE TABLE vaccines CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	return db
}

func makeVaccinationRepoTestPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, name string) *model.Pet {
	t.Helper()
	return makeSpeciesAndPet(t, db, clinicID, ownerID, name)
}

func TestVaccinationRepository_Create_FindByID(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "ワクチン飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "ワクチンペット")
	vaccine := makeVaccineMaster(t, db, clinicA, "混合ワクチン")
	doctor := makeDoctor(t, db, clinicA, "接種医師")

	t.Run("作成した記録を関連Preload付きで取得できる", func(t *testing.T) {
		did := doctor.ID
		v := &model.Vaccination{
			ClinicID:  clinicA,
			PetID:     &pet.ID,
			VaccineID: vaccine.ID,
			Date:      time.Now(),
			DoctorID:  &did,
			Lot1:      "LOT-001",
		}
		require.NoError(t, repo.Create(ctx, v))
		require.NotZero(t, v.ID)

		got, err := repo.FindByID(ctx, clinicA, v.ID)
		require.NoError(t, err)
		require.NotNil(t, got.Vaccine)
		assert.Equal(t, vaccine.ID, got.Vaccine.ID)
		require.NotNil(t, got.Pet)
		assert.Equal(t, pet.ID, got.Pet.ID)
		require.NotNil(t, got.Pet.Owner)
		assert.Equal(t, owner.ID, got.Pet.Owner.ID)
		require.NotNil(t, got.Doctor)
		assert.Equal(t, doctor.ID, got.Doctor.ID)
	})

	t.Run("別クリニックからはNotFound", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		_, err := repo.FindByID(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestVaccinationRepository_FindAll(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA1 := makeTestOwner(t, db, clinicA, "ワクチン飼主A1")
	ownerA2 := makeTestOwner(t, db, clinicA, "ワクチン飼主A2")
	petA1 := makeVaccinationRepoTestPet(t, db, clinicA, ownerA1.ID, "ペットA1")
	petA2 := makeVaccinationRepoTestPet(t, db, clinicA, ownerA2.ID, "ペットA2")
	ownerB := makeTestOwner(t, db, clinicB, "ワクチン飼主B")
	petB := makeVaccinationRepoTestPet(t, db, clinicB, ownerB.ID, "ペットB")
	vaccineA := makeVaccineMaster(t, db, clinicA, "医院Aワクチン")
	vaccineB := makeVaccineMaster(t, db, clinicB, "医院Bワクチン")

	old := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	mid := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)

	makeVaccinationOnDate := func(clinicID, petID, vaccineID uint64, date time.Time) *model.Vaccination {
		v := &model.Vaccination{ClinicID: clinicID, PetID: &petID, VaccineID: vaccineID, Date: date}
		require.NoError(t, db.WithContext(ctx).Create(v).Error)
		return v
	}

	makeVaccinationOnDate(clinicA, petA1.ID, vaccineA.ID, old)
	makeVaccinationOnDate(clinicA, petA1.ID, vaccineA.ID, recent)
	makeVaccinationOnDate(clinicA, petA2.ID, vaccineA.ID, mid)
	makeVaccinationOnDate(clinicB, petB.ID, vaccineB.ID, mid)

	t.Run("クリニックで隔離され全件返る", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, got, 3)
		for _, v := range got {
			assert.Equal(t, clinicA, v.ClinicID)
		}
	})

	t.Run("petIDで絞り込める", func(t *testing.T) {
		pid := petA1.ID
		got, total, err := repo.FindAll(ctx, clinicA, &pid, nil, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		for _, v := range got {
			assert.Equal(t, petA1.ID, *v.PetID)
		}
	})

	t.Run("ownerIDでJOIN絞り込める", func(t *testing.T) {
		oid := ownerA2.ID
		got, total, err := repo.FindAll(ctx, clinicA, nil, &oid, nil, nil, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, petA2.ID, *got[0].PetID)
	})

	t.Run("startDate/endDateで期間絞り込める", func(t *testing.T) {
		start := "2026-02-01"
		end := "2026-04-01"
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, &start, &end, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total, "midのみが期間内")
		require.Len(t, got, 1)
	})

	t.Run("ページネーションはtotalを保ちつつ件数を制限する", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, got, 1)
	})
}

func TestVaccinationRepository_FindByOwner(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "複数ペット飼主")
	petAlive := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "生存ペット")
	vaccine := makeVaccineMaster(t, db, clinicA, "狂犬病ワクチン")

	aliveRec := makeVaccinationRecord(t, db, clinicA, petAlive.ID, vaccine.ID)

	// 削除済みペットのワクチン記録は FindByOwner の JOIN 条件（pets.deleted_at IS NULL）で除外される
	species := &model.AnimalSpecies{Name: "猫"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	deletedPet := &model.Pet{ClinicID: clinicA, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "削除済みペット"}
	require.NoError(t, db.WithContext(ctx).Create(deletedPet).Error)
	makeVaccinationRecord(t, db, clinicA, deletedPet.ID, vaccine.ID)
	require.NoError(t, db.WithContext(ctx).Delete(&model.Pet{}, deletedPet.ID).Error)

	other := makeTestOwner(t, db, clinicA, "無関係の飼主")

	t.Run("生存ペットの記録のみ返しVaccineがPreloadされる", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, owner.ID)
		require.NoError(t, err)
		require.Len(t, got, 1, "削除済みペットの記録は除外される")
		assert.Equal(t, aliveRec.ID, got[0].ID)
		require.NotNil(t, got[0].Vaccine)
		assert.Equal(t, vaccine.ID, got[0].Vaccine.ID)
	})

	t.Run("ペットが全く無い飼主は空配列を返す", func(t *testing.T) {
		got, err := repo.FindByOwner(ctx, clinicA, other.ID)
		require.NoError(t, err)
		assert.Len(t, got, 0)
	})
}

func TestVaccinationRepository_Update(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "更新用飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "更新用ペット")
	vaccine := makeVaccineMaster(t, db, clinicA, "更新用ワクチン")

	t.Run("同一クリニックの更新は反映される", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		got, err := repo.Update(ctx, clinicA, rec.ID, map[string]any{"remarks": "更新後の備考"})
		require.NoError(t, err)
		assert.Equal(t, "更新後の備考", got.Remarks)
	})

	t.Run("別クリニックの更新はNotFound", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		_, err := repo.Update(ctx, clinicB, rec.ID, map[string]any{"remarks": "越境更新"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新はNotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"remarks": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestVaccinationRepository_Delete(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "削除用飼主")
	pet := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "削除用ペット")
	vaccine := makeVaccineMaster(t, db, clinicA, "削除用ワクチン")

	t.Run("同一クリニックの削除は成功しその後取得できない", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		require.NoError(t, repo.Delete(ctx, clinicA, rec.ID))

		_, err := repo.FindByID(ctx, clinicA, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの削除はNotFound", func(t *testing.T) {
		rec := makeVaccinationRecord(t, db, clinicA, pet.ID, vaccine.ID)
		err := repo.Delete(ctx, clinicB, rec.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの削除はNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestVaccinationRepository_FindOwnersByVaccineDeadline(t *testing.T) {
	db := setupVaccinationRepoTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	target := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	other := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

	species := &model.AnimalSpecies{Name: "犬"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)

	makeVaccinationWithNextDate := func(clinicID, petID uint64, vaccineID uint64, next time.Time) {
		v := &model.Vaccination{ClinicID: clinicID, PetID: &petID, VaccineID: vaccineID, Date: time.Now(), NextDate: &next}
		require.NoError(t, db.WithContext(ctx).Create(v).Error)
	}

	t.Run("next_dateが対象日の飼主IDのみ重複なく返る", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "期限飼主")
		petA := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "期限ペットA")
		petB := makeVaccinationRepoTestPet(t, db, clinicA, owner.ID, "期限ペットB")
		vaccine := makeVaccineMaster(t, db, clinicA, "期限ワクチン")

		// 同じ飼主の2匹とも対象日 → Distinct で1件のみ
		makeVaccinationWithNextDate(clinicA, petA.ID, vaccine.ID, target)
		makeVaccinationWithNextDate(clinicA, petB.ID, vaccine.ID, target)

		// 対象日ではない記録は含まれない
		otherOwner := makeTestOwner(t, db, clinicA, "対象外飼主")
		otherPet := makeVaccinationRepoTestPet(t, db, clinicA, otherOwner.ID, "対象外ペット")
		makeVaccinationWithNextDate(clinicA, otherPet.ID, vaccine.ID, other)

		ids, err := repo.FindOwnersByVaccineDeadline(ctx, clinicA, target)
		require.NoError(t, err)
		assert.ElementsMatch(t, []uint64{owner.ID}, ids)
	})

	t.Run("死亡ペットの記録は除外される", func(t *testing.T) {
		owner := makeTestOwner(t, db, clinicA, "死亡ペット飼主")
		deceasedAt := target.Add(-24 * time.Hour)
		deceasedPet := &model.Pet{ClinicID: clinicA, OwnerID: owner.ID, AnimalSpeciesID: species.ID, Name: "死亡ペット", DeceasedAt: &deceasedAt}
		require.NoError(t, db.WithContext(ctx).Create(deceasedPet).Error)
		vaccine := makeVaccineMaster(t, db, clinicA, "死亡ペット用ワクチン")
		makeVaccinationWithNextDate(clinicA, deceasedPet.ID, vaccine.ID, target)

		ids, err := repo.FindOwnersByVaccineDeadline(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, ids, owner.ID, "死亡ペットの記録は除外されるべき")
	})

	t.Run("別クリニックの記録は含まれない", func(t *testing.T) {
		ownerB := makeTestOwner(t, db, clinicB, "医院B飼主")
		petB := makeVaccinationRepoTestPet(t, db, clinicB, ownerB.ID, "医院Bペット")
		vaccineB := makeVaccineMaster(t, db, clinicB, "医院Bワクチン")
		makeVaccinationWithNextDate(clinicB, petB.ID, vaccineB.ID, target)

		ids, err := repo.FindOwnersByVaccineDeadline(ctx, clinicA, target)
		require.NoError(t, err)
		assert.NotContains(t, ids, ownerB.ID)
	})
}
