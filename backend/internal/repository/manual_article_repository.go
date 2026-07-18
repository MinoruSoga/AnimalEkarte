package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/manualarticle"
)

// ManualArticleRepository is a stable facade alias for the manualarticle domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type ManualArticleRepository = manualarticle.Repository

// NewManualArticleRepository constructs the manual article repository.
func NewManualArticleRepository(db *gorm.DB) ManualArticleRepository {
	return manualarticle.New(db)
}
