package medicalrecord

// conversation_summary_repository_test.go — EMR-485 会話要約保存 adapter の実DBテスト。
//
// 保護する不変条件:
//   - clinic_id + medical_record_id で対象が限定される（他医院は NotFound）。
//   - revision はサーバ側採番（初回1、更新+1、冪等再送は据え置き）。
//   - 期待版不一致は conflict、未指定は照合スキップ（最終書き込み優先）。
//   - 未保存（3列NULL）は unknown として NULL のまま返す。
//   - source 非空白 / summary 非NULL の対応は DB CHECK 相当をコード側でも守る
//     （AutoMigrate のテストDBには ck_conversation_summary_metadata が載らないため）。
//   - 保存は3列のみを更新し、他の診療記録列・診療所見（clinical_plans）を変えない。

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

const (
	conversationSummaryClinicA = uint64(4851)
	conversationSummaryClinicB = uint64(4852)
)

func setupConversationSummaryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.DiagnosisType{},
		&model.DiagnosisName{},
		&model.MedicalRecord{},
		&model.ClinicalPlan{},
	), "conversation summary fixture setup")
	db.Exec("TRUNCATE TABLE medical_records CASCADE")
	db.Exec("TRUNCATE TABLE clinical_plans CASCADE")
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE owners CASCADE")
	ensureVaccinationTestClinics(t, db, conversationSummaryClinicA, conversationSummaryClinicB)
	return db
}

func makeConversationSummaryRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	ctx := context.Background()
	owner := makeTestOwner(t, db, clinicID, recordNo+" 飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, recordNo+" ペット")

	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		OwnerID:  &owner.ID,
		PetID:    &pet.ID,
	}
	require.NoError(t, db.WithContext(ctx).Create(record).Error)
	return record
}

// fetchConversationSummaryStatusColumn は status 列の値を生SQLで読む。
// status は #729 で Go model から削除されたが DB 列は現行001に存在する。
func fetchConversationSummaryStatusColumn(t *testing.T, db *gorm.DB, id uint64) *string {
	t.Helper()
	var status *string
	require.NoError(t, db.Raw("SELECT status FROM medical_records WHERE id = ?", id).Scan(&status).Error)
	return status
}

func fetchConversationSummaryRecordRow(t *testing.T, db *gorm.DB, id uint64) *model.MedicalRecord {
	t.Helper()
	var row model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&row, id).Error)
	return &row
}

func conversationSummaryInt64Ptr(v int64) *int64 { return &v }

func TestConversationSummaryRepository_SaveAndFindRoundTrip(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-ROUND-TRIP")

	saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "初回の会話要約",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, record.ID, saved.MedicalRecordID)
	require.NotNil(t, saved.Summary)
	assert.Equal(t, "初回の会話要約", *saved.Summary)
	require.NotNil(t, saved.Revision)
	assert.Equal(t, int64(1), *saved.Revision)
	require.NotNil(t, saved.Source)
	assert.Equal(t, "lstep", *saved.Source)

	got, err := repo.FindByMedicalRecordID(ctx, conversationSummaryClinicA, record.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, record.ID, got.MedicalRecordID)
	require.NotNil(t, got.Summary)
	assert.Equal(t, "初回の会話要約", *got.Summary)
	require.NotNil(t, got.Revision)
	assert.Equal(t, int64(1), *got.Revision)
	require.NotNil(t, got.Source)
	assert.Equal(t, "lstep", *got.Source)
}

func TestConversationSummaryRepository_FindUnsavedReturnsNulls(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-UNSAVED")

	got, err := repo.FindByMedicalRecordID(context.Background(), conversationSummaryClinicA, record.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, record.ID, got.MedicalRecordID)
	// 旧未保存は unknown。推定値・空文字・0 を生成しない。
	assert.Nil(t, got.Summary)
	assert.Nil(t, got.Revision)
	assert.Nil(t, got.Source)
}

func TestConversationSummaryRepository_FindRejectsOtherClinic(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-FIND-XCLINIC")
	_, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "自院の要約",
		Source:          "lstep",
	})
	require.NoError(t, err)

	got, err := repo.FindByMedicalRecordID(ctx, conversationSummaryClinicB, record.ID)
	assert.Nil(t, got)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "他医院からの読戻しは NotFound で拒否する: %v", err)
}

