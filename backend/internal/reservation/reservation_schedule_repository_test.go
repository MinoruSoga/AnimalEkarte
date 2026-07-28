package reservation

// BE9-2C R②: 実装は internal/reservation へ移動済み。本テストは意図的に repository 残置 —
// production ctor（facade）が staff domain の shift_entries write 書き込み者を注入する配線そのものを
// 実 DB で検証するため、両 domain を構築できる本 package が唯一の置き場（count_clinic_scope 先例）。

// reservation_schedule_repository_test.go
// ReservationScheduleRepository（shift_entries + shift_entry_breaks）の CRUD メソッドを実 DB で検証する。
// clinic_id 隔離の回帰テストは reservation_schedule_clinic_isolation_test.go が別途カバーしているため、
// 本ファイルは Breaks 系メソッド・Save の新規作成/更新分岐・バリデーション・NotFound に焦点を当てる。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	staffpkg "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationScheduleCRUDTestDB は ReservationScheduleRepository の CRUD テスト用に DB を整備する。
func setupReservationScheduleCRUDTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{}, &model.Staff{},
		&model.StaffClinicAssignment{}, &model.ShiftEntry{}, &model.ShiftEntryBreak{},
	))
	require.NoError(t, db.Exec(
		"TRUNCATE TABLE shift_entries, staff_clinic_assignments, staffs CASCADE",
	).Error)
	seedClinicsForFK(t, db, 1, 2)
	// shift_entry_breaks.break_start/break_end can be left as "timestamp with time zone" in a
	// pre-existing ekarte_db_test if the table was ever created before model.ShiftEntryBreak's
	// `gorm:"type:time"` tag existed — AutoMigrate never ALTERs an existing column's type, only
	// adds missing ones. Force the correct type every run (no-op cast if already "time").
	db.Exec(`ALTER TABLE shift_entry_breaks ALTER COLUMN break_start TYPE time USING break_start::time`)
	db.Exec(`ALTER TABLE shift_entry_breaks ALTER COLUMN break_end TYPE time USING break_end::time`)
	return db
}

func TestReservationScheduleRepository_FindAllByMonth_InvalidFormat(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()

	got, err := repo.FindAllByMonth(ctx, 1, 1, "2026/06")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err), "不正な month フォーマットは InvalidInput であるべき: %v", err)
}

func TestReservationScheduleRepository_FindAllByDate_NotFound(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "予定なしスタッフ")
	got, err := repo.FindAllByDate(ctx, clinicA, staffA.ID, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestReservationScheduleRepository_FindAllBreaksByEntryID(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "休憩確認スタッフ")
	date := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	makeShiftEntry(t, db, clinicA, staffA.ID, date)

	var entry model.ShiftEntry
	require.NoError(t, db.Where("clinic_id = ? AND staff_id = ? AND date = ?", clinicA, staffA.ID, date.Format("2006-01-02")).First(&entry).Error)

	require.NoError(t, db.Create(&model.ShiftEntryBreak{ShiftEntryID: entry.ID, BreakStart: "12:00", BreakEnd: "13:00"}).Error)

	t.Run("紐づく休憩を取得できる", func(t *testing.T) {
		breaks, err := repo.FindAllBreaksByEntryID(ctx, clinicA, entry.ID)
		require.NoError(t, err)
		require.Len(t, breaks, 1)
		assert.Equal(t, "12:00:00", breaks[0].BreakStart)
	})

	t.Run("紐づく休憩がない場合は空スライス", func(t *testing.T) {
		breaks, err := repo.FindAllBreaksByEntryID(ctx, clinicA, entry.ID+999999)
		require.NoError(t, err)
		assert.Empty(t, breaks)
	})
}

