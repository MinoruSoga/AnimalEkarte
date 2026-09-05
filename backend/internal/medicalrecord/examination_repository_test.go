package medicalrecord

// examination_repository_test.go — ExaminationRepository の統合テスト（内部カバレッジ向上）。
//
// 対象: FindAll / FindByID / FindByJobID / Create / Update / Delete /
//       CountItemsByExamID / FindAllItemsByExamID / ReplaceItemsByExamID
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ、フィルタ/ページネーション。
//
// makeExamTypeMaster / makeExaminationRec は exam_type_repository_test.go で定義され、
// このファイルからも再利用する（同ファイルのコメントに明記済み）。

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupExaminationTestDB は exams / exam_results / exam_types / pets / owners / staffs を整備する。
func setupExaminationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ExaminationType{}, &model.ExamTypeField{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{},
		&model.MedicalRecord{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Examination{}, &model.ExamResult{},
		&model.ExaminationRevision{}, &model.ExaminationRevisionItem{},
	))
	db.Exec("TRUNCATE TABLE examination_revision_items CASCADE")
	db.Exec("TRUNCATE TABLE examination_revisions CASCADE")
	db.Exec("TRUNCATE TABLE exam_results CASCADE")
	db.Exec("TRUNCATE TABLE exams CASCADE")
	db.Exec("TRUNCATE TABLE exam_type_fields CASCADE")
	db.Exec("TRUNCATE TABLE exam_types CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	return db
}

func makeExaminationActor(t *testing.T, db *gorm.DB, clinicID uint64, name string) uint64 {
	t.Helper()
	actor := &model.Staff{
		ClinicID:  clinicID,
		Name:      name,
		IsActive:  true,
		StaffType: model.StaffTypeNurse,
	}
	require.NoError(t, db.Create(actor).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  actor.ID,
		ClinicID: clinicID,
	}).Error)
	return actor.ID
}

func TestExaminationRepository_FindAll_ClinicIsolationAndFilters(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	etA := makeExamTypeMaster(t, db, clinicA, "血液検査")
	etB := makeExamTypeMaster(t, db, clinicB, "医院Bの検査")

	ownerA := makeTestOwner(t, db, clinicA, "飼主A")
	petA1 := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ペットA1")
	petA2 := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "ペットA2")

	jun10 := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	jun20 := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	jul01 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	examA1 := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: etA.ID, PetID: &petA1.ID, Date: jun10, Status: model.ExaminationStatusPending})
	examA2 := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: etA.ID, PetID: &petA2.ID, Date: jun20, Status: model.ExaminationStatusCompleted})
	examB := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicB, ExamTypeID: etB.ID, Date: jul01, Status: model.ExaminationStatusPending})

	t.Run("clinic_id 隔離: 他院の exam は含まれない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total)
		require.Len(t, got, 2)
		for _, e := range got {
			assert.NotEqual(t, examB.ID, e.ID)
		}
		// 日付降順で先頭が examA2
		assert.Equal(t, examA2.ID, got[0].ID)
	})

	t.Run("petID フィルタ", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, &petA1.ID, nil, nil, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, examA1.ID, got[0].ID)
	})

	t.Run("ownerID フィルタ（pets を JOIN）", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, &ownerA.ID, nil, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total)
		assert.Len(t, got, 2)
	})

	t.Run("status フィルタ", func(t *testing.T) {
		status := string(model.ExaminationStatusCompleted)
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, &status, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, examA2.ID, got[0].ID)
	})

	t.Run("日付範囲フィルタ", func(t *testing.T) {
		start := "2026-06-15"
		end := "2026-06-30"
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, &start, &end, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, examA2.ID, got[0].ID)
	})

	t.Run("ページネーション", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, nil, 1, 1, false)
		require.NoError(t, err)
		assert.EqualValues(t, 2, total, "total は limit に依存せず全件数")
		require.Len(t, got, 1, "limit=1 で1件のみ返る")

		got2, _, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, nil, 2, 1, false)
		require.NoError(t, err)
		require.Len(t, got2, 1)
		assert.NotEqual(t, got[0].ID, got2[0].ID, "2ページ目は異なるレコード")
	})

	t.Run("Preload されたリレーションが取得できる", func(t *testing.T) {
		got, _, err := repo.FindAll(ctx, clinicA, &petA1.ID, nil, nil, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].ExaminationType)
		assert.Equal(t, "血液検査", got[0].ExaminationType.Name)
		require.NotNil(t, got[0].Pet)
		assert.Equal(t, "ペットA1", got[0].Pet.Name)
	})

	t.Run("medicalRecordID フィルタ", func(t *testing.T) {
		record := makeHistoryMedicalRecord(t, db, clinicA, petA1.ID, "MR-EXAM-FILTER", jun10)
		linkedExam := makeExaminationRec(t, db, &model.Examination{
			ClinicID:        clinicA,
			ExamTypeID:      etA.ID,
			PetID:           &petA1.ID,
			MedicalRecordID: &record.ID,
			Date:            jun10,
			Status:          model.ExaminationStatusPending,
		})
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, &record.ID, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, linkedExam.ID, got[0].ID)

		// カルテ検査タブの契約: pet + record 併用時は「そのカルテの検査 +
		// 同じペットの未取り込み検査」。別カルテへ取り込んだ検査は含めない。
		otherRecord := makeHistoryMedicalRecord(t, db, clinicA, petA1.ID, "MR-EXAM-FILTER-2", jun10)
		otherLinked := makeExaminationRec(t, db, &model.Examination{
			ClinicID:        clinicA,
			ExamTypeID:      etA.ID,
			PetID:           &petA1.ID,
			MedicalRecordID: &otherRecord.ID,
			Date:            jun10,
			Status:          model.ExaminationStatusPending,
		})
		merged, mergedTotal, err := repo.FindAll(ctx, clinicA, &petA1.ID, nil, &record.ID, nil, nil, nil, 1, 10, false)
		require.NoError(t, err)
		assert.EqualValues(t, 2, mergedTotal)
		mergedIDs := make([]uint64, 0, len(merged))
		for _, e := range merged {
			mergedIDs = append(mergedIDs, e.ID)
		}
		assert.ElementsMatch(t, []uint64{linkedExam.ID, examA1.ID}, mergedIDs,
			"record 取り込み済み + 未取り込み（機器由来）を返し、別カルテの検査は返さない")
		assert.NotContains(t, mergedIDs, otherLinked.ID)
	})
}

