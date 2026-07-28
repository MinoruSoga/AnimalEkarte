package auth

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockAccountRepository は account_service_test.go で定義済み

// ---- PasswordResetToken モック ----

type mockPasswordResetTokenRepository struct {
	findLatestByAccountIDForUpdateFn func(
		ctx context.Context,
		accountID uint64,
	) (*model.PasswordResetToken, error)
	createFn                   func(ctx context.Context, token *model.PasswordResetToken) error
	findByTokenHashFn          func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	findByTokenHashForUpdateFn func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	deleteByAccountIDFn        func(ctx context.Context, accountID uint64) error
	deleteByIDFn               func(ctx context.Context, id uint64) error
	deleteIssuedFn             func(ctx context.Context, id uint64, tokenHash string) error
	consumeByIDFn              func(ctx context.Context, id uint64) error
}

func (m *mockPasswordResetTokenRepository) FindLatestByAccountIDForUpdate(
	ctx context.Context,
	accountID uint64,
) (*model.PasswordResetToken, error) {
	if m.findLatestByAccountIDForUpdateFn != nil {
		return m.findLatestByAccountIDForUpdateFn(ctx, accountID)
	}
	return nil, apperrors.WrapNotFound("password_reset_token", "latest")
}

func (m *mockPasswordResetTokenRepository) Create(ctx context.Context, token *model.PasswordResetToken) error {
	if m.createFn != nil {
		return m.createFn(ctx, token)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*model.PasswordResetToken, error) {
	return m.findByTokenHashFn(ctx, hash)
}

func (m *mockPasswordResetTokenRepository) FindByTokenHashForUpdate(
	ctx context.Context,
	hash string,
) (*model.PasswordResetToken, error) {
	var token *model.PasswordResetToken
	var err error
	if m.findByTokenHashForUpdateFn != nil {
		token, err = m.findByTokenHashForUpdateFn(ctx, hash)
	} else {
		token, err = m.FindByTokenHash(ctx, hash)
	}
	if err == nil && token != nil && token.CreatedAt.IsZero() {
		copyToken := *token
		copyToken.CreatedAt = time.Now()
		token = &copyToken
	}
	return token, err
}

func (m *mockPasswordResetTokenRepository) DeleteByAccountID(ctx context.Context, accountID uint64) error {
	if m.deleteByAccountIDFn != nil {
		return m.deleteByAccountIDFn(ctx, accountID)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) DeleteByID(ctx context.Context, id uint64) error {
	if m.deleteByIDFn != nil {
		return m.deleteByIDFn(ctx, id)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) DeleteIssued(
	ctx context.Context,
	id uint64,
	tokenHash string,
) error {
	if m.deleteIssuedFn != nil {
		return m.deleteIssuedFn(ctx, id, tokenHash)
	}
	return nil
}

func (m *mockPasswordResetTokenRepository) ConsumeByID(ctx context.Context, id uint64) error {
	if m.consumeByIDFn != nil {
		return m.consumeByIDFn(ctx, id)
	}
	return m.DeleteByID(ctx, id)
}

type passwordResetTransactorFunc func(context.Context, func(context.Context) error) error

func (f passwordResetTransactorFunc) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return f(ctx, fn)
}

type passwordResetTxMarker struct{}

func immediatePasswordResetTransactor() Transactor {
	return passwordResetTransactorFunc(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(context.WithValue(ctx, passwordResetTxMarker{}, true))
	})
}

func immediateForgotPasswordResponseTiming() forgotPasswordResponseTiming {
	return forgotPasswordResponseTiming{
		now: time.Now,
		sleep: func(context.Context, time.Duration) error {
			return nil
		},
		jitter: func() time.Duration {
			return 0
		},
	}
}

func newTestPasswordResetService(accountRepo *mockAccountRepository, tokenRepo *mockPasswordResetTokenRepository) PasswordResetService {
	service := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{FrontendURL: "https://example.com"},
		accountRepo,
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			return nil
		},
		immediatePasswordResetTransactor(),
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)
	service.(*passwordResetService).responseTiming =
		immediateForgotPasswordResponseTiming()
	return service
}

// ---- ForgotPassword ----

