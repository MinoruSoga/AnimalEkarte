package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// defaultInput は最小限の共通設定（mode: "allow_gaps"）を返すヘルパー。
func defaultInput(dur int) TimeSlotsInput {
	return TimeSlotsInput{
		BusinessHours:     BusinessHours{Start: "0900", End: "1900"},
		DefaultBreaks:     []BreakPeriod{{Start: "1200", End: "1300"}},
		CourseDuration:    dur,
		IntervalMinutes:   dur,
		Mode:              "allow_gaps",
		MinCourseDuration: dur,
		Staffs: []StaffSlotInput{
			{StaffID: 1},
		},
	}
}

// ---- テストケース ----

// TC1: 基本: 営業時間内（09:00-19:00）で15分枠を生成
func TestGenerateTimeSlots_Basic(t *testing.T) {
	input := defaultInput(15)
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	assert.NotEmpty(t, slots)
	assert.Equal(t, "0900", slots[0].StartTime)
	assert.Equal(t, "0915", slots[0].EndTime)
	// 昼休憩（12:00-13:00）をまたいで最後のスロットは 18:45-19:00 のはず
	last := slots[len(slots)-1]
	assert.Equal(t, "1845", last.StartTime)
	assert.Equal(t, "1900", last.EndTime)
}

// TC2: 休憩時間をまたぐ枠が除外される（12:00-13:00 の休憩）
func TestGenerateTimeSlots_BreakExclusion(t *testing.T) {
	input := defaultInput(15)
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	for _, s := range slots {
		// 休憩時間内（1200-1300）に開始する枠がないこと
		assert.False(t, s.StartTime >= "1200" && s.StartTime < "1300",
			"slot %s-%s starts during break", s.StartTime, s.EndTime)
		// 休憩時間にまたがる枠がないこと（例: 1155-1210 は不可）
		assert.False(t, s.StartTime < "1200" && s.EndTime > "1200",
			"slot %s-%s overlaps break start", s.StartTime, s.EndTime)
	}
}

