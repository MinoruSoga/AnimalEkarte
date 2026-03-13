package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// --- Input DTOs（Service層の公開I/F） ---

// CreatePetForOwnerInput は飼主登録時に同時作成するペットの入力DTO
type CreatePetForOwnerInput struct {
	Name            string     `json:"name" binding:"required"`
	AnimalSpeciesID uint64     `json:"animal_species_id" binding:"required"`
	PetNumber       string     `json:"pet_number"`
	PetNameKana     string     `json:"pet_name_kana"`
	Breed           string     `json:"breed"`
	Gender          string     `json:"gender"`
	BirthDate       *time.Time `json:"birth_date"`
	Remarks         string     `json:"remarks"`
}

// CreateOwnerInput は飼主作成の入力DTO
type CreateOwnerInput struct {
	OwnerName      string                   `json:"owner_name" binding:"required"`
	OwnerNameKana  string                   `json:"owner_name_kana"`
	BirthDate      *time.Time               `json:"birth_date"`
	Company        string                   `json:"company"`
	PostalCode     string                   `json:"postal_code"`
	Address1       string                   `json:"address1"`
	Address2       string                   `json:"address2"`
	HomePostalCode string                   `json:"home_postal_code"`
	HomeAddress1   string                   `json:"home_address1"`
	HomeAddress2   string                   `json:"home_address2"`
	Phone          string                   `json:"phone"`
	CompanyPhone   string                   `json:"company_phone"`
	Email          string                   `json:"email"`
	Remarks        string                   `json:"remarks"`
	IsDangerous    bool                     `json:"is_dangerous"`
	DiscountRate   float64                  `json:"discount_rate"`
	MembershipType model.MembershipType     `json:"membership_type"`
	Pets           []CreatePetForOwnerInput `json:"pets"`
}

// UpdateOwnerInput は飼主更新の入力DTO
type UpdateOwnerInput struct {
	OwnerName      string               `json:"owner_name"`
	OwnerNameKana  string               `json:"owner_name_kana"`
	BirthDate      *time.Time           `json:"birth_date"`
	Company        string               `json:"company"`
	PostalCode     string               `json:"postal_code"`
	Address1       string               `json:"address1"`
	Address2       string               `json:"address2"`
	HomePostalCode string               `json:"home_postal_code"`
	HomeAddress1   string               `json:"home_address1"`
	HomeAddress2   string               `json:"home_address2"`
	Phone          string               `json:"phone"`
	CompanyPhone   string               `json:"company_phone"`
	Email          string               `json:"email"`
	Remarks        string               `json:"remarks"`
	IsDangerous    bool                 `json:"is_dangerous"`
	DiscountRate   float64              `json:"discount_rate"`
	MembershipType model.MembershipType `json:"membership_type"`
}

// --- Interface ---

type OwnerService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	CreateWithPets(ctx context.Context, clinicID uint64, input CreateOwnerInput) (*model.Owner, error)
	Update(ctx context.Context, clinicID, id uint64, input UpdateOwnerInput) (*model.Owner, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// --- Implementation ---

type ownerService struct {
	repo repository.OwnerRepository
}

func NewOwnerService(repo repository.OwnerRepository) OwnerService {
	return &ownerService{repo: repo}
}

func (s *ownerService) List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return s.repo.FindAll(ctx, clinicID, page, limit, search)
}

func (s *ownerService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return s.repo.FindByID(ctx, clinicID, id)
}

func (s *ownerService) CreateWithPets(ctx context.Context, clinicID uint64, input CreateOwnerInput) (*model.Owner, error) {
	// ビジネスルールバリデーション
	if err := validateDiscountRate(input.DiscountRate); err != nil {
		return nil, err
	}
	if err := validateMembershipType(input.MembershipType); err != nil {
		return nil, err
	}
	for i, p := range input.Pets {
		if err := validatePetGender(p.Gender); err != nil {
			return nil, fmt.Errorf("pets[%d]: %w", i, err)
		}
	}

	// DTO → Model 変換
	owner := &model.Owner{
		ClinicID:       clinicID,
		OwnerName:      input.OwnerName,
		OwnerNameKana:  input.OwnerNameKana,
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
	}

	pets := make([]model.Pet, 0, len(input.Pets))
	for _, p := range input.Pets {
		pet := model.Pet{
			Name:            p.Name,
			AnimalSpeciesID: p.AnimalSpeciesID,
			PetNumber:       p.PetNumber,
			PetNameKana:     p.PetNameKana,
			Breed:           p.Breed,
			BirthDate:       p.BirthDate,
			Remarks:         p.Remarks,
		}
		if p.Gender != "" {
			pet.Gender = model.PetGender(p.Gender)
		}
		pets = append(pets, pet)
	}

	if err := s.repo.CreateWithPets(ctx, owner, pets); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "owner created with pets",
		slog.Uint64("owner_id", owner.ID),
		slog.Uint64("clinic_id", clinicID),
		slog.Int("pets_count", len(pets)))

	return owner, nil
}

func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input UpdateOwnerInput) (*model.Owner, error) {
	// ビジネスルールバリデーション
	if err := validateDiscountRate(input.DiscountRate); err != nil {
		return nil, err
	}
	if err := validateMembershipType(input.MembershipType); err != nil {
		return nil, err
	}

	// DTO → Model 変換
	owner := &model.Owner{
		ID:             id,
		ClinicID:       clinicID,
		OwnerName:      input.OwnerName,
		OwnerNameKana:  input.OwnerNameKana,
		BirthDate:      input.BirthDate,
		Company:        input.Company,
		PostalCode:     input.PostalCode,
		Address1:       input.Address1,
		Address2:       input.Address2,
		HomePostalCode: input.HomePostalCode,
		HomeAddress1:   input.HomeAddress1,
		HomeAddress2:   input.HomeAddress2,
		Phone:          input.Phone,
		CompanyPhone:   input.CompanyPhone,
		Email:          input.Email,
		Remarks:        input.Remarks,
		IsDangerous:    input.IsDangerous,
		DiscountRate:   input.DiscountRate,
		MembershipType: input.MembershipType,
	}

	if err := s.repo.Update(ctx, owner); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))

	return owner, nil
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}

// --- バリデーションヘルパー ---

func validateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("discount_rate must be between 0 and 100")
	}
	return nil
}

func validateMembershipType(t model.MembershipType) error {
	if t == "" {
		return nil
	}
	switch t {
	case model.MembershipTypeNonMember, model.MembershipTypeMember,
		model.MembershipTypeDeceased, model.MembershipTypeTransferred:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid membership_type: %s", t))
	}
}
