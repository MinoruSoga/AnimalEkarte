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

	// LINE予約用フィールド
	DurationMinutes      int                        `json:"duration_minutes"`
	ShortName            string                     `json:"short_name"`
	ShowShortName        bool                       `json:"show_short_name"`
	ReservationVisible   bool                       `json:"reservation_visible"`
	ReservationComment   string                     `json:"reservation_comment"`
	ReservationImageURL  string                     `json:"reservation_image_url"`
	ReservationDayOption model.ReservationDayOption `json:"reservation_day_option"`
	IsInternal           bool                       `json:"is_internal"`
}

func toServiceTypeResponse(st *model.ServiceType) serviceTypeResponse {
	return serviceTypeResponse{
		ID:                   st.ID,
		ClinicID:             st.ClinicID,
		Name:                 st.Name,
		Color:                st.Color,
		IsActive:             st.IsActive,
		Description:          st.Description,
		SortOrder:            st.SortOrder,
		CreatedAt:            st.CreatedAt,
		UpdatedAt:            st.UpdatedAt,
		DurationMinutes:      st.DurationMinutes,
		ShortName:            st.ShortName,
		ShowShortName:        st.ShowShortName,
		ReservationVisible:   st.ReservationVisible,
		ReservationComment:   st.ReservationComment,
		ReservationImageURL:  st.ReservationImageURL,
		ReservationDayOption: st.ReservationDayOption,
		IsInternal:           st.IsInternal,
	}
}

func toServiceTypeResponseList(items []model.ServiceType) []serviceTypeResponse {
	list := make([]serviceTypeResponse, 0, len(items))
	for i := range items {
		list = append(list, toServiceTypeResponse(&items[i]))
	}
	return list
}
