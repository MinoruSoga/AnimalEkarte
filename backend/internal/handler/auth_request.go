package handler

// LoginInput はログインリクエストのボディ
type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// forgotPasswordRequest はパスワードリセットリクエストのボディ
type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// resetPasswordRequest はパスワードリセット実行のボディ
type resetPasswordRequest struct {
	Token       string `json:"token"        binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// changePasswordRequest はパスワード変更リクエストのボディ
type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password"     binding:"required,min=8"`
}
