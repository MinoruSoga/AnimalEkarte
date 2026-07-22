package reservation

// reservation_repository_test.go — reservation_repository.go の実 DB 結合テスト（#212）。
//
// 対象: FindAll / ExistsByReservationTypeID / ExistsByStaffID / LockAndFindByID /
//   HasDoctorConflict / CountOnDutyDoctors / CountByTypeAndStartTime /
//   CountByCustomerAndDateRange / CountByDateAndSource / FindNoShowCandidates。
// 異常系・clinic_id 隔離・境界値（0件/複数件）を優先し、正常系の重複でカバレッジを稼がない。
//
// makeReservationType / makeReservation は reservation_clinic_isolation_test.go、
// makeDoctor は isolation_test_helpers_test.go、makeShiftEntry は
// reservation_schedule_clinic_isolation_test.go、makeLineCustomerForAdmin は
// appointment_admin_repository_test.go を再利用する。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

// setupReservationRepoTestDB は reservation_repository.go の全メソッドをカバーするために
// ReservationType/Reservation/Staff/ShiftEntry/LineCustomer を整備する。
// shift_entries は共有テスト DB に残るため TRUNCATE を AutoMigrate より先に実行する
// （reservation_schedule_clinic_isolation_test.go と同じ順序）。
func setupReservationRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := repotest.SetupTestDB(t)
	db.Exec("TRUNCATE TABLE shift_entries CASCADE")
	require.NoError(t, repotest.EnsureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.Staff{}, &model.ShiftEntry{}, &model.LineCustomer{},
	))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE") // appointments も連鎖クリア
	db.Exec("TRUNCATE TABLE staffs CASCADE")            // shift_entries/staff_clinic_assignments も連鎖クリア
	db.Exec("TRUNCATE TABLE line_customers CASCADE")
	return db
}

// makeReservationForReservationRepoTest は開始時刻・ステータス・source・line_customer_id・
// doctor_id を自由に指定できる予約生成ヘルパー（本ファイル専用、他エージェントとの衝突回避のため
// "ForReservationRepoTest" サフィックスを付与）。
func makeReservationForReservationRepoTest(
	t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64,
	start time.Time, status model.ReservationStatus, source model.ReservationSource,
	lineCustomerID, doctorID *uint64,
) *model.Reservation {
	t.Helper()
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		ReservationTypeID: reservationTypeID,
		DoctorID:          doctorID,
		LineCustomerID:    lineCustomerID,
		VisitType:         model.VisitTypeRevisit,
		Status:            status,
		Source:            source,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

func TestReservationRepository_FindAll(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	resA := makeReservation(t, db, clinicA)
	makeReservation(t, db, clinicB)

	t.Run("clinic_id隔離: 指定クリニックのみ返す", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 10, nil, nil, nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, resA.ID, got[0].ID)
	})

	t.Run("空のclinicIDsは0件を返す（フェイルセーフ）", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, []uint64{}, 1, 10, nil, nil, nil, nil, nil, nil, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 0, total)
		assert.Empty(t, got)
	})

	t.Run("statusフィルタで絞り込む", func(t *testing.T) {
		rt := makeReservationType(t, db, clinicA)
		start := time.Now().UTC().Truncate(time.Minute).Add(time.Hour)
		confirmed := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, start,
			model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil)

		statusFilter := string(model.ReservationStatusConfirmed)
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 10, nil, nil, nil, &statusFilter, nil, nil, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, got, 1)
		assert.Equal(t, confirmed.ID, got[0].ID)
	})
}

