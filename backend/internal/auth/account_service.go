package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountService provides login-account lookup.
type AccountService interface {
	FindByEmail(ctx context.Context, email string) (*model.Account, error)
	GetByID(ctx context.Context, id uint64) (*model.Account, error)
}

// AccountPasswordChanger is the atomic credential-change capability consumed
// only by the authenticated password endpoint. It remains separate from the
// broader compatibility interface while BE9 consumers are cut over.
type AccountPasswordChanger interface {
	ChangePassword(
		ctx context.Context,
		accountID uint64,
		currentPassword, newPassword string,
		audit CredentialMutationAudit,
	) error
}

type accountService struct {
	repo        AccountRepository
	resetTokens PasswordResetTokenInvalidator
	transactor  Transactor
	audit       CredentialAuditTxLogger
}

// NewAccountService constructs the auth account use case.
func NewAccountService(repo AccountRepository) AccountService {
	return &accountService{repo: repo}
}

// NewAccountServiceWithCredentialInvalidation constructs the credential
// mutation capability. Every successful password change revokes outstanding
// reset tokens in the same account-first transaction.
func NewAccountServiceWithCredentialInvalidation(
	repo AccountRepository,
	resetTokens PasswordResetTokenInvalidator,
	transactor Transactor,
) AccountService {
	return NewAccountServiceWithCredentialAudit(
		repo,
		resetTokens,
		transactor,
		nil,
	)
}

// NewAccountServiceWithCredentialAudit constructs the credential mutation use
// case with a transaction-bound success audit. A nil audit logger is retained
// only for read-only compatibility and causes ChangePassword to fail closed.
func NewAccountServiceWithCredentialAudit(
	repo AccountRepository,
	resetTokens PasswordResetTokenInvalidator,
	transactor Transactor,
	audit CredentialAuditTxLogger,
) AccountService {
	return &accountService{
		repo:        repo,
		resetTokens: resetTokens,
		transactor:  transactor,
		audit:       audit,
	}
}

func (s *accountService) FindByEmail(ctx context.Context, email string) (*model.Account, error) {
	account, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find account by email", "error_type", fmt.Sprintf("%T", err))
		return nil, apperrors.Wrap(err, "failed to find account by email")
	}
	return account, nil
}

func (s *accountService) GetByID(ctx context.Context, id uint64) (*model.Account, error) {
	account, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(err, fmt.Sprintf("failed to get account: %d", id))
	}
	return account, nil
}

// ChangePassword verifies the caller's current password and atomically replaces
// the exact hash that was verified. A concurrent winner is indistinguishable
// from an invalid current password.
func (s *accountService) ChangePassword(
	ctx context.Context,
	accountID uint64,
	currentPassword, newPassword string,
	audit CredentialMutationAudit,
) error {
	if s.repo == nil ||
		s.resetTokens == nil ||
		s.transactor == nil ||
		s.audit == nil {
		return apperrors.WrapInternalServerError(
			"password change dependencies are not configured",
		)
	}
	if err := validateCredentialMutationAudit(
		audit,
		model.AuditActionAuthPasswordChange,
	); err != nil {
		return err
	}
	account, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return apperrors.Wrap(err, fmt.Sprintf("failed to get account: %d", accountID))
	}
	if account == nil || account.ID != accountID {
		return apperrors.WrapInternalServerError("account lookup returned an invalid result")
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(account.PasswordHash),
		[]byte(currentPassword),
	); err != nil {
		return apperrors.WrapUnauthorized("現在のパスワードが正しくありません")
	}

	newHash, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		config.BcryptCost,
	)
	if err != nil {
		return apperrors.Wrap(err, "failed to hash password")
	}

	updatedAt := nextAccountSessionEpoch(account.UpdatedAt, time.Now())
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		swapped, swapErr := s.repo.CompareAndSwapPasswordHash(
			txCtx,
			accountID,
			account.PasswordHash,
			string(newHash),
			updatedAt,
		)
		if swapErr != nil {
			return apperrors.Wrap(swapErr, "failed to update password hash")
		}
		if !swapped {
			return apperrors.WrapUnauthorized("現在のパスワードが正しくありません")
		}
		if invalidateErr := s.resetTokens.DeleteByAccountID(
			txCtx,
			accountID,
		); invalidateErr != nil {
			return apperrors.Wrap(
				invalidateErr,
				"failed to revoke password reset tokens",
			)
		}
		if auditErr := s.audit.LogEntryTx(
			txCtx,
			credentialAuditEntry(audit, accountID),
		); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to write password change audit",
			)
		}
		return nil
	})
	if err != nil {
		slog.ErrorContext(
			ctx,
			"password change transaction failed",
			"error_type", fmt.Sprintf("%T", err),
			"account_id", accountID,
		)
		return apperrors.Wrap(err, "failed to change password")
	}

	slog.InfoContext(ctx, "password hash updated",
		slog.Uint64("account_id", accountID))
	return nil
}

func nextAccountSessionEpoch(previous, now time.Time) time.Time {
	minimum := previous.Add(time.Microsecond)
	if now.Before(minimum) {
		return minimum
	}
	return now
}
