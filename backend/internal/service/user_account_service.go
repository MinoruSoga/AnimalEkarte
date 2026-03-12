package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/repository"
)

// UserAccountService はユーザーアカウントのビジネスロジック
type UserAccountService interface {
	FindByEmail(ctx context.Context, email string) (*UserAccountResult, error)
	GetMemberships(ctx context.Context, userID uuid.UUID) ([]MembershipResult, error)
	GetWithMemberships(ctx context.Context, userIDStr string) (*repository.UserAccountWithMemberships, error)
}

// UserAccountResult は FindByEmail の結果
type UserAccountResult struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	UserType     string
	Status       string
	PasswordHash string
}

// MembershipResult は GetMemberships の結果
type MembershipResult struct {
	ClinicID uuid.UUID
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
		return nil, err
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
func (s *userAccountService) GetMemberships(ctx context.Context, userID uuid.UUID) ([]MembershipResult, error) {
	data, err := s.repo.FindByIDWithMemberships(ctx, userID)
	if err != nil {
		return nil, err
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
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByIDWithMemberships(ctx, id)
}
