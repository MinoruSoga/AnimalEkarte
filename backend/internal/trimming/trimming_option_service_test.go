package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingOption モック ----

type mockTrimmingOptionRepository struct {
	findAllFn           func(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	findByIDFn          func(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	createFn            func(ctx context.Context, option *model.TrimmingOption) error
	updateFieldsFn      func(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error)
	deleteFn            func(ctx context.Context, clinicID, id uint64) error
	reorderErr          error
	countRecordsByOptFn func(ctx context.Context, clinicID, optionID uint64) (int64, error)
}

func (m *mockTrimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingOptionRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	return m.createFn(ctx, option)
}

func (m *mockTrimmingOptionRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
	return m.updateFieldsFn(ctx, clinicID, id, cmd)
}

func (m *mockTrimmingOptionRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockTrimmingOptionRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockTrimmingOptionRepository) CountUsageByTrimmingOptionID(ctx context.Context, clinicID, optionID uint64) (int64, error) {
	if m.countRecordsByOptFn != nil {
		return m.countRecordsByOptFn(ctx, clinicID, optionID)
	}
	return 0, nil
}

// ---- TrimmingOptionService テスト ----

func TestBuildTrimmingOptionUpdate(t *testing.T) {
	name := "爪切り"
	price := int64(1000)
	isActive := true
	description := "説明文"
	duration := 15
	isCombinable := true
	sortOrder := 2

	tests := []struct {
		name  string
		input *UpdateTrimmingOptionInput
		want  map[string]any
	}{
		{
			name: "all fields set",
			input: &UpdateTrimmingOptionInput{
				Name: &name, Price: &price, IsActive: &isActive, Description: &description,
				Duration: &duration, IsCombinable: &isCombinable, SortOrder: &sortOrder,
			},
			want: map[string]any{
				colTrimmingOptionName:         name,
				colTrimmingOptionPrice:        price,
				colTrimmingOptionIsActive:     isActive,
				colTrimmingOptionDescription:  description,
				colTrimmingOptionDuration:     duration,
				colTrimmingOptionIsCombinable: isCombinable,
				colTrimmingOptionSortOrder:    sortOrder,
			},
		},
		{
			name:  "no fields set returns empty map",
			input: &UpdateTrimmingOptionInput{},
			want:  map[string]any{},
		},
		{
			name:  "only price set",
			input: &UpdateTrimmingOptionInput{Price: &price},
			want:  map[string]any{colTrimmingOptionPrice: price},
		},
		{
			name:  "only description set",
			input: &UpdateTrimmingOptionInput{Description: &description},
			want:  map[string]any{colTrimmingOptionDescription: description},
		},
		{
			name:  "only duration set",
			input: &UpdateTrimmingOptionInput{Duration: &duration},
			want:  map[string]any{colTrimmingOptionDuration: duration},
		},
		{
			name:  "only is_combinable set",
			input: &UpdateTrimmingOptionInput{IsCombinable: &isCombinable},
			want:  map[string]any{colTrimmingOptionIsCombinable: isCombinable},
		},
		{
			name:  "only sort_order set",
			input: &UpdateTrimmingOptionInput{SortOrder: &sortOrder},
			want:  map[string]any{colTrimmingOptionSortOrder: sortOrder},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTrimmingOptionUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTrimmingOptionService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.TrimmingOption
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns option list",
			repoData: []model.TrimmingOption{
				{ID: 1, Name: "爪切り"},
				{ID: 2, Name: "耳掃除"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no options exist",
			repoData: []model.TrimmingOption{},
			repoErr:  nil,
			wantLen:  0,
			wantErr:  false,
		},
		{
			name:     "propagates repository error",
			repoData: nil,
			repoErr:  errors.New("db connection error"),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			options, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, options, tt.wantLen)
			}
		})
	}
}

func TestTrimmingOptionService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoOption   *model.TrimmingOption
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns option when found",
			id:           1,
			repoOption:   &model.TrimmingOption{ID: 1, Name: "爪切り"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when option does not exist",
			id:           999,
			repoOption:   nil,
			repoErr:      apperrors.WrapNotFound("trimming_option", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoOption:   nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingOption, error) {
					return tt.repoOption, tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			option, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, option)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOption, option)
			}
		})
	}
}

func TestTrimmingOptionService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateTrimmingOptionInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates option successfully",
			input:   &CreateTrimmingOptionInput{Name: "新規オプション", IsActive: true, IsCombinable: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates option with duration",
			input: &CreateTrimmingOptionInput{
				Name:         "爪切り",
				IsActive:     true,
				IsCombinable: true,
				Duration:     func() *int { d := 15; return &d }(),
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when option already exists",
			input:   &CreateTrimmingOptionInput{Name: "重複オプション", IsActive: true},
			repoErr: apperrors.WrapAlreadyExists("trimming_option", "重複オプション"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateTrimmingOptionInput{Name: "エラーオプション", IsActive: true},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns validation error when name is blank",
			input:   &CreateTrimmingOptionInput{Name: "", IsActive: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				createFn: func(_ context.Context, _ *model.TrimmingOption) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			result, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.Name, result.Name)
				assert.Equal(t, uint64(1), result.ClinicID)
			}
		})
	}
}

func TestTrimmingOptionService_Update(t *testing.T) {
	optName := "更新後オプション名"
	tests := []struct {
		name    string
		input   *UpdateTrimmingOptionInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates option successfully",
			input:   &UpdateTrimmingOptionInput{Name: &optName},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateTrimmingOptionInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "returns not found error when option does not exist",
			input:   &UpdateTrimmingOptionInput{Name: &optName},
			repoErr: apperrors.WrapNotFound("trimming_option", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &UpdateTrimmingOptionInput{Name: &optName},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns validation error when name is blank",
			input:   &UpdateTrimmingOptionInput{Name: strPtr("")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingOption, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.TrimmingOption{ID: 1, Name: optName}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateTrimmingOptionInput) (*model.TrimmingOption, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.TrimmingOption{ID: 1, Name: optName}, nil
				},
			}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			option, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, option)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, option)
			}
		})
	}
}

func TestTrimmingOptionService_Update_NilInput(t *testing.T) {
	repo := &mockTrimmingOptionRepository{}
	svc := NewTrimmingOptionService(repo, &mockTransactor{})
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestTrimmingOptionService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{2, 3, 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "returns error when ids is empty",
			ids:     []uint64{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{reorderErr: tt.repoErr}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingOptionService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		usageCount   int64
		usageErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:       "deletes option successfully",
			id:         1,
			usageCount: 0,
			repoErr:    nil,
			wantErr:    false,
			wantNF:     false,
		},
		{
			name:         "returns conflict error when option is used in trimming records",
			id:           1,
			usageCount:   2,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:     "returns error when usage count check fails",
			id:       1,
			usageErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:    "returns not found error when option does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("trimming_option", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:       "returns error on repository failure",
			id:         1,
			usageCount: 0,
			repoErr:    errors.New("db error"),
			wantErr:    true,
			wantNF:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				countRecordsByOptFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.usageCount, tt.usageErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo, &mockTransactor{})

			err := svc.Delete(context.Background(), 1, tt.id)

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