func TestPasswordResetService_ForgotPassword(t *testing.T) {
	tests := []struct {
		name                string
		findByEmailFn       func(ctx context.Context, email string) (*model.Account, error)
		deleteByAccountIDFn func(ctx context.Context, accountID uint64) error
		createFn            func(ctx context.Context, token *model.PasswordResetToken) error
		wantErr             bool
	}{
		{
			name: "silently succeeds when account is not found (no info leak)",
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return nil, apperrors.WrapNotFound("account", email)
			},
			wantErr: false,
		},
		{
			name: "propagates unexpected FindByEmail error",
			findByEmailFn: func(_ context.Context, _ string) (*model.Account, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "propagates error cleaning up existing tokens",
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return &model.Account{ID: 1, Email: email}, nil
			},
			deleteByAccountIDFn: func(_ context.Context, _ uint64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "propagates error creating reset token",
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return &model.Account{ID: 1, Email: email}, nil
			},
			createFn: func(_ context.Context, _ *model.PasswordResetToken) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "issues reset token successfully",
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return &model.Account{ID: 1, Email: email}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mockAccountRepository{findByEmailFn: tt.findByEmailFn}
			tokenRepo := &mockPasswordResetTokenRepository{
				deleteByAccountIDFn: tt.deleteByAccountIDFn,
				createFn:            tt.createFn,
			}
			svc := newTestPasswordResetService(accountRepo, tokenRepo)

			err := svc.ForgotPassword(context.Background(), "owner@example.com")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordResetService_ForgotPassword_LocksAccountAndReplacesTokenInOneTransaction(
	t *testing.T,
) {
	events := make([]string, 0, 3)
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 7, Email: email}, nil
		},
		findByIDForUpdateFn: func(ctx context.Context, id uint64) (*model.Account, error) {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			assert.Equal(t, uint64(7), id)
			events = append(events, "lock-account")
			return &model.Account{ID: id}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		deleteByAccountIDFn: func(ctx context.Context, accountID uint64) error {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			assert.Equal(t, uint64(7), accountID)
			events = append(events, "delete-existing")
			return nil
		},
		createFn: func(ctx context.Context, token *model.PasswordResetToken) error {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			assert.Equal(t, uint64(7), token.AccountID)
			events = append(events, "create")
			return nil
		},
	}
	svc := newTestPasswordResetService(accountRepo, tokenRepo)

	require.NoError(t, svc.ForgotPassword(context.Background(), "owner@example.com"))
	assert.Equal(t, []string{"lock-account", "delete-existing", "create"}, events)
}

func TestPasswordResetService_ForgotPassword_UsesThirtyMinuteExpiry(t *testing.T) {
	var expiresAt time.Time
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 7, Email: email}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		createFn: func(_ context.Context, token *model.PasswordResetToken) error {
			expiresAt = token.ExpiresAt
			return nil
		},
	}
	beforeIssue := time.Now()
	svc := newTestPasswordResetService(accountRepo, tokenRepo)

	require.NoError(t, svc.ForgotPassword(context.Background(), "owner@example.com"))
	afterIssue := time.Now()

	require.False(t, expiresAt.IsZero())
	assert.False(t, expiresAt.Before(beforeIssue.Add(passwordResetTokenExpiry)))
	assert.False(t, expiresAt.After(afterIssue.Add(passwordResetTokenExpiry)))
}

func TestPasswordResetService_ForgotPassword_SuppressesActiveRecentReissue(t *testing.T) {
	now := time.Unix(1_727_123_456, 0)
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 7, Email: email}, nil
		},
	}
	mutationCalled := false
	mailCalled := false
	tokenRepo := &mockPasswordResetTokenRepository{
		findLatestByAccountIDForUpdateFn: func(
			ctx context.Context,
			accountID uint64,
		) (*model.PasswordResetToken, error) {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			assert.Equal(t, uint64(7), accountID)
			return &model.PasswordResetToken{
				AccountID: accountID,
				CreatedAt: now.Add(-passwordResetReissueCooldown / 2),
				ExpiresAt: now.Add(passwordResetTokenExpiry),
			}, nil
		},
		deleteByAccountIDFn: func(context.Context, uint64) error {
			mutationCalled = true
			return nil
		},
		createFn: func(context.Context, *model.PasswordResetToken) error {
			mutationCalled = true
			return nil
		},
	}
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
		},
		accountRepo,
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			mailCalled = true
			return nil
		},
		immediatePasswordResetTransactor(),
	).(*passwordResetService)
	service.now = func() time.Time { return now }
	service.responseTiming = immediateForgotPasswordResponseTiming()

	require.NoError(t, service.ForgotPassword(context.Background(), "owner@example.com"))
	service.Wait()

	assert.False(t, mutationCalled)
	assert.False(t, mailCalled)
}

func TestPasswordResetService_ForgotPassword_BoundsAcceptedMailWorkers(t *testing.T) {
	var accountSequence atomic.Uint64
	var created atomic.Int32
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{
				ID:    accountSequence.Add(1),
				Email: email,
			}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		createFn: func(context.Context, *model.PasswordResetToken) error {
			created.Add(1)
			return nil
		},
	}
	mailStarted := make(chan struct{}, passwordResetMailWorkers)
	releaseMail := make(chan struct{})
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
		},
		accountRepo,
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			mailStarted <- struct{}{}
			<-releaseMail
			return nil
		},
		immediatePasswordResetTransactor(),
	).(*passwordResetService)
	service.responseTiming = immediateForgotPasswordResponseTiming()

	for index := range passwordResetMailWorkers {
		require.NoError(
			t,
			service.ForgotPassword(
				context.Background(),
				"owner-"+string(rune('a'+index))+"@example.com",
			),
		)
	}
	for range passwordResetMailWorkers {
		<-mailStarted
	}

	require.NoError(
		t,
		service.ForgotPassword(context.Background(), "overflow-owner@example.com"),
	)
	assert.Equal(t, int32(passwordResetMailWorkers), created.Load())

	close(releaseMail)
	service.Wait()
}

