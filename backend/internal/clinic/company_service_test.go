package clinic

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
			want:  map[string]any{"name": name},
		},
		{
			name:  "only postal_code set",
			input: &UpdateCompanyInput{PostalCode: &postalCode},
			want:  map[string]any{"postal_code": postalCode},
		},
		{
			name:  "only address set",
			input: &UpdateCompanyInput{Address: &address},
			want:  map[string]any{"address": address},
		},
		{
			name:  "only phone_number set",
			input: &UpdateCompanyInput{PhoneNumber: &phone},
			want:  map[string]any{"phone_number": phone},
		},
		{
			name:  "only fax_number set",
			input: &UpdateCompanyInput{FaxNumber: &fax},
			want:  map[string]any{"fax_number": fax},
		},
		{
			name:  "only email set",
			input: &UpdateCompanyInput{Email: &email},
			want:  map[string]any{"email": email},
		},
		{
			name:  "only website set",
			input: &UpdateCompanyInput{Website: &website},
			want:  map[string]any{"website": website},
		},
		{
			name:  "only director_name set",
			input: &UpdateCompanyInput{DirectorName: &director},
			want:  map[string]any{"director_name": director},
		},
		{
			name:  "only registration_number set",
			input: &UpdateCompanyInput{RegistrationNumber: &registrationNumber},
			want:  map[string]any{"registration_number": registrationNumber},
		},
		{
			name:  "only invoice_registration_number set",
			input: &UpdateCompanyInput{InvoiceRegistrationNumber: &invoiceRegistrationNumber},
			want:  map[string]any{"invoice_registration_number": invoiceRegistrationNumber},
		},
		{
			name:  "only logo_url set",
			input: &UpdateCompanyInput{LogoURL: &logoURL},
			want:  map[string]any{"logo_url": logoURL},
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
				"name": name, "postal_code": postalCode, "address": address,
				"phone_number": phone, "fax_number": fax, "email": email,
				"website": website, "director_name": director,
				"registration_number": registrationNumber, "invoice_registration_number": invoiceRegistrationNumber,
				"logo_url": logoURL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildCompanyUpdate(tt.input)
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

// TestCompanyService_Update_PostUpdateFindSingletonError は Update 成功後の再取得失敗で
// 5xx へ反転せず、pre-image に適用済み fields を載せた成功応答を返すこと（POC-02 / X-01）。
func TestCompanyService_Update_PostUpdateFindSingletonError(t *testing.T) {
	newName := "更新動物病院"
	findCalls := 0
	repo := &mockCompanyRepository{
		updateFn: func(_ context.Context, _ map[string]any) error {
			return nil
		},
		getFn: func(_ context.Context) (*model.Company, error) {
			findCalls++
			if findCalls == 1 {
				return &model.Company{ID: 1, Name: "旧名称"}, nil
			}
			return nil, errors.New("db error on post-update fetch")
		},
	}
	svc := NewCompanyService(repo)

	company, err := svc.Update(context.Background(), &UpdateCompanyInput{Name: &newName})

	assert.NoError(t, err)
	if assert.NotNil(t, company) {
		assert.Equal(t, newName, company.Name)
		assert.Equal(t, uint64(1), company.ID)
	}
}