func TestReservationScheduleRepository_FindAllBreaksByEntryIDs(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "複数休憩スタッフ")
	dateA := time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	dateB := time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)
	makeShiftEntry(t, db, clinicA, staffA.ID, dateA)
	makeShiftEntry(t, db, clinicA, staffA.ID, dateB)

	var entryA, entryB model.ShiftEntry
	require.NoError(t, db.Where("staff_id = ? AND date = ?", staffA.ID, dateA.Format("2006-01-02")).First(&entryA).Error)
	require.NoError(t, db.Where("staff_id = ? AND date = ?", staffA.ID, dateB.Format("2006-01-02")).First(&entryB).Error)

	require.NoError(t, db.Create(&model.ShiftEntryBreak{ShiftEntryID: entryA.ID, BreakStart: "12:00", BreakEnd: "13:00"}).Error)
	require.NoError(t, db.Create(&model.ShiftEntryBreak{ShiftEntryID: entryB.ID, BreakStart: "10:00", BreakEnd: "10:15"}).Error)

	t.Run("複数エントリの休憩をまとめて取得できる", func(t *testing.T) {
		result, err := repo.FindAllBreaksByEntryIDs(ctx, clinicA, []uint64{entryA.ID, entryB.ID})
		require.NoError(t, err)
		require.Len(t, result[entryA.ID], 1)
		require.Len(t, result[entryB.ID], 1)
		assert.Equal(t, "12:00:00", result[entryA.ID][0].BreakStart)
	})

	t.Run("空スライスは空マップを返す（クエリを発行しない）", func(t *testing.T) {
		result, err := repo.FindAllBreaksByEntryIDs(ctx, clinicA, []uint64{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	// RSV-08: parent clinic correlation — foreign clinic_id must not return breaks for entry IDs.
	t.Run("他院 clinic_id では休憩を返さない", func(t *testing.T) {
		result, err := repo.FindAllBreaksByEntryIDs(ctx, clinicA+1, []uint64{entryA.ID, entryB.ID})
		require.NoError(t, err)
		assert.Empty(t, result[entryA.ID])
		assert.Empty(t, result[entryB.ID])
	})
}

// TestReservationScheduleRepository_FindAllByStaffIDsAndDateRange は G7-1(LIFF日付ループN+1解消)の
// プリフェッチ用バッチメソッドを検証する。
func TestReservationScheduleRepository_FindAllByStaffIDsAndDateRange(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)
	const clinicB = uint64(2)

	staff1 := makeDoctor(t, db, clinicA, "スタッフ1")
	staff2 := makeDoctor(t, db, clinicA, "スタッフ2")
	otherClinicStaff := makeDoctor(t, db, clinicB, "他院スタッフ")

	inRange1 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	inRange2 := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	outOfRange := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	makeShiftEntry(t, db, clinicA, staff1.ID, inRange1)
	makeShiftEntry(t, db, clinicA, staff2.ID, inRange2)
	makeShiftEntry(t, db, clinicA, staff1.ID, outOfRange)
	makeShiftEntry(t, db, clinicB, otherClinicStaff.ID, inRange1)

	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	t.Run("複数スタッフの期間内シフトを1クエリでまとめて返す", func(t *testing.T) {
		entries, err := repo.FindAllByStaffIDsAndDateRange(ctx, clinicA, []uint64{staff1.ID, staff2.ID}, from, to)
		require.NoError(t, err)
		require.Len(t, entries, 2, "範囲内かつ同クリニックの2件のみ(範囲外1件・別クリニック1件は除外)")
	})

	t.Run("toは排他的上限（指定日そのものは含まれない）", func(t *testing.T) {
		entries, err := repo.FindAllByStaffIDsAndDateRange(ctx, clinicA, []uint64{staff1.ID}, from, outOfRange)
		require.NoError(t, err)
		for _, e := range entries {
			assert.False(t, e.Date.Equal(outOfRange), "to日付そのものは半開区間の外")
		}
	})

	t.Run("staffIDsが空の場合は空スライスを即返す", func(t *testing.T) {
		entries, err := repo.FindAllByStaffIDsAndDateRange(ctx, clinicA, []uint64{}, from, to)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})

	t.Run("別クリニックIDでは0件（clinic_id分離）", func(t *testing.T) {
		entries, err := repo.FindAllByStaffIDsAndDateRange(ctx, clinicB, []uint64{staff1.ID, staff2.ID}, from, to)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

func TestReservationScheduleRepository_Save_CreatesNewEntry(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "新規保存スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID:  staffA.ID,
		ClinicID: clinicA,
		IsMain:   true,
	}).Error)
	date := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

	entry := &model.ShiftEntry{
		ClinicID:  clinicA,
		StaffID:   staffA.ID,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
		Notes:     "初回作成",
	}
	breaks := []model.ShiftEntryBreak{{BreakStart: "12:00", BreakEnd: "13:00"}}

	saved, savedBreaks, created, err := repo.Save(ctx, clinicA, entry, breaks)
	require.NoError(t, err)
	assert.True(t, created)
	require.NotNil(t, saved)
	assert.NotZero(t, saved.ID)
	require.Len(t, savedBreaks, 1)

	got, err := repo.FindAllByDate(ctx, clinicA, staffA.ID, date)
	require.NoError(t, err)
	assert.Equal(t, "初回作成", got.Notes)

	reloadedBreaks, err := repo.FindAllBreaksByEntryID(ctx, clinicA, got.ID)
	require.NoError(t, err)
	require.Len(t, reloadedBreaks, 1)
}

func TestReservationScheduleRepository_Save_UpdatesExistingEntry(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "更新保存スタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID:  staffA.ID,
		ClinicID: clinicA,
		IsMain:   true,
	}).Error)
	date := time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)

	// 初回作成
	first := &model.ShiftEntry{
		ClinicID:  clinicA,
		StaffID:   staffA.ID,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
		Notes:     "旧メモ",
	}
	savedFirst, savedFirstBreaks, firstCreated, err := repo.Save(ctx, clinicA, first, nil)
	require.NoError(t, err)
	assert.True(t, firstCreated)
	require.NotNil(t, savedFirst)
	assert.Empty(t, savedFirstBreaks)
	firstID := savedFirst.ID

	// 同一 staff/date で再度 Save → 更新分岐
	second := &model.ShiftEntry{
		ClinicID:  clinicA,
		StaffID:   staffA.ID,
		Date:      date,
		ShiftType: model.ShiftTypeMorning,
		Notes:     "新メモ",
	}
	savedSecond, savedSecondBreaks, secondCreated, err := repo.Save(
		ctx,
		clinicA,
		second,
		[]model.ShiftEntryBreak{{BreakStart: "09:00", BreakEnd: "09:15"}},
	)
	require.NoError(t, err)
	assert.False(t, secondCreated)
	require.NotNil(t, savedSecond)
	require.Len(t, savedSecondBreaks, 1)

	assert.Equal(t, firstID, savedSecond.ID, "既存エントリの ID を引き継ぐべき")

	got, err := repo.FindAllByDate(ctx, clinicA, staffA.ID, date)
	require.NoError(t, err)
	assert.Equal(t, "新メモ", got.Notes)
	assert.Equal(t, model.ShiftTypeMorning, got.ShiftType)

	savedBreaks, err := repo.FindAllBreaksByEntryID(ctx, clinicA, got.ID)
	require.NoError(t, err)
	require.Len(t, savedBreaks, 1, "既存の休憩は削除され新しい休憩のみ残るべき")
	assert.Equal(t, "09:00:00", savedBreaks[0].BreakStart)
}

func TestReservationScheduleRepository_Delete_NotFound(t *testing.T) {
	db := setupReservationScheduleCRUDTestDB(t)
	repo := NewReservationScheduleRepository(db, staffpkg.NewShiftEntryRepository(db))
	ctx := context.Background()
	const clinicA = uint64(1)

	staffA := makeDoctor(t, db, clinicA, "削除対象なしスタッフ")
	require.NoError(t, db.WithContext(ctx).Create(&model.StaffClinicAssignment{
		StaffID:  staffA.ID,
		ClinicID: clinicA,
		IsMain:   true,
	}).Error)
	err := repo.Delete(ctx, clinicA, staffA.ID, time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC))
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
