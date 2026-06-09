package service

import (
	"math"

	"github.com/animal-ekarte/backend/internal/model"
)

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
