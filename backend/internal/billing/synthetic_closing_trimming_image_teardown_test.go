package billing

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestDeleteSyntheticClosingFixture_TrimmingImageDeletionsAreClinicScopedAndAtomic(t *testing.T) {
	db := testdbSetupSyntheticClosing(t)
	ctx := context.Background()
	jst, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	day := time.Date(2026, 9, 7, 0, 0, 0, 0, jst)
	ddl, err := os.ReadFile(filepath.Join("..", "..", "migrations", "032_trimming_image_deletions.sql"))
	require.NoError(t, err)
	tableDDL, _, found := strings.Cut(string(ddl), "SELECT app_private.apply_rls_policy")
	require.True(t, found)
	tableDDL = strings.Replace(tableDDL, "CREATE TABLE trimming_image_deletions", "CREATE TABLE IF NOT EXISTS trimming_image_deletions", 1)
	policySentinel := errors.New("injected audit policy sentinel")
	cleanupSentinel := errors.New("rollback synthetic trimming image fixture")

	bctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	err = db.WithContext(bctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(bctx, tx)
		require.NoError(t, tx.Exec(tableDDL).Error)
		newFixture := func() *SyntheticClosingResult {
			got, createErr := CreateSyntheticClosingFixture(txCtx, tx, SyntheticClosingRequest{
				AppEnv: "development", DBHost: "db", TargetDate: day, PasswordHash: "test-hash-not-for-login",
				CleanupSecret: testSyntheticClosingSecret,
			})
			require.NoError(t, createErr)
			return got
		}
		target := newFixture()
		control := newFixture()
		require.NoError(t, tx.Exec(
			"INSERT INTO trimming_image_deletions (clinic_id, object_key, actor_staff_id) VALUES (?, ?, ?), (?, ?, ?)",
			target.ClinicID, fmt.Sprintf("trimming-images/%d/teardown.jpg", target.ClinicID), 42,
			control.ClinicID, fmt.Sprintf("trimming-images/%d/teardown.jpg", control.ClinicID), 42,
		).Error)

		policyErr := DeleteSyntheticClosingFixtureWithAuditPolicy(
			txCtx, tx, "development", "db", target.ClinicID, testSyntheticClosingSecret, target.CleanupToken,
			func(context.Context, *gorm.DB, uint64) error { return policySentinel })
		require.ErrorIs(t, policyErr, policySentinel)
		countTable := func(table string, clinicID uint64) int64 {
			var n int64
			require.NoError(t, tx.Table(table).Where("clinic_id = ?", clinicID).Count(&n).Error)
			return n
		}
		countClinic := func(clinicID uint64) int64 {
			var n int64
			require.NoError(t, tx.Model(&model.Clinic{}).Where("id = ?", clinicID).Count(&n).Error)
			return n
		}
		assert.EqualValues(t, 1, countTable("trimming_image_deletions", target.ClinicID))
		assert.EqualValues(t, 1, countTable("trimming_image_deletions", control.ClinicID))
		assert.EqualValues(t, 1, countClinic(target.ClinicID))

		require.NoError(t, DeleteSyntheticClosingFixture(txCtx, tx, "development", "db", target.ClinicID, testSyntheticClosingSecret, target.CleanupToken))
		assert.Zero(t, countTable("trimming_image_deletions", target.ClinicID))
		assert.Zero(t, countClinic(target.ClinicID))
		assert.EqualValues(t, 1, countTable("trimming_image_deletions", control.ClinicID))
		assert.EqualValues(t, 1, countClinic(control.ClinicID))
		return cleanupSentinel
	})
	require.ErrorIs(t, err, cleanupSentinel)
}
