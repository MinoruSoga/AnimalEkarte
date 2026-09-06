package trimming

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse モック ----

type mockTrimmingCourseRepository struct {
	findAllFn              func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	findByIDFn             func(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	createFn               func(ctx context.Context, course *model.TrimmingCourse) error
	updateFieldsFn         func(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error)
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	countUsageByCourseIDFn func(ctx context.Context, clinicID, courseID uint64) (int64, error)
	reorderErr             error
}

// mockMinimalCourseTypeRepo is a minimal mock for TrimmingCourseTypeRepository
type mockMinimalCourseTypeRepo struct{}

func (m *mockMinimalCourseTypeRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourseType, error) {
	return &model.TrimmingCourseType{ID: id, ClinicID: clinicID}, nil
}

func (m *mockMinimalCourseTypeRepo) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourseType, error) {
	return nil, nil
}

func (m *mockMinimalCourseTypeRepo) Create(ctx context.Context, t *model.TrimmingCourseType) (*model.TrimmingCourseType, error) {
	return t, nil
}

func (m *mockMinimalCourseTypeRepo) Update(ctx context.Context, clinicID, id uint64, _ UpdateTrimmingCourseTypeInput) (*model.TrimmingCourseType, error) {
	return nil, nil
}

func (m *mockMinimalCourseTypeRepo) Delete(ctx context.Context, clinicID, id uint64) error {
	return nil
}

func (m *mockMinimalCourseTypeRepo) CountUsageByCourseTypeID(ctx context.Context, clinicID, id uint64) (int64, error) {
	return 0, nil
}

func (m *mockMinimalCourseTypeRepo) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return nil
}

func (m *mockTrimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingCourseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return m.createFn(ctx, course)
}

func (m *mockTrimmingCourseRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, cmd)
	}
	return nil, nil
}

func (m *mockTrimmingCourseRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockTrimmingCourseRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockTrimmingCourseRepository) CountUsageByTrimmingCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error) {
	if m.countUsageByCourseIDFn != nil {
		return m.countUsageByCourseIDFn(ctx, clinicID, courseID)
	}
	return 0, nil
}

// ---- TrimmingCourseService テスト ----

func TestTrimmingCourseService_List(t *testing.T) {
	tests := []struct {
		name     string
		repoData []model.TrimmingCourse
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns course list",
			repoData: []model.TrimmingCourse{
				{ID: 1, Name: "スタンダードコース"},
				{ID: 2, Name: "プレミアムコース"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:     "returns empty list when no courses exist",
			repoData: []model.TrimmingCourse{},
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
			repo := &mockTrimmingCourseRepository{
				findAllFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
					return tt.repoData, tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

			courses, err := svc.List(context.Background(), 1)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, courses, tt.wantLen)
			}
		})
	}
}

func TestTrimmingCourseService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		repoCourse   *model.TrimmingCourse
		repoErr      error
		wantErr      bool
		wantNotFound bool
	}{
		{
			name:         "returns course when found",
			id:           1,
			repoCourse:   &model.TrimmingCourse{ID: 1, Name: "スタンダードコース"},
			repoErr:      nil,
			wantErr:      false,
			wantNotFound: false,
		},
		{
			name:         "returns not found error when course does not exist",
			id:           999,
			repoCourse:   nil,
			repoErr:      apperrors.WrapNotFound("trimming_course", "999"),
			wantErr:      true,
			wantNotFound: true,
		},
		{
			name:         "returns error on repository failure",
			id:           1,
			repoCourse:   nil,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourse, error) {
					return tt.repoCourse, tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

			course, err := svc.GetByID(context.Background(), 1, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					assert.True(t, apperrors.IsNotFound(err))
				}
				assert.Nil(t, course)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoCourse, course)
			}
		})
	}
}

func TestTrimmingCourseService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateTrimmingCourseInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates course successfully",
			input:   &CreateTrimmingCourseInput{Name: "新規コース", IsActive: true},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "creates course with target size",
			input: &CreateTrimmingCourseInput{
				Name:       "小型犬コース",
				TargetSize: "small",
				IsActive:   true,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when course already exists",
			input:   &CreateTrimmingCourseInput{Name: "重複コース", IsActive: true},
			repoErr: apperrors.WrapAlreadyExists("trimming_course", "重複コース"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &CreateTrimmingCourseInput{Name: "エラーコース", IsActive: true},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				createFn: func(_ context.Context, _ *model.TrimmingCourse) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

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

func TestTrimmingCourseService_Update(t *testing.T) {
	name := "更新後コース名"
	tests := []struct {
		name    string
		input   *UpdateTrimmingCourseInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates course successfully",
			input:   &UpdateTrimmingCourseInput{Name: &name},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   &UpdateTrimmingCourseInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "returns not found error when course does not exist",
			input:   &UpdateTrimmingCourseInput{Name: &name},
			repoErr: apperrors.WrapNotFound("trimming_course", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   &UpdateTrimmingCourseInput{Name: &name},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourse, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.TrimmingCourse{ID: 1, Name: name}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateTrimmingCourseInput) (*model.TrimmingCourse, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.TrimmingCourse{ID: 1, Name: name}, nil
				},
			}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

			course, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, course)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, course)
			}
		})
	}
}

