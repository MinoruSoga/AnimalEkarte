package clinic

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// --- DB column constants ---

const (
	colCompanyName                      = "name"
	colCompanyPostalCode                = "postal_code"
	colCompanyAddress                   = "address"
	colCompanyPhoneNumber               = "phone_number"
	colCompanyFaxNumber                 = "fax_number"
	colCompanyEmail                     = "email"
	colCompanyWebsite                   = "website"
	colCompanyDirectorName              = "director_name"
	colCompanyRegistrationNumber        = "registration_number"
	colCompanyInvoiceRegistrationNumber = "invoice_registration_number"
	colCompanyLogoURL                   = "logo_url"
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

func BuildCompanyUpdate(input *UpdateCompanyInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields[colCompanyName] = *input.Name
	}
	if input.PostalCode != nil {
		fields[colCompanyPostalCode] = *input.PostalCode
	}
	if input.Address != nil {
		fields[colCompanyAddress] = *input.Address
	}
	if input.PhoneNumber != nil {
		fields[colCompanyPhoneNumber] = *input.PhoneNumber
	}
	if input.FaxNumber != nil {
		fields[colCompanyFaxNumber] = *input.FaxNumber
	}
	if input.Email != nil {
		fields[colCompanyEmail] = *input.Email
	}
	if input.Website != nil {
		fields[colCompanyWebsite] = *input.Website
	}
	if input.DirectorName != nil {
		fields[colCompanyDirectorName] = *input.DirectorName
	}
	if input.RegistrationNumber != nil {
		fields[colCompanyRegistrationNumber] = *input.RegistrationNumber
	}
	if input.InvoiceRegistrationNumber != nil {
		fields[colCompanyInvoiceRegistrationNumber] = *input.InvoiceRegistrationNumber
	}
	if input.LogoURL != nil {
		fields[colCompanyLogoURL] = *input.LogoURL
	}
	return fields
}

// CompanyService は法人情報のビジネスロジックインターフェース
type CompanyService interface {
	Get(ctx context.Context) (*model.Company, error)
	Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error)
}

type companyService struct {
	repo CompanyRepository
}

// NewCompanyService は CompanyService を初期化して返す
func NewCompanyService(repo CompanyRepository) CompanyService {
	return &companyService{repo: repo}
}

// Get は法人情報シングルトンを取得する
func (s *companyService) Get(ctx context.Context) (*model.Company, error) {
	result, err := s.repo.FindSingleton(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get company", "error", err)
		return nil, apperrors.Wrap(err, "failed to get company")
	}
	return result, nil
}

// Update は法人情報を部分更新する
func (s *companyService) Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error) {
	fields := BuildCompanyUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(sharedkernel.ErrMsgAtLeastOneField)
	}
	// Pre-image for CODING_RULES.md:78 non-invert fallback when post-update reload fails
	// (same-tx UpdateAndFind would be preferred once companyRepository joins DBOrTx).
	pre, preErr := s.repo.FindSingleton(ctx)
	if preErr != nil {
		slog.ErrorContext(ctx, "failed to get company before update", "error", preErr)
		return nil, apperrors.Wrap(preErr, "failed to get company before update")
	}
	if err := s.repo.Update(ctx, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update company", "error", err)
		return nil, apperrors.Wrap(err, "failed to update company")
	}
	result, err := s.repo.FindSingleton(ctx)
	if err != nil {
		// POC-02 / X-01: write already committed — do not invert success into 5xx.
		slog.ErrorContext(ctx, "failed to get company after update; returning applied fields on pre-image", "error", err)
		merged := *pre
		applyCompanyFields(&merged, fields)
		return &merged, nil
	}
	slog.InfoContext(ctx, "company updated", slog.Uint64("company_id", result.ID))
	return result, nil
}

func applyCompanyFields(c *model.Company, fields map[string]any) {
	if v, ok := fields[colCompanyName].(string); ok {
		c.Name = v
	}
	if v, ok := fields[colCompanyPostalCode].(string); ok {
		c.PostalCode = v
	}
	if v, ok := fields[colCompanyAddress].(string); ok {
		c.Address = v
	}
	if v, ok := fields[colCompanyPhoneNumber].(string); ok {
		c.PhoneNumber = v
	}
	if v, ok := fields[colCompanyFaxNumber].(string); ok {
		c.FaxNumber = v
	}
	if v, ok := fields[colCompanyEmail].(string); ok {
		c.Email = v
	}
	if v, ok := fields[colCompanyWebsite].(string); ok {
		c.Website = v
	}
	if v, ok := fields[colCompanyDirectorName].(string); ok {
		c.DirectorName = v
	}
	if v, ok := fields[colCompanyRegistrationNumber].(string); ok {
		c.RegistrationNumber = v
	}
	if v, ok := fields[colCompanyInvoiceRegistrationNumber].(string); ok {
		c.InvoiceRegistrationNumber = v
	}
	if v, ok := fields[colCompanyLogoURL].(string); ok {
		c.LogoURL = v
	}
}
