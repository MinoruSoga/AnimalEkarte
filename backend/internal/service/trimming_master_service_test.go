package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- TrimmingCourse モック ----

type mockTrimmingCourseRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.TrimmingCourse, error)
	createFn   func(ctx context.Context, course *model.TrimmingCourse) error
	updateFn   func(ctx context.Context, course *model.TrimmingCourse) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderErr error
}

func (m *mockTrimmingCourseRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingCourseRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingCourse, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockTrimmingCourseRepository) Create(ctx context.Context, course *model.TrimmingCourse) error {
	return m.createFn(ctx, course)
}

func (m *mockTrimmingCourseRepository) Update(ctx context.Context, course *model.TrimmingCourse) error {
	return m.updateFn(ctx, course)
}

func (m *mockTrimmingCourseRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockTrimmingCourseRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

// ---- TrimmingOption モック ----

type mockTrimmingOptionRepository struct {
	findAllFn  func(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	findByIDFn func(ctx context.Context, id uint64) (*model.TrimmingOption, error)
	createFn   func(ctx context.Context, option *model.TrimmingOption) error
	updateFn   func(ctx context.Context, option *model.TrimmingOption) error
	deleteFn   func(ctx context.Context, id uint64) error
	reorderErr error
}

func (m *mockTrimmingOptionRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockTrimmingOptionRepository) FindByID(ctx context.Context, id uint64) (*model.TrimmingOption, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockTrimmingOptionRepository) Create(ctx context.Context, option *model.TrimmingOption) error {
	return m.createFn(ctx, option)
}

func (m *mockTrimmingOptionRepository) Update(ctx context.Context, option *model.TrimmingOption) error {
	return m.updateFn(ctx, option)
}

func (m *mockTrimmingOptionRepository) Delete(ctx context.Context, id uint64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockTrimmingOptionRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
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
			svc := NewTrimmingCourseService(repo)

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
				findByIDFn: func(_ context.Context, _ uint64) (*model.TrimmingCourse, error) {
					return tt.repoCourse, tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo)

			course, err := svc.GetByID(context.Background(), tt.id)

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
		course  *model.TrimmingCourse
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates course successfully",
			course:  &model.TrimmingCourse{Name: "新規コース", ClinicID: 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when course already exists",
			course:  &model.TrimmingCourse{Name: "重複コース", ClinicID: 1},
			repoErr: apperrors.WrapAlreadyExists("trimming_course", "重複コース"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			course:  &model.TrimmingCourse{Name: "エラーコース", ClinicID: 1},
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
			svc := NewTrimmingCourseService(repo)

			err := svc.Create(context.Background(), tt.course)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingCourseService_Update(t *testing.T) {
	tests := []struct {
		name    string
		course  *model.TrimmingCourse
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates course successfully",
			course:  &model.TrimmingCourse{ID: 1, Name: "更新後コース名"},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when course does not exist",
			course:  &model.TrimmingCourse{ID: 999, Name: "存在しないコース"},
			repoErr: apperrors.WrapNotFound("trimming_course", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			course:  &model.TrimmingCourse{ID: 1, Name: "エラーケース"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				updateFn: func(_ context.Context, _ *model.TrimmingCourse) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo)

			err := svc.Update(context.Background(), tt.course)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingCourseService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes course successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when course does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("trimming_course", "999"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingCourseRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingCourseService(repo)

			err := svc.Delete(context.Background(), tt.id)

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
			svc := NewTrimmingCourseService(repo)

			err := svc.Reorder(context.Background(), 1, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- TrimmingOptionService テスト ----

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
			svc := NewTrimmingOptionService(repo)

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
				findByIDFn: func(_ context.Context, _ uint64) (*model.TrimmingOption, error) {
					return tt.repoOption, tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo)

			option, err := svc.GetByID(context.Background(), tt.id)

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
		option  *model.TrimmingOption
		repoErr error
		wantErr bool
	}{
		{
			name:    "creates option successfully",
			option:  &model.TrimmingOption{Name: "新規オプション", ClinicID: 1},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when option already exists",
			option:  &model.TrimmingOption{Name: "重複オプション", ClinicID: 1},
			repoErr: apperrors.WrapAlreadyExists("trimming_option", "重複オプション"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			option:  &model.TrimmingOption{Name: "エラーオプション", ClinicID: 1},
			repoErr: errors.New("db error"),
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
			svc := NewTrimmingOptionService(repo)

			err := svc.Create(context.Background(), tt.option)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrimmingOptionService_Update(t *testing.T) {
	tests := []struct {
		name    string
		option  *model.TrimmingOption
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates option successfully",
			option:  &model.TrimmingOption{ID: 1, Name: "更新後オプション名"},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns not found error when option does not exist",
			option:  &model.TrimmingOption{ID: 999, Name: "存在しないオプション"},
			repoErr: apperrors.WrapNotFound("trimming_option", "999"),
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			option:  &model.TrimmingOption{ID: 1, Name: "エラーケース"},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				updateFn: func(_ context.Context, _ *model.TrimmingOption) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo)

			err := svc.Update(context.Background(), tt.option)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{reorderErr: tt.repoErr}
			svc := NewTrimmingOptionService(repo)

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
		name    string
		id      uint64
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name:    "deletes option successfully",
			id:      1,
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns not found error when option does not exist",
			id:      999,
			repoErr: apperrors.WrapNotFound("trimming_option", "999"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockTrimmingOptionRepository{
				deleteFn: func(_ context.Context, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewTrimmingOptionService(repo)

			err := svc.Delete(context.Background(), tt.id)

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