func TestPasswordResetService_ForgotPassword_SendFailureRemovesIssuedTokenAndAllowsRetry(
	t *testing.T,
) {
	var mu sync.Mutex
	var liveToken *model.PasswordResetToken
	var nextTokenID uint64
	var created atomic.Int32
	var cleanupCalls atomic.Int32
	var sendCalls atomic.Int32
	cleanupDone := make(chan struct{}, 1)
	secondSendDone := make(chan struct{}, 1)

	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 81, Email: email}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		findLatestByAccountIDForUpdateFn: func(
			context.Context,
			uint64,
		) (*model.PasswordResetToken, error) {
			mu.Lock()
			defer mu.Unlock()
			if liveToken == nil {
				return nil, apperrors.WrapNotFound(
					"password_reset_token",
					"latest",
				)
			}
			copyToken := *liveToken
			return &copyToken, nil
		},
		createFn: func(_ context.Context, token *model.PasswordResetToken) error {
			mu.Lock()
			defer mu.Unlock()
			nextTokenID++
			token.ID = nextTokenID
			copyToken := *token
			liveToken = &copyToken
			created.Add(1)
			return nil
		},
		deleteIssuedFn: func(
			_ context.Context,
			id uint64,
			tokenHash string,
		) error {
			mu.Lock()
			if liveToken != nil &&
				liveToken.ID == id &&
				liveToken.TokenHash == tokenHash {
				liveToken = nil
			}
			mu.Unlock()
			cleanupCalls.Add(1)
			cleanupDone <- struct{}{}
			return nil
		},
	}
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
		},
		accountRepo,
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			if sendCalls.Add(1) == 1 {
				return errors.New("smtp unavailable")
			}
			secondSendDone <- struct{}{}
			return nil
		},
		immediatePasswordResetTransactor(),
	).(*passwordResetService)
	service.responseTiming = immediateForgotPasswordResponseTiming()

	require.NoError(
		t,
		service.ForgotPassword(context.Background(), "owner@example.com"),
	)
	select {
	case <-cleanupDone:
	case <-time.After(time.Second):
		t.Fatal("failed reset email token was not cleaned up")
	}

	require.NoError(
		t,
		service.ForgotPassword(context.Background(), "owner@example.com"),
	)
	select {
	case <-secondSendDone:
	case <-time.After(time.Second):
		t.Fatal("retry email was suppressed by the failed send cooldown")
	}
	service.Wait()

	assert.Equal(t, int32(2), created.Load())
	assert.Equal(t, int32(1), cleanupCalls.Load())
	assert.Equal(t, int32(2), sendCalls.Load())
	mu.Lock()
	require.NotNil(t, liveToken)
	assert.Equal(t, uint64(2), liveToken.ID)
	mu.Unlock()
}

func TestPasswordResetService_ForgotPassword_SendFailureCleanupDoesNotDeleteNewerToken(
	t *testing.T,
) {
	var mu sync.Mutex
	var issuedToken *model.PasswordResetToken
	var liveToken *model.PasswordResetToken
	cleanupDone := make(chan struct{})

	tokenRepo := &mockPasswordResetTokenRepository{
		createFn: func(_ context.Context, token *model.PasswordResetToken) error {
			mu.Lock()
			defer mu.Unlock()
			token.ID = 1
			copyToken := *token
			issuedToken = &copyToken
			liveToken = &copyToken
			return nil
		},
		deleteIssuedFn: func(
			_ context.Context,
			id uint64,
			tokenHash string,
		) error {
			mu.Lock()
			defer mu.Unlock()
			if liveToken != nil &&
				liveToken.ID == id &&
				liveToken.TokenHash == tokenHash {
				liveToken = nil
			}
			close(cleanupDone)
			return nil
		},
	}
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
		},
		&mockAccountRepository{
			findByEmailFn: func(
				_ context.Context,
				email string,
			) (*model.Account, error) {
				return &model.Account{ID: 81, Email: email}, nil
			},
		},
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			mu.Lock()
			liveToken = &model.PasswordResetToken{
				ID:        2,
				AccountID: 81,
				TokenHash: "newer-token-hash",
				ExpiresAt: time.Now().Add(passwordResetTokenExpiry),
				CreatedAt: time.Now(),
			}
			mu.Unlock()
			return errors.New("smtp unavailable")
		},
		immediatePasswordResetTransactor(),
	).(*passwordResetService)
	service.responseTiming = immediateForgotPasswordResponseTiming()

	require.NoError(
		t,
		service.ForgotPassword(context.Background(), "owner@example.com"),
	)
	select {
	case <-cleanupDone:
	case <-time.After(time.Second):
		t.Fatal("failed reset email cleanup did not complete")
	}
	service.Wait()

	mu.Lock()
	require.NotNil(t, issuedToken)
	require.NotNil(t, liveToken)
	assert.Equal(t, uint64(2), liveToken.ID)
	assert.Equal(t, "newer-token-hash", liveToken.TokenHash)
	mu.Unlock()
}

