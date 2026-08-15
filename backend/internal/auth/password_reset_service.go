package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const (
	passwordResetTokenExpiry     = 30 * time.Minute
	passwordResetReissueCooldown = 5 * time.Minute
	passwordResetMailWorkers     = 4
	passwordResetMailSendTimeout = 30 * time.Second
	passwordResetCleanupTimeout  = 5 * time.Second
)

// PasswordResetService provides password-reset request and completion use cases.
type PasswordResetService interface {
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken, newPassword string) error
	Wait()
}

// PasswordResetCompletion identifies the credential subject after a successful
// reset without exposing the raw token or any credential material.
type PasswordResetCompletion struct {
	AccountID uint64
}

// PasswordResetCompletionService accepts trusted HTTP audit metadata while the
// use case resolves the subject and persists the audit inside its transaction.
type PasswordResetCompletionService interface {
	ResetPasswordWithResult(
		ctx context.Context,
		rawToken, newPassword string,
		audit CredentialMutationAudit,
	) (*PasswordResetCompletion, error)
}

// PasswordResetConfig contains SMTP and frontend URL settings.
type PasswordResetConfig struct {
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
	FrontendURL string
}

// PasswordResetMailSender is the auth-owned transport boundary for reset mail.
type PasswordResetMailSender func(
	ctx context.Context,
	cfg *PasswordResetConfig,
	from, to string,
	message []byte,
) error

// Transactor owns the transaction boundary required for single-use reset token consumption.
type Transactor interface {
	WithTx(ctx context.Context, fn func(context.Context) error) error
}

// PasswordCredentialUpdater is the credential-only persistence capability used
// after a caller locks the account row. Implementations must suppress SQL
// parameter logging because newHash is authentication material.
type PasswordCredentialUpdater interface {
	UpdatePasswordHash(
		ctx context.Context,
		accountID uint64,
		newHash string,
		updatedAt time.Time,
	) error
}

type passwordResetService struct {
	cfg               *PasswordResetConfig
	accountRepo       AccountRepository
	credentialUpdater PasswordCredentialUpdater
	tokenRepo         PasswordResetTokenRepository
	sendMail          PasswordResetMailSender
	transactor        Transactor
	auditSubject      CredentialAuditSubjectResolver
	audit             CredentialAuditTxLogger
	comparePassword   passwordCompareFunc
	responseTiming    forgotPasswordResponseTiming
	now               func() time.Time
	mailSlotsOnce     sync.Once
	mailWorkerSlots   chan struct{}
	mailGate          backgroundWorkGate
}

// NewPasswordResetService preserves the transitional constructor. ResetPassword
// fails closed until production composition supplies a Transactor through
// NewPasswordResetServiceWithTransactor.
func NewPasswordResetService(
	cfg *PasswordResetConfig,
	accountRepo AccountRepository,
	tokenRepo PasswordResetTokenRepository,
	sendMail PasswordResetMailSender,
) PasswordResetService {
	return NewPasswordResetServiceWithTransactor(cfg, accountRepo, tokenRepo, sendMail, nil)
}

// NewPasswordResetServiceWithTransactor constructs the password-reset use case
// with the transaction boundary required for strict single-use token consumption.
func NewPasswordResetServiceWithTransactor(
	cfg *PasswordResetConfig,
	accountRepo AccountRepository,
	tokenRepo PasswordResetTokenRepository,
	sendMail PasswordResetMailSender,
	transactor Transactor,
) PasswordResetService {
	return NewPasswordResetServiceWithCredentialAudit(
		cfg,
		accountRepo,
		tokenRepo,
		sendMail,
		transactor,
		nil,
		nil,
	)
}

// NewPasswordResetServiceWithCredentialAudit constructs password reset with
// transaction-bound credential-subject resolution and success auditing.
func NewPasswordResetServiceWithCredentialAudit(
	cfg *PasswordResetConfig,
	accountRepo AccountRepository,
	tokenRepo PasswordResetTokenRepository,
	sendMail PasswordResetMailSender,
	transactor Transactor,
	auditSubject CredentialAuditSubjectResolver,
	audit CredentialAuditTxLogger,
) PasswordResetService {
	credentialUpdater, _ := accountRepo.(PasswordCredentialUpdater)
	service := &passwordResetService{
		cfg:               cfg,
		accountRepo:       accountRepo,
		credentialUpdater: credentialUpdater,
		tokenRepo:         tokenRepo,
		sendMail:          sendMail,
		transactor:        transactor,
		auditSubject:      auditSubject,
		audit:             audit,
		comparePassword:   defaultPasswordCompare,
		responseTiming:    forgotPasswordResponseTiming{}.withDefaults(),
		now:               time.Now,
	}
	service.ensureMailWorkerSlots()
	return service
}

