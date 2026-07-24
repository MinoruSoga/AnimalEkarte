package billing

// estimate_owner_preload_clinic_isolation_test.go — AUD-005
// 汚染された Owner FK を持つ見積から、別 clinic の飼主情報が Preload されないことを検証する。
// Callers: go test ./internal/repository/ -run 'TestEstimateRepository_.*Owner'
// Existing equivalent: none. Synthetic fixtures only (clinicA=1, clinicB=2, owner names).

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupEstimateOwnerPreloadDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupEstimateRepoTestDB(t)
	db.Exec("TRUNCATE TABLE owners CASCADE")
	return db
}

func TestEstimateRepository_FindByID_FindAll_DoesNotPreloadForeignOwner(t *testing.T) {
	db := setupEstimateOwnerPreloadDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	ownerB := testdb.MakeTestOwner(t, db, clinicB, "医院Bの飼主")
	ownerBID := ownerB.ID
	contaminated := &model.Estimate{
		ClinicID: clinicA,
		OwnerID:  &ownerBID,
		Status:   model.EstimateStatusDraft,
		Title:    "汚染見積",
	}
	require.NoError(t, repo.Create(ctx, contaminated))

	t.Run("FindByID does not preload foreign owner", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, contaminated.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.OwnerID)
		assert.Equal(t, ownerBID, *got.OwnerID)
		assert.Nil(t, got.Owner, "foreign Owner must not be preloaded")
	})

	t.Run("FindAll does not preload foreign owner", func(t *testing.T) {
		items, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, 1, 50)
		require.NoError(t, err)
		require.GreaterOrEqual(t, total, int64(1))
		var found *model.Estimate
		for i := range items {
			if items[i].ID == contaminated.ID {
				found = &items[i]
				break
			}
		}
		require.NotNil(t, found)
		assert.Nil(t, found.Owner, "foreign Owner must not be preloaded")
	})
}
