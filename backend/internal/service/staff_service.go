// Package service provides business logic implementations for Staff entity.
package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// ---- StaffService ----

// CreateStaffInput はスタッフ登録時の入力データ。
// スタッフ情報とシステムアカウント情報を同時に受け取る。
type CreateStaffInput struct {
	// Staff fields
	ClinicID      uint64
	Name          string
	StaffRole     model.StaffRole
	LicenseNumber string
	JobTitleID    *uint64
	SortOrder     int
	// Account fields
	Email    string
	Password string
}

type StaffService interface {
	List(ctx context.Context, role *string) ([]model.Staff, error)
	GetByID(ctx context.Context, id uint64) (*model.Staff, error)
	// CreateWithAccount はスタッフとシステムアカウントをアトミックに作成する。
	CreateWithAccount(ctx context.Context, input *CreateStaffInput) (*model.Staff, error)
	Update(ctx context.Context, staff *model.Staff) error
	Delete(ctx context.Context, id uint64) error
}

type staffService struct{ repo repository.StaffRepository }

func NewStaffService(repo repository.StaffRepository) StaffService {
	return &staffService{repo: repo}
}

func (s *staffService) List(ctx context.Context, role *string) ([]model.Staff, error) {
	return s.repo.FindAll(ctx, role)
}

func (s *staffService) GetByID(ctx context.Context, id uint64) (*model.Staff, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *staffService) CreateWithAccount(ctx context.Context, input *CreateStaffInput) (*model.Staff, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	staff := &model.Staff{
		ClinicID:      input.ClinicID,
		Name:          input.Name,
		StaffRole:     input.StaffRole,
		LicenseNumber: input.LicenseNumber,
		JobTitleID:    input.JobTitleID,
		SortOrder:     input.SortOrder,
		IsActive:      true,
	}

	account := &model.UserAccount{
		Email:        input.Email,
		DisplayName:  input.Name,
		UserType:     model.UserTypeStaff,
		Status:       model.AccountStatusActive,
		PasswordHash: string(hash),
	}

	membership := &model.UserClinicMembership{
		ClinicID: input.ClinicID,
		IsMain:   true,
	}

	if err := s.repo.CreateWithAccount(ctx, staff, account, membership); err != nil {
		return nil, err
	}

	return staff, nil
}

func (s *staffService) Update(ctx context.Context, staff *model.Staff) error {
	return s.repo.Update(ctx, staff)
}

func (s *staffService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
