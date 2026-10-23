package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/model"
)

type procedureResponse struct {
	Version     int64                 `json:"version"`
	ID          uint64                `json:"id"`
	ClinicID    uint64                `json:"clinic_id"`
	Name        string                `json:"name"`
	Price       *int64                `json:"price,omitempty"`
	IsActive    bool                  `json:"is_active"`
	Description string                `json:"description"`
	Duration    *int                  `json:"duration,omitempty"`
	Anesthesia  *model.AnesthesiaType `json:"anesthesia"`
	ParentID    *uint64               `json:"parent_id,omitempty"`
	TaxType     string                `json:"tax_type"`
	TaxRate     model.ExactNumeric    `json:"tax_rate"`
	SortOrder   *int                  `json:"sort_order"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`

	// EMR-454: 品目別保険設定。true=保険対象外 / false=保険対象 / null=未保存・旧欠落。
	IsNonInsurance *bool `json:"is_non_insurance"`

	// EMR-447 A01: 全体割引除外。null=未保存・旧欠落を raw のまま返し false に書換えない。
	IsBillDiscountExcluded *bool `json:"is_bill_discount_excluded"`
}

func toProcedureResponse(p *model.Procedure) procedureResponse {
	return procedureResponse{
		Version:     p.Version,
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		Name:        p.Name,
		Price:       p.Price,
		IsActive:    p.IsActive,
		Description: p.Description,
		Duration:    p.Duration,
		Anesthesia:  p.Anesthesia,
		ParentID:    p.ParentID,
		TaxType:     string(p.TaxType),
		TaxRate:     p.TaxRate,
		SortOrder:   p.SortOrder,
		CreatedAt:   httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:   httpapi.LocalTime(p.UpdatedAt),

		IsNonInsurance: p.IsNonInsurance,

		IsBillDiscountExcluded: p.IsBillDiscountExcluded,
	}
}
