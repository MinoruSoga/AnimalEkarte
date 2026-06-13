package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *staffService) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByAccountID(ctx, accountID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find staff by account id", "error", err, "id", accountID)
		return nil, apperrors.Wrap(err, "failed to find staff by account id")
	}
	return staff, nil
}
func (s *staffService) CreateWithAccount(ctx context.Context, input *CreateStaffWithAccountInput) (*model.Staff, error) {
	if err := validateRequiredName(input.Name); err != nil {
		return nil, apperrors.Wrap(err, "failed to validate required name")
	}
	if input.ClinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id is required")
	}
	name := strings.TrimSpace(input.Name)

	// パスワードバリデーション（email が指定されている場合は必須）
	if input.Email != "" && input.Password == "" {
		return nil, apperrors.WrapInvalidInput("password is required when email is provided")
	}
	if input.Password != "" {
		if err := validatePassword(input.Password); err != nil {
			return nil, apperrors.Wrap(err, "failed to validate password")
		}
	}

	// email 重複チェック: FindByEmail が NotFound 以外のエラーを返した場合は伝播する
	existing, err := s.accountRepo.FindByEmail(ctx, input.Email)
	if err != nil && !apperrors.IsNotFound(err) {
		slog.ErrorContext(ctx, "failed to check email uniqueness", "error", err, "clinic_id", input.ClinicID)
		return nil, apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil {
		return nil, apperrors.WrapAlreadyExists("account", input.Email)
	}

	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), config.BcryptCost)
	if hashErr != nil {
		return nil, apperrors.Wrap(hashErr, "failed to hash password")
	}

	staffType := model.StaffType(input.StaffType)
	if staffType == "" {
		staffType = model.StaffTypeDoctor
	}

	reservationVisible := true
	if input.ReservationVisible != nil {
		reservationVisible = *input.ReservationVisible
	}

	var staff *model.Staff
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		account := &model.Account{
			Email:        input.Email,
			PasswordHash: string(hashed),
			IsActive:     true,
		}
		if createErr := s.accountRepo.Create(ctx, account); createErr != nil {
			slog.ErrorContext(ctx, "failed to create account", "error", createErr)
			return apperrors.Wrap(createErr, "failed to create account")
		}
		staff = &model.Staff{
			ClinicID:               input.ClinicID,
			Name:                   name,
			LicenseNumber:          input.LicenseNumber,
			OccupationID:           input.OccupationID,
			SortOrder:              input.SortOrder,
			IsActive:               true,
			AccountID:              &account.ID,
			StaffType:              staffType,
			ReservationDisplayName: input.ReservationDisplayName,
			ReservationVisible:     reservationVisible,
			ReservationComment:     input.ReservationComment,
			ReservationImageURL:    input.ReservationImageURL,
		}
		if err := s.repo.Create(ctx, staff); err != nil {
			slog.ErrorContext(ctx, "failed to create staff", "error", err)
			return apperrors.Wrap(err, "failed to create staff")
		}
		if input.ClinicID != 0 {
			if err := s.assignmentRepo.Create(ctx, &model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: input.ClinicID,
				IsMain:   true,
			}); err != nil {
				slog.ErrorContext(ctx, "failed to assign staff to clinic", "error", err, "clinic_id", input.ClinicID)
				return apperrors.Wrap(err, "failed to assign staff to clinic")
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create staff", "error", err)
		return nil, apperrors.Wrap(err, "failed to create staff")
	}

	slog.InfoContext(ctx, "staff with account created", slog.Uint64("clinic_id", input.ClinicID), slog.Uint64("staff_id", staff.ID))
	return staff, nil
}

// UpdatePassword はスタッフに紐づくアカウントのパスワードを更新する。
func (s *staffService) UpdatePassword(ctx context.Context, accountID uint64, newPassword string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.BcryptCost)
	if err != nil {
		return apperrors.Wrap(err, "failed to hash password")
	}
	if err := s.accountRepo.Update(ctx, accountID, map[string]any{"password_hash": string(hashed)}); err != nil {
		slog.ErrorContext(ctx, "failed to update account password", "error", err, "id", accountID)
		return apperrors.Wrap(err, "failed to update account password")
	}
	slog.InfoContext(ctx, "password updated", slog.Uint64("account_id", accountID))
	return nil
}

// SetClinicAssignments はスタッフのクリニック割当をトランザクション内で差し替える。
func (s *staffService) SetClinicAssignments(ctx context.Context, staffID uint64, clinicIDs []uint64) error {
	if len(clinicIDs) == 0 {
		return apperrors.WrapInvalidInput("clinic_ids must not be empty")
	}
	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.assignmentRepo.Delete(ctx, staffID); err != nil {
			slog.ErrorContext(ctx, "failed to delete existing clinic assignments", "error", err, "staff_id", staffID)
			return apperrors.Wrap(err, "failed to delete existing clinic assignments")
		}
		for i, clinicID := range clinicIDs {
			assignment := &model.StaffClinicAssignment{
				StaffID:  staffID,
				ClinicID: clinicID,
				IsMain:   i == 0,
			}
			if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
				slog.ErrorContext(ctx, "failed to create clinic assignment", "error", err, "staff_id", staffID, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to create clinic assignment")
			}
		}
		if err := s.repo.UpdatePrimaryClinicID(ctx, staffID, clinicIDs[0]); err != nil {
			slog.ErrorContext(ctx, "failed to update staff primary clinic", "error", err, "staff_id", staffID, "clinic_id", clinicIDs[0])
			return apperrors.Wrap(err, "failed to update staff primary clinic")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to set clinic assignments")
	}
	slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", staffID), slog.Int("count", len(clinicIDs)))
	return nil
}
func (s *staffService) VerifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error {
	count, err := s.assignmentRepo.CountByStaffAndClinic(ctx, staffID, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify staff clinic membership", "error", err, "id", staffID, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to verify staff clinic membership")
	}
	if count == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	return nil
}
