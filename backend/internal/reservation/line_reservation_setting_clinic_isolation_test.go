package reservation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupLineSettingIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LineReservationSetting{}))
	db.Exec("TRUNCATE TABLE line_reservation_settings CASCADE")
	return db
}

func makeLineReservationSetting(t *testing.T, db *gorm.DB, clinicID uint64) *model.LineReservationSetting {
	t.Helper()
	setting := &model.LineReservationSetting{
		ClinicID:         clinicID,
		ClosedWeekdays:   []byte(`[]`),
		ClosedDates:      []byte(`[]`),
		BusinessHours:    []byte(`{"start":"0900","end":"1900"}`),
		BreakHours:       []byte(`[{"start":"1200","end":"1300"}]`),
		AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(setting).Error)
	return setting
}

func TestLineReservationSettingRepository_FindByClinicID_ClinicIsolation(t *testing.T) {
	db := setupLineSettingIsolationTestDB(t)
	repo := NewLineReservationSettingRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	makeLineReservationSetting(t, db, clinicA)

	t.Run("same clinic can read", func(t *testing.T) {
		got, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("other clinic cannot read", func(t *testing.T) {
		got, err := repo.FindByClinicID(ctx, clinicB)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got: %v", err)
	})
}
