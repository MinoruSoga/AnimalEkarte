package medicalrecord

// lab_import_repository_test.go — moved from internal/repository (BE9-2D sub-batch③).
// LabImportJobRepository / LabImportEventRepository / LabImportDuplicateCheckerDB の統合テスト。
//
// 保護する不変条件:
//   - LabImportJobRepository.Update は clinic_id スコープ（別クリニックからは更新できない）。
//   - FindByID は clinic_id で正しく分離される。
//   - LabImportEventRepository.FindByJob は job_id + clinic_id で分離され created_at 昇順。
//   - LabImportDuplicateCheckerDB.IsDuplicate は完全同一ペイロード（header + items）のみ
//     重複とし、同日・同検査種別で内容が異なる再検査は重複にしない（Issue #249 R-3）。
//     ソフトデリート済み exam は重複扱いしない。clinic_id 隔離を維持する。
//
// setupTestDB → testdb.SetupTestDB, ensureAutoMigrated → testdb.EnsureAutoMigrated に置換
// （medicalrecord-local セットアップ・sub-batch①/②先例）。makeTestOwner / makeSpeciesAndPet は
// medicalrecord 内の既存ヘルパーを流用。makeExamRec と makeLabImportExamTypeMaster はフラット側の
// 定義を本 package へ移植（前者は preload テスト用に repository 側にも残る別 package コピー、後者は
// IsActive を立てない lab 固有の fixture なので exam_type_repository_test.go の makeExamTypeMaster とは別物）。

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

// makeExamRec はテスト用の Examination を作成して返す（重複判定の下地・Date は now()）。
// フラット側 preload_followup_clinic_isolation_test.go と同型のヘルパーを本 package に移植したもの。
func makeExamRec(t *testing.T, db *gorm.DB, clinicID, petID, examTypeID uint64) uint64 {
	t.Helper()
	pid := petID
	e := &model.Examination{ClinicID: clinicID, PetID: &pid, ExamTypeID: examTypeID, Date: time.Now()}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e.ID
}

