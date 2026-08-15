package reservation

import (
	"fmt"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ---- 時間表現ヘルパー ----

// MinutesSinceMidnight は "HHMM" 形式の文字列を午前0時からの分数に変換する。
func MinutesSinceMidnight(hhmm string) (int, error) {
	if len(hhmm) != 4 {
		return 0, apperrors.WrapInvalidInput(fmt.Sprintf("invalid time format %q: must be HHMM", hhmm))
	}
	// byte 型の unsigned wraparound（'0' 未満の文字は減算で200+に、'9' 超の文字は10+に
	// ラップする）に依存せず、ASCII 数字4桁であることを明示的に検証する。桁が偶然レンジ内に
	// 収まる非数字（例: ':' 0x3A）は下の h/m レンジチェックだけではすり抜けるため必須。
	for i := 0; i < len(hhmm); i++ {
		if hhmm[i] < '0' || hhmm[i] > '9' {
			return 0, apperrors.WrapInvalidInput(fmt.Sprintf("invalid time format %q: must be HHMM", hhmm))
		}
	}
	h := int(hhmm[0]-'0')*10 + int(hhmm[1]-'0')
	m := int(hhmm[2]-'0')*10 + int(hhmm[3]-'0')
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, apperrors.WrapInvalidInput(fmt.Sprintf("invalid time value %q", hhmm))
	}
	return h*60 + m, nil
}

// minutesToHHMM は午前0時からの分数を "HHMM" 形式に変換する。
func minutesToHHMM(total int) string {
	h := total / 60
	m := total % 60
	return fmt.Sprintf("%02d%02d", h, m)
}

// ---- 時間区間 ----

type interval struct {
	start int // minutes since midnight
	end   int // minutes since midnight (exclusive)
}

// subtract は base から excl の時間区間を除外した区間リストを返す。
func subtract(base, excl []interval) []interval {
	result := make([]interval, len(base))
	copy(result, base)
	for _, ex := range excl {
		next := make([]interval, 0, len(result))
		for _, iv := range result {
			if ex.end <= iv.start || ex.start >= iv.end {
				// 重複なし
				next = append(next, iv)
			} else {
				// 左側が残る場合
				if iv.start < ex.start {
					next = append(next, interval{iv.start, ex.start})
				}
				// 右側が残る場合
				if ex.end < iv.end {
					next = append(next, interval{ex.end, iv.end})
				}
			}
		}
		result = next
	}
	return result
}

// ---- 入力型 ----

// BusinessHours は営業時間を表す。
type BusinessHours struct {
	Start string // "HHMM"
	End   string // "HHMM"
}

// BreakPeriod は休憩時間帯を表す。
type BreakPeriod struct {
	Start string // "HHMM"
	End   string // "HHMM"
}

// ExistingReservation は当日の既存予約を表す（重複除外用）。
type ExistingReservation struct {
	StaffID   uint64
	StartTime string // "HHMM"
	EndTime   string // "HHMM"
}

// StaffScheduleOverride はスタッフ個人の勤務設定を表す。
// nil の場合は基本設定（BusinessHours, Breaks）を使用する。
type StaffScheduleOverride struct {
	ShiftType string        // "full" | "morning" | "afternoon" | "off" | "paid_leave"
	WorkStart *string       // nil なら shift_type のデフォルトを使用
	WorkEnd   *string       // nil なら shift_type のデフォルトを使用
	Breaks    []BreakPeriod // 個人設定の休憩時間
}

// StaffSlotInput は1スタッフ分の時間枠計算入力。
type StaffSlotInput struct {
	StaffID          uint64
	ScheduleOverride *StaffScheduleOverride // nil = 基本設定を使用
	ExistingResvs    []ExistingReservation
}

// TimeSlotsInput は時間枠計算全体の入力。
type TimeSlotsInput struct {
	BusinessHours     BusinessHours
	DefaultBreaks     []BreakPeriod
	CourseDuration    int    // 分
	IntervalMinutes   int    // 枠の生成間隔（allow_gaps用）
	Mode              string // "allow_gaps" | "minimize_gaps"
	MinCourseDuration int    // 最短コース時間（minimize_gaps用）
	Staffs            []StaffSlotInput
}

