package medicalrecord

// vaccination_master_preload_clinic_isolation_test.go
// クロステナント READ IDOR 監査（72e8887c write 監査の read 版）回帰テスト。
//
// 保護する不変条件: vaccinations.clinic_id でテナント隔離されるが、Preload する
// Vaccine マスタは FK 値 (vaccination.vaccine_id) で引かれる。vaccine_id が別クリニックの
// vaccine を指す場合（write 側が vaccine_id の clinic 所有権を未検証 / 過去のデータ汚染 #125
// フィラリアワクチン誤参照 vaccine_id 5→2）、Preload に clinic_id 述語が無いと
// 別クリニックの vaccine 名・価格が応答に混入する。
//
// このテストは vaccination_repository.go の Vaccine Preload から "clinic_id = ?" を外すと失敗する。

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

func setupVaccinationMasterPreloadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.Vaccine{}, &model.Vaccination{},
	))
	db.Exec("TRUNCATE TABLE vaccinations CASCADE")
	db.Exec("TRUNCATE TABLE vaccines CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeVaccineMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Vaccine {
	t.Helper()
	v := &model.Vaccine{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(v).Error)
	return v
}

func makeVaccinationRecord(t *testing.T, db *gorm.DB, clinicID, petID, vaccineID uint64) *model.Vaccination {
	t.Helper()
	pid := petID
	rec := &model.Vaccination{
		ClinicID:  clinicID,
		PetID:     &pid,
		VaccineID: vaccineID,
		Date:      time.Now(),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rec).Error)
	return rec
}

// TestVaccinationRepository_FindByID_VaccinePreloadClinicIsolation は
// clinic A のワクチン記録が別クリニック(B)の vaccine を指す FK を持つ場合は、
// 汚染行自体を not-found として別クリニックの生IDも漏らさないことを検証する（#125 同型）。
func TestVaccinationRepository_FindByID_VaccinePreloadClinicIsolation(t *testing.T) {
	db := setupVaccinationMasterPreloadTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	ownerA := makeTestOwner(t, db, clinicA, "ワクチン隔離飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ワクチンポチA")
	vaccineB := makeVaccineMaster(t, db, clinicB, "医院Bの狂犬病ワクチン")

	// clinic A の記録に別クリニックの vaccine_id を植え付ける（#125 再現）。
	cross := makeVaccinationRecord(t, db, clinicA, petA.ID, vaccineB.ID)

	got, err := repo.FindByID(ctx, clinicA, cross.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got, "別クリニックの vaccine ID を持つ汚染行は返してはならない")
}

// TestVaccinationRepository_FindByID_SameClinicVaccinePreloaded は
// clinic 隔離 Preload を追加しても同一クリニックの vaccine は従来どおり Preload されることを検証する。
func TestVaccinationRepository_FindByID_SameClinicVaccinePreloaded(t *testing.T) {
	db := setupVaccinationMasterPreloadTestDB(t)
	repo := NewVaccinationRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	ownerA := makeTestOwner(t, db, clinicA, "正常ワクチン飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "正常ワクチンタマA")
	vaccineA := makeVaccineMaster(t, db, clinicA, "医院Aの狂犬病ワクチン")

	legit := makeVaccinationRecord(t, db, clinicA, petA.ID, vaccineA.ID)

	got, err := repo.FindByID(ctx, clinicA, legit.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Vaccine, "同一クリニックの vaccine マスタは Preload されるべき")
	assert.Equal(t, vaccineA.ID, got.Vaccine.ID)
}
