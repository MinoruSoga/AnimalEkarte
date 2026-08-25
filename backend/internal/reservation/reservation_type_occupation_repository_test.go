package reservation

// reservation_type_occupation_repository_test.go
// ReservationTypeOccupationRepository（予約区分×職種の中間テーブル）の CRUD・Occupation Preload の
// clinic_id 隔離・CountWorkingStaffByReservationTypeIDs の raw SQL 集計を実 DB で検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationTypeOccupationTestDB は ReservationTypeOccupationRepository のテスト用に DB を整備する。
// AutoMigrate より先に shift_entries / reservation_types をクリアする（setupTestDB が
// shift_type / reservation_type_category ENUM を DROP CASCADE し列が消えるため、NULL 違反回避）。
// 最後の occupations CASCADE は staffs.occupation_id FK / reservation_type_occupations.occupation_id FK
// 経由で両テーブルも連鎖 TRUNCATE する。
func setupReservationTypeOccupationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	db.Exec("TRUNCATE TABLE shift_entries CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Occupation{}, &model.Staff{}, &model.ShiftEntry{},
		&model.ReservationType{}, &model.ReservationTypeOccupation{},
	))
	db.Exec("TRUNCATE TABLE occupations CASCADE")
	return db
}

// makeOccupation は repository/staff_occupation_preload_clinic_isolation_test.go の
// 同名ヘルパーの最小限の複製（BE9-2C R①: package 跨ぎで test helper import 不能のため）。
func makeOccupation(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Occupation {
	t.Helper()
	o := &model.Occupation{ClinicID: clinicID, Name: name}
	require.NoError(t, db.WithContext(context.Background()).Create(o).Error)
	return o
}

// makeStaffWithOccupation も同ファイルからの最小限の複製（同上）。
// CountWorkingStaff は is_active / reservation_visible を条件にするため明示する。
func makeStaffWithOccupation(t *testing.T, db *gorm.DB, clinicID, occupationID uint64, name string) *model.Staff {
	t.Helper()
	return makeStaffWithOccupationVisibility(t, db, clinicID, occupationID, name, true, true)
}

func makeStaffWithOccupationVisibility(t *testing.T, db *gorm.DB, clinicID, occupationID uint64, name string, isActive, reservationVisible bool) *model.Staff {
	t.Helper()
	oid := occupationID
	s := &model.Staff{
		ClinicID:           clinicID,
		Name:               name,
		StaffType:          model.StaffTypeDoctor,
		OccupationID:       &oid,
		IsActive:           true,
		ReservationVisible: true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(s).Error)
	if isActive && reservationVisible {
		return s
	}
	// gorm:"default:true" はゼロ値 false を DEFAULT に置き換えるため、false は UpdateColumns で書く。
	require.NoError(t, db.WithContext(context.Background()).Model(s).UpdateColumns(map[string]interface{}{
		"is_active":           isActive,
		"reservation_visible": reservationVisible,
	}).Error)
	s.IsActive = isActive
	s.ReservationVisible = reservationVisible
	return s
}

// makeShiftEntryWithType は repository/shiftentry の同名ヘルパーの最小限の複製
// （BE9-2C R①: 移動先 package から shiftentry の test helper は import 不能のため）。
func makeShiftEntryWithType(t *testing.T, db *gorm.DB, clinicID, staffID uint64, date time.Time, shiftType model.ShiftType) *model.ShiftEntry {
	t.Helper()
	e := &model.ShiftEntry{
		ClinicID:  clinicID,
		StaffID:   staffID,
		Date:      date,
		ShiftType: shiftType,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(e).Error)
	return e
}

func makeReservationTypeOccupationLink(t *testing.T, db *gorm.DB, clinicID, reservationTypeID, occupationID uint64) *model.ReservationTypeOccupation {
	t.Helper()
	rto := &model.ReservationTypeOccupation{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		OccupationID:      occupationID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rto).Error)
	return rto
}

