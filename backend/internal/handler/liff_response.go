package handler

import (
	"encoding/json"
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

func toLiffSettingsResponse(s *model.LineReservationSetting) liffSettingsResponse {
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

func toLiffCourseResponse(st *model.ReservationType) liffCourseResponse {
	// 名前のフォールバックチェーン:
	// 1. ReservationDisplayName（LINE表示名）
	// 2. ShortName（ShowShortName が true の場合）
	// 3. Name（元の名称）
	name := st.ReservationDisplayName
	if name == "" {
		if st.ShowShortName && st.ShortName != "" {
			name = st.ShortName
		} else {
			name = st.Name
		}
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

func toLiffStaffResponse(st *model.Staff) liffStaffResponse {
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

type liffProfileOwnerPetResponse struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	AnimalSpecies *struct {
		Name string `json:"name"`
	} `json:"animal_species,omitempty"`
}

type liffProfileOwnerResponse struct {
	OwnerName string                        `json:"owner_name"`
	Phone     string                        `json:"phone"`
	Pets      []liffProfileOwnerPetResponse `json:"pets,omitempty"`
}

type liffProfileResponse struct {
	LineUserID       string                    `json:"line_user_id"`
	DisplayName      string                    `json:"display_name"`
	AdditionalFields json.RawMessage           `json:"additional_fields"`
	Owner            *liffProfileOwnerResponse `json:"owner,omitempty"`
}

func toLiffProfileResponse(c *model.LineCustomer) liffProfileResponse {
	resp := liffProfileResponse{
		LineUserID:       c.LineUserID,
		DisplayName:      c.DisplayName,
		AdditionalFields: json.RawMessage(c.AdditionalFields),
	}
	if c.Owner != nil {
		pets := make([]liffProfileOwnerPetResponse, 0, len(c.Owner.Pets))
		for i := range c.Owner.Pets {
			pet := liffProfileOwnerPetResponse{
				ID:   c.Owner.Pets[i].ID,
				Name: c.Owner.Pets[i].Name,
			}
			if c.Owner.Pets[i].AnimalSpecies != nil {
				pet.AnimalSpecies = &struct {
					Name string `json:"name"`
				}{Name: c.Owner.Pets[i].AnimalSpecies.Name}
			}
			pets = append(pets, pet)
		}
		resp.Owner = &liffProfileOwnerResponse{
			OwnerName: c.Owner.Name,
			Phone:     c.Owner.Phone,
			Pets:      pets,
		}
	}
	return resp
}

func toLiffReservationResponse(r *model.Reservation) liffReservationResponse {
	res := liffReservationResponse{
		ID:        r.ID,
		Date:      r.StartTime.In(time.Local).Format("2006-01-02"),
		StartTime: r.StartTime.In(time.Local).Format("15:04"),
		EndTime:   r.EndTime.In(time.Local).Format("15:04"),
		Status:    string(r.Status),
		Notes:     r.Notes,
		CreatedAt: r.CreatedAt.In(time.Local),
	}
	if r.ReservationType != nil {
		res.CourseName = r.ReservationType.ReservationDisplayName
		if res.CourseName == "" {
			if r.ReservationType.ShowShortName && r.ReservationType.ShortName != "" {
				res.CourseName = r.ReservationType.ShortName
			} else {
				res.CourseName = r.ReservationType.Name
			}
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

// liffReservationCreatedResponse はLIFF予約作成（201）専用レスポンス。
// 一覧用の liffReservationResponse とは別型（id/notes の2フィールドのみ露出）。
type liffReservationCreatedResponse struct {
	ID    uint64 `json:"id"`
	Notes string `json:"notes"`
}

func toLiffReservationCreatedResponse(r *model.Reservation) liffReservationCreatedResponse {
	return liffReservationCreatedResponse{
		ID:    r.ID,
		Notes: r.Notes,
	}
}

// liffTrimmingCourseResponse はLIFF向けトリミングコースレスポンス。
type liffTrimmingCourseResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       *int64 `json:"price"`
	SortOrder   int    `json:"sort_order"`
}

func toLiffTrimmingCourseResponse(c *model.TrimmingCourse) liffTrimmingCourseResponse {
	return liffTrimmingCourseResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Price:       c.Price,
		SortOrder:   c.SortOrder,
	}
}

// liffTrimmingOptionResponse はLIFF向けトリミングオプションレスポンス。
type liffTrimmingOptionResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       *int64 `json:"price"`
	SortOrder   int    `json:"sort_order"`
}

func toLiffTrimmingOptionResponse(o *model.TrimmingOption) liffTrimmingOptionResponse {
	return liffTrimmingOptionResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		Price:       o.Price,
		SortOrder:   o.SortOrder,
	}
}
