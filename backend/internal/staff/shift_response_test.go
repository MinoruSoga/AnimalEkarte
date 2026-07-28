package staff

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToShiftResponse(t *testing.T) {
	created := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	startTime := "09:00"
	endTime := "18:00"

	tests := []struct {
		name  string
		shift *model.ShiftEntry
		check func(t *testing.T, resp shiftResponse)
	}{
		{
			name: "full shift with breaks and staff",
			shift: &model.ShiftEntry{
				ID:        1,
				ClinicID:  2,
				StaffID:   3,
				Date:      date,
				ShiftType: model.ShiftType("full"),
				StartTime: &startTime,
				EndTime:   &endTime,
				Notes:     "notes",
				Breaks: []model.ShiftEntryBreak{
					{ID: 5, BreakStart: "12:00", BreakEnd: "13:00"},
				},
				CreatedAt: created,
				UpdatedAt: updated,
				Staff:     &model.Staff{ID: 3, Name: "山田太郎"},
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Equal(t, "1", resp.ID)
				assert.Equal(t, "2", resp.ClinicID)
				assert.Equal(t, "3", resp.StaffID)
				assert.Equal(t, "2026-05-28", resp.Date)
				assert.Equal(t, "09:00", resp.StartTime)
				assert.Equal(t, "18:00", resp.EndTime)
				assert.Equal(t, "notes", resp.Notes)
				require.Len(t, resp.Breaks, 1)
				assert.Equal(t, "5", resp.Breaks[0].ID)
				assert.Equal(t, "12:00", resp.Breaks[0].BreakStart)
				assert.Equal(t, "13:00", resp.Breaks[0].BreakEnd)
				assert.Equal(t, "山田太郎", resp.StaffName)
				require.NotNil(t, resp.Staff)
				assert.Equal(t, uint64(3), resp.Staff.ID)
				assert.Equal(t, "山田太郎", resp.Staff.Name)
			},
		},
		{
			name: "shift without start/end time, breaks, or a named staff",
			shift: &model.ShiftEntry{
				ID:        10,
				ClinicID:  1,
				StaffID:   4,
				Date:      date,
				ShiftType: model.ShiftType("off"),
				CreatedAt: created,
				UpdatedAt: updated,
				Staff:     &model.Staff{},
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Empty(t, resp.StartTime)
				assert.Empty(t, resp.EndTime)
				assert.Empty(t, resp.Breaks, "breaks should be empty slice, not nil, when no breaks present")
				assert.NotNil(t, resp.Breaks)
				assert.Empty(t, resp.StaffName)
				assert.Nil(t, resp.Staff, "Staff must stay nil in response when Staff.ID is zero value")
			},
		},
		{
			// AUS-05: Preload may leave Staff nil when assignment is inactive for the clinic.
			name: "nil Staff association does not panic and omits staff summary",
			shift: &model.ShiftEntry{
				ID:        11,
				ClinicID:  1,
				StaffID:   4,
				Date:      date,
				ShiftType: model.ShiftType("off"),
				CreatedAt: created,
				UpdatedAt: updated,
				Staff:     nil,
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Empty(t, resp.StaffName)
				assert.Nil(t, resp.Staff)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toShiftResponse(tt.shift)
			tt.check(t, resp)
		})
	}
}
