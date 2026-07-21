package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// CashRegisterCloseRepository は internal/billing への移行facade（BE9-2C B⑤・BE9-2F削除予定）。
type CashRegisterCloseRepository = billing.CashRegisterCloseRepository

// NewCashRegisterCloseRepository は internal/billing の実装を返す（BE9-2C B⑤ facade）。
func NewCashRegisterCloseRepository(db *gorm.DB) CashRegisterCloseRepository {
	return billing.NewCashRegisterCloseRepository(db)
}
