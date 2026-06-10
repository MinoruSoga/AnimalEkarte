package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- 予約不可時間レスポンス ----

type unavailableTimeResponse struct {
	ID                uint64  `json:"id"`
	ClinicID          uint64  `json:"clinic_id"`
	ReservationTypeID uint64  `json:"reservation_type_id"`
	UnavailableType   string  `json:"unavailable_type"`
	DayOfWeek         *int8   `json:"day_of_week,omitempty"`
	SpecificDate      *string `json:"specific_date,omitempty"` // "YYYY-MM-DD"
	StartTime         string  `json:"start_time"`
	EndTime           string  `json:"end_time"`
}

func toUnavailableTimeResponse(t *model.ReservationTypeUnavailableTime) unavailableTimeResponse {
	resp := unavailableTimeResponse{
		ID:                t.ID,
		ClinicID:          t.ClinicID,
		ReservationTypeID: t.ReservationTypeID,
		UnavailableType:   string(t.UnavailableType),
		DayOfWeek:         t.DayOfWeek,
		StartTime:         t.StartTime,
		EndTime:           t.EndTime,
	}
	if t.SpecificDate != nil {
		s := t.SpecificDate.In(time.Local).Format("2006-01-02")
		resp.SpecificDate = &s
	}
	return resp
}

// ---- 予約可能開始時刻レスポンス ----

type availableSlotResponse struct {
	ID                uint64  `json:"id"`
	ClinicID          uint64  `json:"clinic_id"`
	ReservationTypeID uint64  `json:"reservation_type_id"`
	AvailableType     string  `json:"available_type"`
	DayOfWeek         *int8   `json:"day_of_week,omitempty"`
	SpecificDate      *string `json:"specific_date,omitempty"` // "YYYY-MM-DD"
	StartTime         string  `json:"start_time"`
	IsActive          bool    `json:"is_active"`
}

func toAvailableSlotResponse(slot *model.ReservationTypeAvailableSlot) availableSlotResponse {
	resp := availableSlotResponse{
		ID:                slot.ID,
		ClinicID:          slot.ClinicID,
		ReservationTypeID: slot.ReservationTypeID,
		AvailableType:     string(slot.AvailableType),
		DayOfWeek:         slot.DayOfWeek,
		StartTime:         slot.StartTime,
		IsActive:          slot.IsActive,
	}
	if slot.SpecificDate != nil {
		s := slot.SpecificDate.In(time.Local).Format("2006-01-02")
		resp.SpecificDate = &s
	}
	return resp
}

// ---- 職種紐付けレスポンス ----

type occupationSummary struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type reservationTypeOccupationResponse struct {
	ID                uint64             `json:"id"`
	ClinicID          uint64             `json:"clinic_id"`
	ReservationTypeID uint64             `json:"reservation_type_id"`
	OccupationID      uint64             `json:"occupation_id"`
	Occupation        *occupationSummary `json:"occupation,omitempty"`
	CreatedAt         time.Time          `json:"created_at"`
}

func toReservationTypeOccupationResponse(o *model.ReservationTypeOccupation) reservationTypeOccupationResponse {
	resp := reservationTypeOccupationResponse{
		ID:                o.ID,
		ClinicID:          o.ClinicID,
		ReservationTypeID: o.ReservationTypeID,
		OccupationID:      o.OccupationID,
		CreatedAt:         localTime(o.CreatedAt),
	}
	if o.Occupation != nil {
		resp.Occupation = &occupationSummary{
			ID:   o.Occupation.ID,
			Name: o.Occupation.Name,
		}
	}
	return resp
}

type reservationTypeResponse struct {
	ID          uint64                        `json:"id"`
	ClinicID    uint64                        `json:"clinic_id"`
	Name        string                        `json:"name"`
	Color       string                        `json:"color"`
	Category    model.ReservationTypeCategory `json:"category"`
	IsActive    bool                          `json:"is_active"`
	Description string                        `json:"description"`
	SortOrder   int                           `json:"sort_order"`
	GroupID     *uint64                       `json:"group_id,omitempty"`
	Group       *groupSummary                 `json:"group,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`

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

func toReservationTypeResponse(st *model.ReservationType) reservationTypeResponse {
	resp := reservationTypeResponse{
		ID:                     st.ID,
		ClinicID:               st.ClinicID,
		Name:                   st.Name,
		Color:                  st.Color,
		Category:               st.Category,
		IsActive:               st.IsActive,
		Description:            st.Description,
		SortOrder:              st.SortOrder,
		GroupID:                st.GroupID,
		CreatedAt:              localTime(st.CreatedAt),
		UpdatedAt:              localTime(st.UpdatedAt),
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
