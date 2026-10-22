package inventory

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// merchandiseTaxRateString は保存値を十進文字列として応答する（EMR-530、採択405・483の十進輸送）。
// float 経由の丸めを行わない。列は NOT NULL であり到達しないが、異常時は既定 0 を返す。
func merchandiseTaxRateString(v model.ExactNumeric) string {
	value, err := v.Value()
	if err != nil || value == nil {
		return "0"
	}
	text, _ := value.(string)
	return text
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

func toMerchandiseItemResponse(m *model.MerchandiseItem) merchandiseItemResponse {
	return merchandiseItemResponse{
		ID:        m.ID,
		ClinicID:  m.ClinicID,
		Name:      m.Name,
		Category:  m.Category,
		UnitPrice: m.UnitPrice,
		TaxType:   m.TaxType,
		TaxRate:   merchandiseTaxRateString(m.TaxRate),
		IsActive:  m.IsActive,
		SortOrder: m.SortOrder,
		CreatedAt: httpapi.LocalTime(m.CreatedAt),
		UpdatedAt: httpapi.LocalTime(m.UpdatedAt),
	}
}
