package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type reservationCategoryResponse struct {
	ID          uint64        `json:"id"`
	ClinicID    uint64        `json:"clinic_id"`
	Name        string        `json:"name"`
	Color       string        `json:"color"`
	IsActive    bool          `json:"is_active"`
	Description string        `json:"description"`
	SortOrder   int           `json:"sort_order"`
	GroupID     *uint64       `json:"group_id"`
	Group       *groupSummary `json:"group,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// LINE予約用フィールド
	ReservationDisplayName string                     `json:"reservation_display_name"`
	DurationMinutes        int                        `json:"duration_minutes"`
	ShortName              string                     `json:"short_name"`
	ShowShortName          bool                       `json:"show_short_name"`
	ReservationVisible     bool                       `json:"reservation_visible"`
	ReservationComment     string                     `json:"reservation_comment"`
	ReservationImageURL    string                     `json:"reservation_image_url"`
	ReservationDayOption   model.ReservationDayOption `json:"reservation_day_option"`
	IsInternal             bool                       `json:"is_internal"`
}

type groupSummary struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func toReservationTypeResponse(st *model.ReservationType) reservationCategoryResponse {
	resp := reservationCategoryResponse{
		ID:                     st.ID,
		ClinicID:               st.ClinicID,
		Name:                   st.Name,
		Color:                  st.Color,
		IsActive:               st.IsActive,
		Description:            st.Description,
		SortOrder:              st.SortOrder,
		GroupID:                st.GroupID,
		CreatedAt:              st.CreatedAt,
		UpdatedAt:              st.UpdatedAt,
		ReservationDisplayName: st.ReservationDisplayName,
		DurationMinutes:        st.DurationMinutes,
		ShortName:              st.ShortName,
		ShowShortName:          st.ShowShortName,
		ReservationVisible:     st.ReservationVisible,
		ReservationComment:     st.ReservationComment,
		ReservationImageURL:    st.ReservationImageURL,
		ReservationDayOption:   st.ReservationDayOption,
		IsInternal:             st.IsInternal,
	}
	if st.Group != nil {
		resp.Group = &groupSummary{
			ID:    st.Group.ID,
			Name:  st.Group.Name,
			Color: st.Group.Color,
		}
	}
	return resp
}

func toReservationTypeResponseList(items []model.ReservationType) []reservationCategoryResponse {
	list := make([]reservationCategoryResponse, 0, len(items))
	for i := range items {
		list = append(list, toReservationTypeResponse(&items[i]))
	}
	return list
}
