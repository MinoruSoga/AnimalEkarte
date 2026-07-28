package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func buildSlotsTestSvc(schedule *mockLiffScheduleRepository, admin *mockLiffAdminRepository) *liffService {
	return newLiffSvc(
		&mockLiffSettingRepository{},
		&mockLiffTypeRepository{},
		&mockLiffStaffRepository{},
		schedule,
		admin,
		&mockLiffCustomerRepository{},
		&mockLiffValidators{},
		&mockLiffNotifier{},
	)
}

// TestBuildStaffSlotInputsForDate は PERF-FOLLOWUP-08 で buildStaffSlotInputs を置き換えた
// buildStaffSlotInputsForDate（1日ウィンドウ prefetch + buildStaffSlotInputsFromWindow）を検証する。
// 旧実装は staff ごとに FindAllByDate/FindAllBreaksByEntryID を呼んでいたため per-staff にエラーを
// 注入できたが、新実装はバッチ1回のプリフェッチのため、staff 単位の差異は「返ってきたバッチ結果に
// その staff のエントリが含まれるか」で表現する（観測可能な出力 = ScheduleOverride の有無は不変）。
func TestBuildStaffSlotInputsForDate(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	staffs := []model.Staff{{ID: 1}, {ID: 2}}

	t.Run("予約prefetch失敗はラップされて返る", func(t *testing.T) {
		admin := &mockLiffAdminRepository{
			findTimeRangesByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]model.Reservation, error) {
				return nil, errors.New("db error")
			},
		}
		svc := buildSlotsTestSvc(&mockLiffScheduleRepository{}, admin)

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, date)

		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("スタッフが空なら空スライスを返す", func(t *testing.T) {
		svc := buildSlotsTestSvc(&mockLiffScheduleRepository{}, &mockLiffAdminRepository{})

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, nil, date)

		assert.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("シフトprefetch失敗は非致命的で全staffScheduleOverrideなしのまま続行", func(t *testing.T) {
		schedule := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return nil, errors.New("shift lookup failed")
			},
		}
		svc := buildSlotsTestSvc(schedule, &mockLiffAdminRepository{})

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, date)

		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Nil(t, got[0].ScheduleOverride)
		assert.Nil(t, got[1].ScheduleOverride)
	})

	t.Run("staff1のみシフトを持つ場合staff2はScheduleOverrideなしのまま設定される", func(t *testing.T) {
		start := "09:00:00"
		end := "18:00:00"
		schedule := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{
					{ID: 99, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartTime: &start, EndTime: &end},
				}, nil
			},
		}
		svc := buildSlotsTestSvc(schedule, &mockLiffAdminRepository{})

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, date)

		require.NoError(t, err)
		require.Len(t, got, 2)
		require.NotNil(t, got[0].ScheduleOverride)
		assert.Equal(t, string(model.ShiftTypeFull), got[0].ScheduleOverride.ShiftType)
		require.NotNil(t, got[0].ScheduleOverride.WorkStart)
		assert.Equal(t, "0900", *got[0].ScheduleOverride.WorkStart)
		require.NotNil(t, got[0].ScheduleOverride.WorkEnd)
		assert.Equal(t, "1800", *got[0].ScheduleOverride.WorkEnd)
		assert.Empty(t, got[0].ScheduleOverride.Breaks)
		assert.Nil(t, got[1].ScheduleOverride, "staff2はバッチ結果にエントリがないためoverrideなし")
	})

	t.Run("休憩prefetch失敗は非致命的でBreaksが空のまま設定される", func(t *testing.T) {
		start := "09:00:00"
		end := "18:00:00"
		schedule := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{
					{ID: 7, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartTime: &start, EndTime: &end},
				}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				return nil, errors.New("breaks lookup failed")
			},
		}
		svc := buildSlotsTestSvc(schedule, &mockLiffAdminRepository{})

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs[:1], date)

		require.NoError(t, err)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].ScheduleOverride)
		assert.Empty(t, got[0].ScheduleOverride.Breaks)
	})

	t.Run("シフト+休憩取得成功でBreaksがHHMM形式に変換される", func(t *testing.T) {
		start := "09:00:00"
		end := "18:00:00"
		schedule := &mockLiffScheduleRepository{
			findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, _, _ time.Time) ([]model.ShiftEntry, error) {
				return []model.ShiftEntry{
					{ID: 7, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartTime: &start, EndTime: &end},
				}, nil
			},
			findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
				return map[uint64][]model.ShiftEntryBreak{
					7: {{BreakStart: "12:00:00", BreakEnd: "13:00:00"}},
				}, nil
			},
		}
		svc := buildSlotsTestSvc(schedule, &mockLiffAdminRepository{})

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs[:1], date)

		require.NoError(t, err)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].ScheduleOverride)
		require.Len(t, got[0].ScheduleOverride.Breaks, 1)
		assert.Equal(t, "1200", got[0].ScheduleOverride.Breaks[0].Start)
		assert.Equal(t, "1300", got[0].ScheduleOverride.Breaks[0].End)
	})

	t.Run("当日予約はキャンセル済みを除外し担当スタッフでフィルタする", func(t *testing.T) {
		staffID1 := uint64(1)
		staffID2 := uint64(2)
		startA := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
		endA := time.Date(2026, 5, 1, 10, 30, 0, 0, time.UTC)
		startB := time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)
		endB := time.Date(2026, 5, 1, 11, 30, 0, 0, time.UTC)
		admin := &mockLiffAdminRepository{
			findTimeRangesByDateRangeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]model.Reservation, error) {
				return []model.Reservation{
					{ID: 1, DoctorID: &staffID1, Status: model.ReservationStatusConfirmed, StartTime: startA, EndTime: endA},
					{ID: 2, DoctorID: &staffID1, Status: model.ReservationStatusCancelled, StartTime: startB, EndTime: endB},
					{ID: 3, DoctorID: &staffID2, Status: model.ReservationStatusConfirmed, StartTime: startB, EndTime: endB},
					{ID: 4, DoctorID: nil, Status: model.ReservationStatusConfirmed, StartTime: startB, EndTime: endB},
				}, nil
			},
		}
		svc := buildSlotsTestSvc(&mockLiffScheduleRepository{}, admin)

		got, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, date)

		require.NoError(t, err)
		require.Len(t, got, 2)
		// staff 1: only the confirmed reservation (id=1), cancelled (id=2) excluded
		require.Len(t, got[0].ExistingResvs, 1)
		assert.Equal(t, "1000", got[0].ExistingResvs[0].StartTime)
		assert.Equal(t, "1030", got[0].ExistingResvs[0].EndTime)
		// staff 2: only the confirmed reservation matching DoctorID (id=3)
		require.Len(t, got[1].ExistingResvs, 1)
		assert.Equal(t, "1100", got[1].ExistingResvs[0].StartTime)
	})
}
