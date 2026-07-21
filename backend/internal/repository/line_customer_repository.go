package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// LineCustomerRepository は internal/lstep への移行facade（BE9-2C L②・BE9-2F削除予定）。
type LineCustomerRepository = lstep.LineCustomerRepository

// NewLineCustomerRepository は internal/lstep の実装を返す（BE9-2C L② facade）。
func NewLineCustomerRepository(db *gorm.DB) LineCustomerRepository {
	return lstep.NewLineCustomerRepository(db)
}
