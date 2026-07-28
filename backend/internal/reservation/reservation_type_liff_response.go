package reservation

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type reservationTypeLiffResponse struct {
	ID                   uint64    `json:"id"`
	ClinicID             uint64    `json:"clinic_id"`
	Name                 string    `json:"name"`
	Color                string    `json:"color"`
	Category             string    `json:"category"`
	IsActive             bool      `json:"is_active"`
	Description          string    `json:"description"`
	SortOrder            int       `json:"sort_order"`
	ParentID             *uint64   `json:"parent_id,omitempty"`
	DurationMinutes      int       `json:"duration_minutes"`
	MaxConcurrent        *int      `json:"max_concurrent,omitempty"`
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
		Category:             string(st.Category),
		IsActive:             st.IsActive,
		Description:          st.Description,
		SortOrder:            st.SortOrder,
		ParentID:             st.ParentID,
		DurationMinutes:      st.DurationMinutes,
		MaxConcurrent:        st.MaxConcurrent,
		ShortName:            st.ShortName,
		ShowShortName:        st.ShowShortName,
		ReservationVisible:   st.ReservationVisible,
		ReservationComment:   st.ReservationComment,
		ReservationImageURL:  st.ReservationImageURL,
		ReservationDayOption: string(st.ReservationDayOption),
		IsInternal:           st.IsInternal,
		CreatedAt:            httpapi.LocalTime(st.CreatedAt),
		UpdatedAt:            httpapi.LocalTime(st.UpdatedAt),
	}
}
