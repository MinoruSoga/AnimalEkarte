package validation

import (
	"time"

	"github.com/google/uuid"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ValidateCreatePet validates the create pet request
func ValidateCreatePet(req *model.CreatePetRequest) error {
	if req.Name == "" {
		return apperrors.WrapInvalidInput("pet name is required")
	}
	if len(req.Name) > 100 {
		return apperrors.WrapInvalidInput("pet name must be less than 100 characters")
	}

	if req.Species == "" {
		return apperrors.WrapInvalidInput("pet species is required")
	}
	if len(req.Species) > 50 {
		return apperrors.WrapInvalidInput("pet species must be less than 50 characters")
	}

	if req.Breed != nil && len(*req.Breed) > 100 {
		return apperrors.WrapInvalidInput("pet breed must be less than 100 characters")
	}

	if req.Gender != nil && *req.Gender != "" && *req.Gender != "雄" && *req.Gender != "雌" && *req.Gender != "不明" {
		return apperrors.WrapInvalidInput("pet gender must be '雄', '雌', or '不明'")
	}

	if req.OwnerID == "" {
		return apperrors.WrapInvalidInput("owner ID is required")
	}
	if _, err := uuid.Parse(req.OwnerID); err != nil {
		return apperrors.WrapInvalidInput("invalid owner ID format")
	}

	if req.BirthDate != nil && *req.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", *req.BirthDate); err != nil {
			return apperrors.WrapInvalidInput("invalid birth date format, expected YYYY-MM-DD")
		}
	}

	if req.Environment != nil && *req.Environment != "" && *req.Environment != "室内" && *req.Environment != "室外" && *req.Environment != "混合" {
		return apperrors.WrapInvalidInput("environment must be '室内', '室外', or '混合'")
	}

	if req.Status != "" && req.Status != "生存" && req.Status != "死亡" {
		return apperrors.WrapInvalidInput("status must be '生存' or '死亡'")
	}

	return nil
}

// ValidateUpdatePet validates the update pet request
func ValidateUpdatePet(req *model.UpdatePetRequest) error {
	if req.Name != nil && *req.Name != "" {
		if len(*req.Name) > 100 {
			return apperrors.WrapInvalidInput("pet name must be less than 100 characters")
		}
	}

	if req.Species != nil && *req.Species != "" {
		if len(*req.Species) > 50 {
			return apperrors.WrapInvalidInput("pet species must be less than 50 characters")
		}
	}

	if req.Breed != nil && len(*req.Breed) > 100 {
		return apperrors.WrapInvalidInput("pet breed must be less than 100 characters")
	}

	if req.Gender != nil && *req.Gender != "" && *req.Gender != "雄" && *req.Gender != "雌" && *req.Gender != "不明" {
		return apperrors.WrapInvalidInput("pet gender must be '雄', '雌', or '不明'")
	}

	if req.OwnerID != nil && *req.OwnerID != "" {
		if _, err := uuid.Parse(*req.OwnerID); err != nil {
			return apperrors.WrapInvalidInput("invalid owner ID format")
		}
	}

	if req.BirthDate != nil && *req.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", *req.BirthDate); err != nil {
			return apperrors.WrapInvalidInput("invalid birth date format, expected YYYY-MM-DD")
		}
	}

	if req.Environment != nil && *req.Environment != "" && *req.Environment != "室内" && *req.Environment != "室外" && *req.Environment != "混合" {
		return apperrors.WrapInvalidInput("environment must be '室内', '室外', or '混合'")
	}

	if req.Status != nil && *req.Status != "" && *req.Status != "生存" && *req.Status != "死亡" {
		return apperrors.WrapInvalidInput("status must be '生存' or '死亡'")
	}

	return nil
}
