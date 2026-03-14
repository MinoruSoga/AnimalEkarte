package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type procedureResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	Duration    *int      `json:"duration,omitempty"`
	Anesthesia  string    `json:"anesthesia"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
		Anesthesia:  string(p.Anesthesia),
		SortOrder:   p.SortOrder,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProcedureResponseList(procedures []model.Procedure) []procedureResponse {
	list := make([]procedureResponse, 0, len(procedures))
	for i := range procedures {
		list = append(list, toProcedureResponse(&procedures[i]))
	}
	return list
}
