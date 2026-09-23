package medicalrecord

// inquiry_repository_test.go — InquiryRepository の統合テスト（内部カバレッジ向上）。
//
// 対象: SaveByMedicalRecordID / CountByChiefComplaintTypeID
// 検証観点: 正常系（新規作成/既存更新）、clinic_id 隔離（medical_records JOIN 経由）、NotFound ラップ。

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

// setupInquiryTestDB は inquiries テーブルを整備する。
// medical_records は testdb.SetupTestDB のコア AutoMigrate で既に用意済みのため再定義しない。
// ChiefComplaintType は inquiries.chief_complaint_type_id の FK 参照先のため、
// (存在しないIDを直接書き込むとFK違反になる) AutoMigrate 対象に含める。
func setupInquiryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.Inquiry{}, &model.ChiefComplaintType{}))
	db.Exec("TRUNCATE TABLE inquiries CASCADE")
	db.Exec("TRUNCATE TABLE chief_complaint_types CASCADE")
	return db
}

// makeInquiryMedicalRecord はテスト用の MedicalRecord を作成して返す（pet/owner の紐付け無し）。
func makeInquiryMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
	t.Helper()
	mr := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: recordNo,
		Date:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(mr).Error)
	return mr
}

func TestInquiryRepository_SaveByMedicalRecordID(t *testing.T) {
	db := setupInquiryTestDB(t)
	repo := NewInquiryRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeInquiryMedicalRecord(t, db, clinicA, "MR-A-001")

	t.Run("medical_record が別クリニックの場合 NotFound", func(t *testing.T) {
		_, err := repo.SaveByMedicalRecordID(ctx, clinicB, InquiryUpsertFields{MedicalRecordID: mrA.ID, ChiefComplaint: strPtr("嘔吐")})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("medical_record が存在しない場合 NotFound", func(t *testing.T) {
		_, err := repo.SaveByMedicalRecordID(ctx, clinicA, InquiryUpsertFields{MedicalRecordID: 999999, ChiefComplaint: strPtr("嘔吐")})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	var savedID uint64
	t.Run("新規作成される", func(t *testing.T) {
		got, err := repo.SaveByMedicalRecordID(ctx, clinicA, InquiryUpsertFields{MedicalRecordID: mrA.ID, ChiefComplaint: strPtr("嘔吐"), Notes: strPtr("初回")})
		require.NoError(t, err)
		assert.Equal(t, mrA.ID, got.MedicalRecordID)
		assert.Equal(t, "嘔吐", got.ChiefComplaint)
		assert.Equal(t, "初回", got.Notes)
		assert.NotZero(t, got.ID)
		savedID = got.ID
	})

	t.Run("既存レコードは新規作成せず更新される（同一 medical_record_id は1件のみ）", func(t *testing.T) {
		got, err := repo.SaveByMedicalRecordID(ctx, clinicA, InquiryUpsertFields{MedicalRecordID: mrA.ID, ChiefComplaint: strPtr("下痢"), Notes: strPtr("再診")})
		require.NoError(t, err)
		assert.Equal(t, savedID, got.ID, "同じ medical_record_id に対して同一レコードが更新される")
		assert.Equal(t, "下痢", got.ChiefComplaint)
		assert.Equal(t, "再診", got.Notes)

		var count int64
		require.NoError(t, db.WithContext(ctx).Model(&model.Inquiry{}).Where("medical_record_id = ?", mrA.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count, "重複作成されない")
	})

	t.Run("medical_record が確定済みの場合 Conflict（MRC-12: 空 inquiry 行を残さない）", func(t *testing.T) {
		mrFinalized := &model.MedicalRecord{
			ClinicID: clinicA,
			RecordNo: "MR-A-FINALIZED",
			Date:     time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			Status:   model.MedicalRecordStatusFinalized,
		}
		require.NoError(t, db.WithContext(ctx).Create(mrFinalized).Error)

		_, err := repo.SaveByMedicalRecordID(ctx, clinicA, InquiryUpsertFields{MedicalRecordID: mrFinalized.ID, ChiefComplaint: strPtr("嘔吐")})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの問診保存は Conflict(409) であるべき: %v", err)

		var count int64
		require.NoError(t, db.WithContext(ctx).Model(&model.Inquiry{}).Where("medical_record_id = ?", mrFinalized.ID).Count(&count).Error)
		assert.Zero(t, count, "確定済みカルテには問診データが残ってはならない")
	})
}

// TestInquiryRepository_SaveByMedicalRecordID_ChiefComplaintTypeNullPersistence は
// SLACK-COMPLAINT (EMR-87) の残存 UNKNOWN を実 DB で閉じる回帰テスト。
// SaveByMedicalRecordID は送信フィールドのみ map[string]any に載せる:
//   - ChiefComplaintTypeID = &nil（明示 JSON null）なら列を SQL NULL へ書き込む（意図的解除）
//   - ChiefComplaintTypeID = nil（未送信）なら列に触れず既存値を保持する
//   - 未送信の chief_complaint / notes は既存値を保持する（区分のみ保存で本文を消さない）
func TestInquiryRepository_SaveByMedicalRecordID_ChiefComplaintTypeNullPersistence(t *testing.T) {
	db := setupInquiryTestDB(t)
	repo := NewInquiryRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	mr := makeInquiryMedicalRecord(t, db, clinicID, "MR-NULL-001")
	ctype := makeChiefComplaintType(t, db, clinicID, "嘔吐")

	// clearTypeID は明示的な「区分なし」(&nil) を表す。
	var clearTypeID *uint64
	// persistedTypeID は DB 列を直接読んで返す（返却 struct ではなく永続化値を検証する）。
	// COALESCE で NULL を 0 へ写像する（chief_complaint_type_id の実 ID は 1 始まりのため 0 は NULL のみを意味する）。
	persistedTypeID := func(t *testing.T) uint64 {
		t.Helper()
		var id uint64
		require.NoError(t, db.WithContext(ctx).
			Raw("SELECT COALESCE(chief_complaint_type_id, 0) FROM inquiries WHERE medical_record_id = ?", mr.ID).
			Row().Scan(&id))
		return id
	}

	t.Run("区分未選択のまま保存すると列は NULL で主訴本文は残る", func(t *testing.T) {
		got, err := repo.SaveByMedicalRecordID(ctx, clinicID, InquiryUpsertFields{
			MedicalRecordID: mr.ID,
			ChiefComplaint:  ptr("元気がない"),
		})
		require.NoError(t, err)
		assert.Nil(t, got.ChiefComplaintTypeID)
		assert.Equal(t, "元気がない", got.ChiefComplaint)
		assert.Zero(t, persistedTypeID(t), "未選択保存は chief_complaint_type_id を NULL で永続化する")
	})

	t.Run("既存区分を明示 null (&nil) で再保存すると列が NULL に戻り本文は保持される", func(t *testing.T) {
		setType := &ctype.ID
		got, err := repo.SaveByMedicalRecordID(ctx, clinicID, InquiryUpsertFields{
			MedicalRecordID:      mr.ID,
			ChiefComplaintTypeID: &setType,
			ChiefComplaint:       ptr("嘔吐気味"),
		})
		require.NoError(t, err)
		require.NotNil(t, got.ChiefComplaintTypeID)
		assert.Equal(t, ctype.ID, *got.ChiefComplaintTypeID)
		assert.Equal(t, ctype.ID, persistedTypeID(t))

		// 区分のみを解除する保存（chief_complaint は未送信）: 列は NULL、本文は保持。
		got, err = repo.SaveByMedicalRecordID(ctx, clinicID, InquiryUpsertFields{
			MedicalRecordID:      mr.ID,
			ChiefComplaintTypeID: &clearTypeID,
		})
		require.NoError(t, err)
		assert.Nil(t, got.ChiefComplaintTypeID)
		assert.Zero(t, persistedTypeID(t), "明示 null で再保存したら DB 列も NULL になる")
		assert.Equal(t, "嘔吐気味", got.ChiefComplaint, "区分解除のみの保存で主訴本文を失わない")
	})

	t.Run("chief_complaint_type_id 未送信は既存区分を保持する（PATCH 部分更新）", func(t *testing.T) {
		setType := &ctype.ID
		_, err := repo.SaveByMedicalRecordID(ctx, clinicID, InquiryUpsertFields{
			MedicalRecordID:      mr.ID,
			ChiefComplaintTypeID: &setType,
		})
		require.NoError(t, err)
		require.Equal(t, ctype.ID, persistedTypeID(t))

		// 本文だけの保存（chief_complaint_type_id 未送信）: 既存区分を消さない。
		got, err := repo.SaveByMedicalRecordID(ctx, clinicID, InquiryUpsertFields{
			MedicalRecordID: mr.ID,
			ChiefComplaint:  ptr("再診テキスト"),
		})
		require.NoError(t, err)
		require.NotNil(t, got.ChiefComplaintTypeID)
		assert.Equal(t, ctype.ID, *got.ChiefComplaintTypeID)
		assert.Equal(t, ctype.ID, persistedTypeID(t), "未送信フィールドは既存値を保持する")
		assert.Equal(t, "再診テキスト", got.ChiefComplaint)
	})
}