func (s *passwordResetService) ForgotPassword(
	ctx context.Context,
	email string,
) (resultErr error) {
	timing := s.responseTiming.withDefaults()
	startedAt := timing.now()
	minimumDuration := forgotPasswordResponseFloor +
		boundedForgotPasswordJitter(timing.jitter())
	defer func() {
		elapsed := timing.now().Sub(startedAt)
		remaining := minimumDuration - elapsed
		if remaining < 0 {
			remaining = 0
		}
		if waitErr := timing.sleep(ctx, remaining); waitErr != nil && resultErr == nil {
			resultErr = apperrors.Wrap(
				waitErr,
				"forgot password response floor interrupted",
			)
		}
	}()

	performDummyPasswordComparison(s.comparePassword, dummyPasswordCandidate)

	if s.transactor == nil {
		return apperrors.WrapInternalServerError("password reset transaction is not configured")
	}

	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			slog.InfoContext(ctx, "forgot password: account not found, silently returning")
			return nil
		}
		slog.ErrorContext(ctx, "failed to find account", "error_type", fmt.Sprintf("%T", err))
		return apperrors.Wrap(err, "failed to find account")
	}
	if account == nil {
		slog.ErrorContext(ctx, "forgot password account lookup returned no account")
		return apperrors.WrapInternalServerError("password reset account lookup failed")
	}

	if !s.tryAcquireMailWorker() {
		slog.ErrorContext(
			ctx,
			"password reset mail queue is full",
			slog.Uint64("account_id", account.ID),
		)
		return nil
	}
	mailSlotOwned := true
	mailRegistrationOwned := false
	defer func() {
		if mailSlotOwned {
			s.releaseMailWorker()
		}
		if mailRegistrationOwned {
			s.mailGate.Done()
		}
	}()
	if !s.mailGate.Register() {
		slog.ErrorContext(
			ctx,
			"password reset mail registration is closed",
			slog.Uint64("account_id", account.ID),
		)
		return nil
	}
	mailRegistrationOwned = true

	rawToken, tokenHash, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate token", "error_type", fmt.Sprintf("%T", err))
		return apperrors.Wrap(err, "failed to generate token")
	}

	now := s.currentTime()
	resetToken := &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(passwordResetTokenExpiry),
	}
	issueSuppressed := false
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		lockedAccount, lockErr := s.accountRepo.FindByIDForUpdate(txCtx, account.ID)
		if lockErr != nil {
			return apperrors.Wrap(lockErr, "failed to lock password reset account")
		}
		if lockedAccount == nil || lockedAccount.ID != account.ID {
			return apperrors.WrapUnauthorized("password reset account is unavailable")
		}
		latestToken, latestErr := s.tokenRepo.FindLatestByAccountIDForUpdate(
			txCtx,
			account.ID,
		)
		if latestErr != nil && !apperrors.IsNotFound(latestErr) {
			return apperrors.Wrap(
				latestErr,
				"failed to inspect existing password reset token",
			)
		}
		now := s.currentTime()
		if latestErr == nil && activeRecentResetToken(latestToken, now) {
			issueSuppressed = true
			return nil
		}
		resetToken.CreatedAt = nextAccountSessionEpoch(
			lockedAccount.UpdatedAt,
			now,
		)
		if deleteErr := s.tokenRepo.DeleteByAccountID(txCtx, account.ID); deleteErr != nil {
			return apperrors.Wrap(deleteErr, "failed to clean up existing tokens")
		}
		if createErr := s.tokenRepo.Create(txCtx, resetToken); createErr != nil {
			return apperrors.Wrap(createErr, "failed to create reset token")
		}
		return nil
	}); err != nil {
		slog.ErrorContext(
			ctx,
			"failed to issue password reset token",
			"error_type", fmt.Sprintf("%T", err),
			"account_id", account.ID,
		)
		return apperrors.Wrap(err, "password reset token transaction failed")
	}
	if issueSuppressed {
		slog.InfoContext(
			ctx,
			"password reset token reissue suppressed",
			slog.Uint64("account_id", account.ID),
		)
		return nil
	}

	resetURL := buildPasswordResetURL(s.cfg.FrontendURL, rawToken)
	mailSlotOwned = false
	mailRegistrationOwned = false
	sharedkernel.GoSafe("password reset email", func() { //nolint:contextcheck // request cancellation must not abort an accepted reset mail
		defer s.mailGate.Done()
		defer s.releaseMailWorker()
		backgroundContext, cancel := context.WithTimeout(context.Background(), passwordResetMailSendTimeout) //nolint:gosec // detached accepted work with bounded lifetime
		defer cancel()
		if sendErr := s.sendResetEmail(backgroundContext, email, resetURL); sendErr != nil {
			slog.ErrorContext(
				backgroundContext,
				"failed to send password reset email",
				slog.Uint64("account_id", account.ID),
				slog.String("error_type", fmt.Sprintf("%T", sendErr)),
			)
			cleanupContext, cleanupCancel := context.WithTimeout(
				context.Background(),
				passwordResetCleanupTimeout,
			)
			defer cleanupCancel()
			if cleanupErr := s.tokenRepo.DeleteIssued(
				cleanupContext,
				resetToken.ID,
				tokenHash,
			); cleanupErr != nil {
				slog.ErrorContext(
					cleanupContext,
					"failed to remove undelivered password reset token",
					slog.Uint64("account_id", account.ID),
					slog.Uint64("token_id", resetToken.ID),
					slog.String("error_type", fmt.Sprintf("%T", cleanupErr)),
				)
			}
		}
	})

	slog.InfoContext(ctx, "password reset token issued",
		slog.Uint64("account_id", account.ID))
	return nil
}

