package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// ChangeMyPassword は認証済みユーザーが自分のパスワードを変更する（BUG-148）
func (h *Handler) ChangeMyPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	// 新しいパスワードの複雑性チェック
	if err := validatePassword(req.NewPassword); err != nil {
		RespondError(c, err)
		return
	}

	staffID, ok := extractStaffID(c)
	if !ok {
		return
	}
	staff, err := h.svc.Staff.GetByID(ctx, staffID)
	if err != nil {
		RespondError(c, err)
		return
	}
	if staff.AccountID == nil {
		RespondError(c, apperrors.WrapInvalidInput("このスタッフにはアカウントが紐づいていません"))
		return
	}

	account, err := h.svc.Account.GetByID(ctx, *staff.AccountID)
	if err != nil {
		RespondError(c, err)
		return
	}

	// 現在のパスワード検証
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		RespondError(c, apperrors.WrapUnauthorized("現在のパスワードが正しくありません"))
		return
	}

	// 新しいパスワードをハッシュ化して更新
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), config.BcryptCost)
	if err != nil {
		RespondError(c, apperrors.Wrap(err, "failed to hash password"))
		return
	}
	if err := h.svc.Account.UpdatePasswordHash(ctx, *staff.AccountID, string(hashed)); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// ForgotPassword はパスワードリセットメールを送信する。
// アカウントが存在しない場合も 200 を返す（メール存在有無の漏洩防止）。
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	ctx := c.Request.Context()
	if err := h.svc.PasswordReset.ForgotPassword(ctx, req.Email); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a reset link has been sent."})
}

// ResetPassword は rawToken と新パスワードでパスワードを更新する。
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
		return
	}

	if err := validatePassword(req.Password); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	ctx := c.Request.Context()
	if err := h.svc.PasswordReset.ResetPassword(ctx, req.Token, req.Password); err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}
