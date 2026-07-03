package repository

// billing_confirmation_repository_test.go — BillingConfirmationRepository の統合テスト
// （内部カバレッジ向上）。
//
// 対象: FindByMedicalRecordID / Create / Update
// 検証観点: 正常系、medical_records JOIN 経由の clinic_id 隔離、NotFound ラップ。

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

// setupBillingConfirmationTestDB は billing_confirmations と、その FK 先である staffs を整備する
// （medical_records は core AutoMigrate 済み）。
func setupBillingConfirmationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Staff{}, &model.BillingConfirmation{}))
	db.Exec("TRUNCATE TABLE billing_confirmations CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	return db
}

func makeBillingConfirmationMedicalRecord(t *testing.T, db *gorm.DB, clinicID uint64, recordNo string) *model.MedicalRecord {
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

// TestBillingConfirmationRepository_Create_FindByMedicalRecordID は作成と取得（clinic_id 隔離・
// NotFound）を検証する。
func TestBillingConfirmationRepository_Create_FindByMedicalRecordID(t *testing.T) {
	db := setupBillingConfirmationTestDB(t)
	repo := NewBillingConfirmationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeBillingConfirmationMedicalRecord(t, db, clinicA, "MR-BC-A-001")

	review := &model.BillingConfirmation{
		MedicalRecordID: mrA.ID,
		Status:          model.ConfirmationStatusPending,
	}

	t.Run("Create で新規作成される", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, review))
		assert.NotZero(t, review.ID)
	})

	t.Run("自クリニックの medical_record_id で取得できる", func(t *testing.T) {
		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, mrA.ID, got.MedicalRecordID)
		assert.Equal(t, model.ConfirmationStatusPending, got.Status)
	})

	t.Run("別クリニックからの取得は NotFound（medical_records JOIN 経由の隔離）", func(t *testing.T) {
		_, err := repo.FindByMedicalRecordID(ctx, clinicB, mrA.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない medical_record_id は NotFound", func(t *testing.T) {
		_, err := repo.FindByMedicalRecordID(ctx, clinicA, 9999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestBillingConfirmationRepository_Update は部分更新（サブクエリによる clinic_id 隔離・
// NotFound）を検証する。
func TestBillingConfirmationRepository_Update(t *testing.T) {
	db := setupBillingConfirmationTestDB(t)
	repo := NewBillingConfirmationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	mrA := makeBillingConfirmationMedicalRecord(t, db, clinicA, "MR-BC-A-002")
	review := &model.BillingConfirmation{MedicalRecordID: mrA.ID, Status: model.ConfirmationStatusPending}
	require.NoError(t, repo.Create(ctx, review))
	// confirmed_by は staffs への FK。存在しない staff_id を渡すと FK 制約違反になるため、
	// 実在する staff 行を作成してその ID を使う。
	confirmer := makeDoctor(t, db, clinicA, "確定担当医")

	t.Run("正常系: status/confirmed_by が更新される", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, clinicA, review.ID, map[string]any{
			"status":       model.ConfirmationStatusConfirmed,
			"confirmed_by": confirmer.ID,
		}))

		got, err := repo.FindByMedicalRecordID(ctx, clinicA, mrA.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ConfirmationStatusConfirmed, got.Status)
		require.NotNil(t, got.ConfirmedBy)
		assert.Equal(t, confirmer.ID, *got.ConfirmedBy)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicB, review.ID, map[string]any{"memo": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの更新は NotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 9999999, map[string]any{"memo": "存在しない"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
