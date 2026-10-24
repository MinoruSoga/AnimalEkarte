package staff

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

func TestNewListShiftEntriesQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   listShiftEntriesQuery
	}{
		{
			name:   "parses date and staff_id",
			values: url.Values{"date": {"2026-05"}, "staff_id": {"10"}},
			want:   listShiftEntriesQuery{Date: "2026-05", StaffID: "10"},
		},
		{
			name:   "empty values yield zero query",
			values: url.Values{},
			want:   listShiftEntriesQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newListShiftEntriesQuery(tt.values))
		})
	}
}

func TestNewOnDutyStaffsQuery(t *testing.T) {
	tests := []struct {
		name   string
		values url.Values
		want   onDutyStaffsQuery
	}{
		{
			name:   "parses date",
			values: url.Values{"date": {"2026-05-28"}},
			want:   onDutyStaffsQuery{Date: "2026-05-28"},
		},
		{
			name:   "empty values yield empty date",
			values: url.Values{},
			want:   onDutyStaffsQuery{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newOnDutyStaffsQuery(tt.values))
		})
	}
}

func TestOnDutyStaffsQuery_ToDate(t *testing.T) {
	tests := []struct {
		name    string
		query   onDutyStaffsQuery
		wantErr bool
	}{
		{
			name:  "valid date",
			query: onDutyStaffsQuery{Date: "2026-05-28"},
		},
		{
			name:    "missing date returns error",
			query:   onDutyStaffsQuery{Date: ""},
			wantErr: true,
		},
		{
			name:    "invalid format returns error",
			query:   onDutyStaffsQuery{Date: "2026/05/28"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := tt.query.toDate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, date.IsZero())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.query.Date, date.Format(time.DateOnly))
		})
	}
}

func TestListShiftEntriesQuery_ToServiceFilters(t *testing.T) {
	filters, err := (&listShiftEntriesQuery{
		Date:    "2026-05",
		StaffID: "10",
	}).toServiceFilters()
	if err != nil {
		t.Fatalf("toServiceFilters returned error: %v", err)
	}

	if filters.YearMonth != "2026-05" {
		t.Fatalf("YearMonth = %q, want 2026-05", filters.YearMonth)
	}
	if filters.StaffID == nil || *filters.StaffID != 10 {
		t.Fatalf("StaffID = %v, want 10", filters.StaffID)
	}
}

func TestListShiftEntriesQuery_ToServiceFilters_InvalidStaffID(t *testing.T) {
	filters, err := (&listShiftEntriesQuery{StaffID: "abc"}).toServiceFilters()
	if err == nil {
		t.Fatal("toServiceFilters returned nil error")
	}
	if filters != (listShiftEntriesFilters{}) {
		t.Fatalf("filters = %#v, want zero value", filters)
	}
	if !apperrors.IsInvalidInput(err) {
		t.Fatalf("error = %v, want invalid input", err)
	}
}

const testShiftOperationKey = "11111111-1111-4111-8111-111111111111"

// newCreateShiftBody は夜勤（22:00 JST から翌 07:00 JST）の有効な作成本文を返す。
func newCreateShiftBody() map[string]any {
	body := map[string]any{}
	body["operation_key"] = testShiftOperationKey
	body["staff_id"] = "10"
	body["date"] = "2026-10-21"
	body["shift_type"] = "full"
	body["starts_at"] = "2026-10-21T22:00:00+09:00"
	body["ends_at"] = "2026-10-22T07:00:00+09:00"
	body["breaks"] = []any{map[string]any{
		"starts_at": "2026-10-22T01:00:00+09:00",
		"ends_at":   "2026-10-22T02:00:00+09:00",
	}}
	body["notes"] = "夜勤"
	return body
}

// newUpdateShiftBody は終日非勤務（off）への有効な更新本文を返す。
func newUpdateShiftBody() map[string]any {
	body := map[string]any{}
	body["operation_key"] = testShiftOperationKey
	body["expected_version"] = 3
	body["shift_type"] = "off"
	body["starts_at"] = nil
	body["ends_at"] = nil
	body["breaks"] = []any{}
	return body
}

func bindShiftRequestBody(t *testing.T, body map[string]any, destination any) error {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return bindStaffJSON(newStaffBindingContext(t, string(raw)), destination)
}

