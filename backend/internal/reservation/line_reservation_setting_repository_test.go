package reservation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
