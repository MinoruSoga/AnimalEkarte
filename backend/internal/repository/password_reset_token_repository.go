package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/repository/passwordreset"
)

// PasswordResetTokenRepository is a stable facade alias for the passwordreset domain package.
type PasswordResetTokenRepository = passwordreset.Repository

// NewPasswordResetTokenRepository constructs the password reset token repository.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return passwordreset.New(db)
}
