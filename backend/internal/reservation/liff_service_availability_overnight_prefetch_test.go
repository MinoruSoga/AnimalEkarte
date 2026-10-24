package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// overlapAwareScheduleRepo は FindAllByStaffIDsAndDateRange の実クエリと同じ選択規則を再現するモックを返す。
// 勤務日が [from, to) の暦日内にある行、または明示日時 [starts_at, ends_at) が [from, to) と重なる行を返す。
// 手で索引を組み替えず、取得 → 索引化 → 評価日ごとの構築という prefetch 経路をそのまま通すために使う。
func overlapAwareScheduleRepo(entries []model.ShiftEntry, breaks map[uint64][]model.ShiftEntryBreak) *mockLiffScheduleRepository {
	return &mockLiffScheduleRepository{
		findAllByStaffIDsAndDateRangeFn: func(_ context.Context, _ uint64, _ []uint64, from, to time.Time) ([]model.ShiftEntry, error) {
			fromDate, toDate := from.Format(time.DateOnly), to.Format(time.DateOnly)
			got := make([]model.ShiftEntry, 0, len(entries))
			for _, e := range entries {
				workDate := e.Date.Format(time.DateOnly)
				dateInRange := workDate >= fromDate && workDate < toDate
				overlaps := e.StartsAt != nil && e.EndsAt != nil && e.StartsAt.Before(to) && e.EndsAt.After(from)
				if dateInRange || overlaps {
					got = append(got, e)
				}
			}
			return got, nil
		},
		findAllBreaksByEntryIDsFn: func(_ context.Context, _ uint64, entryIDs []uint64) (map[uint64][]model.ShiftEntryBreak, error) {
			got := make(map[uint64][]model.ShiftEntryBreak, len(entryIDs))
			for _, id := range entryIDs {
				if b, ok := breaks[id]; ok {
					got[id] = b
				}
			}
			return got, nil
		},
	}
}

// overnightSlotsFor は営業時間 09:00-18:00・1 時間刻みで枠を計算する。勤務窓があればそちらが優先される。
func overnightSlotsFor(t *testing.T, inputs []StaffSlotInput) []TimeSlot {
	t.Helper()
	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  60,
		IntervalMinutes: 60,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})
	require.NoError(t, err)
	return slots
}

// EMR-549 A5-1 / Codex P1: 日跨ぎ勤務の秒付き境界を、評価日ごとに当日分だけ切り出して丸める。
// 勤務行の date は初日のまま取得される（実際の prefetch 経路を通し、2 日目の索引は手で組み替えない）。
func TestBuildStaffSlotInputsForDate_ClipsOvernightShiftSecondsPerEvaluatedDay(t *testing.T) {
	// Arrange: 8/1 22:00:30 から 8/2 07:00:30 の勤務。休憩は 8/1 23:00:30-23:30:30 と 8/2 01:00:30-02:00:30。
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	entry := model.ShiftEntry{ID: 10, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00:30"), EndsAt: rsvTS(t, nextDay, "07:00:30")}
	breaks := map[uint64][]model.ShiftEntryBreak{
		10: {rsvBreak(t, date, "23:00:30", "23:30:30"), rsvBreak(t, nextDay, "01:00:30", "02:00:30")},
	}

	tests := []struct {
		name       string
		evaluated  time.Time
		wantStart  string
		wantEnd    string
		wantBreaks []BreakPeriod
	}{
		{"first day keeps the tail of the day", date, "2201", "2359", []BreakPeriod{{Start: "2300", End: "2331"}}},
		{"second day starts at midnight", nextDay, "0000", "0700", []BreakPeriod{{Start: "0100", End: "0201"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := buildSlotsTestSvc(overlapAwareScheduleRepo([]model.ShiftEntry{entry}, breaks), &mockLiffAdminRepository{})

			// Act
			inputs, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, tt.evaluated)

			// Assert
			require.NoError(t, err)
			require.Len(t, inputs, 1)
			override := inputs[0].ScheduleOverride
			require.NotNil(t, override)
			require.NotNil(t, override.WorkStart)
			require.NotNil(t, override.WorkEnd)
			assert.Equal(t, tt.wantStart, *override.WorkStart)
			assert.Equal(t, tt.wantEnd, *override.WorkEnd)
			assert.Equal(t, tt.wantBreaks, override.Breaks)
		})
	}
}

// 複数日パス: 予約受付期間の初日に、窓の前日から続く夜勤の続きと休憩が反映される（取得窓より前の勤務日の行）。
func TestBuildAvailableDatesStaffInputsFn_OvernightShiftFromBeforeTheWindowReachesItsFirstDay(t *testing.T) {
	// Arrange
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}}
	settings := AvailableDatesSettings{BookingWindowMinDays: 1, BookingWindowMaxDays: 2}
	minDate, _ := BookingWindowDates(settings)
	prevDay := minDate.AddDate(0, 0, -1)
	entry := model.ShiftEntry{ID: 21, StaffID: 1, Date: prevDay, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, prevDay, "22:00"), EndsAt: rsvTS(t, minDate, "07:00")}
	breaks := map[uint64][]model.ShiftEntryBreak{21: {rsvBreak(t, minDate, "01:00", "02:00")}}
	svc := &liffService{scheduleRepo: overlapAwareScheduleRepo([]model.ShiftEntry{entry}, breaks), adminRepo: &mockLiffAdminRepository{}}

	fn, err := svc.buildAvailableDatesStaffInputsFn(ctx, 1, staffs, settings)
	require.NoError(t, err)

	// Act
	first, err := fn(ctx, minDate, 0, 0)
	require.NoError(t, err)
	following, err := fn(ctx, minDate.AddDate(0, 0, 1), 0, 0)
	require.NoError(t, err)

	// Assert
	require.Len(t, first, 1)
	override := first[0].ScheduleOverride
	require.NotNil(t, override)
	require.NotNil(t, override.WorkStart)
	require.NotNil(t, override.WorkEnd)
	assert.Equal(t, "0000", *override.WorkStart)
	assert.Equal(t, "0700", *override.WorkEnd)
	assert.Equal(t, []BreakPeriod{{Start: "0100", End: "0200"}}, override.Breaks)
	require.Len(t, following, 1)
	assert.Nil(t, following[0].ScheduleOverride, "夜勤は 07:00 で終わり、その翌日には掛からない")
}

