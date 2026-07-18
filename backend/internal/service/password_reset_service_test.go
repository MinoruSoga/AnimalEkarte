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

	"github.com/animal-ekarte/backend/internal/apperrors"
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

// ---- Wait (PERF-FOLLOWUP-05: shutdown drain) ----

// TestPasswordResetService_Wait_DrainsInFlightEmailGoroutine は、ForgotPassword が起動する
// fire-and-forget メール送信 goroutine を Wait() が実際に待機することを証明する。
// 修正前（wg なし）は Wait() 自体が存在せずコンパイル不可 — 本テストは修正と一体で追加する
// 加算的信頼性修正のため、RED は「Wait 未実装」という形で表現される。
func TestPasswordResetService_Wait_DrainsInFlightEmailGoroutine(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	const serverDelay = 150 * time.Millisecond
	handled := make(chan struct{})
	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		// SMTP ハンドシェイクを遅延させ、goroutine がまだ実行中の状態を作る。
		time.Sleep(serverDelay)
		close(handled)
	}()

	host, port, splitErr := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, splitErr)

	accountRepo := &mockAccountRepository{
		findByEmailFn: func(_ context.Context, email string) (*model.Account, error) {
			return &model.Account{ID: 1, Email: email}, nil
		},
	}
	tokenRepo := &mockPasswordResetTokenRepository{}

	svc := &passwordResetService{
		cfg: &PasswordResetConfig{
			FrontendURL: "https://example.com",
			SMTPHost:    host,
			SMTPPort:    port,
			SMTPFrom:    "noreply@example.com",
		},
		accountRepo: accountRepo,
		tokenRepo:   tokenRepo,
	}

	require.NoError(t, svc.ForgotPassword(context.Background(), "owner@example.com"))

	select {
	case <-handled:
		t.Fatal("background goroutine already finished before Wait() was called — test setup invalid (delay too short), cannot prove drain behavior")
	default:
	}

	start := time.Now()
	svc.Wait()
	elapsed := time.Since(start)

	select {
	case <-handled:
		// OK: サーバ側の遅延処理が完了した後で Wait() が返った。
	default:
		t.Fatal("Wait() returned before the in-flight email goroutine finished")
	}
	assert.GreaterOrEqual(t, elapsed, serverDelay/2,
		"Wait() should block roughly until the background SMTP goroutine completes")
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

		err := svc.sendResetEmail(context.Background(), "owner@example.com", "https://example.com/reset-password?token=abc")

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

		sendErr := svc.sendResetEmail(context.Background(), "owner@example.com", "https://example.com/reset-password?token=abc")

		assert.Error(t, sendErr)
	})

	t.Run("returns error within ctx deadline when SMTP server accepts but never responds", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer ln.Close()

		accepted := make(chan struct{})
		go func() {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}
			defer conn.Close()
			close(accepted)
			// 応答を返さず接続だけ保持する（deadline が効いていなければ
			// 呼び出し元は OS の TCP タイムアウトまで無期限にブロックする）。
			time.Sleep(5 * time.Second)
		}()

		host, port, splitErr := net.SplitHostPort(ln.Addr().String())
		require.NoError(t, splitErr)

		svc := &passwordResetService{cfg: &PasswordResetConfig{
			SMTPHost: host,
			SMTPPort: port,
			SMTPFrom: "noreply@example.com",
		}}

		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		start := time.Now()
		sendErr := svc.sendResetEmail(ctx, "owner@example.com", "https://example.com/reset-password?token=abc")
		elapsed := time.Since(start)

		<-accepted // サーバ側が実際に接続を受け付けたことを確認（フレーク防止）
		assert.Error(t, sendErr)
		assert.Less(t, elapsed, 3*time.Second,
			"deadline が smtp 送信に伝播していれば OS の TCP タイムアウトを待たずに復帰するはず")
	})

	t.Run("rejects recipient address containing CRLF before dialing (SMTP command injection guard)", func(t *testing.T) {
		svc := &passwordResetService{cfg: &PasswordResetConfig{
			SMTPHost: "127.0.0.1",
			SMTPPort: "25",
			SMTPFrom: "noreply@example.com",
		}}

		sendErr := svc.sendResetEmail(context.Background(), "victim@example.com\r\nRCPT TO:<attacker@evil.example>", "https://example.com/reset-password?token=abc")

		assert.Error(t, sendErr)
		assert.Contains(t, sendErr.Error(), "CR or LF")
	})
}

func TestValidateSMTPLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{name: "plain address is accepted", line: "owner@example.com", wantErr: false},
		{name: "CR is rejected", line: "owner@example.com\rSubject: injected", wantErr: true},
		{name: "LF is rejected", line: "owner@example.com\nSubject: injected", wantErr: true},
		{name: "CRLF is rejected", line: "owner@example.com\r\nRCPT TO:<attacker@evil.example>", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSMTPLine(tt.line)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
