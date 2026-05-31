package handler

import (
	"strings"

	"github.com/animal-ekarte/backend/internal/service"
)

// createStaffRequest はスタッフ登録リクエスト。
type createStaffRequest struct {
	Name          string  `json:"name"           binding:"required"`
	Email         string  `json:"email"     binding:"omitempty,email"`
	Password      string  `json:"password"  binding:"omitempty,min=8"`
	LicenseNumber string  `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     int     `json:"sort_order"`

	// LINE予約用フィールド
	StaffType              string `json:"staff_type"`
	ReservationDisplayName string `json:"reservation_display_name"`
	ReservationVisible     *bool  `json:"reservation_visible"`
	ReservationComment     string `json:"reservation_comment"`
	ReservationImageURL    string `json:"reservation_image_url"`
}

func (r *createStaffRequest) hasAccountEmail() bool {
	return strings.TrimSpace(r.Email) != ""
}

func (r *createStaffRequest) toCreateServiceInput(clinicID uint64) *service.CreateStaffInput {
	return &service.CreateStaffInput{
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

func (r *createStaffRequest) toCreateWithAccountServiceInput(clinicID uint64) *service.CreateStaffWithAccountInput {
	return &service.CreateStaffWithAccountInput{
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
	Name          *string `json:"name"`
	LicenseNumber *string `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
	Password      *string `json:"password" binding:"omitempty,min=8"`

	// LINE予約用フィールド
	StaffType              *string `json:"staff_type"`
	ReservationDisplayName *string `json:"reservation_display_name"`
	ReservationVisible     *bool   `json:"reservation_visible"`
	ReservationComment     *string `json:"reservation_comment"`
	ReservationImageURL    *string `json:"reservation_image_url"`
}

func (r *updateStaffRequest) toServiceInput() *service.UpdateStaffInput {
	return &service.UpdateStaffInput{
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
	GroupIDs []uint64 `json:"group_ids"`
}

// setStaffClinicAssignmentsRequest はクリニック割当リクエスト。
type setStaffClinicAssignmentsRequest struct {
	ClinicIDs []uint64 `json:"clinic_ids"`
}

// setStaffExcludedReservationTypesRequest は除外予約種別リクエスト。
type setStaffExcludedReservationTypesRequest struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids"`
}

// setStaffCapableReservationTypesRequest は対応可能予約種別リクエスト。
type setStaffCapableReservationTypesRequest struct {
	ReservationTypeIDs []uint64 `json:"reservation_type_ids"`
}
