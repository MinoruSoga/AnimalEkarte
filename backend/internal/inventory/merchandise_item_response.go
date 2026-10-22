package inventory

import (
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
)

// merchandiseTaxRateString は保存値を十進文字列として応答する（EMR-530、採択405・483の十進輸送）。
// float 経由の丸めを行わない。NULL・NaN・無限大・変換失敗は既知の異常として伝播し、
// 有限契約外の値を応答せず、正常な "0" を捏造して補完しない（stage2計画 §3C「未知/NULLを0補完しない」）。
func merchandiseTaxRateString(v model.ExactNumeric) (string, error) {
	if !v.Valid || v.NaN || v.InfinityModifier != pgtype.Finite {
		return "", apperrors.WrapInternalServerError("merchandise tax_rate の保存値が不正です（NULL / NaN / 無限大）")
	}
	value, err := v.Value()
	if err != nil || value == nil {
		return "", apperrors.WrapInternalServerError("merchandise tax_rate の保存値を文字列化できません")
	}
	text, _ := value.(string)
	return text, nil
}

type merchandiseItemResponse struct {
	ID        uint64             `json:"id"`
	ClinicID  uint64             `json:"clinic_id"`
	Name      string             `json:"name"`
	Category  model.ItemCategory `json:"category"`
	UnitPrice int64              `json:"unit_price"`
	TaxType   model.TaxType      `json:"tax_type"`
	TaxRate   string             `json:"tax_rate"`
	IsActive  bool               `json:"is_active"`
	SortOrder int                `json:"sort_order"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func toMerchandiseItemResponse(m *model.MerchandiseItem) (merchandiseItemResponse, error) {
	taxRate, err := merchandiseTaxRateString(m.TaxRate)
	if err != nil {
		return merchandiseItemResponse{}, err
	}
	return merchandiseItemResponse{
		ID:        m.ID,
		ClinicID:  m.ClinicID,
		Name:      m.Name,
		Category:  m.Category,
		UnitPrice: m.UnitPrice,
		TaxType:   m.TaxType,
		TaxRate:   taxRate,
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: httpapi.LocalTime(m.CreatedAt),
		UpdatedAt: httpapi.LocalTime(m.UpdatedAt),
	}, nil
}
