package handler

type createServiceTypeRequest struct {
	Name        string `json:"name"        binding:"required"`
	Color       string `json:"color"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`

	// LINE予約用フィールド
	ReservationDisplayName string `json:"reservation_display_name"`
	DurationMinutes        *int   `json:"duration_minutes"`
	ShortName              string `json:"short_name"`
	ShowShortName          bool   `json:"show_short_name"`
	ReservationVisible     *bool  `json:"reservation_visible"`
	ReservationComment     string `json:"reservation_comment"`
	ReservationDayOption   string `json:"reservation_day_option"`
	IsInternal             bool   `json:"is_internal"`
}

type updateServiceTypeRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`

	// LINE予約用フィールド
	ReservationDisplayName *string `json:"reservation_display_name"`
	DurationMinutes        *int    `json:"duration_minutes"`
	ShortName              *string `json:"short_name"`
	ShowShortName          *bool   `json:"show_short_name"`
	ReservationVisible     *bool   `json:"reservation_visible"`
	ReservationComment     *string `json:"reservation_comment"`
	ReservationDayOption   *string `json:"reservation_day_option"`
	IsInternal             *bool   `json:"is_internal"`
}

type reorderServiceTypeRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
