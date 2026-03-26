package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateUserAccountInput はユーザー作成の入力DTO
type CreateUserAccountInput struct {
	Email       string
	Password    string
	DisplayName string
	UserType    model.UserType
	StaffID     *uint64
	IsMain      bool
}

// UpdateUserAccountInput はユーザー更新の入力DTO（nil = 未変更）
type UpdateUserAccountInput struct {
	DisplayName *string
	UserType    *model.UserType
	AvatarURL   *string
	Status      *string
	StaffID     *uint64
}

// SetPermissionGroupsInput はグループ割当の入力DTO
type SetPermissionGroupsInput struct {
	GroupIDs []uint64
}

// UserAccountService はユーザーアカウントのビジネスロジック
type UserAccountService interface {
	FindByEmail(ctx context.Context, email string) (*UserAccountResult, error)
	GetMemberships(ctx context.Context, userID uint64) ([]MembershipResult, error)
	GetWithMemberships(ctx context.Context, userIDStr string) (*repository.UserAccountWithMemberships, error)
	ListUsers(ctx context.Context, clinicID uint64) ([]model.UserAccount, error)
	CreateUser(ctx context.Context, clinicID uint64, input *CreateUserAccountInput) (*model.UserAccount, error)
	UpdateUser(ctx context.Context, id uint64, input *UpdateUserAccountInput) error
	DeleteUser(ctx context.Context, id uint64) error
	SetPermissionGroups(ctx context.Context, userID uint64, input SetPermissionGroupsInput) error
	GetPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error)
}

// UserAccountResult は FindByEmail の結果
type UserAccountResult struct {
	ID           uint64
	Email        string
	DisplayName  string
	UserType     string
	Status       string
	PasswordHash string
}

// MembershipResult は GetMemberships の結果
type MembershipResult struct {
	ClinicID uint64
	IsMain   bool
}

type userAccountService struct {
	repo repository.UserAccountRepository
}

// NewUserAccountService は UserAccountService を初期化して返す
func NewUserAccountService(repo repository.UserAccountRepository) UserAccountService {
	return &userAccountService{repo: repo}
}

// FindByEmail はメールアドレスでユーザーアカウントを取得する
func (s *userAccountService) FindByEmail(ctx context.Context, email string) (*UserAccountResult, error) {
	account, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user account by email: %w", err)
	}
	return &UserAccountResult{
		ID:           account.ID,
		Email:        account.Email,
		DisplayName:  account.DisplayName,
		UserType:     string(account.UserType),
		Status:       string(account.Status),
		PasswordHash: account.PasswordHash,
	}, nil
}

// GetMemberships はユーザーの所属クリニック一覧を返す
func (s *userAccountService) GetMemberships(ctx context.Context, userID uint64) ([]MembershipResult, error) {
	data, err := s.repo.FindByIDWithMemberships(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user memberships: %w", err)
	}
	results := make([]MembershipResult, 0, len(data.Memberships))
	for _, m := range data.Memberships {
		results = append(results, MembershipResult{
			ClinicID: m.ClinicID,
			IsMain:   m.IsMain,
		})
	}
	return results, nil
}

// GetWithMemberships はユーザー・所属クリニック・権限を一括取得する
func (s *userAccountService) GetWithMemberships(ctx context.Context, userIDStr string) (*repository.UserAccountWithMemberships, error) {
	id, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, apperrors.Wrap(fmt.Errorf("invalid user id: %s", userIDStr), "parse user id")
	}
	return s.repo.FindByIDWithMemberships(ctx, id)
}

// ListUsers はクリニックに所属するユーザー一覧を返す
func (s *userAccountService) ListUsers(ctx context.Context, clinicID uint64) ([]model.UserAccount, error) {
	return s.repo.FindByClinicID(ctx, clinicID)
}

// CreateUser はユーザーアカウントとクリニック所属を作成する
func (s *userAccountService) CreateUser(ctx context.Context, clinicID uint64, input *CreateUserAccountInput) (*model.UserAccount, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	account := &model.UserAccount{
		Email:        input.Email,
		DisplayName:  input.DisplayName,
		UserType:     input.UserType,
		Status:       model.AccountStatusActive,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(ctx, account, clinicID, input.StaffID, input.IsMain); err != nil {
		return nil, fmt.Errorf("failed to create user account: %w", err)
	}

	slog.InfoContext(ctx, "user account created",
		slog.Uint64("user_id", account.ID),
		slog.Uint64("clinic_id", clinicID))
	return account, nil
}

// UpdateUser はユーザーアカウントを部分更新する
func (s *userAccountService) UpdateUser(ctx context.Context, id uint64, input *UpdateUserAccountInput) error {
	fields := map[string]any{}
	if input.DisplayName != nil {
		fields["display_name"] = *input.DisplayName
	}
	if input.UserType != nil {
		fields["user_type"] = *input.UserType
	}
	if input.AvatarURL != nil {
		fields["avatar_url"] = *input.AvatarURL
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.StaffID != nil {
		fields["staff_id"] = *input.StaffID
	}
	if len(fields) == 0 {
		return apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, id, fields); err != nil {
		return fmt.Errorf("failed to update user account: %w", err)
	}
	slog.InfoContext(ctx, "user account updated", slog.Uint64("user_id", id))
	return nil
}

// DeleteUser はユーザーアカウントを論理削除する
func (s *userAccountService) DeleteUser(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user account: %w", err)
	}
	slog.InfoContext(ctx, "user account deleted", slog.Uint64("user_id", id))
	return nil
}

// SetPermissionGroups はユーザーのグループ割当を全置換する
func (s *userAccountService) SetPermissionGroups(ctx context.Context, userID uint64, input SetPermissionGroupsInput) error {
	slog.InfoContext(ctx, "setting user permission groups", slog.Uint64("user_id", userID), slog.Any("group_ids", input.GroupIDs))
	return s.repo.SetPermissionGroups(ctx, userID, input.GroupIDs)
}

// GetPermissionGroupIDs はユーザーの割当グループIDリストを返す
func (s *userAccountService) GetPermissionGroupIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	return s.repo.FindPermissionGroupIDs(ctx, userID)
}
