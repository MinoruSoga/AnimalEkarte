package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/campaign"
)

// CampaignRepository is a stable facade alias for the campaign domain
// package (BE8-4). Service/handler imports keep using repository.* so the split
// does not churn all importers.
type CampaignRepository = campaign.Repository

// NewCampaignRepository constructs the campaign repository.
func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return campaign.New(db)
}
