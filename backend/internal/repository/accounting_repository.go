package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// AccountingRepository は internal/billing への移行facade（BE9-2C B④・BE9-2F削除予定）。
type AccountingRepository = billing.AccountingRepository

// NewAccountingRepository は internal/billing の実装を返す（BE9-2C B④ facade）。
func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return billing.NewAccountingRepository(db)
}
