package service

import "github.com/animal-ekarte/backend/internal/inventory"

// CreateInventoryInput is a BE9-2F compatibility alias. Delete after the
// legacy Services/Handler aggregators construct and consume inventory directly.
type CreateInventoryInput = inventory.CreateInventoryInput

// UpdateInventoryInput is a BE9-2F compatibility alias.
type UpdateInventoryInput = inventory.UpdateInventoryInput

// InventoryService is a BE9-2F compatibility alias.
type InventoryService = inventory.InventoryService

// NewInventoryService is a BE9-2F compatibility constructor.
func NewInventoryService(repo inventory.Repository) InventoryService {
	return inventory.NewInventoryService(repo)
}
