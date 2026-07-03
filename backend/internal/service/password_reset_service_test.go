package service

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockAccountRepository は account_service_test.go で定義済み

// ---- PasswordResetToken モック ----

type mockPasswordResetTokenRepository struct {
	createFn            func(ctx context.Context, token *model.PasswordResetToken) error
	findByTokenHashFn   func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
	deleteByAccountIDFn func(ctx context.Context, accountID uint64) error
	deleteByIDFn        func(ctx context.Context, id uint64) error
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

func newTestPasswordResetService(accountRepo *mockAccountRepository, tokenRepo *mockPasswordResetTokenRepository) PasswordResetService {
	return NewPasswordResetService(&PasswordResetConfig{FrontendURL: "https://example.com"}, accountRepo, tokenRepo)
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

// ---- ResetPassword ----

func TestPasswordResetService_ResetPassword(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)
	tooLongPassword := strings.Repeat("a", 73) // bcrypt はUTF-8バイト長72を超えるパスワードを拒否する

	tests := []struct {
		name              string
		findByTokenHashFn func(ctx context.Context, hash string) (*model.PasswordResetToken, error)
		deleteByIDFn      func(ctx context.Context, id uint64) error
		updateFn          func(ctx context.Context, id uint64, fields map[string]any) error
		newPassword       string
		wantErr           bool
		wantInvalidInput  bool
	}{
		{
			name: "returns invalid input when token is not found",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return nil, errors.New("not found")
			},
			newPassword:      "newpass123",
			wantErr:          true,
			wantInvalidInput: true,
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
			name: "returns invalid input even when expired token deletion fails",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: past}, nil
			},
			deleteByIDFn: func(_ context.Context, _ uint64) error {
				return errors.New("db error")
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
			updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
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
			name: "succeeds even when used token deletion fails",
			findByTokenHashFn: func(_ context.Context, _ string) (*model.PasswordResetToken, error) {
				return &model.PasswordResetToken{ID: 1, AccountID: 1, ExpiresAt: future}, nil
			},
			deleteByIDFn: func(_ context.Context, _ uint64) error {
				return errors.New("db error")
			},
			newPassword: "newpass123",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := &mockAccountRepository{updateFn: tt.updateFn}
			tokenRepo := &mockPasswordResetTokenRepository{
				findByTokenHashFn: tt.findByTokenHashFn,
				deleteByIDFn:      tt.deleteByIDFn,
			}
			svc := newTestPasswordResetService(accountRepo, tokenRepo)

			err := svc.ResetPassword(context.Background(), "raw-token", tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalidInput {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
		svc := &passwordResetService{cfg: &PasswordResetConfig{}}

		err := svc.sendResetEmail("owner@example.com", "https://example.com/reset-password?token=abc")

		assert.NoError(t, err)
	})

	t.Run("returns error when SMTP server is unreachable", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := ln.Addr().String()
		require.NoError(t, ln.Close()) // 直後に閉じて接続拒否させる

		host, port, splitErr := net.SplitHostPort(addr)
		require.NoError(t, splitErr)

		svc := &passwordResetService{cfg: &PasswordResetConfig{
			SMTPHost: host,
			SMTPPort: port,
			SMTPUser: "user",
			SMTPPass: "pass",
			SMTPFrom: "noreply@example.com",
		}}

		sendErr := svc.sendResetEmail("owner@example.com", "https://example.com/reset-password?token=abc")

		assert.Error(t, sendErr)
	})
}
