package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockCageRepository は CageRepository のテスト用モック実装
type mockCageRepository struct {
	findAllFn            func(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error)
	findByIDFn           func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	lockByIDForUpdateFn  func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
	countUsageByCageIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	createFn             func(ctx context.Context, cage *model.Cage) error
	updateFieldsFn       func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error)
	deleteFn             func(ctx context.Context, clinicID, id uint64) error
	reorderFn            func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockCageRepository) FindAll(ctx context.Context, clinicID uint64, cageType *string) ([]model.Cage, error) {
	return m.findAllFn(ctx, clinicID, cageType)
}

func (m *mockCageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockCageRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, id)
	}
	// Default: fall back to FindByID so existing tests stay simple.
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockCageRepository) Create(ctx context.Context, cage *model.Cage) error {
	return m.createFn(ctx, cage)
}

func (m *mockCageRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockCageRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockCageRepository) CountUsageByCageID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByCageIDFn != nil {
		return m.countUsageByCageIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockCageRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func newTestCageService(repo *mockCageRepository) CageService {
	return NewCageService(repo, &mockTransactor{})
}

func newTestCageServiceWithTx(repo *mockCageRepository, tx Transactor) CageService {
	return NewCageService(repo, tx)
}

func TestCageService_List(t *testing.T) {
	cageType := string(model.CageTypeDog)

	tests := []struct {
		name      string
		cageType  *string
		repoCages []model.Cage
		repoErr   error
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns all cage list",
			repoCages: []model.Cage{
				{ID: 1, ClinicID: 1, Name: "犬用ケージA", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "猫用ケージA", CageType: model.CageTypeCat, CageSize: model.CageSizeMedium, IsActive: true},
				{ID: 3, ClinicID: 1, Name: "ICU-1", CageType: model.CageTypeICU, CageSize: model.CageSizeLarge, IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:      "returns empty list when no cages exist",
			repoCages: []model.Cage{},
			repoErr:   nil,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:     "filters by cage_type",
			cageType: &cageType,
			repoCages: []model.Cage{
				{ID: 1, ClinicID: 1, Name: "犬用ケージA", CageType: model.CageTypeDog, CageSize: model.CageSizeSmall},
			},
			repoErr: nil,
			wantLen: 1,
			wantErr: false,
		},
		{
			name:      "propagates repository error",
			repoCages: nil,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *string) ([]model.Cage, error) {
					return tt.repoCages, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cages, err := svc.List(context.Background(), 1, tt.cageType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cages, tt.wantLen)
			}
		})
	}
}

func TestCageService_GetByID(t *testing.T) {
	price := int64(5000)
	tests := []struct {
		name     string
		id       uint64
		repoCage *model.Cage
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name: "returns cage when found",
			id:   1,
			repoCage: &model.Cage{
				ID:       1,
				ClinicID: 1,
				Name:     "犬用ケージA",
				CageType: model.CageTypeDog,
				CageSize: model.CageSizeSmall,
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when cage does not exist",
			id:       999,
			repoCage: nil,
			repoErr:  apperrors.WrapNotFound("cage", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			id:       1,
			repoCage: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					return tt.repoCage, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cage, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCage, cage)
			}
		})
	}
}

