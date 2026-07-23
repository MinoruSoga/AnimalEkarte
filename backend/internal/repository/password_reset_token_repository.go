package repository

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/auth"
)

// PasswordResetTokenRepository is a compatibility alias for auth persistence.
type PasswordResetTokenRepository = auth.PasswordResetTokenRepository

// NewPasswordResetTokenRepository constructs the password reset token repository.
func NewPasswordResetTokenRepository(db *gorm.DB) PasswordResetTokenRepository {
	return auth.NewPasswordResetTokenRepository(db)
}