func TestPasswordResetService_ForgotPassword_FailsClosedWithoutTransactor(t *testing.T) {
	tokenMutationCalled := false
	svc := NewPasswordResetService(
		&PasswordResetConfig{FrontendURL: "https://example.com"},
		&mockAccountRepository{
			findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
				return &model.Account{ID: 7, Email: email}, nil
			},
		},
		&mockPasswordResetTokenRepository{
			createFn: func(_ context.Context, _ *model.PasswordResetToken) error {
				tokenMutationCalled = true
				return nil
			},
		},
		nil,
	)
	svc.(*passwordResetService).responseTiming =
		immediateForgotPasswordResponseTiming()

	err := svc.ForgotPassword(context.Background(), "owner@example.com")

	require.Error(t, err)
	assert.False(t, tokenMutationCalled)
}

func TestPasswordResetService_ForgotPassword_PerformsOneFixedCost12ComparisonOnEveryOutcome(
	t *testing.T,
) {
	type serviceCase struct {
		name         string
		accountRepo  AccountRepository
		transactor   Transactor
		expectError  bool
		privateError string
	}
	const privateEmail = "private-owner@example.test"
	const privateToken = "private-reset-token"
	cases := []serviceCase{
		{
			name: "existing account",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return &model.Account{ID: 7}, nil
				},
			},
			transactor: immediatePasswordResetTransactor(),
		},
		{
			name: "nonexistent account",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, apperrors.WrapNotFound("account", "lookup")
				},
			},
			transactor: immediatePasswordResetTransactor(),
		},
		{
			name: "nil account result",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, nil
				},
			},
			transactor:  immediatePasswordResetTransactor(),
			expectError: true,
		},
		{
			name: "repository error",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return nil, errors.New(
						"lookup failed for " + privateEmail + " token=" + privateToken,
					)
				},
			},
			transactor:   immediatePasswordResetTransactor(),
			expectError:  true,
			privateError: privateEmail,
		},
		{
			name: "transaction error",
			accountRepo: &mockAccountRepository{
				findByEmailFn: func(context.Context, string) (*model.Account, error) {
					return &model.Account{ID: 7}, nil
				},
			},
			transactor: passwordResetTransactorFunc(
				func(context.Context, func(context.Context) error) error {
					return errors.New("transaction failed token=" + privateToken)
				},
			),
			expectError:  true,
			privateError: privateToken,
		},
		{
			name:        "missing transaction boundary",
			transactor:  nil,
			expectError: true,
		},
	}

	cost, err := bcrypt.Cost([]byte(dummyPasswordHash))
	require.NoError(t, err)
	require.Equal(t, config.BcryptCost, cost)

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var logs bytes.Buffer
			originalLogger := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			t.Cleanup(func() {
				slog.SetDefault(originalLogger)
			})

			service := &passwordResetService{
				cfg:            &PasswordResetConfig{FrontendURL: "https://example.com"},
				accountRepo:    test.accountRepo,
				tokenRepo:      &mockPasswordResetTokenRepository{},
				transactor:     test.transactor,
				responseTiming: immediateForgotPasswordResponseTiming(),
			}
			comparisonCount := 0
			service.comparePassword = func(hashedPassword, password []byte) error {
				comparisonCount++
				assert.Equal(t, dummyPasswordHash, string(hashedPassword))
				assert.Equal(t, dummyPasswordCandidate, string(password))
				return bcrypt.ErrMismatchedHashAndPassword
			}

			forgotPasswordError := service.ForgotPassword(
				context.Background(),
				privateEmail,
			)
			service.Wait()

			if test.expectError {
				require.Error(t, forgotPasswordError)
			} else {
				require.NoError(t, forgotPasswordError)
			}
			assert.Equal(t, 1, comparisonCount)
			assert.NotContains(t, logs.String(), privateEmail)
			assert.NotContains(t, logs.String(), privateToken)
			if test.privateError != "" {
				assert.NotContains(t, logs.String(), test.privateError)
			}
		})
	}
}

// ---- Wait (PERF-FOLLOWUP-05: shutdown drain) ----

