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

func TestLineReservationSettingRepository_FindByLineBotUserID(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	a := makeLineReservationSetting(t, db, clinicA)
	a.LineBotUserID = "bot-A"
	require.NoError(t, db.WithContext(ctx).Model(a).Update("line_bot_user_id", "bot-A").Error)

	b := makeLineReservationSetting(t, db, clinicB)
	b.LineBotUserID = "bot-B"
	require.NoError(t, db.WithContext(ctx).Model(b).Update("line_bot_user_id", "bot-B").Error)

	// Unprovisioned clinic (empty bot user id) must never match empty lookup.
	makeLineReservationSetting(t, db, 3)

	t.Run("returns the matching clinic", func(t *testing.T) {
		got, err := repo.FindByLineBotUserID(ctx, "bot-A")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicA, got.ClinicID)
		assert.Equal(t, "bot-A", got.LineBotUserID)
	})

	t.Run("unknown bot id is not found", func(t *testing.T) {
		_, err := repo.FindByLineBotUserID(ctx, "bot-unknown")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("empty bot id is not found without matching unprovisioned rows", func(t *testing.T) {
		_, err := repo.FindByLineBotUserID(ctx, "")
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
