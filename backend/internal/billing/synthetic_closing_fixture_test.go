package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestCreateSyntheticClosingFixture_RejectsUnsafeRequest(t *testing.T) {
	db := testdb.SetupTestDB(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)

	t.Run("staging env", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "staging", DBHost: "db", TargetDate: day,
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "APP_ENV")
	})

	t.Run("existing billing ids", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day, ExistingBillingIDs: []uint64{3},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "existing billing")
	})
}

func TestCreateSyntheticClosingFixture_CreatesFiveNewCompletedBillings(t *testing.T) {
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{}, &model.Billing{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)

	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: day,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, RejectReservedClinicID(got.ClinicID))
	require.Len(t, got.BillingIDs, 5)
	require.Len(t, got.CompletedAt, 5)

	wantHours := [][2]int{{10, 0}, {13, 30}, {14, 0}, {20, 0}, {2, 0}}
	var persisted []model.Billing
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Order("completed_at ASC").Find(&persisted).Error)
	require.Len(t, persisted, 5)

	for i, b := range persisted {
		require.NotNil(t, b.CompletedAt)
		at := b.CompletedAt.In(jst)
		assert.Equal(t, wantHours[i][0], at.Hour(), "billing %d hour", i)
		assert.Equal(t, wantHours[i][1], at.Minute(), "billing %d minute", i)
		assert.Equal(t, model.BillingStatusCompleted, b.Status)
		assert.Equal(t, "s09-synthetic", b.Memo)
		assert.NotZero(t, b.ID)
	}
	assert.Equal(t, 8, persisted[4].CompletedAt.In(jst).Day(), "overnight EMG is next calendar day")

	var settings model.ClinicSettings
	require.NoError(t, db.WithContext(ctx).First(&settings, "clinic_id = ?", got.ClinicID).Error)
	assert.Equal(t, "09:00:00", settings.ClosingAmStart)
	assert.Equal(t, "13:30:00", settings.ClosingAmPmBoundary)
	assert.Equal(t, "19:00:00", settings.ClosingWeekdayEnd)
}
