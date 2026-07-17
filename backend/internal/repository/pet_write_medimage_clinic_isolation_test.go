package repository

// pet_write_medimage_clinic_isolation_test.go — BE-refactor.md R3-4 (D6) の repository テスト補強。
//
// 既存の owner_pet_clinic_isolation_test.go は Pet の FindByID / Delete のみをカバーしていた
// （D6: 「pet CRUD は R/D のみ」）。本ファイルは write 経路の clinic スコープ（Update）と
// medical_record_image の隔離（テスト 0 件だった層）を補い、最弱層 repository の回帰検出力を上げる。
//
// pool 枯渇（D6 の別懸念）は getTestDatabaseConnection で解消済み（mainDB close + MaxOpenConns(10)
// + t.Cleanup(close)）のため本補強では触れない。

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

// TestPetRepository_Update_ClinicIsolation は clinic A のペットを clinic B のスコープで更新できない
// ことを検証する（Update の clinicScope を外すと別クリニックのペットを改ざんできてしまう）。
func TestPetRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := NewPetRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "医院Aの飼主")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "医院Aのポチ")

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, petA.ID, map[string]any{"name": "改ざん"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックの Update は NotFound: %v", err)
	})

	t.Run("ペットの値は変更されていない（データ改ざん防止）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.Equal(t, "医院Aのポチ", got.Name, "越境 Update で値が書き換わってはならない")
	})

	t.Run("同一クリニックIDからの Update は成功する（false-reject なし）", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, petA.ID, map[string]any{"name": "正規更新"})
		require.NoError(t, err)
		got, err := repo.FindByID(ctx, clinicA, petA.ID)
		require.NoError(t, err)
		assert.Equal(t, "正規更新", got.Name)
	})
}

// setupMedImageIsolationTestDB は medical_record_image 隔離テスト用に DB を整える。
func setupMedImageIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db,
		&model.AnimalSpecies{}, &model.Pet{}, &model.MedicalRecord{}, &model.MedicalRecordImage{},
	))
	db.Exec("TRUNCATE TABLE medical_record_images CASCADE")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

// makeMedRecordImage は指定 medical_record に画像を1件作る。
func makeMedRecordImage(t *testing.T, db *gorm.DB, medicalRecordID uint64, fileName string) *model.MedicalRecordImage {
	t.Helper()
	img := &model.MedicalRecordImage{
		MedicalRecordID: medicalRecordID,
		ImageURL:        "https://example.test/" + fileName,
		FileName:        fileName,
		ImageType:       model.MedicalImageTypeOther,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(img).Error)
	return img
}

// TestMedicalRecordImageRepository_FindByID_ClinicIsolation は medical_record_image が clinic_id 列を
// 持たず親 medical_records.clinic_id への JOIN でスコープすることの隔離を検証する（テスト 0 件だった層）。
func TestMedicalRecordImageRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "画像飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "画像ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-IMG-A", time.Now())
	imgA := makeMedRecordImage(t, db, mrA.ID, "a.jpg")

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, imgA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, imgA.ID, got.ID)
	})

	t.Run("別クリニックIDでは取得できない（親カルテ JOIN スコープ）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, imgA.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックからは NotFound: %v", err)
	})
}

// TestMedicalRecordImageRepository_Delete_ClinicIsolation は clinic A の画像を clinic B のスコープで
// 削除できないことを検証する。
func TestMedicalRecordImageRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	ownerA := makeTestOwner(t, db, clinicA, "画像飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "画像ポチA")
	mrA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-IMG-DEL", time.Now())
	imgA := makeMedRecordImage(t, db, mrA.ID, "del.jpg")

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, imgA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "別クリニックの Delete は NotFound: %v", err)
	})

	t.Run("画像はまだ存在する（不正削除防止）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, imgA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("同一クリニックIDからの Delete は成功する", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, imgA.ID)
		require.NoError(t, err)
		_, err = repo.FindByID(ctx, clinicA, imgA.ID)
		assert.True(t, apperrors.IsNotFound(err), "削除後は取得できない")
	})
}
