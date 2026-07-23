package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
)

// PermissionGroupRepository is a compatibility alias for auth persistence.
type PermissionGroupRepository = auth.PermissionGroupRepository

// NewPermissionGroupRepository constructs permission-group persistence.
func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return auth.NewPermissionGroupRepository(db)
}