// TestPasswordResetService_Wait_DrainsInFlightEmailGoroutine は、ForgotPassword が起動する
// fire-and-forget メール送信 goroutine を Wait() が実際に待機することを証明する。
// 修正前（wg なし）は Wait() 自体が存在せずコンパイル不可 — 本テストは修正と一体で追加する
// 加算的信頼性修正のため、RED は「Wait 未実装」という形で表現される。
func TestPasswordResetService_Wait_DrainsInFlightEmailGoroutine(t *testing.T) {
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 1, Email: email}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{}
	mailStarted := make(chan struct{})
	releaseMail := make(chan struct{})

	svc := &passwordResetService{
		cfg: &PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
			SMTPFrom:    "noreply@example.com",
		},
		accountRepo:    accountRepo,
		tokenRepo:      tokenRepo,
		transactor:     immediatePasswordResetTransactor(),
		responseTiming: immediateForgotPasswordResponseTiming(),
		sendMail: func(
			context.Context,
			*PasswordResetConfig,
			string,
			string,
			[]byte,
		) error {
			close(mailStarted)
			<-releaseMail
			return nil
		},
	}

	require.NoError(t, svc.ForgotPassword(context.Background(), "owner@example.com"))
	<-mailStarted

	waitReturned := make(chan struct{})
	go func() {
		svc.Wait()
		close(waitReturned)
	}()

	select {
	case <-waitReturned:
		t.Fatal("Wait() returned before the in-flight email sender was released")
	default:
	}

	close(releaseMail)
	select {
	case <-waitReturned:
	case <-time.After(time.Second):
		t.Fatal("Wait() did not return after the in-flight email sender completed")
	}
}

func TestPasswordResetService_WaitClosesMailWorkerRegistrationBeforeTokenIssue(
	t *testing.T,
) {
	var created atomic.Int32
	var mailCalls atomic.Int32
	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 71, Email: email}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		createFn: func(context.Context, *model.PasswordResetToken) error {
			created.Add(1)
			return nil
		},
	}
	service := NewPasswordResetServiceWithTransactor(
		&PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    "smtp.example.com",
		},
		accountRepo,
		tokenRepo,
		func(context.Context, *PasswordResetConfig, string, string, []byte) error {
			mailCalls.Add(1)
			return nil
		},
		immediatePasswordResetTransactor(),
	).(*passwordResetService)
	service.responseTiming = immediateForgotPasswordResponseTiming()

	service.Wait()
	require.NoError(
		t,
		service.ForgotPassword(context.Background(), "owner@example.com"),
	)

	assert.Zero(t, created.Load())
	assert.Zero(t, mailCalls.Load())
	assert.Empty(t, service.mailWorkerSlots)
}

func TestResetTokenExpired_ExactExpiryBoundaryIsExpired(t *testing.T) {
	now := time.Unix(1_727_123_456, 789_012_345)

	assert.True(t, resetTokenExpired(time.Time{}, now))
	assert.True(t, resetTokenExpired(now, now))
	assert.True(t, resetTokenExpired(now.Add(-time.Nanosecond), now))
	assert.False(t, resetTokenExpired(now.Add(time.Nanosecond), now))
}

// ---- ResetPassword ----

