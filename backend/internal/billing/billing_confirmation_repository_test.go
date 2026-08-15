package billing

// billing_confirmation_repository_test.go — BillingConfirmationRepository の統合テスト
// （内部カバレッジ向上）。
//
// 対象: FindByMedicalRecordID / Create / Update
// 検証観点: 正常系、medical_records JOIN 経由の clinic_id 隔離、NotFound ラップ。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupBillingConfirmationTestDB は billing_confirmations と actor assignment の
// FK 先である staffs / staff_clinic_assignments を整備する
// （medical_records は core AutoMigrate 済み、clinic は assignment fixture が作成する）。
func setupBillingConfirmationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Staff{},
		&model.StaffClinicAssignment{},
		&model.BillingConfirmation{},
	))
	db.Exec("TRUNCATE TABLE billing_confirmations CASCADE")
	db.Exec("TRUNCATE TABLE staff_clinic_assignments CASCADE")
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

func assignBillingConfirmationStaff(
	t *testing.T,
	db *gorm.DB,
	staffID, clinicID uint64,
) *model.StaffClinicAssignment {
	t.Helper()
	seedBillingClinicForFK(t, db, clinicID)
	assignment := &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID}
	require.NoError(t, db.Create(assignment).Error)
	return assignment
}

func lockBillingConfirmationActor(
	db *gorm.DB,
	repo BillingConfirmationRepository,
	clinicID, staffID uint64,
) error {
	return testNewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		return repo.LockActiveStaffAssignment(txCtx, clinicID, staffID)
	})
}

func TestBillingConfirmationRepository_LockActiveStaffAssignment(t *testing.T) {
	db := setupBillingConfirmationTestDB(t)
	repo := NewBillingConfirmationRepository(db)
	const clinicID = uint64(1)

	validStaff := makeDoctor(t, db, clinicID, "valid assigned actor")
	assignBillingConfirmationStaff(t, db, validStaff.ID, clinicID)
	require.NoError(t, lockBillingConfirmationActor(db, repo, clinicID, validStaff.ID))

	unassignedStaff := makeDoctor(t, db, clinicID, "unassigned actor")
	err := lockBillingConfirmationActor(db, repo, clinicID, unassignedStaff.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))

	inactiveStaff := makeDoctor(t, db, clinicID, "inactive assigned actor")
	assignBillingConfirmationStaff(t, db, inactiveStaff.ID, clinicID)
	require.NoError(t, db.Model(&model.Staff{}).Where("id = ?", inactiveStaff.ID).Update("is_active", false).Error)
	err = lockBillingConfirmationActor(db, repo, clinicID, inactiveStaff.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))

	deletedAssignmentStaff := makeDoctor(t, db, clinicID, "deleted assignment actor")
	deletedAssignment := assignBillingConfirmationStaff(t, db, deletedAssignmentStaff.ID, clinicID)
	require.NoError(t, db.Delete(deletedAssignment).Error)
	err = lockBillingConfirmationActor(db, repo, clinicID, deletedAssignmentStaff.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))

	err = repo.LockActiveStaffAssignment(context.Background(), clinicID, validStaff.ID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code, "actor validation must never fall back to a non-transactional DB handle")
}

func TestBillingConfirmationService_RuntimeActorIsolation(t *testing.T) {
	db := setupBillingConfirmationTestDB(t)
	repo := NewBillingConfirmationRepository(db)
	const clinicID = uint64(1)

	actor := makeDoctor(t, db, clinicID, "runtime assigned actor")
	assignBillingConfirmationStaff(t, db, actor.ID, clinicID)
	medicalRecord := makeBillingConfirmationMedicalRecord(t, db, clinicID, "MR-BC-runtime-valid")

	lockCalls := 0
	medicalRecordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(ctx context.Context, gotClinicID, gotMedicalRecordID uint64) (*model.MedicalRecord, error) {
			require.NotNil(t, persistence.TxFromContext(ctx), "GetOrCreate parent validation must use the confirmation transaction")
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, medicalRecord.ID, gotMedicalRecordID)
			return medicalRecord, nil
		},
		lockByIDForUpdateFn: func(ctx context.Context, gotClinicID, gotMedicalRecordID uint64) (*model.MedicalRecord, error) {
			lockCalls++
			require.NotNil(t, persistence.TxFromContext(ctx), "medical-record and actor locks must share the write transaction")
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, medicalRecord.ID, gotMedicalRecordID)
			return medicalRecord, nil
		},
	}
	svc := NewBillingConfirmationService(repo, medicalRecordRepo, testNewTransactor(db))

	confirmed, err := svc.Confirm(context.Background(), clinicID, medicalRecord.ID, &ConfirmBillingConfirmationInput{
		ConfirmedBy: actor.ID,
		Memo:        "trusted actor",
	})
	require.NoError(t, err)
	require.NotNil(t, confirmed)
	assert.Equal(t, model.ConfirmationStatusConfirmed, confirmed.Status)
	require.NotNil(t, confirmed.ConfirmedBy)
	assert.Equal(t, actor.ID, *confirmed.ConfirmedBy)

	returned, err := svc.Return(context.Background(), clinicID, medicalRecord.ID, &ReturnBillingConfirmationInput{
		ReturnedBy:   actor.ID,
		ReturnReason: "runtime return",
	})
	require.NoError(t, err)
	require.NotNil(t, returned)
	assert.Equal(t, model.ConfirmationStatusReturned, returned.Status)
	require.NotNil(t, returned.ReturnedBy)
	assert.Equal(t, actor.ID, *returned.ReturnedBy)
	assert.Equal(t, 2, lockCalls)

	unassignedActor := makeDoctor(t, db, clinicID, "runtime unassigned actor")
	unassignedRecord := makeBillingConfirmationMedicalRecord(t, db, clinicID, "MR-BC-runtime-denied")
	deniedMedRecRepo := &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(ctx context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			require.NotNil(t, persistence.TxFromContext(ctx))
			return unassignedRecord, nil
		},
	}
	deniedSvc := NewBillingConfirmationService(repo, deniedMedRecRepo, testNewTransactor(db))

	denied, err := deniedSvc.Confirm(context.Background(), clinicID, unassignedRecord.ID, &ConfirmBillingConfirmationInput{
		ConfirmedBy: unassignedActor.ID,
	})

	require.Error(t, err)
	assert.Nil(t, denied)
	assert.True(t, errors.Is(err, apperrors.ErrForbidden))
	var count int64
	require.NoError(t, db.Model(&model.BillingConfirmation{}).
		Where("medical_record_id = ?", unassignedRecord.ID).
		Count(&count).Error)
	assert.Zero(t, count, "actor validation must fail before GetOrCreate persists a pending row")
}

