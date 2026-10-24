package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
