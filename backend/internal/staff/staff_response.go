package staff

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type occupationInStaffResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type staffResponse struct {
	ID            uint64                     `json:"id"`
	Name          string                     `json:"name"`
	IsActive      bool                       `json:"is_active"`
	OccupationID  *uint64                    `json:"occupation_id,omitempty"`
	LicenseNumber string                     `json:"license_number"`
	SortOrder     int                        `json:"sort_order"`
	Email         string                     `json:"email"`
	Occupation    *occupationInStaffResponse `json:"occupation,omitempty"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`

	// LINE予約用フィールド
	StaffType              model.StaffType `json:"staff_type"`
	ReservationDisplayName string          `json:"reservation_display_name"`
	ReservationVisible     bool            `json:"reservation_visible"`
	ReservationComment     string          `json:"reservation_comment"`
	ReservationImageURL    string          `json:"reservation_image_url"`
}

func toOccupationInStaffResponse(occ *model.Occupation) *occupationInStaffResponse {
	if occ == nil {
		return nil
	}
	return &occupationInStaffResponse{
		ID:        occ.ID,
		ClinicID:  occ.ClinicID,
		Name:      occ.Name,
		SortOrder: occ.SortOrder,
		IsActive:  occ.IsActive,
		CreatedAt: localTime(occ.CreatedAt),
		UpdatedAt: localTime(occ.UpdatedAt),
	}
}

func toStaffResponse(s *model.Staff) staffResponse {
	email := ""
	if s.Account != nil {
		email = s.Account.Email
	}
	return staffResponse{
		ID:                     s.ID,
		Name:                   s.Name,
		IsActive:               s.IsActive,
		OccupationID:           s.OccupationID,
		LicenseNumber:          s.LicenseNumber,
		SortOrder:              s.SortOrder,
		Email:                  email,
		Occupation:             toOccupationInStaffResponse(s.Occupation),
		CreatedAt:              localTime(s.CreatedAt),
		UpdatedAt:              localTime(s.UpdatedAt),
		StaffType:              s.StaffType,
		ReservationDisplayName: s.ReservationDisplayName,
		ReservationVisible:     s.ReservationVisible,
		ReservationComment:     s.ReservationComment,
		ReservationImageURL:    s.ReservationImageURL,
	}
}

// staffSummaryResponse はネストされたレスポンスで使用するスタッフの要約型
type staffSummaryResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// toStaffSummary は *model.Staff を *staffSummaryResponse に変換する。nilの場合はnilを返す。
func toStaffSummary(s *model.Staff) *staffSummaryResponse {
	if s == nil {
		return nil
	}
	return &staffSummaryResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

// 関連付けエンドポイント（権限グループ / 所属医院 / 予約種別）の ID リスト応答型。
// 旧 gin.H{"...": ids} を P7 準拠の typed struct に置き換える（JSON 表現は同一）。
type staffPermissionGroupsResponse struct {
	GroupIDs []uint64 `json:"group_ids"`
}

type staffClinicAssignmentsResponse struct {
	ClinicIDs []uint64 `json:"clinic_ids"`
}

type staffReservationTypesResponse struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids"`
}