func TestExaminationRepository_FindByID(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, exam.ID, got.ID)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, exam.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestExaminationRepository_LockByIDForUpdateRequiresAmbientTransaction(t *testing.T) {
	repo := NewExaminationRepository(nil)

	got, err := repo.LockByIDForUpdate(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestExaminationRepository_PatientRelationsAreClinicScoped(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	typeA := makeExamTypeMaster(t, db, clinicA, "検査関係スコープ種別A")
	ownerA := makeTestOwner(t, db, clinicA, "検査関係スコープ飼主A")
	petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "検査関係スコープペットA")
	recordA := makeHistoryMedicalRecord(t, db, clinicA, petA.ID, "MR-EXAM-SCOPE-A", time.Now())
	require.NoError(t, db.Model(recordA).Update("owner_id", ownerA.ID).Error)
	recordA.OwnerID = &ownerA.ID
	ensureVaccinationTestClinics(t, db, clinicA, clinicB)
	validDoctor := makeDoctor(t, db, clinicB, "検査関係スコープ有効医師")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: validDoctor.ID, ClinicID: clinicA,
	}).Error)
	valid := makeExaminationRec(t, db, &model.Examination{
		ClinicID:        clinicA,
		MedicalRecordID: &recordA.ID,
		PetID:           &petA.ID,
		ExamTypeID:      typeA.ID,
		DoctorID:        &validDoctor.ID,
		Date:            time.Now(),
	})

	unassignedDoctor := makeDoctor(t, db, clinicA, "検査関係スコープ未所属医師")
	inactiveDoctor := makeDoctor(t, db, clinicA, "検査関係スコープ無効医師")
	require.NoError(t, db.Model(inactiveDoctor).UpdateColumn("is_active", false).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: inactiveDoctor.ID, ClinicID: clinicA,
	}).Error)
	nurse := makeDoctor(t, db, clinicA, "検査関係スコープ看護師")
	require.NoError(t, db.Model(nurse).UpdateColumn("staff_type", model.StaffTypeNurse).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID: nurse.ID, ClinicID: clinicA,
	}).Error)

	typeB := makeExamTypeMaster(t, db, clinicB, "検査関係スコープ種別B")
	ownerB := makeTestOwner(t, db, clinicB, "検査関係スコープ飼主B")
	petB := makeSpeciesAndPet(t, db, clinicB, ownerB.ID, "検査関係スコープペットB")
	recordB := makeHistoryMedicalRecord(t, db, clinicB, petB.ID, "MR-EXAM-SCOPE-B", time.Now())
	require.NoError(t, db.Model(recordB).Update("owner_id", ownerB.ID).Error)
	recordB.OwnerID = &ownerB.ID
	pollutedOwnerPet := makeSpeciesAndPet(
		t,
		db,
		clinicA,
		ownerB.ID,
		"検査関係スコープ別院飼主ペット",
	)

	polluted := make([]*model.Examination, 0, 7)
	polluted = append(polluted,
		makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicA, PetID: &petB.ID, ExamTypeID: typeA.ID, Date: time.Now(),
		}),
		makeExaminationRec(t, db, &model.Examination{
			ClinicID:        clinicA,
			MedicalRecordID: &recordB.ID,
			PetID:           &petA.ID,
			ExamTypeID:      typeA.ID,
			Date:            time.Now(),
		}),
		makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicA, PetID: &pollutedOwnerPet.ID, ExamTypeID: typeA.ID, Date: time.Now(),
		}),
		makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicA, PetID: &petA.ID, ExamTypeID: typeB.ID, Date: time.Now(),
		}),
	)
	for _, doctorID := range []uint64{unassignedDoctor.ID, inactiveDoctor.ID, nurse.ID} {
		polluted = append(polluted, makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicA, PetID: &petA.ID, ExamTypeID: typeA.ID,
			DoctorID: &doctorID, Date: time.Now(),
		}))
	}
	jobID := uuid.New()
	jobScopedIDs := make([]uint64, 0, 1+len(polluted))
	jobScopedIDs = append(jobScopedIDs, valid.ID)
	for _, item := range polluted {
		jobScopedIDs = append(jobScopedIDs, item.ID)
	}
	require.NoError(t, db.Model(&model.Examination{}).
		Where("id IN ?", jobScopedIDs).
		Update("job_id", jobID).Error)

	for _, item := range polluted {
		got, err := repo.FindByID(ctx, clinicA, item.ID)
		require.Error(t, err, "polluted examination %d must fail closed", item.ID)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, got)
	}

	listed, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, nil, nil, nil, 1, 100, false)
	require.NoError(t, err)
	require.EqualValues(t, 1+len(polluted), total, "COUNT is clinic-scoped; isolation stays on Find")
	require.Len(t, listed, 1)
	assert.Empty(t, listed[0].Items)
	assert.Equal(t, valid.ID, listed[0].ID)
	require.NotNil(t, listed[0].Pet)
	require.NotNil(t, listed[0].Pet.Owner)
	assert.Equal(t, ownerA.ID, listed[0].Pet.Owner.ID)
	require.NotNil(t, listed[0].Doctor)
	assert.Equal(t, validDoctor.ID, listed[0].Doctor.ID)

	byJob, err := repo.FindByJobID(ctx, clinicA, jobID)
	require.NoError(t, err)
	require.Len(t, byJob, 1)
	assert.Equal(t, valid.ID, byJob[0].ID)
}

