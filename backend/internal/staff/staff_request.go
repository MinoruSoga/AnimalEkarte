package staff

import (
	"strings"
)

// createStaffRequest はスタッフ登録リクエスト。
type createStaffRequest struct {
	Name          string  `json:"name"           binding:"required,max=255"`
	Email         string  `json:"email"          binding:"omitempty,email,max=254"`
	Password      string  `json:"password"       binding:"omitempty,min=8,max=72"`
	LicenseNumber string  `json:"license_number" binding:"max=255"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     int     `json:"sort_order"`

	// LINE予約用フィールド
	StaffType              string `json:"staff_type"               binding:"omitempty,max=16,oneof=doctor nurse trimmer resource"`
	ReservationDisplayName string `json:"reservation_display_name" binding:"max=255"`
	ReservationVisible     *bool  `json:"reservation_visible"`
	ReservationComment     string `json:"reservation_comment"      binding:"max=2000"`
	ReservationImageURL    string `json:"reservation_image_url"    binding:"max=2048"`
}

func (r *createStaffRequest) hasAccountEmail() bool {
	return strings.TrimSpace(r.Email) != ""
}

func (r *createStaffRequest) toCreateServiceInput(clinicID uint64) *CreateStaffInput {
	return &CreateStaffInput{
		ClinicID:               clinicID,
		Name:                   r.Name,
		LicenseNumber:          r.LicenseNumber,
		OccupationID:           r.OccupationID,
		SortOrder:              r.SortOrder,
		StaffType:              r.StaffType,
		ReservationDisplayName: r.ReservationDisplayName,
		ReservationVisible:     r.ReservationVisible,
		ReservationComment:     r.ReservationComment,
		ReservationImageURL:    r.ReservationImageURL,
	}
}

func (r *createStaffRequest) toCreateWithAccountServiceInput(clinicID uint64) *CreateStaffWithAccountInput {
	return &CreateStaffWithAccountInput{
		ClinicID:               clinicID,
		Name:                   r.Name,
		LicenseNumber:          r.LicenseNumber,
		OccupationID:           r.OccupationID,
		SortOrder:              r.SortOrder,
		Email:                  strings.TrimSpace(r.Email),
		Password:               r.Password,
		StaffType:              r.StaffType,
		ReservationDisplayName: r.ReservationDisplayName,
		ReservationVisible:     r.ReservationVisible,
		ReservationComment:     r.ReservationComment,
		ReservationImageURL:    r.ReservationImageURL,
	}
}

// updateStaffRequest はスタッフ更新リクエスト。nil = 未送信として扱う。
type updateStaffRequest struct {
	Name          *string `json:"name"           binding:"omitempty,max=255"`
	LicenseNumber *string `json:"license_number" binding:"omitempty,max=255"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
	Password      *string `json:"password" binding:"omitempty,min=8,max=72"`

	// LINE予約用フィールド
	StaffType              *string `json:"staff_type"               binding:"omitempty,max=16,oneof=doctor nurse trimmer resource"`
	ReservationDisplayName *string `json:"reservation_display_name" binding:"omitempty,max=255"`
	ReservationVisible     *bool   `json:"reservation_visible"`
	ReservationComment     *string `json:"reservation_comment"      binding:"omitempty,max=2000"`
	ReservationImageURL    *string `json:"reservation_image_url"    binding:"omitempty,max=2048"`
}

func (r *updateStaffRequest) toServiceInput() *UpdateStaffInput {
	return &UpdateStaffInput{
		Name:                   r.Name,
		LicenseNumber:          r.LicenseNumber,
		OccupationID:           r.OccupationID,
		SortOrder:              r.SortOrder,
		IsActive:               r.IsActive,
		Password:               r.Password,
		StaffType:              r.StaffType,
		ReservationDisplayName: r.ReservationDisplayName,
		ReservationVisible:     r.ReservationVisible,
		ReservationComment:     r.ReservationComment,
		ReservationImageURL:    r.ReservationImageURL,
	}
}

// setStaffPermissionGroupsRequest は権限グループ割当リクエスト。
type setStaffPermissionGroupsRequest struct {
	GroupIDs []uint64 `json:"group_ids" binding:"max=50,dive"`
}

// setStaffClinicAssignmentsRequest はクリニック割当リクエスト。
type setStaffClinicAssignmentsRequest struct {
	ClinicIDs []uint64 `json:"clinic_ids" binding:"max=50,dive"`
}

// setStaffExcludedReservationTypesRequest は除外予約種別リクエスト。
type setStaffExcludedReservationTypesRequest struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids" binding:"max=50,dive"`
}

// setStaffCapableReservationTypesRequest は対応可能予約種別リクエスト。
type setStaffCapableReservationTypesRequest struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids" binding:"max=50,dive"`
}