func TestPasswordResetService_ResetPassword(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	tooLongPassword := strings.Repeat("a", 73) // bcrypt はUTF-8バイト長72を超えるパスワードを拒否する
	lookupErr := errors.New("password reset token lookup unavailable")
	lockErr := errors.New("password reset token lock unavailable")

	tests := []struct {
		name                       string
		findByTokenHashFn          func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
		findByTokenHashForUpdateFn func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
		deleteByIDFn               func(ctx context.Context, id uint64) error
		consumeByIDFn              func(ctx context.Context, id uint64) error
		updatePasswordHashFn       func(
			ctx context.Context,
			id uint64,
			newHash string,
			updatedAt time.Time,
		) error
		newPassword      string
		wantErr          bool
		wantInvalidInput bool
		wantCause        error
	}{
		{
			name: "returns invalid input when token is not found",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return nil, apperrors.WrapNotFound("password_reset_token", "hash")
			},
			newPassword:      "newpass123",
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name: "preserves unexpected initial token lookup error",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return nil, lookupErr
			},
			newPassword: "newpass123",
			wantErr:     true,
			wantCause:   lookupErr,
		},
		{
			name: "returns invalid input when locked token is no longer present",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			findByTokenHashForUpdateFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return nil, apperrors.WrapNotFound("password_reset_token", "hash")
			},
			newPassword:      "newpass123",
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name: "preserves unexpected locked token lookup error",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			findByTokenHashForUpdateFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return nil, lockErr
			},
			newPassword: "newpass123",
			wantErr:     true,
			wantCause:   lockErr,
		},
		{
			name: "returns invalid input and cleans up expired token",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: past}, nil
			},
			newPassword:      "newpass123",
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name: "returns error when expired token consumption fails",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: past}, nil
			},
			consumeByIDFn: func(_ context.Context, _ uint64) error {
				return errors.New("db error")
			},
			newPassword: "newpass123",
			wantErr:     true,
		},
		{
			name: "returns invalid input and consumes token with missing expiry",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1}, nil
			},
			newPassword:      "newpass123",
			wantErr:          true,
			wantInvalidInput: true,
		},
		{
			name: "returns error when password exceeds bcrypt length limit",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			newPassword: tooLongPassword,
			wantErr:     true,
		},
		{
			name: "propagates account update error",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			updatePasswordHashFn: func(
				_ context.Context,
				_ uint64,
				_ string,
				_ time.Time,
			) error {
				return errors.New("db error")
			},
			newPassword: "newpass123",
			wantErr:     true,
		},
		{
			name: "resets password successfully",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			newPassword: "newpass123",
			wantErr:     false,
		},
		{
			name: "returns error when used token consumption fails",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			consumeByIDFn: func(_ context.Context, _ uint64) error {
				return errors.New("db error")
			},
			newPassword: "newpass123",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mockAccountRepository{
				updatePasswordHashFn: tt.updatePasswordHashFn,
			}
			tokenRepo := &mockPasswordResetTokenRepository{
				findByTokenHashFn:          tt.findByTokenHashFn,
				findByTokenHashForUpdateFn: tt.findByTokenHashForUpdateFn,
				deleteByIDFn:               tt.deleteByIDFn,
				consumeByIDFn:              tt.consumeByIDFn,
			}
			svc := newTestPasswordResetService(accountRepo, tokenRepo)

			err := svc.ResetPassword(context.Background(), "raw-token", tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				if tt.wantCause != nil {
					assert.ErrorIs(t, err, tt.wantCause)
					assert.False(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordResetService_ResetPasswordWithResultReturnsAccountID(t *testing.T) {
	const accountID = uint64(73)
	future := time.Now().Add(time.Hour)
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(
			context.Context,
			string,
		) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{
				ID:        11,
				AccountID: accountID,
				ExpiresAt: future,
			}, nil
		},
	}
	service := newTestPasswordResetService(
		&mockAccountRepository{},
		tokenRepo,
	)
	completionService, ok := service.(PasswordResetCompletionService)
	require.True(t, ok)

	result, err := completionService.ResetPasswordWithResult(
		context.Background(),
		"raw-token",
		"newpass123",
		testPasswordResetAudit(),
	)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, accountID, result.AccountID)
}

func TestPasswordResetService_ResetPasswordWithResultReturnsNoSubjectOnFailure(
	t *testing.T,
) {
	service := newTestPasswordResetService(
		&mockAccountRepository{},
		&mockPasswordResetTokenRepository{
			findByTokenHashFn: func(
				context.Context,
				string,
			) (*model.PasswordResetToken, error) {
				return nil, apperrors.WrapNotFound(
					"password_reset_token",
					"hash",
				)
			},
		},
	)
	completionService, ok := service.(PasswordResetCompletionService)
	require.True(t, ok)

	result, err := completionService.ResetPasswordWithResult(
		context.Background(),
		"raw-token",
		"newpass123",
		testPasswordResetAudit(),
	)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestPasswordResetService_ResetPassword_FailsClosedWithoutTransactor(t *testing.T) {
	repositoryCalled := false
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
			repositoryCalled = true
			return nil, errors.New("must not be reached")
		},
	}
	svc := NewPasswordResetService(
		&PasswordResetConfig{FrontendURL: "https://example.com"},
		&mockAccountRepository{},
		tokenRepo,
		nil,
	)

	err := svc.ResetPassword(context.Background(), "raw-token", "newpass123")

	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "INTERNAL", appErr.Code)
	assert.False(t, repositoryCalled)
}

func TestPasswordResetService_ResetPassword_UsesOneTransactionInAccountTokenUpdateConsumeOrder(t *testing.T) {
	future := time.Now().Add(time.Hour)
	events := make([]string, 0, 4)
	assertInTx := func(t *testing.T, ctx context.Context) {
		t.Helper()
		assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
	}

	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{ID: 7, AccountID: 11, ExpiresAt: future}, nil
		},
		findByTokenHashForUpdateFn: func(ctx context.Context, _ string) (*model.PasswordResetToken, error) {
			assertInTx(t, ctx)
			events = append(events, "token_lock")
			return &model.PasswordResetToken{ID: 7, AccountID: 11, ExpiresAt: future}, nil
		},
		consumeByIDFn: func(ctx context.Context, id uint64) error {
			assertInTx(t, ctx)
			assert.Equal(t, uint64(7), id)
			events = append(events, "consume")
			return nil
		},
	}
	accountRepo := &mockAccountRepository{
		findByIDForUpdateFn: func(ctx context.Context, id uint64) (*model.Account, error) {
			assertInTx(t, ctx)
			assert.Equal(t, uint64(11), id)
			events = append(events, "account_lock")
			return &model.Account{ID: id, UpdatedAt: time.Now().Add(-time.Hour)}, nil
		},
		updatePasswordHashFn: func(
			ctx context.Context,
			id uint64,
			newHash string,
			updatedAt time.Time,
		) error {
			assertInTx(t, ctx)
			assert.Equal(t, uint64(11), id)
			assert.NotEmpty(t, newHash)
			assert.False(t, updatedAt.IsZero())
			events = append(events, "update")
			return nil
		},
	}
	svc := newTestPasswordResetService(accountRepo, tokenRepo)

	require.NoError(t, svc.ResetPassword(context.Background(), "raw-token", "newpass123"))
	assert.Equal(t, []string{"account_lock", "token_lock", "update", "consume"}, events)
}