func (s *passwordResetService) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *passwordResetService) ensureMailWorkerSlots() {
	s.mailSlotsOnce.Do(func() {
		s.mailWorkerSlots = make(chan struct{}, passwordResetMailWorkers)
	})
}

func (s *passwordResetService) tryAcquireMailWorker() bool {
	s.ensureMailWorkerSlots()
	select {
	case s.mailWorkerSlots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *passwordResetService) releaseMailWorker() {
	<-s.mailWorkerSlots
}

func activeRecentResetToken(token *model.PasswordResetToken, now time.Time) bool {
	if token == nil || token.CreatedAt.IsZero() || token.ExpiresAt.IsZero() {
		return false
	}
	return token.ExpiresAt.After(now) &&
		token.CreatedAt.Add(passwordResetReissueCooldown).After(now)
}

func buildPasswordResetURL(frontendURL, rawToken string) string {
	fragment := url.Values{"token": []string{rawToken}}.Encode()
	return fmt.Sprintf("%s/reset-password#%s", frontendURL, fragment)
}

func (s *passwordResetService) ResetPassword(
	ctx context.Context,
	rawToken, newPassword string,
) error {
	_, err := s.ResetPasswordWithResult(
		ctx,
		rawToken,
		newPassword,
		CredentialMutationAudit{
			ActorType: model.AuditActorTypeSystem,
			Action:    model.AuditActionAuthPasswordReset,
		},
	)
	return err
}

func (s *passwordResetService) ResetPasswordWithResult(
	ctx context.Context,
	rawToken, newPassword string,
	audit CredentialMutationAudit,
) (*PasswordResetCompletion, error) {
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError(
			"password reset transaction is not configured",
		)
	}
	if s.credentialUpdater == nil {
		return nil, apperrors.WrapInternalServerError(
			"password reset credential update is not configured",
		)
	}
	if s.auditSubject == nil || s.audit == nil {
		return nil, apperrors.WrapInternalServerError(
			"password reset credential audit is not configured",
		)
	}
	if err := validateCredentialMutationAudit(
		audit,
		model.AuditActionAuthPasswordReset,
	); err != nil {
		return nil, err
	}

	tokenHash := hashToken(rawToken)
	candidate, err := s.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.WrapInvalidInput("invalid or expired token")
		}
		return nil, apperrors.Wrap(err, "failed to find password reset token")
	}
	if candidate == nil {
		return nil, apperrors.WrapInvalidInput("invalid or expired token")
	}

	if resetTokenExpired(candidate.ExpiresAt, time.Now()) {
		expired, consumeErr := s.consumeResetToken(
			ctx,
			candidate.AccountID,
			tokenHash,
			nil,
			nil,
		)
		if consumeErr != nil {
			return nil, consumeErr
		}
		if expired {
			return nil, apperrors.WrapInvalidInput("token has expired")
		}
		return nil, apperrors.WrapInvalidInput("invalid or expired token")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.BcryptCost)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error_type", fmt.Sprintf("%T", err))
		return nil, apperrors.Wrap(err, "failed to hash password")
	}

	expired, err := s.consumeResetToken(
		ctx,
		candidate.AccountID,
		tokenHash,
		hashedPassword,
		&audit,
	)
	if err != nil {
		return nil, err
	}
	if expired {
		return nil, apperrors.WrapInvalidInput("token has expired")
	}

	slog.InfoContext(ctx, "password reset completed",
		slog.Uint64("account_id", candidate.AccountID))
	return &PasswordResetCompletion{AccountID: candidate.AccountID}, nil
}