func TestReservationRepository_ExistsByReservationTypeID(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	usedType := makeReservationType(t, db, clinicA)
	unusedType := makeReservationType(t, db, clinicA)
	makeReservationForReservationRepoTest(t, db, clinicA, usedType.ID, time.Now().UTC(),
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil)

	t.Run("使用中の予約区分はtrue", func(t *testing.T) {
		got, err := repo.ExistsByReservationTypeID(ctx, clinicA, usedType.ID)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("未使用の予約区分はfalse", func(t *testing.T) {
		got, err := repo.ExistsByReservationTypeID(ctx, clinicA, unusedType.ID)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestReservationRepository_ExistsByStaffID(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	assignedDoctor := makeDoctor(t, db, clinicA, "担当医師")
	unassignedDoctor := makeDoctor(t, db, clinicA, "未担当医師")
	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, time.Now().UTC(),
		model.ReservationStatusPending, model.ReservationSourceManual, nil, &assignedDoctor.ID)

	t.Run("予約を担当しているスタッフはtrue", func(t *testing.T) {
		got, err := repo.ExistsByStaffID(ctx, clinicA, assignedDoctor.ID)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("予約を担当していないスタッフはfalse", func(t *testing.T) {
		got, err := repo.ExistsByStaffID(ctx, clinicA, unassignedDoctor.ID)
		require.NoError(t, err)
		assert.False(t, got)
	})
}

func TestReservationRepository_LockAndFindByID(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	resA := makeReservation(t, db, clinicA)

	t.Run("正常系: tx内でFOR UPDATE取得できる", func(t *testing.T) {
		var got *model.Reservation
		err := tx.WithTx(ctx, func(ctx context.Context) error {
			var innerErr error
			got, innerErr = repo.LockAndFindByID(ctx, clinicA, resA.ID)
			return innerErr
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, resA.ID, got.ID)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := tx.WithTx(ctx, func(ctx context.Context) error {
			_, innerErr := repo.LockAndFindByID(ctx, clinicA, 9999999)
			return innerErr
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

func TestReservationRepository_HasDoctorConflict(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	doctor := makeDoctor(t, db, clinicA, "医師A")
	start := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	existing := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, start,
		model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, &doctor.ID)
	checkConflict := func(startTime, endTime time.Time, excludeID *uint64) (bool, error) {
		var conflict bool
		err := tx.WithTx(ctx, func(txCtx context.Context) error {
			var innerErr error
			conflict, innerErr = repo.HasDoctorConflict(txCtx, clinicA, doctor.ID, startTime, endTime, excludeID)
			return innerErr
		})
		return conflict, err
	}

	t.Run("重複する時間枠はtrue", func(t *testing.T) {
		conflict, err := checkConflict(start.Add(15*time.Minute), end.Add(15*time.Minute), nil)
		require.NoError(t, err)
		assert.True(t, conflict)
	})

	t.Run("重複しない時間枠はfalse", func(t *testing.T) {
		conflict, err := checkConflict(end.Add(time.Hour), end.Add(90*time.Minute), nil)
		require.NoError(t, err)
		assert.False(t, conflict)
	})

	t.Run("excludeIDで自身を除外するとfalse", func(t *testing.T) {
		conflict, err := checkConflict(start, end, &existing.ID)
		require.NoError(t, err)
		assert.False(t, conflict)
	})
}

func TestReservationRepository_CountOnDutyDoctors(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	doctor1 := makeDoctor(t, db, clinicA, "医師1")
	doctor2 := makeDoctor(t, db, clinicA, "医師2")
	shiftDate := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	makeShiftEntry(t, db, clinicA, doctor1.ID, shiftDate)
	makeShiftEntry(t, db, clinicA, doctor2.ID, shiftDate)

	// JST 正午（同日）= UTC 03:00。CountOnDutyDoctors は `DATE(? AT TIME ZONE 'Asia/Tokyo')` で
	// JST の日付に変換して比較するため、UTC 日跨ぎを避けるためにこの時刻を選ぶ。
	queryDate := time.Date(2026, 7, 20, 3, 0, 0, 0, time.UTC)

	t.Run("出勤医師数を返す（複数件）", func(t *testing.T) {
		count, err := repo.CountOnDutyDoctors(ctx, clinicA, queryDate)
		require.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})

	t.Run("シフトの無い日は0件", func(t *testing.T) {
		count, err := repo.CountOnDutyDoctors(ctx, clinicA, queryDate.AddDate(0, 0, 1))
		require.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})
}

func TestReservationRepository_CountByTypeAndStartTime(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	startTime := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)

	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, startTime,
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil)
	res2 := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, startTime,
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil)
	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, startTime,
		model.ReservationStatusCancelled, model.ReservationSourceManual, nil, nil)

	t.Run("同一区分・同一開始時刻はキャンセル除いてカウント（複数件）", func(t *testing.T) {
		count, err := repo.CountByTypeAndStartTime(ctx, clinicA, rt.ID, startTime, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 2, count)
	})

	t.Run("excludeIDを除外してカウント", func(t *testing.T) {
		count, err := repo.CountByTypeAndStartTime(ctx, clinicA, rt.ID, startTime, &res2.ID)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("開始時刻が異なれば0件", func(t *testing.T) {
		count, err := repo.CountByTypeAndStartTime(ctx, clinicA, rt.ID, startTime.Add(time.Hour), nil)
		require.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})
}

func TestReservationRepository_CountByCustomerAndDateRange(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	lc := makeLineCustomerForAdmin(t, db, clinicA, "line-user-repo-test")

	inRange := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	outOfRange := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	rangeStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	rangeEnd := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, inRange,
		model.ReservationStatusPending, model.ReservationSourceLine, &lc.ID, nil)
	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, outOfRange,
		model.ReservationStatusPending, model.ReservationSourceLine, &lc.ID, nil)
	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, inRange,
		model.ReservationStatusCancelled, model.ReservationSourceLine, &lc.ID, nil)

	t.Run("期間内・キャンセル除くカウント", func(t *testing.T) {
		count, err := repo.CountByCustomerAndDateRange(ctx, clinicA, lc.ID, rangeStart, rangeEnd)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("該当なしは0件", func(t *testing.T) {
		count, err := repo.CountByCustomerAndDateRange(ctx, clinicA, lc.ID,
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})
}

func TestReservationRepository_CountByDateAndSource(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	// JST 2026-07-15 12:00 = UTC 2026-07-15 03:00
	dayStart := time.Date(2026, 7, 15, 3, 0, 0, 0, time.UTC)

	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, dayStart,
		model.ReservationStatusPending, model.ReservationSourceLine, nil, nil)
	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, dayStart.Add(2*time.Hour),
		model.ReservationStatusPending, model.ReservationSourceManual, nil, nil)

	t.Run("日付・sourceが一致する件数を返す", func(t *testing.T) {
		count, err := repo.CountByDateAndSource(ctx, clinicA, dayStart, model.ReservationSourceLine)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count)
	})

	t.Run("該当日に予約が無ければ0件", func(t *testing.T) {
		count, err := repo.CountByDateAndSource(ctx, clinicA, dayStart.AddDate(0, 0, 1), model.ReservationSourceLine)
		require.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})
}

func TestReservationRepository_FindNoShowCandidates(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	now := time.Now().UTC()
	oldEnoughEnd := now.Add(-5 * time.Hour)
	tooRecentEnd := now.Add(-1 * time.Hour)

	// makeReservationForReservationRepoTest は EndTime = start + 30分 で作成するため、
	// start を「望む end_time - 30分」に設定する。
	candidate := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, oldEnoughEnd.Add(-30*time.Minute),
		model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil)
	notOldEnough := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, tooRecentEnd.Add(-30*time.Minute),
		model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil)
	cancelledOld := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, oldEnoughEnd.Add(-30*time.Minute),
		model.ReservationStatusCancelled, model.ReservationSourceManual, nil, nil)
	finalized := makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, oldEnoughEnd.Add(-30*time.Minute),
		model.ReservationStatusConfirmed, model.ReservationSourceManual, nil, nil)
	record := &model.MedicalRecord{
		ClinicID:      clinicA,
		RecordNo:      "MR-NO-SHOW-CANDIDATE-EXCLUSION",
		Date:          now,
		AppointmentID: &finalized.ID,
		Status:        model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.Create(record).Error)

	t.Run("4時間以上経過したconfirmed/pendingのみ該当", func(t *testing.T) {
		got, err := repo.FindNoShowCandidates(ctx, clinicA)
		require.NoError(t, err)
		ids := make([]uint64, 0, len(got))
		for _, r := range got {
			ids = append(ids, r.ID)
		}
		assert.Contains(t, ids, candidate.ID)
		assert.NotContains(t, ids, notOldEnough.ID, "4時間未満はノーショウ候補外のはず")
		assert.NotContains(t, ids, cancelledOld.ID, "cancelledはノーショウ候補外のはず")
		assert.NotContains(t, ids, finalized.ID, "確定済みカルテがある予約はノーショウ候補外のはず")
	})

	t.Run("該当データが無いクリニックは空を返す", func(t *testing.T) {
		got, err := repo.FindNoShowCandidates(ctx, uint64(999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}