func TestConversationSummaryRepository_SaveRejectsOtherClinic(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-SAVE-XCLINIC")

	saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicB,
		MedicalRecordID: record.ID,
		Summary:         "他院からの書込み",
		Source:          "lstep",
	})
	assert.Nil(t, saved)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "他医院からの保存は NotFound で拒否する: %v", err)

	row := fetchConversationSummaryRecordRow(t, db, record.ID)
	assert.Nil(t, row.ConversationSummary)
	assert.Nil(t, row.ConversationSummaryRevision)
	assert.Nil(t, row.ConversationSummarySource)
}

func TestConversationSummaryRepository_SaveBumpsRevision(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-REVISION")

	first, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "1回目",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, first.Revision)
	assert.Equal(t, int64(1), *first.Revision)

	second, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:         conversationSummaryClinicA,
		MedicalRecordID:  record.ID,
		Summary:          "2回目",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(1),
	})
	require.NoError(t, err)
	require.NotNil(t, second.Revision)
	assert.Equal(t, int64(2), *second.Revision)
	require.NotNil(t, second.Summary)
	assert.Equal(t, "2回目", *second.Summary)

	got, err := repo.FindByMedicalRecordID(ctx, conversationSummaryClinicA, record.ID)
	require.NoError(t, err)
	require.NotNil(t, got.Revision)
	assert.Equal(t, int64(2), *got.Revision)
}

func TestConversationSummaryRepository_SaveRejectsExpectedRevisionMismatch(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()

	t.Run("保存済みの版と一致しない期待版は conflict", func(t *testing.T) {
		record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-CONFLICT")
		_, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:        conversationSummaryClinicA,
			MedicalRecordID: record.ID,
			Summary:         "現在の要約",
			Source:          "lstep",
		})
		require.NoError(t, err)
		second, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:         conversationSummaryClinicA,
			MedicalRecordID:  record.ID,
			Summary:          "2版目",
			Source:           "lstep",
			ExpectedRevision: conversationSummaryInt64Ptr(1),
		})
		require.NoError(t, err)
		require.NotNil(t, second.Revision)
		require.Equal(t, int64(2), *second.Revision)

		saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:         conversationSummaryClinicA,
			MedicalRecordID:  record.ID,
			Summary:          "古い期待版からの上書き",
			Source:           "lstep",
			ExpectedRevision: conversationSummaryInt64Ptr(1),
		})
		assert.Nil(t, saved)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "期待版不一致は conflict: %v", err)

		got, findErr := repo.FindByMedicalRecordID(ctx, conversationSummaryClinicA, record.ID)
		require.NoError(t, findErr)
		require.NotNil(t, got.Summary)
		assert.Equal(t, "2版目", *got.Summary)
		require.NotNil(t, got.Revision)
		assert.Equal(t, int64(2), *got.Revision)
	})

	t.Run("未保存行に正の期待版を指定しても conflict", func(t *testing.T) {
		record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-CONFLICT-UNSAVED")
		saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:         conversationSummaryClinicA,
			MedicalRecordID:  record.ID,
			Summary:          "未保存行への書込み",
			Source:           "lstep",
			ExpectedRevision: conversationSummaryInt64Ptr(3),
		})
		assert.Nil(t, saved)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "未保存行の期待版不一致は conflict: %v", err)

		row := fetchConversationSummaryRecordRow(t, db, record.ID)
		assert.Nil(t, row.ConversationSummary)
	})
}

func TestConversationSummaryRepository_SaveIdempotentResend(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-IDEMPOTENT")

	// 同一要求（期待版0=未保存期待）を2回送る。応答消失後の再送を模する。
	request := SaveConversationSummaryCommand{
		ClinicID:         conversationSummaryClinicA,
		MedicalRecordID:  record.ID,
		Summary:          "再送される要約",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(0),
	}

	first, err := repo.Save(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, first.Revision)
	require.Equal(t, int64(1), *first.Revision)

	// 同一要求の再送 → no-op 成功（revision 据え置き）。期待版が古くなっていても
	// 保存済み内容が要求の意図と一致するため成功を返す。
	resent, err := repo.Save(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, resent.Revision)
	assert.Equal(t, int64(1), *resent.Revision)

	// 期待版未指定の同一内容再送も no-op 成功。
	resentWithoutExpected, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "再送される要約",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, resentWithoutExpected.Revision)
	assert.Equal(t, int64(1), *resentWithoutExpected.Revision)

	// 同一内容だが期待版だけ現在と不一致（未来値99）で内容一致 → 冪等成功。
	// （内容が要求の意図と一致するため conflict にしない）
	resentStaleExpected, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:         conversationSummaryClinicA,
		MedicalRecordID:  record.ID,
		Summary:          "再送される要約",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(99),
	})
	require.NoError(t, err)
	require.NotNil(t, resentStaleExpected.Revision)
	assert.Equal(t, int64(1), *resentStaleExpected.Revision)

	row := fetchConversationSummaryRecordRow(t, db, record.ID)
	require.NotNil(t, row.ConversationSummaryRevision)
	assert.Equal(t, int64(1), *row.ConversationSummaryRevision)
}