func TestPasswordResetService_ResetPassword_RejectsTokenIssuedBeforeCredentialEpoch(
	t *testing.T,
) {
	now := time.Now()
	token := &model.PasswordResetToken{
		ID:        7,
		AccountID: 11,
		CreatedAt: now.Add(-10 * time.Minute),
		ExpiresAt: now.Add(time.Hour),
	}
	credentialUpdatedAt := now.Add(-5 * time.Minute)
	passwordUpdated := false
	tokenConsumed := false

	accountRepo := &mockAccountRepository{
		findByIDForUpdateFn: func(context.Context, uint64) (*model.Account, error) {
			return &model.Account{
				ID:        token.AccountID,
				UpdatedAt: credentialUpdatedAt,
			}, nil
		},
		updatePasswordHashFn: func(
			context.Context,
			uint64,
			string,
			time.Time,
		) error {
			passwordUpdated = true
			return nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(context.Context, string) (*model.PasswordResetToken, error) {
			return token, nil
		},
		consumeByIDFn: func(_ context.Context, id uint64) error {
			assert.Equal(t, token.ID, id)
			tokenConsumed = true
			return nil
		},
	}
	svc := newTestPasswordResetService(accountRepo, tokenRepo)

	err := svc.ResetPassword(context.Background(), "raw-token", "newpass123")

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.False(t, passwordUpdated)
	assert.True(t, tokenConsumed)
}

func TestPasswordResetService_ResetPassword_ConsumeFailureRollsBackPasswordUpdate(t *testing.T) {
	future := time.Now().Add(time.Hour)
	consumeErr := errors.New("consume failed")
	committedPassword := "old-password-hash"
	stagedPassword := ""
	rolledBack := false
	transactor := passwordResetTransactorFunc(func(ctx context.Context, fn func(context.Context) error) error {
		stagedPassword = committedPassword
		txCtx := context.WithValue(ctx, passwordResetTxMarker{}, true)
		if err := fn(txCtx); err != nil {
			rolledBack = true
			stagedPassword = ""
			return err
		}
		committedPassword = stagedPassword
		return nil
	})
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{ID: 3, AccountID: 5, ExpiresAt: future}, nil
		},
		consumeByIDFn: func(ctx context.Context, _ uint64) error {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			return consumeErr
		},
	}
	accountRepo := &mockAccountRepository{
		updatePasswordHashFn: func(ctx context.Context, _ uint64, newHash string, _ time.Time) error {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			stagedPassword = newHash
			return nil
		},
	}
	svc := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{FrontendURL: "https://example.com"},
		accountRepo,
		tokenRepo,
		nil,
		transactor,
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)

	err := svc.ResetPassword(context.Background(), "raw-token", "newpass123")

	require.Error(t, err)
	require.ErrorIs(t, err, consumeErr)
	assert.True(t, rolledBack)
	assert.Equal(t, "old-password-hash", committedPassword)
	assert.Empty(t, stagedPassword)
}

func TestPasswordResetService_ResetPassword_ExpiredTokenConsumptionCommitsBeforeInvalidResponse(t *testing.T) {
	expired := time.Now().Add(-time.Minute)
	tokenConsumed := false
	transactorCommitted := false
	transactor := passwordResetTransactorFunc(func(ctx context.Context, fn func(context.Context) error) error {
		txCtx := context.WithValue(ctx, passwordResetTxMarker{}, true)
		if err := fn(txCtx); err != nil {
			return err
		}
		transactorCommitted = true
		return nil
	})
	tokenRepo := &mockPasswordResetTokenRepository{
		findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
			return &model.PasswordResetToken{ID: 13, AccountID: 17, ExpiresAt: expired}, nil
		},
		consumeByIDFn: func(ctx context.Context, id uint64) error {
			assert.Equal(t, true, ctx.Value(passwordResetTxMarker{}))
			assert.Equal(t, uint64(13), id)
			tokenConsumed = true
			return nil
		},
	}
	svc := NewPasswordResetServiceWithCredentialAudit(
		&PasswordResetConfig{FrontendURL: "https://example.com"},
		&mockAccountRepository{},
		tokenRepo,
		nil,
		transactor,
		noopCredentialAuditSubjectResolver{},
		noopCredentialAuditTxLogger{},
	)

	err := svc.ResetPassword(context.Background(), "raw-token", "newpass123")

	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.True(t, tokenConsumed)
	assert.True(t, transactorCommitted)
}

