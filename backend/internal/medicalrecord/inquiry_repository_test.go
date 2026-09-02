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
		_, err := repo.SaveByMedicalRecordID(ctx, clinicB, &model.Inquiry{MedicalRecordID: mrA.ID, ChiefComplaint: "嘔吐"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("medical_record が存在しない場合 NotFound", func(t *testing.T) {
		_, err := repo.SaveByMedicalRecordID(ctx, clinicA, &model.Inquiry{MedicalRecordID: 999999, ChiefComplaint: "嘔吐"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	var savedID uint64
	t.Run("新規作成される", func(t *testing.T) {
		got, err := repo.SaveByMedicalRecordID(ctx, clinicA, &model.Inquiry{MedicalRecordID: mrA.ID, ChiefComplaint: "嘔吐", Notes: "初回"})
		require.NoError(t, err)
		assert.Equal(t, mrA.ID, got.MedicalRecordID)
		assert.Equal(t, "嘔吐", got.ChiefComplaint)
		assert.Equal(t, "初回", got.Notes)
		assert.NotZero(t, got.ID)
		savedID = got.ID
	})

	t.Run("既存レコードは新規作成せず更新される（同一 medical_record_id は1件のみ）", func(t *testing.T) {
		got, err := repo.SaveByMedicalRecordID(ctx, clinicA, &model.Inquiry{MedicalRecordID: mrA.ID, ChiefComplaint: "下痢", Notes: "再診"})
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

		_, err := repo.SaveByMedicalRecordID(ctx, clinicA, &model.Inquiry{MedicalRecordID: mrFinalized.ID, ChiefComplaint: "嘔吐"})
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの問診保存は Conflict(409) であるべき: %v", err)

		var count int64
		require.NoError(t, db.WithContext(ctx).Model(&model.Inquiry{}).Where("medical_record_id = ?", mrFinalized.ID).Count(&count).Error)
		assert.Zero(t, count, "確定済みカルテには問診データが残ってはならない")
	})
}
