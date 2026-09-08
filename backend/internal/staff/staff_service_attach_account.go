package staff

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/seedlogin"
)

const attachAccountSecretBytes = 32

func (s *staffService) AttachAccount(
	ctx context.Context,
	clinicID, staffID uint64,
	input *AttachStaffAccountInput,
) (*model.Staff, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
	}
	if !input.IsSystemAdmin {
		return nil, apperrors.WrapForbidden("forbidden")
	}
	if s.credentialAudit == nil {
		return nil, apperrors.WrapInternalServerError("staff credential audit is not configured")
	}
	email := strings.TrimSpace(input.Email)
	if email == "" {
		return nil, apperrors.WrapInvalidInput("email is required")
	}
	if seedlogin.IsCatalogEmail(email) {
		return nil, apperrors.WrapInvalidInput("demo catalog email cannot be attached")
	}
	if err := validateStaffCredentialAudit(clinicID, staffID, input.CredentialAudit); err != nil {
		return nil, err
	}
	secret, err := generateAttachAccountSecret()
	if err != nil {
		return nil, err
	}
	if err := validatePassword(secret); err != nil {
		return nil, apperrors.WrapInternalServerError("generated account secret failed validation")
	}
	hashed, hashErr := bcrypt.GenerateFromPassword([]byte(secret), config.BcryptCost)
	if hashErr != nil {
		return nil, apperrors.Wrap(hashErr, "failed to hash staff account secret")
	}

	var attached *model.Staff
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		created, attachErr := s.attachAccountInTx(
			txCtx,
			clinicID,
			staffID,
			email,
			string(hashed),
			input,
		)
		attached = created
		return attachErr
	}); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "staff account attached",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("staff_id", staffID))
	return attached, nil
}

func (s *staffService) attachAccountInTx(
	txCtx context.Context,
	clinicID, staffID uint64,
	email, passwordHash string,
	input *AttachStaffAccountInput,
) (*model.Staff, error) {
	lockedStaff, err := s.repo.LockActiveByIDForUpdateInClinic(txCtx, clinicID, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to lock staff for account attach")
	}
	if lockedStaff == nil || lockedStaff.ID != staffID {
		return nil, apperrors.WrapInternalServerError("staff lock returned an invalid record")
	}
	if !lockedStaff.IsActive {
		return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("%d", staffID))
	}
	if lockedStaff.AccountID != nil {
		return nil, apperrors.WrapConflict("staff already has an account")
	}

	existing, err := s.accountRepo.FindByEmail(txCtx, email)
	if err != nil && !apperrors.IsNotFound(err) {
		return nil, apperrors.Wrap(err, "failed to check email uniqueness")
	}
	if existing != nil {
		return nil, apperrors.WrapAlreadyExists("account", email)
	}

	account := &model.Account{
		Email:         email,
		PasswordHash:  passwordHash,
		IsActive:      true,
		IsSystemAdmin: false,
	}
	if createErr := s.accountRepo.Create(txCtx, account); createErr != nil {
		if apperrors.IsAlreadyExists(createErr) {
			return nil, apperrors.WrapAlreadyExists("account", email)
		}
		return nil, apperrors.Wrap(createErr, "failed to create staff account")
	}
	if account.ID == 0 {
		return nil, apperrors.WrapInternalServerError("created account is missing an id")
	}
	if attachErr := s.repo.AttachAccountID(txCtx, clinicID, staffID, account.ID); attachErr != nil {
		return nil, attachErr
	}
	if auditErr := s.credentialAudit.LogEntryTx(
		txCtx,
		staffAccountAttachAuditEntry(*input.CredentialAudit, account.ID),
	); auditErr != nil {
		return nil, apperrors.Wrap(auditErr, "failed to write staff account attach audit")
	}
	updated, err := s.repo.FindByIDInClinic(txCtx, clinicID, staffID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get staff after account attach")
	}
	return updated, nil
}

func generateAttachAccountSecret() (string, error) {
	for range 8 {
		buf := make([]byte, attachAccountSecretBytes)
		if _, err := rand.Read(buf); err != nil {
			return "", apperrors.Wrap(err, "failed to generate staff account secret")
		}
		secret := hex.EncodeToString(buf)
		if secretHasLetterAndDigit(secret) {
			return secret, nil
		}
	}
	return "", apperrors.WrapInternalServerError("generated account secret is missing required character classes")
}

func secretHasLetterAndDigit(secret string) bool {
	var hasLetter, hasDigit bool
	for _, r := range secret {
		hasLetter = hasLetter || unicode.IsLetter(r)
		hasDigit = hasDigit || unicode.IsDigit(r)
	}
	return hasLetter && hasDigit
}
