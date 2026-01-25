package validation

import (
	"regexp"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^[0-9+\-\s()]+$`)
)

// ValidateCreateOwner validates the create owner request
func ValidateCreateOwner(req *model.CreateOwnerRequest) error {
	if req.Name == "" {
		return apperrors.WrapInvalidInput("owner name is required")
	}
	if len(req.Name) > 100 {
		return apperrors.WrapInvalidInput("owner name must be less than 100 characters")
	}

	if req.Phone != "" {
		if len(req.Phone) > 20 {
			return apperrors.WrapInvalidInput("phone number must be less than 20 characters")
		}
		if !phoneRegex.MatchString(req.Phone) {
			return apperrors.WrapInvalidInput("invalid phone number format")
		}
	}

	if req.Email != "" {
		if len(req.Email) > 255 {
			return apperrors.WrapInvalidInput("email must be less than 255 characters")
		}
		if !emailRegex.MatchString(req.Email) {
			return apperrors.WrapInvalidInput("invalid email format")
		}
	}

	return nil
}

// ValidateUpdateOwner validates the update owner request
func ValidateUpdateOwner(req *model.UpdateOwnerRequest) error {
	if req.Name != "" {
		if len(req.Name) > 100 {
			return apperrors.WrapInvalidInput("owner name must be less than 100 characters")
		}
	}

	if req.Phone != "" {
		if len(req.Phone) > 20 {
			return apperrors.WrapInvalidInput("phone number must be less than 20 characters")
		}
		if !phoneRegex.MatchString(req.Phone) {
			return apperrors.WrapInvalidInput("invalid phone number format")
		}
	}

	if req.Email != "" {
		if len(req.Email) > 255 {
			return apperrors.WrapInvalidInput("email must be less than 255 characters")
		}
		if !emailRegex.MatchString(req.Email) {
			return apperrors.WrapInvalidInput("invalid email format")
		}
	}

	return nil
}
