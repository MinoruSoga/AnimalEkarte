package handler

import "github.com/animal-ekarte/backend/internal/service"

type createReservationTypeLiffRequest struct {
	Name                 string `json:"name"                   binding:"required"`
	Color                string `json:"color"`
	Description          string `json:"description"`
	SortOrder            int    `json:"sort_order"`
	DurationMinutes      int    `json:"duration_minutes"`
	ShortName            string `json:"short_name"`
	ShowShortName        bool   `json:"show_short_name"`
	ReservationVisible   bool   `json:"reservation_visible"`
	ReservationComment   string `json:"reservation_comment"`
	ReservationDayOption string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal           bool   `json:"is_internal"`
}

func (r *createReservationTypeLiffRequest) toServiceInput() *service.CreateReservationTypeLiffInput {
	return &service.CreateReservationTypeLiffInput{
		Name:                 r.Name,
		Color:                r.Color,
		Description:          r.Description,
		SortOrder:            r.SortOrder,
		DurationMinutes:      r.DurationMinutes,
		ShortName:            r.ShortName,
		ShowShortName:        r.ShowShortName,
		ReservationVisible:   r.ReservationVisible,
		ReservationComment:   r.ReservationComment,
		ReservationDayOption: r.ReservationDayOption,
		IsInternal:           r.IsInternal,
	}
}

type updateReservationTypeLiffRequest struct {
	Name                 *string `json:"name"`
	Color                *string `json:"color"`
	Description          *string `json:"description"`
	SortOrder            *int    `json:"sort_order"`
	DurationMinutes      *int    `json:"duration_minutes"`
	ShortName            *string `json:"short_name"`
	ShowShortName        *bool   `json:"show_short_name"`
	ReservationVisible   *bool   `json:"reservation_visible"`
	ReservationComment   *string `json:"reservation_comment"`
	ReservationDayOption *string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal           *bool   `json:"is_internal"`
}

func (r *updateReservationTypeLiffRequest) toServiceInput() *service.UpdateReservationTypeLiffInput {
	return &service.UpdateReservationTypeLiffInput{
		Name:                 r.Name,
		Color:                r.Color,
		Description:          r.Description,
		SortOrder:            r.SortOrder,
		DurationMinutes:      r.DurationMinutes,
		ShortName:            r.ShortName,
		ShowShortName:        r.ShowShortName,
		ReservationVisible:   r.ReservationVisible,
		ReservationComment:   r.ReservationComment,
		ReservationDayOption: r.ReservationDayOption,
		IsInternal:           r.IsInternal,
	}
}

type patchReservationTypeLiffStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type patchReservationTypeLiffSortOrderRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}
