package billing

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// campaignResponse は割引キャンペーンのレスポンス (#81)
type campaignResponse struct {
	ID               uint64   `json:"id"`
	ClinicID         uint64   `json:"clinic_id"`
	Name             string   `json:"name"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	DiscountType     string   `json:"discount_type"`
	DiscountValue    string   `json:"discount_value"`
	IsActive         bool     `json:"is_active"`
	SortOrder        int      `json:"sort_order"`
	TargetCategories []string `json:"target_categories"`
	TargetItemIDs    []uint64 `json:"target_item_ids"`
}

// exactNumericString は保存値を十進文字列として応答する（EMR-530、採択483の十進輸送）。
// float 経由の丸めを行わない。NULL・NaN・無限大・変換失敗は既知の異常として伝播し、
// 有限契約外の値を応答せず、正常な "0" を捏造して補完しない（stage2計画 §3C「未知/NULLを0補完しない」）。
func exactNumericString(v model.ExactNumeric) (string, error) {
	if !v.Valid || v.NaN || v.InfinityModifier != pgtype.Finite {
		return "", apperrors.WrapInternalServerError("campaign discount_value の保存値が不正です（NULL / NaN / 無限大）")
	}
	value, err := v.Value()
	if err != nil || value == nil {
		return "", apperrors.WrapInternalServerError("campaign discount_value の保存値を文字列化できません")
	}
	text, _ := value.(string)
	return text, nil
}

func toCampaignResponse(m *model.Campaign) (campaignResponse, error) {
	cats := make([]string, 0, len(m.TargetCategories))
	for _, tc := range m.TargetCategories {
		cats = append(cats, string(tc.Category))
	}
	ids := make([]uint64, 0, len(m.TargetItems))
	for _, ti := range m.TargetItems {
		ids = append(ids, ti.MerchandiseItemID)
	}
	discountValue, err := exactNumericString(m.DiscountValue)
	if err != nil {
		return campaignResponse{}, err
	}
	return campaignResponse{
		ID:               m.ID,
		ClinicID:         m.ClinicID,
		Name:             m.Name,
		StartDate:        m.StartDate.In(time.Local).Format(time.DateOnly),
		EndDate:          m.EndDate.In(time.Local).Format(time.DateOnly),
		DiscountType:     string(m.DiscountType),
		DiscountValue:    discountValue,
		IsActive:         m.IsActive,
		SortOrder:        m.SortOrder,
		TargetCategories: cats,
		TargetItemIDs:    ids,
	}, nil
}
