package handler

import "github.com/animal-ekarte/backend/internal/model"

// createStaffRequest はスタッフ登録リクエスト。
// スタッフ情報とシステムアカウント情報を同時に受け取る。
type createStaffRequest struct {
	Name          string          `json:"name"           binding:"required"`
	StaffRole     model.StaffRole `json:"staff_role"     binding:"required"`
	Email         string          `json:"email"          binding:"required,email"`
	Password      string          `json:"password"       binding:"required,min=8"`
	LicenseNumber string          `json:"license_number"`
	JobTitleID    *uint64         `json:"job_title_id"`
	SortOrder     int             `json:"sort_order"`
}

// updateStaffRequest はスタッフ更新リクエスト。nil = 未送信として扱う。
type updateStaffRequest struct {
	Name          *string          `json:"name"`
	StaffRole     *model.StaffRole `json:"staff_role"`
	LicenseNumber *string          `json:"license_number"`
	JobTitleID    *uint64          `json:"job_title_id"`
	SortOrder     *int             `json:"sort_order"`
	IsActive      *bool            `json:"is_active"`
}