func TestCreateShiftRequest_OvernightToServiceInput(t *testing.T) {
	// Arrange
	var req createShiftRequest
	require.NoError(t, bindShiftRequestBody(t, newCreateShiftBody(), &req))

	// Act
	input, err := req.toServiceInput()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, uint64(10), input.StaffID)
	assert.Equal(t, "2026-10-21", input.Date.Format(time.DateOnly))
	assert.Equal(t, "full", input.ShiftType)
	require.NotNil(t, input.StartsAt)
	require.NotNil(t, input.EndsAt)
	assert.True(t, input.StartsAt.Equal(time.Date(2026, 10, 21, 22, 0, 0, 0, config.JST)))
	assert.True(t, input.EndsAt.Equal(time.Date(2026, 10, 22, 7, 0, 0, 0, config.JST)), "日跨ぎ勤務は翌日の明示日時のまま保持する")
	require.Len(t, input.Breaks, 1)
	assert.True(t, input.Breaks[0].StartsAt.Equal(time.Date(2026, 10, 22, 1, 0, 0, 0, config.JST)))
	assert.True(t, input.Breaks[0].EndsAt.Equal(time.Date(2026, 10, 22, 2, 0, 0, 0, config.JST)))
	assert.Equal(t, "夜勤", input.Notes)
}

func TestCreateShiftRequest_OffRequiresExplicitNullTimes(t *testing.T) {
	t.Run("explicit null times and an empty breaks array are accepted", func(t *testing.T) {
		body := newCreateShiftBody()
		body["shift_type"] = "off"
		body["starts_at"] = nil
		body["ends_at"] = nil
		body["breaks"] = []any{}
		body["notes"] = ""
		var req createShiftRequest
		require.NoError(t, bindShiftRequestBody(t, body, &req))

		input, err := req.toServiceInput()

		require.NoError(t, err)
		assert.Nil(t, input.StartsAt)
		assert.Nil(t, input.EndsAt)
		assert.Empty(t, input.Breaks)
		assert.Empty(t, input.Notes)
	})

	t.Run("omitted times are rejected instead of being treated as null", func(t *testing.T) {
		body := newCreateShiftBody()
		body["shift_type"] = "off"
		delete(body, "starts_at")
		var req createShiftRequest

		err := bindShiftRequestBody(t, body, &req)

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

func TestCreateShiftRequest_RejectsMalformedBodies(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(body map[string]any)
	}{
		{name: "legacy start_time key", mutate: func(body map[string]any) { body["start_time"] = "09:00" }},
		{name: "legacy end_time key", mutate: func(body map[string]any) { body["end_time"] = "18:00" }},
		{name: "legacy break keys", mutate: func(body map[string]any) {
			body["breaks"] = []any{map[string]any{"break_start": "12:00", "break_end": "13:00"}}
		}},
		{name: "unknown key", mutate: func(body map[string]any) { body["clinic_id"] = "1" }},
		{name: "missing operation_key", mutate: func(body map[string]any) { delete(body, "operation_key") }},
		{name: "missing breaks", mutate: func(body map[string]any) { delete(body, "breaks") }},
		{name: "missing notes", mutate: func(body map[string]any) { delete(body, "notes") }},
		{name: "null breaks", mutate: func(body map[string]any) { body["breaks"] = nil }},
		{name: "null notes", mutate: func(body map[string]any) { body["notes"] = nil }},
		{name: "starts_at without an offset", mutate: func(body map[string]any) { body["starts_at"] = "2026-10-21T22:00:00" }},
		{name: "ends_at that is not a date-time", mutate: func(body map[string]any) { body["ends_at"] = "07:00" }},
		{name: "staff_id as a JSON number", mutate: func(body map[string]any) { body["staff_id"] = 10 }},
		{name: "break without ends_at", mutate: func(body map[string]any) {
			body["breaks"] = []any{map[string]any{"starts_at": "2026-10-22T01:00:00+09:00"}}
		}},
		{name: "unknown shift_type", mutate: func(body map[string]any) { body["shift_type"] = "night" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := newCreateShiftBody()
			tt.mutate(body)
			var req createShiftRequest

			err := bindShiftRequestBody(t, body, &req)

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "unexpected error: %v", err)
		})
	}
}

func TestCreateShiftRequest_ToServiceInputRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(body map[string]any)
	}{
		{name: "staff_id with a leading zero", mutate: func(body map[string]any) { body["staff_id"] = "010" }},
		{name: "staff_id zero", mutate: func(body map[string]any) { body["staff_id"] = "0" }},
		{name: "staff_id with a sign", mutate: func(body map[string]any) { body["staff_id"] = "+10" }},
		{name: "staff_id above the bigint range", mutate: func(body map[string]any) { body["staff_id"] = "9223372036854775808" }},
		{name: "date not in YYYY-MM-DD", mutate: func(body map[string]any) { body["date"] = "2026/10/21" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := newCreateShiftBody()
			tt.mutate(body)
			var req createShiftRequest
			require.NoError(t, bindShiftRequestBody(t, body, &req))

			input, err := req.toServiceInput()

			require.Error(t, err)
			assert.Nil(t, input)
		})
	}
}

