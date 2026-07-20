package medicalrecord

// medical_record_image_repository_test.go — MedicalRecordImageRepository の統合テスト。
// 移動元 internal/repository（BE9-2D sub-batch④a）。
//
// pet_write_medimage_clinic_isolation_test.go（internal/repository に残留）が FindByID / Delete の
// clinic_id 隔離（親 medical_records への JOIN スコープ）を既にカバーしているため、本ファイルはそこで
// 未カバーだった Create / FindByMedicalRecordID（並び順・Staff preload・clinic_id 隔離）と、
// 存在しないID系の NotFound ケースを補う。
//
// setupMedImageIsolationTestDB / makeMedRecordImage は残留側 pet_write_medimage_clinic_isolation_test.go
// で定義されており package をまたいで import できないため、medicalrecord 側でローカルに再構築する
// （挙動同一。setupTestDB/ensureAutoMigrated → repotest.SetupTestDB/EnsureAutoMigrated 直呼びのみ差分）。
// makeTestOwner / makeSpeciesAndPet / makeDoctor / makeHistoryMedicalRecord は既存の medicalrecord
// 共有ヘルパーを再利用する（重複定義しない）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

// setupMedImageIsolationTestDB は medical_record_image テスト用に DB を整える。
// 残留 pet_write_medimage_clinic_isolation_test.go の同名ヘルパーの再構築版
// （setupTestDB/ensureAutoMigrated → repotest 直呼びのみ差分・AutoMigrate model set は同一）。
func setupMedImageIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := repotest.SetupTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db,
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

func TestMedicalRecordImageRepository_Create(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.Staff{}))
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "画像作成飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "画像作成犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-IMG-CREATE", time.Now())

	img := &model.MedicalRecordImage{
		MedicalRecordID: mr.ID,
		ImageURL:        "https://example.test/created.jpg",
		FileName:        "created.jpg",
		ImageType:       model.MedicalImageTypeXray,
	}
	require.NoError(t, repo.Create(ctx, img))
	assert.NotZero(t, img.ID)

	var stored model.MedicalRecordImage
	require.NoError(t, db.First(&stored, img.ID).Error)
	assert.Equal(t, "created.jpg", stored.FileName)
	assert.Equal(t, model.MedicalImageTypeXray, stored.ImageType)
}

func TestMedicalRecordImageRepository_FindByMedicalRecordID(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.Staff{}))
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "画像一覧飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "画像一覧犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-IMG-LIST", time.Now())
	staff := makeDoctor(t, db, clinicA, "画像担当医")

	first := makeMedRecordImage(t, db, mr.ID, "1-first.jpg")
	require.NoError(t, db.Model(first).Update("staff_id", staff.ID).Error)
	time.Sleep(2 * time.Millisecond)
	second := makeMedRecordImage(t, db, mr.ID, "2-second.jpg")

	// 別クリニックのカルテに紐づく画像は混入してはならない
	ownerB := makeTestOwner(t, db, clinicB, "画像一覧飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "画像一覧犬B")
	mrB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-IMG-LIST-B", time.Now())
	makeMedRecordImage(t, db, mrB.ID, "other-clinic.jpg")

	t.Run("同一クリニック・同一カルテの画像のみ sort_order/created_at 順で返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, first.ID, got[0].ID)
		assert.Equal(t, second.ID, got[1].ID)
		require.NotNil(t, got[0].Staff, "staff_id 設定済みの画像は Staff を preload するべき")
		assert.Equal(t, staff.ID, got[0].Staff.ID)
		assert.Nil(t, got[1].Staff, "staff_id 未設定の画像は Staff が nil であるべき")
	})

	t.Run("別クリニックIDでは0件を返す（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicB, mr.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("画像が無いカルテIDは空スライスを返す", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, uint64(9999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestMedicalRecordImageRepository_FindByID_NotFound(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	got, err := repo.FindByID(ctx, clinicA, uint64(9999999))
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestMedicalRecordImageRepository_Delete_NotFound(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	err := repo.Delete(ctx, clinicA, uint64(9999999))
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
