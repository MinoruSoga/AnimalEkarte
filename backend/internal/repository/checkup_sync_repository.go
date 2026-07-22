package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/lstep"
)

// CheckupSyncPreviewRow is a stable facade alias for the lstep domain package.
// Remove the facade after residual repository-aggregate consumers are migrated.
type CheckupSyncPreviewRow = lstep.CheckupSyncPreviewRow

// FindCheckupSyncPreviewParams is a stable facade alias for the lstep domain package.
type FindCheckupSyncPreviewParams = lstep.FindCheckupSyncPreviewParams

// CheckupSyncRepository is a stable facade alias for the lstep domain package.
type CheckupSyncRepository = lstep.CheckupSyncRepository

// NewCheckupSyncRepository constructs the checkup sync preview repository.
func NewCheckupSyncRepository(db *gorm.DB) CheckupSyncRepository {
	return lstep.NewCheckupSyncRepository(db)
}
