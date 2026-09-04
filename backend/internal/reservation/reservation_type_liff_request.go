package reservation

type createReservationTypeLiffRequest struct {
	Name            string `json:"name"                   binding:"required"`
	Color           string `json:"color"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
	DurationMinutes int    `json:"duration_minutes"`
	MaxConcurrent   *int   `json:"max_concurrent"`
	ShortName       string `json:"short_name"`
	ShowShortName   bool   `json:"show_short_name"`
	// ReservationVisible is *bool so omitted stays the LIFF default (true).
	// Explicit false is preserved; this avoids silently flipping omit→false after compensation.
	ReservationVisible   *bool  `json:"reservation_visible"`
	ReservationComment   string `json:"reservation_comment"`
	ReservationDayOption string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal           bool   `json:"is_internal"`
}

func (r *createReservationTypeLiffRequest) toServiceInput() *CreateReservationTypeLiffInput {
	return &CreateReservationTypeLiffInput{
		Name:                 r.Name,
		Color:                r.Color,
		Description:          r.Description,
		SortOrder:            r.SortOrder,
		DurationMinutes:      r.DurationMinutes,
		MaxConcurrent:        r.MaxConcurrent,
		ShortName:            r.ShortName,
		ShowShortName:        r.ShowShortName,
		ReservationVisible:   resolveBoolDefaultTrue(r.ReservationVisible),
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
	MaxConcurrent        *int    `json:"max_concurrent"`
	ClearMaxConcurrent   bool    `json:"clear_max_concurrent"`
	ShortName            *string `json:"short_name"`
	ShowShortName        *bool   `json:"show_short_name"`
	ReservationVisible   *bool   `json:"reservation_visible"`
	ReservationComment   *string `json:"reservation_comment"`
	ReservationDayOption *string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal           *bool   `json:"is_internal"`
}

func (r *updateReservationTypeLiffRequest) toServiceInput() *UpdateReservationTypeLiffInput {
	return &UpdateReservationTypeLiffInput{
		Name:                 r.Name,
		Color:                r.Color,
		Description:          r.Description,
		SortOrder:            r.SortOrder,
		DurationMinutes:      r.DurationMinutes,
		MaxConcurrent:        r.MaxConcurrent,
		ClearMaxConcurrent:   r.ClearMaxConcurrent,
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
