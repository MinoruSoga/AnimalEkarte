package medicalrecord

// medical_record_image_repository_test.go — MedicalRecordImageRepository の統合テスト。
// 移動元 internal/repository（BE9-2D sub-batch④a）。
//
// internal/repository から移動した FindByID / Delete の clinic_id 隔離に加え、
// Create / FindByMedicalRecordID（並び順・Staff preload・clinic_id 隔離）と、
// 存在しないID系の NotFound ケースを同じ target package でカバーする。
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
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupMedImageIsolationTestDB は medical_record_image テスト用に DB を整える。
// 残留 pet_write_medimage_clinic_isolation_test.go の同名ヘルパーの再構築版
// （setupTestDB/ensureAutoMigrated → repotest 直呼びのみ差分・AutoMigrate model set は同一）。
func setupMedImageIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Owner{},
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.MedicalRecord{},
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.ExaminationType{},
		&model.Examination{},
		&model.MedicalRecordImage{},
	))
	require.NoError(t, db.Exec(`
		TRUNCATE TABLE
			medical_record_images,
			exams,
			exam_types,
			staff_clinic_assignments,
			staffs,
			medical_records,
			pets,
			animal_species
		CASCADE
	`).Error)
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
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Staff{}))
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
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Staff{}))
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	owner := makeTestOwner(t, db, clinicA, "画像一覧飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "画像一覧犬")
	mr := makeHistoryMedicalRecord(t, db, clinicA, pet.ID, "MR-IMG-LIST", time.Now())
	staff := makeDoctor(t, db, clinicA, "画像担当医")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: staff.ID, ClinicID: clinicA,
	}).Error)

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

func TestDB_MedicalRecordImageRepositoryRejectsPollutedExamAndStaffRelations(t *testing.T) {
	db := setupMedImageIsolationTestDB(t)
	repo := NewMedicalRecordImageRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)

	ownerA := makeTestOwner(t, db, clinicA, "画像read関係飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "画像read関係患者A")
	recordA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-IMG-READ-A", time.Now())
	otherRecordA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-IMG-READ-A-OTHER", time.Now())
	ownerB := makeTestOwner(t, db, clinicB, "画像read関係飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "画像read関係患者B")
	recordB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-IMG-READ-B", time.Now())

	examTypeA := makeExamTypeMaster(t, db, clinicA, "画像read関係検査A")
	examTypeB := makeExamTypeMaster(t, db, clinicB, "画像read関係検査B")
	validExam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, MedicalRecordID: &recordA.ID, PetID: &petA.ID,
		ExamTypeID: examTypeA.ID, Date: time.Now(), Status: model.ExaminationStatusPending,
	})
	mismatchedExam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, MedicalRecordID: &otherRecordA.ID, PetID: &petA.ID,
		ExamTypeID: examTypeA.ID, Date: time.Now(), Status: model.ExaminationStatusPending,
	})
	foreignExam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicB, MedicalRecordID: &recordB.ID, PetID: &petB.ID,
		ExamTypeID: examTypeB.ID, Date: time.Now(), Status: model.ExaminationStatusPending,
	})
	deletedExam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, MedicalRecordID: &recordA.ID, PetID: &petA.ID,
		ExamTypeID: examTypeA.ID, Date: time.Now(), Status: model.ExaminationStatusPending,
	})
	require.NoError(t, db.Delete(deletedExam).Error)

	validStaff := makeImageRelationStaff(t, db, clinicB, "画像read有効担当者", true)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: validStaff.ID, ClinicID: clinicA,
	}).Error)
	foreignStaff := makeImageRelationStaff(t, db, clinicB, "画像read別院担当者", true)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: foreignStaff.ID, ClinicID: clinicB,
	}).Error)
	unassignedStaff := makeImageRelationStaff(t, db, clinicA, "画像read未所属担当者", true)
	inactiveStaff := makeImageRelationStaff(t, db, clinicA, "画像read無効担当者", true)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: inactiveStaff.ID, ClinicID: clinicA,
	}).Error)
	require.NoError(t, db.Model(inactiveStaff).UpdateColumn("is_active", false).Error)
	deletedStaff := makeImageRelationStaff(t, db, clinicA, "画像read削除担当者", true)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: deletedStaff.ID, ClinicID: clinicA,
	}).Error)
	require.NoError(t, db.Delete(deletedStaff).Error)
	deletedAssignmentStaff := makeImageRelationStaff(t, db, clinicA, "画像read削除所属担当者", true)
	deletedAssignment := &model.StaffClinicAssignment{
		StaffID: deletedAssignmentStaff.ID, ClinicID: clinicA,
	}
	require.NoError(t, db.Create(deletedAssignment).Error)
	require.NoError(t, db.Delete(deletedAssignment).Error)

	valid := makeMedRecordImageWithRelations(
		t, db, recordA.ID, "valid-relations.jpg", &validExam.ID, &validStaff.ID,
	)
	withoutRelations := makeMedRecordImage(t, db, recordA.ID, "without-relations.jpg")
	polluted := []*model.MedicalRecordImage{
		makeMedRecordImageWithRelations(t, db, recordA.ID, "foreign-exam.jpg", &foreignExam.ID, nil),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "mismatched-exam.jpg", &mismatchedExam.ID, nil),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "deleted-exam.jpg", &deletedExam.ID, nil),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "foreign-staff.jpg", nil, &foreignStaff.ID),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "unassigned-staff.jpg", nil, &unassignedStaff.ID),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "inactive-staff.jpg", nil, &inactiveStaff.ID),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "deleted-staff.jpg", nil, &deletedStaff.ID),
		makeMedRecordImageWithRelations(t, db, recordA.ID, "deleted-assignment.jpg", nil, &deletedAssignmentStaff.ID),
	}

	listed, err := repo.FindByMedicalRecordID(ctx, clinicA, recordA.ID)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	byID := make(map[uint64]model.MedicalRecordImage, len(listed))
	for _, image := range listed {
		byID[image.ID] = image
	}
	validResult, ok := byID[valid.ID]
	require.True(t, ok)
	require.NotNil(t, validResult.Staff)
	assert.Equal(t, validStaff.ID, validResult.Staff.ID)
	_, ok = byID[withoutRelations.ID]
	assert.True(t, ok)

	for _, image := range polluted {
		t.Run(image.FileName, func(t *testing.T) {
			got, findErr := repo.FindByID(ctx, clinicA, image.ID)
			require.Error(t, findErr)
			assert.True(t, apperrors.IsNotFound(findErr))
			assert.Nil(t, got)
		})
	}
}

func makeImageRelationStaff(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
	name string,
	isActive bool,
) *model.Staff {
	t.Helper()
	staff := &model.Staff{
		ClinicID: clinicID, Name: name, IsActive: isActive, StaffType: model.StaffTypeDoctor,
	}
	require.NoError(t, db.Create(staff).Error)
	return staff
}

func makeMedRecordImageWithRelations(
	t *testing.T,
	db *gorm.DB,
	medicalRecordID uint64,
	fileName string,
	examID, staffID *uint64,
) *model.MedicalRecordImage {
	t.Helper()
	image := &model.MedicalRecordImage{
		MedicalRecordID: medicalRecordID,
		ImageURL:        "https://example.test/" + fileName,
		FileName:        fileName,
		ImageType:       model.MedicalImageTypeOther,
		ExamID:          examID,
		StaffID:         staffID,
	}
	require.NoError(t, db.Create(image).Error)
	return image
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
