package persistence

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueConstraintErr reports PostgreSQL unique-constraint violations.
func IsUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsFKConstraintErr reports PostgreSQL foreign-key violations.
func IsFKConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
