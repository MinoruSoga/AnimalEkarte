package reservation

// reservation_booking_lock_concurrency_test.go — BE-refactor.md Appendix A X-9
// (`resv-slot-phantom-toctou`) の実DB並行性証明。
//
// 背景: CountConflicts（SELECT ... FOR UPDATE）は条件に合致する既存行が 0 件（空枠）の場合は
// 何もロックしない。空き枠に対する同時2リクエストは双方とも「空き」と判定し双方 INSERT
// できてしまう（ファントム）。AcquireBookingLock は clinic 単位の pg_advisory_xact_lock で
// この check-then-act を直列化する。
//
// 検証方法: 同一 clinic・同一時間枠に対し、AcquireBookingLock → CountConflicts → Create
// という reservation_service.go の実フローと同型の手順を 2 goroutine で同時実行し、
// ちょうど 1 件だけ成功する（他方は Conflict）ことを実 DB のロック挙動で証明する。

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// bookAttempt は reservation_service.go の Create/ValidateAndCreate が WithTx 内で踏む
// AcquireBookingLock → 競合チェック → Create の実フローを repository 層だけで再現する
// （advisory lock 導入前は両者成功しうる、導入後は直列化される）。
func bookAttempt(ctx context.Context, tx Transactor, repo ReservationRepository, clinicID uint64, start, end time.Time, reservationTypeID uint64) error {
	return tx.WithTx(ctx, func(ctx context.Context) error {
		if err := repo.AcquireBookingLock(ctx, clinicID); err != nil {
			return err
		}
		count, err := repo.CountConflicts(ctx, clinicID, start, end, nil)
		if err != nil {
			return err
		}
		if count > 0 {
			return apperrors.WrapConflict("この時間枠は既に予約が入っています")
		}
		return repo.Create(ctx, &model.Reservation{
			ClinicID:          clinicID,
			StartTime:         start,
			EndTime:           end,
			ReservationTypeID: reservationTypeID,
			VisitType:         model.VisitTypeRevisit,
			Status:            model.ReservationStatusConfirmed,
			Source:            model.ReservationSourceManual,
			CustomerFields:    []byte(`{}`),
		})
	})
}

// TestReservationRepository_AcquireBookingLock_SerializesPhantomBooking は X-9 の受け入れ条件:
// 空き枠に対する同時2リクエストのうち、advisory lock 導入後はちょうど1件だけ成功し、
// 他方は Conflict になる（ファントム重複予約が発生しない）ことを検証する。
func TestReservationRepository_AcquireBookingLock_SerializesPhantomBooking(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	const clinicID = uint64(90001)
	rt := makeReservationType(t, db, clinicID)

	repo := NewReservationRepository(db)
	tx := testNewTransactor(db)

	start := time.Date(2027, 6, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)

	const attempts = 2
	results := make([]error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := range attempts {
		go func(idx int) {
			defer wg.Done()
			results[idx] = bookAttempt(context.Background(), tx, repo, clinicID, start, end, rt.ID)
		}(i)
	}
	wg.Wait()

	successCount, conflictCount := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successCount++
		case apperrors.IsConflict(err):
			conflictCount++
		default:
			t.Fatalf("unexpected error from concurrent booking attempt: %v", err)
		}
	}
	assert.Equal(t, 1, successCount, "ちょうど1件だけ成功するべき（ファントム重複予約防止）")
	assert.Equal(t, 1, conflictCount, "もう1件は Conflict になるべき")

	var persisted int64
	require.NoError(t, db.Model(&model.Reservation{}).
		Where("clinic_id = ? AND start_time = ? AND end_time = ?", clinicID, start, end).
		Count(&persisted).Error)
	assert.Equal(t, int64(1), persisted, "DBに保存された予約は1件のみであるべき")
}
