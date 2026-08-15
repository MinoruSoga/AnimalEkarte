package billing

import (
	"math"

	"github.com/animal-ekarte/backend/internal/model"
)

// DiscountSuggestion は明細に適用可能な割引の候補（#81 Q-I スタッフ選択用）。
type DiscountSuggestion struct {
	Type          string  `json:"type"`           // "campaign" | "owner"
	CampaignID    *uint64 `json:"campaign_id"`    // キャンペーンID（"owner" 時は nil）
	Name          string  `json:"name"`           // キャンペーン名 or "飼主割引"
	DiscountType  string  `json:"discount_type"`  // "rate" | "amount" | ""
	DiscountValue float64 `json:"discount_value"` // 割引率(%) or 固定額(円)
	Amount        int64   `json:"amount"`         // 円換算の割引額
}

// BuildDiscountSuggestions は明細小計・全適用キャンペーン・飼主割引率から選択肢を構築する。
// 同一 calculated_amount の重複は除去しない（名前が異なるキャンペーンは個別に表示）。
// Amount=0 の候補は除外する（割引なしは選択肢として意味がない）。
func BuildDiscountSuggestions(itemSubtotal int64, campaigns []*model.Campaign, ownerRate float64) []DiscountSuggestion {
	var suggestions []DiscountSuggestion
	for _, c := range campaigns {
		if c == nil {
			continue
		}
		var amount int64
		switch c.DiscountType {
		case model.CampaignDiscountTypeRate:
			amount = int64(math.Round(float64(itemSubtotal) * c.DiscountValue / 100))
		case model.CampaignDiscountTypeAmount:
			amount = int64(math.Round(c.DiscountValue))
		}
		amount = min(amount, itemSubtotal) // クランプ
		if amount <= 0 {
			continue
		}
		id := c.ID
		suggestions = append(suggestions, DiscountSuggestion{
			Type:          "campaign",
			CampaignID:    &id,
			Name:          c.Name,
			DiscountType:  string(c.DiscountType),
			DiscountValue: c.DiscountValue,
			Amount:        amount,
		})
	}
	if ownerRate > 0 {
		ownerAmount := min(int64(math.Round(float64(itemSubtotal)*ownerRate/100)), itemSubtotal)
		if ownerAmount > 0 {
			suggestions = append(suggestions, DiscountSuggestion{
				Type:          "owner",
				CampaignID:    nil,
				Name:          "飼主割引",
				DiscountType:  "rate",
				DiscountValue: ownerRate,
				Amount:        ownerAmount,
			})
		}
	}
	return suggestions
}

// CalculateItemCampaignDiscount は明細に適用する割引額(円)を計算する (#81 段階2)。
//
// Q-G=A(自動適用) / Q3=B / Q-I: キャンペーン割引と飼主割引(率)を円換算し、
// 高い方を採用する。割引額は明細小計(itemSubtotal = 単価×数量)を超えない。
//
//   - campaign が nil の場合はキャンペーン割引 0（飼主割引のみ）
//   - キャンペーン rate: itemSubtotal × discountValue / 100
//   - キャンペーン amount: discountValue(固定額、明細単位)
//   - 飼主割引: itemSubtotal × ownerDiscountRate / 100
func CalculateItemCampaignDiscount(itemSubtotal int64, campaign *model.Campaign, ownerDiscountRate float64) int64 {
	if itemSubtotal <= 0 {
		return 0
	}

	var campaignDiscount int64
	if campaign != nil {
		switch campaign.DiscountType {
		case model.CampaignDiscountTypeRate:
			campaignDiscount = int64(math.Round(float64(itemSubtotal) * campaign.DiscountValue / 100))
		case model.CampaignDiscountTypeAmount:
			campaignDiscount = int64(math.Round(campaign.DiscountValue))
		}
	}

	ownerDiscount := int64(math.Round(float64(itemSubtotal) * ownerDiscountRate / 100))

	// Q3=B: 高い方を採用。明細小計を超えないようクランプ。
	return min(max(campaignDiscount, ownerDiscount), itemSubtotal)
}