func TestCageService_Create(t *testing.T) {
	price := int64(3000)
	tests := []struct {
		name    string
		input   *CreateCageInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates cage successfully",
			input: &CreateCageInput{
				Name:     "新規ケージ",
				CageType: string(model.CageTypeDog),
				CageSize: string(model.CageSizeMedium),
				Price:    &price,
				IsActive: true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when cage already exists",
			input: &CreateCageInput{
				Name:     "既存ケージ",
				CageType: string(model.CageTypeCat),
				CageSize: string(model.CageSizeSmall),
			},
			repoErr: apperrors.WrapAlreadyExists("cage", "既存ケージ"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateCageInput{
				Name:     "エラーケージ",
				CageType: string(model.CageTypeGeneral),
				CageSize: string(model.CageSizeLarge),
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "returns error when name is empty",
			input: &CreateCageInput{
				Name:     "",
				CageType: string(model.CageTypeDog),
				CageSize: string(model.CageSizeSmall),
			},
			wantErr: true,
		},
		{
			name: "returns error when cage_type is invalid",
			input: &CreateCageInput{
				Name:     "不正種別ケージ",
				CageType: "invalid_type",
				CageSize: string(model.CageSizeSmall),
			},
			wantErr: true,
		},
		{
			name: "returns error when cage_size is invalid",
			input: &CreateCageInput{
				Name:     "不正サイズケージ",
				CageType: string(model.CageTypeDog),
				CageSize: "invalid_size",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				createFn: func(_ context.Context, _ *model.Cage) error {
					return tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cage, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cage)
			}
		})
	}
}

func TestCageService_Update(t *testing.T) {
	price := int64(4500)
	name := "更新後ケージ"
	cageType := string(model.CageTypeDog)
	cageSize := string(model.CageSizeLarge)
	isActive := true
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		input    UpdateCageInput
		repoCage *model.Cage
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "updates cage successfully",
			clinicID: 1,
			id:       1,
			input: UpdateCageInput{
				Name:     &name,
				CageType: &cageType,
				CageSize: &cageSize,
				Price:    &price,
				IsActive: &isActive,
			},
			repoCage: &model.Cage{
				ID:       1,
				ClinicID: 1,
				Name:     name,
				CageType: model.CageType(cageType),
				CageSize: model.CageSize(cageSize),
				Price:    &price,
				IsActive: isActive,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    UpdateCageInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       999,
			input:    UpdateCageInput{Name: &name},
			repoCage: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
					return &model.Cage{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Cage, error) {
					return tt.repoCage, tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			cage, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cage)
			}
		})
	}
}

func TestCageService_Update_NilInput(t *testing.T) {
	repo := &mockCageRepository{}
	svc := newTestCageService(repo)
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCageService_Update_FindByIDError(t *testing.T) {
	name := "更新後ケージ"
	repo := &mockCageRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Cage, error) {
			return nil, apperrors.WrapNotFound("cage", "999")
		},
	}
	svc := newTestCageService(repo)

	cage, err := svc.Update(context.Background(), 1, 999, &UpdateCageInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, cage)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestCageService_Update_InvalidName(t *testing.T) {
	empty := ""
	repo := &mockCageRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
			return &model.Cage{ID: id}, nil
		},
	}
	svc := newTestCageService(repo)

	cage, err := svc.Update(context.Background(), 1, 1, &UpdateCageInput{Name: &empty})

	assert.Error(t, err)
	assert.Nil(t, cage)
}

func TestCageService_Update_InvalidCageType(t *testing.T) {
	invalid := "invalid_type"
	repo := &mockCageRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
			return &model.Cage{ID: id}, nil
		},
	}
	svc := newTestCageService(repo)

	cage, err := svc.Update(context.Background(), 1, 1, &UpdateCageInput{CageType: &invalid})

	assert.Error(t, err)
	assert.Nil(t, cage)
}

func TestCageService_Update_InvalidCageSize(t *testing.T) {
	invalid := "invalid_size"
	repo := &mockCageRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
			return &model.Cage{ID: id}, nil
		},
	}
	svc := newTestCageService(repo)

	cage, err := svc.Update(context.Background(), 1, 1, &UpdateCageInput{CageSize: &invalid})

	assert.Error(t, err)
	assert.Nil(t, cage)
}

func TestCageService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		usageCount   int64
		usageErr     error
		findByIDErr  error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes cage successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:        "returns not found error when cage does not exist",
			id:          999,
			findByIDErr: apperrors.WrapNotFound("cage", "999"),
			wantErr:     true,
			wantNF:      true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:         "使用中のケージは削除できない",
			id:           2,
			usageCount:   1,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when usage count check fails",
			id:       3,
			usageErr: errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var locked, counted, deleted bool
			inTx := false
			tx := &mockTransactor{withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				inTx = true
				defer func() { inTx = false }()
				return fn(ctx)
			}}
			repo := &mockCageRepository{
				lockByIDForUpdateFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
					assert.True(t, inTx, "LockByIDForUpdate must run inside WithTx")
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					locked = true
					return &model.Cage{ID: id}, nil
				},
				countUsageByCageIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					assert.True(t, inTx, "CountUsage must run inside WithTx")
					assert.True(t, locked, "lock must precede usage count")
					counted = true
					return tt.usageCount, tt.usageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					assert.True(t, inTx, "Delete must run inside WithTx")
					assert.True(t, counted, "usage count must precede soft-delete")
					deleted = true
					return tt.repoErr
				},
			}
			svc := newTestCageServiceWithTx(repo, tx)

			err := svc.Delete(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
					assert.True(t, locked || tt.findByIDErr != nil)
					assert.False(t, deleted, "must not soft-delete on not-found")
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
					assert.False(t, deleted, "must not soft-delete when usage > 0")
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, locked)
				assert.True(t, counted)
				assert.True(t, deleted)
			}
		})
	}
}

