package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestBuildDiscountSuggestions(t *testing.T) {
	rateCampaign := func(id uint64, name string, v float64) *model.Campaign {
		return &model.Campaign{ID: id, Name: name, DiscountType: model.CampaignDiscountTypeRate, DiscountValue: v}
	}
	amountCampaign := func(id uint64, name string, v float64) *model.Campaign {
		return &model.Campaign{ID: id, Name: name, DiscountType: model.CampaignDiscountTypeAmount, DiscountValue: v}
	}

	tests := []struct {
		name         string
		itemSubtotal int64
		campaigns    []*model.Campaign
		ownerRate    float64
		wantLen      int
		verify       func(t *testing.T, got []DiscountSuggestion)
	}{
		{
			name:         "率型キャンペーン1件 → 候補1件を返す",
			itemSubtotal: 1000,
			campaigns:    []*model.Campaign{rateCampaign(1, "夏季キャンペーン", 20)},
			ownerRate:    0,
			wantLen:      1,
			verify: func(t *testing.T, got []DiscountSuggestion) {
				assert.Equal(t, "campaign", got[0].Type)
				assert.Equal(t, uint64(1), *got[0].CampaignID)
				assert.Equal(t, "夏季キャンペーン", got[0].Name)
				assert.Equal(t, "rate", got[0].DiscountType)
				assert.Equal(t, int64(200), got[0].Amount)
			},
		},
		{
			name:         "額型キャンペーン1件 → 固定額で候補を返す",
			itemSubtotal: 1000,
			campaigns:    []*model.Campaign{amountCampaign(2, "固定割引", 300)},
			ownerRate:    0,
			wantLen:      1,
			verify: func(t *testing.T, got []DiscountSuggestion) {
				assert.Equal(t, int64(300), got[0].Amount)
			},
		},
		{
			name:         "キャンペーン額が明細小計を超える → 小計でクランプされる",
			itemSubtotal: 200,
			campaigns:    []*model.Campaign{amountCampaign(3, "大幅割引", 800)},
			ownerRate:    0,
			wantLen:      1,
			verify: func(t *testing.T, got []DiscountSuggestion) {
				assert.Equal(t, int64(200), got[0].Amount)
			},
		},
		{
			name:         "nil キャンペーンはスキップされる",
			itemSubtotal: 1000,
			campaigns:    []*model.Campaign{nil, rateCampaign(4, "有効キャンペーン", 10)},
			ownerRate:    0,
			wantLen:      1,
		},
		{
			name:         "Amount<=0 のキャンペーンは除外される",
			itemSubtotal: 1000,
			campaigns:    []*model.Campaign{rateCampaign(5, "ゼロ割引", 0)},
			ownerRate:    0,
			wantLen:      0,
		},
		{
			name:         "飼主割引率 > 0 → 候補に追加される",
			itemSubtotal: 1000,
			campaigns:    nil,
			ownerRate:    15,
			wantLen:      1,
			verify: func(t *testing.T, got []DiscountSuggestion) {
				assert.Equal(t, "owner", got[0].Type)
				assert.Nil(t, got[0].CampaignID)
				assert.Equal(t, "飼主割引", got[0].Name)
				assert.Equal(t, int64(150), got[0].Amount)
			},
		},
		{
			name:         "飼主割引率 0 → 候補に追加されない",
			itemSubtotal: 1000,
			campaigns:    nil,
			ownerRate:    0,
			wantLen:      0,
		},
		{
			name:         "キャンペーン + 飼主割引の両方 → 両方候補に含まれる",
			itemSubtotal: 1000,
			campaigns:    []*model.Campaign{rateCampaign(6, "キャンペーンA", 10)},
			ownerRate:    5,
			wantLen:      2,
		},
		{
			name:         "飼主割引額が明細小計を超える → 小計でクランプされる",
			itemSubtotal: 100,
			campaigns:    nil,
			ownerRate:    500,
			wantLen:      1,
			verify: func(t *testing.T, got []DiscountSuggestion) {
				assert.Equal(t, int64(100), got[0].Amount)
			},
		},
		{
			name:         "キャンペーンなし・飼主割引なし → 空リストを返す",
			itemSubtotal: 1000,
			campaigns:    nil,
			ownerRate:    0,
			wantLen:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDiscountSuggestions(tt.itemSubtotal, tt.campaigns, tt.ownerRate)
			assert.Len(t, got, tt.wantLen)
			if tt.verify != nil {
				tt.verify(t, got)
			}
		})
	}
}

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
