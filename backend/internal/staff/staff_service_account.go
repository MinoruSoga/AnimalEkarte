package staff

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

func (s *staffService) FindByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error) {
	staff, err := s.repo.FindByAccountID(ctx, accountID)
	if err != nil {
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
		return nil, apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil {
		return nil, apperrors.WrapAlreadyExists("account", input.Email)
	}

	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), config.BcryptCost)
	if hashErr != nil {
		return nil, apperrors.Wrap(hashErr, "failed to hash password")
	}

	if err := validateStaffType(input.StaffType); err != nil {
		return nil, err
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
		created, err := s.createStaffWithAccountInTx(ctx, input, name, staffType, reservationVisible, string(hashed))
		staff = created
		return err
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to create staff")
	}

	slog.InfoContext(ctx, "staff with account created", slog.Uint64("clinic_id", input.ClinicID), slog.Uint64("staff_id", staff.ID))
	return staff, nil
}

func (s *staffService) createStaffWithAccountInTx(
	ctx context.Context,
	input *CreateStaffWithAccountInput,
	name string,
	staffType model.StaffType,
	reservationVisible bool,
	passwordHash string,
) (*model.Staff, error) {
	if err := s.lockOccupationOwnership(ctx, input.ClinicID, input.OccupationID); err != nil {
		return nil, err
	}
	account := &model.Account{
		Email:        input.Email,
		PasswordHash: passwordHash,
		IsActive:     true,
	}
	if createErr := s.accountRepo.Create(ctx, account); createErr != nil {
		return nil, apperrors.Wrap(createErr, "failed to create account")
	}
	staff := &model.Staff{
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
		return nil, apperrors.Wrap(err, "failed to create staff")
	}
	if input.ClinicID != 0 {
		if err := s.assignmentRepo.Create(ctx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: input.ClinicID,
			IsMain:   true,
		}); err != nil {
			return nil, apperrors.Wrap(err, "failed to assign staff to clinic")
		}
	}
	return staff, nil
}

func (s *staffService) VerifyClinicMembership(ctx context.Context, staffID, clinicID uint64) error {
	count, err := s.assignmentRepo.CountByStaffAndClinic(ctx, staffID, clinicID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify staff clinic membership")
	}
	if count == 0 {
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	return nil
}