func TestCageService_Delete_RequiresTransactor(t *testing.T) {
	svc := NewCageService(&mockCageRepository{}, nil)
	err := svc.Delete(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.False(t, apperrors.IsNotFound(err))
	assert.False(t, apperrors.IsConflict(err))
}

func TestCageService_Reorder(t *testing.T) {
	tests := []struct {
		name        string
		ids         []uint64
		repoErr     error
		wantErr     bool
		wantInvalid bool
		repoCalled  bool
	}{
		{
			name:       "reorders successfully",
			ids:        []uint64{3, 1, 2},
			repoCalled: true,
		},
		{
			name:        "returns error when ids is empty",
			ids:         []uint64{},
			wantErr:     true,
			wantInvalid: true,
			repoCalled:  false,
		},
		{
			name:        "rejects duplicate ids without calling repo",
			ids:         []uint64{1, 2, 1},
			wantErr:     true,
			wantInvalid: true,
			repoCalled:  false,
		},
		{
			name:        "rejects over-limit ids without calling repo",
			ids:         overLimitReorderIDs(),
			wantErr:     true,
			wantInvalid: true,
			repoCalled:  false,
		},
		{
			name:       "propagates repository error",
			ids:        []uint64{1, 2},
			repoErr:    errors.New("db error"),
			wantErr:    true,
			repoCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoCalled := false
			repo := &mockCageRepository{
				reorderFn: func(_ context.Context, _ uint64, _ []uint64) error {
					repoCalled = true
					return tt.repoErr
				},
			}
			svc := newTestCageService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err), "expected InvalidInput, got %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.repoCalled, repoCalled, "repository call expectation")
		})
	}
}

// overLimitReorderIDs builds a unique ID list one past the shared reorder bound.
func overLimitReorderIDs() []uint64 {
	ids := make([]uint64, httpapi.MaxReorderIDs+1)
	for i := range ids {
		ids[i] = uint64(i + 1)
	}
	return ids
}

func TestBuildCageUpdate(t *testing.T) {
	name := "テストケージ"
	cageType := string(model.CageTypeDog)
	cageSize := string(model.CageSizeSmall)
	price := int64(1000)
	isActive := true
	description := "説明文"
	sortOrder := 3

	tests := []struct {
		name  string
		input *UpdateCageInput
		want  map[string]any
	}{
		{
			name:  "全フィールド未指定なら空map",
			input: &UpdateCageInput{},
			want:  map[string]any{},
		},
		{
			name: "全フィールド指定で全キーが含まれる",
			input: &UpdateCageInput{
				Name:        &name,
				CageType:    &cageType,
				CageSize:    &cageSize,
				Price:       &price,
				IsActive:    &isActive,
				Description: &description,
				SortOrder:   &sortOrder,
			},
			want: map[string]any{
				colCageName:        name,
				colCageCageType:    model.CageType(cageType),
				colCageCageSize:    model.CageSize(cageSize),
				colCagePrice:       price,
				colCageIsActive:    isActive,
				colCageDescription: description,
				colCageSortOrder:   sortOrder,
			},
		},
		{
			name:  "Description のみ指定",
			input: &UpdateCageInput{Description: &description},
			want:  map[string]any{colCageDescription: description},
		},
		{
			name:  "SortOrder のみ指定",
			input: &UpdateCageInput{SortOrder: &sortOrder},
			want:  map[string]any{colCageSortOrder: sortOrder},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCageUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
