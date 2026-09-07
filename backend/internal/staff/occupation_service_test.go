package staff

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Tests ----

func TestOccupationService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.Occupation
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns occupations list",
			repoData: []model.Occupation{
				{ID: 1, ClinicID: 1, Name: "獣医師", SortOrder: 1, IsActive: true},
				{ID: 2, ClinicID: 1, Name: "看護師", SortOrder: 2, IsActive: true},
				{ID: 3, ClinicID: 1, Name: "受付", SortOrder: 3, IsActive: true},
			},
			repoErr: nil,
			wantLen: 3,
			wantErr: false,
		},
		{
			name:     "returns empty list when no occupations exist",
			repoData: []model.Occupation{},
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
			repo := &mockOccupationRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.Occupation, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occupations, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, occupations, tt.wantLen)
			}
		})
	}
}

func TestOccupationService_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		id             uint64
		repoOccupation *model.Occupation
		repoErr        error
		wantErr        bool
		wantNotFound   bool
	}{
		{
			name: "returns occupation when found",
			id:   1,
			repoOccupation: &model.Occupation{
				ID:          1,
				ClinicID:    1,
				Name:        "獣医師",
				Description: "診療を行う医師",
				SortOrder:   1,
				IsActive:    true,
			},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:           "returns not found error when occupation does not exist",
			id:             999,
			repoOccupation: nil,
			repoErr:        apperrors.WrapNotFound("occupation", "999"),
			wantErr:        true,
			wantNotFound:   true,
		},
		{
			name:           "returns error on repository failure",
			id:             1,
			repoOccupation: nil,
			repoErr:        errors.New("db error"),
			wantErr:        true,
			wantNotFound:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
					return tt.repoOccupation, tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occupation, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, occupation)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOccupation, occupation)
			}
		})
	}
}

func TestOccupationService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateOccupationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates occupation successfully",
			input: &CreateOccupationInput{
				Name:        "新規職種",
				Description: "新しい職種の説明",
				SortOrder:   4,
				IsActive:    true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns error when occupation already exists",
			input: &CreateOccupationInput{
				Name: "獣医師",
			},
			repoErr: apperrors.WrapAlreadyExists("occupation", "獣医師"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: &CreateOccupationInput{
				Name: "エラー職種",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				createFn: func(_ context.Context, _ *model.Occupation) error {
					return tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

			occ, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, occ)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, occ)
			}
		})
	}
}

func TestOccupationService_Update(t *testing.T) {
	updateName := "更新後職種名"
	updateSortOrder := 2
	anotherName := "更新職種"
	errorName := "エラー職種"

	tests := []struct {
		name           string
		clinicID       uint64
		id             uint64
		input          *UpdateOccupationInput
		repoOccupation *model.Occupation
		repoErr        error
		wantErr        bool
	}{
		{
			name:     "updates occupation successfully",
			clinicID: 1,
			id:       1,
			input: &UpdateOccupationInput{
				Name:      &updateName,
				SortOrder: &updateSortOrder,
			},
			repoOccupation: &model.Occupation{
				ID:        1,
				ClinicID:  1,
				Name:      "更新後職種名",
				SortOrder: 2,
				IsActive:  true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when no fields provided",
			clinicID: 1,
			id:       1,
			input:    &UpdateOccupationInput{},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns not found error when occupation does not exist",
			clinicID: 1,
			id:       999,
			input: &UpdateOccupationInput{
				Name: &anotherName,
			},
			repoErr: apperrors.WrapNotFound("occupation", "999"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: &UpdateOccupationInput{
				Name: &errorName,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateOccupationInput) (*model.Occupation, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return tt.repoOccupation, nil
				},
			}
			svc := NewOccupationService(repo)

			occupation, err := svc.Update(context.Background(), tt.clinicID, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoOccupation, occupation)
			}
		})
	}
}

