package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Company モック ----

type mockCompanyRepository struct {
	getFn    func(ctx context.Context) (*model.Company, error)
	updateFn func(ctx context.Context, fields map[string]any) error
}

func (m *mockCompanyRepository) FindSingleton(ctx context.Context) (*model.Company, error) {
	return m.getFn(ctx)
}

func (m *mockCompanyRepository) Update(ctx context.Context, fields map[string]any) error {
	return m.updateFn(ctx, fields)
}

// ---- Tests ----

func TestCompanyService_Get(t *testing.T) {
	tests := []struct {
		name     string
		repoData *model.Company
		repoErr  error
		wantErr  bool
	}{
		{
			name: "returns company information",
			repoData: &model.Company{
				Name:               "テスト動物病院",
				PostalCode:         "100-0001",
				Address:            "東京都千代田区千代田1-1",
				PhoneNumber:        "03-1234-5678",
				Email:              "info@example.com",
				DirectorName:       "院長太郎",
				RegistrationNumber: "1234567890123",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error on repository failure",
			repoData: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCompanyRepository{
				getFn: func(_ context.Context) (*model.Company, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewCompanyService(repo)

			company, err := svc.Get(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, company)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoData, company)
			}
		})
	}
}

func TestCompanyService_Update(t *testing.T) {
	newName := "更新動物病院"
	newPhone := "03-9999-8888"
	newEmail := "updated@example.com"

	tests := []struct {
		name        string
		input       *UpdateCompanyInput
		repoErr     error
		repoCompany *model.Company
		wantErr     bool
	}{
		{
			name: "updates company successfully",
			input: &UpdateCompanyInput{
				Name:        &newName,
				PhoneNumber: &newPhone,
			},
			repoErr: nil,
			repoCompany: &model.Company{
				Name:        "更新動物病院",
				PhoneNumber: "03-9999-8888",
			},
			wantErr: false,
		},
		{
			name: "updates multiple fields",
			input: &UpdateCompanyInput{
				Name:    &newName,
				Email:   &newEmail,
				Address: func(s string) *string { return &s }("東京都渋谷区"),
			},
			repoErr: nil,
			repoCompany: &model.Company{
				Name:    "更新動物病院",
				Email:   "updated@example.com",
				Address: "東京都渋谷区",
			},
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateCompanyInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &UpdateCompanyInput{
				Name: &newName,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCompanyRepository{
				updateFn: func(_ context.Context, _ map[string]any) error {
					return tt.repoErr
				},
				getFn: func(_ context.Context) (*model.Company, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.repoCompany, nil
				},
			}
			svc := NewCompanyService(repo)

			company, err := svc.Update(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCompany, company)
			}
		})
	}
}
