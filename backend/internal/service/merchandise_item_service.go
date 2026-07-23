package service

import "github.com/animal-ekarte/backend/internal/inventory"

// CreateMerchandiseItemInput is a BE9-2F compatibility alias. Delete after
// the legacy Services/Handler aggregators consume inventory directly.
type CreateMerchandiseItemInput = inventory.CreateMerchandiseItemInput

// UpdateMerchandiseItemInput is a BE9-2F compatibility alias.
type UpdateMerchandiseItemInput = inventory.UpdateMerchandiseItemInput

// MerchandiseItemService is a BE9-2F compatibility alias.
type MerchandiseItemService = inventory.MerchandiseItemService

// NewMerchandiseItemService is a BE9-2F compatibility constructor.
func NewMerchandiseItemService(repo inventory.MerchandiseItemRepository) MerchandiseItemService {
	return inventory.NewMerchandiseItemService(repo)
}
