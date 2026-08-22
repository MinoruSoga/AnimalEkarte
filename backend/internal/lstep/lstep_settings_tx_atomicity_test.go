package lstep

// lstep_settings_tx_atomicity_test.go — BE-X06-LSTEP-SETTINGS-01 / LSA-06 (X-06)
//
// UpdateSettings は credentials / sync / clinic_settings を同一 ambient tx に載せる。
// 本ファイルは repository が DBOrTx 経由で参加し、後続失敗時に部分成功が残らないことを証明する。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

var errSentinelLstepSettingsTx = errors.New("simulated post-write failure in ambient tx")

func setupLstepSettingsAtomicityDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupLstepSettingsTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepSettings{}, &model.ClinicSettings{}))
	db.Exec("TRUNCATE TABLE lstep_settings CASCADE")
	db.Exec("TRUNCATE TABLE clinic_settings CASCADE")
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_test_lstep_settings_clinic ON lstep_settings (clinic_id)`)
	return db
}

func TestLstepSettingsRepository_Upsert_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupLstepSettingsAtomicityDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()
	tx := persistence.NewTransactor(db)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.Upsert(txCtx, &model.ClinicIntegration{
			ClinicID: 1,
			Service:  model.IntegrationServiceLstep,
			KeyName:  model.IntegrationKeyLineChannelAccessToken,
			KeyValue: "token-should-rollback",
		}); err != nil {
			return err
		}
		return errSentinelLstepSettingsTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelLstepSettingsTx)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).
		Where("clinic_id = ? AND key_name = ?", 1, model.IntegrationKeyLineChannelAccessToken).
		Count(&count).Error)
	assert.EqualValues(t, 0, count, "ambient tx 失敗時、credential Upsert はロールバックされる")
}

// TestLstepSettingsRepository_DeleteByClinicServiceAndKey_RollsBackWhenAmbientTxFails
// proves DeleteByClinicServiceAndKey joins ambient tx via DBOrTx: a seeded
// ClinicIntegration row deleted inside WithTx is restored when the tx fails.
//
// temp-revert RED: DeleteByClinicServiceAndKey DBOrTx → r.db.WithContext(ctx)
// → delete auto-commits and the row is gone after the sentinel rollback.
func TestLstepSettingsRepository_DeleteByClinicServiceAndKey_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupLstepSettingsAtomicityDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()
	tx := persistence.NewTransactor(db)
	const clinicID = uint64(1)

	require.NoError(t, repo.Upsert(ctx, &model.ClinicIntegration{
		ClinicID: clinicID,
		Service:  model.IntegrationServiceLstep,
		KeyName:  model.IntegrationKeyLiffID,
		KeyValue: "liff-should-remain",
	}))

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.DeleteByClinicServiceAndKey(
			txCtx, clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLiffID,
		); err != nil {
			return err
		}
		return errSentinelLstepSettingsTx
	})
	require.Error(t, txErr)
	require.ErrorIs(t, txErr, errSentinelLstepSettingsTx)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).
		Where("clinic_id = ? AND key_name = ?", clinicID, model.IntegrationKeyLiffID).
		Count(&count).Error)
	assert.EqualValues(t, 1, count, "ambient tx 失敗時、DeleteByClinicServiceAndKey はロールバックされる")
}

func TestLstepSyncSettingsRepository_Upsert_DoesNotRequirePostCreateFind(t *testing.T) {
	// G2B-02: Upsert returns the input record without a separate Find that can invert success.
	db := setupLstepSettingsAtomicityDB(t)
	repo := NewLstepSyncSettingsRepository(db)
	ctx := context.Background()

	got, err := repo.Upsert(ctx, &model.LstepSettings{
		ClinicID:      7,
		IsSyncEnabled: true,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, uint64(7), got.ClinicID)
	assert.True(t, got.IsSyncEnabled)

	var stored model.LstepSettings
	require.NoError(t, db.Where("clinic_id = ?", 7).First(&stored).Error)
	assert.True(t, stored.IsSyncEnabled)
}

func TestLstepSettingsService_UpdateSettings_RollsBackCredentialsWhenClinicConfigFails(t *testing.T) {
	db := setupLstepSettingsAtomicityDB(t)
	ctx := context.Background()
	const clinicID = uint64(3)

	settingsRepo := NewLstepSettingsRepository(db)
	syncRepo := NewLstepSyncSettingsRepository(db)
	failingClinic := &failingClinicSettingsRepo{err: errors.New("forced clinic_settings failure")}
	svc := NewLstepSettingsService(settingsRepo, syncRepo, nil, nil, failingClinic, persistence.NewTransactor(db))

	cpm := "v2"
	_, err := svc.UpdateSettings(ctx, clinicID, &UpdateLstepSettingsInput{
		LineChannelAccessToken: "line-token-partial",
		LineChannelSecret:      "line-secret-partial",
		CPMVersion:             &cpm,
	}, nil)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).
		Where("clinic_id = ?", clinicID).
		Count(&count).Error)
	assert.EqualValues(t, 0, count, "clinic_settings 失敗時に credential の部分成功は残らない（LSA-06）")
}

// failingClinicSettingsRepo implements lstepClinicSettingsRepo with forced write errors.
type failingClinicSettingsRepo struct {
	err error
}

func (f *failingClinicSettingsRepo) FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error) {
	return &model.ClinicSettings{ClinicID: clinicID}, nil
}
func (f *failingClinicSettingsRepo) UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error {
	return f.err
}
func (f *failingClinicSettingsRepo) UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error {
	return f.err
}
func (f *failingClinicSettingsRepo) UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error {
	return f.err
}
func (f *failingClinicSettingsRepo) UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error {
	return f.err
}
func (f *failingClinicSettingsRepo) UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error {
	return f.err
}

// TestLstepSettingsRepository_FindCredentialByClinicServiceKey_SeesUncommittedUpsert
// proves FindCredentialByClinicServiceKey joins ambient tx via DBOrTx: an Upsert in
// the same WithTx is visible, then forced rollback leaves the credential absent outside.
//
// temp-revert RED: FindCredential DBOrTx → r.db.WithContext(ctx) → uncommitted Upsert
// is invisible inside ambient and the assert fails before rollback.
func TestLstepSettingsRepository_FindCredentialByClinicServiceKey_SeesUncommittedUpsert(t *testing.T) {
	// ClinicIntegration only — avoid ClinicSettings AutoMigrate time-default issues
	// that setupLstepSettingsAtomicityDB can hit on a cold schema.
	db := setupLstepSettingsTestDB(t)
	repo := NewLstepSettingsRepository(db)
	ctx := context.Background()
	tx := persistence.NewTransactor(db)
	const clinicID = uint64(88)
	forced := errors.New("forced find-credential ambient rollback")

	// Precondition: not found outside.
	_, err := repo.FindCredentialByClinicServiceKey(
		ctx, clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLineChannelAccessToken,
	)
	require.Error(t, err)

	txErr := tx.WithTx(ctx, func(txCtx context.Context) error {
		if persistence.TxFromContext(txCtx) == nil {
			return errors.New("expected ambient tx installed")
		}
		if err := repo.Upsert(txCtx, &model.ClinicIntegration{
			ClinicID: clinicID,
			Service:  model.IntegrationServiceLstep,
			KeyName:  model.IntegrationKeyLineChannelAccessToken,
			KeyValue: "token-uncommitted-visible",
		}); err != nil {
			return err
		}
		got, err := repo.FindCredentialByClinicServiceKey(
			txCtx, clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLineChannelAccessToken,
		)
		if err != nil {
			return err
		}
		if got == nil || got.KeyValue != "token-uncommitted-visible" {
			return errors.New("FindCredentialByClinicServiceKey did not see uncommitted Upsert")
		}
		return forced
	})
	require.ErrorIs(t, txErr, forced)

	// Outside after rollback: still not found.
	_, err = repo.FindCredentialByClinicServiceKey(
		ctx, clinicID, model.IntegrationServiceLstep, model.IntegrationKeyLineChannelAccessToken,
	)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&model.ClinicIntegration{}).
		Where("clinic_id = ? AND key_name = ?", clinicID, model.IntegrationKeyLineChannelAccessToken).
		Count(&count).Error)
	assert.Zero(t, count, "credential Upsert must roll back with ambient tx")
}
