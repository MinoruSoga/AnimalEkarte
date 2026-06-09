package service

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestCalculateItemCampaignDiscount(t *testing.T) {
	rateCampaign := func(v float64) *model.Campaign {
		return &model.Campaign{DiscountType: model.CampaignDiscountTypeRate, DiscountValue: v}
	}
	amountCampaign := func(v float64) *model.Campaign {
		return &model.Campaign{DiscountType: model.CampaignDiscountTypeAmount, DiscountValue: v}
	}

	tests := []struct {
		name              string
		itemSubtotal      int64
		campaign          *model.Campaign
		ownerDiscountRate float64
		want              int64
	}{
		{
			name:              "キャンペーン率 > 飼主率 → キャンペーン採用",
			itemSubtotal:      1000,
			campaign:          rateCampaign(20), // 200
			ownerDiscountRate: 10,               // 100
			want:              200,
		},
		{
			name:              "飼主率 > キャンペーン率 → 飼主採用 (Q3=B)",
			itemSubtotal:      1000,
			campaign:          rateCampaign(5), // 50
			ownerDiscountRate: 15,              // 150
			want:              150,
		},
		{
			name:              "キャンペーン額 vs 飼主率 → 高い方を円換算で比較",
			itemSubtotal:      1000,
			campaign:          amountCampaign(300), // 固定300
			ownerDiscountRate: 10,                  // 100
			want:              300,
		},
		{
			name:              "キャンペーン nil → 飼主割引のみ",
			itemSubtotal:      1000,
			campaign:          nil,
			ownerDiscountRate: 10,
			want:              100,
		},
		{
			name:              "割引が明細小計を超える → 小計でクランプ",
			itemSubtotal:      500,
			campaign:          amountCampaign(800),
			ownerDiscountRate: 0,
			want:              500,
		},
		{
			name:              "小計 0 → 割引 0",
			itemSubtotal:      0,
			campaign:          rateCampaign(20),
			ownerDiscountRate: 10,
			want:              0,
		},
		{
			name:              "割引なし(両方0)",
			itemSubtotal:      1000,
			campaign:          nil,
			ownerDiscountRate: 0,
			want:              0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateItemCampaignDiscount(tt.itemSubtotal, tt.campaign, tt.ownerDiscountRate)
			if got != tt.want {
				t.Errorf("CalculateItemCampaignDiscount() = %d, want %d", got, tt.want)
			}
		})
	}
}
