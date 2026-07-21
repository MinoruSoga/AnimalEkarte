package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// BillingItemRepository は internal/billing への移行facade（BE9-2C B③・BE9-2F削除予定）。
type BillingItemRepository = billing.BillingItemRepository

// NewBillingItemRepository は internal/billing の実装を返す（BE9-2C B③ facade）。
func NewBillingItemRepository(db *gorm.DB) BillingItemRepository {
	return billing.NewBillingItemRepository(db)
}
