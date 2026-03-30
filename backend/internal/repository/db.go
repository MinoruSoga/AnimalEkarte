package repository

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func NewDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open database connection")
	}

	// DB接続プール設定
	sqlDB, err := db.DB()
	if err != nil {
		return nil, apperrors.Wrap(err, "get sql.DB")
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// isUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