func (s *passwordResetService) consumeResetToken(
	ctx context.Context,
	accountID uint64,
	tokenHash string,
	hashedPassword []byte,
	audit *CredentialMutationAudit,
) (bool, error) {
	expired := false
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		lockedAccount, err := s.accountRepo.FindByIDForUpdate(txCtx, accountID)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("invalid or expired token")
			}
			return apperrors.Wrap(err, "failed to lock password reset account")
		}
		if lockedAccount == nil || lockedAccount.ID != accountID {
			return apperrors.WrapInternalServerError(
				"password reset account lock returned an invalid result",
			)
		}

		resetToken, err := s.tokenRepo.FindByTokenHashForUpdate(txCtx, tokenHash)
		if err != nil {
			if apperrors.IsNotFound(err) {
				return apperrors.WrapInvalidInput("invalid or expired token")
			}
			return apperrors.Wrap(err, "failed to lock password reset token")
		}
		if resetToken == nil {
			return apperrors.WrapInvalidInput("invalid or expired token")
		}
		if resetToken.AccountID != accountID {
			return apperrors.WrapInternalServerError(
				"password reset token account changed during consumption",
			)
		}

		if resetToken.CreatedAt.IsZero() ||
			lockedAccount.UpdatedAt.IsZero() ||
			!resetToken.CreatedAt.After(lockedAccount.UpdatedAt) {
			if err := s.tokenRepo.ConsumeByID(txCtx, resetToken.ID); err != nil {
				return apperrors.Wrap(err, "failed to consume stale reset token")
			}
			expired = true
			return nil
		}

		if resetTokenExpired(resetToken.ExpiresAt, time.Now()) {
			if err := s.tokenRepo.ConsumeByID(txCtx, resetToken.ID); err != nil {
				return apperrors.Wrap(err, "failed to consume expired reset token")
			}
			expired = true
			return nil
		}

		if len(hashedPassword) == 0 {
			return apperrors.WrapInvalidInput("invalid or expired token")
		}
		updatedAt := nextAccountSessionEpoch(lockedAccount.UpdatedAt, time.Now())
		if err := s.credentialUpdater.UpdatePasswordHash(
			txCtx,
			resetToken.AccountID,
			string(hashedPassword),
			updatedAt,
		); err != nil {
			slog.ErrorContext(
				ctx,
				"failed to update password",
				"error_type", fmt.Sprintf("%T", err),
				"account_id", resetToken.AccountID,
			)
			return apperrors.Wrap(err, "failed to update password")
		}
		if err := s.tokenRepo.ConsumeByID(txCtx, resetToken.ID); err != nil {
			return apperrors.Wrap(err, "failed to consume used reset token")
		}
		if audit == nil {
			return apperrors.WrapInternalServerError(
				"password reset credential audit metadata is unavailable",
			)
		}
		subject, resolveErr := s.auditSubject.ResolveCredentialAuditSubject(
			txCtx,
			resetToken.AccountID,
		)
		if resolveErr != nil {
			return apperrors.Wrap(
				resolveErr,
				"failed to resolve password reset audit subject",
			)
		}
		if subject.ClinicID == 0 || subject.StaffID == 0 {
			return apperrors.WrapInternalServerError(
				"password reset audit subject is invalid",
			)
		}
		entryAudit := *audit
		entryAudit.ClinicID = subject.ClinicID
		entryAudit.TargetStaffID = subject.StaffID
		if auditErr := s.audit.LogEntryTx(
			txCtx,
			credentialAuditEntry(entryAudit, resetToken.AccountID),
		); auditErr != nil {
			return apperrors.Wrap(
				auditErr,
				"failed to write password reset audit",
			)
		}
		return nil
	}); err != nil {
		return false, apperrors.Wrap(err, "password reset transaction failed")
	}
	return expired, nil
}

func resetTokenExpired(expiresAt, now time.Time) bool {
	return expiresAt.IsZero() || !expiresAt.After(now)
}

func (s *passwordResetService) Wait() {
	s.mailGate.CloseAndWait()
}

func generateToken() (rawToken, tokenHash string, err error) {
	tokenBytes := make([]byte, 32)
	if _, err = rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("crypto/rand: %w", err)
	}
	rawToken = hex.EncodeToString(tokenBytes)
	tokenHash = hashToken(rawToken)
	return rawToken, tokenHash, nil
}

func hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}

func (s *passwordResetService) sendResetEmail(
	ctx context.Context,
	to, resetURL string,
) error {
	if s.cfg.SMTPHost == "" {
		return nil
	}
	if s.sendMail == nil {
		return errors.New("password reset email sender is not configured")
	}

	subject := "パスワードリセットのご案内"
	body := fmt.Sprintf(
		"パスワードリセットのリクエストを受け付けました。\r\n\r\n"+
			"以下のリンクからパスワードを変更してください（有効期限：30分）：\r\n\r\n"+
			"%s\r\n\r\n"+
			"このメールに心当たりがない場合は無視してください。",
		resetURL,
	)

	from := s.cfg.SMTPFrom
	message := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	return s.sendMail(ctx, s.cfg, from, to, message)
}
