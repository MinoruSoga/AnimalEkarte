package reservation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLineReservationSettingRepository_FindAll(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	makeLineReservationSetting(t, db, clinicA)
	makeLineReservationSetting(t, db, clinicB)

	got, err := repo.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	clinicIDs := []uint64{got[0].ClinicID, got[1].ClinicID}
	assert.Contains(t, clinicIDs, clinicA)
	assert.Contains(t, clinicIDs, clinicB)
}

func TestLineReservationSettingRepository_FindWebhookRouteByLineBotUserID(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	a := makeLineReservationSetting(t, db, clinicA)
	a.LineBotUserID = "bot-A"
	require.NoError(t, db.WithContext(ctx).Model(a).Updates(map[string]any{
		"line_bot_user_id":    "bot-A",
		"line_channel_secret": "legacy-present-placeholder",
	}).Error)

	b := makeLineReservationSetting(t, db, clinicB)
	b.LineBotUserID = "bot-B"
	require.NoError(t, db.WithContext(ctx).Model(b).Update("line_bot_user_id", "bot-B").Error)

	// Unprovisioned clinic (empty bot user id) must never match empty lookup.
	makeLineReservationSetting(t, db, 3)

	t.Run("returns only matching clinic identity and legacy credential presence", func(t *testing.T) {
		gotClinicID, legacyCredentialPresent, err := repo.FindWebhookRouteByLineBotUserID(ctx, "bot-A")
		require.NoError(t, err)
		assert.Equal(t, clinicA, gotClinicID)
		assert.True(t, legacyCredentialPresent)
	})

	t.Run("reports absence without returning any credential payload", func(t *testing.T) {
		gotClinicID, legacyCredentialPresent, err := repo.FindWebhookRouteByLineBotUserID(ctx, "bot-B")
		require.NoError(t, err)
		assert.Equal(t, clinicB, gotClinicID)
		assert.False(t, legacyCredentialPresent)
	})

	t.Run("unknown bot id is not found", func(t *testing.T) {
		_, _, err := repo.FindWebhookRouteByLineBotUserID(ctx, "bot-unknown")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("empty bot id is not found without matching unprovisioned rows", func(t *testing.T) {
		_, _, err := repo.FindWebhookRouteByLineBotUserID(ctx, "")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLineReservationSettingRepository_Save(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)

	t.Run("creates a new setting row", func(t *testing.T) {
		setting := &model.LineReservationSetting{
			ClinicID:         clinicA,
			Status:           "stopped",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
		}
		require.NoError(t, repo.Save(ctx, clinicA, setting))

		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, "stopped", got.Status)

		var count int64
		require.NoError(t, db.Model(&model.LineReservationSetting{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("persists explicit show_no_staff_option false on create upsert", func(t *testing.T) {
		const clinicC = uint64(3)
		setting := &model.LineReservationSetting{
			ClinicID:          clinicC,
			Status:            "stopped",
			ClosedWeekdays:    []byte(`[]`),
			ClosedDates:       []byte(`[]`),
			BusinessHours:     []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:        []byte(`[]`),
			AdditionalFields:  []byte(`{}`),
			ShowNoStaffOption: false,
		}
		require.NoError(t, repo.Save(ctx, clinicC, setting))
		assert.False(t, setting.ShowNoStaffOption)

		got, err := repo.FindByClinicID(ctx, clinicC)
		require.NoError(t, err)
		assert.False(t, got.ShowNoStaffOption)

		var raw bool
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("show_no_staff_option").
			Where("clinic_id = ?", clinicC).
			Scan(&raw).Error)
		assert.False(t, raw, "raw show_no_staff_option must be false")
	})

	t.Run("updates the existing row without duplication", func(t *testing.T) {
		updated := &model.LineReservationSetting{
			ClinicID:         clinicA,
			Status:           "active",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"1000","end":"1800"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
			PhoneNumber:      "03-1234-5678",
		}
		require.NoError(t, repo.Save(ctx, clinicA, updated))

		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, "active", got.Status)
		assert.Equal(t, "03-1234-5678", got.PhoneNumber)

		var count int64
		require.NoError(t, db.Model(&model.LineReservationSetting{}).Where("clinic_id = ?", clinicA).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Save does not wipe provisioned line_bot_user_id (SEC-CS-F05-R1)", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Where("clinic_id = ?", clinicA).
			Update("line_bot_user_id", "bot-provisioned").Error)

		wiped := &model.LineReservationSetting{
			ClinicID:         clinicA,
			Status:           "active",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"1000","end":"1800"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
			// LineBotUserID intentionally empty — API Save path never sends it.
		}
		require.NoError(t, repo.Save(ctx, clinicA, wiped))

		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, "bot-provisioned", got.LineBotUserID)

		var rawBot string
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("line_bot_user_id").
			Where("clinic_id = ?", clinicA).
			Scan(&rawBot).Error)
		assert.Equal(t, "bot-provisioned", rawBot, "raw line_bot_user_id must survive Save without bot id")
	})

	// BUG-030: GORM injects DefaultValueInterface for zero default-tagged fields on
	// Create; former OnConflict excluded.* then rewrote explicit 0 with DB defaults.
	t.Run("persists explicit zeros on first insert create path", func(t *testing.T) {
		const clinicNew = uint64(29)
		setting := &model.LineReservationSetting{
			ClinicID:             clinicNew,
			Status:               "stopped",
			ClosedWeekdays:       []byte(`[]`),
			ClosedDates:          []byte(`[]`),
			BusinessHours:        []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:           []byte(`[]`),
			AdditionalFields:     []byte(`{}`),
			BookingWindowMinDays: 0,
			BookingWindowMaxDays: 0,
			CalendarMonths:       0,
			ShowNoStaffOption:    false,
		}
		require.NoError(t, repo.Save(ctx, clinicNew, setting))
		assert.Equal(t, 0, setting.BookingWindowMinDays)
		assert.Equal(t, 0, setting.BookingWindowMaxDays)
		assert.Equal(t, 0, setting.CalendarMonths)
		assert.False(t, setting.ShowNoStaffOption)

		got, err := repo.FindByClinicID(ctx, clinicNew)
		require.NoError(t, err)
		assert.Equal(t, 0, got.BookingWindowMinDays)
		assert.Equal(t, 0, got.BookingWindowMaxDays)
		assert.Equal(t, 0, got.CalendarMonths)
		assert.False(t, got.ShowNoStaffOption)

		var rawMin, rawMax, rawCal int
		var rawShow bool
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("booking_window_min_days").
			Where("clinic_id = ?", clinicNew).
			Scan(&rawMin).Error)
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("booking_window_max_days").
			Where("clinic_id = ?", clinicNew).
			Scan(&rawMax).Error)
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("calendar_months").
			Where("clinic_id = ?", clinicNew).
			Scan(&rawCal).Error)
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("show_no_staff_option").
			Where("clinic_id = ?", clinicNew).
			Scan(&rawShow).Error)
		assert.Equal(t, 0, rawMin)
		assert.Equal(t, 0, rawMax)
		assert.Equal(t, 0, rawCal)
		assert.False(t, rawShow)
	})

	t.Run("persists explicit zero for booking window and calendar columns", func(t *testing.T) {
		const clinicZ = uint64(30)

		seed := &model.LineReservationSetting{
			ClinicID:             clinicZ,
			Status:               "stopped",
			ClosedWeekdays:       []byte(`[]`),
			ClosedDates:          []byte(`[]`),
			BusinessHours:        []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:           []byte(`[]`),
			AdditionalFields:     []byte(`{}`),
			BookingWindowMinDays: 2,
			BookingWindowMaxDays: 30,
			CalendarMonths:       2,
		}
		require.NoError(t, repo.Save(ctx, clinicZ, seed))

		// Peer clinic must stay untouched while clinicZ is zeroed.
		const clinicPeer = uint64(31)
		peer := makeLineReservationSetting(t, db, clinicPeer)
		require.NoError(t, db.WithContext(ctx).
			Model(peer).
			Updates(map[string]any{
				"booking_window_min_days": 5,
				"booking_window_max_days": 40,
				"calendar_months":         3,
			}).Error)

		zeroCols := []struct {
			name   string
			apply  func(*model.LineReservationSetting)
			column string
		}{
			{
				name: "booking_window_min_days",
				apply: func(s *model.LineReservationSetting) {
					s.BookingWindowMinDays = 0
				},
				column: "booking_window_min_days",
			},
			{
				name: "booking_window_max_days",
				apply: func(s *model.LineReservationSetting) {
					s.BookingWindowMaxDays = 0
				},
				column: "booking_window_max_days",
			},
			{
				name: "calendar_months",
				apply: func(s *model.LineReservationSetting) {
					s.CalendarMonths = 0
				},
				column: "calendar_months",
			},
		}

		for _, tc := range zeroCols {
			t.Run(tc.name+"=0 after non-zero seed", func(t *testing.T) {
				setting := &model.LineReservationSetting{
					ClinicID:             clinicZ,
					Status:               "stopped",
					ClosedWeekdays:       []byte(`[]`),
					ClosedDates:          []byte(`[]`),
					BusinessHours:        []byte(`{"start":"0900","end":"1900"}`),
					BreakHours:           []byte(`[]`),
					AdditionalFields:     []byte(`{}`),
					BookingWindowMinDays: 2,
					BookingWindowMaxDays: 30,
					CalendarMonths:       2,
				}
				tc.apply(setting)
				require.NoError(t, repo.Save(ctx, clinicZ, setting))

				got, err := repo.FindByClinicID(ctx, clinicZ)
				require.NoError(t, err)
				switch tc.column {
				case "booking_window_min_days":
					assert.Equal(t, 0, got.BookingWindowMinDays)
				case "booking_window_max_days":
					assert.Equal(t, 0, got.BookingWindowMaxDays)
				case "calendar_months":
					assert.Equal(t, 0, got.CalendarMonths)
				}

				var raw int
				require.NoError(t, db.WithContext(ctx).
					Model(&model.LineReservationSetting{}).
					Select(tc.column).
					Where("clinic_id = ?", clinicZ).
					Scan(&raw).Error)
				assert.Equal(t, 0, raw, "raw %s must be 0", tc.column)
			})
		}

		// 2 → 0 → 2 round-trip on min days (BUG-030 primary column).
		t.Run("booking_window_min_days 2→0→2 round-trip", func(t *testing.T) {
			toZero := &model.LineReservationSetting{
				ClinicID:             clinicZ,
				Status:               "stopped",
				ClosedWeekdays:       []byte(`[]`),
				ClosedDates:          []byte(`[]`),
				BusinessHours:        []byte(`{"start":"0900","end":"1900"}`),
				BreakHours:           []byte(`[]`),
				AdditionalFields:     []byte(`{}`),
				BookingWindowMinDays: 0,
				BookingWindowMaxDays: 30,
				CalendarMonths:       2,
			}
			require.NoError(t, repo.Save(ctx, clinicZ, toZero))
			var rawZero int
			require.NoError(t, db.WithContext(ctx).
				Model(&model.LineReservationSetting{}).
				Select("booking_window_min_days").
				Where("clinic_id = ?", clinicZ).
				Scan(&rawZero).Error)
			assert.Equal(t, 0, rawZero)

			toTwo := &model.LineReservationSetting{
				ClinicID:             clinicZ,
				Status:               "stopped",
				ClosedWeekdays:       []byte(`[]`),
				ClosedDates:          []byte(`[]`),
				BusinessHours:        []byte(`{"start":"0900","end":"1900"}`),
				BreakHours:           []byte(`[]`),
				AdditionalFields:     []byte(`{}`),
				BookingWindowMinDays: 2,
				BookingWindowMaxDays: 30,
				CalendarMonths:       2,
			}
			require.NoError(t, repo.Save(ctx, clinicZ, toTwo))
			var rawTwo int
			require.NoError(t, db.WithContext(ctx).
				Model(&model.LineReservationSetting{}).
				Select("booking_window_min_days").
				Where("clinic_id = ?", clinicZ).
				Scan(&rawTwo).Error)
			assert.Equal(t, 2, rawTwo)
		})

		// Peer clinic isolation after zero writes on clinicZ.
		gotPeer, err := repo.FindByClinicID(ctx, clinicPeer)
		require.NoError(t, err)
		assert.Equal(t, 5, gotPeer.BookingWindowMinDays)
		assert.Equal(t, 40, gotPeer.BookingWindowMaxDays)
		assert.Equal(t, 3, gotPeer.CalendarMonths)
	})

	// R-05 Phase B: OnConflict update must not overwrite legacy line_channel_secret column
	// (column remains for presence SELECT until inventory-zero DROP packet).
	t.Run("Save does not wipe existing line_channel_secret on update (R-05 Phase B)", func(t *testing.T) {
		const placeholder = "legacy-present-placeholder"
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Where("clinic_id = ?", clinicA).
			Update("line_channel_secret", placeholder).Error)

		withoutSecret := &model.LineReservationSetting{
			ClinicID:         clinicA,
			Status:           "active",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"1000","end":"1800"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
			// LineChannelSecret intentionally zero — write path no longer sets it.
		}
		require.NoError(t, repo.Save(ctx, clinicA, withoutSecret))

		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, placeholder, got.LineChannelSecret)
	})

	t.Run("updatable columns exclude credential and routing identity columns", func(t *testing.T) {
		// Frozen set: DoUpdates must stay identical to this list (BUG-030 safety).
		// line_channel_secret / line_bot_user_id / id must never appear.
		want := []string{
			"status",
			"header_text",
			"reservation_notice",
			"cancel_notice",
			"privacy_policy",
			"closed_weekdays",
			"closed_dates",
			"national_holiday_closed",
			"business_hours",
			"business_hours_by_weekday",
			"break_hours",
			"daily_limit",
			"monthly_limit",
			"booking_window_max_days",
			"booking_window_min_days",
			"calendar_months",
			"phone_number",
			"notification_email",
			"request_example",
			"time_slot_mode",
			"time_slot_interval_minutes",
			"no_staff_mode",
			"show_no_staff_option",
			"additional_fields",
			"line_channel_id",
			"liff_id",
			"line_access_token",
			"updated_at",
		}
		got := lineReservationSettingUpdatableColumns()
		assert.Equal(t, want, got)
		for _, col := range got {
			assert.NotEqual(t, "line_channel_secret", col)
			assert.NotEqual(t, "line_bot_user_id", col)
			assert.NotEqual(t, "id", col)
		}
	})

	t.Run("does not affect another clinic", func(t *testing.T) {
		const clinicB = uint64(2)
		makeLineReservationSetting(t, db, clinicB)

		other := &model.LineReservationSetting{
			ClinicID:         clinicA,
			Status:           "stopped",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
		}
		require.NoError(t, repo.Save(ctx, clinicA, other))

		gotB, err := repo.FindByClinicID(ctx, clinicB)
		require.NoError(t, err)
		assert.Equal(t, clinicB, gotB.ClinicID)
	})
}
