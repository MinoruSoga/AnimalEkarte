package reservation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
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

// EMR-504 Q7: line_bot_user_id was dropped and the replacement routing storage is undecided, so the
// repository fails closed with the fixed 422 instead of reading dropped columns or guessing a clinic.
func TestLineReservationSettingRepository_FindWebhookRouteByLineBotUserID_FailsClosed(t *testing.T) {
	repo := NewLineReservationSettingRepository(nil)

	for _, destination := range []string{"bot-A", "bot-unknown", ""} {
		gotClinicID, err := repo.FindWebhookRouteByLineBotUserID(context.Background(), destination)

		require.Error(t, err)
		assert.Zero(t, gotClinicID)
		assert.True(t, apperrors.HasCode(err, apperrors.CodeLineWebhookRoutingPending))
		assert.ErrorIs(t, err, apperrors.ErrUnprocessable)
	}
}

// FindIntegrationValue reads only the allow-listed non-secret key (liff_id) from clinic_integrations,
// scoped to the clinic.
func TestLineReservationSettingRepository_FindIntegrationValue(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClinicIntegration{}))
	db.Exec("TRUNCATE TABLE clinic_integrations CASCADE")
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLiffID, KeyValue: "1111111111-aaaa",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicB, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLiffID, KeyValue: "9999999999-liffB",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLineChannelSecret, KeyValue: "ciphertext",
	}).Error)

	t.Run("returns the clinic's own liff_id", func(t *testing.T) {
		got, err := repo.FindIntegrationValue(ctx, clinicA, model.IntegrationKeyLiffID)
		require.NoError(t, err)
		assert.Equal(t, "1111111111-aaaa", got)
	})

	t.Run("a clinic without the row gets an empty value, not an error", func(t *testing.T) {
		got, err := repo.FindIntegrationValue(ctx, 3, model.IntegrationKeyLiffID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("returns another clinic's own value", func(t *testing.T) {
		got, err := repo.FindIntegrationValue(ctx, clinicB, model.IntegrationKeyLiffID)
		require.NoError(t, err)
		assert.Equal(t, "9999999999-liffB", got)
	})

	t.Run("secret and other keys are not readable", func(t *testing.T) {
		for _, key := range []string{model.IntegrationKeyLineChannelSecret, model.IntegrationKeyLineChannelAccessToken, model.IntegrationKeyLstepAPIKey} {
			got, err := repo.FindIntegrationValue(ctx, clinicA, key)
			require.Error(t, err, key)
			assert.Empty(t, got)
		}
	})
}

// FindLineChannelAccessTokenCiphertext is the purpose-named reader behind reservation notifications:
// it returns the stored (still encrypted) channel access token of the clinic only, and does not widen
// FindIntegrationValue's allow-list.
func TestLineReservationSettingRepository_FindLineChannelAccessTokenCiphertext(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ClinicIntegration{}))
	db.Exec("TRUNCATE TABLE clinic_integrations CASCADE")
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLineChannelAccessToken, KeyValue: "ciphertext-aaaa",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicB, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLineChannelAccessToken, KeyValue: "ciphertext-bbbb",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&model.ClinicIntegration{
		ClinicID: clinicA, Service: model.IntegrationServiceLstep, KeyName: model.IntegrationKeyLineChannelSecret, KeyValue: "ciphertext-secret",
	}).Error)

	t.Run("returns the clinic's own stored ciphertext untouched", func(t *testing.T) {
		got, err := repo.FindLineChannelAccessTokenCiphertext(ctx, clinicA)
		require.NoError(t, err)
		assert.Equal(t, "ciphertext-aaaa", got)
	})

	t.Run("another clinic gets its own value, never a neighbour's", func(t *testing.T) {
		got, err := repo.FindLineChannelAccessTokenCiphertext(ctx, clinicB)
		require.NoError(t, err)
		assert.Equal(t, "ciphertext-bbbb", got)
	})

	t.Run("a clinic without the row gets an empty value, not an error", func(t *testing.T) {
		got, err := repo.FindLineChannelAccessTokenCiphertext(ctx, 3)
		require.NoError(t, err)
		assert.Empty(t, got)
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

	// BUG-030 follow-up: OnConflict DO NOTHING does not RETURNING on conflict, so
	// update path left setting.ID at 0 and PUT response id became 0 (RSV-03 no re-fetch).
	t.Run("populates setting.ID matching DB row on create path", func(t *testing.T) {
		const clinicCreateID = uint64(40)
		setting := &model.LineReservationSetting{
			ClinicID:         clinicCreateID,
			Status:           "stopped",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
		}
		require.NoError(t, repo.Save(ctx, clinicCreateID, setting))

		var rawID uint64
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("id").
			Where("clinic_id = ?", clinicCreateID).
			Scan(&rawID).Error)
		require.NotZero(t, rawID, "DB row must have a non-zero primary key")
		assert.Equal(t, rawID, setting.ID, "Save must reflect persisted id on create path")
	})

	t.Run("populates setting.ID matching DB row on update path", func(t *testing.T) {
		const clinicUpdateID = uint64(41)
		seed := makeLineReservationSetting(t, db, clinicUpdateID)
		require.NotZero(t, seed.ID)
		require.False(t, seed.CreatedAt.IsZero(), "seed must have created_at")

		updated := &model.LineReservationSetting{
			ClinicID:         clinicUpdateID,
			Status:           "active",
			ClosedWeekdays:   []byte(`[]`),
			ClosedDates:      []byte(`[]`),
			BusinessHours:    []byte(`{"start":"1000","end":"1800"}`),
			BreakHours:       []byte(`[]`),
			AdditionalFields: []byte(`{}`),
			PhoneNumber:      "03-9999-0000",
			// ID/CreatedAt intentionally zero — service builds request struct without them.
		}
		require.NoError(t, repo.Save(ctx, clinicUpdateID, updated))

		var rawID uint64
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("id").
			Where("clinic_id = ?", clinicUpdateID).
			Scan(&rawID).Error)
		require.Equal(t, seed.ID, rawID, "update must not create a second row / change PK")
		assert.Equal(t, rawID, updated.ID, "Save must reflect persisted id on update path")

		// created_at: BUG-030's *setting = intended zeroed timestamps; reflect DB value.
		var rawCreatedAt time.Time
		require.NoError(t, db.WithContext(ctx).
			Model(&model.LineReservationSetting{}).
			Select("created_at").
			Where("clinic_id = ?", clinicUpdateID).
			Scan(&rawCreatedAt).Error)
		require.False(t, rawCreatedAt.IsZero())
		assert.WithinDuration(t, rawCreatedAt, updated.CreatedAt, time.Second,
			"Save must reflect persisted created_at on update path")
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

	t.Run("updatable columns hold no LINE credential or routing columns", func(t *testing.T) {
		// Frozen set: DoUpdates must stay identical to this list (BUG-030 safety).
		// EMR-504: line_channel_id / liff_id / line_access_token / line_channel_secret / line_bot_user_id
		// were dropped from the table, so none may appear (they live in clinic_integrations).
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
			"updated_at",
		}
		got := lineReservationSettingUpdatableColumns()
		assert.Equal(t, want, got)
		for _, col := range got {
			for _, dropped := range []string{"line_channel_id", "line_channel_secret", "liff_id", "line_access_token", "line_bot_user_id", "id"} {
				assert.NotEqual(t, dropped, col)
			}
		}
	})

	// EMR-504 contract A: business_hours_by_weekday is a nullable jsonb. NULL (override released),
	// a valid JSON object and the JSON literal null (normalised to NULL by the service) must round-trip
	// without ever writing an empty (invalid) jsonb.
	t.Run("business_hours_by_weekday distinguishes NULL from a stored JSON object", func(t *testing.T) {
		const clinicWeekday = uint64(41)
		base := func() *model.LineReservationSetting {
			return &model.LineReservationSetting{
				ClinicID:         clinicWeekday,
				Status:           "stopped",
				ClosedWeekdays:   []byte(`[]`),
				ClosedDates:      []byte(`[]`),
				BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
				BreakHours:       []byte(`[]`),
				AdditionalFields: []byte(`{}`),
			}
		}
		isNull := func() bool {
			var n bool
			require.NoError(t, db.WithContext(ctx).
				Model(&model.LineReservationSetting{}).
				Select("business_hours_by_weekday IS NULL").
				Where("clinic_id = ?", clinicWeekday).
				Scan(&n).Error)
			return n
		}

		first := base()
		require.NoError(t, repo.Save(ctx, clinicWeekday, first))
		assert.True(t, isNull(), "a nil pointer must be stored as SQL NULL")
		got, err := repo.FindByClinicID(ctx, clinicWeekday)
		require.NoError(t, err)
		assert.Nil(t, got.BusinessHoursByWeekday)

		withValue := base()
		withValue.BusinessHoursByWeekday = nullableJSONPtr([]byte(`{"6":{"start":"0900","end":"1300"}}`))
		require.NoError(t, repo.Save(ctx, clinicWeekday, withValue))
		assert.False(t, isNull())
		got, err = repo.FindByClinicID(ctx, clinicWeekday)
		require.NoError(t, err)
		require.NotNil(t, got.BusinessHoursByWeekday)
		assert.JSONEq(t, `{"6":{"start":"0900","end":"1300"}}`, string(*got.BusinessHoursByWeekday))

		released := base()
		require.NoError(t, repo.Save(ctx, clinicWeekday, released))
		assert.True(t, isNull(), "saving nil again releases the weekday override (NULL)")
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
