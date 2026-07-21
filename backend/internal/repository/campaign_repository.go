package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/billing"
)

// CampaignRepository は internal/billing への移行facade（BE9-2C B①・BE9-2F削除予定）。
type CampaignRepository = billing.CampaignRepository

// NewCampaignRepository は internal/billing の実装を返す（BE9-2C B① facade）。
func NewCampaignRepository(db *gorm.DB) CampaignRepository {
	return billing.NewCampaignRepository(db)
}
