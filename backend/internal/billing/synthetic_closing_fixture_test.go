package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func testdbSetupSyntheticClosing(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Company{}, &model.Clinic{},
		&model.Account{}, &model.Staff{}, &model.StaffClinicAssignment{},
		&model.Owner{}, &model.AnimalSpecies{}, &model.Pet{},
		&model.Billing{}, &model.BillingItem{}, &model.Payment{}, &model.PaymentSplit{},
		&model.PaymentMethodMaster{},
	))
	testdb.EnsureClinicSettingsTable(t, db)
	return db
}

func TestCreateSyntheticClosingFixture_RejectsUnsafeRequest(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)

	t.Run("staging env", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "staging", DBHost: "db", TargetDate: day, PasswordHash: "x",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "APP_ENV")
	})

	t.Run("existing billing ids", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "x", ExistingBillingIDs: []uint64{3},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "existing billing")
	})

	t.Run("empty password hash", func(t *testing.T) {
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: day,
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "password hash")
	})

	t.Run("weekend", func(t *testing.T) {
		saturday := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		_, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
			AppEnv: "development", DBHost: "db", TargetDate: saturday, PasswordHash: "x",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "weekday")
	})
}

func TestCreateSyntheticClosingFixture_CreatesFiveNewCompletedBillings(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)

	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "test-hash-not-for-login",
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NoError(t, RejectReservedClinicID(got.ClinicID))
	require.Len(t, got.BillingIDs, 5)
	require.Len(t, got.CompletedAt, 5)
	assert.Equal(t, SyntheticClosingLoginEmail(got.ClinicID), got.LoginEmail)
	assert.Equal(t, SyntheticClosingCleanupToken(got.ClinicID), got.CleanupToken)

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

	var items []model.BillingItem
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Find(&items).Error)
	require.Len(t, items, 5)
	var splits []model.PaymentSplit
	require.NoError(t, db.WithContext(ctx).Where("clinic_id = ?", got.ClinicID).Find(&splits).Error)
	require.Len(t, splits, 5)
	var account model.Account
	require.NoError(t, db.WithContext(ctx).Where("email = ?", got.LoginEmail).First(&account).Error)
	assert.True(t, account.IsSystemAdmin)

	require.NoError(t, DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, got.CleanupToken))
	var remaining int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Billing{}).Where("clinic_id = ?", got.ClinicID).Count(&remaining).Error)
	assert.Zero(t, remaining)
	require.Error(t, db.WithContext(ctx).First(&model.Clinic{}, got.ClinicID).Error)
}

func TestDeleteSyntheticClosingFixture_RejectsWrongToken(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	got, err := CreateSyntheticClosingFixture(ctx, db, SyntheticClosingRequest{
		AppEnv: "development", DBHost: "db", TargetDate: time.Date(2026, 9, 7, 0, 0, 0, 0, jst), PasswordHash: "x",
	})
	require.NoError(t, err)
	err = DeleteSyntheticClosingFixture(ctx, db, "development", "db", got.ClinicID, "deadbeef")
	require.Error(t, err)
	assert.ErrorContains(t, err, "cleanup token")
}
