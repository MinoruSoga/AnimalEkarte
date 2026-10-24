package reservation

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToLineReservationSettingResponse(t *testing.T) {
	t.Run("converts full setting with all fields populated", func(t *testing.T) {
		dailyLimit := 5
		monthlyLimit := 100
		created := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
		updated := time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC)
		m := &model.LineReservationSetting{
			ID:                      1,
			ClinicID:                2,
			Status:                  "open",
			HeaderText:              "ご予約はこちら",
			ReservationNotice:       "当日キャンセルはご遠慮ください",
			CancelNotice:            "前日までにご連絡ください",
			PrivacyPolicy:           "個人情報保護方針",
			ClosedWeekdays:          []byte(`[0]`),
			ClosedDates:             []byte(`["2026-01-01"]`),
			NationalHolidayClosed:   true,
			BusinessHours:           []byte(`{"start":"0900","end":"1900"}`),
			BusinessHoursByWeekday:  nullableJSONPtr([]byte(`{"mon":{"start":"0900","end":"1800"}}`)),
			BreakHours:              []byte(`[{"start":"1200","end":"1300"}]`),
			DailyLimit:              &dailyLimit,
			MonthlyLimit:            &monthlyLimit,
			BookingWindowMaxDays:    30,
			BookingWindowMinDays:    2,
			CalendarMonths:          2,
			PhoneNumber:             "03-1234-5678",
			NotificationEmail:       "notify@example.com",
			RequestExample:          "例: 爪切り希望",
			TimeSlotMode:            "minimize_gaps",
			TimeSlotIntervalMinutes: 15,
			NoStaffMode:             "first_available",
			ShowNoStaffOption:       true,
			AdditionalFields:        []byte(`{}`),
			CreatedAt:               created,
			UpdatedAt:               updated,
		}

		resp := toLineReservationSettingResponse(m)

		assert.Equal(t, uint64(1), resp.ID)
		assert.Equal(t, uint64(2), resp.ClinicID)
		assert.Equal(t, "open", resp.Status)
		assert.Equal(t, "ご予約はこちら", resp.HeaderText)
		assert.True(t, resp.NationalHolidayClosed)
		assert.JSONEq(t, `[0]`, string(resp.ClosedWeekdays))
		assert.JSONEq(t, `["2026-01-01"]`, string(resp.ClosedDates))
		assert.JSONEq(t, `{"start":"0900","end":"1900"}`, string(resp.BusinessHours))
		assert.JSONEq(t, `{"mon":{"start":"0900","end":"1800"}}`, string(resp.BusinessHoursByWeekday))
		require := assert.New(t)
		require.NotNil(resp.DailyLimit)
		require.Equal(5, *resp.DailyLimit)
		require.NotNil(resp.MonthlyLimit)
		require.Equal(100, *resp.MonthlyLimit)
		assert.Equal(t, 30, resp.BookingWindowMaxDays)
		assert.Equal(t, 2, resp.BookingWindowMinDays)

		// EMR-504: LINE 資格情報（channel ID / LIFF ID / token / secret）は応答に存在しない。
		raw, err := json.Marshal(resp)
		assert.NoError(t, err)
		for _, key := range []string{"line_channel_id", "line_channel_secret", "liff_id", "line_access_token"} {
			assert.NotContains(t, string(raw), `"`+key+`"`)
		}
	})

	t.Run("converts setting with nil optional limits and empty byte fields", func(t *testing.T) {
		m := &model.LineReservationSetting{
			ID:                     10,
			ClinicID:               20,
			Status:                 "stopped",
			ClosedWeekdays:         []byte(`[]`),
			ClosedDates:            []byte(`[]`),
			BusinessHours:          []byte(`{}`),
			BusinessHoursByWeekday: nil,
			BreakHours:             []byte(`[]`),
			DailyLimit:             nil,
			MonthlyLimit:           nil,
			AdditionalFields:       []byte(`{}`),
		}

		resp := toLineReservationSettingResponse(m)

		assert.Nil(t, resp.DailyLimit)
		assert.Nil(t, resp.MonthlyLimit)
		assert.Equal(t, "stopped", resp.Status)
		assert.Empty(t, resp.BusinessHoursByWeekday)
	})

	// EMR-504 contract A: NULL (override released) is omitted from the wire; a stored JSON value is kept.
	t.Run("business_hours_by_weekday NULL is omitted and a stored value is returned", func(t *testing.T) {
		base := model.LineReservationSetting{
			ClosedWeekdays: []byte(`[]`), ClosedDates: []byte(`[]`), BusinessHours: []byte(`{}`),
			BreakHours: []byte(`[]`), AdditionalFields: []byte(`{}`),
		}
		withNull := base
		raw, err := json.Marshal(toLineReservationSettingResponse(&withNull))
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "business_hours_by_weekday")

		withValue := base
		withValue.BusinessHoursByWeekday = nullableJSONPtr([]byte(`{"6":{"start":"0900","end":"1300"}}`))
		raw, err = json.Marshal(toLineReservationSettingResponse(&withValue))
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"business_hours_by_weekday":{"6":{"start":"0900","end":"1300"}}`)
	})
}
