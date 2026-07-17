package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/paymentmethod"
)

// PaymentMethodMasterRepository is a stable facade alias for the paymentmethod
// domain package. Service/handler imports keep using repository.* so the first
// domain split does not churn all importers.
type PaymentMethodMasterRepository = paymentmethod.Repository

// NewPaymentMethodMasterRepository constructs the payment-method master repository.
func NewPaymentMethodMasterRepository(db *gorm.DB) PaymentMethodMasterRepository {
	return paymentmethod.New(db)
}