func TestTrimmingCourseService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		id           uint64
		countRecords int64
		countErr     error
		deleteErr    error
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name:         "deletes course successfully",
			id:           1,
			countRecords: 0,
			deleteErr:    nil,
			wantErr:      false,
		},
		{
			name:         "returns conflict error when course is in use",
			id:           1,
			countRecords: 1,
			wantErr:      true,
			wantConflict: true,
		},
		{
			name:      "returns not found error when course does not exist",
			id:        999,
			deleteErr: apperrors.WrapNotFound("trimming_course", "999"),
			wantErr:   true,
			wantNF:    true,
		},
		{
			name:      "returns error on repository failure",
			id:        1,
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				countUsageByCourseIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.countRecords, tt.countErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

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

func TestTrimmingCourseService_Reorder(t *testing.T) {
	tests := []struct {
		name    string
		ids     []uint64
		repoErr error
		wantErr bool
	}{
		{
			name:    "reorders successfully",
			ids:     []uint64{3, 1, 2},
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
			repo := &mockTrimmingCourseRepository{reorderErr: tt.repoErr}
			svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingCourseService_Update_NilInput(t *testing.T) {
	repo := &mockTrimmingCourseRepository{}
	svc := NewTrimmingCourseService(repo, &mockTrimmingCourseTypeRepository{}, &mockTransactor{})
	result, err := svc.Update(context.Background(), 1, 1, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- buildTrimmingCourseUpdate 直接テスト ----

func TestBuildTrimmingCourseUpdate(t *testing.T) {
	name := "更新後コース"
	price := int64(5000)
	isActive := true
	description := "更新説明"
	targetSize := "large"
	courseTypeID := uint64(3)
	duration := 90
	sortOrder := 2

	tests := []struct {
		name  string
		input *UpdateTrimmingCourseInput
		want  map[string]any
	}{
		{
			name:  "全フィールドnilの場合は空map",
			input: &UpdateTrimmingCourseInput{},
			want:  map[string]any{},
		},
		{
			name: "全フィールド設定",
			input: &UpdateTrimmingCourseInput{
				Name:         &name,
				Price:        &price,
				IsActive:     &isActive,
				Description:  &description,
				TargetSize:   &targetSize,
				CourseTypeID: &courseTypeID,
				Duration:     &duration,
				SortOrder:    &sortOrder,
			},
			want: map[string]any{
				colTrimmingCourseName:         name,
				colTrimmingCoursePrice:        price,
				colTrimmingCourseIsActive:     isActive,
				colTrimmingCourseDescription:  description,
				colTrimmingCourseTargetSize:   model.TargetSize(targetSize),
				colTrimmingCourseCourseTypeID: courseTypeID,
				colTrimmingCourseDuration:     duration,
				colTrimmingCourseSortOrder:    sortOrder,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTrimmingCourseUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---- Create 追加分岐 ----

func TestTrimmingCourseService_Create_EmptyNameError(t *testing.T) {
	svc := NewTrimmingCourseService(&mockTrimmingCourseRepository{}, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

	result, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{Name: ""})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestTrimmingCourseService_Create_CourseTypeNotFound(t *testing.T) {
	courseTypeID := uint64(5)
	courseTypeRepo := &mockTrimmingCourseTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.TrimmingCourseType, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewTrimmingCourseService(&mockTrimmingCourseRepository{}, courseTypeRepo, &mockTransactor{})

	result, err := svc.Create(context.Background(), 1, &CreateTrimmingCourseInput{
		Name:         "新規コース",
		CourseTypeID: &courseTypeID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestTrimmingCourseService_Create_RejectsCrossClinicCourseTypeBeforeWrite(t *testing.T) {
	const clinicID = uint64(1)
	courseTypeID := uint64(5)
	createCalled := false
	courseRepo := &mockTrimmingCourseRepository{
		createFn: func(context.Context, *model.TrimmingCourse) error {
			createCalled = true
			return nil
		},
	}
	courseTypeRepo := &mockTrimmingCourseTypeRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotID uint64) (*model.TrimmingCourseType, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, courseTypeID, gotID)
			return nil, apperrors.WrapNotFound("trimming course type", "5")
		},
	}
	svc := NewTrimmingCourseService(courseRepo, courseTypeRepo, &mockTransactor{})

	result, err := svc.Create(context.Background(), clinicID, &CreateTrimmingCourseInput{
		Name:         "新規コース",
		CourseTypeID: &courseTypeID,
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.False(t, createCalled)
}

// ---- Update 追加分岐 ----

func TestTrimmingCourseService_Update_InvalidName(t *testing.T) {
	invalidName := "無効\x00文字"
	repo := &mockTrimmingCourseRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id}, nil
		},
	}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

	result, err := svc.Update(context.Background(), 1, 1, &UpdateTrimmingCourseInput{Name: &invalidName})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, apperrors.IsInvalidInput(err))
}

// ---- Delete 追加分岐 ----

func TestTrimmingCourseService_Delete_CountUsageError(t *testing.T) {
	repo := &mockTrimmingCourseRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.TrimmingCourse, error) {
			return &model.TrimmingCourse{ID: id}, nil
		},
		countUsageByCourseIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, errors.New("db error")
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	svc := NewTrimmingCourseService(repo, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
}

// ---- Reorder 追加分岐 ----

func TestTrimmingCourseService_Reorder_EmptyIDs(t *testing.T) {
	svc := NewTrimmingCourseService(&mockTrimmingCourseRepository{}, &mockMinimalCourseTypeRepo{}, &mockTransactor{})

	err := svc.Reorder(context.Background(), 1, []uint64{})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}
