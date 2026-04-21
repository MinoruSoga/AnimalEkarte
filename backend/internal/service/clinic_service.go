package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateClinicInput はクリニック作成の入力DTO
type CreateClinicInput struct {
	Name               string
	PostalCode         string
	Address            string
	PhoneNumber        string
	FaxNumber          string
	RegistrationNumber string
	DirectorName       string
	Email              string
	Website            string
}

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

const (
	colClinicName               = "name"
	colClinicPostalCode         = "postal_code"
	colClinicAddress            = "address"
	colClinicPhoneNumber        = "phone_number"
	colClinicFaxNumber          = "fax_number"
	colClinicRegistrationNumber = "registration_number"
	colClinicDirectorName       = "director_name"
	colClinicEmail              = "email"
	colClinicWebsite            = "website"
	colClinicLogoURL            = "logo_url"
	colClinicIsActive           = "is_active"
	colClinicStandardTaxRate    = "standard_tax_rate"
	colClinicReducedTaxRate     = "reduced_tax_rate"
)

// buildClinicUpdate は PATCH 用 map を構築する。
// GORM のゼロ値スキップ問題を回避するために使用する。
// 税率が [0, 1] の範囲外の場合は error を返す。
func buildClinicUpdate(input *UpdateClinicInput) (map[string]any, error) {
	fields := make(map[string]any)
	if input.Name != nil {
		fields[colClinicName] = *input.Name
	}
	if input.PostalCode != nil {
		fields[colClinicPostalCode] = *input.PostalCode
	}
	if input.Address != nil {
		fields[colClinicAddress] = *input.Address
	}
	if input.PhoneNumber != nil {
		fields[colClinicPhoneNumber] = *input.PhoneNumber
	}
	if input.FaxNumber != nil {
		fields[colClinicFaxNumber] = *input.FaxNumber
	}
	if input.RegistrationNumber != nil {
		fields[colClinicRegistrationNumber] = *input.RegistrationNumber
	}
	if input.DirectorName != nil {
		fields[colClinicDirectorName] = *input.DirectorName
	}
	if input.Email != nil {
		fields[colClinicEmail] = *input.Email
	}
	if input.Website != nil {
		fields[colClinicWebsite] = *input.Website
	}
	if input.LogoURL != nil {
		fields[colClinicLogoURL] = *input.LogoURL
	}
	if input.IsActive != nil {
		fields[colClinicIsActive] = *input.IsActive
	}
	if input.StandardTaxRate != nil {
		r := *input.StandardTaxRate
		if r < 0 || r > 1 {
			return nil, apperrors.WrapInvalidInput("standard_tax_rate must be between 0 and 1")
		}
		fields[colClinicStandardTaxRate] = r
	}
	if input.ReducedTaxRate != nil {
		r := *input.ReducedTaxRate
		if r < 0 || r > 1 {
			return nil, apperrors.WrapInvalidInput("reduced_tax_rate must be between 0 and 1")
		}
		fields[colClinicReducedTaxRate] = r
	}
	return fields, nil
}

type ClinicService interface {
	ListClinics(ctx context.Context) ([]model.Clinic, error)
	ListClinicsByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id uint64) (*model.Clinic, error)
	CreateClinic(ctx context.Context, input *CreateClinicInput) (*model.Clinic, error)
	UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error)
	DeleteClinic(ctx context.Context, id uint64) error
}

type clinicService struct {
	repo                repository.ClinicRepository
	permissionGroupRepo repository.PermissionGroupRepository
}

func NewClinicService(repo repository.ClinicRepository, permissionGroupRepo repository.PermissionGroupRepository) ClinicService {
	return &clinicService{repo: repo, permissionGroupRepo: permissionGroupRepo}
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

func (s *clinicService) CreateClinic(ctx context.Context, input *CreateClinicInput) (*model.Clinic, error) {
	company, err := s.repo.GetCompany(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get company")
	}
	clinic := &model.Clinic{
		CompanyID:          company.ID,
		Name:               input.Name,
		PostalCode:         input.PostalCode,
		Address:            input.Address,
		PhoneNumber:        input.PhoneNumber,
		FaxNumber:          input.FaxNumber,
		RegistrationNumber: input.RegistrationNumber,
		DirectorName:       input.DirectorName,
		Email:              input.Email,
		Website:            input.Website,
		IsActive:           true,
	}

	if err := s.repo.Create(ctx, clinic); err != nil {
		return nil, apperrors.Wrap(err, "failed to create clinic")
	}

	// Initialize default permission groups: "執行" (executive) and "一般" (general)
	defaultGroups := []struct {
		name        string
		description string
		sortOrder   int
	}{
		{name: "執行", description: "執行権限", sortOrder: 1},
		{name: "一般", description: "一般スタッフ権限", sortOrder: 2},
	}

	for _, groupDef := range defaultGroups {
		group := &model.PermissionGroup{
			ClinicID:    clinic.ID,
			Name:        groupDef.name,
			Description: groupDef.description,
			IsActive:    true,
			SortOrder:   groupDef.sortOrder,
		}
		if err := s.permissionGroupRepo.Create(ctx, group); err != nil {
			return nil, apperrors.Wrap(err, "failed to create default permission group")
		}
		slog.InfoContext(ctx, "default permission group created",
			slog.Uint64("clinic_id", clinic.ID),
			slog.String("group_name", group.Name),
			slog.Uint64("group_id", group.ID))
	}

	slog.InfoContext(ctx, "clinic created",
		slog.Uint64("clinic_id", clinic.ID))
	return clinic, nil
}

func (s *clinicService) UpdateClinic(ctx context.Context, id uint64, input *UpdateClinicInput) (*model.Clinic, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	// 存在確認（NotFound を早期返却）
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to find clinic for update")
	}
	fields, err := buildClinicUpdate(input)
	if err != nil {
		return nil, err
	}
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
	slog.InfoContext(ctx, "clinic updated",
		slog.Uint64("clinic_id", id))
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
