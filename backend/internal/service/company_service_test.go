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

func TestBuildCompanyUpdate(t *testing.T) {
	name := "更新法人"
	postalCode := "100-0001"
	address := "東京都千代田区"
	phone := "03-0000-0000"
	fax := "03-1111-1111"
	email := "info@example.com"
	website := "https://example.com"
	director := "院長"
	registrationNumber := "1234567890123"
	invoiceRegistrationNumber := "T1234567890123"
	logoURL := "https://example.com/logo.png"

	tests := []struct {
		name  string
		input *UpdateCompanyInput
		want  map[string]any
	}{
		{
			name:  "no fields set returns empty map",
			input: &UpdateCompanyInput{},
			want:  map[string]any{},
		},
		{
			name:  "only name set",
			input: &UpdateCompanyInput{Name: &name},
			want:  map[string]any{colCompanyName: name},
		},
		{
			name:  "only postal_code set",
			input: &UpdateCompanyInput{PostalCode: &postalCode},
			want:  map[string]any{colCompanyPostalCode: postalCode},
		},
		{
			name:  "only address set",
			input: &UpdateCompanyInput{Address: &address},
			want:  map[string]any{colCompanyAddress: address},
		},
		{
			name:  "only phone_number set",
			input: &UpdateCompanyInput{PhoneNumber: &phone},
			want:  map[string]any{colCompanyPhoneNumber: phone},
		},
		{
			name:  "only fax_number set",
			input: &UpdateCompanyInput{FaxNumber: &fax},
			want:  map[string]any{colCompanyFaxNumber: fax},
		},
		{
			name:  "only email set",
			input: &UpdateCompanyInput{Email: &email},
			want:  map[string]any{colCompanyEmail: email},
		},
		{
			name:  "only website set",
			input: &UpdateCompanyInput{Website: &website},
			want:  map[string]any{colCompanyWebsite: website},
		},
		{
			name:  "only director_name set",
			input: &UpdateCompanyInput{DirectorName: &director},
			want:  map[string]any{colCompanyDirectorName: director},
		},
		{
			name:  "only registration_number set",
			input: &UpdateCompanyInput{RegistrationNumber: &registrationNumber},
			want:  map[string]any{colCompanyRegistrationNumber: registrationNumber},
		},
		{
			name:  "only invoice_registration_number set",
			input: &UpdateCompanyInput{InvoiceRegistrationNumber: &invoiceRegistrationNumber},
			want:  map[string]any{colCompanyInvoiceRegistrationNumber: invoiceRegistrationNumber},
		},
		{
			name:  "only logo_url set",
			input: &UpdateCompanyInput{LogoURL: &logoURL},
			want:  map[string]any{colCompanyLogoURL: logoURL},
		},
		{
			name: "all fields set",
			input: &UpdateCompanyInput{
				Name: &name, PostalCode: &postalCode, Address: &address, PhoneNumber: &phone,
				FaxNumber: &fax, Email: &email, Website: &website, DirectorName: &director,
				RegistrationNumber: &registrationNumber, InvoiceRegistrationNumber: &invoiceRegistrationNumber,
				LogoURL: &logoURL,
			},
			want: map[string]any{
				colCompanyName: name, colCompanyPostalCode: postalCode, colCompanyAddress: address,
				colCompanyPhoneNumber: phone, colCompanyFaxNumber: fax, colCompanyEmail: email,
				colCompanyWebsite: website, colCompanyDirectorName: director,
				colCompanyRegistrationNumber: registrationNumber, colCompanyInvoiceRegistrationNumber: invoiceRegistrationNumber,
				colCompanyLogoURL: logoURL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCompanyUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

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

// TestCompanyService_Update_PostUpdateFindSingletonError は Update 自体は成功したが
// 直後の再取得（FindSingleton）が失敗した場合にエラーを返すことを検証する
// （Update が repoErr を Update と FindSingleton 両方に流用する既存テーブルでは分離できない分岐）。
func TestCompanyService_Update_PostUpdateFindSingletonError(t *testing.T) {
	newName := "更新動物病院"
	repo := &mockCompanyRepository{
		updateFn: func(_ context.Context, _ map[string]any) error {
			return nil
		},
		getFn: func(_ context.Context) (*model.Company, error) {
			return nil, errors.New("db error on post-update fetch")
		},
	}
	svc := NewCompanyService(repo)

	company, err := svc.Update(context.Background(), &UpdateCompanyInput{Name: &newName})

	assert.Error(t, err)
	assert.Nil(t, company)
}
