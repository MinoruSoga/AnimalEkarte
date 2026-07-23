package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/inventory"
)

// MerchandiseItemRepository is a BE9-2F compatibility alias for the legacy
// Repositories aggregator and its billing composition caller. Delete after
// those callers construct internal/inventory directly.
type MerchandiseItemRepository = inventory.MerchandiseItemRepository

// NewMerchandiseItemRepository is a BE9-2F compatibility constructor.
func NewMerchandiseItemRepository(db *gorm.DB) MerchandiseItemRepository {
	return inventory.NewMerchandiseItemRepository(db)
}
