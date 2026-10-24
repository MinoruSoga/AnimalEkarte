package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-549 F2: 勤務の starts_at / ends_at が NULL（interval_state=unknown）のスタッフを、
// 営業時間や shift_type の既定時間帯へ戻して勤務可能とみなさない（api.yaml ShiftEntry）。
func TestBuildStaffSlotInputsFromWindow_ExcludesUnknownShift(t *testing.T) {
	staffs := []model.Staff{{ID: 1}, {ID: 2}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		entry  *model.ShiftEntry
		breaks []model.ShiftEntryBreak
	}{
		{name: "full shift with both timestamps NULL", entry: &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeFull}},
		{name: "morning shift with both timestamps NULL", entry: &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeMorning}},
		{name: "afternoon shift with only starts_at", entry: &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeAfternoon, StartsAt: rsvTS(t, date, "13:00")}},
		{name: "known shift with a break whose times are NULL", entry: &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "09:00"), EndsAt: rsvTS(t, date, "18:00")}, breaks: []model.ShiftEntryBreak{{}}},
		{name: "off shift with leftover timestamps", entry: &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeOff, StartsAt: rsvTS(t, date, "09:00"), EndsAt: rsvTS(t, date, "18:00")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := map[string]map[uint64]*model.ShiftEntry{"2026-08-01": {1: tt.entry}}
			breaksByEntry := map[uint64][]model.ShiftEntryBreak{10: tt.breaks}

			inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, breaksByEntry, nil)

			require.Len(t, inputs, 1, "unknown の staff 1 は候補外、勤務未登録の staff 2 は従来どおり残る")
			assert.Equal(t, uint64(2), inputs[0].StaffID)
		})
	}
}

func TestBuildStaffSlotInputsFromWindow_UnknownShiftDoesNotFallBackToBusinessHours(t *testing.T) {
	// Arrange: 営業時間 09:00-18:00 の医院で、勤務の日時が NULL のスタッフ 1 人だけ。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	entries := map[string]map[uint64]*model.ShiftEntry{
		"2026-08-01": {1: {ID: 10, ShiftType: model.ShiftTypeFull}},
	}
	inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, nil, nil)

	// Act
	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  60,
		IntervalMinutes: 60,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})

	// Assert
	require.NoError(t, err)
	assert.Empty(t, slots, "unknown を営業時間の既定で勤務可能と扱わない")
}

func TestBuildStaffSlotInputsFromWindow_ClipsOvernightShiftToTheEvaluatedDay(t *testing.T) {
	// Arrange: 8/1 22:00 から 8/2 07:00 の勤務。休憩は 8/1 23:00-23:30 と 8/2 01:00-02:00。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	entries := map[string]map[uint64]*model.ShiftEntry{
		"2026-08-01": {1: {ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00"), EndsAt: rsvTS(t, nextDay, "07:00")}},
	}
	breaksByEntry := map[uint64][]model.ShiftEntryBreak{
		10: {rsvBreak(t, date, "23:00", "23:30"), rsvBreak(t, nextDay, "01:00", "02:00")},
	}

	// Act
	inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, breaksByEntry, nil)

	// Assert
	require.Len(t, inputs, 1)
	override := inputs[0].ScheduleOverride
	require.NotNil(t, override)
	require.NotNil(t, override.WorkStart)
	require.NotNil(t, override.WorkEnd)
	assert.Equal(t, "2200", *override.WorkStart)
	assert.Equal(t, "2359", *override.WorkEnd, "評価日の末尾は HHMM で表せる 23:59 に丸める")
	assert.Equal(t, []BreakPeriod{{Start: "2300", End: "2330"}}, override.Breaks, "翌日側の休憩は評価日に掛からない")

	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  30,
		IntervalMinutes: 30,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})
	require.NoError(t, err)
	assert.Equal(t, []TimeSlot{{StartTime: "2200", EndTime: "2230"}, {StartTime: "2230", EndTime: "2300"}}, slots)
}

func TestBuildStaffSlotInputsFromWindow_ShiftOnAnotherDayHasNoWorkWindow(t *testing.T) {
	// Arrange: 勤務行の日付キーと starts_at の暦日が食い違う旧データは、営業時間へ戻さず空の勤務窓にする。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	other := date.AddDate(0, 0, 3)
	entries := map[string]map[uint64]*model.ShiftEntry{
		"2026-08-01": {1: {ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, other, "09:00"), EndsAt: rsvTS(t, other, "18:00")}},
	}
	inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, nil, nil)

	// Act
	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  60,
		IntervalMinutes: 60,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Empty(t, slots)
}

// rsvAt は rsvTS の時刻に秒未満の端数を足した時刻を返す（秒未満を含む境界のテスト用）。
func rsvAt(t testing.TB, date time.Time, clock string, extra time.Duration) time.Time {
	t.Helper()
	return rsvTS(t, date, clock).Add(extra)
}

