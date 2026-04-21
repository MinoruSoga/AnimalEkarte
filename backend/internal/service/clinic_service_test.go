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
	findAllFn               func(ctx context.Context) ([]model.Clinic, error)
	findByStaffIDFn         func(ctx context.Context, staffID uint64) ([]model.Clinic, error)
	findByIDFn              func(ctx context.Context, id uint64) (*model.Clinic, error)
	getCompanyFn            func(ctx context.Context) (*model.Company, error)
	createFn                func(ctx context.Context, clinic *model.Clinic) error
	updateFn                func(ctx context.Context, id uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, id uint64) error
	countOwnersByClinicIDFn func(ctx context.Context, clinicID uint64) (int64, error)
	countStaffByClinicIDFn  func(ctx context.Context, clinicID uint64) (int64, error)
}

// mockPermissionGroupRepositoryForClinic は PermissionGroupRepository の最小限モック（clinic_service_test用）
type mockPermissionGroupRepositoryForClinic struct {
	createFn                  func(ctx context.Context, group *model.PermissionGroup) error
	countUsageByGroupIDFn     func(ctx context.Context, clinicID, groupID uint64) (int64, error)
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	updateFieldsFn            func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	setRulesFn                func(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error
	reorderFn                 func(ctx context.Context, clinicID uint64, ids []uint64) error
	getEffectivePermissionsFn func(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error)
	getGroupIDsByStaffIDFn    func(ctx context.Context, staffID uint64) ([]uint64, error)
	setStaffGroupsFn          func(ctx context.Context, staffID uint64, groupIDs []uint64) error
}

func (m *mockPermissionGroupRepositoryForClinic) Create(ctx context.Context, group *model.PermissionGroup) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, group)
}

func (m *mockPermissionGroupRepositoryForClinic) CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	if m.countUsageByGroupIDFn == nil {
		return 0, nil
	}
	return m.countUsageByGroupIDFn(ctx, clinicID, groupID)
}

func (m *mockPermissionGroupRepositoryForClinic) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	if m.findAllFn == nil {
		return nil, nil
	}
	return m.findAllFn(ctx, clinicID)
}

func (m *mockPermissionGroupRepositoryForClinic) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	if m.findByIDFn == nil {
		return nil, nil
	}
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockPermissionGroupRepositoryForClinic) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error) {
	if m.updateFieldsFn == nil {
		return nil, nil
	}
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockPermissionGroupRepositoryForClinic) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockPermissionGroupRepositoryForClinic) SetRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if m.setRulesFn == nil {
		return nil
	}
	return m.setRulesFn(ctx, groupID, rules)
}

func (m *mockPermissionGroupRepositoryForClinic) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn == nil {
		return nil
	}
	return m.reorderFn(ctx, clinicID, ids)
}

func (m *mockPermissionGroupRepositoryForClinic) FindEffectivePermissionsByStaffID(ctx context.Context, staffID uint64) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsFn == nil {
		return nil, nil
	}
	return m.getEffectivePermissionsFn(ctx, staffID)
}

func (m *mockPermissionGroupRepositoryForClinic) FindGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error) {
	if m.getGroupIDsByStaffIDFn == nil {
		return nil, nil
	}
	return m.getGroupIDsByStaffIDFn(ctx, staffID)
}

func (m *mockPermissionGroupRepositoryForClinic) ReplaceStaffGroups(ctx context.Context, staffID uint64, groupIDs []uint64) error {
	if m.setStaffGroupsFn == nil {
		return nil
	}
	return m.setStaffGroupsFn(ctx, staffID, groupIDs)
}

func (m *mockClinicRepository) FindAll(ctx context.Context) ([]model.Clinic, error) {
	return m.findAllFn(ctx)
}

func (m *mockClinicRepository) FindByStaffID(ctx context.Context, staffID uint64) ([]model.Clinic, error) {
	if m.findByStaffIDFn == nil {
		return nil, nil
	}
	return m.findByStaffIDFn(ctx, staffID)
}

func (m *mockClinicRepository) FindByID(ctx context.Context, id uint64) (*model.Clinic, error) {
 if m.findByIDFn != nil {
	return m.findByIDFn(ctx, id)
 }
 return nil, nil
}

func (m *mockClinicRepository) GetCompany(ctx context.Context) (*model.Company, error) {
	return m.getCompanyFn(ctx)
}

