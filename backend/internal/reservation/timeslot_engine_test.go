package reservation

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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

// TC6: paid_leave も空リスト
func TestGenerateTimeSlots_PaidLeave(t *testing.T) {
	input := defaultInput(15)
	input.Staffs[0].ScheduleOverride = &StaffScheduleOverride{
		ShiftType: "paid_leave",
	}
	slots, err := GenerateTimeSlots(&input)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

// TC7: allow_gaps モード — 指定間隔で生成
func TestGenerateTimeSlots_AllowGapsMode(t *testing.T) {
	input := defaultInput(15)
	input.IntervalMinutes = 30 // 30分間隔で生成
	slots, err := GenerateTimeSlots(&input)
	require.NoError(t, err)
	// 間隔が30分になっていること
	if len(slots) >= 2 {
		s0, _ := MinutesSinceMidnight(slots[0].StartTime)
		s1, _ := MinutesSinceMidnight(slots[1].StartTime)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
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
	slots, err := GenerateTimeSlots(&input)
	require.NoError(t, err)
	// 午後のみ（13:00-19:00）
	assert.NotEmpty(t, slots)
	for _, s := range slots {
		assert.False(t, s.StartTime < "1300", "afternoon shift should not have morning slots")
	}
}

func tsStrPtr(s string) *string { return &s }

// ---- MinutesSinceMidnight 直接テスト ----

func TestMinutesSinceMidnight(t *testing.T) {
	tests := []struct {
		name    string
		hhmm    string
		want    int
		wantErr bool
	}{
		{name: "valid time 09:00", hhmm: "0900", want: 540},
		{name: "valid time 23:59", hhmm: "2359", want: 1439},
		{name: "valid time 00:00", hhmm: "0000", want: 0},
		{name: "invalid length (too short)", hhmm: "900", wantErr: true},
		{name: "invalid length (too long)", hhmm: "09000", wantErr: true},
		{name: "invalid hour", hhmm: "2500", wantErr: true},
		{name: "invalid minute", hhmm: "0060", wantErr: true},
		// R1-3: 非ASCII数字混入は明示的に拒否する。以下は byte 型の unsigned wraparound
		// （'0'=0x30 未満の文字は減算で 200+ に、'9' 超の文字は 10+ にラップする）により
		// 現行実装が「たまたま」レンジ外へ弾いているだけの例。桁が偶然レンジ内に収まる非数字
		// （':' などの 0x3A-0x47 台の文字）は素通りするため、明示的な ASCII 数字チェックが必要。
		{name: "non-digit letter in hour tens (caught by range, kept as regression)", hhmm: "1a00", wantErr: true},
		{name: "non-digit letter in minute tens (caught by range, kept as regression)", hhmm: "12x0", wantErr: true},
		{name: "full-width digits (caught by length, kept as regression)", hhmm: "１２00", wantErr: true},
		{name: "leading space (caught by range, kept as regression)", hhmm: " 900", wantErr: true},
		{name: "colon in hour ones digit slips through range check", hhmm: "1:00", wantErr: true},
		{name: "colon in hour ones digit with zero tens slips through range check", hhmm: "0:00", wantErr: true},
		{name: "colon in minute ones digit slips through range check", hhmm: "090:", wantErr: true},
		{name: "semicolon in hour ones digit slips through range check", hhmm: "1;00", wantErr: true},
		{name: "less-than sign in hour ones digit slips through range check", hhmm: "0<00", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MinutesSinceMidnight(tt.hhmm)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// ---- resolveWorkIntervals 直接テスト ----

func TestResolveWorkIntervals(t *testing.T) {
	baseInput := func() *TimeSlotsInput {
		return &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "1900"},
			DefaultBreaks: []BreakPeriod{{Start: "1200", End: "1300"}},
		}
	}

	t.Run("業務時間開始が不正な場合はエラー", func(t *testing.T) {
		input := baseInput()
		input.BusinessHours.Start = "invalid"
		_, err := resolveWorkIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("業務時間終了が不正な場合はエラー", func(t *testing.T) {
		input := baseInput()
		input.BusinessHours.End = "invalid"
		_, err := resolveWorkIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("デフォルト休憩開始が不正な場合はエラー", func(t *testing.T) {
		input := baseInput()
		input.DefaultBreaks = []BreakPeriod{{Start: "invalid", End: "1300"}}
		_, err := resolveWorkIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("デフォルト休憩終了が不正な場合はエラー", func(t *testing.T) {
		input := baseInput()
		input.DefaultBreaks = []BreakPeriod{{Start: "1200", End: "invalid"}}
		_, err := resolveWorkIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("DefaultBreaksが空の場合はbsEndを既定の休憩境界に使用", func(t *testing.T) {
		input := baseInput()
		input.DefaultBreaks = nil
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{ScheduleOverride: &StaffScheduleOverride{ShiftType: "morning"}})
		assert.NoError(t, err)
		assert.Equal(t, []interval{{540, 1140}}, intervals)
	})

	t.Run("ScheduleOverrideがnilの場合は基本営業時間を使用", func(t *testing.T) {
		input := baseInput()
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{})
		assert.NoError(t, err)
		assert.Equal(t, []interval{{540, 1140}}, intervals)
	})

	t.Run("ShiftType=off は空", func(t *testing.T) {
		input := baseInput()
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{ScheduleOverride: &StaffScheduleOverride{ShiftType: "off"}})
		assert.NoError(t, err)
		assert.Nil(t, intervals)
	})

	t.Run("ShiftType=paid_leave は空", func(t *testing.T) {
		input := baseInput()
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{ScheduleOverride: &StaffScheduleOverride{ShiftType: "paid_leave"}})
		assert.NoError(t, err)
		assert.Nil(t, intervals)
	})

	t.Run("WorkStartが不正な文字列の場合はデフォルトにフォールバック", func(t *testing.T) {
		input := baseInput()
		invalidStart := "invalid"
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{ShiftType: "full", WorkStart: &invalidStart},
		})
		assert.NoError(t, err)
		assert.Equal(t, 540, intervals[0].start) // bsStart にフォールバック
	})

	t.Run("WorkEndが不正な文字列の場合はデフォルトにフォールバック", func(t *testing.T) {
		input := baseInput()
		invalidEnd := "invalid"
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{ShiftType: "full", WorkEnd: &invalidEnd},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1140, intervals[0].end) // bsEnd にフォールバック
	})

	t.Run("afternoonシフトでWorkStart/WorkEndが不正な場合はデフォルトにフォールバック", func(t *testing.T) {
		input := baseInput()
		invalid := "invalid"
		intervals, err := resolveWorkIntervals(input, &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{ShiftType: "afternoon", WorkStart: &invalid, WorkEnd: &invalid},
		})
		assert.NoError(t, err)
		assert.Equal(t, []interval{{780, 1140}}, intervals) // defaultBreakEnd(13:00) 〜 bsEnd(19:00)
	})
}

// ---- resolveBreakIntervals 直接テスト ----

func TestResolveBreakIntervals(t *testing.T) {
	t.Run("ScheduleOverrideがnilの場合はDefaultBreaksを使用", func(t *testing.T) {
		input := &TimeSlotsInput{DefaultBreaks: []BreakPeriod{{Start: "1200", End: "1300"}}}
		got, err := resolveBreakIntervals(input, &StaffSlotInput{})
		assert.NoError(t, err)
		assert.Equal(t, []interval{{720, 780}}, got)
	})

	t.Run("ScheduleOverrideにBreaksがある場合は個人設定を優先", func(t *testing.T) {
		input := &TimeSlotsInput{DefaultBreaks: []BreakPeriod{{Start: "1200", End: "1300"}}}
		got, err := resolveBreakIntervals(input, &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{Breaks: []BreakPeriod{{Start: "1400", End: "1430"}}},
		})
		assert.NoError(t, err)
		assert.Equal(t, []interval{{840, 870}}, got)
	})

	t.Run("休憩開始が不正な場合はエラー", func(t *testing.T) {
		input := &TimeSlotsInput{DefaultBreaks: []BreakPeriod{{Start: "invalid", End: "1300"}}}
		_, err := resolveBreakIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("休憩終了が不正な場合はエラー", func(t *testing.T) {
		input := &TimeSlotsInput{DefaultBreaks: []BreakPeriod{{Start: "1200", End: "invalid"}}}
		_, err := resolveBreakIntervals(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("休憩が空の場合は空リスト", func(t *testing.T) {
		input := &TimeSlotsInput{}
		got, err := resolveBreakIntervals(input, &StaffSlotInput{})
		assert.NoError(t, err)
		assert.Empty(t, got)
	})
}

// ---- generateAllowGaps 直接テスト ----

func TestGenerateAllowGaps(t *testing.T) {
	t.Run("intervalMinutesが0以下の場合はdurにフォールバック", func(t *testing.T) {
		available := []interval{{540, 600}} // 09:00-10:00
		slots, err := generateAllowGaps(available, 15, 0)
		assert.NoError(t, err)
		assert.Len(t, slots, 4) // 60分/15分=4枠
	})

	t.Run("区間に枠が収まらない場合は空", func(t *testing.T) {
		available := []interval{{540, 550}} // 10分しかない
		slots, err := generateAllowGaps(available, 15, 15)
		assert.NoError(t, err)
		assert.Empty(t, slots)
	})

	t.Run("複数区間から生成", func(t *testing.T) {
		available := []interval{{540, 570}, {600, 630}} // 09:00-09:30, 10:00-10:30
		slots, err := generateAllowGaps(available, 15, 15)
		assert.NoError(t, err)
		assert.Len(t, slots, 4)
	})
}

// ---- generateForStaff 直接テスト ----

func TestGenerateForStaff(t *testing.T) {
	t.Run("resolveWorkIntervalsのエラーを伝播", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "invalid", End: "1900"},
			Mode:          "allow_gaps",
		}
		_, err := generateForStaff(input, &StaffSlotInput{})
		assert.Error(t, err)
	})

	t.Run("resolveBreakIntervalsのエラーを伝播", func(t *testing.T) {
		// DefaultBreaks 自体は valid にして resolveWorkIntervals を通過させ、
		// staffInput.ScheduleOverride.Breaks（優先される）を invalid にして resolveBreakIntervals 側のみを失敗させる。
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "1900"},
			DefaultBreaks: []BreakPeriod{{Start: "1200", End: "1300"}},
			Mode:          "allow_gaps",
		}
		staffInput := &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{Breaks: []BreakPeriod{{Start: "invalid", End: "1300"}}},
		}
		_, err := generateForStaff(input, staffInput)
		assert.Error(t, err)
	})

	t.Run("既存予約のStartTimeが不正な場合はエラー", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "1900"},
			Mode:          "allow_gaps",
		}
		staffInput := &StaffSlotInput{
			StaffID: 1,
			ExistingResvs: []ExistingReservation{
				{StaffID: 1, StartTime: "invalid", EndTime: "1000"},
			},
		}
		_, err := generateForStaff(input, staffInput)
		assert.Error(t, err)
	})

	t.Run("既存予約のEndTimeが不正な場合はエラー", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "1900"},
			Mode:          "allow_gaps",
		}
		staffInput := &StaffSlotInput{
			StaffID: 1,
			ExistingResvs: []ExistingReservation{
				{StaffID: 1, StartTime: "0900", EndTime: "invalid"},
			},
		}
		_, err := generateForStaff(input, staffInput)
		assert.Error(t, err)
	})

	t.Run("StaffID=0の場合は全既存予約を除外対象にする", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours:   BusinessHours{Start: "0900", End: "1000"},
			CourseDuration:  15,
			IntervalMinutes: 15,
			Mode:            "allow_gaps",
		}
		staffInput := &StaffSlotInput{
			StaffID: 0,
			ExistingResvs: []ExistingReservation{
				{StaffID: 99, StartTime: "0900", EndTime: "0930"},
			},
		}
		slots, err := generateForStaff(input, staffInput)
		assert.NoError(t, err)
		for _, s := range slots {
			assert.False(t, s.StartTime >= "0900" && s.StartTime < "0930")
		}
	})

	t.Run("休日の場合はnilを返す", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "1900"},
			Mode:          "allow_gaps",
		}
		staffInput := &StaffSlotInput{
			ScheduleOverride: &StaffScheduleOverride{ShiftType: "off"},
		}
		slots, err := generateForStaff(input, staffInput)
		assert.NoError(t, err)
		assert.Nil(t, slots)
	})

	t.Run("CourseDuration未指定時は15分デフォルト", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours: BusinessHours{Start: "0900", End: "0930"},
			Mode:          "allow_gaps",
		}
		slots, err := generateForStaff(input, &StaffSlotInput{})
		assert.NoError(t, err)
		assert.Len(t, slots, 2) // 30分/15分=2枠
	})

	t.Run("minimize_gapsモードでMinCourseDuration未指定時はdurを使用", func(t *testing.T) {
		input := &TimeSlotsInput{
			BusinessHours:  BusinessHours{Start: "0900", End: "1000"},
			CourseDuration: 15,
			Mode:           "minimize_gaps",
		}
		slots, err := generateForStaff(input, &StaffSlotInput{})
		assert.NoError(t, err)
		assert.NotEmpty(t, slots)
	})
}
