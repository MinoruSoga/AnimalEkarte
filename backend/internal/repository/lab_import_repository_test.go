package repository

// lab_import_repository_test.go — LabImportJobRepository / LabImportEventRepository /
// LabImportDuplicateCheckerDB の統合テスト。
//
// 保護する不変条件:
//   - LabImportJobRepository.Update は clinic_id スコープ（別クリニックからは更新できない）。
//   - FindByID は clinic_id で正しく分離される。
//   - LabImportEventRepository.FindByJob は job_id + clinic_id で分離され created_at 昇順。
//   - LabImportDuplicateCheckerDB.IsDuplicate は (clinic_id, exam_type_id, date, pet_id) の
//     複合キーで判定し、ソフトデリート済み exam は重複扱いしない。

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupLabImportTestDB は lab_import_jobs / lab_import_events と、重複チェックに使う
// exams 周辺のテーブルを整備する。
func setupLabImportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	// setupTestDB の共通 ENUM リストには lab_import_jobs 専用の2型
	// (lab_import_job_status / lab_import_source_type、migrations/001_init.sql に定義。
	// 旧 005_add_lab_import_tables.sql は 2026-07-04 に統合済み・docs/ERD.md §4.3) が
	// 含まれないため、AutoMigrate 前にここで明示的に作成する（無ければ CREATE TABLE が
	// 42704 type does not exist で失敗する）。
	for _, stmt := range []string{
		`DO $$ BEGIN
			CREATE TYPE lab_import_job_status AS ENUM ('received','validated','mapped','persisted','duplicate','needs_review','failed');
		EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN
			CREATE TYPE lab_import_source_type AS ENUM ('fixture','drwan','manual');
		EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}
	require.NoError(t, db.AutoMigrate(
		&model.LabImportJob{}, &model.LabImportEvent{},
		&model.AnimalSpecies{}, &model.Pet{}, &model.ExaminationType{}, &model.Examination{},
	))
	db.Exec("TRUNCATE TABLE lab_import_events CASCADE")
	db.Exec("TRUNCATE TABLE lab_import_jobs CASCADE")
	db.Exec("TRUNCATE TABLE exams CASCADE")
	db.Exec("TRUNCATE TABLE exam_types CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	return db
}

func makeLabImportJob(t *testing.T, db *gorm.DB, clinicID uint64, status model.LabImportJobStatus) *model.LabImportJob {
	t.Helper()
	job := &model.LabImportJob{
		ClinicID:   clinicID,
		SourceType: model.LabImportSourceTypeFixture,
		Status:     status,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(job).Error)
	return job
}

func makeLabImportExamTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.ExaminationType {
	t.Helper()
	et := &model.ExaminationType{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(et).Error)
	return et
}

func TestLabImportJobRepository_Create(t *testing.T) {
	db := setupLabImportTestDB(t)
	repo := NewLabImportJobRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	job := &model.LabImportJob{ClinicID: clinicA, SourceType: model.LabImportSourceTypeFixture, Status: model.LabImportJobStatusReceived}
	require.NoError(t, repo.Create(ctx, job))
	assert.NotEqual(t, uuid.Nil, job.ID, "UUID が生成されるべき")

	var stored model.LabImportJob
	require.NoError(t, db.First(&stored, "id = ?", job.ID).Error)
	assert.Equal(t, clinicA, stored.ClinicID)
	assert.Equal(t, model.LabImportJobStatusReceived, stored.Status)
}

func TestLabImportJobRepository_Update(t *testing.T) {
	db := setupLabImportTestDB(t)
	repo := NewLabImportJobRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの更新は成功する", func(t *testing.T) {
		job := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)
		job.Status = model.LabImportJobStatusValidated
		job.RowCount = 10
		require.NoError(t, repo.Update(ctx, job))

		var stored model.LabImportJob
		require.NoError(t, db.First(&stored, "id = ?", job.ID).Error)
		assert.Equal(t, model.LabImportJobStatusValidated, stored.Status)
		assert.Equal(t, 10, stored.RowCount)
	})

	t.Run("別クリニックIDを指定した更新はNotFoundを返す", func(t *testing.T) {
		job := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)
		// clinic_id を書き換えた構造体で Update を呼ぶと、job.ClinicID が Scope に使われるため
		// 別クリニック扱いになり対象行が見つからない。
		mismatched := *job
		mismatched.ClinicID = clinicB
		mismatched.Status = model.LabImportJobStatusFailed

		err := repo.Update(ctx, &mismatched)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stored model.LabImportJob
		require.NoError(t, db.First(&stored, "id = ?", job.ID).Error)
		assert.Equal(t, model.LabImportJobStatusReceived, stored.Status, "別クリニックからの更新で状態が変わってはならない")
	})
}

func TestLabImportJobRepository_FindByID(t *testing.T) {
	db := setupLabImportTestDB(t)
	repo := NewLabImportJobRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	job := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, job.ID)
		require.NoError(t, err)
		assert.Equal(t, job.ID, got.ID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, job.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uuid.New())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLabImportEventRepository_Create(t *testing.T) {
	db := setupLabImportTestDB(t)
	repo := NewLabImportEventRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	job := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)
	event := &model.LabImportEvent{
		ClinicID:  clinicA,
		JobID:     job.ID,
		EventType: model.LabImportEventTypeStatusTransition,
	}
	require.NoError(t, repo.Create(ctx, event))
	assert.NotZero(t, event.ID)

	var stored model.LabImportEvent
	require.NoError(t, db.First(&stored, event.ID).Error)
	assert.Equal(t, job.ID, stored.JobID)
}

func TestLabImportEventRepository_FindByJob(t *testing.T) {
	db := setupLabImportTestDB(t)
	repo := NewLabImportEventRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	job := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)
	otherJob := makeLabImportJob(t, db, clinicA, model.LabImportJobStatusReceived)

	first := &model.LabImportEvent{ClinicID: clinicA, JobID: job.ID, EventType: model.LabImportEventTypeStatusTransition}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)
	time.Sleep(2 * time.Millisecond)
	second := &model.LabImportEvent{ClinicID: clinicA, JobID: job.ID, EventType: model.LabImportEventTypeValidationResult}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	// 別ジョブのイベントは混入してはならない
	require.NoError(t, db.WithContext(ctx).Create(&model.LabImportEvent{ClinicID: clinicA, JobID: otherJob.ID, EventType: model.LabImportEventTypeStatusTransition}).Error)

	t.Run("同一ジョブのイベントのみ時系列昇順で返す", func(t *testing.T) {
		got, err := repo.FindByJob(ctx, clinicA, job.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, first.ID, got[0].ID)
		assert.Equal(t, second.ID, got[1].ID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByJob(ctx, clinicB, job.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestLabImportDuplicateCheckerDB_IsDuplicate(t *testing.T) {
	db := setupLabImportTestDB(t)
	checker := NewLabImportDuplicateCheckerDB(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "重複判定飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "重複判定犬")
	examType := makeLabImportExamTypeMaster(t, db, clinicA, "血液検査")
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	makeExamRec(t, db, clinicA, pet.ID, examType.ID) // Date は now() だが以下の未紐付ケースで別途 exam を作る

	// pet_id ありの exam を明示的に日付指定で作成
	examWithPet := &model.Examination{ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: date}
	require.NoError(t, db.WithContext(ctx).Create(examWithPet).Error)

	// pet_id なしの exam
	examNoPet := &model.Examination{ClinicID: clinicA, ExamTypeID: examType.ID, Date: date}
	require.NoError(t, db.WithContext(ctx).Create(examNoPet).Error)

	t.Run("同一(clinic,exam_type,date,pet_id)は重複と判定する", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, clinicA, examType.ID, date, &pet.ID)
		require.NoError(t, err)
		assert.True(t, dup)
	})

	t.Run("pet_idがnilの行はpet_id IS NULLとして重複判定する", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, clinicA, examType.ID, date, nil)
		require.NoError(t, err)
		assert.True(t, dup)
	})

	t.Run("該当日に記録が無ければ重複ではない", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, clinicA, examType.ID, date.AddDate(0, 0, 1), &pet.ID)
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("別クリニックの同条件は重複ではない（clinic_id隔離）", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, clinicB, examType.ID, date, &pet.ID)
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("ソフトデリート済みexamは重複判定に含まれない", func(t *testing.T) {
		deletedDate := date.AddDate(0, 0, 2)
		deletedExam := &model.Examination{ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: deletedDate}
		require.NoError(t, db.WithContext(ctx).Create(deletedExam).Error)
		require.NoError(t, db.WithContext(ctx).Delete(deletedExam).Error) // soft delete

		dup, err := checker.IsDuplicate(ctx, clinicA, examType.ID, deletedDate, &pet.ID)
		require.NoError(t, err)
		assert.False(t, dup, "ソフトデリート済みexamは重複扱いされないべき")
	})
}
