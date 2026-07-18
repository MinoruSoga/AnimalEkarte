package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/inventory"
)

// InventoryRepository is a stable facade alias for the inventory domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type InventoryRepository = inventory.Repository

// NewInventoryRepository constructs the inventory repository.
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return inventory.New(db)
}
