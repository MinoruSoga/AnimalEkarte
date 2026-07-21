package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// RefundRepository は internal/billing への移行facade（BE9-2C B③・BE9-2F削除予定）。
type RefundRepository = billing.RefundRepository

// NewRefundRepository は internal/billing の実装を返す（BE9-2C B③ facade）。
func NewRefundRepository(db *gorm.DB) RefundRepository {
	return billing.NewRefundRepository(db)
}