func TestReservationTypeOccupationRepository_FindAll(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "職種紐付け区分A", nil, nil)
	occA := makeOccupation(t, db, clinicA, "医院Aの職種")
	occB := makeOccupation(t, db, clinicB, "医院Bの職種")

	legit := makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)
	// (cross) clinic A の紐付けに clinic B の occupation_id を植え付け（汚染データ模擬）
	cross := makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occB.ID)

	t.Run("同一クリニックの紐付けを取得し、Occupation は clinic_id が一致する場合のみ Preload される", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, rtA.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)

		byID := make(map[uint64]model.ReservationTypeOccupation, len(got))
		for _, rto := range got {
			byID[rto.ID] = rto
		}

		gotLegit, ok := byID[legit.ID]
		require.True(t, ok)
		require.NotNil(t, gotLegit.Occupation, "同一クリニックの Occupation は Preload されるべき")
		assert.Equal(t, occA.ID, gotLegit.Occupation.ID)

		gotCross, ok := byID[cross.ID]
		require.True(t, ok, "越境FKの紐付け行自体は base クエリで返るべき")
		assert.Nil(t, gotCross.Occupation, "別クリニックの Occupation マスタが混入してはならない")
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB, rtA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationTypeOccupationRepository_FindByID(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "職種単体取得区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "単体取得用職種")
	link := makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)

	t.Run("同一クリニックIDで取得でき、Occupation が Preload される", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, rtA.ID, occA.ID)
		require.NoError(t, err)
		assert.Equal(t, link.ID, got.ID)
		require.NotNil(t, got.Occupation)
		assert.Equal(t, occA.ID, got.Occupation.ID)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, rtA.ID, occA.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない組み合わせは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, rtA.ID, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeOccupationRepository_Create(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rtA := makeReservationTypeLinked(t, db, clinicA, "新規紐付け区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "新規紐付け職種")

	rto := &model.ReservationTypeOccupation{ClinicID: clinicA, ReservationTypeID: rtA.ID, OccupationID: occA.ID}
	require.NoError(t, repo.Create(ctx, rto))
	assert.NotZero(t, rto.ID)

	got, err := repo.FindByID(ctx, clinicA, rtA.ID, occA.ID)
	require.NoError(t, err)
	assert.Equal(t, rto.ID, got.ID)
}

func TestReservationTypeOccupationRepository_Delete(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "削除対象紐付け区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "削除対象職種")
	makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, rtA.ID, occA.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正しいクリニックIDで削除できる（物理削除）", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, rtA.ID, occA.ID))
		_, err := repo.FindByID(ctx, clinicA, rtA.ID, occA.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("既に削除済みの再削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, rtA.ID, occA.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs は G7-1(LIFF日付ループ
// N+1解消)のバッチ版を検証する。単日版と同じ off/paid_leave除外・clinic_id隔離ロジックを複数日分まとめて確認する。
func TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	dateA := time.Date(2026, 6, 15, 0, 0, 0, 0, config.JST)
	dateB := time.Date(2026, 6, 16, 0, 0, 0, 0, config.JST)
	dateNoShift := time.Date(2026, 6, 17, 0, 0, 0, 0, config.JST)

	rtA := makeReservationTypeLinked(t, db, clinicA, "バッチ集計対象区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "バッチ集計対象職種")
	makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)

	workingA := makeStaffWithOccupation(t, db, clinicA, occA.ID, "dateA出勤スタッフ")
	makeShiftEntryWithType(t, db, clinicA, workingA.ID, dateA, model.ShiftTypeFull)

	working1B := makeStaffWithOccupation(t, db, clinicA, occA.ID, "dateB出勤スタッフ1")
	makeShiftEntryWithType(t, db, clinicA, working1B.ID, dateB, model.ShiftTypeFull)
	working2B := makeStaffWithOccupation(t, db, clinicA, occA.ID, "dateB出勤スタッフ2")
	makeShiftEntryWithType(t, db, clinicA, working2B.ID, dateB, model.ShiftTypeMorning)
	afternoon := makeStaffWithOccupation(t, db, clinicA, occA.ID, "dateB午後スタッフ")
	makeShiftEntryWithType(t, db, clinicA, afternoon.ID, dateB, model.ShiftTypeAfternoon)

	off := makeStaffWithOccupation(t, db, clinicA, occA.ID, "dateB休日スタッフ")
	makeShiftEntryWithType(t, db, clinicA, off.ID, dateB, model.ShiftTypeOff)

	_ = makeStaffWithOccupation(t, db, clinicB, occA.ID, "他院スタッフ")
	_ = makeStaffWithOccupationVisibility(t, db, clinicA, occA.ID, "無効スタッフ", false, true)
	_ = makeStaffWithOccupationVisibility(t, db, clinicA, occA.ID, "非公開スタッフ", true, false)

	const eligibleClinicA = int64(5) // workingA, working1B, working2B, afternoon, off。他院・無効・非公開は除外。

	t.Run("複数日分の出勤スタッフ数を1クエリでまとめて返す", func(t *testing.T) {
		result, err := repo.CountWorkingStaffByReservationTypeIDs(ctx, clinicA, rtA.ID, []time.Time{dateA, dateB, dateNoShift})
		require.NoError(t, err)
		assert.Equal(t, eligibleClinicA, result[dateA.Format("2006-01-02")], "dateAはoff/paid_leaveなしなので候補全員（シフトなしも含む）")
		assert.Equal(t, eligibleClinicA-1, result[dateB.Format("2006-01-02")], "dateBはoff 1名を除き、full/morning/afternoonとシフトなしを含む")
		noShiftKey := dateNoShift.Format("2006-01-02")
		assert.Equal(t, eligibleClinicA, result[noShiftKey], "シフトが無い日も出勤候補スタッフ数を返す")
		_, hasNoShiftKey := result[noShiftKey]
		assert.True(t, hasNoShiftKey, "候補が1人以上ならシフト無し日のキーも埋める")
	})

	t.Run("別クリニックIDでは空map（clinic_id 隔離）", func(t *testing.T) {
		result, err := repo.CountWorkingStaffByReservationTypeIDs(ctx, clinicB, rtA.ID, []time.Time{dateA, dateB})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("datesが空の場合は空mapを即返す", func(t *testing.T) {
		result, err := repo.CountWorkingStaffByReservationTypeIDs(ctx, clinicA, rtA.ID, []time.Time{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs_OccupationGuardNoShiftEntries
// は BUG-002: 職種紐付けあり・該当スタッフが is_active/reservation_visible・その日の shift_entries が
// 0件でも出勤候補として数え、applyOccupationGuard が Available を staff_off で潰さないことを固定する。
func TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs_OccupationGuardNoShiftEntries(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	date := time.Date(2026, 7, 1, 0, 0, 0, 0, config.JST)
	dateStr := date.Format(time.DateOnly)

	rtA := makeReservationTypeLinked(t, db, clinicA, "シフトなし職種ガード区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "シフトなし職種")
	makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)
	_ = makeStaffWithOccupation(t, db, clinicA, occA.ID, "シフトなし出勤候補")
	_ = makeStaffWithOccupation(t, db, clinicB, occA.ID, "他院シフトなし")
	_ = makeStaffWithOccupationVisibility(t, db, clinicA, occA.ID, "無効シフトなし", false, true)
	_ = makeStaffWithOccupationVisibility(t, db, clinicA, occA.ID, "非公開シフトなし", true, false)

	t.Run("シフト0件でも候補スタッフを1以上カウントする", func(t *testing.T) {
		result, err := repo.CountWorkingStaffByReservationTypeIDs(ctx, clinicA, rtA.ID, []time.Time{date})
		require.NoError(t, err)
		require.Equal(t, int64(1), result[dateStr], "他院・無効・非公開を除いた1名がシフトなしでも出勤候補")
	})

	t.Run("applyOccupationGuardはシフトなし日をAvailableのまま残す", func(t *testing.T) {
		svc := &liffService{occupationRepo: repo}
		results := []AvailableDateResult{{Date: dateStr, Available: true}}
		require.NoError(t, svc.applyOccupationGuard(ctx, clinicA, rtA.ID, results))
		assert.True(t, results[0].Available)
		assert.Empty(t, results[0].Reason)
	})
}

// TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs_OccupationGuardAllOff
// は該当職種スタッフが全員 off/paid_leave の日は count=0 で applyOccupationGuard が staff_off にすることを固定する。
func TestReservationTypeOccupationRepository_CountWorkingStaffByReservationTypeIDs_OccupationGuardAllOff(t *testing.T) {
	db := setupReservationTypeOccupationTestDB(t)
	repo := NewReservationTypeOccupationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	date := time.Date(2026, 7, 2, 0, 0, 0, 0, config.JST)
	dateStr := date.Format(time.DateOnly)

	rtA := makeReservationTypeLinked(t, db, clinicA, "全員休日職種ガード区分", nil, nil)
	occA := makeOccupation(t, db, clinicA, "全員休日職種")
	makeReservationTypeOccupationLink(t, db, clinicA, rtA.ID, occA.ID)
	offStaff := makeStaffWithOccupation(t, db, clinicA, occA.ID, "休日スタッフ")
	leaveStaff := makeStaffWithOccupation(t, db, clinicA, occA.ID, "有給スタッフ")
	makeShiftEntryWithType(t, db, clinicA, offStaff.ID, date, model.ShiftTypeOff)
	makeShiftEntryWithType(t, db, clinicA, leaveStaff.ID, date, model.ShiftTypePaidLeave)

	t.Run("全員off/paid_leaveの日はcount=0でキーを埋める", func(t *testing.T) {
		result, err := repo.CountWorkingStaffByReservationTypeIDs(ctx, clinicA, rtA.ID, []time.Time{date})
		require.NoError(t, err)
		assert.Equal(t, int64(0), result[dateStr])
		_, ok := result[dateStr]
		assert.True(t, ok, "候補はいるが全員休みの日もキーを0で埋める")
	})

	t.Run("applyOccupationGuardは全員休みの日をstaff_offにする", func(t *testing.T) {
		svc := &liffService{occupationRepo: repo}
		results := []AvailableDateResult{{Date: dateStr, Available: true}}
		require.NoError(t, svc.applyOccupationGuard(ctx, clinicA, rtA.ID, results))
		assert.False(t, results[0].Available)
		assert.Equal(t, "staff_off", results[0].Reason)
	})
}
