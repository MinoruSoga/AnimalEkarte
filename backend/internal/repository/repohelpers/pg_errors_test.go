package repohelpers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueConstraintErr(t *testing.T) {
	t.Run("23505は一意制約違反として検出される", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.True(t, IsUniqueConstraintErr(err))
	})

	t.Run("fmt.Errorfでラップされていてもerrors.Asで解除して検出される", func(t *testing.T) {
		wrapped := fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"})
		assert.True(t, IsUniqueConstraintErr(wrapped))
	})

	t.Run("別のPGエラーコードはfalse", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.False(t, IsUniqueConstraintErr(err))
	})

	t.Run("PG以外のエラーはfalse", func(t *testing.T) {
		assert.False(t, IsUniqueConstraintErr(errors.New("plain error")))
	})

	t.Run("nilはfalse", func(t *testing.T) {
		assert.False(t, IsUniqueConstraintErr(nil))
	})
}

func TestIsFKConstraintErr(t *testing.T) {
	t.Run("23503はFK制約違反として検出される", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.True(t, IsFKConstraintErr(err))
	})

	t.Run("fmt.Errorfでラップされていてもerrors.Asで解除して検出される", func(t *testing.T) {
		wrapped := fmt.Errorf("delete failed: %w", &pgconn.PgError{Code: "23503"})
		assert.True(t, IsFKConstraintErr(wrapped))
	})

	t.Run("別のPGエラーコードはfalse", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.False(t, IsFKConstraintErr(err))
	})

	t.Run("PG以外のエラーはfalse", func(t *testing.T) {
		assert.False(t, IsFKConstraintErr(errors.New("plain error")))
	})

	t.Run("nilはfalse", func(t *testing.T) {
		assert.False(t, IsFKConstraintErr(nil))
	})
}
