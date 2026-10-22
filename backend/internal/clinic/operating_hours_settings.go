package clinic

// 共通医院時間 aggregate の Settings（A03 §2.1）。週次営業区間・個別休診日・祝日方針の検証済み値。
// 行なしを全日営業/全日休診へ補完しない。集合 field は全置換で、未指定 field の default は持たない。
// 保存表 clinic_operating_hours の DDL/model と API 正本は①（B/A）受領後に結線する。

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// OperatingHoursSchemaVersion は保存・GET の schema_version。
const OperatingHoursSchemaVersion = "clinic-operating-hours/v1"

// closedDateReasonMaxRunes は既存休診理由（clinic_holiday_handler.go の max=500）と同じ上限。
const closedDateReasonMaxRunes = 500

const daysPerWeek = 7

// IntervalInput は厳密 HH:MM の区間入力。
type IntervalInput struct {
	Start string
	End   string
}

// WeeklyHoursInput は1曜日分の入力。Intervals が空なら明示休診。
type WeeklyHoursInput struct {
	Weekday   int
	Intervals []IntervalInput
}

// ClosedDateInput は個別休診日の入力。Reason は空可。
type ClosedDateInput struct {
	Date   string
	Reason string
}

// OperatingHoursFields は Settings の field 単位の入力。nil は「その field を送っていない」。
// 初期化（initialize）は全 field 必須、更新（update）の changes は1 field 以上。
// JSON null・未知 key の拒否は HTTP 境界の closed decode が行い、ここへは渡さない。
type OperatingHoursFields struct {
	WeeklyHours           *[]WeeklyHoursInput
	ClosedDates           *[]ClosedDateInput
	NationalHolidayClosed *bool
}

func (f OperatingHoursFields) isComplete() bool {
	return f.WeeklyHours != nil && f.ClosedDates != nil && f.NationalHolidayClosed != nil
}

func (f OperatingHoursFields) isEmpty() bool {
	return f.WeeklyHours == nil && f.ClosedDates == nil && f.NationalHolidayClosed == nil
}

// ClosedDate は検証済みの個別休診日。
type ClosedDate struct {
	Date   CalendarDate
	Reason string
}

// OperatingHoursSettings は検証済みの共通医院時間。値は不変で、getter は複製を返す。
type OperatingHoursSettings struct {
	weekly                [daysPerWeek][]TimeInterval
	closedDates           []ClosedDate
	nationalHolidayClosed bool
}

// NewOperatingHoursSettings は全 field 明示の入力から Settings を作る（初期化・確認用）。
func NewOperatingHoursSettings(f OperatingHoursFields) (OperatingHoursSettings, error) {
	if !f.isComplete() {
		return OperatingHoursSettings{}, apperrors.WrapInvalidInput(
			"weekly_hours・closed_dates・national_holiday_closed はすべて明示する必要があります")
	}
	return OperatingHoursSettings{}.withFields(f)
}

// ApplyChanges は changes の field だけを全置換した新しい Settings を返す。changes は1 field 以上。
func (s OperatingHoursSettings) ApplyChanges(changes OperatingHoursFields) (OperatingHoursSettings, error) {
	if changes.isEmpty() {
		return OperatingHoursSettings{}, apperrors.WrapInvalidInput("changes は1項目以上必要です")
	}
	return s.withFields(changes)
}

func (s OperatingHoursSettings) withFields(f OperatingHoursFields) (OperatingHoursSettings, error) {
	next := s.clone()
	if f.WeeklyHours != nil {
		weekly, err := parseWeeklyHours(*f.WeeklyHours)
		if err != nil {
			return OperatingHoursSettings{}, err
		}
		next.weekly = weekly
	}
	if f.ClosedDates != nil {
		closed, err := parseClosedDates(*f.ClosedDates)
		if err != nil {
			return OperatingHoursSettings{}, err
		}
		next.closedDates = closed
	}
	if f.NationalHolidayClosed != nil {
		next.nationalHolidayClosed = *f.NationalHolidayClosed
	}
	return next, nil
}

