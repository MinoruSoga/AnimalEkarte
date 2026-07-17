package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/merchandiseitem"
)

// MerchandiseItemRepository is a stable facade alias for merchandiseitem.
type MerchandiseItemRepository = merchandiseitem.Repository

// NewMerchandiseItemRepository constructs the merchandise item repository.
func NewMerchandiseItemRepository(db *gorm.DB) MerchandiseItemRepository {
	return merchandiseitem.New(db)
}
