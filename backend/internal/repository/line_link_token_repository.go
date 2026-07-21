package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LineLinkTokenRepository は internal/lstep への移行facade（BE9-2C L②・BE9-2F削除予定）。
type LineLinkTokenRepository = lstep.LineLinkTokenRepository

// NewLineLinkTokenRepository は internal/lstep の実装を返す（BE9-2C L② facade）。
func NewLineLinkTokenRepository(db *gorm.DB) LineLinkTokenRepository {
	return lstep.NewLineLinkTokenRepository(db)
}
