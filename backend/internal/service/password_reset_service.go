package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/config"
	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const passwordResetTokenExpiry = time.Hour

// PasswordResetService はパスワードリセットのユースケースを定義する。
type PasswordResetService interface {
	// ForgotPassword はリセットメールを送信する。
	// アカウントが存在しない場合も 200 を返す（メール存在有無の漏洩防止）。
	ForgotPassword(ctx context.Context, email string) error
	// ResetPassword は rawToken と新パスワードでパスワードを更新する。
	ResetPassword(ctx context.Context, rawToken, newPassword string) error
	// Wait は送信中のパスワードリセットメール goroutine（fire-and-forget）の完了を待つ。
	// graceful shutdown で server.Shutdown(ctx) の後に呼び出し、goroutine の孤児化を防ぐ
	// （PERF-FOLLOWUP-05）。
	Wait()
}

// PasswordResetConfig は SMTP とフロントエンド URL の設定を保持する。
type PasswordResetConfig struct {
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
	FrontendURL string
}

type passwordResetService struct {
	cfg         *PasswordResetConfig
	accountRepo repository.AccountRepository
	tokenRepo   repository.PasswordResetTokenRepository
	// wg は fire-and-forget のメール送信 goroutine を追跡する（PERF-FOLLOWUP-05）。
	// shutdown 時に Wait() で drain し、goroutine の孤児化を防ぐ。
	wg sync.WaitGroup
}

// NewPasswordResetService は PasswordResetService の実装を返す。
func NewPasswordResetService(
	cfg *PasswordResetConfig,
	accountRepo repository.AccountRepository,
	tokenRepo repository.PasswordResetTokenRepository,
) PasswordResetService {
	return &passwordResetService{
		cfg:         cfg,
		accountRepo: accountRepo,
		tokenRepo:   tokenRepo,
	}
}

// ForgotPassword はリセットリンクメールを送信する。
// アカウントが見つからない場合もエラーを返さず正常終了する（情報漏洩防止）。
func (s *passwordResetService) ForgotPassword(ctx context.Context, email string) error {
	account, err := s.accountRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperrors.IsNotFound(err) {
			// アカウント不在でも呼び出し元には成功を返す
			slog.InfoContext(ctx, "forgot password: account not found, silently returning",
				slog.String("email", email))
			return nil
		}
		slog.ErrorContext(ctx, "failed to find account", "error", err)
		return apperrors.Wrap(err, "failed to find account")
	}

	// 既存トークンを削除してから新規発行
	if err := s.tokenRepo.DeleteByAccountID(ctx, account.ID); err != nil {
		slog.ErrorContext(ctx, "failed to clean up existing tokens", "error", err, "account_id", account.ID)
		return apperrors.Wrap(err, "failed to clean up existing tokens")
	}

	rawToken, tokenHash, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate token", "error", err)
		return apperrors.Wrap(err, "failed to generate token")
	}

	prt := &model.PasswordResetToken{
		AccountID: account.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(passwordResetTokenExpiry),
	}
	if err := s.tokenRepo.Create(ctx, prt); err != nil {
		slog.ErrorContext(ctx, "failed to create reset token", "error", err, "account_id", account.ID)
		return apperrors.Wrap(err, "failed to create reset token")
	}

	// メール送信は非同期（fire-and-forget）。リクエスト ctx はすでにキャンセル済みの
	// 可能性があるため context.Background() + 独立タイムアウトを使用する。
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, rawToken)
	s.wg.Add(1)
	goSafe("password reset email", func() { //nolint:contextcheck // fire-and-forget: request ctx キャンセル後も送信継続が必要なため context.Background を使用
		defer s.wg.Done()
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:gosec // 上記と同理由
		defer cancel()
		if sendErr := s.sendResetEmail(bgCtx, email, resetURL); sendErr != nil {
			slog.ErrorContext(bgCtx, "failed to send password reset email",
				slog.String("email", email),
				slog.String("error", sendErr.Error()))
		}
	})

	slog.InfoContext(ctx, "password reset token issued",
		slog.Uint64("account_id", account.ID))
	return nil
}

// ResetPassword は rawToken を検証してパスワードを更新する。
func (s *passwordResetService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	tokenHash := hashToken(rawToken)

	prt, err := s.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		// トークン不在も 400 として扱う（not found を伝える必要はない）
		return apperrors.WrapInvalidInput("invalid or expired token")
	}

	if time.Now().After(prt.ExpiresAt) {
		// 期限切れトークンを削除してからエラーを返す
		if delErr := s.tokenRepo.DeleteByID(ctx, prt.ID); delErr != nil {
			slog.WarnContext(ctx, "failed to delete expired reset token", "error", delErr, "token_id", prt.ID)
		}
		return apperrors.WrapInvalidInput("token has expired")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), config.BcryptCost)
	if err != nil {
		slog.ErrorContext(ctx, "failed to hash password", "error", err)
		return apperrors.Wrap(err, "failed to hash password")
	}

	if err := s.accountRepo.Update(ctx, prt.AccountID, map[string]any{
		"password_hash": string(hashedPassword),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to update password", "error", err, "account_id", prt.AccountID)
		return apperrors.Wrap(err, "failed to update password")
	}

	// 使用済みトークンを削除
	if err := s.tokenRepo.DeleteByID(ctx, prt.ID); err != nil {
		slog.WarnContext(ctx, "failed to delete used reset token",
			slog.Uint64("token_id", prt.ID),
			slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "password reset completed",
		slog.Uint64("account_id", prt.AccountID))
	return nil
}

// Wait は送信中のパスワードリセットメール goroutine の完了を待つ（PERF-FOLLOWUP-05）。
// main.go の graceful shutdown から server.Shutdown(ctx) の後に呼び出す。
func (s *passwordResetService) Wait() {
	s.wg.Wait()
}

// ---- ヘルパー ----

// generateToken は URL 用の rawToken と DB 保存用の SHA256 ハッシュを返す。
func generateToken() (rawToken, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("crypto/rand: %w", err)
	}
	rawToken = hex.EncodeToString(b)
	tokenHash = hashToken(rawToken)
	return rawToken, tokenHash, nil
}

// hashToken は rawToken の SHA256 ハッシュを hex 文字列で返す。
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

func (s *passwordResetService) sendResetEmail(ctx context.Context, to, resetURL string) error {
	if s.cfg.SMTPHost == "" {
		return nil
	}

	subject := "パスワードリセットのご案内"
	body := fmt.Sprintf(
		"パスワードリセットのリクエストを受け付けました。\r\n\r\n"+
			"以下のリンクからパスワードを変更してください（有効期限：1時間）：\r\n\r\n"+
			"%s\r\n\r\n"+
			"このメールに心当たりがない場合は無視してください。",
		resetURL,
	)

	from := s.cfg.SMTPFrom
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	cfg := smtpConfig{Host: s.cfg.SMTPHost, Port: s.cfg.SMTPPort, User: s.cfg.SMTPUser, Pass: s.cfg.SMTPPass}
	return sendSMTPMail(ctx, cfg, from, to, msg)
}
