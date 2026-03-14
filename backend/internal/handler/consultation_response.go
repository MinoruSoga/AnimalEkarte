package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type consultationResponse struct {
	ID            uint64    `json:"id"`
	ClinicID      uint64    `json:"clinic_id"`
	Name          string    `json:"name"`
	Price         *float64  `json:"price,omitempty"`
	IsActive      bool      `json:"is_active"`
	Description   string    `json:"description"`
	TimeCondition string    `json:"time_condition"`
	Duration      *int      `json:"duration,omitempty"`
	SortOrder     int       `json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toConsultationResponse(con *model.Consultation) consultationResponse {
	return consultationResponse{
		ID:            con.ID,
		ClinicID:      con.ClinicID,
		Name:          con.Name,
		Price:         con.Price,
		IsActive:      con.IsActive,
		Description:   con.Description,
		TimeCondition: con.TimeCondition,
		Duration:      con.Duration,
		SortOrder:     con.SortOrder,
		CreatedAt:     con.CreatedAt,
		UpdatedAt:     con.UpdatedAt,
	}
}

func toConsultationResponseList(items []model.Consultation) []consultationResponse {
	list := make([]consultationResponse, 0, len(items))
	for i := range items {
		list = append(list, toConsultationResponse(&items[i]))
	}
	return list
}
