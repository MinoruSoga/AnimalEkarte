package clinic

// 共通医院時間（A03・emr-clinic-line-time-selected-contract-20261022 §1）の時刻・区間・暦日の値型。
// DB model・HTTP DTO・共有操作 Q に依存しない純粋な領域値で、LINE 受付（reservation）と
// 担当者別開始 policy も同じ型を使う。HH:MM の wire 文字列は共通医院時間専用であり、
// 既存 LINE の HHMM 表記へ機械転用しない（LINE 側は境界 adapter で解釈する）。

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

// errZeroCalendarDate はゼロ値の CalendarDate を使おうとした誤用（内部エラー）。
// 入力不正ではないため apperrors.WrapInvalidInput にはせず、OperatingHoursRejection も確定拒否へ写像しない。
var errZeroCalendarDate = errors.New("calendar date was not built by ParseCalendarDate or CalendarDateOf")

// maxEchoedInputBytes は検証エラーへ写すクライアント入力の最大バイト数。長い入力をメッセージへ複製しない。
const maxEchoedInputBytes = 16

// echoInput はエラーメッセージ用に raw を先頭 maxEchoedInputBytes バイトへ切り詰めて %q 化する。
func echoInput(raw string) string {
	if len(raw) > maxEchoedInputBytes {
		return fmt.Sprintf("%q…", raw[:maxEchoedInputBytes])
	}
	return fmt.Sprintf("%q", raw)
}

// MinutesPerDay は JST 暦日1日の分数。区間終端の 24:00 はこの値で表す。
const MinutesPerDay = 24 * 60

// ClockMinute は JST 暦日の 00:00 からの分数。開始は 0..1439、終端は 1..1440。
type ClockMinute int

// String は厳密 HH:MM を返す。終端 1440 は "24:00"。
func (m ClockMinute) String() string {
	return fmt.Sprintf("%02d:%02d", int(m)/60, int(m)%60)
}

// ParseClockStart は開始時刻（00:00..23:59）の厳密 HH:MM を解釈する。
func ParseClockStart(raw string) (ClockMinute, error) {
	m, err := parseStrictHHMM(raw)
	if err != nil {
		return 0, err
	}
	if m >= MinutesPerDay {
		return 0, apperrors.WrapInvalidInput(fmt.Sprintf("開始時刻 %q は 00:00〜23:59 である必要があります", raw))
	}
	return m, nil
}

// ParseClockEnd は終了時刻（00:01..24:00）の厳密 HH:MM を解釈する。
// 24:00 は当日の終端であり、実日時は翌 JST 暦日 00:00（InstantsOn）として扱う。
func ParseClockEnd(raw string) (ClockMinute, error) {
	m, err := parseStrictHHMM(raw)
	if err != nil {
		return 0, err
	}
	if m == 0 {
		return 0, apperrors.WrapInvalidInput(fmt.Sprintf("終了時刻 %q は 00:01〜24:00 である必要があります", raw))
	}
	return m, nil
}

// parseStrictHHMM は "HH:MM"（ASCII 数字・コロン固定・5バイト）だけを受け付け、00:00..24:00 を返す。
func parseStrictHHMM(raw string) (ClockMinute, error) {
	invalid := func() error {
		return apperrors.WrapInvalidInput(fmt.Sprintf("時刻 %s は HH:MM 形式である必要があります", echoInput(raw)))
	}
	if len(raw) != 5 || raw[2] != ':' {
		return 0, invalid()
	}
	for _, i := range []int{0, 1, 3, 4} {
		if raw[i] < '0' || raw[i] > '9' {
			return 0, invalid()
		}
	}
	h := int(raw[0]-'0')*10 + int(raw[1]-'0')
	mm := int(raw[3]-'0')*10 + int(raw[4]-'0')
	if mm > 59 || h > 24 || (h == 24 && mm != 0) {
		return 0, invalid()
	}
	return ClockMinute(h*60 + mm), nil
}

// TimeInterval は同一 JST 暦日内の半開区間 [Start, End)。日跨ぎは隣接日の2区間で表し、
// 片側から自動生成しない。
//
// 前提条件: フィールドは NewTimeInterval（または ParseClockStart／ParseClockEnd）で得た値でなければならない
// （0 <= Start < End <= MinutesPerDay）。リテラルで作った不正値は検証されず、区間代数は結果を保証しない。
type TimeInterval struct {
	Start ClockMinute
	End   ClockMinute
}

// NewTimeInterval は厳密 HH:MM の開始・終了から区間を作る（start < end 必須）。
func NewTimeInterval(start, end string) (TimeInterval, error) {
	s, err := ParseClockStart(start)
	if err != nil {
		return TimeInterval{}, err
	}
	e, err := ParseClockEnd(end)
	if err != nil {
		return TimeInterval{}, err
	}
	if s >= e {
		return TimeInterval{}, apperrors.WrapInvalidInput(fmt.Sprintf("区間 %s-%s は開始 < 終了である必要があります", start, end))
	}
	return TimeInterval{Start: s, End: e}, nil
}

// Contains は開始時刻 t が区間内（Start <= t < End）かを返す。
func (iv TimeInterval) Contains(t ClockMinute) bool {
	return iv.Start <= t && t < iv.End
}