// EMR-549 A5-1: 秒・秒未満を含む境界は、使える時間を広げない向きに分単位へ丸める。
// 勤務は開始を切り上げ・終了を切り捨て、休憩は開始を切り捨て・終了を切り上げる。ちょうどの分は変えない。
func TestClipToDayHHMM_RoundsBoundariesWithoutWideningAvailability(t *testing.T) {
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	prevDay := date.AddDate(0, 0, -1)
	dayStart := time.Date(2026, 8, 1, 0, 0, 0, 0, config.JST)
	dayEnd := dayStart.AddDate(0, 0, 1)
	const half = 500 * time.Millisecond
	at := func(d time.Time, clock string, extra time.Duration) time.Time { return rsvAt(t, d, clock, extra) }

	tests := []struct {
		name       string
		startsAt   time.Time
		endsAt     time.Time
		startRound minuteRounding
		endRound   minuteRounding
		wantStart  string
		wantEnd    string
		wantOK     bool
	}{
		// 勤務（開始は切り上げ・終了は切り捨て）
		{"work exact minutes are unchanged", at(date, "09:00", 0), at(date, "18:00", 0), roundMinuteUp, roundMinuteDown, "0900", "1800", true},
		{"work start with seconds rounds up", at(date, "09:00:30", 0), at(date, "18:00", 0), roundMinuteUp, roundMinuteDown, "0901", "1800", true},
		{"work start with sub-second rounds up", at(date, "09:00", half), at(date, "18:00", 0), roundMinuteUp, roundMinuteDown, "0901", "1800", true},
		{"work end with seconds rounds down", at(date, "09:00", 0), at(date, "18:00:30", 0), roundMinuteUp, roundMinuteDown, "0900", "1800", true},
		{"work end with sub-second rounds down", at(date, "09:00", 0), at(date, "18:00", half), roundMinuteUp, roundMinuteDown, "0900", "1800", true},
		{"work end just before a minute rounds down", at(date, "09:00", 0), at(date, "17:59:59", half), roundMinuteUp, roundMinuteDown, "0900", "1759", true},
		// 休憩（開始は切り捨て・終了は切り上げ）
		{"break exact minutes are unchanged", at(date, "12:00", 0), at(date, "12:30", 0), roundMinuteDown, roundMinuteUp, "1200", "1230", true},
		{"break start with seconds rounds down", at(date, "12:00:30", 0), at(date, "12:30", 0), roundMinuteDown, roundMinuteUp, "1200", "1230", true},
		{"break start with sub-second rounds down", at(date, "12:00", half), at(date, "12:30", 0), roundMinuteDown, roundMinuteUp, "1200", "1230", true},
		{"break end with seconds rounds up", at(date, "12:00", 0), at(date, "12:30:30", 0), roundMinuteDown, roundMinuteUp, "1200", "1231", true},
		{"break end with sub-second rounds up", at(date, "12:00", 0), at(date, "12:30", half), roundMinuteDown, roundMinuteUp, "1200", "1231", true},
		{"zero-length exact break is unchanged", at(date, "12:00", 0), at(date, "12:00", 0), roundMinuteDown, roundMinuteUp, "1200", "1200", true},
		{"break shorter than a minute widens instead of vanishing", at(date, "12:00:10", 0), at(date, "12:00:50", 0), roundMinuteDown, roundMinuteUp, "1200", "1201", true},
		// 評価日の境界（翌 0:00 は "2359"、日跨ぎは当日分だけを切り出す）
		{"work ending exactly at the end of the day keeps 2359", at(date, "22:00", 0), at(nextDay, "00:00", 0), roundMinuteUp, roundMinuteDown, "2200", "2359", true},
		{"break end rounded up to the next midnight keeps 2359", at(date, "23:00", 0), at(date, "23:59:30", 0), roundMinuteDown, roundMinuteUp, "2300", "2359", true},
		{"work start rounded up to the next midnight leaves nothing", at(date, "23:59:30", 0), at(nextDay, "07:00", 0), roundMinuteUp, roundMinuteDown, "", "", false},
		{"work start at exactly 23:59 is unchanged", at(date, "23:59", 0), at(nextDay, "07:00", 0), roundMinuteUp, roundMinuteDown, "2359", "2359", true},
		{"overnight work is clipped to the evaluated day", at(date, "22:00:30", 0), at(nextDay, "07:00:30", 0), roundMinuteUp, roundMinuteDown, "2201", "2359", true},
		{"overnight work from the previous day starts at midnight", at(prevDay, "22:00:30", 0), at(date, "07:00:30", 0), roundMinuteUp, roundMinuteDown, "0000", "0700", true},
		{"overnight break from the previous day starts at midnight", at(prevDay, "23:30:30", 0), at(date, "00:30:30", 0), roundMinuteDown, roundMinuteUp, "0000", "0031", true},
		{"work that does not touch the day is ignored", at(nextDay, "09:00", 0), at(nextDay, "18:00", 0), roundMinuteUp, roundMinuteDown, "", "", false},
		// 丸めで空になる勤務は可用時間を足さない
		{"work inside a single minute rounds to empty", at(date, "09:00:10", 0), at(date, "09:00:50", 0), roundMinuteUp, roundMinuteDown, "", "", false},
		{"work whose rounded start meets the rounded end is empty", at(date, "09:00:30", 0), at(date, "09:01", 0), roundMinuteUp, roundMinuteDown, "", "", false},
		{"work spanning no whole minute is empty", at(date, "09:00:10", 0), at(date, "09:01:50", 0), roundMinuteUp, roundMinuteDown, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, ok := clipToDayHHMM(tt.startsAt, tt.endsAt, dayStart, dayEnd, tt.startRound, tt.endRound)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}

// EMR-549 A5-1: 秒を含む勤務・休憩は、勤務窓を狭め休憩を広げる向きに丸め、枠が休憩へ食い込まない。
func TestBuildStaffSlotInputsFromWindow_RoundsShiftSecondsWithoutWideningAvailability(t *testing.T) {
	// Arrange: 勤務 09:00:30〜18:00:30、休憩 12:00:30〜12:30:30。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	entries := map[string]map[uint64]*model.ShiftEntry{
		"2026-08-01": {1: {ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "09:00:30"), EndsAt: rsvTS(t, date, "18:00:30")}},
	}
	breaksByEntry := map[uint64][]model.ShiftEntryBreak{
		10: {rsvBreak(t, date, "12:00:30", "12:30:30")},
	}

	// Act
	inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, breaksByEntry, nil)

	// Assert
	require.Len(t, inputs, 1)
	override := inputs[0].ScheduleOverride
	require.NotNil(t, override)
	require.NotNil(t, override.WorkStart)
	require.NotNil(t, override.WorkEnd)
	assert.Equal(t, "0901", *override.WorkStart, "勤務開始 09:00:30 は切り上げて勤務窓を狭める")
	assert.Equal(t, "1800", *override.WorkEnd, "勤務終了 18:00:30 は切り捨てる")
	assert.Equal(t, []BreakPeriod{{Start: "1200", End: "1231"}}, override.Breaks, "休憩は開始を切り捨て・終了を切り上げて広げる")

	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  30,
		IntervalMinutes: 30,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})
	require.NoError(t, err)
	require.NotEmpty(t, slots)
	assert.Equal(t, "0901", slots[0].StartTime, "勤務開始 09:00:30 より前の枠は出ない")
	assert.LessOrEqual(t, slots[len(slots)-1].EndTime, "1800")
	for _, slot := range slots {
		assert.False(t, slot.StartTime < "1231" && slot.EndTime > "1200", "休憩 12:00:30〜12:30:30 に食い込む枠は出ない: %v", slot)
	}
}

