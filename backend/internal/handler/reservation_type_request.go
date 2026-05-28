package handler

import "github.com/animal-ekarte/backend/internal/service"

type createReservationTypeRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Color       string  `json:"color"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	SortOrder   int     `json:"sort_order"`
	GroupID     *uint64 `json:"group_id"`
	Category    string  `json:"category"    binding:"omitempty,oneof=general trimming"`

	// LINE予約用フィールド
	ReservationDisplayName string `json:"reservation_display_name"`
	DurationMinutes        *int   `json:"duration_minutes"`
	ShortName              string `json:"short_name"`
	ShowShortName          bool   `json:"show_short_name"`
	ReservationVisible     *bool  `json:"reservation_visible"`
	ReservationComment     string `json:"reservation_comment"`
	ReservationImageURL    string `json:"reservation_image_url"`
	ReservationDayOption   string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal             bool   `json:"is_internal"`
}

func (r createReservationTypeRequest) toServiceInput() *service.CreateReservationTypeInput {
	return &service.CreateReservationTypeInput{
		Name:                   r.Name,
		Color:                  r.Color,
		IsActive:               r.IsActive,
		Description:            r.Description,
		SortOrder:              r.SortOrder,
		GroupID:                r.GroupID,
		Category:               r.Category,
		ReservationDisplayName: r.ReservationDisplayName,
		DurationMinutes:        r.DurationMinutes,
		ShortName:              r.ShortName,
		ShowShortName:          r.ShowShortName,
		ReservationVisible:     r.ReservationVisible,
		ReservationComment:     r.ReservationComment,
		ReservationImageURL:    r.ReservationImageURL,
		ReservationDayOption:   r.ReservationDayOption,
		IsInternal:             r.IsInternal,
	}
}

type updateReservationTypeRequest struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	SortOrder    *int    `json:"sort_order"`
	GroupID      *uint64 `json:"group_id"`
	ClearGroupID bool    `json:"clear_group_id"`
	Category     *string `json:"category"    binding:"omitempty,oneof=general trimming"`

	// LINE予約用フィールド
	ReservationDisplayName *string `json:"reservation_display_name"`
	DurationMinutes        *int    `json:"duration_minutes"`
	ShortName              *string `json:"short_name"`
	ShowShortName          *bool   `json:"show_short_name"`
	ReservationVisible     *bool   `json:"reservation_visible"`
	ReservationComment     *string `json:"reservation_comment"`
	ReservationImageURL    *string `json:"reservation_image_url"`
	ReservationDayOption   *string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal             *bool   `json:"is_internal"`
}

func (r updateReservationTypeRequest) toServiceInput() *service.UpdateReservationTypeInput {
	return &service.UpdateReservationTypeInput{
		Name:                   r.Name,
		Color:                  r.Color,
		IsActive:               r.IsActive,
		Description:            r.Description,
		SortOrder:              r.SortOrder,
		GroupID:                r.GroupID,
		ClearGroupID:           r.ClearGroupID,
		Category:               r.Category,
		ReservationDisplayName: r.ReservationDisplayName,
		DurationMinutes:        r.DurationMinutes,
		ShortName:              r.ShortName,
		ShowShortName:          r.ShowShortName,
		ReservationVisible:     r.ReservationVisible,
		ReservationComment:     r.ReservationComment,
		ReservationImageURL:    r.ReservationImageURL,
		ReservationDayOption:   r.ReservationDayOption,
		IsInternal:             r.IsInternal,
	}
}

// CreateUnavailableTimeRequest は予約不可時間の作成リクエスト
type createUnavailableTimeRequest struct {
	UnavailableType string  `json:"unavailable_type" binding:"required,oneof=weekly specific"`
	DayOfWeek       *int8   `json:"day_of_week"`
	SpecificDate    *string `json:"specific_date"` // "YYYY-MM-DD"
	StartTime       string  `json:"start_time"     binding:"required"`
	EndTime         string  `json:"end_time"       binding:"required"`
}

func (r createUnavailableTimeRequest) toServiceInput() (service.CreateUnavailableTimeInput, error) {
	input := service.CreateUnavailableTimeInput{
		UnavailableType: r.UnavailableType,
		DayOfWeek:       r.DayOfWeek,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
	}
	if r.SpecificDate != nil {
		specificDate, err := parseDate(r.SpecificDate)
		if err != nil {
			return service.CreateUnavailableTimeInput{}, err
		}
		input.SpecificDate = specificDate
	}
	return input, nil
}

// LinkOccupationRequest は職種紐付けリクエスト
type linkOccupationRequest struct {
	OccupationID uint64 `json:"occupation_id" binding:"required"`
}