// TC3: 既存予約と重複する枠が除外される
func TestGenerateTimeSlots_ExistingReservationExclusion(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ExistingResvs = []ExistingReservation{
		{StaffID: 1, StartTime: "0900", EndTime: "0930"},
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	for _, s := range slots {
		assert.False(t, s.StartTime >= "0900" && s.StartTime < "0930",
			"slot %s should be excluded by existing reservation", s.StartTime)
	}
	// 09:30 以降は存在すること
	found0930 := false
	for _, s := range slots {
		if s.StartTime == "0930" {
			found0930 = true
		}
	}
	assert.True(t, found0930, "slot 0930 should be available")
}

// TC4: 個人設定で営業時間が変更された場合
func TestGenerateTimeSlots_StaffScheduleOverride(t *testing.T) {
	input := defaultInput(15)
	workStart := "1000"
	workEnd := "1600"
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "full",
		WorkStart: &workStart,
		WorkEnd:   &workEnd,
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	assert.NotEmpty(t, slots)
	assert.Equal(t, "1000", slots[0].StartTime)
	last := slots[len(slots)-1]
	assert.Equal(t, "1600", last.EndTime)
	// 09:00-10:00 の枠がないこと
	for _, s := range slots {
		assert.False(t, s.StartTime < "1000", "slot %s before work start", s.StartTime)
		assert.False(t, s.EndTime > "1600", "slot %s after work end", s.EndTime)
	}
}

// TC5: 個人設定で休日の場合は空リスト
func TestGenerateTimeSlots_StaffDayOff(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "off",
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

// TC6: paid_leave も空リスト
func TestGenerateTimeSlots_PaidLeave(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "paid_leave",
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

// TC7: allow_gaps モード — 指定間隔で生成
func TestGenerateTimeSlots_AllowGapsMode(t *testing.T) {
	input := defaultInput(15)
	input.IntervalMinutes = 30 // 30分間隔で生成
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// 間隔が30分になっていること
	if len(slots) >= 2 {
		s0, _ := minutesSinceMidnight(slots[0].StartTime)
		s1, _ := minutesSinceMidnight(slots[1].StartTime)
		assert.Equal(t, 30, s1-s0, "slots should be 30 minutes apart")
	}
}

// TC8: minimize_gaps モード — 最短コース時間を考慮して空きを最小化
func TestGenerateTimeSlots_MinimizeGapsMode(t *testing.T) {
	input := TimeSlotsInput{
		BusinessHours:     BusinessHours{Start: "0900", End: "1200"},
		DefaultBreaks:     nil,
		CourseDuration:    60, // 60分コース
		IntervalMinutes:   60,
		Mode:              "minimize_gaps",
		MinCourseDuration: 15, // 最短コース15分 → 余り15分未満なら無視
		Staffs:            []StaffSlotInput{{StaffID: 1}},
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// 09:00-12:00 = 180分。60分コースなら最大3枠だが…
	// minimize_gaps: 60分詰めて 09:00, 10:00, 11:00 → 3枠
	assert.Len(t, slots, 3)
	assert.Equal(t, "0900", slots[0].StartTime)
	assert.Equal(t, "1000", slots[1].StartTime)
	assert.Equal(t, "1100", slots[2].StartTime)
}

// TC9: 複数スタッフ（指名なし）— 全スタッフの空き時間を統合（UNION）
func TestGenerateTimeSlots_MultipleStaffs(t *testing.T) {
	// Staff1: 09:00-12:00 のみ勤務
	// Staff2: 13:00-17:00 のみ勤務
	end1 := "1200"
	start2 := "1300"
	end2 := "1700"
	input := TimeSlotsInput{
		BusinessHours:     BusinessHours{Start: "0900", End: "1900"},
		DefaultBreaks:     nil,
		CourseDuration:    15,
		IntervalMinutes:   15,
		Mode:              "allow_gaps",
		MinCourseDuration: 15,
		Staffs: []StaffSlotInput{
			{
				StaffID: 1,
				ScheduleOverride: &StaffScheduleOverride{
					ShiftType: "full",
					WorkStart: tsStrPtr("0900"),
					WorkEnd:   &end1,
				},
			},
			{
				StaffID: 2,
				ScheduleOverride: &StaffScheduleOverride{
					ShiftType: "full",
					WorkStart: &start2,
					WorkEnd:   &end2,
				},
			},
		},
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// Staff1の枠と Staff2 の枠が統合されること
	has0900 := false
	has1300 := false
	for _, s := range slots {
		if s.StartTime == "0900" {
			has0900 = true
		}
		if s.StartTime == "1300" {
			has1300 = true
		}
	}
	assert.True(t, has0900, "staff1's 09:00 slot should be present")
	assert.True(t, has1300, "staff2's 13:00 slot should be present")
}

// TC10: 60分コース（手術）の枠生成
func TestGenerateTimeSlots_60MinCourse(t *testing.T) {
	input := defaultInput(60)
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	assert.NotEmpty(t, slots)
	// 最初の枠は09:00-10:00
	assert.Equal(t, "0900", slots[0].StartTime)
	assert.Equal(t, "1000", slots[0].EndTime)
	// 11:00-12:00 は存在するが 11:30-12:30 は休憩をまたぐので存在しないこと
	for _, s := range slots {
		assert.False(t, s.StartTime < "1200" && s.EndTime > "1200",
			"slot %s-%s should not overlap break", s.StartTime, s.EndTime)
	}
}

// TC11: 15分コース（一般診察）の枠生成 — 計算件数確認
func TestGenerateTimeSlots_15MinCourse_Count(t *testing.T) {
	input := defaultInput(15)
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// 09:00-12:00 = 180分 → 12枠
	// 13:00-19:00 = 360分 → 24枠
	// 合計 36枠
	assert.Equal(t, 36, len(slots))
}

// TC12: 個人設定で morning シフト（午前のみ）
func TestGenerateTimeSlots_MorningShift(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "morning",
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// 午前のみ（09:00-12:00）
	assert.NotEmpty(t, slots)
	for _, s := range slots {
		assert.False(t, s.StartTime >= "1200", "morning shift should not have afternoon slots")
	}
}

// TC13: 個人設定で afternoon シフト（午後のみ）
func TestGenerateTimeSlots_AfternoonShift(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "afternoon",
	}
	slots, err := GenerateTimeSlots(input)
	require.NoError(t, err)
	// 午後のみ（13:00-19:00）
	assert.NotEmpty(t, slots)
	for _, s := range slots {
		assert.False(t, s.StartTime < "1300", "afternoon shift should not have morning slots")
	}
}

func tsStrPtr(s string) *string { return &s }
