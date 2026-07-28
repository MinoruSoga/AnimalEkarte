package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- filterApplicableUnavailableTimes (純粋関数) ----

func TestFilterApplicableUnavailableTimes(t *testing.T) {
	// 2026-04-15 は水曜日 (Weekday()==3)
	targetDate := time.Date(2026, 4, 15, 10, 0, 0, 0, config.JST)
	otherDate := time.Date(2026, 4, 16, 10, 0, 0, 0, config.JST)
	wednesday := int8(3)
	thursday := int8(4)

	tests := []struct {
		name  string
		times []model.ReservationTypeUnavailableTime
		date  time.Time
		want  []string // StartTime の一覧で比較（順序含む）
	}{
		{
			name:  "空リストは空を返す",
			times: nil,
			date:  targetDate,
			want:  []string{},
		},
		{
			name: "weekly のみ一致する場合はそれを返す",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &wednesday, StartTime: "12:00", EndTime: "13:00"},
			},
			date: targetDate,
			want: []string{"12:00"},
		},
		{
			name: "weekly の曜日が一致しない場合は除外される",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &thursday, StartTime: "12:00", EndTime: "13:00"},
			},
			date: targetDate,
			want: []string{},
		},
		{
			name: "specific が weekly より優先される",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &wednesday, StartTime: "12:00", EndTime: "13:00"},
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &targetDate, StartTime: "09:00", EndTime: "10:00"},
			},
			date: targetDate,
			want: []string{"09:00"},
		},
		{
			name: "specific の日付が一致しない場合は除外され weekly にフォールバックする",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: &wednesday, StartTime: "12:00", EndTime: "13:00"},
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &otherDate, StartTime: "09:00", EndTime: "10:00"},
			},
			date: targetDate,
			want: []string{"12:00"},
		},
		{
			name: "SpecificDate が nil の specific エントリは無視される",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: nil, StartTime: "09:00", EndTime: "10:00"},
			},
			date: targetDate,
			want: []string{},
		},
		{
			name: "DayOfWeek が nil の weekly エントリは無視される",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeWeekly, DayOfWeek: nil, StartTime: "12:00", EndTime: "13:00"},
			},
			date: targetDate,
			want: []string{},
		},
		{
			name: "複数の specific が一致する場合は全件返す",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &targetDate, StartTime: "09:00", EndTime: "10:00"},
				{UnavailableType: model.UnavailableTypeSpecific, SpecificDate: &targetDate, StartTime: "14:00", EndTime: "15:00"},
			},
			date: targetDate,
			want: []string{"09:00", "14:00"},
		},
		{
			name: "未知の UnavailableType は無視される",
			times: []model.ReservationTypeUnavailableTime{
				{UnavailableType: model.UnavailableType("unknown"), StartTime: "09:00", EndTime: "10:00"},
			},
			date: targetDate,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterApplicableUnavailableTimes(tt.times, tt.date)
			gotStarts := make([]string, 0, len(got))
			for _, g := range got {
				gotStarts = append(gotStarts, g.StartTime)
			}
			assert.Equal(t, tt.want, gotStarts)
		})
	}
}
