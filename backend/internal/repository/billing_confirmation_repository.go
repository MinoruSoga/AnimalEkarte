package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// BillingConfirmationRepository は internal/billing への移行facade（BE9-2C B②・BE9-2F削除予定）。
type BillingConfirmationRepository = billing.BillingConfirmationRepository

// NewBillingConfirmationRepository は internal/billing の実装を返す（BE9-2C B② facade）。
func NewBillingConfirmationRepository(db *gorm.DB) BillingConfirmationRepository {
	return billing.NewBillingConfirmationRepository(db)
}
