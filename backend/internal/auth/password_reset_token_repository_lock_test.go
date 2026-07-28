package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestPasswordResetTokenRepository_LockingOperationsRequireAmbientTransaction(t *testing.T) {
	repo := NewPasswordResetTokenRepository(nil)

	t.Run("find for update", func(t *testing.T) {
		token, err := repo.FindByTokenHashForUpdate(context.Background(), "hash")

		assert.Nil(t, token)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("find latest for account", func(t *testing.T) {
		token, err := repo.FindLatestByAccountIDForUpdate(context.Background(), 1)

		assert.Nil(t, token)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INTERNAL", appErr.Code)
	})

	t.Run("strict consume", func(t *testing.T) {
		err := repo.ConsumeByID(context.Background(), 1)

		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INTERNAL", appErr.Code)
	})
}
