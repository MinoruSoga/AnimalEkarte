package staff_test

// repository_test.go — Repository の統合テスト（実 Postgres テスト DB）。
// スタッフ-クリニック中間テーブルの CRUD と GORM SoftDelete スコープの自動適用
// （コメントに記載された前提）を実際の DB 動作で検証する。
//
// makeDoctor / seedClinicsForFK は親 repository パッケージの
// isolation_test_helpers_test.go / staff_preload_clinic_isolation_test.go 内の同名ヘルパーの複製
// （BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	. "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var New = NewStaffClinicAssignmentRepository

// setupStaffClinicAssignmentRepositoryTestDB は repository.go のテスト用に
// FK 親（companies/clinics）と staffs/staff_clinic_assignments を整備する。
func setupStaffClinicAssignmentRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
	))
	// model.StaffClinicAssignment has no `uniqueIndex` gorm tag, so AutoMigrate does not recreate
	// the composite UNIQUE(staff_id, clinic_id) constraint that migrations/001_init.sql defines
	// (uk_staff_clinic). Recreate a full unique index explicitly; a partial active-row index would
	// hide the production behavior where a soft-deleted pair must be restored instead of inserted.
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs CASCADE").Error)
	require.NoError(t, db.Exec("DROP INDEX IF EXISTS ux_test_staff_clinic_assignment").Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX ux_test_staff_clinic_assignment
		ON staff_clinic_assignments (staff_id, clinic_id)`).Error)
	return db
}

func TestStaffClinicAssignmentRepository_Create_HappyPath(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "新規所属スタッフ")

	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
	require.NoError(t, repo.Create(ctx, assignment))
	require.NotZero(t, assignment.ID)

	found, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, clinicID, found[0].ClinicID)
	assert.True(t, found[0].IsMain)
}

func TestStaffClinicAssignmentRepository_FindByStaffID_ReturnsAllClinicsForMultiAssignedStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "多医院所属スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}))
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}))

	found, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestStaffClinicAssignmentRepository_FindByStaffAndClinicReturnsOnlyRequestedClinic(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "医院スコープ所属スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicA,
		IsMain:   true,
	}))
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicB,
	}))

	found, err := repo.FindByStaffAndClinic(ctx, staff.ID, clinicB)

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, staff.ID, found.StaffID)
	assert.Equal(t, clinicB, found.ClinicID)
}

func TestStaffClinicAssignmentRepository_FindByStaffID_EmptyForUnassignedStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "未所属スタッフ")

	found, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestStaffClinicAssignmentRepository_FindByStaffID_ExcludesSoftDeleted(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "一部脱退スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA}))
	assignmentB := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, repo.Create(ctx, assignmentB))
	// clinicB への所属をソフトデリート
	require.NoError(t, db.WithContext(ctx).Delete(&model.StaffClinicAssignment{}, assignmentB.ID).Error)

	found, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	require.Len(t, found, 1, "ソフトデリート済み所属は GORM SoftDelete スコープで自動除外される")
	assert.Equal(t, clinicA, found[0].ClinicID)

	// 行自体は DB に残っている
	var rawCount int64
	db.Unscoped().Model(&model.StaffClinicAssignment{}).Where("id = ?", assignmentB.ID).Count(&rawCount)
	assert.Equal(t, int64(1), rawCount)
}

func TestStaffClinicAssignmentRepository_CountByStaffAndClinic_ReturnsOneWhenAssigned(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "カウント確認スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}))

	count, err := repo.CountByStaffAndClinic(ctx, staff.ID, clinicID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestStaffClinicAssignmentRepository_CountByStaffAndClinic_ZeroWhenNotAssigned(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "未所属確認スタッフ")

	count, err := repo.CountByStaffAndClinic(ctx, staff.ID, clinicB)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestStaffClinicAssignmentRepository_CountByStaffAndClinic_ExcludesSoftDeleted(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "脱退確認スタッフ")
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}
	require.NoError(t, repo.Create(ctx, assignment))
	require.NoError(t, db.WithContext(ctx).Delete(&model.StaffClinicAssignment{}, assignment.ID).Error)

	count, err := repo.CountByStaffAndClinic(ctx, staff.ID, clinicID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "ソフトデリート済み所属はカウントされない")
}

func TestStaffClinicAssignmentRepository_LockActiveByStaff_ReturnsOnlyActiveAssignmentsInClinicOrder(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)
	seedClinicsForFK(t, db, clinicA, clinicB, clinicC)

	staff := makeDoctor(t, db, clinicA, "全所属ロック対象スタッフ")
	assignmentC := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicC}
	assignmentA := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}
	assignmentB := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, repo.Create(ctx, assignmentC))
	require.NoError(t, repo.Create(ctx, assignmentA))
	require.NoError(t, repo.Create(ctx, assignmentB))
	require.NoError(t, db.WithContext(ctx).Delete(assignmentB).Error)

	require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		assignments, err := repo.LockActiveByStaff(persistence.WithTxValue(ctx, tx), staff.ID)
		require.NoError(t, err)
		require.Len(t, assignments, 2)
		assert.Equal(t, []uint64{clinicA, clinicC}, []uint64{
			assignments[0].ClinicID,
			assignments[1].ClinicID,
		})
		assert.Equal(t, []uint64{assignmentA.ID, assignmentC.ID}, []uint64{
			assignments[0].ID,
			assignments[1].ID,
		})
		return nil
	}))
}

func TestStaffClinicAssignmentRepository_LockActiveByStaff_SourceContract(t *testing.T) {
	source, err := os.ReadFile("staff_clinic_assignment_repository.go")
	require.NoError(t, err)

	const methodSignature = "func (r *staffClinicAssignmentRepository) LockActiveByStaff("
	const nextMethodSignature = "func (r *staffClinicAssignmentRepository) LockActiveByStaffAndClinic("
	methodStart := bytes.Index(source, []byte(methodSignature))
	require.NotEqual(t, -1, methodStart)
	methodEndOffset := bytes.Index(source[methodStart:], []byte(nextMethodSignature))
	require.NotEqual(t, -1, methodEndOffset)
	methodSource := string(source)[methodStart : methodStart+methodEndOffset]

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, `clause.Locking{Strength: "UPDATE"}`)
	assert.Contains(t, methodSource, `Where("staff_id = ?", staffID)`)
	assert.Contains(t, methodSource, `Order("clinic_id ASC, id ASC")`)
}

func TestStaffClinicAssignmentRepository_LockActiveByStaff_RequiresAmbientTransaction(t *testing.T) {
	repo := New(nil)

	assignments, err := repo.LockActiveByStaff(context.Background(), 1)

	assert.Nil(t, assignments)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestStaffClinicAssignmentRepository_LockActiveByStaff_HoldsUpdateLocksUntilTransactionEnds(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "全所属更新ロックスタッフ")
	assignment := &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   true,
	}
	require.NoError(t, repo.Create(ctx, assignment))

	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if _, err := repo.LockActiveByStaff(
				persistence.WithTxValue(ctx, tx),
				staff.ID,
			); err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-locked:
	case <-ctx.Done():
		t.Fatal("assignment UPDATE lock was not acquired")
	}

	contenderErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL lock_timeout = '100ms'").Error; err != nil {
			return err
		}
		return tx.Model(&model.StaffClinicAssignment{}).
			Where("id = ?", assignment.ID).
			Update("is_main", false).Error
	})
	close(release)
	require.NoError(t, <-holderDone)
	require.Error(t, contenderErr, "concurrent assignment update must time out behind FOR UPDATE")

	var unchanged model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).First(&unchanged, assignment.ID).Error)
	assert.True(t, unchanged.IsMain)
}

func TestStaffClinicAssignmentRepository_LockActiveByStaffAndClinic_ReturnsTargetActiveAssignment(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "所属ロック対象スタッフ")
	assignmentA := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}
	assignmentB := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, repo.Create(ctx, assignmentA))
	require.NoError(t, repo.Create(ctx, assignmentB))

	require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := repo.LockActiveByStaffAndClinic(
			persistence.WithTxValue(ctx, tx),
			staff.ID,
			clinicA,
		)
		require.NoError(t, err)
		require.NotNil(t, locked)
		assert.Equal(t, assignmentA.ID, locked.ID)
		assert.Equal(t, clinicA, locked.ClinicID)
		assert.NotEqual(t, assignmentB.ID, locked.ID)
		return nil
	}))
}

func TestStaffClinicAssignmentRepository_LockActiveByStaffAndClinic_RejectsOtherClinicAndSoftDeleted(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "所属ロック拒否スタッフ")
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}
	require.NoError(t, repo.Create(ctx, assignment))

	t.Run("other clinic", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			locked, err := repo.LockActiveByStaffAndClinic(
				persistence.WithTxValue(ctx, tx),
				staff.ID,
				clinicB,
			)
			assert.Nil(t, locked)
			assert.True(t, apperrors.IsNotFound(err))
			return nil
		}))
	})

	require.NoError(t, db.WithContext(ctx).Delete(assignment).Error)
	t.Run("soft deleted assignment", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			locked, err := repo.LockActiveByStaffAndClinic(
				persistence.WithTxValue(ctx, tx),
				staff.ID,
				clinicA,
			)
			assert.Nil(t, locked)
			assert.True(t, apperrors.IsNotFound(err))
			return nil
		}))
	})
}

func TestStaffClinicAssignmentRepository_LockActiveByStaffAndClinic_HoldsShareLockUntilTransactionEnds(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "所属共有ロックスタッフ")
	assignment := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID, IsMain: true}
	require.NoError(t, repo.Create(ctx, assignment))

	locked := make(chan struct{})
	release := make(chan struct{})
	holderDone := make(chan error, 1)
	go func() {
		holderDone <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if _, err := repo.LockActiveByStaffAndClinic(
				persistence.WithTxValue(ctx, tx),
				staff.ID,
				clinicID,
			); err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-locked:
	case <-ctx.Done():
		t.Fatal("assignment lock was not acquired")
	}

	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- db.WithContext(ctx).Delete(assignment).Error
	}()

	var deleteErr error
	completedBeforeRelease := false
	select {
	case deleteErr = <-deleteDone:
		completedBeforeRelease = true
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-holderDone)
	if !completedBeforeRelease {
		select {
		case deleteErr = <-deleteDone:
		case <-ctx.Done():
			t.Fatal("assignment deletion did not resume after the lock transaction ended")
		}
	}
	require.NoError(t, deleteErr)
	assert.False(t, completedBeforeRelease, "soft delete must wait for the assignment SHARE lock")
}

func TestStaffClinicAssignmentRepository_LockActiveByStaffAndClinic_SourceContract(t *testing.T) {
	source, err := os.ReadFile("staff_clinic_assignment_repository.go")
	require.NoError(t, err)

	const methodSignature = "func (r *staffClinicAssignmentRepository) LockActiveByStaffAndClinic("
	const nextMethodSignature = "func (r *staffClinicAssignmentRepository) RestoreOrCreate("
	methodStart := bytes.Index(source, []byte(methodSignature))
	require.NotEqual(t, -1, methodStart)
	methodEndOffset := bytes.Index(source[methodStart:], []byte(nextMethodSignature))
	require.NotEqual(t, -1, methodEndOffset)
	methodSource := string(source)[methodStart : methodStart+methodEndOffset]

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, `clause.Locking{Strength: "SHARE"}`)
	assert.Contains(t, methodSource, `Where("staff_id = ? AND clinic_id = ?", staffID, clinicID)`)
}

func TestStaffClinicAssignmentRepository_LockActiveByStaffAndClinic_RequiresAmbientTransaction(t *testing.T) {
	repo := New(nil)

	assignment, err := repo.LockActiveByStaffAndClinic(context.Background(), 1, 1)

	assert.Nil(t, assignment)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_RestoresSoftDeletedPair(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "所属復元スタッフ")
	original := &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   false,
	}
	require.NoError(t, repo.Create(ctx, original))
	originalID := original.ID
	require.NoError(t, db.WithContext(ctx).Delete(original).Error)

	require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return repo.RestoreOrCreate(
			persistence.WithTxValue(ctx, tx),
			&model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: clinicID,
				IsMain:   true,
			},
		)
	}))

	var restored model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
		First(&restored).Error)
	assert.Equal(t, originalID, restored.ID, "full unique pair must reuse the soft-deleted row")
	assert.False(t, restored.DeletedAt.Valid)
	assert.True(t, restored.IsMain)

	var pairCount int64
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
		Count(&pairCount).Error)
	assert.Equal(t, int64(1), pairCount)
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_ReusesRowsForNoOpReplacement(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "同一所属差替スタッフ")
	assignmentA := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}
	assignmentB := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, repo.Create(ctx, assignmentA))
	require.NoError(t, repo.Create(ctx, assignmentB))
	originalIDs := []uint64{assignmentA.ID, assignmentB.ID}

	require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		locked, err := repo.LockActiveByStaff(txCtx, staff.ID)
		if err != nil {
			return err
		}
		if len(locked) != 2 {
			return fmt.Errorf("lock active assignments: got %d rows, want 2", len(locked))
		}
		if err := repo.Delete(txCtx, staff.ID); err != nil {
			return err
		}
		if err := repo.RestoreOrCreate(txCtx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicA,
			IsMain:   true,
		}); err != nil {
			return err
		}
		return repo.RestoreOrCreate(txCtx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicB,
			IsMain:   false,
		})
	}))

	active, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	require.Len(t, active, 2)
	assert.ElementsMatch(t, originalIDs, []uint64{active[0].ID, active[1].ID})

	var allRows int64
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ?", staff.ID).
		Count(&allRows).Error)
	assert.Equal(t, int64(2), allRows, "no-op replacement must not insert duplicate pairs")
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_ReusesOverlapAndCreatesNewPair(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB, clinicC = uint64(1), uint64(2), uint64(3)
	seedClinicsForFK(t, db, clinicA, clinicB, clinicC)

	staff := makeDoctor(t, db, clinicA, "一部重複所属差替スタッフ")
	assignmentA := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA, IsMain: true}
	assignmentB := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}
	require.NoError(t, repo.Create(ctx, assignmentA))
	require.NoError(t, repo.Create(ctx, assignmentB))

	require.NoError(t, db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		locked, err := repo.LockActiveByStaff(txCtx, staff.ID)
		if err != nil {
			return err
		}
		if len(locked) != 2 {
			return fmt.Errorf("lock active assignments: got %d rows, want 2", len(locked))
		}
		if err := repo.Delete(txCtx, staff.ID); err != nil {
			return err
		}
		if err := repo.RestoreOrCreate(txCtx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicB,
			IsMain:   true,
		}); err != nil {
			return err
		}
		return repo.RestoreOrCreate(txCtx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: clinicC,
			IsMain:   false,
		})
	}))

	var all []model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ?", staff.ID).
		Order("clinic_id ASC").
		Find(&all).Error)
	require.Len(t, all, 3)
	assert.Equal(t, assignmentA.ID, all[0].ID)
	assert.True(t, all[0].DeletedAt.Valid, "removed clinic must remain soft-deleted")
	assert.Equal(t, assignmentB.ID, all[1].ID, "overlapping clinic must reuse its row")
	assert.False(t, all[1].DeletedAt.Valid)
	assert.True(t, all[1].IsMain)
	assert.Equal(t, clinicC, all[2].ClinicID)
	assert.False(t, all[2].DeletedAt.Valid)
	assert.False(t, all[2].IsMain)
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_SourceContract(t *testing.T) {
	source, err := os.ReadFile("staff_clinic_assignment_repository.go")
	require.NoError(t, err)

	const methodSignature = "func (r *staffClinicAssignmentRepository) RestoreOrCreate("
	const nextMethodSignature = "func (r *staffClinicAssignmentRepository) Create("
	methodStart := bytes.Index(source, []byte(methodSignature))
	require.NotEqual(t, -1, methodStart)
	methodEndOffset := bytes.Index(source[methodStart:], []byte(nextMethodSignature))
	require.NotEqual(t, -1, methodEndOffset)
	methodSource := string(source)[methodStart : methodStart+methodEndOffset]

	assert.Contains(t, methodSource, "persistence.TxFromContext(ctx)")
	assert.Contains(t, methodSource, "persistence.DBOrTx(ctx, r.db)")
	assert.Contains(t, methodSource, `Name: "staff_id"`)
	assert.Contains(t, methodSource, `Name: "clinic_id"`)
	assert.Contains(t, methodSource, `"deleted_at"`)
	assert.Contains(t, methodSource, `"is_main"`)
	assert.Contains(t, methodSource, "apperrors.FromGORM")
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_RequiresAmbientTransaction(t *testing.T) {
	repo := New(nil)

	err := repo.RestoreOrCreate(context.Background(), &model.StaffClinicAssignment{
		StaffID:  1,
		ClinicID: 1,
		IsMain:   true,
	})

	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestStaffClinicAssignmentRepository_RestoreOrCreate_ParticipatesInAmbientRollback(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "所属復元ロールバックスタッフ")
	original := &model.StaffClinicAssignment{
		StaffID:  staff.ID,
		ClinicID: clinicID,
		IsMain:   false,
	}
	require.NoError(t, repo.Create(ctx, original))
	require.NoError(t, db.WithContext(ctx).Delete(original).Error)

	rollbackErr := errors.New("force assignment restore rollback")
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := repo.RestoreOrCreate(
			persistence.WithTxValue(ctx, tx),
			&model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: clinicID,
				IsMain:   true,
			},
		); err != nil {
			return err
		}
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)

	var persisted model.StaffClinicAssignment
	require.NoError(t, db.WithContext(ctx).Unscoped().
		Where("staff_id = ? AND clinic_id = ?", staff.ID, clinicID).
		First(&persisted).Error)
	assert.True(t, persisted.DeletedAt.Valid, "restore must roll back with the ambient transaction")
	assert.False(t, persisted.IsMain)
}

func TestStaffClinicAssignmentRepository_Delete_RemovesAllAssignmentsForStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	seedClinicsForFK(t, db, clinicA, clinicB)

	staff := makeDoctor(t, db, clinicA, "全所属削除対象スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicA}))
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicB}))

	require.NoError(t, repo.Delete(ctx, staff.ID))

	found, err := repo.FindByStaffID(ctx, staff.ID)
	require.NoError(t, err)
	assert.Empty(t, found, "Delete は staff の全所属を削除する")
}

func TestStaffClinicAssignmentRepository_Delete_DoesNotAffectOtherStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staffA := makeDoctor(t, db, clinicID, "削除対象スタッフ")
	staffB := makeDoctor(t, db, clinicID, "無関係スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staffA.ID, ClinicID: clinicID}))
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staffB.ID, ClinicID: clinicID}))

	require.NoError(t, repo.Delete(ctx, staffA.ID))

	foundB, err := repo.FindByStaffID(ctx, staffB.ID)
	require.NoError(t, err)
	assert.Len(t, foundB, 1, "無関係スタッフの所属は削除されない")
}

func TestStaffClinicAssignmentRepository_Create_DuplicateStaffClinicViolatesUniqueConstraint(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "重複所属スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}))

	err := repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID})
	require.Error(t, err, "同一 staff_id + clinic_id の重複所属は一意制約違反になるべき")
	assert.False(t, apperrors.IsNotFound(err))
}

func TestStaffClinicAssignmentRepository_Create_SoftDeletedPairStillViolatesFullUniqueConstraint(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "論理削除済み重複所属スタッフ")
	original := &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}
	require.NoError(t, repo.Create(ctx, original))
	require.NoError(t, db.WithContext(ctx).Delete(original).Error)

	err := repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID})

	require.Error(t, err, "production full unique constraint also rejects a duplicate soft-deleted pair")
	assert.False(t, apperrors.IsNotFound(err))
}