func parseWeeklyHours(in []WeeklyHoursInput) ([daysPerWeek][]TimeInterval, error) {
	var weekly [daysPerWeek][]TimeInterval
	if len(in) != daysPerWeek {
		return weekly, apperrors.WrapInvalidInput("weekly_hours は曜日 0〜6 を各1件ずつ含む必要があります")
	}
	for i, day := range in {
		if day.Weekday != i {
			return weekly, apperrors.WrapInvalidInput("weekly_hours は曜日 0〜6 の昇順で各1件である必要があります")
		}
		intervals := make([]TimeInterval, 0, len(day.Intervals))
		for _, raw := range day.Intervals {
			iv, err := NewTimeInterval(raw.Start, raw.End)
			if err != nil {
				return weekly, err
			}
			if n := len(intervals); n > 0 && iv.Start < intervals[n-1].End {
				return weekly, apperrors.WrapInvalidInput(fmt.Sprintf(
					"weekly_hours の曜日 %d の区間は開始昇順かつ重複なしである必要があります", day.Weekday))
			}
			intervals = append(intervals, iv)
		}
		weekly[i] = intervals
	}
	return weekly, nil
}

func parseClosedDates(in []ClosedDateInput) ([]ClosedDate, error) {
	out := make([]ClosedDate, 0, len(in))
	for _, raw := range in {
		cd, err := newClosedDate(raw.Date, raw.Reason)
		if err != nil {
			return nil, err
		}
		if n := len(out); n > 0 && cd.Date.Compare(out[n-1].Date) <= 0 {
			return nil, apperrors.WrapInvalidInput("closed_dates は日付の一意昇順である必要があります")
		}
		out = append(out, cd)
	}
	return out, nil
}

func newClosedDate(date, reason string) (ClosedDate, error) {
	d, err := ParseCalendarDate(date)
	if err != nil {
		return ClosedDate{}, err
	}
	if err := validateClosedDateReason(reason); err != nil {
		return ClosedDate{}, err
	}
	return ClosedDate{Date: d, Reason: reason}, nil
}

func validateClosedDateReason(reason string) error {
	if !utf8.ValidString(reason) {
		return apperrors.WrapInvalidInput("休診理由は有効な UTF-8 である必要があります")
	}
	if utf8.RuneCountInString(reason) > closedDateReasonMaxRunes {
		return apperrors.WrapInvalidInput(fmt.Sprintf("休診理由は %d 文字以内である必要があります", closedDateReasonMaxRunes))
	}
	return nil
}

func (s OperatingHoursSettings) clone() OperatingHoursSettings {
	next := OperatingHoursSettings{
		closedDates:           slices.Clone(s.closedDates),
		nationalHolidayClosed: s.nationalHolidayClosed,
	}
	for i := range s.weekly {
		next.weekly[i] = slices.Clone(s.weekly[i])
	}
	return next
}

// WeeklyIntervals は曜日 w の営業区間の複製を返す。空は明示休診。
func (s OperatingHoursSettings) WeeklyIntervals(w time.Weekday) []TimeInterval {
	if w < time.Sunday || w > time.Saturday {
		return []TimeInterval{}
	}
	return append([]TimeInterval{}, s.weekly[w]...)
}

// ClosedDates は個別休診日の複製（日付昇順）を返す。
func (s OperatingHoursSettings) ClosedDates() []ClosedDate {
	return append([]ClosedDate{}, s.closedDates...)
}

// NationalHolidayClosed は祝日休診方針。
func (s OperatingHoursSettings) NationalHolidayClosed() bool {
	return s.nationalHolidayClosed
}

// ClosedDateOn は暦日 d の個別休診を返す。
func (s OperatingHoursSettings) ClosedDateOn(d CalendarDate) (ClosedDate, bool) {
	i, found := slices.BinarySearchFunc(s.closedDates, d, func(cd ClosedDate, t CalendarDate) int {
		return cd.Date.Compare(t)
	})
	if !found {
		return ClosedDate{}, false
	}
	return s.closedDates[i], true
}

// IntervalsOn は暦日 d の共通基準 C(d) を返す。個別休診日と（方針が休診なら）祝日は正常な空。
// isNationalHoliday は既存の祝日判定から呼出側が渡す。旧 clinic_holidays は参照しない（§2.4）。
func (s OperatingHoursSettings) IntervalsOn(d CalendarDate, isNationalHoliday bool) []TimeInterval {
	if _, closed := s.ClosedDateOn(d); closed {
		return []TimeInterval{}
	}
	if isNationalHoliday && s.nationalHolidayClosed {
		return []TimeInterval{}
	}
	return s.WeeklyIntervals(d.Weekday())
}

