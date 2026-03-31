package service

import (
	"context"
	apperrors "github.com/animal-ekarte/backend/internal/errors"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// UpdateClinicInput はクリニック更新の入力DTO（nil = 未指定）
type UpdateClinicInput struct {
	Name               *string
	PostalCode         *string
	Address            *string
	PhoneNumber        *string
	FaxNumber          *string
	RegistrationNumber *string
	DirectorName       *string
	Email              *string
	Website            *string
	LogoURL            *string
	IsActive           *bool
	StandardTaxRate    *float64
	ReducedTaxRate     *float64
}

// buildClinicUpdateFields は PATCH 用 map を構築する。
// GORM のゼロ値スキップ問題を回避するために使用する。
func buildClinicUpdateFields(input *UpdateClinicInput) map[string]any {
	fields := make(map[string]any)
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
	if input.RegistrationNumber != nil {
		fields["registration_number"] = *input.RegistrationNumber
	}
	if input.DirectorName != nil {
		fields["director_name"] = *input.DirectorName
	}
	if input.Email != nil {
		fields["email"] = *input.Email
	}
	if input.Website != nil {
		fields["website"] = *input.Website
	}
	if input.LogoURL != nil {
		fields["logo_url"] = *input.LogoURL
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	if input.StandardTaxRate != nil {
		r := *input.StandardTaxRate
		if r > 0 && r <= 1 {
			fields["standard_tax_rate"] = r
		}
	}
	if input.ReducedTaxRate != nil {
		r := *input.ReducedTaxRate
		if r > 0 && r <= 1 {
			fields["reduced_tax_rate"] = r
		}
	}
	return fields
}

type ClinicService interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error)
	CreateClinic(ctx context.Context, clinic *model.Clinic) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error)
	DeleteClinic(ctx context.Context, id uint64) error
}

type clinicService struct {
	repo repository.ClinicRepository
}

func NewClinicService(repo repository.ClinicRepository) ClinicService {
	return &clinicService{repo: repo}
}

func (s *clinicService) ListClinics(ctx context.Context) ([]model.Clinic, error) {
	return s.repo.FindAll(ctx)
}

func (s *clinicService) GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *clinicService) CreateClinic(ctx context.Context, clinic *model.Clinic) (*model.Clinic, error) {
	// company はシングルトンなので自動設定する
	company, err := s.repo.GetCompany(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get company")
	}
	clinic.CompanyID = company.ID
	if err := s.repo.Create(ctx, clinic); err != nil {
		return nil, apperrors.Wrap(err, "failed to create clinic")
	}
	return clinic, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error) {
	// 存在確認（NotFound を早期返却）
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}
	fields := buildClinicUpdateFields(input)
	if len(fields) == 0 {
		return s.repo.FindByID(ctx, id)
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update clinic")
	}
	// 更新後の完全なレコードを DB から取得して返す（created_at 等のサーバー管理フィールドを正しく反映）
	return s.repo.FindByID(ctx, id)
}

func (s *clinicService) DeleteClinic(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