// TimeSlot は空き時間枠を表す。
type TimeSlot struct {
	StartTime string // "HHMM"
	EndTime   string // "HHMM"
}

// ---- エンジン本体 ----

// GenerateTimeSlots は入力に基づいて空き時間枠を計算して返す。
// 複数スタッフが指定された場合は全スタッフの空き枠を統合して返す（UNION）。
func GenerateTimeSlots(input *TimeSlotsInput) ([]TimeSlot, error) {
	slotSet := make(map[string]struct{})

	for i := range input.Staffs {
		slots, err := generateForStaff(input, &input.Staffs[i])
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to generate time slots")
		}
		for _, s := range slots {
			key := s.StartTime + "-" + s.EndTime
			slotSet[key] = struct{}{}
		}
	}

	// ソートして返す
	result := make([]TimeSlot, 0, len(slotSet))
	for key := range slotSet {
		start := key[:4]
		end := key[5:]
		result = append(result, TimeSlot{StartTime: start, EndTime: end})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime < result[j].StartTime
	})
	return result, nil
}

// generateForStaff は1スタッフ分の空き時間枠を計算する。
func generateForStaff(input *TimeSlotsInput, staffInput *StaffSlotInput) ([]TimeSlot, error) {
	// 1. 勤務時間を決定
	workIntervals, err := resolveWorkIntervals(input, staffInput)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve work intervals")
	}
	if len(workIntervals) == 0 {
		return nil, nil // 休日
	}

	// 2. 休憩時間を除外
	breaks, err := resolveBreakIntervals(input, staffInput)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve break intervals")
	}
	available := subtract(workIntervals, breaks)

	// 3. 既存予約を除外
	resvIntervals := make([]interval, 0, len(staffInput.ExistingResvs))
	for _, r := range staffInput.ExistingResvs {
		if r.StaffID != staffInput.StaffID && staffInput.StaffID != 0 {
			continue
		}
		s, err := MinutesSinceMidnight(r.StartTime)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to parse reservation start time")
		}
		e, err := MinutesSinceMidnight(r.EndTime)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to parse reservation end time")
		}
		resvIntervals = append(resvIntervals, interval{s, e})
	}
	available = subtract(available, resvIntervals)

	// 4. 空き区間からコース所要時間が収まる枠を列挙
	dur := input.CourseDuration
	if dur <= 0 {
		dur = 15
	}

	switch input.Mode {
	case "allow_gaps":
		return generateAllowGaps(available, dur, input.IntervalMinutes)
	default: // minimize_gaps
		minDur := input.MinCourseDuration
		if minDur <= 0 {
			minDur = dur
		}
		return generateMinimizeGaps(available, dur, minDur)
	}
}

