package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToShiftEntryBreakResponse(t *testing.T) {
	tests := []struct {
		name  string
		input model.ShiftEntryBreak
		want  shiftEntryBreakResponse
	}{
		{
			name:  "maps all fields",
			input: model.ShiftEntryBreak{ID: 5, BreakStart: "12:00", BreakEnd: "13:00"},
			want:  shiftEntryBreakResponse{ID: 5, BreakStart: "12:00", BreakEnd: "13:00"},
		},
		{
			name:  "zero value",
			input: model.ShiftEntryBreak{},
			want:  shiftEntryBreakResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toShiftEntryBreakResponse(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToScheduleEntryResponse(t *testing.T) {
	// withJSTLocal は package 内共有ヘルパー（pet_birthdate_consistency_test.go）。
	// time.Local を JST に固定し、Date/CreatedAt/UpdatedAt のローカル変換を決定的にする。
	withJSTLocal(t)

	workStart := "09:00"
	workEnd := "18:00"
	date := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 14, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		entry *ScheduleEntry
		check func(t *testing.T, got scheduleEntryResponse)
	}{
		{
			name: "with breaks and work times",
			entry: &ScheduleEntry{
				Entry: model.ShiftEntry{
					ID:        1,
					ClinicID:  2,
					StaffID:   3,
					Date:      date,
					ShiftType: model.ShiftTypeFull,
					StartTime: &workStart,
					EndTime:   &workEnd,
					Notes:     "note",
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
				Breaks: []model.ShiftEntryBreak{
					{ID: 10, BreakStart: "12:00", BreakEnd: "13:00"},
				},
			},
			check: func(t *testing.T, got scheduleEntryResponse) {
				assert.Equal(t, uint64(1), got.ID)
				assert.Equal(t, uint64(2), got.ClinicID)
				assert.Equal(t, uint64(3), got.StaffID)
				assert.Equal(t, date.In(time.Local).Format("2006-01-02"), got.Date)
				assert.Equal(t, "full", got.ShiftType)
				assert.Equal(t, &workStart, got.WorkStart)
				assert.Equal(t, &workEnd, got.WorkEnd)
				assert.Equal(t, "note", got.Note)
				assert.Equal(t, createdAt.In(time.Local).Format(time.RFC3339), got.CreatedAt)
				assert.Equal(t, updatedAt.In(time.Local).Format(time.RFC3339), got.UpdatedAt)
				if assert.Len(t, got.Breaks, 1) {
					assert.Equal(t, uint64(10), got.Breaks[0].ID)
					assert.Equal(t, "12:00", got.Breaks[0].BreakStart)
					assert.Equal(t, "13:00", got.Breaks[0].BreakEnd)
				}
			},
		},
		{
			name: "no breaks and nil work times",
			entry: &ScheduleEntry{
				Entry: model.ShiftEntry{
					ID:        2,
					ClinicID:  2,
					StaffID:   4,
					Date:      date,
					ShiftType: model.ShiftTypeOff,
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
				Breaks: nil,
			},
			check: func(t *testing.T, got scheduleEntryResponse) {
				assert.Equal(t, "off", got.ShiftType)
				assert.Nil(t, got.WorkStart)
				assert.Nil(t, got.WorkEnd)
				assert.NotNil(t, got.Breaks, "toScheduleEntryResponse should return an empty slice, not nil")
				assert.Empty(t, got.Breaks)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toScheduleEntryResponse(tt.entry)
			tt.check(t, got)
		})
	}
}
