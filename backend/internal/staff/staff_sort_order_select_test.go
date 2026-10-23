package staff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// selectReservationStaffSwap の並び順は DB の ORDER BY staffs.sort_order ASC と同じく
// NULL を末尾に置く（NULLS LAST）。NULL を 0 へ推測補完して先頭へ並べない。
func TestSelectReservationStaffSwap_NullSortOrderIsLast(t *testing.T) {
	one, two := 1, 2
	locked := []model.Staff{
		{ID: 10, SortOrder: nil},
		{ID: 20, SortOrder: &two},
		{ID: 30, SortOrder: &one},
	}

	// 並び順は 30(1) → 20(2) → 10(NULL)。20 の down の相手は NULL の 10。
	target, neighbor, err := selectReservationStaffSwap(locked, 20, "down")
	require.NoError(t, err)
	require.NotNil(t, neighbor)
	assert.Equal(t, uint64(20), target.ID)
	assert.Equal(t, uint64(10), neighbor.ID)

	// 30 の up は先頭なので相手なし。
	_, neighbor, err = selectReservationStaffSwap(locked, 30, "up")
	require.NoError(t, err)
	assert.Nil(t, neighbor)
}

func TestSelectReservationStaffSwap_NullTieBreaksByID(t *testing.T) {
	locked := []model.Staff{{ID: 7, SortOrder: nil}, {ID: 3, SortOrder: nil}}

	target, neighbor, err := selectReservationStaffSwap(locked, 3, "down")
	require.NoError(t, err)
	require.NotNil(t, neighbor)
	assert.Equal(t, uint64(3), target.ID)
	assert.Equal(t, uint64(7), neighbor.ID)
}
