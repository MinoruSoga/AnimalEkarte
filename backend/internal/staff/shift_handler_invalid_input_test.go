package staff

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestPassthroughOrInvalidDateTime(t *testing.T) {
	t.Run("AppError is not double-wrapped", func(t *testing.T) {
		inner := apperrors.WrapInvalidInput("date query parameter is required (YYYY-MM-DD)")
		got := passthroughOrInvalidDateTime(inner)
		require.Equal(t, inner, got)
		var appErr *apperrors.AppError
		require.True(t, errors.As(got, &appErr))
		assert.Equal(t, "date query parameter is required (YYYY-MM-DD)", appErr.Message)
		assert.NotContains(t, appErr.Message, ": invalid input")
	})

	t.Run("raw parse error becomes fixed Japanese", func(t *testing.T) {
		got := passthroughOrInvalidDateTime(fmt.Errorf("invalid date: use YYYY-MM-DD"))
		var appErr *apperrors.AppError
		require.True(t, errors.As(got, &appErr))
		assert.Equal(t, errMsgInvalidDateTime, appErr.Message)
		assert.NotContains(t, appErr.Message, "invalid date")
	})
}
