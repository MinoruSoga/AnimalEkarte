package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// NewDB temporarily preserves the repository package API for existing
// consumers. Remove this delegate when BE9-2F moves them to dbconn.OpenGORM.
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	return dbconn.OpenGORM(cfg)
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
