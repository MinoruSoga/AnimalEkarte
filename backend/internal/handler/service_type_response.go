package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type serviceTypeResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toServiceTypeResponse(st *model.ServiceType) serviceTypeResponse {
	return serviceTypeResponse{
		ID:          st.ID,
		ClinicID:    st.ClinicID,
		Name:        st.Name,
		Color:       st.Color,
		IsActive:    st.IsActive,
		Description: st.Description,
		SortOrder:   st.SortOrder,
		CreatedAt:   st.CreatedAt,
		UpdatedAt:   st.UpdatedAt,
	}
}

func toServiceTypeResponseList(items []model.ServiceType) []serviceTypeResponse {
	list := make([]serviceTypeResponse, 0, len(items))
	for i := range items {
		list = append(list, toServiceTypeResponse(&items[i]))
	}
	return list
}
