package handler

// createStaffRequest はスタッフ登録リクエスト。
type createStaffRequest struct {
	Name          string  `json:"name"           binding:"required"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	LicenseNumber string  `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     int     `json:"sort_order"`

	// LINE予約用フィールド
	StaffType          string `json:"staff_type"`
	ReservationVisible *bool  `json:"reservation_visible"`
	ReservationComment string `json:"reservation_comment"`
}

// updateStaffRequest はスタッフ更新リクエスト。nil = 未送信として扱う。
type updateStaffRequest struct {
	Name          *string `json:"name"`
	LicenseNumber *string `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
	Password      *string `json:"password"`

	// LINE予約用フィールド
	StaffType          *string `json:"staff_type"`
	ReservationVisible *bool   `json:"reservation_visible"`
	ReservationComment *string `json:"reservation_comment"`
}

// reorderStaffRequest はスタッフ並び替えリクエスト。
type reorderStaffRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
