package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// createCampaignRequest は割引キャンペーン作成リクエスト (#81)
type createCampaignRequest struct {
	Name             string                  `json:"name"           binding:"required,max=255"`
	StartDate        string                  `json:"start_date"     binding:"required"` // YYYY-MM-DD
	EndDate          string                  `json:"end_date"       binding:"required"` // YYYY-MM-DD
	DiscountType     string                  `json:"discount_type"  binding:"required,oneof=rate amount"`
	DiscountValue    campaignDiscountValueDTO `json:"discount_value"`
	IsActive         bool                    `json:"is_active"`
	SortOrder        int                     `json:"sort_order"`
	TargetCategories []string                `json:"target_categories" binding:"omitempty,dive,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training house_call"`
	TargetItemIDs    []uint64                `json:"target_item_ids"`
}

func (r *createCampaignRequest) toServiceInput() (*CreateCampaignInput, error) {
	start, err := time.ParseInLocation(time.DateOnly, r.StartDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
	}
	end, err := time.ParseInLocation(time.DateOnly, r.EndDate, time.Local)
	if err != nil {
		return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
	}
	cats := make([]model.ItemCategory, 0, len(r.TargetCategories))
	for _, c := range r.TargetCategories {
		cats = append(cats, model.ItemCategory(c))
	}
	return &CreateCampaignInput{
		Name:             r.Name,
		StartDate:        start,
		EndDate:          end,
		DiscountType:     model.CampaignDiscountType(r.DiscountType),
		DiscountValue:    r.DiscountValue.orZero(),
		IsActive:         r.IsActive,
		SortOrder:        r.SortOrder,
		TargetCategories: cats,
		TargetItemIDs:    r.TargetItemIDs,
	}, nil
}

// updateCampaignRequest は割引キャンペーン更新リクエスト（nil = 未指定）。
// discount_value は省略/null で保持（*campaignDiscountValueDTO は null で nil になるため省略と同じ扱い）。
type updateCampaignRequest struct {
	Name             *string                   `json:"name" binding:"omitempty,max=255"`
	StartDate        *string                   `json:"start_date"`
	EndDate          *string                   `json:"end_date"`
	DiscountType     *string                   `json:"discount_type"  binding:"omitempty,oneof=rate amount"`
	DiscountValue    *campaignDiscountValueDTO `json:"discount_value"`
	IsActive         *bool                     `json:"is_active"`
	SortOrder        *int                      `json:"sort_order"`
	TargetCategories *[]string                 `json:"target_categories" binding:"omitempty,dive,oneof=examination test procedure surgery medicine food goods other vaccine trimming hotel training house_call"`
	TargetItemIDs    *[]uint64                 `json:"target_item_ids"`
}

func (r *updateCampaignRequest) toServiceInput() (*UpdateCampaignInput, error) {
	input := &UpdateCampaignInput{
		Name:          r.Name,
		IsActive:      r.IsActive,
		SortOrder:     r.SortOrder,
		TargetItemIDs: r.TargetItemIDs,
	}
	if r.DiscountValue != nil {
		v := r.DiscountValue.value
		input.DiscountValue = &v
	}
	if r.StartDate != nil {
		t, err := time.ParseInLocation(time.DateOnly, *r.StartDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("start_date は YYYY-MM-DD 形式で指定してください")
		}
		input.StartDate = &t
	}
	if r.EndDate != nil {
		t, err := time.ParseInLocation(time.DateOnly, *r.EndDate, time.Local)
		if err != nil {
			return nil, apperrors.WrapInvalidInput("end_date は YYYY-MM-DD 形式で指定してください")
		}
		input.EndDate = &t
	}
	if r.DiscountType != nil {
		dt := model.CampaignDiscountType(*r.DiscountType)
		input.DiscountType = &dt
	}
	if r.TargetCategories != nil {
		cats := make([]model.ItemCategory, 0, len(*r.TargetCategories))
		for _, c := range *r.TargetCategories {
			cats = append(cats, model.ItemCategory(c))
		}
		input.TargetCategories = &cats
	}
	return input, nil
}
