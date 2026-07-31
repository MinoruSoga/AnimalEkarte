package reservation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapableIDsFromExcluded(t *testing.T) {
	t.Parallel()
	universe := []uint64{1, 2, 3}

	t.Run("empty exclusion yields full universe as capable", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []uint64{1, 2, 3}, capableIDsFromExcluded(universe, nil))
		assert.Equal(t, []uint64{1, 2, 3}, capableIDsFromExcluded(universe, []uint64{}))
	})

	t.Run("partial exclusion", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []uint64{1, 3}, capableIDsFromExcluded(universe, []uint64{2}))
	})

	t.Run("full exclusion yields empty capable", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, capableIDsFromExcluded(universe, []uint64{1, 2, 3}))
	})

	t.Run("excluded IDs outside universe are ignored", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []uint64{2, 3}, capableIDsFromExcluded(universe, []uint64{1, 99}))
	})

	t.Run("empty universe", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, capableIDsFromExcluded(nil, []uint64{1}))
	})
}

func TestExcludedIDsFromCapable(t *testing.T) {
	t.Parallel()
	universe := []uint64{1, 2, 3}

	t.Run("empty capable yields full universe as excluded", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []uint64{1, 2, 3}, excludedIDsFromCapable(universe, nil))
	})

	t.Run("partial capable", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, []uint64{2}, excludedIDsFromCapable(universe, []uint64{1, 3}))
	})

	t.Run("full capable yields empty excluded", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, excludedIDsFromCapable(universe, []uint64{1, 2, 3}))
	})

	t.Run("round-trip inverse", func(t *testing.T) {
		t.Parallel()
		excluded := []uint64{2}
		capable := capableIDsFromExcluded(universe, excluded)
		assert.Equal(t, excluded, excludedIDsFromCapable(universe, capable))
	})
}