// 前日からの夜勤の続き（00:00-07:00）と当日の勤務（09:00-18:00）は別々の勤務窓になり、07:00-09:00 は勤務可能にならない。
func TestBuildStaffSlotInputsForDate_CombinesOvernightTailAndOwnShift(t *testing.T) {
	// Arrange: 8/1 22:00-8/2 07:00 の夜勤（休憩 8/2 01:00-02:00）と 8/2 09:00-18:00 の勤務（休憩 12:00-13:00）。
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	night := model.ShiftEntry{ID: 10, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00"), EndsAt: rsvTS(t, nextDay, "07:00")}
	day := model.ShiftEntry{ID: 11, StaffID: 1, Date: nextDay, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, nextDay, "09:00"), EndsAt: rsvTS(t, nextDay, "18:00")}
	breaks := map[uint64][]model.ShiftEntryBreak{
		10: {rsvBreak(t, nextDay, "01:00", "02:00")},
		11: {rsvBreak(t, nextDay, "12:00", "13:00")},
	}
	svc := buildSlotsTestSvc(overlapAwareScheduleRepo([]model.ShiftEntry{night, day}, breaks), &mockLiffAdminRepository{})

	// Act
	inputs, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, nextDay)

	// Assert
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	override := inputs[0].ScheduleOverride
	require.NotNil(t, override)
	assert.Equal(t, []WorkPeriod{{Start: "0000", End: "0700"}, {Start: "0900", End: "1800"}}, override.WorkWindows)
	assert.ElementsMatch(t, []BreakPeriod{{Start: "0100", End: "0200"}, {Start: "1200", End: "1300"}}, override.Breaks)
	want := []TimeSlot{
		{StartTime: "0000", EndTime: "0100"},
		{StartTime: "0200", EndTime: "0300"},
		{StartTime: "0300", EndTime: "0400"},
		{StartTime: "0400", EndTime: "0500"},
		{StartTime: "0500", EndTime: "0600"},
		{StartTime: "0600", EndTime: "0700"},
		{StartTime: "0900", EndTime: "1000"},
		{StartTime: "1000", EndTime: "1100"},
		{StartTime: "1100", EndTime: "1200"},
		{StartTime: "1300", EndTime: "1400"},
		{StartTime: "1400", EndTime: "1500"},
		{StartTime: "1500", EndTime: "1600"},
		{StartTime: "1600", EndTime: "1700"},
		{StartTime: "1700", EndTime: "1800"},
	}
	assert.Equal(t, want, overnightSlotsFor(t, inputs), "07:00-09:00 は勤務窓の間なので枠が出ない")
}

