package errors

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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

func TestWrapAlreadyExists(t *testing.T) {
	err := WrapAlreadyExists("owner", "identifier")
	assert.NotNil(t, err)
	assert.True(t, IsAlreadyExists(err))
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "ALREADY_EXISTS", appErr.Code)
}

func TestIsAlreadyExists(t *testing.T) {
	assert.True(t, IsAlreadyExists(ErrAlreadyExists))
	assert.True(t, IsAlreadyExists(Wrap(ErrAlreadyExists, "context")))
	assert.False(t, IsAlreadyExists(ErrNotFound))
}

func TestWrapConflict(t *testing.T) {
	err := WrapConflict("conflict message")
	assert.NotNil(t, err)
	assert.True(t, IsConflict(err))
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "CONFLICT", appErr.Code)
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(ErrConflict))
	assert.True(t, IsConflict(Wrap(ErrConflict, "context")))
	assert.False(t, IsConflict(ErrNotFound))
}

func TestWrapForbidden(t *testing.T) {
	err := WrapForbidden("forbidden message")
	assert.NotNil(t, err)
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "FORBIDDEN", appErr.Code)
}

func TestWrapUnauthorized(t *testing.T) {
	err := WrapUnauthorized("unauthorized message")
	assert.NotNil(t, err)
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "UNAUTHORIZED", appErr.Code)
}

func TestWrapInternalServerError(t *testing.T) {
	err := WrapInternalServerError("internal message")
	assert.NotNil(t, err)
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code)
}

func TestWrapNotImplemented(t *testing.T) {
	err := WrapNotImplemented("not implemented message")
	assert.NotNil(t, err)
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "NOT_IMPLEMENTED", appErr.Code)
}

func TestWrapBadGateway(t *testing.T) {
	err := WrapBadGateway("bad gateway message")
	assert.NotNil(t, err)
	assert.True(t, IsBadGateway(err))
	var appErr *AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "BAD_GATEWAY", appErr.Code)
}

func TestIsBadGateway(t *testing.T) {
	assert.True(t, IsBadGateway(ErrBadGateway))
	assert.True(t, IsBadGateway(Wrap(ErrBadGateway, "context")))
	assert.False(t, IsBadGateway(ErrNotFound))
}

func TestFromGORM(t *testing.T) {
	t.Run("returns nil for nil error", func(t *testing.T) {
		assert.Nil(t, FromGORM(nil, "resource", "1"))
	})

	t.Run("maps gorm.ErrRecordNotFound to WrapNotFound", func(t *testing.T) {
		err := FromGORM(gorm.ErrRecordNotFound, "pet", "123")
		assert.True(t, IsNotFound(err))
	})

	t.Run("maps encode error to WrapInvalidInput", func(t *testing.T) {
		err1 := FromGORM(errors.New("unable to encode something"), "pet", "123")
		assert.True(t, IsInvalidInput(err1))

		err2 := FromGORM(errors.New("value greater than maximum value"), "pet", "123")
		assert.True(t, IsInvalidInput(err2))

		err3 := FromGORM(errors.New("value less than minimum value"), "pet", "123")
		assert.True(t, IsInvalidInput(err3))
	})

	t.Run("maps pgconn.PgError unique_violation to WrapAlreadyExists", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: "23505",
		}
		err := FromGORM(pgErr, "pet", "123")
		assert.True(t, IsAlreadyExists(err))
	})

	t.Run("maps pgconn.PgError foreign_key_violation to WrapInvalidInput", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: "23503",
		}
		err := FromGORM(pgErr, "pet", "123")
		assert.True(t, IsInvalidInput(err))
	})

	t.Run("maps pgconn.PgError numeric_value_out_of_range to WrapInvalidInput", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: "22003",
		}
		err := FromGORM(pgErr, "pet", "123")
		assert.True(t, IsInvalidInput(err))
	})

	t.Run("maps pgconn.PgError invalid_text_representation to WrapInvalidInput", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: "22P02",
		}
		err := FromGORM(pgErr, "pet", "123")
		assert.True(t, IsInvalidInput(err))
	})

	t.Run("maps other pgconn.PgError to generic database error", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code: "99999",
		}
		err := FromGORM(pgErr, "pet", "123")
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("maps generic error to generic database error", func(t *testing.T) {
		genericErr := errors.New("some DB failure")
		err := FromGORM(genericErr, "pet", "123")
		assert.Contains(t, err.Error(), "database error")
	})
}