func (m *mockClinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {
	return m.createFn(ctx, clinic)
}

func (m *mockClinicRepository) Update(ctx context.Context, id uint64, fields map[string]any) error {
	return m.updateFn(ctx, id, fields)
}

func (m *mockClinicRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockClinicRepository) CountOwnersByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	if m.countOwnersByClinicIDFn == nil {
		return 0, nil
	}
	return m.countOwnersByClinicIDFn(ctx, clinicID)
}

func (m *mockClinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
	if m.countStaffByClinicIDFn == nil {
		return 0, nil
	}
	return m.countStaffByClinicIDFn(ctx, clinicID)
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
			pgRepo := &mockPermissionGroupRepositoryForClinic{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo)

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
			pgRepo := &mockPermissionGroupRepositoryForClinic{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo)

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
		input          *CreateClinicInput
		repoCompany    *model.Company
		repoCompanyErr error
		repoCreateErr  error
		wantErr        bool
		wantCompanyID  uint64
	}{
		{
			name:          "creates clinic successfully with company id set",
			input:         &CreateClinicInput{Name: "新規院"},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name:           "returns error when company retrieval fails",
			input:          &CreateClinicInput{Name: "新規院"},
			repoCompanyErr: apperrors.WrapNotFound("company", "singleton"),
			wantErr:        true,
		},
		{
			name:          "returns error when clinic creation fails",
			input:         &CreateClinicInput{Name: "既存院"},
			repoCompany:   &model.Company{ID: 5, Name: "グループ本社"},
			repoCreateErr: apperrors.WrapAlreadyExists("clinic", "既存院"),
			wantErr:       true,
		},
		{
			name:          "returns error on repository failure",
			input:         &CreateClinicInput{Name: "エラー院"},
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
			pgRepo := &mockPermissionGroupRepositoryForClinic{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo)

			result, err := svc.CreateClinic(context.Background(), tt.input)

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
		input         *UpdateClinicInput
		repoClinic    *model.Clinic
		repoFindErr   error
		repoUpdateErr error
		wantErr       bool
		wantNF        bool
		wantCompanyID uint64
	}{
		{
			name: "updates clinic successfully and returns fresh record from DB",
			id:   1,
			input: &UpdateClinicInput{
				Name:    strPtr("更新後院"),
				Address: strPtr("東京都渋谷区"),
			},
			repoClinic: &model.Clinic{
				ID:        1,
				CompanyID: 5,
				Name:      "更新後院",
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			wantErr:       false,
			wantCompanyID: 5,
		},
		{
			name:        "returns not found error when clinic does not exist",
			id:          999,
			input:       &UpdateClinicInput{Name: strPtr("存在しない院")},
			repoClinic:  nil,
			repoFindErr: apperrors.WrapNotFound("clinic", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:  "returns error on update failure",
			id:    1,
			input: &UpdateClinicInput{Name: strPtr("更新後院")},
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
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			pgRepo := &mockPermissionGroupRepositoryForClinic{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo)

			result, err := svc.UpdateClinic(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				// 更新後に FindByID でリフレッシュした結果が返ることを確認
				assert.Equal(t, tt.wantCompanyID, result.CompanyID)
				assert.Equal(t, tt.repoClinic.ID, result.ID)
			}
		})
	}
}

func TestClinicService_DeleteClinic(t *testing.T) {
	tests := []struct {
		name          string
		id            uint64
		ownerCount    int64
		staffCount    int64
		countOwnerErr error
		countStaffErr error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes clinic successfully when no dependencies exist",
			id:            1,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       false,
		},
		{
			name:          "returns conflict error when clinic has owners",
			id:            1,
			ownerCount:    5,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns conflict error when clinic has staff",
			id:            1,
			ownerCount:    0,
			staffCount:    3,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns conflict error when clinic has both owners and staff",
			id:            1,
			ownerCount:    5,
			staffCount:    3,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:          "returns error when owner count check fails",
			id:            1,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: errors.New("db error"),
			countStaffErr: nil,
			repoErr:       nil,
			wantErr:       true,
		},
		{
			name:          "returns not found error when clinic does not exist",
			id:            999,
			ownerCount:    0,
			staffCount:    0,
			countOwnerErr: nil,
			countStaffErr: nil,
			repoErr:       apperrors.WrapNotFound("clinic", "999"),
			wantErr:       true,
			wantNF:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicRepository{
				countOwnersByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.ownerCount, tt.countOwnerErr
				},
				countStaffByClinicIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.staffCount, tt.countStaffErr
				},
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			pgRepo := &mockPermissionGroupRepositoryForClinic{
				createFn: func(_ context.Context, _ *model.PermissionGroup) error { return nil },
			}
			svc := NewClinicService(repo, pgRepo)

			err := svc.DeleteClinic(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
