package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"

	"github.com/animal-ekarte/backend/internal/model"
)

type consultationResponse struct {
	ID            uint64             `json:"id"`
	ClinicID      uint64             `json:"clinic_id"`
	Name          string             `json:"name"`
	Price         *int64             `json:"price,omitempty"`
	IsActive      bool               `json:"is_active"`
	Description   string             `json:"description"`
	TimeCondition string             `json:"time_condition"`
	Duration      *int               `json:"duration,omitempty"`
	ParentID      *uint64            `json:"parent_id,omitempty"`
	TaxType       model.TaxType      `json:"tax_type"`
	TaxRate       model.ExactNumeric `json:"tax_rate"`
	SortOrder     *int               `json:"sort_order"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`

	// EMR-454: 品目別保険設定。true=保険対象外 / false=保険対象 / null=未保存・旧欠落。
	IsNonInsurance *bool `json:"is_non_insurance"`
}

func toConsultationResponse(c *model.Consultation) consultationResponse {
	return consultationResponse{
		ID:            c.ID,
		ClinicID:      c.ClinicID,
		Name:          c.Name,
		Price:         c.Price,
		IsActive:      c.IsActive,
		Description:   c.Description,
		TimeCondition: c.TimeCondition,
		Duration:      c.Duration,
		ParentID:      c.ParentID,
		TaxType:       c.TaxType,
		TaxRate:       c.TaxRate,
		SortOrder:     c.SortOrder,
		CreatedAt:     httpapi.LocalTime(c.CreatedAt),
		UpdatedAt:     httpapi.LocalTime(c.UpdatedAt),

		IsNonInsurance: c.IsNonInsurance,
	}
}
