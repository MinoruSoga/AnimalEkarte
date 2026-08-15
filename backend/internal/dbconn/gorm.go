package dbconn

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
)

const (
	connMaxLifetime = 30 * time.Minute
	connMaxIdleTime = 5 * time.Minute
)

type gormOpener func(dsn string) (*gorm.DB, error)

type poolSettingsSetter interface {
	SetMaxOpenConns(int)
	SetMaxIdleConns(int)
	SetConnMaxLifetime(time.Duration)
	SetConnMaxIdleTime(time.Duration)
}

// OpenGORM opens the application database and applies the configured pool
// limits together with the established connection lifetime defaults.
func OpenGORM(cfg *config.Config) (*gorm.DB, error) {
	return openGORMWith(cfg, openPostgresGORM)
}

func openPostgresGORM(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{}) //nolint:wrapcheck // openGORMWith applies the legacy error context.
}

func openGORMWith(cfg *config.Config, open gormOpener) (*gorm.DB, error) {
	db, err := open(cfg.DSN())
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open database connection")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, apperrors.Wrap(err, "get sql.DB")
	}
	applyPoolSettings(sqlDB, cfg)

	return db, nil
}

func applyPoolSettings(pool poolSettingsSetter, cfg *config.Config) {
	pool.SetMaxOpenConns(cfg.DBMaxOpenConns)
	pool.SetMaxIdleConns(cfg.DBMaxIdleConns)
	pool.SetConnMaxLifetime(connMaxLifetime)
	pool.SetConnMaxIdleTime(connMaxIdleTime)
}
