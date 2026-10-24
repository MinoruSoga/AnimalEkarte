package sharedkernel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// jstAt は 2026-10-<day> の JST 時刻を返す。
func jstAt(day, hour, minute int) *time.Time {
	ts := time.Date(2026, 10, day, hour, minute, 0, 0, config.JST)
	return &ts
}

func windowOf(from, to *time.Time) ShiftWindow {
	return ShiftWindow{StartsAt: *from, EndsAt: *to}
}

func TestValidateShiftInterval(t *testing.T) {
	tests := []struct {
		name      string
		shiftType model.ShiftType
		startsAt  *time.Time
		endsAt    *time.Time
		wantErr   bool
	}{
		{name: "same-day working shift", shiftType: model.ShiftTypeFull, startsAt: jstAt(21, 9, 0), endsAt: jstAt(21, 18, 0)},
		{name: "overnight 22:00 to 07:00 next day", shiftType: model.ShiftTypeFull, startsAt: jstAt(21, 22, 0), endsAt: jstAt(22, 7, 0)},
		{name: "ends_at equal to starts_at", shiftType: model.ShiftTypeMorning, startsAt: jstAt(21, 9, 0), endsAt: jstAt(21, 9, 0), wantErr: true},
		{name: "earlier clock on the same date is not read as next day", shiftType: model.ShiftTypeFull, startsAt: jstAt(21, 22, 0), endsAt: jstAt(21, 7, 0), wantErr: true},
		{name: "working shift without starts_at", shiftType: model.ShiftTypeFull, endsAt: jstAt(21, 18, 0), wantErr: true},
		{name: "working shift without ends_at", shiftType: model.ShiftTypeAfternoon, startsAt: jstAt(21, 13, 0), wantErr: true},
		{name: "working shift without both", shiftType: model.ShiftTypeFull, wantErr: true},
		{name: "off with both NULL", shiftType: model.ShiftTypeOff},
		{name: "paid_leave with both NULL", shiftType: model.ShiftTypePaidLeave},
		{name: "off with timestamps", shiftType: model.ShiftTypeOff, startsAt: jstAt(21, 9, 0), endsAt: jstAt(21, 18, 0), wantErr: true},
		{name: "paid_leave with one timestamp", shiftType: model.ShiftTypePaidLeave, startsAt: jstAt(21, 9, 0), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShiftInterval(tt.shiftType, tt.startsAt, tt.endsAt)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestValidateShiftDate(t *testing.T) {
	jstDate := time.Date(2026, 10, 21, 0, 0, 0, 0, config.JST)
	// DATE 列は UTC 0 時で読まれうる。JST 暦日で比較する。
	utcDate := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)

	assert.NoError(t, ValidateShiftDate(jstDate, jstAt(21, 22, 0)), "日跨ぎ勤務は開始側の日付で登録する")
	assert.NoError(t, ValidateShiftDate(utcDate, jstAt(21, 0, 30)))
	assert.NoError(t, ValidateShiftDate(jstDate, nil), "終日非勤務は starts_at を持たない")
	err := ValidateShiftDate(jstDate, jstAt(22, 7, 0))
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Error(t, ValidateShiftDate(jstDate, jstAt(20, 23, 59)))
}

func TestValidateShiftBreaks(t *testing.T) {
	startsAt, endsAt := jstAt(21, 22, 0), jstAt(22, 7, 0)
	tests := []struct {
		name    string
		breaks  []ShiftWindow
		wantErr bool
	}{
		{name: "no breaks"},
		{name: "break after midnight inside an overnight shift", breaks: []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0))}},
		{name: "breaks touching the shift boundaries", breaks: []ShiftWindow{windowOf(jstAt(21, 22, 0), jstAt(21, 23, 0)), windowOf(jstAt(22, 6, 0), jstAt(22, 7, 0))}},
		{name: "adjacent breaks do not overlap", breaks: []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0)), windowOf(jstAt(22, 2, 0), jstAt(22, 3, 0))}},
		{name: "unsorted breaks are checked after sorting", breaks: []ShiftWindow{windowOf(jstAt(22, 3, 0), jstAt(22, 4, 0)), windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0))}},
		{name: "break starts before the shift", breaks: []ShiftWindow{windowOf(jstAt(21, 21, 0), jstAt(21, 23, 0))}, wantErr: true},
		{name: "break ends after the shift", breaks: []ShiftWindow{windowOf(jstAt(22, 6, 0), jstAt(22, 8, 0))}, wantErr: true},
		{name: "break with ends_at equal to starts_at", breaks: []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 1, 0))}, wantErr: true},
		{name: "break with ends_at before starts_at", breaks: []ShiftWindow{windowOf(jstAt(22, 2, 0), jstAt(22, 1, 0))}, wantErr: true},
		{name: "overlapping breaks", breaks: []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 30)), windowOf(jstAt(22, 2, 0), jstAt(22, 3, 0))}, wantErr: true},
		{name: "breaks with the same start", breaks: []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0)), windowOf(jstAt(22, 1, 0), jstAt(22, 1, 30))}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShiftBreaks(startsAt, endsAt, tt.breaks)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err))
				return
			}
			assert.NoError(t, err)
		})
	}

	t.Run("breaks without a shift interval", func(t *testing.T) {
		err := ValidateShiftBreaks(nil, nil, []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0))})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

