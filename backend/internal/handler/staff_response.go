package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type jobTitleInStaffResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type staffResponse struct {
	ID            uint64                   `json:"id"`
	ClinicID      uint64                   `json:"clinic_id"`
	Name          string                   `json:"name"`
	IsActive      bool                     `json:"is_active"`
	StaffRole     string                   `json:"staff_role"`
	JobTitleID    *uint64                  `json:"job_title_id,omitempty"`
	LicenseNumber string                   `json:"license_number"`
	SortOrder     int                      `json:"sort_order"`
	JobTitle      *jobTitleInStaffResponse `json:"job_title,omitempty"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
}

func toJobTitleInStaffResponse(jt *model.JobTitle) *jobTitleInStaffResponse {
	if jt == nil {
		return nil
	}
	return &jobTitleInStaffResponse{
		ID:        jt.ID,
		ClinicID:  jt.ClinicID,
		Name:      jt.Name,
		SortOrder: jt.SortOrder,
		IsActive:  jt.IsActive,
		CreatedAt: jt.CreatedAt,
		UpdatedAt: jt.UpdatedAt,
	}
}

func toStaffResponse(s *model.Staff) staffResponse {
	return staffResponse{
		ID:            s.ID,
		ClinicID:      s.ClinicID,
		Name:          s.Name,
		IsActive:      s.IsActive,
		StaffRole:     string(s.StaffRole),
		JobTitleID:    s.JobTitleID,
		LicenseNumber: s.LicenseNumber,
		SortOrder:     s.SortOrder,
		JobTitle:      toJobTitleInStaffResponse(s.JobTitle),
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func toStaffResponseList(staffs []model.Staff) []staffResponse {
	result := make([]staffResponse, 0, len(staffs))
	for i := range staffs {
		result = append(result, toStaffResponse(&staffs[i]))
	}
	return result
}

// staffSummaryResponse はネストされたレスポンスで使用するスタッフの要約型
type staffSummaryResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	StaffRole string `json:"staff_role"`
}

// toStaffSummary は *model.Staff を *staffSummaryResponse に変換する。nilの場合はnilを返す。
func toStaffSummary(s *model.Staff) *staffSummaryResponse {
	if s == nil {
		return nil
	}
	return &staffSummaryResponse{
		ID:        s.ID,
		Name:      s.Name,
		StaffRole: string(s.StaffRole),
	}
}
