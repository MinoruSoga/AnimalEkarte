package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// PaymentMethodMasterRepository は internal/billing への移行facade（BE9-2C B①・BE9-2F削除予定）。
type PaymentMethodMasterRepository = billing.PaymentMethodMasterRepository

// NewPaymentMethodMasterRepository は internal/billing の実装を返す（BE9-2C B① facade）。
func NewPaymentMethodMasterRepository(db *gorm.DB) PaymentMethodMasterRepository {
	return billing.NewPaymentMethodMasterRepository(db)
}
