package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxStaffClinicAssignments = 50

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
	if err := s.validateOccupationOwnership(ctx, input.ClinicID, input.OccupationID); err != nil {
		return nil, err
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

func validateAndDedupeClinicAssignments(input *SetClinicAssignmentsInput) ([]uint64, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("clinic assignment input is required")
	}
	if input.StaffID == 0 {
		return nil, apperrors.WrapInvalidInput("staff_id is required")
	}
	if len(input.ClinicIDs) == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_ids must not be empty")
	}
	if len(input.ClinicIDs) > maxStaffClinicAssignments {
		return nil, apperrors.WrapInvalidInput("clinic_ids must contain at most 50 ids")
	}

	authorized := make(map[uint64]struct{}, len(input.AuthorizedClinicIDs))
	for _, clinicID := range input.AuthorizedClinicIDs {
		authorized[clinicID] = struct{}{}
	}

	seen := make(map[uint64]struct{}, len(input.ClinicIDs))
	clinicIDs := make([]uint64, 0, len(input.ClinicIDs))
	for _, clinicID := range input.ClinicIDs {
		if clinicID == 0 {
			return nil, apperrors.WrapInvalidInput("clinic_ids must contain positive ids")
		}
		if _, duplicate := seen[clinicID]; duplicate {
			continue
		}
		if !input.IsSystemAdmin {
			if _, ok := authorized[clinicID]; !ok {
				return nil, apperrors.WrapForbidden("cannot assign staff outside authorized clinics")
			}
		}
		seen[clinicID] = struct{}{}
		clinicIDs = append(clinicIDs, clinicID)
	}
	return clinicIDs, nil
}

func authorizeExistingClinicAssignments(
	input *SetClinicAssignmentsInput,
	assignments []model.StaffClinicAssignment,
) error {
	if input.IsSystemAdmin {
		return nil
	}
	authorized := make(map[uint64]struct{}, len(input.AuthorizedClinicIDs))
	for _, clinicID := range input.AuthorizedClinicIDs {
		authorized[clinicID] = struct{}{}
	}
	for _, assignment := range assignments {
		if _, ok := authorized[assignment.ClinicID]; !ok {
			return apperrors.WrapForbidden("cannot replace staff assignments outside authorized clinics")
		}
	}
	return nil
}

func removedClinicAssignmentIDs(
	existingAssignments []model.StaffClinicAssignment,
	targetClinicIDs []uint64,
) []uint64 {
	targets := make(map[uint64]struct{}, len(targetClinicIDs))
	for _, clinicID := range targetClinicIDs {
		targets[clinicID] = struct{}{}
	}

	removed := make([]uint64, 0, len(existingAssignments))
	for _, assignment := range existingAssignments {
		if _, retained := targets[assignment.ClinicID]; !retained {
			removed = append(removed, assignment.ClinicID)
		}
	}
	return removed
}

// SetClinicAssignments は認可・存在確認後にスタッフのクリニック割当を
// トランザクション内で差し替える。
func (s *staffService) SetClinicAssignments(ctx context.Context, input *SetClinicAssignmentsInput) error {
	clinicIDs, err := validateAndDedupeClinicAssignments(input)
	if err != nil {
		return err
	}
	if s.repo == nil || s.assignmentRepo == nil || s.shiftEntryRepo == nil ||
		s.clinicRepo == nil || s.tx == nil {
		return apperrors.WrapInternalServerError("clinic assignment dependencies are not configured")
	}

	if err := s.tx.WithTx(ctx, func(ctx context.Context) error {
		staff, lockStaffErr := s.repo.LockActiveByIDForUpdate(ctx, input.StaffID)
		if lockStaffErr != nil {
			return apperrors.Wrap(lockStaffErr, "failed to lock staff for clinic assignment replacement")
		}
		if staff == nil || staff.ID != input.StaffID {
			return apperrors.WrapInternalServerError("staff lock returned an invalid record")
		}

		existingAssignments, lockAssignmentsErr := s.assignmentRepo.LockActiveByStaff(ctx, input.StaffID)
		if lockAssignmentsErr != nil {
			return apperrors.Wrap(lockAssignmentsErr, "failed to lock existing clinic assignments")
		}
		for _, assignment := range existingAssignments {
			if assignment.StaffID != input.StaffID || assignment.ClinicID == 0 {
				return apperrors.WrapInternalServerError("clinic assignment lock returned an invalid record")
			}
		}
		if authorizeErr := authorizeExistingClinicAssignments(input, existingAssignments); authorizeErr != nil {
			return authorizeErr
		}

		// Lock every active target clinic before the destructive replacement.
		for _, clinicID := range clinicIDs {
			clinic, lockClinicErr := s.clinicRepo.LockActiveByID(ctx, clinicID)
			if lockClinicErr != nil {
				return apperrors.Wrap(lockClinicErr, "failed to lock clinic assignment target")
			}
			if clinic == nil || clinic.ID != clinicID {
				return apperrors.WrapInternalServerError("clinic lock returned an invalid record")
			}
		}

		for _, clinicID := range removedClinicAssignmentIDs(existingAssignments, clinicIDs) {
			hasShift, dependencyErr := s.shiftEntryRepo.ExistsByStaffID(ctx, clinicID, input.StaffID)
			if dependencyErr != nil {
				slog.ErrorContext(
					ctx,
					"failed to check shift dependency before removing clinic assignment",
					"error", dependencyErr,
					"staff_id", input.StaffID,
					"clinic_id", clinicID,
				)
				return apperrors.Wrap(
					dependencyErr,
					"failed to check shift dependency before removing clinic assignment",
				)
			}
			if hasShift {
				return apperrors.WrapConflict("シフトデータがあるクリニック所属は解除できません")
			}
		}

		if err := s.assignmentRepo.Delete(ctx, input.StaffID); err != nil {
			slog.ErrorContext(ctx, "failed to delete existing clinic assignments", "error", err, "staff_id", input.StaffID)
			return apperrors.Wrap(err, "failed to delete existing clinic assignments")
		}
		for i, clinicID := range clinicIDs {
			assignment := &model.StaffClinicAssignment{
				StaffID:  input.StaffID,
				ClinicID: clinicID,
				IsMain:   i == 0,
			}
			if err := s.assignmentRepo.RestoreOrCreate(ctx, assignment); err != nil {
				slog.ErrorContext(ctx, "failed to restore or create clinic assignment", "error", err, "staff_id", input.StaffID, "clinic_id", clinicID)
				return apperrors.Wrap(err, "failed to restore or create clinic assignment")
			}
		}
		if err := s.repo.UpdatePrimaryClinicID(ctx, input.StaffID, clinicIDs[0]); err != nil {
			slog.ErrorContext(ctx, "failed to update staff primary clinic", "error", err, "staff_id", input.StaffID, "clinic_id", clinicIDs[0])
			return apperrors.Wrap(err, "failed to update staff primary clinic")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to set clinic assignments")
	}
	slog.InfoContext(ctx, "clinic assignments updated", slog.Uint64("staff_id", input.StaffID), slog.Int("count", len(clinicIDs)))
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
