package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationTypeLiffResponse struct {
	ID                   uint64    `json:"id"`
	ClinicID             uint64    `json:"clinic_id"`
	Name                 string    `json:"name"`
	Color                string    `json:"color"`
	IsActive             bool      `json:"is_active"`
	Description          string    `json:"description"`
	SortOrder            int       `json:"sort_order"`
	DurationMinutes      int       `json:"duration_minutes"`
	ShortName            string    `json:"short_name"`
	ShowShortName        bool      `json:"show_short_name"`
	ReservationVisible   bool      `json:"reservation_visible"`
	ReservationComment   string    `json:"reservation_comment"`
	ReservationImageURL  string    `json:"reservation_image_url"`
	ReservationDayOption string    `json:"reservation_day_option"`
	IsInternal           bool      `json:"is_internal"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func toReservationTypeLiffResponse(st *model.ReservationType) reservationTypeLiffResponse {
	return reservationTypeLiffResponse{
		ID:                   st.ID,
		ClinicID:             st.ClinicID,
		Name:                 st.Name,
		Color:                st.Color,
		IsActive:             st.IsActive,
		Description:          st.Description,
		SortOrder:            st.SortOrder,
		DurationMinutes:      st.DurationMinutes,
		ShortName:            st.ShortName,
		ShowShortName:        st.ShowShortName,
		ReservationVisible:   st.ReservationVisible,
		ReservationComment:   st.ReservationComment,
		ReservationImageURL:  st.ReservationImageURL,
		ReservationDayOption: string(st.ReservationDayOption),
		IsInternal:           st.IsInternal,
		CreatedAt:            st.CreatedAt,
		UpdatedAt:            st.UpdatedAt,
	}
}

func toReservationTypeLiffResponseList(items []model.ReservationType) []reservationTypeLiffResponse {
	list := make([]reservationTypeLiffResponse, 0, len(items))
	for i := range items {
		list = append(list, toReservationTypeLiffResponse(&items[i]))
	}
	return list
}
