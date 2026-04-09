package handler

type createReservationCourseRequest struct {
	Name                 string `json:"name"                   binding:"required"`
	Color                string `json:"color"`
	Description          string `json:"description"`
	SortOrder            int    `json:"sort_order"`
	DurationMinutes      int    `json:"duration_minutes"`
	ShortName            string `json:"short_name"`
	ShowShortName        bool   `json:"show_short_name"`
	ReservationVisible   bool   `json:"reservation_visible"`
	ReservationComment   string `json:"reservation_comment"`
	ReservationDayOption string `json:"reservation_day_option"`
	IsInternal           bool   `json:"is_internal"`
}

type updateReservationCourseRequest struct {
	Name                 *string `json:"name"`
	Color                *string `json:"color"`
	Description          *string `json:"description"`
	SortOrder            *int    `json:"sort_order"`
	DurationMinutes      *int    `json:"duration_minutes"`
	ShortName            *string `json:"short_name"`
	ShowShortName        *bool   `json:"show_short_name"`
	ReservationVisible   *bool   `json:"reservation_visible"`
	ReservationComment   *string `json:"reservation_comment"`
	ReservationDayOption *string `json:"reservation_day_option"`
	IsInternal           *bool   `json:"is_internal"`
}

type patchReservationCourseStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type patchReservationCourseSortOrderRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down"`
}
