package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/refund"
)

// RefundRepository is a stable facade alias for the refund domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type RefundRepository = refund.Repository

// NewRefundRepository constructs the refund repository.
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return refund.New(db)
}