// ---- generateToken / hashToken ----

// NOTE: crypto/rand.Read() の異常系（Reader 差し替えによるエラー注入）は Go 1.24+ では
// テスト不可能。crypto/rand.Read はエラーを返さない契約になっており、内部 Reader が失敗すると
// runtime.fatal でプロセスごと即死する（recover 不可）ため、generateToken() 内の
// `if _, err = rand.Read(b); err != nil` 分岐は現行 Go では到達不能な防御的コードであり、
// この分岐をテストで再現しようとするとテストバイナリ自体がクラッシュする。
// 該当ロジックは変更せず（本タスクは振る舞い変更禁止）、到達不能分岐のテストのみ省略する。
func TestGenerateToken(t *testing.T) {
	t.Run("returns raw token and matching sha256 hash", func(t *testing.T) {
		raw, hash, err := generateToken()

		require.NoError(t, err)
		assert.NotEmpty(t, raw)
		assert.Len(t, raw, 64) // 32 バイトを hex エンコードすると 64 文字
		assert.Equal(t, hashToken(raw), hash)
	})
}

func TestHashToken(t *testing.T) {
	h1 := hashToken("raw-token-abc")
	h2 := hashToken("raw-token-abc")
	h3 := hashToken("raw-token-xyz")

	assert.Equal(t, h1, h2, "同じ入力からは同じハッシュが得られる")
	assert.NotEqual(t, h1, h3, "異なる入力からは異なるハッシュが得られる")
	assert.Len(t, h1, 64) // sha256 を hex エンコードすると 64 文字
}

// ---- sendResetEmail ----

func TestPasswordResetService_SendResetEmail(t *testing.T) {
	t.Run("returns nil immediately when SMTP host is not configured", func(t *testing.T) {
		senderCalled := false
		svc := &passwordResetService{
			cfg: &PasswordResetConfig{},
			sendMail: func(
				context.Context,
				*PasswordResetConfig,
				string,
				string,
				[]byte,
			) error {
				senderCalled = true
				return nil
			},
		}

		err := svc.sendResetEmail(context.Background(), "owner@example.com", "https://example.com/reset-password?token=abc")

		assert.NoError(t, err)
		assert.False(t, senderCalled)
	})

	t.Run("propagates transport error", func(t *testing.T) {
		transportErr := errors.New("smtp unavailable")
		svc := &passwordResetService{cfg: &PasswordResetConfig{
			SMTPHost: "smtp.example.com",
			SMTPFrom: "noreply@example.com",
		}, sendMail: func(
			context.Context,
			*PasswordResetConfig,
			string,
			string,
			[]byte,
		) error {
			return transportErr
		}}

		sendErr := svc.sendResetEmail(context.Background(), "owner@example.com", "https://example.com/reset-password?token=abc")

		assert.ErrorIs(t, sendErr, transportErr)
	})

	t.Run("passes the reset message to the configured transport", func(t *testing.T) {
		var gotFrom, gotTo string
		var gotMessage []byte
		cfg := &PasswordResetConfig{
			SMTPHost: "smtp.example.com",
			SMTPFrom: "noreply@example.com",
		}
		svc := &passwordResetService{cfg: cfg, sendMail: func(
			_ context.Context,
			gotConfig *PasswordResetConfig,
			from, to string,
			message []byte,
		) error {
			assert.Same(t, cfg, gotConfig)
			gotFrom = from
			gotTo = to
			gotMessage = message
			return nil
		}}

		err := svc.sendResetEmail(
			context.Background(),
			"owner@example.com",
			"https://example.com/reset-password#token=abc",
		)

		require.NoError(t, err)
		assert.Equal(t, "noreply@example.com", gotFrom)
		assert.Equal(t, "owner@example.com", gotTo)
		assert.Contains(t, string(gotMessage), "https://example.com/reset-password#token=abc")
		assert.Contains(t, string(gotMessage), "有効期限：30分")
		assert.NotContains(t, string(gotMessage), "有効期限：1時間")
	})
}

func TestBuildPasswordResetURL(t *testing.T) {
	t.Parallel()

	got := buildPasswordResetURL("https://example.com", "secret+token/value")

	assert.Equal(t, "https://example.com/reset-password#token=secret%2Btoken%2Fvalue", got)
	assert.NotContains(t, got, "?token=")
}