// InstantsOn は暦日 date 上の実日時 [start, end) を返す。End=24:00 は翌 JST 暦日 00:00。
// レジ営業日・日界へは付け替えない。
func (iv TimeInterval) InstantsOn(date CalendarDate) (time.Time, time.Time) {
	base := date.StartInstant()
	return base.Add(time.Duration(iv.Start) * time.Minute), base.Add(time.Duration(iv.End) * time.Minute)
}

// NormalizeIntervals は区間を開始昇順に並べ、重複・隣接を併合した新しいスライスを返す。
// 空区間は捨てる。引数は変更しない。
func NormalizeIntervals(in []TimeInterval) []TimeInterval {
	sorted := slices.Clone(in)
	slices.SortFunc(sorted, func(a, b TimeInterval) int { return int(a.Start - b.Start) })
	out := make([]TimeInterval, 0, len(sorted))
	for _, iv := range sorted {
		if iv.Start >= iv.End {
			continue
		}
		if n := len(out); n > 0 && iv.Start <= out[n-1].End {
			if iv.End > out[n-1].End {
				out[n-1].End = iv.End
			}
			continue
		}
		out = append(out, iv)
	}
	return out
}

// IntersectIntervals は a ∩ b を正規化済み区間で返す。
func IntersectIntervals(a, b []TimeInterval) []TimeInterval {
	na, nb := NormalizeIntervals(a), NormalizeIntervals(b)
	out := make([]TimeInterval, 0)
	for i, j := 0, 0; i < len(na) && j < len(nb); {
		start := max(na[i].Start, nb[j].Start)
		end := min(na[i].End, nb[j].End)
		if start < end {
			out = append(out, TimeInterval{Start: start, End: end})
		}
		if na[i].End < nb[j].End {
			i++
		} else {
			j++
		}
	}
	return out
}

// SubtractIntervals は base から excl を除いた正規化済み区間を返す。
func SubtractIntervals(base, excl []TimeInterval) []TimeInterval {
	result := NormalizeIntervals(base)
	for _, ex := range NormalizeIntervals(excl) {
		next := make([]TimeInterval, 0, len(result))
		for _, iv := range result {
			if ex.End <= iv.Start || ex.Start >= iv.End {
				next = append(next, iv)
				continue
			}
			if iv.Start < ex.Start {
				next = append(next, TimeInterval{Start: iv.Start, End: ex.Start})
			}
			if ex.End < iv.End {
				next = append(next, TimeInterval{Start: ex.End, End: iv.End})
			}
		}
		result = next
	}
	return result
}

// IntervalsCover は inner の全区間が outer の和集合に含まれるかを返す。
func IntervalsCover(outer, inner []TimeInterval) bool {
	return len(SubtractIntervals(inner, outer)) == 0
}

// CalendarDate は JST 暦日。時刻・タイムゾーンを持たない比較可能な値。
//
// ゼロ値は暦日ではなく無効（IsZero）で、ParseCalendarDate／CalendarDateOf だけが有効な値を作る。
// 年の範囲は厳密な 4 桁 Gregorian（0000–9999 のうち time が解釈できるもの）以外に制限しない。契約が上下限を
// 定めていないため、独自の年範囲を発明しない。
type CalendarDate struct {
	valid bool
	year  int
	month time.Month
	day   int
}

// ParseCalendarDate は厳密 Gregorian YYYY-MM-DD を解釈する（2026-02-30 や桁不足は拒否）。
func ParseCalendarDate(raw string) (CalendarDate, error) {
	t, err := time.ParseInLocation(time.DateOnly, raw, config.JST)
	if err != nil || t.Format(time.DateOnly) != raw {
		return CalendarDate{}, apperrors.WrapInvalidInput(fmt.Sprintf("日付 %s は YYYY-MM-DD 形式の実在日である必要があります", echoInput(raw)))
	}
	return CalendarDate{valid: true, year: t.Year(), month: t.Month(), day: t.Day()}, nil
}

// CalendarDateOf は実日時 t を JST 暦日へ変換する。
func CalendarDateOf(t time.Time) CalendarDate {
	j := t.In(config.JST)
	return CalendarDate{valid: true, year: j.Year(), month: j.Month(), day: j.Day()}
}

// IsZero はゼロ値（無効な暦日）かを返す。休診日の設定・検索・営業区間の評価はゼロ値を拒否する。
func (d CalendarDate) IsZero() bool {
	return !d.valid
}

// StartInstant は暦日の JST 00:00。
func (d CalendarDate) StartInstant() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, config.JST)
}

// String は YYYY-MM-DD を返す。ゼロ値は空文字（偽の日付を作らない）。
func (d CalendarDate) String() string {
	if !d.valid {
		return ""
	}
	return d.StartInstant().Format(time.DateOnly)
}

// Weekday は JST の曜日（0=日曜..6=土曜）。
func (d CalendarDate) Weekday() time.Weekday {
	return d.StartInstant().Weekday()
}

// Compare は d と o の前後を -1/0/+1 で返す。
func (d CalendarDate) Compare(o CalendarDate) int {
	if c := cmp.Compare(d.year, o.year); c != 0 {
		return c
	}
	if c := cmp.Compare(d.month, o.month); c != 0 {
		return c
	}
	return cmp.Compare(d.day, o.day)
}
