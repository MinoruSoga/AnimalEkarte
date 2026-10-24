package billing

import (
	"math/big"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// DiscountSuggestion は明細に適用可能な割引の候補（#81 Q-I スタッフ選択用）。
// DiscountValue は保存値を exact のまま返す JSON number（float64 へ変換しない）。
type DiscountSuggestion struct {
	Type          string             `json:"type"`           // "campaign" | "owner"
	CampaignID    *uint64            `json:"campaign_id"`    // キャンペーンID（"owner" 時は nil）
	Name          string             `json:"name"`           // キャンペーン名 or "飼主割引"
	DiscountType  string             `json:"discount_type"`  // "rate" | "amount" | ""
	DiscountValue model.ExactNumeric `json:"discount_value"` // 割引率(%) or 固定額(円)
	Amount        int64              `json:"amount"`         // 円換算の割引額
}

// discountBaseAmount は割引額を計算する基数（円）を返す。
// 旧実装（int64(float64(単価)*数量)）と同じく単価×数量の 1 円未満切捨てで、行金額を確定する算定器
// （billing_service.go。1 円未満四捨五入）とは別の段階。既存の割引額の意味を変えないために維持しており、
// 両者の統一は割引式の exact 化（EMR-505 後続。inventory §2.D の「formula N later」）で扱う。
func discountBaseAmount(unitPrice, quantity model.ExactNumeric) (int64, error) {
	unit, unitOK := calcRat(unitPrice)
	qty, qtyOK := calcRat(quantity)
	if !unitOK || !qtyOK || unit.Sign() < 0 || qty.Sign() < 0 {
		return 0, apperrors.WrapCalculationUnsupported("単価または数量が有限の0以上の数値ではないため、割引額の基数を算定できません")
	}
	base, err := calcInt64(ratFloor(new(big.Rat).Mul(unit, qty)))
	if err != nil {
		return 0, apperrors.WrapCalculationUnsupported("割引額の基数が扱える範囲を超えています")
	}
	return base, nil
}

// discountPercentAmount は基数 × 割合(%) ÷ 100 を 1 円未満四捨五入した割引額（円）を返す。
// 旧 math.Round（非負は 0.5 切上げ）と同じ丸めを exact で行う。割合が NULL・非有限・0 以下なら 0。
// 結果が int64 を超える場合は基数を返す（呼出し側が基数で上限を掛けるため結果は同じ）。
func discountPercentAmount(base int64, percent model.ExactNumeric) int64 {
	rate, ok := calcRat(percent)
	if !ok || rate.Sign() <= 0 {
		return 0
	}
	amount := new(big.Rat).SetInt64(base)
	amount.Mul(amount, rate)
	amount.Quo(amount, big.NewRat(100, 1))
	rounded, err := calcInt64(ratRoundHalfUp(amount))
	if err != nil {
		return base
	}
	return rounded
}

// discountFixedAmount は固定額キャンペーンの割引額（円）。旧 math.Round と同じく 1 円未満四捨五入。
// 値が NULL・非有限・0 以下なら 0。int64 を超える場合は基数を返す（呼出し側が基数で上限を掛ける）。
func discountFixedAmount(base int64, value model.ExactNumeric) int64 {
	amount, ok := calcRat(value)
	if !ok || amount.Sign() <= 0 {
		return 0
	}
	rounded, err := calcInt64(ratRoundHalfUp(amount))
	if err != nil {
		return base
	}
	return rounded
}

// BuildDiscountSuggestions は明細小計・全適用キャンペーン・飼主割引率から選択肢を構築する。
// 同一 calculated_amount の重複は除去しない（名前が異なるキャンペーンは個別に表示）。
// Amount=0 の候補は除外する（割引なしは選択肢として意味がない）。
func BuildDiscountSuggestions(itemSubtotal int64, campaigns []*model.Campaign, ownerRate model.ExactNumeric) []DiscountSuggestion {
	var suggestions []DiscountSuggestion
	for _, c := range campaigns {
		if c == nil {
			continue
		}
		var amount int64
		switch c.DiscountType {
		case model.CampaignDiscountTypeRate:
			amount = discountPercentAmount(itemSubtotal, c.DiscountValue)
		case model.CampaignDiscountTypeAmount:
			amount = discountFixedAmount(itemSubtotal, c.DiscountValue)
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
	if ownerAmount := min(discountPercentAmount(itemSubtotal, ownerRate), itemSubtotal); ownerAmount > 0 {
		suggestions = append(suggestions, DiscountSuggestion{
			Type:          "owner",
			CampaignID:    nil,
			Name:          "飼主割引",
			DiscountType:  "rate",
			DiscountValue: ownerRate,
			Amount:        ownerAmount,
		})
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
func CalculateItemCampaignDiscount(itemSubtotal int64, campaign *model.Campaign, ownerDiscountRate model.ExactNumeric) int64 {
	amount, _ := ResolveItemDiscount(itemSubtotal, campaign, ownerDiscountRate)
	return amount
}

// ResolveItemDiscount は明細に適用する割引を（割引額[円], 適用割引率[%]）で返す。
// appliedRate は割引の出所を表す: 飼主割引採用時は飼主率、率方式キャンペーン採用時は
// キャンペーン率、額方式キャンペーン採用時・割引なし時は 0。
//
// EMR-229: 割引額だけでなく適用率を返し、billing_items.discount_rate に正しく記録できる
// ようにする（従来は自動割引時に rate が 0 のまま保存され、画面に割引率が出なかった）。
func ResolveItemDiscount(itemSubtotal int64, campaign *model.Campaign, ownerDiscountRate model.ExactNumeric) (int64, model.ExactNumeric) {
	noDiscount := model.ExactNumericFromInt64(0)
	if itemSubtotal <= 0 {
		return 0, noDiscount
	}

	var campaignDiscount int64
	campaignRate := noDiscount
	if campaign != nil {
		switch campaign.DiscountType {
		case model.CampaignDiscountTypeRate:
			campaignDiscount = discountPercentAmount(itemSubtotal, campaign.DiscountValue)
			campaignRate = campaign.DiscountValue
		case model.CampaignDiscountTypeAmount:
			campaignDiscount = discountFixedAmount(itemSubtotal, campaign.DiscountValue)
		}
	}

	ownerDiscount := discountPercentAmount(itemSubtotal, ownerDiscountRate)

	// Q3=B: 高い方を採用。明細小計を超えないようクランプ。
	if campaignDiscount > ownerDiscount {
		return min(campaignDiscount, itemSubtotal), campaignRate
	}
	return min(ownerDiscount, itemSubtotal), ownerDiscountRate
}
