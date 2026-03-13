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

// --- Owner DB column constants ---

const (
	colOwnerName      = "owner_name"
	colOwnerNameKana  = "owner_name_kana"
	colBirthDate      = "birth_date"
	colCompany        = "company"
	colPostalCode     = "postal_code"
	colAddress1       = "address1"
	colAddress2       = "address2"
	colHomePostalCode = "home_postal_code"
	colHomeAddress1   = "home_address1"
	colHomeAddress2   = "home_address2"
	colPhone          = "phone"
	colCompanyPhone   = "company_phone"
	colEmail          = "email"
	colRemarks        = "remarks"
	colIsDangerous    = "is_dangerous"
	colDiscountRate   = "discount_rate"
	colMembershipType = "membership_type"
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
// ビジネスルール（DiscountRate 範囲, MembershipType 列挙値）は Service 層で検証する。
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

// UpdateOwnerInput は飼主更新の入力DTO（全フィールドポインタ型: nil = 未指定, 非nil = 更新対象）
// ビジネスルールは Service 層で検証する。
type UpdateOwnerInput struct {
	OwnerName      *string               `json:"owner_name"`
	OwnerNameKana  *string               `json:"owner_name_kana"`
	BirthDate      *time.Time            `json:"birth_date"`
	Company        *string               `json:"company"`
	PostalCode     *string               `json:"postal_code"`
	Address1       *string               `json:"address1"`
	Address2       *string               `json:"address2"`
	HomePostalCode *string               `json:"home_postal_code"`
	HomeAddress1   *string               `json:"home_address1"`
	HomeAddress2   *string               `json:"home_address2"`
	Phone          *string               `json:"phone"`
	CompanyPhone   *string               `json:"company_phone"`
	Email          *string               `json:"email"`
	Remarks        *string               `json:"remarks"`
	IsDangerous    *bool                 `json:"is_dangerous"`
	DiscountRate   *float64              `json:"discount_rate"`
	MembershipType *model.MembershipType `json:"membership_type"`
}

// --- Interface ---

type OwnerService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, search string) ([]model.Owner, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
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

func (s *ownerService) CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
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

func (s *ownerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
	// ビジネスルールバリデーション
	if input.DiscountRate != nil {
		if err := validateDiscountRate(*input.DiscountRate); err != nil {
			return nil, err
		}
	}
	if input.MembershipType != nil {
		if err := validateMembershipType(*input.MembershipType); err != nil {
			return nil, err
		}
	}

	// 更新フィールドマップ構築（nil フィールドはスキップ）
	fields := buildOwnerUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "owner updated",
		slog.Uint64("owner_id", id),
		slog.Uint64("clinic_id", clinicID))

	// DB の最新状態を返す
	return s.repo.FindByID(ctx, clinicID, id)
}

// buildOwnerUpdateFields はポインタが非 nil のフィールドのみ map に追加する
func buildOwnerUpdateFields(input *UpdateOwnerInput) map[string]any {
	fields := make(map[string]any)
	if input.OwnerName != nil {
		fields[colOwnerName] = *input.OwnerName
	}
	if input.OwnerNameKana != nil {
		fields[colOwnerNameKana] = *input.OwnerNameKana
	}
	if input.BirthDate != nil {
		fields[colBirthDate] = *input.BirthDate
	}
	if input.Company != nil {
		fields[colCompany] = *input.Company
	}
	if input.PostalCode != nil {
		fields[colPostalCode] = *input.PostalCode
	}
	if input.Address1 != nil {
		fields[colAddress1] = *input.Address1
	}
	if input.Address2 != nil {
		fields[colAddress2] = *input.Address2
	}
	if input.HomePostalCode != nil {
		fields[colHomePostalCode] = *input.HomePostalCode
	}
	if input.HomeAddress1 != nil {
		fields[colHomeAddress1] = *input.HomeAddress1
	}
	if input.HomeAddress2 != nil {
		fields[colHomeAddress2] = *input.HomeAddress2
	}
	if input.Phone != nil {
		fields[colPhone] = *input.Phone
	}
	if input.CompanyPhone != nil {
		fields[colCompanyPhone] = *input.CompanyPhone
	}
	if input.Email != nil {
		fields[colEmail] = *input.Email
	}
	if input.Remarks != nil {
		fields[colRemarks] = *input.Remarks
	}
	if input.IsDangerous != nil {
		fields[colIsDangerous] = *input.IsDangerous
	}
	if input.DiscountRate != nil {
		fields[colDiscountRate] = *input.DiscountRate
	}
	if input.MembershipType != nil {
		fields[colMembershipType] = *input.MembershipType
	}
	return fields
}

func (s *ownerService) Delete(ctx context.Context, clinicID, id uint64) error {
	return s.repo.Delete(ctx, clinicID, id)
}
