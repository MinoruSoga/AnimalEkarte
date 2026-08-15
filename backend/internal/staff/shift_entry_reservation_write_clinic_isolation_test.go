package staff_test

// reservation_write_clinic_isolation_test.go — ADR-006 論点#1 案A（shift_entries 書込一本化）
//
// テスト対象: repository.SaveByStaffDate / DeleteByStaffDate の clinic_id 境界。
// 保護する不変条件: 一本化後も staff domain（本package）側 write path が clinic-scope
//   検証を維持する（裁定の実装条件ii）。reservation 側の既存テストは delegate 経由で
//   挙動を検証するが、本ファイルは移動先メソッドを直接呼び、スコープ述語の削除で必ず失敗する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
)

func TestRepository_SaveByStaffDate_RejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		entry    *model.ShiftEntry
	}{
		{
			name:     "nil entry",
			clinicID: 1,
			entry:    nil,
		},
		{
			name:     "clinic id mismatch",
			clinicID: 1,
			entry: &model.ShiftEntry{
				ClinicID: 2,
				StaffID:  3,
				Date:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewShiftEntryRepository(nil)

			saved, breaks, created, err := repo.SaveByStaffDate(
				context.Background(),
				tt.clinicID,
				tt.entry,
				nil,
			)

			assert.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
			assert.Nil(t, saved)
			assert.Nil(t, breaks)
			assert.False(t, created)
		})
	}
}

// TestRepository_SaveByStaffDate_RejectsUnauthorizedStaffGraph は、shift write の前に
// active staff identity と requested clinic の active assignment の双方を検証する。
func TestRepository_SaveByStaffDate_RejectsUnauthorizedStaffGraph(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, db *gorm.DB) (clinicID, staffID uint64)
	}{
		{
			name: "foreign clinic assignment",
			prepare: func(t *testing.T, db *gorm.DB) (uint64, uint64) {
				staff := makeShiftEntryDoctor(t, db, 2, "クリニックB医師")
				return 1, staff.ID
			},
		},
		{
			name: "soft deleted requested clinic assignment",
			prepare: func(t *testing.T, db *gorm.DB) (uint64, uint64) {
				staff := makeShiftEntryDoctor(t, db, 1, "所属解除済み医師")
				require.NoError(t, db.Where(
					"staff_id = ? AND clinic_id = ?",
					staff.ID,
					uint64(1),
				).Delete(&model.StaffClinicAssignment{}).Error)
				return 1, staff.ID
			},
		},
		{
			name: "soft deleted staff identity",
			prepare: func(t *testing.T, db *gorm.DB) (uint64, uint64) {
				staff := makeShiftEntryDoctor(t, db, 1, "削除済み医師")
				require.NoError(t, db.Delete(&model.Staff{}, staff.ID).Error)
				return 1, staff.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupShiftEntryTestDB(t)
			repo := NewShiftEntryRepository(db)
			clinicID, staffID := tt.prepare(t, db)
			date := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
			entry := &model.ShiftEntry{
				ClinicID:  clinicID,
				StaffID:   staffID,
				Date:      date,
				ShiftType: model.ShiftTypeMorning,
			}

			_, _, _, err := repo.SaveByStaffDate(context.Background(), clinicID, entry, nil)

			require.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err), "unauthorized graph must be indistinguishable from not found: %v", err)
			var count int64
			require.NoError(t, db.Model(&model.ShiftEntry{}).
				Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicID, staffID, date).
				Count(&count).Error)
			assert.Zero(t, count, "authorization failure must not write a shift entry")
		})
	}
}

// TestRepository_DeleteByStaffDate_ClinicIsolation は clinic A から
// clinic B のシフトエントリを削除できない（NotFound・行残存）ことを検証する。
func TestRepository_DeleteByStaffDate_ClinicIsolation(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	staffB := makeShiftEntryDoctor(t, db, 2, "クリニックB医師")
	date := time.Date(2026, 7, 22, 0, 0, 0, 0, time.UTC)
	entryB := makeShiftEntryWithType(t, db, 2, staffB.ID, date, model.ShiftTypeFull)

	err := repo.DeleteByStaffDate(ctx, 1, staffB.ID, date)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))

	var count int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("id = ?", entryB.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count, "clinic B のエントリが削除されてはならない")
}

// TestRepository_SaveByStaffDate_UpsertsWithinClinic は同一クリニック内の upsert が
// 既存行の更新+breaks 全置換になる（delegate 移動後の挙動保持）ことを検証する。
func TestRepository_SaveByStaffDate_UpsertsWithinClinic(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	staff := makeShiftEntryDoctor(t, db, 1, "医師")
	date := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	existing := makeShiftEntryWithType(t, db, 1, staff.ID, date, model.ShiftTypeFull)

	updated := &model.ShiftEntry{
		ClinicID:  1,
		StaffID:   staff.ID,
		Date:      date,
		ShiftType: model.ShiftTypeMorning,
	}
	breaks := []model.ShiftEntryBreak{{BreakStart: "12:00", BreakEnd: "13:00"}}
	saved, savedBreaks, created, err := repo.SaveByStaffDate(ctx, 1, updated, breaks)
	require.NoError(t, err)

	assert.False(t, created)
	require.NotNil(t, saved)
	assert.Equal(t, existing.ID, saved.ID, "同一 (clinic, staff, date) は既存行の更新になる")
	require.Len(t, savedBreaks, 1)
	assert.Equal(t, existing.ID, savedBreaks[0].ShiftEntryID)
	var reloaded model.ShiftEntry
	require.NoError(t, db.Preload("Breaks").First(&reloaded, existing.ID).Error)
	assert.Equal(t, model.ShiftTypeMorning, reloaded.ShiftType)
	require.Len(t, reloaded.Breaks, 1)
	assert.Equal(t, "12:00:00", reloaded.Breaks[0].BreakStart) // DB の time 型で HH:MM:SS へ正規化される
}

// TestRepository_SaveByStaffDate_RejectsClinicIDMismatch は認証済み clinicID と
// entry.ClinicID の不一致を書込前に拒否する（fail-closed 防御）ことを検証する。
func TestRepository_SaveByStaffDate_RejectsClinicIDMismatch(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := NewShiftEntryRepository(db)
	ctx := context.Background()

	staff := makeShiftEntryDoctor(t, db, 2, "医師")
	entry := &model.ShiftEntry{
		ClinicID:  2,
		StaffID:   staff.ID,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: model.ShiftTypeFull,
	}
	_, _, _, err := repo.SaveByStaffDate(ctx, 1, entry, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	var count int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("staff_id = ?", staff.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count, "不一致時は一切書き込まれない")
}