// setupLabImportTestDB は lab_import_jobs / lab_import_events と、重複チェックに使う
// exams 周辺のテーブルを整備する。
func setupLabImportTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// testdb.SetupTestDB の共通 ENUM リストには lab_import_jobs 専用の2型
	// (lab_import_job_status / lab_import_source_type、migrations/001_init.sql に定義。
	// 旧 005_add_lab_import_tables.sql は 2026-07-04 に統合済み・docs/architecture/erd.md §4.3) が
	// 含まれないため、AutoMigrate 前にここで明示的に作成する（無ければ CREATE TABLE が
	// 42704 type does not exist で失敗する）。
	for _, stmt := range []string{
		`DO $$ BEGIN
			CREATE TYPE lab_import_job_status AS ENUM ('received','validated','mapped','persisted','duplicate','needs_review','failed','reverted');
		EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
		`DO $$ BEGIN
			CREATE TYPE lab_import_source_type AS ENUM ('fixture','drwan','manual');
		EXCEPTION WHEN duplicate_object THEN NULL; END $$;`,
	} {
		require.NoError(t, db.Exec(stmt).Error)
	}
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.LabImportJob{}, &model.LabImportEvent{},
		&model.AnimalSpecies{}, &model.Pet{}, &model.ExaminationType{}, &model.Examination{}, &model.ExamResult{},
	))
	db.Exec("TRUNCATE TABLE lab_import_events CASCADE")
	db.Exec("TRUNCATE TABLE lab_import_jobs CASCADE")
	db.Exec("TRUNCATE TABLE exam_results CASCADE")
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

	// pet_id ありの exam を明示的に日付指定で作成（header のみ・items なし）
	examWithPet := &model.Examination{
		ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: date, Machine: "Analyzer-A",
	}
	require.NoError(t, db.WithContext(ctx).Create(examWithPet).Error)

	// pet_id なしの exam
	examNoPet := &model.Examination{ClinicID: clinicA, ExamTypeID: examType.ID, Date: date, Machine: "Analyzer-A"}
	require.NoError(t, db.WithContext(ctx).Create(examNoPet).Error)

	// 同一 header + 空 items は完全同一再インポートとして重複
	t.Run("完全同一header(空items)は重複と判定する", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, LabExamPersistInput{
			ClinicID: clinicA, ExamTypeID: examType.ID, Date: date, PetID: &pet.ID, Machine: "Analyzer-A",
		})
		require.NoError(t, err)
		assert.True(t, dup)
	})

	t.Run("pet_idがnilの行はpet_id IS NULLとして完全一致判定する", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, LabExamPersistInput{
			ClinicID: clinicA, ExamTypeID: examType.ID, Date: date, PetID: nil, Machine: "Analyzer-A",
		})
		require.NoError(t, err)
		assert.True(t, dup)
	})

	t.Run("該当日に記録が無ければ重複ではない", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, LabExamPersistInput{
			ClinicID: clinicA, ExamTypeID: examType.ID, Date: date.AddDate(0, 0, 1), PetID: &pet.ID, Machine: "Analyzer-A",
		})
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("別クリニックの同条件は重複ではない（clinic_id隔離）", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, LabExamPersistInput{
			ClinicID: clinicB, ExamTypeID: examType.ID, Date: date, PetID: &pet.ID, Machine: "Analyzer-A",
		})
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("ソフトデリート済みexamは重複判定に含まれない", func(t *testing.T) {
		deletedDate := date.AddDate(0, 0, 2)
		deletedExam := &model.Examination{
			ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: deletedDate, Machine: "Analyzer-A",
		}
		require.NoError(t, db.WithContext(ctx).Create(deletedExam).Error)
		require.NoError(t, db.WithContext(ctx).Delete(deletedExam).Error) // soft delete

		dup, err := checker.IsDuplicate(ctx, LabExamPersistInput{
			ClinicID: clinicA, ExamTypeID: examType.ID, Date: deletedDate, PetID: &pet.ID, Machine: "Analyzer-A",
		})
		require.NoError(t, err)
		assert.False(t, dup, "ソフトデリート済みexamは重複扱いされないべき")
	})
}

// TestLabImportDuplicateCheckerDB_IsDuplicate_FullIdenticalOnly は Issue #249 R-3:
// 同日・同検査種別でも内容が異なれば重複ではなく、完全同一ペイロードのみスキップする。
func TestLabImportDuplicateCheckerDB_IsDuplicate_FullIdenticalOnly(t *testing.T) {
	db := setupLabImportTestDB(t)
	checker := NewLabImportDuplicateCheckerDB(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "R3重複判定飼主")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "R3重複判定犬")
	examType := makeLabImportExamTypeMaster(t, db, clinicA, "生化学")
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	refMin, refMax := 8.0, 30.0

	seedItems := func(examID uint64, value string) {
		t.Helper()
		item := model.ExamResult{
			ExamID:          examID,
			Name:            "BUN",
			InspectionValue: value,
			Unit:            "mg/dL",
			ReferenceValue:  "8-30",
			RefMin:          &refMin,
			RefMax:          &refMax,
			SortOrder:       1,
		}
		require.NoError(t, db.WithContext(ctx).Create(&item).Error)
	}

	// 既存 exam: BUN=12.0
	existing := &model.Examination{
		ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: date, Machine: "Fuji",
	}
	require.NoError(t, db.WithContext(ctx).Create(existing).Error)
	seedItems(existing.ID, "12.0")

	baseInput := LabExamPersistInput{
		ClinicID:   clinicA,
		PetID:      &pet.ID,
		ExamTypeID: examType.ID,
		Date:       date,
		Machine:    "Fuji",
		Items: []LabExamItemInput{{
			Name: "BUN", InspectionValue: "12.0", Unit: "mg/dL",
			ReferenceValue: "8-30", RefMin: &refMin, RefMax: &refMax, SortOrder: 1,
		}},
	}

	t.Run("完全同一header+itemsは重複", func(t *testing.T) {
		dup, err := checker.IsDuplicate(ctx, baseInput)
		require.NoError(t, err)
		assert.True(t, dup, "full-identical re-import must be skipped")
	})

	t.Run("同日同typeでInspectionValueが異なれば重複ではない", func(t *testing.T) {
		diff := baseInput
		diff.Items = []LabExamItemInput{{
			Name: "BUN", InspectionValue: "25.5", Unit: "mg/dL",
			ReferenceValue: "8-30", RefMin: &refMin, RefMax: &refMax, SortOrder: 1,
		}}
		dup, err := checker.IsDuplicate(ctx, diff)
		require.NoError(t, err)
		assert.False(t, dup, "same-day re-exam with different content must NOT be duplicate")
	})

	t.Run("同日同typeでMachineが異なれば重複ではない", func(t *testing.T) {
		diff := baseInput
		diff.Machine = "Other-Machine"
		dup, err := checker.IsDuplicate(ctx, diff)
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("同日同typeでmedical_record_idが異なれば重複ではない", func(t *testing.T) {
		mrID := uint64(999)
		diff := baseInput
		diff.MedicalRecordID = &mrID
		dup, err := checker.IsDuplicate(ctx, diff)
		require.NoError(t, err)
		assert.False(t, dup)
	})

	t.Run("同日2回目の異なる再検査は重複ではない", func(t *testing.T) {
		// 2 件目を seed（異なる値）
		second := &model.Examination{
			ClinicID: clinicA, PetID: &pet.ID, ExamTypeID: examType.ID, Date: date, Machine: "Fuji",
		}
		require.NoError(t, db.WithContext(ctx).Create(second).Error)
		seedItems(second.ID, "18.0")

		// 3 回目: さらに別の値 → どちらとも一致しないので false
		third := baseInput
		third.Items = []LabExamItemInput{{
			Name: "BUN", InspectionValue: "99.0", Unit: "mg/dL",
			ReferenceValue: "8-30", RefMin: &refMin, RefMax: &refMax, SortOrder: 1,
		}}
		dup, err := checker.IsDuplicate(ctx, third)
		require.NoError(t, err)
		assert.False(t, dup, "second re-exam with different values must insert")

		// 既存 18.0 と同一なら true
		sameAsSecond := baseInput
		sameAsSecond.Items = []LabExamItemInput{{
			Name: "BUN", InspectionValue: "18.0", Unit: "mg/dL",
			ReferenceValue: "8-30", RefMin: &refMin, RefMax: &refMax, SortOrder: 1,
		}}
		dup, err = checker.IsDuplicate(ctx, sameAsSecond)
		require.NoError(t, err)
		assert.True(t, dup, "re-import of already-stored second exam content must skip")
	})
}