func TestOccupationService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		staffCount   int64
		countErr     error
		repoErr      error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:    "deletes occupation successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when occupation does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("occupation", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name:    "returns error on repository failure",
			id:      1,
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
		{
			name:         "スタッフが所属している職種は削除できない",
			id:           2,
			staffCount:   5,
			wantErr:      true,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{
				countUsageByOccupationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.staffCount, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewOccupationService(repo)

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

// ---- buildOccupationUpdate 直接テスト ----

func TestBuildOccupationUpdate(t *testing.T) {
	name := "更新名"
	description := "更新説明"
	sortOrder := 5
	isActive := true

	tests := []struct {
		name  string
		input *UpdateOccupationInput
		want  map[string]any
	}{
		{
			name:  "全フィールドnilの場合は空map",
			input: &UpdateOccupationInput{},
			want:  map[string]any{},
		},
		{
			name:  "Nameのみ設定",
			input: &UpdateOccupationInput{Name: &name},
			want:  map[string]any{colOccupationName: name},
		},
		{
			name:  "Descriptionのみ設定",
			input: &UpdateOccupationInput{Description: &description},
			want:  map[string]any{colOccupationDescription: description},
		},
		{
			name:  "SortOrderのみ設定",
			input: &UpdateOccupationInput{SortOrder: &sortOrder},
			want:  map[string]any{colOccupationSortOrder: sortOrder},
		},
		{
			name:  "IsActiveのみ設定",
			input: &UpdateOccupationInput{IsActive: &isActive},
			want:  map[string]any{colOccupationIsActive: isActive},
		},
		{
			name: "全フィールド設定",
			input: &UpdateOccupationInput{
				Name:        &name,
				Description: &description,
				SortOrder:   &sortOrder,
				IsActive:    &isActive,
			},
			want: map[string]any{
				colOccupationName:        name,
				colOccupationDescription: description,
				colOccupationSortOrder:   sortOrder,
				colOccupationIsActive:    isActive,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildOccupationUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- Create バリデーションエラー ----

func TestOccupationService_Create_ValidationError(t *testing.T) {
	repo := &mockOccupationRepository{
		createFn: func(_ context.Context, _ *model.Occupation) error {
			t.Fatal("repo.Create should not be called when validation fails")
			return nil
		},
	}
	svc := NewOccupationService(repo)

	occ, err := svc.Create(context.Background(), 1, &CreateOccupationInput{Name: ""})

	assert.Error(t, err)
	assert.Nil(t, occ)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// ---- Update 追加分岐 ----

func TestOccupationService_Update_NilInput(t *testing.T) {
	repo := &mockOccupationRepository{}
	svc := NewOccupationService(repo)

	occ, err := svc.Update(context.Background(), 1, 1, nil)

	assert.Error(t, err)
	assert.Nil(t, occ)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestOccupationService_Update_FindByIDError(t *testing.T) {
	name := "更新名"
	repo := &mockOccupationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
			return nil, apperrors.WrapNotFound("occupation", "1")
		},
	}
	svc := NewOccupationService(repo)

	occ, err := svc.Update(context.Background(), 1, 1, &UpdateOccupationInput{Name: &name})

	assert.Error(t, err)
	assert.Nil(t, occ)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestOccupationService_Update_InvalidName(t *testing.T) {
	invalidName := "無効\x00文字"
	repo := &mockOccupationRepository{}
	svc := NewOccupationService(repo)

	occ, err := svc.Update(context.Background(), 1, 1, &UpdateOccupationInput{Name: &invalidName})

	assert.Error(t, err)
	assert.Nil(t, occ)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// ---- Delete 追加分岐 ----

func TestOccupationService_Delete_CountUsageError(t *testing.T) {
	repo := &mockOccupationRepository{
		countUsageByOccupationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	svc := NewOccupationService(repo)

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}

func TestOccupationService_Delete_LocksAndChecksUsageInsideOneTransaction(t *testing.T) {
	events := make([]string, 0, 3)
	repo := &mockOccupationRepository{
		withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
			events = append(events, "tx")
			return fn(ctx)
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Occupation, error) {
			t.Fatal("delete must not perform an unlocked existence check")
			return nil, nil
		},
		lockForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Occupation, error) {
			events = append(events, "lock")
			return &model.Occupation{ID: id, ClinicID: clinicID}, nil
		},
		countUsageByOccupationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
			events = append(events, "count")
			return 0, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			events = append(events, "delete")
			return nil
		},
	}
	svc := NewOccupationService(repo)

	err := svc.Delete(context.Background(), 10, 7)

	require.NoError(t, err)
	assert.Equal(t, []string{"tx", "lock", "count", "delete"}, events)
	assert.Equal(t, 1, repo.withTxCalls)
}

// ---- Reorder 追加分岐 ----

func TestOccupationService_Reorder_EmptyIDs(t *testing.T) {
	repo := &mockOccupationRepository{}
	svc := NewOccupationService(repo)

	err := svc.Reorder(context.Background(), 1, []uint64{})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestOccupationService_Reorder(t *testing.T) {
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
			name:    "reorders single item",
			ids:     []uint64{1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "propagates repository error",
			ids:     []uint64{1, 2},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockOccupationRepository{reorderErr: tt.repoErr}
			svc := NewOccupationService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
