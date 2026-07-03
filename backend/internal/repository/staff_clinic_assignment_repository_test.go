package repository

// staff_clinic_assignment_repository_test.go — StaffClinicAssignmentRepository の統合テスト
// （実 Postgres テスト DB）。スタッフ-クリニック中間テーブルの CRUD と GORM SoftDelete スコープの
// 自動適用（コメントに記載された前提）を実際の DB 動作で検証する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupStaffClinicAssignmentRepositoryTestDB は staff_clinic_assignment_repository.go のテスト用に
// FK 親（companies/clinics）と staffs/staff_clinic_assignments を整備する。
func setupStaffClinicAssignmentRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.Company{}, &model.Clinic{}, &model.Staff{}, &model.StaffClinicAssignment{},
	))
	// model.StaffClinicAssignment has no `uniqueIndex` gorm tag, so AutoMigrate does not recreate
	// the composite UNIQUE(staff_id, clinic_id) constraint that migrations/001_init.sql defines
	// (uk_staff_clinic). Add it explicitly so duplicate-assignment rejection is actually exercised.
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_staff_clinic_assignment
		ON staff_clinic_assignments (staff_id, clinic_id) WHERE deleted_at IS NULL`)
	require.NoError(t, db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs CASCADE").Error)
	return db
}

func TestStaffClinicAssignmentRepository_Create_HappyPath(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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

func TestStaffClinicAssignmentRepository_FindByStaffID_EmptyForUnassignedStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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

func TestStaffClinicAssignmentRepository_Delete_RemovesAllAssignmentsForStaff(t *testing.T) {
	db := setupStaffClinicAssignmentRepositoryTestDB(t)
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
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
	repo := NewStaffClinicAssignmentRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)
	seedClinicsForFK(t, db, clinicID)

	staff := makeDoctor(t, db, clinicID, "重複所属スタッフ")
	require.NoError(t, repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID}))

	err := repo.Create(ctx, &model.StaffClinicAssignment{StaffID: staff.ID, ClinicID: clinicID})
	require.Error(t, err, "同一 staff_id + clinic_id の重複所属は一意制約違反になるべき")
	assert.False(t, apperrors.IsNotFound(err))
}
