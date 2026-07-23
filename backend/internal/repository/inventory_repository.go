package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/inventory"
)

// InventoryRepository is a BE9-2F compatibility alias for the legacy
// Repositories aggregator and its remaining medicalrecord composition callers.
// Delete after those callers construct internal/inventory directly.
type InventoryRepository = inventory.Repository

// NewInventoryRepository is a BE9-2F compatibility constructor.
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return inventory.New(db)
}
