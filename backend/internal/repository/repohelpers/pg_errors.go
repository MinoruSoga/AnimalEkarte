package repohelpers

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
// （BE9-2C R①: repository/db.go から昇格。owner/pet/staff/reservation 等の恒久ドメイン跨ぎ）。
func IsUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// IsFKConstraintErr はPostgreSQLの外部キー制約違反（23503）を判定する
func IsFKConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
