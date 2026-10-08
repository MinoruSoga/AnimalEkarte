package reservation

// reservation_tentative_exclusion_test.go — EMR-308: tentative（次回予定）は
// 枠・定員・競合・予約上限カウント・LINE 予約一覧を消費しないことを実 DB で固定する。
//
// 対象リーダ（ticket 採用方針「tentative は枠・定員を消費しない」/ Q2「LINE 予約一覧: 非表示」）:
//   - reservation_repository_slots.go: HasDoctorConflict / CountConflicts
//   - reservation_repository_queries.go: CountByTypeAndStartTime / CountByTypeAndStartTimes /
//     CountByCustomerAndDateRange
//   - appointment_admin_repository.go: FindAllByCustomerID（LIFF my-reservations）
//
// tentative 行は status='tentative' + reservation_route='record_intent' で直接作る
// （SetNextVisitIntent は doctor_id/line_customer_id を持たないため、除外フィルタ単体を
// 検証するには生 insert が必要 — doctor/lc を持つ tentative でも必ず除外されることを示す）。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// makeTentativeForExclusionTest は status='tentative' + reservation_route='record_intent' の
// tentative 行を start/end・doctor・line_customer 指定で1件作成する。
// intended_date は実経路（SetNextVisitIntent）と同じく start の暦日を保持する。
func makeTentativeForExclusionTest(
	t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64,
	start, end time.Time, doctorID, lineCustomerID *uint64,
) *model.Reservation {
	t.Helper()
	route := "record_intent"
	intended := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           end,
		ReservationTypeID: reservationTypeID,
		DoctorID:          doctorID,
		LineCustomerID:    lineCustomerID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusTentative,
		Source:            model.ReservationSourceManual,
		ReservationRoute:  &route,
		IntendedDate:      &intended,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

// TestTentativeExclusion_SlotAndCapacityCounts は tentative 行が予約枠占有・
// 時間枠競合・区分定員・顧客予約上限の各カウントへ計上されないことを検証する。
//
// 配置: confirmed 実予約は slotStart（doctor/lc 属性付き）、tentative は翌日の同時刻帯に
// 置く。tentative のみが重なる窓で各リーダを呼び、tentative が計上されないことを示す。
func TestTentativeExclusion_SlotAndCapacityCounts(t *testing.T) {
	db := setupReservationRepoTestDB(t)
	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	doctor := makeDoctor(t, db, clinicA, "tentative除外検証医")
	lc := makeLineCustomerForAdmin(t, db, clinicA, "line-tentative-exclusion")

	slotStart := time.Date(2030, 8, 10, 10, 0, 0, 0, time.UTC)
	slotEnd := slotStart.Add(30 * time.Minute)
	tentativeStart := slotStart.Add(24 * time.Hour)
	tentativeEnd := slotEnd.Add(24 * time.Hour)

	makeReservationForReservationRepoTest(t, db, clinicA, rt.ID, slotStart,
		model.ReservationStatusConfirmed, model.ReservationSourceLine, &lc.ID, &doctor.ID)
	makeTentativeForExclusionTest(t, db, clinicA, rt.ID, tentativeStart, tentativeEnd, &doctor.ID, &lc.ID)

	t.Run("HasDoctorConflict ignores a tentative row overlapping the doctor window", func(t *testing.T) {
		// 基準: confirmed が重なる枠は競合あり。
		var baseline bool
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			baseline, err = repo.HasDoctorConflict(txCtx, clinicA, doctor.ID, slotStart, slotEnd, nil)
			return err
		}))
		assert.True(t, baseline)

		// tentative のみが重なる枠は競合なし。
		var tentativeOnly bool
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			tentativeOnly, err = repo.HasDoctorConflict(txCtx, clinicA, doctor.ID, tentativeStart, tentativeEnd, nil)
			return err
		}))
		assert.False(t, tentativeOnly, "tentative 行は医師時間枠の競合として数えない")
	})

	t.Run("CountConflicts ignores a tentative row overlapping the window", func(t *testing.T) {
		var baseline int64
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			baseline, err = repo.CountConflicts(txCtx, clinicA, slotStart, slotEnd, nil)
			return err
		}))
		assert.EqualValues(t, 1, baseline, "confirmed のみ計上")

		var tentativeOnly int64
		require.NoError(t, tx.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			tentativeOnly, err = repo.CountConflicts(txCtx, clinicA, tentativeStart, tentativeEnd, nil)
			return err
		}))
		assert.EqualValues(t, 0, tentativeOnly, "tentative のみが重なる枠は 0 件")
	})

	t.Run("CountByTypeAndStartTime ignores tentative at the same slot start", func(t *testing.T) {
		count, err := repo.CountByTypeAndStartTime(ctx, clinicA, rt.ID, tentativeStart, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 0, count, "tentative のみの開始時刻は 0 件")

		baseline, err := repo.CountByTypeAndStartTime(ctx, clinicA, rt.ID, slotStart, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 1, baseline, "confirmed は従来どおり計上")
	})

	t.Run("CountByTypeAndStartTimes ignores tentative at the same slot start", func(t *testing.T) {
		concrete, ok := repo.(*reservationRepository)
		require.True(t, ok)
		counts, err := concrete.CountByTypeAndStartTimes(ctx, clinicA, rt.ID,
			[]time.Time{slotStart, tentativeStart}, nil)
		require.NoError(t, err)
		assert.EqualValues(t, 1, counts[slotStart.Unix()], "confirmed は従来どおり計上")
		assert.EqualValues(t, 0, counts[tentativeStart.Unix()], "tentative のみの開始時刻は 0 件")
	})

	t.Run("CountByCustomerAndDateRange ignores tentative in the customer window", func(t *testing.T) {
		rangeStart := time.Date(2030, 8, 1, 0, 0, 0, 0, time.UTC)
		rangeEnd := time.Date(2030, 9, 1, 0, 0, 0, 0, time.UTC)
		count, err := repo.CountByCustomerAndDateRange(ctx, clinicA, lc.ID, rangeStart, rangeEnd)
		require.NoError(t, err)
		assert.EqualValues(t, 1, count, "confirmed のみ計上（tentative は予約上限カウントに含めない）")
	})
}

// TestTentativeExclusion_FindAllByCustomerID は LIFF my-reservations 相当の
// FindAllByCustomerID が tentative 行を返さないことを検証する（Q2: LINE 予約一覧は非表示）。
func TestTentativeExclusion_FindAllByCustomerID(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	customer := makeLineCustomerForAdmin(t, db, clinicA, "U-tentative-hidden")
	rt := makeReservationType(t, db, clinicA)

	real := makeAdminReservationAt(t, db, clinicA, time.Date(2030, 8, 10, 9, 0, 0, 0, time.UTC), nil, nil, nil, &customer.ID)
	tentative := makeTentativeForExclusionTest(t, db, clinicA, rt.ID,
		time.Date(2030, 8, 15, 0, 0, 0, 0, time.UTC), time.Date(2030, 8, 15, 0, 15, 0, 0, time.UTC), nil, &customer.ID)

	t.Run("tentative row is absent from the LIFF customer list", func(t *testing.T) {
		got, err := repo.FindAllByCustomerID(ctx, clinicA, customer.ID)
		require.NoError(t, err)
		ids := adminReservationIDs(got)
		assert.Contains(t, ids, real.ID)
		assert.NotContains(t, ids, tentative.ID, "tentative（次回予定）は LINE 予約一覧に出さない")
	})
}
