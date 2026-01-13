package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := Wrap(originalErr, "context message")

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "context message")
		assert.Contains(t, wrappedErr.Error(), "original error")
		assert.True(t, errors.Is(wrappedErr, originalErr))
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		wrappedErr := Wrap(nil, "context message")
		assert.Nil(t, wrappedErr)
	})
}

func TestWrapNotFound(t *testing.T) {
	err := WrapNotFound("pet", "123")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "pet")
	assert.Contains(t, err.Error(), "123")
	assert.True(t, IsNotFound(err))

	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "NOT_FOUND", appErr.Code)
}

func TestWrapInvalidInput(t *testing.T) {
	err := WrapInvalidInput("name is required")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "name is required")
	assert.True(t, IsInvalidInput(err))

	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INVALID_INPUT", appErr.Code)
}

func TestIsNotFound(t *testing.T) {
	t.Run("returns true for ErrNotFound", func(t *testing.T) {
		assert.True(t, IsNotFound(ErrNotFound))
	})

	t.Run("returns true for wrapped ErrNotFound", func(t *testing.T) {
		wrapped := Wrap(ErrNotFound, "context")
		assert.True(t, IsNotFound(wrapped))
	})

	t.Run("returns true for WrapNotFound error", func(t *testing.T) {
		err := WrapNotFound("pet", "123")
		assert.True(t, IsNotFound(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		assert.False(t, IsNotFound(ErrInvalidInput))
		assert.False(t, IsNotFound(errors.New("random error")))
	})
}

func TestIsInvalidInput(t *testing.T) {
	t.Run("returns true for ErrInvalidInput", func(t *testing.T) {
		assert.True(t, IsInvalidInput(ErrInvalidInput))
	})

	t.Run("returns true for wrapped ErrInvalidInput", func(t *testing.T) {
		wrapped := Wrap(ErrInvalidInput, "context")
		assert.True(t, IsInvalidInput(wrapped))
	})

	t.Run("returns true for WrapInvalidInput error", func(t *testing.T) {
		err := WrapInvalidInput("validation failed")
		assert.True(t, IsInvalidInput(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		assert.False(t, IsInvalidInput(ErrNotFound))
		assert.False(t, IsInvalidInput(errors.New("random error")))
	})
}

func TestAppError(t *testing.T) {
	t.Run("Error() returns message with wrapped error", func(t *testing.T) {
		appErr := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     errors.New("inner error"),
		}
		assert.Contains(t, appErr.Error(), "test message")
		assert.Contains(t, appErr.Error(), "inner error")
	})

	t.Run("Error() returns message without wrapped error", func(t *testing.T) {
		appErr := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     nil,
		}
		assert.Equal(t, "test message", appErr.Error())
	})

	t.Run("Unwrap() returns inner error", func(t *testing.T) {
		inner := errors.New("inner error")
		appErr := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     inner,
		}
		assert.Equal(t, inner, appErr.Unwrap())
	})
}

func TestIs(t *testing.T) {
	t.Run("delegates to errors.Is", func(t *testing.T) {
		assert.True(t, Is(ErrNotFound, ErrNotFound))
		assert.False(t, Is(ErrNotFound, ErrInvalidInput))
	})
}

func TestAs(t *testing.T) {
	t.Run("delegates to errors.As", func(t *testing.T) {
		err := WrapNotFound("pet", "123")
		var appErr *AppError
		assert.True(t, As(err, &appErr))
		assert.Equal(t, "NOT_FOUND", appErr.Code)
	})
}
