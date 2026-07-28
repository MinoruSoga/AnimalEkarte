package owner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOwnerRepository_FindAll_KanaSearchSymmetry(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	repo := newTestRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	tests := []struct {
		name       string
		storedName string
		query      string
	}{
		{
			name:       "katakana stored and katakana query",
			storedName: "カタカナアカネ",
			query:      "カタカナアカネ",
		},
		{
			name:       "katakana stored and hiragana query",
			storedName: "カタカナイツキ",
			query:      "かたかないつき",
		},
		{
			name:       "hiragana stored and katakana query",
			storedName: "ひらがなうらら",
			query:      "ヒラガナウララ",
		},
		{
			name:       "hiragana stored and hiragana query",
			storedName: "ひらがなおと",
			query:      "ひらがなおと",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := makeTestOwner(t, db, clinicID, tt.storedName)
			require.Empty(t, fixture.NameKana)

			got, total, err := repo.FindAll(ctx, []uint64{clinicID}, 1, 100, tt.query)

			require.NoError(t, err)
			require.Equal(t, int64(1), total)
			require.Len(t, got, 1)
			require.Equal(t, fixture.ID, got[0].ID)
		})
	}
}
