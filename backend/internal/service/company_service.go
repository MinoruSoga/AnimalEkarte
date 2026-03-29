package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// UpdateCompanyInput は法人情報部分更新の入力DTO
type UpdateCompanyInput struct {
	Name                      *string
	PostalCode                *string
	Address                   *string
	PhoneNumber               *string
	FaxNumber                 *string
	Email                     *string
	Website                   *string
	DirectorName              *string
	RegistrationNumber        *string
	InvoiceRegistrationNumber *string
	LogoURL                   *string
}

// CompanyService は法人情報のビジネスロジックインターフェース
type CompanyService interface {
	Get(ctx context.Context) (*model.Company, error)
	Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error)
}

type companyService struct {
	repo repository.CompanyRepository
}

// NewCompanyService は CompanyService を初期化して返す
func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

// Get は法人情報シングルトンを取得する
func (s *companyService) Get(ctx context.Context) (*model.Company, error) {
	return s.repo.Get(ctx)
}

// Update は法人情報を部分更新する
func (s *companyService) Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error) {
	fields := buildCompanyUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update company")
	}
	slog.InfoContext(ctx, "company updated")
	return s.repo.Get(ctx)
}

func buildCompanyUpdateFields(input *UpdateCompanyInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.PostalCode != nil {
		fields["postal_code"] = *input.PostalCode
	}
	if input.Address != nil {
		fields["address"] = *input.Address
	}
	if input.PhoneNumber != nil {
		fields["phone_number"] = *input.PhoneNumber
	}
	if input.FaxNumber != nil {
		fields["fax_number"] = *input.FaxNumber
	}
	if input.Email != nil {
		fields["email"] = *input.Email
	}
	if input.Website != nil {
		fields["website"] = *input.Website
	}
	if input.DirectorName != nil {
		fields["director_name"] = *input.DirectorName
	}
	if input.RegistrationNumber != nil {
		fields["registration_number"] = *input.RegistrationNumber
	}
	if input.InvoiceRegistrationNumber != nil {
		fields["invoice_registration_number"] = *input.InvoiceRegistrationNumber
	}
	if input.LogoURL != nil {
		fields["logo_url"] = *input.LogoURL
	}
	return fields
}
