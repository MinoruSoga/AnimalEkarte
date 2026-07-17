package repository

// reservation_schedule_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: ReservationScheduleRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B のシフトを
//   読み取れない、削除できない。
//
// clinicScope を reservation_schedule_repository.go から削除すると必ず失敗するよう設計されている。

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

// setupScheduleIsolationTestDB はシフト clinic_id 隔離テスト用の DB を整備する。
// setupTestDB が DROP TYPE shift_type CASCADE で shift_entries.shift_type 列を落とすため、
// TRUNCATE を AutoMigrate より前に実行して既存行 NULL 違反を防ぐ。
func setupScheduleIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	// AutoMigrate より先にクリア（DROP TYPE CASCADE で列が消えた後の NULL 違反回避）
	db.Exec("TRUNCATE TABLE shift_entries CASCADE")
	require.NoError(t, ensureAutoMigrated(db, &model.Staff{}, &model.ShiftEntry{}))
	return db
}

// makeShiftEntry はテスト用シフトエントリを1件作成する。
func makeShiftEntry(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time) {
	t.Helper()
	entry := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(entry).Error)
}

// TestReservationScheduleRepository_FindAllByDate_ClinicIsolation は
// clinic A のシフトを clinic B の clinicID で取得できないことを検証する。
// clinicScope を削除すると「別クリニックIDでは取得できない」が失敗する。
func TestReservationScheduleRepository_FindAllByDate_ClinicIsolation(t *testing.T) {
	db := setupScheduleIsolationTestDB(t)
	repo := NewReservationScheduleRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)

	staffA := makeDoctor(t, db, clinicA, "スケジュールA用スタッフ")
	makeShiftEntry(t, db, clinicA, staffA.ID, date)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindAllByDate(ctx, clinicA, staffA.ID, date)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicA, got.ClinicID)
		assert.Equal(t, staffA.ID, got.StaffID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindAllByDate(ctx, clinicB, staffA.ID, date)
		assert.Error(t, err, "clinic B から clinic A のシフトを取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestReservationScheduleRepository_FindAllByMonth_ClinicIsolation は
// clinic A のシフト一覧を clinic B の clinicID で取得できないことを検証する。
// clinicScope を削除すると「別クリニックIDでは0件」が失敗する。
func TestReservationScheduleRepository_FindAllByMonth_ClinicIsolation(t *testing.T) {
	db := setupScheduleIsolationTestDB(t)
	repo := NewReservationScheduleRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	staffA := makeDoctor(t, db, clinicA, "月次スケジュールA用スタッフ")
	makeShiftEntry(t, db, clinicA, staffA.ID, date)

	t.Run("同一クリニックIDでは1件取得できる", func(t *testing.T) {
		got, err := repo.FindAllByMonth(ctx, clinicA, staffA.ID, "2026-06")
		require.NoError(t, err)
		assert.Len(t, got, 1, "clinic A のシフトが1件見えるはず")
		assert.Equal(t, clinicA, got[0].ClinicID)
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindAllByMonth(ctx, clinicB, staffA.ID, "2026-06")
		require.NoError(t, err, "FindAllByMonth は空リストを返す（エラーではない）")
		assert.Empty(t, got, "clinic B から clinic A のシフトを見てはならない")
	})
}

// TestReservationScheduleRepository_Delete_ClinicIsolation は
// 別クリニックIDからの Delete が NotFound を返し、シフトが削除されないことを検証する。
// clinicScope を削除すると「シフトはまだ存在する」が失敗する。
func TestReservationScheduleRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupScheduleIsolationTestDB(t)
	repo := NewReservationScheduleRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	date := time.Date(2026, 6, 26, 0, 0, 0, 0, time.UTC)

	staffA := makeDoctor(t, db, clinicA, "削除テストA用スタッフ")
	makeShiftEntry(t, db, clinicA, staffA.ID, date)

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, staffA.ID, date)
		require.Error(t, err, "clinic B から clinic A のシフトを削除できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("シフトはまだ存在する（不正削除防止）", func(t *testing.T) {
		var entry model.ShiftEntry
		require.NoError(t, db.Where("clinic_id = ? AND staff_id = ? AND date = ?",
			clinicA, staffA.ID, date.Format("2006-01-02")).First(&entry).Error)
		assert.Equal(t, staffA.ID, entry.StaffID, "clinic A のシフトは別クリニックの Delete で消えてはならない")
	})

	t.Run("正しいクリニックIDからの Delete は成功する", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, staffA.ID, date)
		require.NoError(t, err)
		var entry model.ShiftEntry
		assert.Error(t, db.Where("clinic_id = ? AND staff_id = ? AND date = ?",
			clinicA, staffA.ID, date.Format("2006-01-02")).First(&entry).Error,
			"削除後は entry が存在してはならない")
	})
}
