package service

import (
	"context"
	"log/slog"

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
	ListClinicsByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
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
	clinics, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list clinics")
	}
	return clinics, nil
}

func (s *clinicService) ListClinicsByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	clinics, err := s.repo.FindByStaffID(ctx, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list clinics by staff")
	}
	return clinics, nil
}

func (s *clinicService) GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	clinic, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get clinic")
	}
	return clinic, nil
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
		return nil, apperrors.Wrap(err, "failed to find clinic for update")
	}
	fields := buildClinicUpdateFields(input)
	if len(fields) == 0 {
		clinic, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to get clinic")
		}
		return clinic, nil
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update clinic")
	}
	// 更新後の完全なレコードを DB から取得して返す（created_at 等のサーバー管理フィールドを正しく反映）
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated clinic")
	}
	return updated, nil
}

func (s *clinicService) DeleteClinic(ctx context.Context, id uint64) error {
	// FK依存チェック: クリニックに関連するオーナーが存在する場合は削除を拒否
	ownerCount, err := s.repo.CountOwnersByClinicID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check owner dependencies")
	}
	if ownerCount > 0 {
		return apperrors.WrapConflict("飼主が紐付いているため削除できません。先に飼主を削除してください")
	}

	// FK依存チェック: クリニックに関連するスタッフが存在する場合は削除を拒否
	staffCount, err := s.repo.CountStaffByClinicID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check staff dependencies")
	}
	if staffCount > 0 {
		return apperrors.WrapConflict("スタッフが紐付いているため削除できません。先にスタッフを削除してください")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, "failed to delete clinic")
	}

	slog.InfoContext(ctx, "clinic deleted",
		slog.Uint64("clinic_id", id))

	return nil
}
