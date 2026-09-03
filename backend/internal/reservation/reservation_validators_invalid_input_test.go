package reservation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestPassthroughOrInvalidDateTime(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		assert.NoError(t, passthroughOrInvalidDateTime(nil))
	})

	t.Run("AppError is not double-wrapped", func(t *testing.T) {
		inner := apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
		got := passthroughOrInvalidDateTime(inner)

		require.Equal(t, inner, got)
		var appErr *apperrors.AppError
		require.True(t, errors.As(got, &appErr))
		assert.Equal(t, "パスワードは英字と数字の両方を含めてください", appErr.Message)
		assert.NotContains(t, appErr.Message, ": invalid input")
	})

	t.Run("raw parse error becomes fixed Japanese", func(t *testing.T) {
		got := passthroughOrInvalidDateTime(fmt.Errorf("invalid date format, use YYYY-MM-DD"))

		var appErr *apperrors.AppError
		require.True(t, errors.As(got, &appErr))
		assert.Equal(t, errMsgInvalidDateTime, appErr.Message)
		assert.NotContains(t, appErr.Message, "invalid date format")
	})

	t.Run("MinutesSinceMidnight AppError is passed through", func(t *testing.T) {
		_, inner := MinutesSinceMidnight("xx")
		require.Error(t, inner)

		got := passthroughOrInvalidDateTime(inner)
		require.Equal(t, inner, got)
		var appErr *apperrors.AppError
		require.True(t, errors.As(got, &appErr))
		assert.Equal(t, `invalid time format "xx": must be HHMM`, appErr.Message)
	})
}