func TestParseShiftOperationKey(t *testing.T) {
	// 大文字の UUID は小文字の正規表記へ揃える。
	got, err := parseShiftOperationKey("AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA")
	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", got)

	for _, bad := range []string{"", "not-a-uuid", "11111111111141118111111111111111", "{11111111-1111-4111-8111-11111111111}"} {
		_, err := parseShiftOperationKey(bad)
		require.Error(t, err, bad)
		assert.True(t, apperrors.IsInvalidInput(err))
	}
}

func TestUpdateShiftRequest_ToServiceInput(t *testing.T) {
	// Arrange
	var req updateShiftRequest
	require.NoError(t, bindShiftRequestBody(t, newUpdateShiftBody(), &req))

	// Act
	input, err := req.toServiceInput()

	// Assert
	require.NoError(t, err)
	require.NotNil(t, input.ShiftType)
	assert.Equal(t, "off", *input.ShiftType)
	assert.True(t, input.StartsAt.Set, "明示 null は未指定と区別する")
	assert.Nil(t, input.StartsAt.Value)
	assert.True(t, input.EndsAt.Set)
	assert.Nil(t, input.EndsAt.Value)
	assert.Nil(t, input.Notes, "未指定の notes は維持する")
	require.NotNil(t, input.Breaks)
	assert.Empty(t, *input.Breaks)
	require.NotNil(t, input.ExpectedVersion)
	assert.Equal(t, int64(3), *input.ExpectedVersion)
}

func TestUpdateShiftRequest_UnspecifiedFieldsKeepStoredValues(t *testing.T) {
	// Arrange
	body := map[string]any{}
	body["operation_key"] = testShiftOperationKey
	body["expected_version"] = 1
	body["notes"] = ""
	var req updateShiftRequest
	require.NoError(t, bindShiftRequestBody(t, body, &req))

	// Act
	input, err := req.toServiceInput()

	// Assert
	require.NoError(t, err)
	assert.False(t, input.StartsAt.Set)
	assert.False(t, input.EndsAt.Set)
	assert.Nil(t, input.ShiftType)
	assert.Nil(t, input.Breaks)
	require.NotNil(t, input.Notes)
	assert.Empty(t, *input.Notes, "空文字は notes の解除")
}

func TestUpdateShiftRequest_RejectsMalformedBodies(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(body map[string]any)
	}{
		{name: "legacy start_time key", mutate: func(body map[string]any) { body["start_time"] = "09:00" }},
		{name: "legacy end_time key", mutate: func(body map[string]any) { body["end_time"] = "18:00" }},
		{name: "staff_id cannot be reassigned", mutate: func(body map[string]any) { body["staff_id"] = "11" }},
		{name: "date cannot be reassigned", mutate: func(body map[string]any) { body["date"] = "2026-10-22" }},
		{name: "missing operation_key", mutate: func(body map[string]any) { delete(body, "operation_key") }},
		{name: "missing expected_version", mutate: func(body map[string]any) { delete(body, "expected_version") }},
		{name: "expected_version zero", mutate: func(body map[string]any) { body["expected_version"] = 0 }},
		{name: "null shift_type", mutate: func(body map[string]any) { body["shift_type"] = nil }},
		{name: "null breaks", mutate: func(body map[string]any) { body["breaks"] = nil }},
		{name: "null notes", mutate: func(body map[string]any) { body["notes"] = nil }},
		{name: "starts_at without an offset", mutate: func(body map[string]any) { body["starts_at"] = "2026-10-21T09:00:00" }},
		{name: "only operation_key and expected_version", mutate: func(body map[string]any) {
			delete(body, "shift_type")
			delete(body, "starts_at")
			delete(body, "ends_at")
			delete(body, "breaks")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := newUpdateShiftBody()
			tt.mutate(body)
			var req updateShiftRequest

			err := bindShiftRequestBody(t, body, &req)

			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err), "unexpected error: %v", err)
		})
	}
}