func TestExaminationRepository_FindByID_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	typeA := makeExamTypeMaster(t, db, clinicID, "譲渡後検査")
	previousOwner := makeTestOwner(t, db, clinicID, "譲渡前飼主")
	currentOwner := makeTestOwner(t, db, clinicID, "譲渡後飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, previousOwner.ID, "譲渡ペット")
	record := makeHistoryMedicalRecord(t, db, clinicID, pet.ID, "MR-EXAM-TRANSFER", time.Now())
	record.OwnerID = &previousOwner.ID
	require.NoError(t, db.Model(record).Update("owner_id", previousOwner.ID).Error)
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicID, ExamTypeID: typeA.ID, PetID: &pet.ID,
		MedicalRecordID: &record.ID, Date: time.Now(),
	})
	require.NoError(t, db.Model(&model.Pet{}).Where("id = ?", pet.ID).Update("owner_id", currentOwner.ID).Error)

	got, err := repo.FindByID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, exam.ID, got.ID)
	var persisted model.MedicalRecord
	require.NoError(t, db.First(&persisted, record.ID).Error)
	require.NotNil(t, persisted.OwnerID)
	assert.Equal(t, previousOwner.ID, *persisted.OwnerID)
}

func TestExaminationRepository_FindByJobID(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	job := uuid.New()
	otherJob := uuid.New()

	jun10 := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	jun20 := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

	e1 := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: jun10, JobID: &job})
	e2 := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: jun20, JobID: &job})
	makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: jun10, JobID: &otherJob})
	makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: jun10}) // job_id なし

	t.Run("job_id + clinic_id で絞り込み、日付降順で返る", func(t *testing.T) {
		got, err := repo.FindByJobID(ctx, clinicA, job)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, e2.ID, got[0].ID)
		assert.Equal(t, e1.ID, got[1].ID)
	})

	t.Run("別クリニックでは空", func(t *testing.T) {
		got, err := repo.FindByJobID(ctx, clinicB, job)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("存在しない job_id では空", func(t *testing.T) {
		got, err := repo.FindByJobID(ctx, clinicA, uuid.New())
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestExaminationRepository_Create(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)}
	require.NoError(t, repo.Create(ctx, exam))
	assert.NotZero(t, exam.ID)

	got, err := repo.FindByID(ctx, clinicA, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, exam.ID, got.ID)
}

func TestExaminationRepository_Update(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, exam.ID, map[string]any{"status": string(model.ExaminationStatusCompleted)})
		require.NoError(t, err)
		assert.Equal(t, model.ExaminationStatusCompleted, got.Status)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, exam.ID, map[string]any{"status": string(model.ExaminationStatusConfirmed)})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"status": string(model.ExaminationStatusConfirmed)})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestExaminationRepository_Delete(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, exam.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, exam.ID))

		_, err := repo.FindByID(ctx, clinicA, exam.ID)
		assert.True(t, apperrors.IsNotFound(err))

		var raw model.Examination
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", exam.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestExaminationRepository_CountItemsByExamID(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	t.Run("items が無ければ 0", func(t *testing.T) {
		count, err := repo.CountItemsByExamID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	require.NoError(t, db.WithContext(ctx).Create(&model.ExamResult{ExamID: exam.ID, Name: "WBC"}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ExamResult{ExamID: exam.ID, Name: "RBC"}).Error)

	t.Run("items が2件あれば 2", func(t *testing.T) {
		count, err := repo.CountItemsByExamID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountItemsByExamID(ctx, clinicB, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("親 exam がソフトデリートされていれば 0", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, exam.ID))
		count, err := repo.CountItemsByExamID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestExaminationRepository_FindAllItemsByExamID(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	require.NoError(t, db.WithContext(ctx).Create(&model.ExamResult{ExamID: exam.ID, Name: "RBC", SortOrder: 2}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ExamResult{ExamID: exam.ID, Name: "WBC", SortOrder: 1}).Error)

	t.Run("sort_order 昇順で返る", func(t *testing.T) {
		got, err := repo.FindAllItemsByExamID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "WBC", got[0].Name)
		assert.Equal(t, "RBC", got[1].Name)
	})

	t.Run("別クリニックIDでは空", func(t *testing.T) {
		got, err := repo.FindAllItemsByExamID(ctx, clinicB, exam.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("親 exam がソフトデリートされていれば空", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, exam.ID))
		got, err := repo.FindAllItemsByExamID(ctx, clinicA, exam.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestExaminationRepository_ReplaceItemsByExamID(t *testing.T) {
	db := setupExaminationTestDB(t)
	repo := NewExaminationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "血液検査")
	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)})

	t.Run("新規 items を一括作成できる（渡した ID は無視され新規採番される）", func(t *testing.T) {
		items := []model.ExamResult{
			{ID: 999999, Name: "WBC", InspectionValue: "10.0"},
			{ID: 999998, Name: "RBC", InspectionValue: "5.0"},
		}
		got, deleted, err := repo.ReplaceItemsByExamID(ctx, clinicA, exam.ID, items)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.EqualValues(t, 0, deleted, "既存 item が無い状態の初回置換は削除 0 件")
		for _, item := range got {
			assert.NotEqual(t, uint64(999999), item.ID)
			assert.NotEqual(t, uint64(999998), item.ID)
			assert.Equal(t, exam.ID, item.ExamID, "ExamID は強制上書きされる")
		}
	})

	t.Run("再度差し替えると旧 items が消え新 items のみ残る", func(t *testing.T) {
		got, deleted, err := repo.ReplaceItemsByExamID(ctx, clinicA, exam.ID, []model.ExamResult{{Name: "Glucose"}})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "Glucose", got[0].Name)
		assert.EqualValues(t, 2, deleted, "直前の2件が削除される（#211 監査ゲートの根拠）")
	})

	t.Run("空スライスへの差し替えで全削除される", func(t *testing.T) {
		got, deleted, err := repo.ReplaceItemsByExamID(ctx, clinicA, exam.ID, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.EqualValues(t, 1, deleted, "直前の1件が削除される")
	})

	t.Run("存在しない exam ID は NotFound", func(t *testing.T) {
		_, _, err := repo.ReplaceItemsByExamID(ctx, clinicA, 999999, []model.ExamResult{{Name: "x"}})
		require.Error(t, err)
	})

	t.Run("別クリニックの exam ID は NotFound", func(t *testing.T) {
		_, _, err := repo.ReplaceItemsByExamID(ctx, clinicB, exam.ID, []model.ExamResult{{Name: "x"}})
		require.Error(t, err)
	})
}