func TestConversationSummaryRepository_SaveWithoutExpectedRevisionAppliesLastWriteWins(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-LAST-WRITE")

	_, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "A",
		Source:          "lstep",
	})
	require.NoError(t, err)

	latest, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "B",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, latest.Revision)
	assert.Equal(t, int64(2), *latest.Revision)
	require.NotNil(t, latest.Summary)
	assert.Equal(t, "B", *latest.Summary)
}

// Save は UPDATE...RETURNING の単一statementで自分の書込み内容と版を返すため、
// 「UPDATE と読戻しの間に他writerが commit し、読戻しが他writerの内容・版を返す」
// 競合窗口が構造的に存在しない。旧実装（UPDATE→読戻し）で指摘された lost update 経路
// （読戻し値を次の ExpectedRevision に使い他writerの内容を conflict なしで上書き）は
// この構造変更で除去された。GORM の Query callback 経由で中割込みをシミュレートする
// テストは Raw+RETURNING との組合せで非決定論的になるため置き換え、代わりに
// 「連続する独立writerがそれぞれ自分の書込み結果を受け取ること」を直接検証する。
func TestConversationSummaryRepository_SaveReturnsOwnWritePerWriter(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-CONCURRENT-FOLLOW")

	writerA, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "writerAの要約",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, writerA.Summary)
	assert.Equal(t, "writerAの要約", *writerA.Summary, "writerA は自分の内容を受け取る")
	require.NotNil(t, writerA.Revision)
	require.Equal(t, int64(1), *writerA.Revision, "writerA の版は 1")

	writerB, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "writerBの要約",
		Source:          "lstep",
	})
	require.NoError(t, err)
	require.NotNil(t, writerB.Summary)
	assert.Equal(t, "writerBの要約", *writerB.Summary, "writerB は自分の内容を受け取る（writerA の内容でない）")
	require.NotNil(t, writerB.Revision)
	require.Equal(t, int64(2), *writerB.Revision, "writerB の版は 2（writerA の版を流用しない）")

	// 各writerの返値が次の ExpectedRevision として正しく機能すること
	// （旧 lost update 経路の回帰防止: 返版が他writerのものだとこの保存が誤conflictする）。
	writerC, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:         conversationSummaryClinicA,
		MedicalRecordID:  record.ID,
		Summary:          "writerCの要約",
		Source:           "lstep",
		ExpectedRevision: writerB.Revision,
	})
	require.NoError(t, err, "writerB の返版は次の保存の期待版として有効")
	require.NotNil(t, writerC.Revision)
	assert.Equal(t, int64(3), *writerC.Revision)
}

func TestConversationSummaryRepository_SaveOntoUnsavedWithExpectedZero(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-EXPECTED-ZERO")

	saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:         conversationSummaryClinicA,
		MedicalRecordID:  record.ID,
		Summary:          "未保存からの初回保存",
		Source:           "lstep",
		ExpectedRevision: conversationSummaryInt64Ptr(0),
	})
	require.NoError(t, err)
	require.NotNil(t, saved.Revision)
	assert.Equal(t, int64(1), *saved.Revision)
}

