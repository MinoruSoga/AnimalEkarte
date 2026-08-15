package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// ForgotPasswordRequest is the password-reset request body.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

// ResetPasswordRequest is the password-reset completion body.
type ResetPasswordRequest struct {
	Token    string `json:"token"    binding:"required,max=128"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// ChangeMyPasswordRequest is the authenticated password-change body.
type ChangeMyPasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,max=72"`
	NewPassword     string `json:"new_password"     binding:"required,min=8,max=72"`
}

// ValidatePassword enforces the shared auth password complexity requirement.
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください")
	}
	if len([]byte(password)) > 72 {
		return apperrors.WrapInvalidInput("パスワードは72バイト以下で入力してください")
	}
	var hasLetter bool
	var hasDigit bool
	for _, character := range password {
		hasLetter = hasLetter || unicode.IsLetter(character)
		hasDigit = hasDigit || unicode.IsDigit(character)
	}
	if !hasLetter || !hasDigit {
		return apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
	}
	return nil
}

// ChangeMyPassword verifies the current password before replacing its hash.
func (h *HTTPHandler) ChangeMyPassword(c *gin.Context) {
	ctx := c.Request.Context()
	var request ChangeMyPasswordRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := ValidatePassword(request.NewPassword); err != nil {
		httpapi.RespondError(c, err)
		return
	}

	staffID, ok := httpapi.ExtractStaffID(c)
	if !ok {
		return
	}
	clinicID, ok := httpapi.ExtractClinicID(c)
	if !ok {
		return
	}
	staff, err := h.deps.Staff.GetByID(ctx, staffID)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if staff.AccountID == nil {
		httpapi.RespondError(
			c,
			apperrors.WrapInvalidInput("このスタッフにはアカウントが紐づいていません"),
		)
		return
	}
	auditClinicID, err := h.resolveSelfPasswordAuditClinic(
		ctx,
		staff,
		clinicID,
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	passwordChanger, ok := h.deps.Accounts.(AccountPasswordChanger)
	if !ok {
		httpapi.RespondError(
			c,
			apperrors.WrapInternalServerError(
				"atomic account password change is not configured",
			),
		)
		return
	}
	if err := passwordChanger.ChangePassword(
		ctx,
		*staff.AccountID,
		request.CurrentPassword,
		request.NewPassword,
		CredentialMutationAudit{
			ClinicID:      auditClinicID,
			ActorID:       &staffID,
			ActorType:     model.AuditActorTypeStaff,
			Action:        model.AuditActionAuthPasswordChange,
			TargetStaffID: staff.ID,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.Header.Get("User-Agent"),
		},
	); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "パスワードを変更しました"})
}

// ForgotPassword accepts reset requests without disclosing account existence.
func (h *HTTPHandler) ForgotPassword(c *gin.Context) {
	var request ForgotPasswordRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := h.deps.PasswordReset.ForgotPassword(
		c.Request.Context(),
		request.Email,
	); err != nil {
		slog.ErrorContext(
			c.Request.Context(),
			"forgot password request processing failed",
			"error_type", fmt.Sprintf("%T", err),
		)
	}
	c.JSON(
		http.StatusOK,
		gin.H{"message": "If the email exists, a reset link has been sent."},
	)
}

// ResetPassword validates password complexity before consuming the reset token.
func (h *HTTPHandler) ResetPassword(c *gin.Context) {
	var request ResetPasswordRequest
	if err := bindAuthJSON(c, &request); err != nil {
		httpapi.RespondError(c, err)
		return
	}
	if err := ValidatePassword(request.Password); err != nil {
		httpapi.RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}
	completionService, ok := h.deps.PasswordReset.(PasswordResetCompletionService)
	if !ok {
		httpapi.RespondError(
			c,
			apperrors.WrapInternalServerError(
				"auditable password reset completion is not configured",
			),
		)
		return
	}
	_, err := completionService.ResetPasswordWithResult(
		c.Request.Context(),
		request.Token,
		request.Password,
		CredentialMutationAudit{
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionAuthPasswordReset,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.Header.Get("User-Agent"),
		},
	)
	if err != nil {
		httpapi.RespondError(c, err)
		return
	}
	c.JSON(
		http.StatusOK,
		gin.H{"message": "Password has been reset successfully."},
	)
}
