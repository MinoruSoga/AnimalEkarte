package shiftentry

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

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TestRepository_SaveByStaffDate_ClinicIsolation は clinic A の upsert が
// 同一 (staff_id, date) の clinic B 既存エントリを更新せず、A 側の新規行になることを検証する
// （スコープは entry.ClinicID — 呼び出し元 service が認証済み clinicID を設定する契約）。
func TestRepository_SaveByStaffDate_ClinicIsolation(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := New(db)
	ctx := context.Background()

	staffB := makeDoctor(t, db, 2, "クリニックB医師")
	date := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	entryB := makeShiftEntryWithType(t, db, 2, staffB.ID, date, model.ShiftTypeFull)

	entryA := &model.ShiftEntry{
		ClinicID:  1,
		StaffID:   staffB.ID,
		Date:      date,
		ShiftType: model.ShiftTypeMorning,
	}
	require.NoError(t, repo.SaveByStaffDate(ctx, 1, entryA, nil))

	var reloadedB model.ShiftEntry
	require.NoError(t, db.First(&reloadedB, entryB.ID).Error)
	assert.Equal(t, model.ShiftTypeFull, reloadedB.ShiftType, "clinic B のエントリが更新されてはならない")
	assert.NotEqual(t, entryB.ID, entryA.ID, "clinic A の upsert は B の既存行ではなく新規行になる")
}

// TestRepository_DeleteByStaffDate_ClinicIsolation は clinic A から
// clinic B のシフトエントリを削除できない（NotFound・行残存）ことを検証する。
func TestRepository_DeleteByStaffDate_ClinicIsolation(t *testing.T) {
	db := setupShiftEntryTestDB(t)
	repo := New(db)
	ctx := context.Background()

	staffB := makeDoctor(t, db, 2, "クリニックB医師")
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
	repo := New(db)
	ctx := context.Background()

	staff := makeDoctor(t, db, 1, "医師")
	date := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	existing := makeShiftEntryWithType(t, db, 1, staff.ID, date, model.ShiftTypeFull)

	updated := &model.ShiftEntry{
		ClinicID:  1,
		StaffID:   staff.ID,
		Date:      date,
		ShiftType: model.ShiftTypeMorning,
	}
	breaks := []model.ShiftEntryBreak{{BreakStart: "12:00", BreakEnd: "13:00"}}
	require.NoError(t, repo.SaveByStaffDate(ctx, 1, updated, breaks))

	assert.Equal(t, existing.ID, updated.ID, "同一 (clinic, staff, date) は既存行の更新になる")
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
	repo := New(db)
	ctx := context.Background()

	staff := makeDoctor(t, db, 2, "医師")
	entry := &model.ShiftEntry{
		ClinicID:  2,
		StaffID:   staff.ID,
		Date:      time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC),
		ShiftType: model.ShiftTypeFull,
	}
	err := repo.SaveByStaffDate(ctx, 1, entry, nil)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	var count int64
	require.NoError(t, db.Model(&model.ShiftEntry{}).Where("staff_id = ?", staff.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count, "不一致時は一切書き込まれない")
}