// Equal は意味値が等しいかを返す（no-op 判定用）。
func (s OperatingHoursSettings) Equal(o OperatingHoursSettings) bool {
	if s.nationalHolidayClosed != o.nationalHolidayClosed || !slices.Equal(s.closedDates, o.closedDates) {
		return false
	}
	for i := range s.weekly {
		if !slices.Equal(s.weekly[i], o.weekly[i]) {
			return false
		}
	}
	return true
}

// WithClosedDate は個別休診日 d を追加した Settings を返す。同日同理由は changed=false（no-op）、
// 同日異理由は ErrClosedDateReasonConflict で拒否し、無言の upsert はしない。
func (s OperatingHoursSettings) WithClosedDate(d CalendarDate, reason string) (OperatingHoursSettings, bool, error) {
	if err := validateClosedDateReason(reason); err != nil {
		return OperatingHoursSettings{}, false, err
	}
	if existing, ok := s.ClosedDateOn(d); ok {
		if existing.Reason == reason {
			return s.clone(), false, nil
		}
		return OperatingHoursSettings{}, false, ErrClosedDateReasonConflict
	}
	next := s.clone()
	i, _ := slices.BinarySearchFunc(next.closedDates, d, func(cd ClosedDate, t CalendarDate) int {
		return cd.Date.Compare(t)
	})
	next.closedDates = slices.Insert(next.closedDates, i, ClosedDate{Date: d, Reason: reason})
	return next, true, nil
}

// WithoutClosedDate は個別休診日 d を除いた Settings を返す。既にない日は changed=false（no-op）。
func (s OperatingHoursSettings) WithoutClosedDate(d CalendarDate) (OperatingHoursSettings, bool) {
	next := s.clone()
	i, found := slices.BinarySearchFunc(next.closedDates, d, func(cd ClosedDate, t CalendarDate) int {
		return cd.Date.Compare(t)
	})
	if !found {
		return next, false
	}
	next.closedDates = slices.Delete(next.closedDates, i, i+1)
	return next, true
}

type operatingHoursIntervalWire struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type operatingHoursWeekdayWire struct {
	Weekday   int                          `json:"weekday"`
	Intervals []operatingHoursIntervalWire `json:"intervals"`
}

type operatingHoursClosedDateWire struct {
	Date   string `json:"date"`
	Reason string `json:"reason"`
}

type operatingHoursSettingsWire struct {
	WeeklyHours           []operatingHoursWeekdayWire    `json:"weekly_hours"`
	ClosedDates           []operatingHoursClosedDateWire `json:"closed_dates"`
	NationalHolidayClosed bool                           `json:"national_holiday_closed"`
}

// MarshalJSON は exact wire `{weekly_hours,closed_dates,national_holiday_closed}` を決定的な順序で返す。
// GET・preview・監査 before/after・command の正規 bytes で同じ表現を使う。
func (s OperatingHoursSettings) MarshalJSON() ([]byte, error) {
	wire := operatingHoursSettingsWire{
		WeeklyHours:           make([]operatingHoursWeekdayWire, 0, daysPerWeek),
		ClosedDates:           make([]operatingHoursClosedDateWire, 0, len(s.closedDates)),
		NationalHolidayClosed: s.nationalHolidayClosed,
	}
	for w := range s.weekly {
		intervals := make([]operatingHoursIntervalWire, 0, len(s.weekly[w]))
		for _, iv := range s.weekly[w] {
			intervals = append(intervals, operatingHoursIntervalWire{Start: iv.Start.String(), End: iv.End.String()})
		}
		wire.WeeklyHours = append(wire.WeeklyHours, operatingHoursWeekdayWire{Weekday: w, Intervals: intervals})
	}
	for _, cd := range s.closedDates {
		wire.ClosedDates = append(wire.ClosedDates, operatingHoursClosedDateWire{Date: cd.Date.String(), Reason: cd.Reason})
	}
	return json.Marshal(wire)
}