// 同じ日の off の行は、前日からの夜勤の続きを打ち消さない。ShiftType が off のままだと勤務が消えて staff_off と誤判定される。
func TestBuildStaffSlotInputsForDate_OvernightTailIsNotReplacedByAnOffEntry(t *testing.T) {
	// Arrange: 8/1 22:00-8/2 07:00 の夜勤と、8/2 の off（終日非勤務の行は日時を持たない）。
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	night := model.ShiftEntry{ID: 10, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00"), EndsAt: rsvTS(t, nextDay, "07:00")}
	off := model.ShiftEntry{ID: 11, StaffID: 1, Date: nextDay, ShiftType: model.ShiftTypeOff}
	svc := buildSlotsTestSvc(overlapAwareScheduleRepo([]model.ShiftEntry{night, off}, nil), &mockLiffAdminRepository{})

	// Act
	inputs, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, nextDay)

	// Assert
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	override := inputs[0].ScheduleOverride
	require.NotNil(t, override)
	assert.Equal(t, string(model.ShiftTypeFull), override.ShiftType)
	require.NotNil(t, override.WorkStart)
	require.NotNil(t, override.WorkEnd)
	assert.Equal(t, "0000", *override.WorkStart)
	assert.Equal(t, "0700", *override.WorkEnd)
	assert.Empty(t, override.WorkWindows)
	slots := overnightSlotsFor(t, inputs)
	require.Len(t, slots, 7)
	assert.Equal(t, "0000", slots[0].StartTime)
	assert.Equal(t, "0700", slots[len(slots)-1].EndTime)
}

// 評価日に掛かる行のどれかが interval_state=unknown なら、その日はスタッフを候補から外す（当日の既知の行があっても）。
func TestBuildStaffSlotInputsForDate_ExcludesStaffWhenAnOverlappingEntryIsUnknown(t *testing.T) {
	// Arrange: 夜勤の休憩が日時 NULL（unknown）。スタッフ 1 は 8/2 に既知の勤務も持つ。スタッフ 2 は勤務未登録。
	ctx := context.Background()
	staffs := []model.Staff{{ID: 1}, {ID: 2}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	night := model.ShiftEntry{ID: 10, StaffID: 1, Date: date, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00"), EndsAt: rsvTS(t, nextDay, "07:00")}
	day := model.ShiftEntry{ID: 11, StaffID: 1, Date: nextDay, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, nextDay, "09:00"), EndsAt: rsvTS(t, nextDay, "18:00")}
	breaks := map[uint64][]model.ShiftEntryBreak{10: {{}}}
	svc := buildSlotsTestSvc(overlapAwareScheduleRepo([]model.ShiftEntry{night, day}, breaks), &mockLiffAdminRepository{})

	// Act
	inputs, err := svc.buildStaffSlotInputsForDate(ctx, 1, staffs, nextDay)

	// Assert
	require.NoError(t, err)
	require.Len(t, inputs, 1, "unknown の行が掛かるスタッフ 1 は候補外、勤務未登録のスタッフ 2 は残る")
	assert.Equal(t, uint64(2), inputs[0].StaffID)
}

func TestIndexShiftEntriesByDate_CoversEveryOverlappedDayWithinTheWindow(t *testing.T) {
	// Arrange: スタッフ 1 は 8/1 22:00 から 8/4 07:00 の長い勤務、スタッフ 2 は 8/1 22:00 から 8/2 00:00 ちょうどまで。
	d1 := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	long := model.ShiftEntry{ID: 1, StaffID: 1, Date: d1, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, d1, "22:00"), EndsAt: rsvTS(t, d1.AddDate(0, 0, 3), "07:00")}
	endsAtMidnight := model.ShiftEntry{ID: 2, StaffID: 2, Date: d1, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, d1, "22:00"), EndsAt: rsvTS(t, d1.AddDate(0, 0, 1), "00:00")}
	windowStart := *rsvTS(t, d1.AddDate(0, 0, 1), "00:00")
	windowEnd := *rsvTS(t, d1.AddDate(0, 0, 3), "00:00")

	// Act
	index := indexShiftEntriesByDate([]model.ShiftEntry{long, endsAtMidnight}, windowStart, windowEnd)

	// Assert
	require.Len(t, index["2026-08-02"][1], 1)
	require.Len(t, index["2026-08-03"][1], 1)
	assert.Empty(t, index["2026-08-04"], "評価窓の外の日は索引化しない")
	assert.Empty(t, index["2026-08-02"][2], "ちょうど 0:00 で終わる勤務は翌日へ掛からない")
	assert.Len(t, index["2026-08-01"][2], 1, "勤務日の下には窓の外でも登録する")
}

func TestGenerateTimeSlots_InvalidWorkWindowIsAnError(t *testing.T) {
	// Arrange: 解析できない勤務窓は勤務可能として扱わない。
	input := &TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  60,
		IntervalMinutes: 60,
		Mode:            "allow_gaps",
		Staffs: []StaffSlotInput{{
			StaffID:          1,
			ScheduleOverride: &StaffScheduleOverride{ShiftType: "full", WorkWindows: []WorkPeriod{{Start: "bad", End: "0700"}}},
		}},
	}

	// Act
	_, err := GenerateTimeSlots(input)

	// Assert
	require.Error(t, err)
}
