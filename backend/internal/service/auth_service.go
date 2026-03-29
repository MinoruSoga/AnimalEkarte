package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const (
	refreshTokenDuration  = 7 * 24 * time.Hour
	passwordResetDuration = 1 * time.Hour
)

// AuthService はトークン管理・パスワードリセットのビジネスロジック
type AuthService interface {
	// CreateRefreshToken は新しいリフレッシュトークンを発行する
	CreateRefreshToken(ctx context.Context, userID, clinicID uint64) (rawToken string, err error)
	// RefreshToken はリフレッシュトークンを検証して新しいアクセストークン用情報とローテーション済みトークンを返す
	RefreshToken(ctx context.Context, rawToken string) (uint64, uint64, string, error)
	// RevokeRefreshToken はリフレッシュトークンを無効化する（ログアウト）
	RevokeRefreshToken(ctx context.Context, rawToken string) error
	// RevokeAllUserTokens はユーザーの全リフレッシュトークンを無効化する（全端末ログアウト）
	RevokeAllUserTokens(ctx context.Context, userID uint64) error

	// ForgotPassword はパスワードリセットトークンを生成してメールを送信する（dev環境ではログ出力）
	ForgotPassword(ctx context.Context, email string) error
	// ResetPassword はリセットトークンを検証してパスワードを更新し、全セッションを無効化する
	ResetPassword(ctx context.Context, rawToken, newPassword string) error

	// ChangePassword は現在のパスワードを検証して新しいパスワードに変更し、全セッションを無効化する
	ChangePassword(ctx context.Context, userID uint64, currentPassword, newPassword string) error
}

type authService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserAccountRepository
}

// NewAuthService はAuthServiceを初期化して返す
func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserAccountRepository) AuthService {
	return &authService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (s *authService) CreateRefreshToken(ctx context.Context, userID, clinicID uint64) (string, error) {
	rawToken, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	token := &model.RefreshToken{
		UserID:    userID,
		ClinicID:  clinicID,
		TokenHash: sha256Hex(rawToken),
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := s.authRepo.CreateRefreshToken(ctx, token); err != nil {
		return "", fmt.Errorf("save refresh token: %w", err)
	}
	return rawToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, rawToken string) (userID uint64, clinicID uint64, newToken string, err error) { //nolint:gocritic // named results for clarity
	hash := sha256Hex(rawToken)
	token, findErr := s.authRepo.FindRefreshToken(ctx, hash)
	if findErr != nil {
		return 0, 0, "", apperrors.WrapInvalidInput("invalid or expired refresh token")
	}
	if !token.IsValid() {
		return 0, 0, "", apperrors.WrapInvalidInput("refresh token is expired or revoked")
	}

	// ローテーション: 古いトークンを無効化して新しいトークンを発行
	if revokeErr := s.authRepo.RevokeRefreshToken(ctx, hash); revokeErr != nil {
		return 0, 0, "", fmt.Errorf("revoke old refresh token: %w", revokeErr)
	}

	newRaw, createErr := s.CreateRefreshToken(ctx, token.UserID, token.ClinicID)
	if createErr != nil {
		return 0, 0, "", fmt.Errorf("create new refresh token: %w", createErr)
	}

	return token.UserID, token.ClinicID, newRaw, nil
}

func (s *authService) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	hash := sha256Hex(rawToken)
	return s.authRepo.RevokeRefreshToken(ctx, hash)
}

func (s *authService) RevokeAllUserTokens(ctx context.Context, userID uint64) error {
	return s.authRepo.RevokeAllUserTokens(ctx, userID)
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	account, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// セキュリティのためアカウントが存在しなくてもエラーを返さない（タイミング攻撃防止）
		slog.InfoContext(ctx, "forgot password requested for unknown or errored email", slog.String("email", email))
		return nil //nolint:nilerr // intentional: hide user existence for security
	}

	rawToken, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("generate password reset token: %w", err)
	}

	resetToken := &model.PasswordResetToken{
		UserID:    account.ID,
		TokenHash: sha256Hex(rawToken),
		ExpiresAt: time.Now().Add(passwordResetDuration),
	}
	if err := s.authRepo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return fmt.Errorf("save password reset token: %w", err)
	}

	// dev環境ではログ出力（本番環境ではメール送信）
	slog.InfoContext(ctx, "password reset token generated",
		slog.String("email", email),
		slog.Uint64("user_id", account.ID),
		slog.String("raw_token", rawToken))

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return apperrors.WrapInvalidInput("password must be at least 8 characters")
	}

	hash := sha256Hex(rawToken)
	token, err := s.authRepo.FindPasswordResetToken(ctx, hash)
	if err != nil {
		return apperrors.WrapInvalidInput("invalid or expired reset token")
	}
	if !token.IsValid() {
		return apperrors.WrapInvalidInput("reset token is expired or already used")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.userRepo.Update(ctx, token.UserID, map[string]any{
		"password_hash": string(passwordHash),
	}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.authRepo.MarkPasswordResetTokenUsed(ctx, hash); err != nil {
		return fmt.Errorf("mark reset token as used: %w", err)
	}

	// 全セッションを無効化する
	if err := s.authRepo.RevokeAllUserTokens(ctx, token.UserID); err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}

	slog.InfoContext(ctx, "password reset completed", slog.Uint64("user_id", token.UserID))
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uint64, currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return apperrors.WrapInvalidInput("new password must be at least 8 characters")
	}

	account, err := s.userRepo.FindActiveByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(currentPassword)); err != nil {
		return apperrors.WrapInvalidInput("current password is incorrect")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	if err := s.userRepo.Update(ctx, userID, map[string]any{
		"password_hash": string(passwordHash),
	}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// 全セッションを無効化する（全端末ログアウト）
	if err := s.authRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}

	slog.InfoContext(ctx, "password changed", slog.Uint64("user_id", userID))
	return nil
}
