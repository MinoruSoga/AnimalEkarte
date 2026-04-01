// Package service provides business logic implementations for Staff entity.
package service

import (
	"context"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- StaffService ----

// CreateStaffInput はスタッフ登録時の入力データ。
// スタッフ情報とシステムアカウント情報を同時に受け取る。
type CreateStaffInput struct {
	// Staff fields
	ClinicID      uint64
	Name          string
	StaffRole     model.StaffRole
	LicenseNumber string
	JobTitleID    *uint64
	SortOrder     int
	// Account fields
	Email    string
	Password string
}

// UpdateStaffInput はスタッフ部分更新の入力DTO。nil = 未送信フィールド。
type UpdateStaffInput struct {
	Name          *string
	StaffRole     *model.StaffRole
	LicenseNumber *string
	JobTitleID    *uint64
	SortOrder     *int
	IsActive      *bool
}

type StaffService interface {
	List(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error)
	GetByID(ctx context.Context, id uint64) (*model.Staff, error)
	// CreateWithAccount はスタッフとシステムアカウントをアトミックに作成する。
	CreateWithAccount(ctx context.Context, input *CreateStaffInput) (*model.Staff, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
}

type staffService struct {
	repo            repository.StaffRepository
	reservationRepo repository.ReservationRepository
	shiftEntryRepo  repository.ShiftEntryRepository
}

func NewStaffService(repo repository.StaffRepository, reservationRepo repository.ReservationRepository, shiftEntryRepo repository.ShiftEntryRepository) StaffService {
	return &staffService{repo: repo, reservationRepo: reservationRepo, shiftEntryRepo: shiftEntryRepo}
}

func (s *staffService) List(ctx context.Context, clinicID uint64, role *string, page, limit int) ([]model.Staff, int64, error) {
	staff, total, err := s.repo.FindAll(ctx, clinicID, role, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list staff")
	}
	return staff, total, nil
}

func (s *staffService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staff")
	}
	return staff, nil
}

func (s *staffService) CreateWithAccount(ctx context.Context, input *CreateStaffInput) (*model.Staff, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, err
	}
	input.Name = strings.TrimSpace(input.Name)

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Wrap(err, "hash password")
	}

	staff := &model.Staff{
		ClinicID:      input.ClinicID,
		Name:          input.Name,
		StaffRole:     input.StaffRole,
		LicenseNumber: input.LicenseNumber,
		JobTitleID:    input.JobTitleID,
		SortOrder:     input.SortOrder,
		IsActive:      true,
	}

	account := &model.UserAccount{
		Email:        input.Email,
		DisplayName:  input.Name,
		UserType:     model.UserTypeStaff,
		Status:       model.AccountStatusActive,
		PasswordHash: string(hash),
	}

	membership := &model.UserClinicMembership{
		ClinicID: input.ClinicID,
		IsMain:   true,
	}

	if err := s.repo.CreateWithAccount(ctx, staff, account, membership); err != nil {
		return nil, apperrors.Wrap(err, "failed to create staff with account")
	}

	return staff, nil
}

func (s *staffService) Update(ctx context.Context, clinicID, id uint64, input *UpdateStaffInput) (*model.Staff, error) {
	if input.Name != nil {
		if err := validateRequiredName(*input.Name); err != nil {
			return nil, err
		}
		trimmed := strings.TrimSpace(*input.Name)
		input.Name = &trimmed
	}
	if input.StaffRole != nil {
		if err := validateStaffRole(*input.StaffRole); err != nil {
			return nil, err
		}
	}
	fields := buildStaffUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, id, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update staff")
	}
	slog.InfoContext(ctx, "staff updated", slog.Uint64("staff_id", id))
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get updated staff")
	}
	return updated, nil
}

func buildStaffUpdateFields(input *UpdateStaffInput) map[string]any {
	fields := map[string]any{}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.StaffRole != nil {
		fields["staff_role"] = *input.StaffRole
	}
	if input.LicenseNumber != nil {
		fields["license_number"] = *input.LicenseNumber
	}
	if input.JobTitleID != nil {
		fields["job_title_id"] = *input.JobTitleID
	}
	if input.SortOrder != nil {
		fields["sort_order"] = *input.SortOrder
	}
	if input.IsActive != nil {
		fields["is_active"] = *input.IsActive
	}
	return fields
}

func (s *staffService) Delete(ctx context.Context, clinicID, id uint64) error {
	reservationExists, err := s.reservationRepo.ExistsByStaffID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependency")
	}
	if reservationExists {
		return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
	}
	shiftExists, err := s.shiftEntryRepo.ExistsByStaffID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check shift dependency")
	}
	if shiftExists {
		return apperrors.WrapConflict("このスタッフはシフト・予約データで使用中のため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete staff")
	}
	return nil
}

func (s *staffService) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("ids must not be empty")
	}
	if err := s.repo.Reorder(ctx, clinicID, ids); err != nil {
		return apperrors.Wrap(err, "failed to reorder staff")
	}
	return nil
}
