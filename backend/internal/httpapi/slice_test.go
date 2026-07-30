package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestValidateReorderIDs(t *testing.T) {
	t.Parallel()

	t.Run("accepts valid unique list within limit", func(t *testing.T) {
		t.Parallel()
		err := ValidateReorderIDs([]uint64{3, 1, 2})
		assert.NoError(t, err)
	})

	t.Run("rejects empty list", func(t *testing.T) {
		t.Parallel()
		err := ValidateReorderIDs(nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "empty must be InvalidInput: %v", err)

		err = ValidateReorderIDs([]uint64{})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("rejects duplicates", func(t *testing.T) {
		t.Parallel()
		err := ValidateReorderIDs([]uint64{1, 2, 1})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "dups must be InvalidInput: %v", err)
	})

	t.Run("rejects over MaxReorderIDs without depending on repo", func(t *testing.T) {
		t.Parallel()
		ids := make([]uint64, MaxReorderIDs+1)
		for i := range ids {
			ids[i] = uint64(i + 1)
		}
		err := ValidateReorderIDs(ids)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "over-limit must be InvalidInput: %v", err)
	})

	t.Run("accepts exactly MaxReorderIDs unique ids", func(t *testing.T) {
		t.Parallel()
		ids := make([]uint64, MaxReorderIDs)
		for i := range ids {
			ids[i] = uint64(i + 1)
		}
		assert.NoError(t, ValidateReorderIDs(ids))
	})
}