// EMR-549 A5-1: 丸めで 1 分も残らない勤務は空の勤務窓にし、営業時間へ戻さない。
func TestBuildStaffSlotInputsFromWindow_ShiftRoundedToEmptyHasNoWorkWindow(t *testing.T) {
	// Arrange: 勤務 09:00:10〜09:00:50（丸めると開始 09:01 が終了 09:00 を超える）。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	entries := map[string]map[uint64]*model.ShiftEntry{
		"2026-08-01": {1: {ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "09:00:10"), EndsAt: rsvTS(t, date, "09:00:50")}},
	}
	inputs := buildStaffSlotInputsFromWindow(staffs, date, entries, nil, nil)

	// Act
	slots, err := GenerateTimeSlots(&TimeSlotsInput{
		BusinessHours:   BusinessHours{Start: "0900", End: "1800"},
		CourseDuration:  15,
		IntervalMinutes: 15,
		Mode:            "allow_gaps",
		Staffs:          inputs,
	})

	// Assert
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	require.NotNil(t, inputs[0].ScheduleOverride)
	require.NotNil(t, inputs[0].ScheduleOverride.WorkStart)
	require.NotNil(t, inputs[0].ScheduleOverride.WorkEnd)
	assert.Equal(t, "0000", *inputs[0].ScheduleOverride.WorkStart)
	assert.Equal(t, "0000", *inputs[0].ScheduleOverride.WorkEnd)
	assert.Empty(t, slots)
}

// EMR-549 A5-1: 日跨ぎ勤務の秒付き境界も、評価日ごとに当日分だけを切り出して丸める。
func TestBuildStaffSlotInputsFromWindow_RoundsOvernightShiftSecondsPerEvaluatedDay(t *testing.T) {
	// Arrange: 8/1 22:00:30 から 8/2 07:00:30 の勤務。休憩は 8/1 23:00:30-23:30:30 と 8/2 01:00:30-02:00:30。
	staffs := []model.Staff{{ID: 1}}
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	nextDay := date.AddDate(0, 0, 1)
	entry := &model.ShiftEntry{ID: 10, ShiftType: model.ShiftTypeFull, StartsAt: rsvTS(t, date, "22:00:30"), EndsAt: rsvTS(t, nextDay, "07:00:30")}
	breaksByEntry := map[uint64][]model.ShiftEntryBreak{
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
			entries := map[string]map[uint64]*model.ShiftEntry{tt.evaluated.Format(time.DateOnly): {1: entry}}

			inputs := buildStaffSlotInputsFromWindow(staffs, tt.evaluated, entries, breaksByEntry, nil)

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