type failNthBillingConfirmationFindRepository struct {
	BillingConfirmationRepository
	failAt int
	calls  int
}

func (r *failNthBillingConfirmationFindRepository) FindByMedicalRecordID(
	ctx context.Context,
	clinicID, medicalRecordID uint64,
) (*model.BillingConfirmation, error) {
	r.calls++
	if r.calls == r.failAt {
		return nil, errors.New("injected confirmation refetch failure")
	}
	return r.BillingConfirmationRepository.FindByMedicalRecordID(
		ctx,
		clinicID,
		medicalRecordID,
	)
}

func TestBillingConfirmationService_RefetchFailureRollsBackWrite(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("confirm create and update are both rolled back", func(t *testing.T) {
		db := setupBillingConfirmationTestDB(t)
		baseRepo := NewBillingConfirmationRepository(db)
		repo := &failNthBillingConfirmationFindRepository{
			BillingConfirmationRepository: baseRepo,
			failAt:                        2,
		}
		actor := makeDoctor(t, db, clinicID, "confirm rollback actor")
		assignBillingConfirmationStaff(t, db, actor.ID, clinicID)
		medicalRecord := makeBillingConfirmationMedicalRecord(
			t,
			db,
			clinicID,
			"MR-BC-refetch-confirm",
		)
		medicalRecordRepo := &mockMedicalRecordRepository{
			findByIDFn: func(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
				return medicalRecord, nil
			},
			lockByIDForUpdateFn: func(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
				return medicalRecord, nil
			},
		}
		svc := NewBillingConfirmationService(
			repo,
			medicalRecordRepo,
			testNewTransactor(db),
		)

		got, err := svc.Confirm(
			context.Background(),
			clinicID,
			medicalRecord.ID,
			&ConfirmBillingConfirmationInput{ConfirmedBy: actor.ID},
		)

		require.Error(t, err)
		assert.Nil(t, got)
		var count int64
		require.NoError(t, db.Model(&model.BillingConfirmation{}).
			Where("medical_record_id = ?", medicalRecord.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("return update is rolled back", func(t *testing.T) {
		db := setupBillingConfirmationTestDB(t)
		baseRepo := NewBillingConfirmationRepository(db)
		actor := makeDoctor(t, db, clinicID, "return rollback actor")
		assignBillingConfirmationStaff(t, db, actor.ID, clinicID)
		medicalRecord := makeBillingConfirmationMedicalRecord(
			t,
			db,
			clinicID,
			"MR-BC-refetch-return",
		)
		confirmedBy := actor.ID
		confirmation := &model.BillingConfirmation{
			MedicalRecordID: medicalRecord.ID,
			Status:          model.ConfirmationStatusConfirmed,
			ConfirmedBy:     &confirmedBy,
		}
		require.NoError(t, baseRepo.Create(context.Background(), confirmation))
		repo := &failNthBillingConfirmationFindRepository{
			BillingConfirmationRepository: baseRepo,
			failAt:                        2,
		}
		medicalRecordRepo := &mockMedicalRecordRepository{
			lockByIDForUpdateFn: func(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
				return medicalRecord, nil
			},
		}
		svc := NewBillingConfirmationService(
			repo,
			medicalRecordRepo,
			testNewTransactor(db),
		)

		got, err := svc.Return(
			context.Background(),
			clinicID,
			medicalRecord.ID,
			&ReturnBillingConfirmationInput{
				ReturnedBy:   actor.ID,
				ReturnReason: "injected rollback",
			},
		)

		require.Error(t, err)
		assert.Nil(t, got)
		persisted, findErr := baseRepo.FindByMedicalRecordID(
			context.Background(),
			clinicID,
			medicalRecord.ID,
		)
		require.NoError(t, findErr)
		assert.Equal(t, model.ConfirmationStatusConfirmed, persisted.Status)
		assert.Nil(t, persisted.ReturnedBy)
	})
}
