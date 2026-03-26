package handler

// createUserRequest はユーザー作成のバインド struct
type createUserRequest struct {
	Email       string  `json:"email"        binding:"required,email"`
	Password    string  `json:"password"     binding:"required,min=8"`
	DisplayName string  `json:"display_name" binding:"required"`
	UserType    string  `json:"user_type"    binding:"required"`
	StaffID     *uint64 `json:"staff_id"`
	IsMain      bool    `json:"is_main"`
}

// updateUserRequest はユーザー更新のバインド struct
type updateUserRequest struct {
	DisplayName *string `json:"display_name"`
	UserType    *string `json:"user_type"`
	StaffID     *uint64 `json:"staff_id"`
	AvatarURL   *string `json:"avatar_url"`
	Status      *string `json:"status"`
}
