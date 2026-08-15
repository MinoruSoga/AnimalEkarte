package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockCapacityCounter struct {
	countFn func(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
}

func (m mockCapacityCounter) CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error) {
	return m.countFn(ctx, clinicID, reservationTypeID, startTime, excludeID)
}

func TestFilterApplicableAvailableSlots_WeeklyAndSpecific(t *testing.T) {
	monday := int8(1)
	tuesday := int8(2)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST)

	result := FilterApplicableAvailableSlots([]model.ReservationTypeAvailableSlot{
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &tuesday, StartTime: "12:30", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "14:00", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "15:00", IsActive: false},
	}, date)

	assert.Len(t, result, 2)
	assert.Equal(t, "09:45", result[0].StartTime)
	assert.Equal(t, "14:00", result[1].StartTime)
}

// TestMergeAvailableTimeSlots: 予約可能枠は加算モード（ホワイトリストなし）を検証する。
// 営業時間内の登録時刻は重複追加されず、営業時間外の登録時刻は追加される。
func TestMergeAvailableTimeSlots(t *testing.T) {
	monday := int8(1)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST) // Monday

	t.Run("営業時間外の時刻のみ追加される", func(t *testing.T) {
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
			{StartTime: "0945", EndTime: "1045"},
		}
		availableSlots := []model.ReservationTypeAvailableSlot{
			// 0945 は既存スロットと重複 → 追加されない
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
			// 0800 は営業時間外 → 追加される
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "08:00", IsActive: true},
		}
		result := MergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Len(t, result, 3)
		assert.Equal(t, "0800", result[0].StartTime)
		assert.Equal(t, "0900", result[1].StartTime) // end = 0800 + 60min = 0900
		assert.Equal(t, "0900", result[0].EndTime)
		assert.Equal(t, "0945", result[2].StartTime)
	})

	t.Run("該当日に枠なしでもスロットはブロックされない", func(t *testing.T) {
		tuesday := int8(2)
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
			{StartTime: "0945", EndTime: "1045"},
		}
		// 火曜日登録の枠を月曜日に適用 → 該当なし → slots がそのまま返る
		availableSlots := []model.ReservationTypeAvailableSlot{
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &tuesday, StartTime: "10:00", IsActive: true},
		}
		result := MergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Equal(t, slots, result)
	})

	t.Run("有効でない枠は無視される", func(t *testing.T) {
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
		}
		availableSlots := []model.ReservationTypeAvailableSlot{
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "08:00", IsActive: false},
		}
		result := MergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Equal(t, slots, result)
	})
}

func TestFilterSlotsByCapacity(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST)
	slots := []TimeSlot{
		{StartTime: "0900", EndTime: "0930"},
		{StartTime: "1000", EndTime: "1030"},
	}

	t.Run("上限未満のスロットは含まれる", func(t *testing.T) {
		result, err := FilterSlotsByCapacity(context.Background(), slots, mockCapacityCounter{
			countFn: func(_ context.Context, _, _ uint64, _ time.Time, _ *uint64) (int64, error) {
				return 1, nil
			},
		}, 1, 2, date, 2)

		assert.NoError(t, err)
		assert.Equal(t, slots, result)
	})

	t.Run("上限以上のスロットは除外される", func(t *testing.T) {
		result, err := FilterSlotsByCapacity(context.Background(), slots, mockCapacityCounter{
			countFn: func(_ context.Context, _, _ uint64, start time.Time, _ *uint64) (int64, error) {
				if start.Hour() == 9 {
					return 2, nil
				}
				return 0, nil
			},
		}, 1, 2, date, 2)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "1000", result[0].StartTime)
	})
}

// mockBatchCapacityCounter は reservationTypeCapacityCounter + reservationTypeCapacityBatchCounter
// を両方満たす（BE-refactor.md R2-4 / D8: N+1 解消のバッチ経路検証用）。
type mockBatchCapacityCounter struct {
	mockCapacityCounter
	batchCallCount int
	lastStartTimes []time.Time
	batchFn        func(ctx context.Context, clinicID, reservationTypeID uint64, startTimes []time.Time, excludeID *uint64) (map[int64]int64, error)
}

func (m *mockBatchCapacityCounter) CountByTypeAndStartTimes(ctx context.Context, clinicID, reservationTypeID uint64, startTimes []time.Time, excludeID *uint64) (map[int64]int64, error) {
	m.batchCallCount++
	m.lastStartTimes = startTimes
	return m.batchFn(ctx, clinicID, reservationTypeID, startTimes, excludeID)
}

// TestFilterSlotsByCapacity_BatchPath は repo が reservationTypeCapacityBatchCounter を実装する
// 場合に、日付ごとの全スロットが単一クエリで処理され（N+1解消）、per-slot 経路と同一のフィルタ結果に
// なることを固定する（BE-refactor.md R2-4 / D8）。
func TestFilterSlotsByCapacity_BatchPath(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, config.JST)
	slots := []TimeSlot{
		{StartTime: "0900", EndTime: "0930"},
		{StartTime: "1000", EndTime: "1030"},
		{StartTime: "1100", EndTime: "1130"},
	}

	t.Run("1クエリで全スロットを判定し、per-slotと同じ結果を返す", func(t *testing.T) {
		nineAM := time.Date(2026, 6, 1, 9, 0, 0, 0, config.JST)
		mock := &mockBatchCapacityCounter{
			batchFn: func(_ context.Context, _, _ uint64, _ []time.Time, _ *uint64) (map[int64]int64, error) {
				// 09:00 のみ満員（2件）、他は 0 件
				return map[int64]int64{nineAM.Unix(): 2}, nil
			},
		}

		result, err := FilterSlotsByCapacity(context.Background(), slots, mock, 1, 2, date, 2)

		assert.NoError(t, err)
		require.Equal(t, 1, mock.batchCallCount, "日付ごとに1クエリのみ発行する（N+1解消の直接証明）")
		require.Len(t, mock.lastStartTimes, 3, "その日の全スロット分の startTime を一括で渡す")
		require.Len(t, result, 2, "満員の09:00のみ除外される")
		assert.Equal(t, "1000", result[0].StartTime)
		assert.Equal(t, "1100", result[1].StartTime)
	})

	t.Run("バッチクエリのエラーは伝播する", func(t *testing.T) {
		mock := &mockBatchCapacityCounter{
			batchFn: func(_ context.Context, _, _ uint64, _ []time.Time, _ *uint64) (map[int64]int64, error) {
				return nil, assert.AnError
			},
		}

		_, err := FilterSlotsByCapacity(context.Background(), slots, mock, 1, 2, date, 2)
		assert.Error(t, err)
	})
}
