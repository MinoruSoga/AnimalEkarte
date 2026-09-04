package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestParseAvailableDatesSettings(t *testing.T) {
	t.Run("success parse JSON", func(t *testing.T) {
		settings, err := ParseAvailableDatesSettings(
			[]byte("[1, 2]"),
			[]byte("[\"2026-01-01\"]"),
			true,
			1, 30, 3,
			"weekday",
		)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, settings.ClosedWeekdays)
		assert.Equal(t, []string{"2026-01-01"}, settings.ClosedDates)
		assert.True(t, settings.NationalHolidayClosed)
		assert.Equal(t, 1, settings.BookingWindowMinDays)
		assert.Equal(t, 30, settings.BookingWindowMaxDays)
		assert.Equal(t, 3, settings.CalendarMonths)
		assert.Equal(t, "weekday", settings.ReservationDayOption)
	})

	t.Run("unmarshal error returns fail-closed error", func(t *testing.T) {
		settings, err := ParseAvailableDatesSettings(
			[]byte("bad json"),
			[]byte("bad json"),
			false,
			1, 30, 3,
			"none",
		)
		assert.Error(t, err)
		assert.Equal(t, AvailableDatesSettings{}, settings)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("壊れたclosed_weekdays JSONはエラーを返す", func(t *testing.T) {
		settings, err := ParseAvailableDatesSettings(
			[]byte("not-valid-json"),
			[]byte("[\"2026-01-01\"]"),
			false,
			1, 30, 3,
			"none",
		)
		assert.Error(t, err)
		assert.Equal(t, AvailableDatesSettings{}, settings)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("unmarshal error does not leak json internals", func(t *testing.T) {
		tests := []struct {
			name               string
			closedWeekdaysJSON []byte
			closedDatesJSON    []byte
		}{
			{
				name:               "壊れたclosed_weekdays JSONは内部エラーを漏らさない",
				closedWeekdaysJSON: []byte("bad json"),
				closedDatesJSON:    []byte("[\"2026-01-01\"]"),
			},
			{
				name:               "壊れたclosed_dates JSONは内部エラーを漏らさない",
				closedWeekdaysJSON: []byte("[1]"),
				closedDatesJSON:    []byte("bad json"),
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				settings, err := ParseAvailableDatesSettings(
					tt.closedWeekdaysJSON,
					tt.closedDatesJSON,
					false,
					1, 30, 3,
					"none",
				)
				assert.Error(t, err)
				assert.Equal(t, AvailableDatesSettings{}, settings)
				assert.True(t, apperrors.IsInvalidInput(err))

				var appErr *apperrors.AppError
				assert.ErrorAs(t, err, &appErr)
				assert.Equal(t, "休診設定の形式が正しくありません", appErr.Message)
				assert.NotContains(t, err.Error(), "invalid character")
				assert.NotContains(t, err.Error(), "json:")
			})
		}
	})
}

// TestOrEmptyJSONArray は orEmptyJSONArray の nil/空/非空バイト列の分岐を直接検証する。
func TestOrEmptyJSONArray(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  []byte
	}{
		{
			name:  "nil slice becomes empty JSON array",
			input: nil,
			want:  []byte("[]"),
		},
		{
			name:  "empty (non-nil) slice becomes empty JSON array",
			input: []byte{},
			want:  []byte("[]"),
		},
		{
			name:  "non-empty slice is returned unchanged",
			input: []byte("[1,2,3]"),
			want:  []byte("[1,2,3]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orEmptyJSONArray(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAvailableDatesResultCap(t *testing.T) {
	assert.Equal(t, 0, availableDatesResultCap(-5))
	assert.Equal(t, 0, availableDatesResultCap(0))
	assert.Equal(t, 30, availableDatesResultCap(30))
	assert.Equal(t, maxAvailableDatesResultCap, availableDatesResultCap(maxAvailableDatesResultCap))
	assert.Equal(t, maxAvailableDatesResultCap, availableDatesResultCap(maxAvailableDatesResultCap+1000))
}

func TestCalcAvailableDates_NegativeWindowDoesNotPanic(t *testing.T) {
	// RSV-04: previously make([]T, 0, negative) panicked; clamp keeps the request alive.
	ctx := context.Background()
	settings := AvailableDatesSettings{
		BookingWindowMinDays: -3,
		BookingWindowMaxDays: -1,
	}
	results, window, err := CalcAvailableDates(ctx, &AvailableDatesInput{Settings: settings})
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.NotEmpty(t, window.Start)
	assert.NotEmpty(t, window.End)
}

func TestCalcAvailableDates(t *testing.T) {
	ctx := context.Background()

	t.Run("basic available dates calculation", func(t *testing.T) {
		settings := AvailableDatesSettings{
			ClosedWeekdays:        []int{0},
			ClosedDates:           []string{time.Now().AddDate(0, 0, 2).Format("2006-01-02")},
			NationalHolidayClosed: false,
			BookingWindowMinDays:  1,
			BookingWindowMaxDays:  5,
			ReservationDayOption:  "anyday",
		}

		input := &AvailableDatesInput{
			Settings: settings,
			TypeID:   1,
			StaffID:  0,
		}

		results, window, err := CalcAvailableDates(ctx, input)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.NotEmpty(t, window.Start)
		assert.NotEmpty(t, window.End)
	})

	t.Run("national holiday closed", func(t *testing.T) {
		settings := AvailableDatesSettings{
			NationalHolidayClosed: true,
			BookingWindowMinDays:  0,
			BookingWindowMaxDays:  2,
		}
		input := &AvailableDatesInput{
			Settings: settings,
		}
		_, _, err := CalcAvailableDates(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("check day options", func(t *testing.T) {
		options := []string{"saturday", "weekday", "anyday", "none", "invalid"}
		for _, opt := range options {
			settings := AvailableDatesSettings{
				BookingWindowMinDays: 1,
				BookingWindowMaxDays: 7,
				ReservationDayOption: opt,
			}
			input := &AvailableDatesInput{
				Settings: settings,
			}
			_, _, err := CalcAvailableDates(ctx, input)
			assert.NoError(t, err)
		}
	})

	t.Run("staff and slot validation", func(t *testing.T) {
		settings := AvailableDatesSettings{
			BookingWindowMinDays: 1,
			BookingWindowMaxDays: 2,
		}

		t.Run("staff all off", func(t *testing.T) {
			input := &AvailableDatesInput{
				Settings: settings,
				TypeID:   1,
				StaffID:  10,
				StaffInputsFn: func(_ context.Context, _ time.Time, _, _ uint64) ([]StaffSlotInput, error) {
					return []StaffSlotInput{
						{
							ScheduleOverride: &StaffScheduleOverride{
								ShiftType: "off",
							},
						},
					}, nil
				},
				SlotSettingsFn: func(_ time.Time) TimeSlotsInput {
					return TimeSlotsInput{}
				},
			}
			res, _, err := CalcAvailableDates(ctx, input)
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
			assert.Equal(t, "staff_off", res[0].Reason)
		})

		t.Run("staff fn error", func(t *testing.T) {
			input := &AvailableDatesInput{
				Settings: settings,
				StaffInputsFn: func(_ context.Context, _ time.Time, _, _ uint64) ([]StaffSlotInput, error) {
					return nil, errors.New("fn error")
				},
				SlotSettingsFn: func(_ time.Time) TimeSlotsInput {
					return TimeSlotsInput{}
				},
			}
			_, _, err := CalcAvailableDates(ctx, input)
			assert.Error(t, err)
		})

		t.Run("generate time slots error", func(t *testing.T) {
			input := &AvailableDatesInput{
				Settings: settings,
				StaffInputsFn: func(_ context.Context, _ time.Time, _, _ uint64) ([]StaffSlotInput, error) {
					return []StaffSlotInput{
						{
							StaffID: 1,
							ExistingResvs: []ExistingReservation{
								{
									StaffID:   1,
									StartTime: "bad-time",
								},
							},
						},
					}, nil
				},
				SlotSettingsFn: func(_ time.Time) TimeSlotsInput {
					return TimeSlotsInput{
						BusinessHours: BusinessHours{
							Start: "0900",
							End:   "1800",
						},
					}
				},
			}
			_, _, err := CalcAvailableDates(ctx, input)
			assert.Error(t, err)
		})

		t.Run("no slots available", func(t *testing.T) {
			input := &AvailableDatesInput{
				Settings: settings,
				StaffInputsFn: func(_ context.Context, _ time.Time, _, _ uint64) ([]StaffSlotInput, error) {
					return []StaffSlotInput{}, nil
				},
				SlotSettingsFn: func(_ time.Time) TimeSlotsInput {
					return TimeSlotsInput{
						BusinessHours: BusinessHours{
							Start: "0900",
							End:   "1800",
						},
						CourseDuration: 30,
					}
				},
				SlotFilterFn: func(_ time.Time, slots []TimeSlot) []TimeSlot {
					return []TimeSlot{}
				},
			}
			res, _, err := CalcAvailableDates(ctx, input)
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
			assert.Equal(t, "no_slots", res[0].Reason)
		})
	})
}
