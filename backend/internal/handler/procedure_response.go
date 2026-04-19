package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type procedureResponse struct {
	ID          uint64               `json:"id"`
	ClinicID    uint64               `json:"clinic_id"`
	Name        string               `json:"name"`
	Price       *int64               `json:"price,omitempty"`
	IsActive    bool                 `json:"is_active"`
	Description string               `json:"description"`
	Duration    *int                 `json:"duration,omitempty"`
	Anesthesia  model.AnesthesiaType `json:"anesthesia"`
	ParentID    *uint64              `json:"parent_id,omitempty"`
	TaxType     model.TaxType        `json:"tax_type"`
	TaxRate     float64              `json:"tax_rate"`
	SortOrder   int                  `json:"sort_order"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func toProcedureResponse(p *model.Procedure) procedureResponse {
	return procedureResponse{
		ID:          p.ID,
		ClinicID:    p.ClinicID,
		Name:        p.Name,
		Price:       p.Price,
		IsActive:    p.IsActive,
		Description: p.Description,
		Duration:    p.Duration,
		Anesthesia:  p.Anesthesia,
		ParentID:    p.ParentID,
		TaxType:     p.TaxType,
		TaxRate:     p.TaxRate,
		SortOrder:   p.SortOrder,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
