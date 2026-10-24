package reservation

import (
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

type createReservationTypeRequest struct {
	Name  string `json:"name"        binding:"required"`
	Color string `json:"color"`
	// IsActive is *bool so JSON binding can distinguish omitted / false / true.
	// Omitted (nil) resolves to true in toServiceInput.
	IsActive    *bool   `json:"is_active"`
	Description string  `json:"description"`
	SortOrder   int     `json:"sort_order"`
	GroupID     *uint64 `json:"group_id"`
	ParentID    *uint64 `json:"parent_id"`
	Category    string  `json:"category"    binding:"omitempty,oneof=general trimming"`

	// LINE予約用フィールド
	ReservationDisplayName string `json:"reservation_display_name"`
	DurationMinutes        *int   `json:"duration_minutes"`
	MaxConcurrent          *int   `json:"max_concurrent"`
	ShortName              string `json:"short_name"`
	ShowShortName          bool   `json:"show_short_name"`
	ReservationVisible     *bool  `json:"reservation_visible"`
	ReservationComment     string `json:"reservation_comment"`
	ReservationImageURL    string `json:"reservation_image_url"`
	ReservationDayOption   string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal             bool   `json:"is_internal"`
}

func (r *createReservationTypeRequest) toServiceInput() *CreateReservationTypeInput {
	return &CreateReservationTypeInput{
		Name:                   r.Name,
		IsActive:               resolveBoolDefaultTrue(r.IsActive),
		Description:            r.Description,
		SortOrder:              r.SortOrder,
		GroupID:                r.GroupID,
		ParentID:               r.ParentID,
		ReservationDisplayName: r.ReservationDisplayName,
		DurationMinutes:        r.DurationMinutes,
		MaxConcurrent:          r.MaxConcurrent,
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
	Name          *string `json:"name"`
	Color         *string `json:"color"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	SortOrder     *int    `json:"sort_order"`
	GroupID       *uint64 `json:"group_id"`
	ClearGroupID  bool    `json:"clear_group_id"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	Category      *string `json:"category"    binding:"omitempty,oneof=general trimming"`

	// LINE予約用フィールド
	ReservationDisplayName *string `json:"reservation_display_name"`
	DurationMinutes        *int    `json:"duration_minutes"`
	MaxConcurrent          *int    `json:"max_concurrent"`
	ClearMaxConcurrent     bool    `json:"clear_max_concurrent"`
	ShortName              *string `json:"short_name"`
	ShowShortName          *bool   `json:"show_short_name"`
	ReservationVisible     *bool   `json:"reservation_visible"`
	ReservationComment     *string `json:"reservation_comment"`
	ReservationImageURL    *string `json:"reservation_image_url"`
	ReservationDayOption   *string `json:"reservation_day_option" binding:"omitempty,oneof=none saturday weekday anyday"`
	IsInternal             *bool   `json:"is_internal"`
}

func (r *updateReservationTypeRequest) toServiceInput() *UpdateReservationTypeInput {
	return &UpdateReservationTypeInput{
		Name:                   r.Name,
		IsActive:               r.IsActive,
		Description:            r.Description,
		SortOrder:              r.SortOrder,
		GroupID:                r.GroupID,
		ClearGroupID:           r.ClearGroupID,
		ParentID:               r.ParentID,
		ClearParentID:          r.ClearParentID,
		ReservationDisplayName: r.ReservationDisplayName,
		DurationMinutes:        r.DurationMinutes,
		MaxConcurrent:          r.MaxConcurrent,
		ClearMaxConcurrent:     r.ClearMaxConcurrent,
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
	DayOfWeek       *int16  `json:"day_of_week"`
	SpecificDate    *string `json:"specific_date"` // "YYYY-MM-DD"
	StartTime       string  `json:"start_time"     binding:"required"`
	EndTime         string  `json:"end_time"       binding:"required"`
}

func (r createUnavailableTimeRequest) toServiceInput() (CreateUnavailableTimeInput, error) {
	input := CreateUnavailableTimeInput{
		UnavailableType: r.UnavailableType,
		DayOfWeek:       r.DayOfWeek,
		StartTime:       r.StartTime,
		EndTime:         r.EndTime,
	}
	if r.SpecificDate != nil {
		specificDate, err := httpapi.ParseDate(r.SpecificDate)
		if err != nil {
			return CreateUnavailableTimeInput{}, err
		}
		input.SpecificDate = specificDate
	}
	return input, nil
}

// CreateAvailableSlotRequest は予約可能開始時刻の作成リクエスト
type createAvailableSlotRequest struct {
	AvailableType string  `json:"available_type" binding:"required,oneof=weekly specific"`
	DayOfWeek     *int16  `json:"day_of_week"`
	SpecificDate  *string `json:"specific_date"` // "YYYY-MM-DD"
	StartTime     string  `json:"start_time"     binding:"required"`
	IsActive      *bool   `json:"is_active"`
}

func (r createAvailableSlotRequest) toServiceInput() (CreateAvailableSlotInput, error) {
	input := CreateAvailableSlotInput{
		AvailableType: r.AvailableType,
		DayOfWeek:     r.DayOfWeek,
		StartTime:     r.StartTime,
		IsActive:      r.IsActive,
	}
	if r.SpecificDate != nil {
		specificDate, err := httpapi.ParseDate(r.SpecificDate)
		if err != nil {
			return CreateAvailableSlotInput{}, err
		}
		input.SpecificDate = specificDate
	}
	return input, nil
}

// retiredReservationTypeFieldError は廃止された color / category の書込を拒否する（400）。
// color は予約区分グループの色、category は business_kind に置き換わり（001_init.sql L10878 で列を DROP 済み）、
// 黙って捨てると利用者が保存済みと思い込む値を失う。null / 空は未送信と同じ扱いにする。
func retiredReservationTypeFieldError(colorSent, categorySent bool) error {
	switch {
	case colorSent:
		return apperrors.WrapInvalidInput("color は廃止されました。予約区分の色は予約区分グループの色を使用します")
	case categorySent:
		return apperrors.WrapInvalidInput("category は廃止されました。業務区分は business_kind で管理します")
	}
	return nil
}

func (r *createReservationTypeRequest) retiredFieldError() error {
	return retiredReservationTypeFieldError(r.Color != "", r.Category != "")
}

func (r *updateReservationTypeRequest) retiredFieldError() error {
	return retiredReservationTypeFieldError(r.Color != nil, r.Category != nil)
}
