package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// campaignResponse は割引キャンペーンのレスポンス (#81)
type campaignResponse struct {
	ID               uint64   `json:"id"`
	ClinicID         uint64   `json:"clinic_id"`
	Name             string   `json:"name"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	DiscountType     string   `json:"discount_type"`
	DiscountValue    float64  `json:"discount_value"`
	IsActive         bool     `json:"is_active"`
	SortOrder        int      `json:"sort_order"`
	TargetCategories []string `json:"target_categories"`
	TargetItemIDs    []uint64 `json:"target_item_ids"`
}

func toCampaignResponse(m *model.Campaign) campaignResponse {
	cats := make([]string, 0, len(m.TargetCategories))
	for _, tc := range m.TargetCategories {
		cats = append(cats, string(tc.Category))
	}
	ids := make([]uint64, 0, len(m.TargetItems))
	for _, ti := range m.TargetItems {
		ids = append(ids, ti.MerchandiseItemID)
	}
	return campaignResponse{
		ID:               m.ID,
		ClinicID:         m.ClinicID,
		Name:             m.Name,
		StartDate:        m.StartDate.In(time.Local).Format(time.DateOnly),
		EndDate:          m.EndDate.In(time.Local).Format(time.DateOnly),
		DiscountType:     string(m.DiscountType),
		DiscountValue:    m.DiscountValue,
		IsActive:         m.IsActive,
		SortOrder:        m.SortOrder,
		TargetCategories: cats,
		TargetItemIDs:    ids,
	}
}
