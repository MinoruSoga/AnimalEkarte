package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/service"
)

// liffSettingsResponse はLIFF向け公開設定レスポンス（機密フィールド除外）。
type liffSettingsResponse struct {
	Status               string `json:"status"`
	HeaderText           string `json:"header_text"`
	ReservationNotice    string `json:"reservation_notice"`
	CancelNotice         string `json:"cancel_notice"`
	PrivacyPolicy        string `json:"privacy_policy"`
	PhoneNumber          string `json:"phone_number"`
	RequestExample       string `json:"request_example"`
	LiffID               string `json:"liff_id"`
	ShowNoStaffOption    bool   `json:"show_no_staff_option"`
	BookingWindowMaxDays int    `json:"booking_window"`
}

func toLiffSettingsResponse(s *model.ReservationSetting) liffSettingsResponse {
	return liffSettingsResponse{
		Status:               s.Status,
		HeaderText:           s.HeaderText,
		ReservationNotice:    s.ReservationNotice,
		CancelNotice:         s.CancelNotice,
		PrivacyPolicy:        s.PrivacyPolicy,
		PhoneNumber:          s.PhoneNumber,
		RequestExample:       s.RequestExample,
		LiffID:               s.LiffID,
		ShowNoStaffOption:    s.ShowNoStaffOption,
		BookingWindowMaxDays: s.BookingWindowMaxDays,
	}
}

// liffCourseResponse はLIFF向けコースレスポンス。
type liffCourseResponse struct {
	ID                  uint64 `json:"id"`
	Name                string `json:"name"`
	ShortName           string `json:"short_name"`
	ShowShortName       bool   `json:"show_short_name"`
	DurationMinutes     int    `json:"duration_minutes"`
	ReservationComment  string `json:"reservation_comment"`
	ReservationImageURL string `json:"reservation_image_url"`
	SortOrder           int    `json:"sort_order"`
}

func toLiffCourseResponse(st model.ServiceType) liffCourseResponse {
	name := st.ReservationDisplayName
	if name == "" {
		name = st.Name
	}
	return liffCourseResponse{
		ID:                  st.ID,
		Name:                name,
		ShortName:           st.ShortName,
		ShowShortName:       st.ShowShortName,
		DurationMinutes:     st.DurationMinutes,
		ReservationComment:  st.ReservationComment,
		ReservationImageURL: st.ReservationImageURL,
		SortOrder:           st.SortOrder,
	}
}

// liffStaffResponse はLIFF向けスタッフレスポンス。
type liffStaffResponse struct {
	ID                  uint64 `json:"id"`
	Name                string `json:"name"`
	ReservationComment  string `json:"reservation_comment"`
	ReservationImageURL string `json:"reservation_image_url"`
	SortOrder           int    `json:"sort_order"`
}

func toLiffStaffResponse(st model.Staff) liffStaffResponse {
	name := st.ReservationDisplayName
	if name == "" {
		name = st.Name
	}
	return liffStaffResponse{
		ID:                  st.ID,
		Name:                name,
		ReservationComment:  st.ReservationComment,
		ReservationImageURL: st.ReservationImageURL,
		SortOrder:           st.SortOrder,
	}
}

// liffAvailableDateResponse は空き日付レスポンス。
type liffAvailableDateResponse struct {
	Date      string `json:"date"`
	Weekday   int    `json:"weekday"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// liffAvailableDatesResponse は空き日付一覧レスポンス。
type liffAvailableDatesResponse struct {
	Dates  []liffAvailableDateResponse `json:"dates"`
	Window service.BookingWindow       `json:"window"`
}

// liffTimeSlotResponse は時間枠レスポンス。
type liffTimeSlotResponse struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// liffReservationResponse は予約レスポンス。
type liffReservationResponse struct {
	ID         uint64    `json:"id"`
	Date       string    `json:"date"`
	StartTime  string    `json:"start_time"` // "HH:MM"
	EndTime    string    `json:"end_time"`   // "HH:MM"
	CourseName string    `json:"course_name"`
	StaffName  string    `json:"staff_name,omitempty"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func toLiffReservationResponse(r model.ReservationAppointment) liffReservationResponse {
	res := liffReservationResponse{
		ID:        r.ID,
		Date:      r.StartTime.Format("2006-01-02"),
		StartTime: r.StartTime.Format("15:04"),
		EndTime:   r.EndTime.Format("15:04"),
		Status:    string(r.Status),
		Notes:     r.Notes,
		CreatedAt: r.CreatedAt,
	}
	if r.ServiceType != nil {
		res.CourseName = r.ServiceType.ReservationDisplayName
		if res.CourseName == "" {
			res.CourseName = r.ServiceType.Name
		}
	}
	if r.Doctor != nil {
		res.StaffName = r.Doctor.ReservationDisplayName
		if res.StaffName == "" {
			res.StaffName = r.Doctor.Name
		}
	}
	return res
}
