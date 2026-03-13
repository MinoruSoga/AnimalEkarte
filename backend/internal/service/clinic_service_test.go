package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockClinicRepository は ClinicRepository のテスト用モック実装
type mockClinicRepository struct {
	findAllFn       func(ctx context.Context) ([]model.Clinic, error)
	findByIDFn      func(ctx context.Context, id uint64) (*model.Clinic, error)
	getCompanyFn    func(ctx context.Context) (*model.Company, error)
	updateCompanyFn func(ctx context.Context, company *model.Company) error
	createFn        func(ctx context.Context, clinic *model.Clinic) error
	updateFn        func(ctx context.Context, clinic *model.Clinic) error
	deleteFn        func(ctx context.Context, id uint64) error
}

func (m *mockClinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	return m.findAllFn(ctx)
}

func (m *mockClinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockClinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	return m.getCompanyFn(ctx)
}

func (m *mockClinicRepository) UpdateCompany(ctx context.Context, company *model.Company) error {
	return m.updateCompanyFn(ctx, company)
}

func (m *mockClinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	return m.createFn(ctx, clinic)
}

func (m *mockClinicRepository) Update(ctx context.Context, clinic *model.Clinic) error {
	return m.updateFn(ctx, clinic)
}

func (m *mockClinicRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func TestClinicService_ListClinics(t *testing.T) {
	tests := []struct {
		name        string
		repoClinics []model.Clinic
		repoErr     error
		wantLen     int
		wantErr     bool
	}{
		{
			name: "returns clinic list",
			repoClinics: []model.Clinic{
				{ID: 1, CompanyID: 1, Name: "本院"},
				{ID: 2, CompanyID: 1, Name: "分院"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:        "returns empty list when no clinics exist",
			repoClinics: []model.Clinic{},
			repoErr:     nil,
			wantLen:     0,
			wantErr:     false,
		},
		{
			name:        "propagates repository error",
			repoClinics: nil,
			repoErr:     errors.New("db connection error"),
			wantLen:     0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findAllFn: func(_ context.Context) ([]model.Clinic, error) {
					return tt.repoClinics, tt.repoErr
				},
			}
			svc := NewClinicService(repo)

			clinics, err := svc.ListClinics(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, clinics, tt.wantLen)
			}
		})
	}
}

func TestClinicService_GetClinicByID(t *testing.T) {
	tests := []struct {
		name       string
		id         uint64
		repoClinic *model.Clinic
		repoErr    error
		wantErr    bool
		wantNF     bool
	}{
		{
			name: "returns clinic when found",
			id:   1,
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 1,
				Name:      "本院",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:       "returns not found error when clinic does not exist",
			id:         999,
			repoClinic: nil,
			repoErr:    apperrors.WrapNotFound("clinic", "999"),
			wantErr:    true,
			wantNF:     true,
		},
		{
			name:       "returns error on repository failure",
			id:         1,
			repoClinic: nil,
			repoErr:    errors.New("db error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
					return tt.repoClinic, tt.repoErr
				},
			}
			svc := NewClinicService(repo)

			clinic, err := svc.GetClinicByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoClinic, clinic)
			}
		})
	}
}

func TestClinicService_CreateClinic(t *testing.T) {
	tests := []struct {
		name           string
		clinic         *model.Clinic
		repoCompany    *model.Company
		repoCompanyErr error
		repoCreateErr  error
		wantErr        bool
		wantCompanyID  uint64
	}{
		{
			name: "creates clinic successfully with company id set",
			clinic: &model.Clinic{
				Name: "新規院",
			},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name: "returns error when company retrieval fails",
			clinic: &model.Clinic{
				Name: "新規院",
			},
			repoCompanyErr: apperrors.WrapNotFound("company", "singleton"),
			wantErr:        true,
		},
		{
			name: "returns error when clinic creation fails",
			clinic: &model.Clinic{
				Name: "既存院",
			},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			repoCreateErr: apperrors.WrapAlreadyExists("clinic", "既存院"),
			wantErr:       true,
		},
		{
			name: "returns error on repository failure",
			clinic: &model.Clinic{
				Name: "エラー院",
			},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			repoCreateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				getCompanyFn: func(_ context.Context) (*model.Company, error) {
					return tt.repoCompany, tt.repoCompanyErr
				},
				createFn: func(_ context.Context, _ *model.Clinic) error {
					return tt.repoCreateErr
				},
			}
			svc := NewClinicService(repo)

			result, err := svc.CreateClinic(context.Background(), tt.clinic)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantCompanyID, result.CompanyID)
			}
		})
	}
}

func TestClinicService_UpdateClinic(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		inputClinic   *model.Clinic
		repoClinic    *model.Clinic
		repoFindErr   error
		repoUpdateErr error
		wantErr       bool
		wantNF        bool
		wantCompanyID uint64
	}{
		{
			name: "updates clinic successfully with immutable fields preserved",
			id:   1,
			inputClinic: &model.Clinic{
				Name:    "更新後院",
				Address: "東京都渋谷区",
			},
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 5,
				Name:      "旧院名",
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name: "returns not found error when clinic does not exist",
			id:   999,
			inputClinic: &model.Clinic{
				Name: "存在しない院",
			},
			repoClinic:  nil,
			repoFindErr: apperrors.WrapNotFound("clinic", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name: "returns error on update failure",
			id:   1,
			inputClinic: &model.Clinic{
				Name: "更新後院",
			},
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 5,
				Name:      "旧院名",
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				findByIDFn: func(_ context.Context, _ uint64) (*model.Clinic, error) {
					return tt.repoClinic, tt.repoFindErr
				},
				updateFn: func(_ context.Context, _ *model.Clinic) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewClinicService(repo)

			result, err := svc.UpdateClinic(context.Background(), tt.id, tt.inputClinic)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// immutable フィールドが既存レコードから引き継がれていることを確認
				assert.Equal(t, tt.wantCompanyID, result.CompanyID)
				assert.Equal(t, tt.repoClinic.ID, result.ID)
			}
		})
	}
}

func TestClinicService_DeleteClinic(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes clinic successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when clinic does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("clinic", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewClinicService(repo)

			err := svc.DeleteClinic(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClinicService_GetCompany(t *testing.T) {
	tests := []struct {
		name        string
		repoCompany *model.Company
		repoErr     error
		wantErr     bool
		wantNF      bool
	}{
		{
			name: "returns company when found",
			repoCompany: &model.Company{
				ID:   1,
				Name: "グループ本社",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns not found error when company does not exist",
			repoCompany: nil,
			repoErr:     apperrors.WrapNotFound("company", "singleton"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:        "returns error on repository failure",
			repoCompany: nil,
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				getCompanyFn: func(_ context.Context) (*model.Company, error) {
					return tt.repoCompany, tt.repoErr
				},
			}
			svc := NewClinicService(repo)

			company, err := svc.GetCompany(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCompany, company)
			}
		})
	}
}

func TestClinicService_UpdateCompany(t *testing.T) {
	tests := []struct {
		name    string
		company *model.Company
		repoErr error
		wantErr bool
	}{
		{
			name: "updates company successfully",
			company: &model.Company{
				Name:    "更新後本社",
				Address: "東京都千代田区",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error on repository failure",
			company: &model.Company{
				Name: "エラー本社",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				updateCompanyFn: func(_ context.Context, _ *model.Company) error {
					return tt.repoErr
				},
			}
			svc := NewClinicService(repo)

			err := svc.UpdateCompany(context.Background(), tt.company)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
