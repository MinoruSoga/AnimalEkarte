package staff

import (
	"encoding/json"
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
	nextDay := date.AddDate(0, 0, 1)

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
				StartsAt:  shiftTS(t, date, "09:00"),
				EndsAt:    shiftTS(t, date, "18:00"),
				Version:   4,
				Notes:     "notes",
				Breaks: []model.ShiftEntryBreak{
					{ID: 5, StartsAt: shiftTS(t, date, "12:00"), EndsAt: shiftTS(t, date, "13:00")},
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
				require.NotNil(t, resp.StartsAt)
				require.NotNil(t, resp.EndsAt)
				assert.Equal(t, "2026-05-28T09:00:00+09:00", *resp.StartsAt)
				assert.Equal(t, "2026-05-28T18:00:00+09:00", *resp.EndsAt)
				assert.Equal(t, int64(4), resp.Version)
				assert.Equal(t, "known", resp.IntervalState)
				assert.Equal(t, "notes", resp.Notes)
				require.Len(t, resp.Breaks, 1)
				require.NotNil(t, resp.Breaks[0].StartsAt)
				require.NotNil(t, resp.Breaks[0].EndsAt)
				assert.Equal(t, "2026-05-28T12:00:00+09:00", *resp.Breaks[0].StartsAt)
				assert.Equal(t, "2026-05-28T13:00:00+09:00", *resp.Breaks[0].EndsAt)
				assert.Equal(t, "known", resp.Breaks[0].IntervalState)
				assert.Equal(t, "山田太郎", resp.StaffName)
			},
		},
		{
			name: "overnight shift keeps explicit instants and is dated by its start side",
			shift: &model.ShiftEntry{
				ID:        2,
				ClinicID:  2,
				StaffID:   3,
				Date:      date,
				ShiftType: model.ShiftTypeFull,
				StartsAt:  shiftTS(t, date, "22:00"),
				EndsAt:    shiftTS(t, nextDay, "07:00"),
				Breaks: []model.ShiftEntryBreak{
					{ID: 6, StartsAt: shiftTS(t, nextDay, "01:00"), EndsAt: shiftTS(t, nextDay, "02:00")},
				},
				CreatedAt: created,
				UpdatedAt: updated,
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Equal(t, "2026-05-28", resp.Date, "date は勤務開始側の JST 日付")
				require.NotNil(t, resp.StartsAt)
				require.NotNil(t, resp.EndsAt)
				assert.Equal(t, "2026-05-28T22:00:00+09:00", *resp.StartsAt)
				assert.Equal(t, "2026-05-29T07:00:00+09:00", *resp.EndsAt)
				assert.Equal(t, "known", resp.IntervalState)
				require.Len(t, resp.Breaks, 1)
				assert.Equal(t, "2026-05-29T01:00:00+09:00", *resp.Breaks[0].StartsAt)
			},
		},
		{
			name: "off shift is known with null times and an empty breaks array",
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
				assert.Nil(t, resp.StartsAt)
				assert.Nil(t, resp.EndsAt)
				assert.Equal(t, "2026-05-28", resp.Date, "終日非勤務は保存された勤務日を返す")
				assert.Equal(t, "known", resp.IntervalState)
				assert.Empty(t, resp.Breaks, "breaks should be empty slice, not nil, when no breaks present")
				assert.NotNil(t, resp.Breaks)
				assert.Empty(t, resp.StaffName)
			},
		},
		{
			// api.yaml ShiftEntry: NULL を休日・勤務可能と扱わず、時刻も補完しない。
			name: "working shift with NULL times is unknown and never defaulted",
			shift: &model.ShiftEntry{
				ID:        11,
				ClinicID:  1,
				StaffID:   4,
				Date:      date,
				ShiftType: model.ShiftTypeFull,
				CreatedAt: created,
				UpdatedAt: updated,
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Nil(t, resp.StartsAt)
				assert.Nil(t, resp.EndsAt)
				assert.Equal(t, "unknown", resp.IntervalState)
				assert.Equal(t, "2026-05-28", resp.Date)
			},
		},
		{
			name: "an unknown break makes the parent unknown and keeps the known side",
			shift: &model.ShiftEntry{
				ID:        12,
				ClinicID:  1,
				StaffID:   4,
				Date:      date,
				ShiftType: model.ShiftTypeFull,
				StartsAt:  shiftTS(t, date, "09:00"),
				EndsAt:    shiftTS(t, date, "18:00"),
				Breaks: []model.ShiftEntryBreak{
					{ID: 7, StartsAt: shiftTS(t, date, "12:00")},
				},
				CreatedAt: created,
				UpdatedAt: updated,
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Equal(t, "unknown", resp.IntervalState)
				require.Len(t, resp.Breaks, 1, "unknown の休憩を配列から除外しない")
				assert.Equal(t, "unknown", resp.Breaks[0].IntervalState)
				require.NotNil(t, resp.Breaks[0].StartsAt)
				assert.Equal(t, "2026-05-28T12:00:00+09:00", *resp.Breaks[0].StartsAt)
				assert.Nil(t, resp.Breaks[0].EndsAt)
			},
		},
		{
			name: "off shift with leftover timestamps is unknown and keeps the stored values",
			shift: &model.ShiftEntry{
				ID:        13,
				ClinicID:  1,
				StaffID:   4,
				Date:      date,
				ShiftType: model.ShiftTypeOff,
				StartsAt:  shiftTS(t, date, "09:00"),
				EndsAt:    shiftTS(t, date, "18:00"),
				CreatedAt: created,
				UpdatedAt: updated,
			},
			check: func(t *testing.T, resp shiftResponse) {
				assert.Equal(t, "unknown", resp.IntervalState)
				require.NotNil(t, resp.StartsAt)
				assert.Equal(t, "2026-05-28T09:00:00+09:00", *resp.StartsAt)
			},
		},
		{
			// AUS-05: Preload may leave Staff nil when assignment is inactive for the clinic.
			name: "nil Staff association does not panic and omits the staff name",
			shift: &model.ShiftEntry{
				ID:        14,
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

func TestToShiftResponse_JSONMatchesTheShiftEntryContract(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	resp := toShiftResponse(&model.ShiftEntry{
		ID:        1,
		ClinicID:  2,
		StaffID:   3,
		Date:      date,
		ShiftType: model.ShiftTypeFull,
		StartsAt:  shiftTS(t, date, "09:00"),
		EndsAt:    shiftTS(t, date, "18:00"),
		Version:   2,
		Breaks:    []model.ShiftEntryBreak{{ID: 5, StartsAt: shiftTS(t, date, "12:00"), EndsAt: shiftTS(t, date, "13:00")}},
		Staff:     &model.Staff{ID: 3, Name: "山田太郎"},
	})

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))

	for _, key := range []string{"id", "clinic_id", "staff_id", "date", "shift_type", "starts_at", "ends_at", "version", "interval_state", "notes", "breaks", "created_at", "updated_at"} {
		assert.Contains(t, decoded, key)
	}
	for _, key := range []string{"start_time", "end_time", "staff"} {
		assert.NotContains(t, decoded, key, "ShiftEntry は旧 wire や staff オブジェクトを返さない")
	}
	breaks, ok := decoded["breaks"].([]any)
	require.True(t, ok)
	require.Len(t, breaks, 1)
	breakItem, ok := breaks[0].(map[string]any)
	require.True(t, ok)
	for _, key := range []string{"starts_at", "ends_at", "interval_state"} {
		assert.Contains(t, breakItem, key)
	}
	assert.Len(t, breakItem, 3, "break の id は返さない")
}

func TestToShiftMutationReceipt(t *testing.T) {
	date := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
	shift := &model.ShiftEntry{ID: 9, ClinicID: 2, StaffID: 3, Date: date, ShiftType: model.ShiftTypeOff, Version: 3}

	receipt := toShiftMutationReceipt("11111111-1111-4111-8111-111111111111", shiftIntentUpdate, shift)

	assert.Equal(t, "11111111-1111-4111-8111-111111111111", receipt.OperationKey)
	assert.Equal(t, "update", receipt.Intent)
	assert.Equal(t, "9", receipt.ShiftID)
	assert.Equal(t, int64(3), receipt.ResultingVersion)
	assert.Equal(t, "9", receipt.Shift.ID)
	assert.Equal(t, int64(3), receipt.Shift.Version)
}