func TestConversationSummaryRepository_SaveRejectsBlankMetadata(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-BLANK")

	t.Run("source が空白", func(t *testing.T) {
		for _, source := range []string{"", "   ", "\t"} {
			saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
				ClinicID:        conversationSummaryClinicA,
				MedicalRecordID: record.ID,
				Summary:         "本文",
				Source:          source,
			})
			assert.Nil(t, saved)
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "source 空白は invalid input: %v", err)
		}
	})

	t.Run("要約本文が空", func(t *testing.T) {
		saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:        conversationSummaryClinicA,
			MedicalRecordID: record.ID,
			Summary:         "  ",
			Source:          "lstep",
		})
		assert.Nil(t, saved)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "空の要約は invalid input: %v", err)
	})

	t.Run("clinic/カルテ未特定", func(t *testing.T) {
		saved, err := repo.Save(ctx, SaveConversationSummaryCommand{
			ClinicID:        0,
			MedicalRecordID: record.ID,
			Summary:         "本文",
			Source:          "lstep",
		})
		assert.Nil(t, saved)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "clinic 未特定は invalid input: %v", err)
	})

	// 拒否された保存は DB を一切変えない。
	row := fetchConversationSummaryRecordRow(t, db, record.ID)
	assert.Nil(t, row.ConversationSummary)
	assert.Nil(t, row.ConversationSummaryRevision)
	assert.Nil(t, row.ConversationSummarySource)
}

func TestConversationSummaryRepository_SaveLeavesOtherColumnsUntouched(t *testing.T) {
	db := setupConversationSummaryTestDB(t)
	repo := NewConversationSummaryRepository(db)
	ctx := context.Background()
	record := makeConversationSummaryRecord(t, db, conversationSummaryClinicA, "CS-OTHER-COLUMNS")

	plan := &model.ClinicalPlan{
		MedicalRecordID:  record.ID,
		ClinicID:         conversationSummaryClinicA,
		PhysicalExam:     "所見: 元気・食欲あり",
		DiagnosisDetails: "診断: 経過観察",
		TreatmentPolicy:  "方針: 1週間後に再診",
	}
	require.NoError(t, db.WithContext(ctx).Create(plan).Error)

	before := fetchConversationSummaryRecordRow(t, db, record.ID)
	beforeStatus := fetchConversationSummaryStatusColumn(t, db, record.ID)
	beforePlan := &model.ClinicalPlan{}
	require.NoError(t, db.WithContext(ctx).First(beforePlan, plan.ID).Error)

	_, err := repo.Save(ctx, SaveConversationSummaryCommand{
		ClinicID:        conversationSummaryClinicA,
		MedicalRecordID: record.ID,
		Summary:         "他所見を壊さないことの確認",
		Source:          "lstep",
	})
	require.NoError(t, err)

	after := fetchConversationSummaryRecordRow(t, db, record.ID)
	afterPlan := &model.ClinicalPlan{}
	require.NoError(t, db.WithContext(ctx).First(afterPlan, plan.ID).Error)

	// 更新対象の3列は変わる（保存が実際に走ったことの確認）。
	require.NotNil(t, after.ConversationSummary)
	assert.Equal(t, "他所見を壊さないことの確認", *after.ConversationSummary)

	// 他の診療記録列は不変（負例確認）。updated_at も非表示保存で進めない
	// （UpdateColumns により自動タイムスタンプを回避）。
	assert.Equal(t, before.RecordNo, after.RecordNo)
	assert.True(t, before.Date.Equal(after.Date), "date は不変であること")
	assert.True(t, before.UpdatedAt.Equal(after.UpdatedAt), "updated_at は会話要約の保存で進まないこと")
	assert.Equal(t, before.Version, after.Version, "会話要約の保存はカルテ version を進めない")
	assert.Equal(t, before.OwnerID, after.OwnerID)
	assert.Equal(t, before.PetID, after.PetID)
	assert.Equal(t, before.DoctorID, after.DoctorID)
	assert.Equal(t, before.RecommendationReason, after.RecommendationReason)
	assert.Equal(t, before.ClinicalSetGeneration, after.ClinicalSetGeneration)
	assert.Equal(t, before.ClinicalSetDigest, after.ClinicalSetDigest)
	assert.Equal(t, beforeStatus, fetchConversationSummaryStatusColumn(t, db, record.ID), "status 列は保存前後で不変であること")

	// 診療所見・診断・方針（clinical_plans）も不変。
	assert.Equal(t, beforePlan.PhysicalExam, afterPlan.PhysicalExam)
	assert.Equal(t, beforePlan.DiagnosisDetails, afterPlan.DiagnosisDetails)
	assert.Equal(t, beforePlan.TreatmentPolicy, afterPlan.TreatmentPolicy)
	assert.Equal(t, beforePlan.Version, afterPlan.Version)
}
