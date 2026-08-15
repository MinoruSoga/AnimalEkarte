package staff_test

// staff_reservation_write_clinic_isolation_test.go — ADR-006 論点#1 案A（staffs 書込一本化）
//
// テスト対象: staffRepository の予約用途 write（CreateForReservation /
//   UpdateForReservation / SwapSortOrderForReservation）の clinic_id 境界。
// 保護する不変条件: 一本化後も staff domain 側 write path が clinic-scope 検証を維持する
//   （裁定の実装条件ii）。reservation 側の既存テストは delegate 経由で挙動を検証するが、
//   本ファイルは移動先の staff domain メソッドを直接呼び、スコープ述語の削除で必ず失敗する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func setupStaffReservationWriteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.Staff{}, &model.StaffClinicAssignment{}))
	db.Exec("TRUNCATE TABLE staff_clinic_assignments, staffs CASCADE")
	return db
}

func makeAssignedDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string, sortOrder int) *model.Staff {
	t.Helper()
	s := &model.Staff{ClinicID: clinicID, Name: name, StaffType: model.StaffTypeDoctor, SortOrder: sortOrder}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	require.NoError(t, db.WithContext(context.Background()).Create(&model.StaffClinicAssignment{
		StaffID: s.ID, ClinicID: clinicID, IsMain: true,
	}).Error)
	return s
}

// TestStaffRepository_CreateForReservation_BindsAssignmentToClinic は
// 作成された StaffClinicAssignment が呼び出し元クリニックへ IsMain で紐づくことを検証する。
func TestStaffRepository_CreateForReservation_BindsAssignmentToClinic(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staff := &model.Staff{ClinicID: 1, Name: "予約医師", StaffType: model.StaffTypeDoctor}
	require.NoError(t, repo.CreateForReservation(ctx, staff, 1))

	var assignment model.StaffClinicAssignment
	require.NoError(t, db.Where("staff_id = ?", staff.ID).First(&assignment).Error)
	assert.Equal(t, uint64(1), assignment.ClinicID)
	assert.True(t, assignment.IsMain)
}

// BUG-455-S7: reservation staff create path must persist explicit reservation_visible=false.
func TestStaffRepository_CreateForReservation_ReservationVisibleFalsePersists(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staff := &model.Staff{
		ClinicID:           1,
		Name:               "hidden reservation staff",
		StaffType:          model.StaffTypeDoctor,
		ReservationVisible: false,
	}
	require.NoError(t, repo.CreateForReservation(ctx, staff, 1))
	assert.False(t, staff.ReservationVisible)

	var raw bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.Staff{}).
		Select("reservation_visible").
		Where("id = ?", staff.ID).
		Scan(&raw).Error)
	assert.False(t, raw, "raw reservation_visible must be false")
}

func TestStaffRepository_CreateForReservation_RejectsClinicIDMismatchWithoutWrite(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()
	staff := &model.Staff{ClinicID: 2, Name: "越境予約用医師", StaffType: model.StaffTypeDoctor}

	err := repo.CreateForReservation(ctx, staff, 1)

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "clinic mismatch must fail closed: %v", err)
	var staffCount int64
	require.NoError(t, db.Model(&model.Staff{}).Where("name = ?", staff.Name).Count(&staffCount).Error)
	assert.Zero(t, staffCount)
	var assignmentCount int64
	require.NoError(t, db.Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ?", staff.ID).
		Count(&assignmentCount).Error)
	assert.Zero(t, assignmentCount)
}

// TestStaffRepository_UpdateForReservation_ClinicIsolation は clinic A から
// clinic B のスタッフを更新できない（NotFound・無変更）ことを検証する。
func TestStaffRepository_UpdateForReservation_ClinicIsolation(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staffB := makeAssignedDoctor(t, db, 2, "クリニックB医師", 1)

	err := repo.UpdateForReservation(ctx, 1, staffB.ID, map[string]any{"name": "改ざん"})
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var reloaded model.Staff
	require.NoError(t, db.First(&reloaded, staffB.ID).Error)
	assert.Equal(t, "クリニックB医師", reloaded.Name)
}

// TestStaffRepository_SwapSortOrderForReservation_ClinicIsolation は clinic A から
// clinic B のスタッフの並び順を操作できない（NotFound・sort_order 無変更）ことを検証する。
func TestStaffRepository_SwapSortOrderForReservation_ClinicIsolation(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staffB1 := makeAssignedDoctor(t, db, 2, "B医師1", 1)
	staffB2 := makeAssignedDoctor(t, db, 2, "B医師2", 2)

	err := repo.SwapSortOrderForReservation(ctx, 1, staffB2.ID, "up")
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var r1, r2 model.Staff
	require.NoError(t, db.First(&r1, staffB1.ID).Error)
	require.NoError(t, db.First(&r2, staffB2.ID).Error)
	assert.Equal(t, 1, r1.SortOrder)
	assert.Equal(t, 2, r2.SortOrder)
}

// TestStaffRepository_SwapSortOrderForReservation_SwapsWithinClinic は同一クリニック内で
// 隣接スタッフと sort_order が入れ替わる（delegate 移動後の挙動保持）ことを検証する。
func TestStaffRepository_SwapSortOrderForReservation_SwapsWithinClinic(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staff1 := makeAssignedDoctor(t, db, 1, "医師1", 1)
	staff2 := makeAssignedDoctor(t, db, 1, "医師2", 2)

	require.NoError(t, repo.SwapSortOrderForReservation(ctx, 1, staff2.ID, "up"))

	var r1, r2 model.Staff
	require.NoError(t, db.First(&r1, staff1.ID).Error)
	require.NoError(t, db.First(&r2, staff2.ID).Error)
	assert.Equal(t, 2, r1.SortOrder)
	assert.Equal(t, 1, r2.SortOrder)
}

// A shared staff row keeps one global sort_order owned by staffs.clinic_id. A
// secondary-clinic assignment grants visibility but must not grant authority to
// mutate the ordering observed by the primary clinic.
func TestStaffRepository_SwapSortOrderForReservation_RejectsSecondaryClinicAssignment(t *testing.T) {
	db := setupStaffReservationWriteTestDB(t)
	repo := NewStaffRepository(db)
	ctx := context.Background()

	staff1 := makeAssignedDoctor(t, db, 2, "共有医師1", 1)
	staff2 := makeAssignedDoctor(t, db, 2, "共有医師2", 2)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff1.ID,
		ClinicID: 1,
		IsMain:   false,
	}).Error)
	require.NoError(t, db.Create(&model.StaffClinicAssignment{
		StaffID:  staff2.ID,
		ClinicID: 1,
		IsMain:   false,
	}).Error)

	err := repo.SwapSortOrderForReservation(ctx, 1, staff2.ID, "up")
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "secondary clinic reorder must fail closed: %v", err)

	var reloaded1, reloaded2 model.Staff
	require.NoError(t, db.First(&reloaded1, staff1.ID).Error)
	require.NoError(t, db.First(&reloaded2, staff2.ID).Error)
	assert.Equal(t, 1, reloaded1.SortOrder)
	assert.Equal(t, 2, reloaded2.SortOrder)
}
