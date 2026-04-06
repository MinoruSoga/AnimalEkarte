package handler

// createStaffRequest はスタッフ登録リクエスト。
type createStaffRequest struct {
	Name          string  `json:"name"           binding:"required"`
	LicenseNumber string  `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     int     `json:"sort_order"`
}

// updateStaffRequest はスタッフ更新リクエスト。nil = 未送信として扱う。
type updateStaffRequest struct {
	Name          *string `json:"name"`
	LicenseNumber *string `json:"license_number"`
	OccupationID  *uint64 `json:"occupation_id"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
	Password      *string `json:"password"`
}

// reorderStaffRequest はスタッフ並び替えリクエスト。
type reorderStaffRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