func TestValidateShiftEntryShape(t *testing.T) {
	date := time.Date(2026, 10, 21, 0, 0, 0, 0, config.JST)
	nightBreak := []ShiftWindow{windowOf(jstAt(22, 1, 0), jstAt(22, 2, 0))}

	t.Run("overnight shift with a break after midnight", func(t *testing.T) {
		assert.NoError(t, ValidateShiftEntryShape(model.ShiftTypeFull, date, jstAt(21, 22, 0), jstAt(22, 7, 0), nightBreak))
	})
	t.Run("off shift is all-day non-working with no breaks", func(t *testing.T) {
		assert.NoError(t, ValidateShiftEntryShape(model.ShiftTypeOff, date, nil, nil, nil))
		err := ValidateShiftEntryShape(model.ShiftTypePaidLeave, date, nil, nil, nightBreak)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("date must be the JST date of starts_at", func(t *testing.T) {
		err := ValidateShiftEntryShape(model.ShiftTypeFull, date, jstAt(22, 9, 0), jstAt(22, 18, 0), nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
	t.Run("break outside the shift is rejected", func(t *testing.T) {
		err := ValidateShiftEntryShape(model.ShiftTypeFull, date, jstAt(21, 22, 0), jstAt(22, 7, 0), []ShiftWindow{windowOf(jstAt(22, 7, 0), jstAt(22, 8, 0))})
		require.Error(t, err)
	})
}

func TestShiftEntryIntervalState(t *testing.T) {
	knownBreak := model.ShiftEntryBreak{StartsAt: jstAt(22, 1, 0), EndsAt: jstAt(22, 2, 0)}
	halfBreak := model.ShiftEntryBreak{StartsAt: jstAt(22, 1, 0)}
	tests := []struct {
		name   string
		entry  *model.ShiftEntry
		breaks []model.ShiftEntryBreak
		want   string
	}{
		{name: "known overnight shift with a known break", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeFull, StartsAt: jstAt(21, 22, 0), EndsAt: jstAt(22, 7, 0)}, breaks: []model.ShiftEntryBreak{knownBreak}, want: ShiftIntervalStateKnown},
		{name: "working shift with NULL timestamps", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeFull}, want: ShiftIntervalStateUnknown},
		{name: "working shift with only starts_at", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeMorning, StartsAt: jstAt(21, 9, 0)}, want: ShiftIntervalStateUnknown},
		{name: "known shift whose break has a NULL side", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeFull, StartsAt: jstAt(21, 22, 0), EndsAt: jstAt(22, 7, 0)}, breaks: []model.ShiftEntryBreak{halfBreak}, want: ShiftIntervalStateUnknown},
		{name: "off without timestamps or breaks", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeOff}, want: ShiftIntervalStateKnown},
		{name: "paid_leave without timestamps or breaks", entry: &model.ShiftEntry{ShiftType: model.ShiftTypePaidLeave}, want: ShiftIntervalStateKnown},
		{name: "off with leftover timestamps", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeOff, StartsAt: jstAt(21, 9, 0), EndsAt: jstAt(21, 18, 0)}, want: ShiftIntervalStateUnknown},
		{name: "off with a leftover break", entry: &model.ShiftEntry{ShiftType: model.ShiftTypeOff}, breaks: []model.ShiftEntryBreak{knownBreak}, want: ShiftIntervalStateUnknown},
		{name: "nil entry", want: ShiftIntervalStateUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ShiftEntryIntervalState(tt.entry, tt.breaks))
		})
	}
}

func TestShiftBreakIntervalState(t *testing.T) {
	assert.Equal(t, ShiftIntervalStateKnown, ShiftBreakIntervalState(model.ShiftEntryBreak{StartsAt: jstAt(22, 1, 0), EndsAt: jstAt(22, 2, 0)}))
	assert.Equal(t, ShiftIntervalStateUnknown, ShiftBreakIntervalState(model.ShiftEntryBreak{EndsAt: jstAt(22, 2, 0)}))
	assert.Equal(t, ShiftIntervalStateUnknown, ShiftBreakIntervalState(model.ShiftEntryBreak{}))
}