// resolveWorkIntervals はスタッフの勤務時間区間を解決する。
func resolveWorkIntervals(input *TimeSlotsInput, staffInput *StaffSlotInput) ([]interval, error) {
	bsStart, err := MinutesSinceMidnight(input.BusinessHours.Start)
	if err != nil {
		return nil, fmt.Errorf("business_hours.start: %w", err)
	}
	bsEnd, err := MinutesSinceMidnight(input.BusinessHours.End)
	if err != nil {
		return nil, fmt.Errorf("business_hours.end: %w", err)
	}

	// 基本の休憩時間（最初のブレーク）
	var defaultBreakStart, defaultBreakEnd int
	if len(input.DefaultBreaks) > 0 {
		var errBreak error
		defaultBreakStart, errBreak = MinutesSinceMidnight(input.DefaultBreaks[0].Start)
		if errBreak != nil {
			return nil, apperrors.Wrap(errBreak, "default_breaks.start")
		}
		defaultBreakEnd, errBreak = MinutesSinceMidnight(input.DefaultBreaks[0].End)
		if errBreak != nil {
			return nil, apperrors.Wrap(errBreak, "default_breaks.end")
		}
	} else {
		defaultBreakStart = bsEnd
		defaultBreakEnd = bsEnd
	}

	if staffInput.ScheduleOverride == nil {
		return []interval{{bsStart, bsEnd}}, nil
	}

	ov := staffInput.ScheduleOverride
	switch ov.ShiftType {
	case "off", "paid_leave":
		return nil, nil
	case "morning":
		s := bsStart
		e := defaultBreakStart
		if ov.WorkStart != nil {
			if s2, err2 := MinutesSinceMidnight(*ov.WorkStart); err2 == nil {
				s = s2
			}
		}
		if ov.WorkEnd != nil {
			if e2, err2 := MinutesSinceMidnight(*ov.WorkEnd); err2 == nil {
				e = e2
			}
		}
		return []interval{{s, e}}, nil
	case "afternoon":
		s := defaultBreakEnd
		e := bsEnd
		if ov.WorkStart != nil {
			if s2, err2 := MinutesSinceMidnight(*ov.WorkStart); err2 == nil {
				s = s2
			}
		}
		if ov.WorkEnd != nil {
			if e2, err2 := MinutesSinceMidnight(*ov.WorkEnd); err2 == nil {
				e = e2
			}
		}
		return []interval{{s, e}}, nil
	default: // "full"
		s := bsStart
		e := bsEnd
		if ov.WorkStart != nil {
			if s2, err2 := MinutesSinceMidnight(*ov.WorkStart); err2 == nil {
				s = s2
			}
		}
		if ov.WorkEnd != nil {
			if e2, err2 := MinutesSinceMidnight(*ov.WorkEnd); err2 == nil {
				e = e2
			}
		}
		return []interval{{s, e}}, nil
	}
}

// resolveBreakIntervals はスタッフの休憩時間区間を解決する。
func resolveBreakIntervals(input *TimeSlotsInput, staffInput *StaffSlotInput) ([]interval, error) {
	var breaks []BreakPeriod
	if staffInput.ScheduleOverride != nil && len(staffInput.ScheduleOverride.Breaks) > 0 {
		breaks = staffInput.ScheduleOverride.Breaks
	} else {
		breaks = input.DefaultBreaks
	}

	result := make([]interval, 0, len(breaks))
	for _, b := range breaks {
		s, err := MinutesSinceMidnight(b.Start)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to parse break start time")
		}
		e, err := MinutesSinceMidnight(b.End)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to parse break end time")
		}
		result = append(result, interval{s, e})
	}
	return result, nil
}

// generateAllowGaps は指定間隔（interval_minutes）で枠を生成する。
func generateAllowGaps(available []interval, dur, intervalMin int) ([]TimeSlot, error) {
	if intervalMin <= 0 {
		intervalMin = dur
	}
	var slots []TimeSlot
	for _, iv := range available {
		for start := iv.start; start+dur <= iv.end; start += intervalMin {
			slots = append(slots, TimeSlot{
				StartTime: minutesToHHMM(start),
				EndTime:   minutesToHHMM(start + dur),
			})
		}
	}
	return slots, nil
}

// generateMinimizeGaps は最短コース時間を考慮し、空き時間を最小化するように枠を生成する。
// 各空き区間の先頭から「コース所要時間」単位で詰めて配置する。
// 余った端数が最短コース時間未満の場合はその端数を無視し、空きを最小化する。
func generateMinimizeGaps(available []interval, dur, minDur int) ([]TimeSlot, error) {
	var slots []TimeSlot
	for _, iv := range available {
		cur := iv.start
		for cur+dur <= iv.end {
			slots = append(slots, TimeSlot{
				StartTime: minutesToHHMM(cur),
				EndTime:   minutesToHHMM(cur + dur),
			})
			cur += dur
			// 残り時間が最短コース時間未満なら次の区間へ
			if iv.end-cur < minDur {
				break
			}
		}
	}
	return slots, nil
}
