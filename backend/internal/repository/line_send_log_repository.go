package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LineSendLogRepository は internal/lstep への移行facade（BE9-2C L②・BE9-2F削除予定）。
type LineSendLogRepository = lstep.LineSendLogRepository

// NewLineSendLogRepository は internal/lstep の実装を返す（BE9-2C L② facade）。
func NewLineSendLogRepository(db *gorm.DB) LineSendLogRepository {
	return lstep.NewLineSendLogRepository(db)
}
