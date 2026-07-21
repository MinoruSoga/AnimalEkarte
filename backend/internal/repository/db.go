package repository

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/repohelpers"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/apperrors"
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
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

// isUniqueConstraintErr はPostgreSQLのユニーク制約違反（23505）を判定する
// （BE9-2C R①: 実装は repohelpers へ昇格・本関数は既存呼び出し面互換の delegate）
func isUniqueConstraintErr(err error) bool {
	return repohelpers.IsUniqueConstraintErr(err)
}

// isFKConstraintErr はPostgreSQLの外部キー制約違反（23503）を判定する
func isFKConstraintErr(err error) bool {
	return repohelpers.IsFKConstraintErr(err)
}
