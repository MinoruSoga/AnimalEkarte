package reservation

// reservation_repository_count_batch_test.go — BE-refactor.md R2-4 (D8) の回帰防止。
//
// CountByTypeAndStartTimes は liff_service.FilterSlotsByCapacity の N+1
// （日付ごとの各スロットで CountByTypeAndStartTime を個別発行）を解消するための
// バッチ経路（GROUP BY start_time）。実 DB に対して以下を検証する:
//   - 複数 start_time の件数を1クエリで一括取得できる
//   - cancelled 予約は除外される（CountByTypeAndStartTime と同じフィルタ）
//   - excludeID で自分自身を除外できる
//   - 該当なしの start_time はマップに現れない（0件想定はキー不在で表現）
//   - startTimes が空なら空マップを返す（クエリを発行しない）

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestReservationRepository_CountByTypeAndStartTimes(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	// CountByTypeAndStartTimes は narrow interface (service 層の
	// reservationTypeCapacityBatchCounter) 経由の任意能力として実装されており、公開
	// ReservationRepository インターフェースには含めていない（BE-refactor.md R2-4 の設計判断:
	// 既存の広いインターフェース実装済みモック群への波及を避けるため）。同一パッケージ内なので
	// 具象型に直接キャストして検証する。
	repo := NewReservationRepository(db).(*reservationRepository)
	ctx := context.Background()
	const clinicID = uint64(1)

	rt := makeReservationType(t, db, clinicID)
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nineAM := base.Add(9 * time.Hour)
	tenAM := base.Add(10 * time.Hour)
	elevenAM := base.Add(11 * time.Hour)

	createReservation := func(startTime time.Time, status model.ReservationStatus) *model.Reservation {
		res := &model.Reservation{
			ClinicID:          clinicID,
			StartTime:         startTime,
			EndTime:           startTime.Add(15 * time.Minute),
			ReservationTypeID: rt.ID,
			VisitType:         model.VisitTypeRevisit,
			Status:            status,
			Source:            model.ReservationSourceManual,
			CustomerFields:    json.RawMessage(`{}`),
		}
		require.NoError(t, db.WithContext(ctx).Create(res).Error)
		return res
	}

	// 09:00 に有効な予約2件、10:00 に有効な予約1件+cancelled1件（cancelledは除外される想定）、
	// 11:00 は予約なし。
	createReservation(nineAM, model.ReservationStatusPending)
	createReservation(nineAM, model.ReservationStatusConfirmed)
	createReservation(tenAM, model.ReservationStatusPending)
	createReservation(tenAM, model.ReservationStatusCancelled)

	t.Run("複数start_timeの件数を1クエリで一括取得する（cancelled除外）", func(t *testing.T) {
		counts, err := repo.CountByTypeAndStartTimes(ctx, clinicID, rt.ID, []time.Time{nineAM, tenAM, elevenAM}, nil)

		require.NoError(t, err)
		assert.Equal(t, int64(2), counts[nineAM.Unix()], "09:00は有効2件")
		assert.Equal(t, int64(1), counts[tenAM.Unix()], "10:00は有効1件（cancelled除外）")
		_, elevenExists := counts[elevenAM.Unix()]
		assert.False(t, elevenExists, "該当なしの時刻はマップにキーを持たない")
	})

	t.Run("excludeIDで自分自身を除外する", func(t *testing.T) {
		pendingAt10, err := repo.CountByTypeAndStartTimes(ctx, clinicID, rt.ID, []time.Time{tenAM}, nil)
		require.NoError(t, err)
		require.Equal(t, int64(1), pendingAt10[tenAM.Unix()])

		var pendingID uint64
		require.NoError(t, db.WithContext(ctx).Model(&model.Reservation{}).
			Where("clinic_id = ? AND reservation_type_id = ? AND start_time = ? AND status = ?", clinicID, rt.ID, tenAM, model.ReservationStatusPending).
			Pluck("id", &pendingID).Error)

		counts, err := repo.CountByTypeAndStartTimes(ctx, clinicID, rt.ID, []time.Time{tenAM}, &pendingID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), counts[tenAM.Unix()], "excludeIDで自分自身を除外すると0件")
	})

	t.Run("startTimesが空なら空マップを返しクエリを発行しない", func(t *testing.T) {
		counts, err := repo.CountByTypeAndStartTimes(ctx, clinicID, rt.ID, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, counts)
	})

	t.Run("別clinicの予約は集計に含まれない", func(t *testing.T) {
		otherClinic := uint64(2)
		otherRT := makeReservationType(t, db, otherClinic)
		otherRes := &model.Reservation{
			ClinicID: otherClinic, StartTime: nineAM, EndTime: nineAM.Add(15 * time.Minute),
			ReservationTypeID: otherRT.ID, VisitType: model.VisitTypeRevisit,
			Status: model.ReservationStatusPending, Source: model.ReservationSourceManual,
			CustomerFields: json.RawMessage(`{}`),
		}
		require.NoError(t, db.WithContext(ctx).Create(otherRes).Error)

		counts, err := repo.CountByTypeAndStartTimes(ctx, clinicID, rt.ID, []time.Time{nineAM}, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(2), counts[nineAM.Unix()], "別クリニックの予約が混入しない")
	})
}
