package owner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnerRepository_FindAll_IdeographicSpaceFourWay(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{name: "DB fullwidth × query fullwidth", storedName: "飼主全角全角姓　飼主全角全角名", query: "飼主全角全角姓　飼主全角全角名"},
		{name: "DB fullwidth × query halfwidth", storedName: "飼主全角半角姓　飼主全角半角名", query: "飼主全角半角姓 飼主全角半角名"},
		{name: "DB halfwidth × query fullwidth", storedName: "飼主半角全角姓 飼主半角全角名", query: "飼主半角全角姓　飼主半角全角名"},
		{name: "DB halfwidth × query halfwidth", storedName: "飼主半角半角姓 飼主半角半角名", query: "飼主半角半角姓 飼主半角半角名"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := makeTestOwner(t, db, clinicA, tt.storedName)
			foreign := makeTestOwner(t, db, clinicB, tt.storedName)

			got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, tt.query)
			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, fixture.ID, got[0].ID)
			require.NotEqual(t, foreign.ID, got[0].ID)
		})
	}

	t.Run("ideographic-space-only query returns no rows", func(t *testing.T) {
		_ = makeTestOwner(t, db, clinicA, "空白除外飼主 名")
		got, total, err := repo.FindAll(ctx, []uint64{clinicA}, 1, 100, "　　")
		require.NoError(t, err)
		require.Equal(t, int64(0), total)
		require.Empty(t, got)
	})
}
