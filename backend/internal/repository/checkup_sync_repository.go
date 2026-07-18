package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/checkupsync"
)

// CheckupSyncPreviewRow is a stable facade alias for the checkupsync domain
// package (BE8-4). Service/handler imports keep using repository.* so the
// split does not churn all importers.
type CheckupSyncPreviewRow = checkupsync.PreviewRow

// FindCheckupSyncPreviewParams is a stable facade alias for the checkupsync
// domain package (BE8-4). Service/handler imports keep using repository.* so
// the split does not churn all importers.
type FindCheckupSyncPreviewParams = checkupsync.FindPreviewParams

// CheckupSyncRepository is a stable facade alias for the checkupsync domain
// package (BE8-4). Service/handler imports keep using repository.* so the
// split does not churn all importers.
type CheckupSyncRepository = checkupsync.Repository

// NewCheckupSyncRepository constructs the checkup sync preview repository.
func NewCheckupSyncRepository(db *gorm.DB) CheckupSyncRepository {
	return checkupsync.New(db)
}
