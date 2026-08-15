package manualarticle

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockManualArticleRepository は ManualArticleRepository のテスト用モック実装
type mockManualArticleRepository struct {
	findAllFn                 func(ctx context.Context) ([]model.ManualArticle, error)
	findByCategoryAndSlugFn   func(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	upsertFn                  func(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error)
	deleteFn                  func(ctx context.Context, category model.ManualCategory, slug string) error
	findVersionsByArticleIDFn func(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

func (m *mockManualArticleRepository) FindAll(ctx context.Context) ([]model.ManualArticle, error) {
	return m.findAllFn(ctx)
}

func (m *mockManualArticleRepository) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	return m.findByCategoryAndSlugFn(ctx, category, slug)
}

func (m *mockManualArticleRepository) Upsert(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error) {
	return m.upsertFn(ctx, article, editorStaffID)
}

func (m *mockManualArticleRepository) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	return m.deleteFn(ctx, category, slug)
}

func (m *mockManualArticleRepository) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	if m.findVersionsByArticleIDFn != nil {
		return m.findVersionsByArticleIDFn(ctx, articleID)
	}
	return []model.ManualArticleVersion{}, nil
}

func TestNewManualArticleService(t *testing.T) {
	repo := &mockManualArticleRepository{}
	svc := NewManualArticleService(repo)

	assert.NotNil(t, svc)
}

func TestManualArticleService_FindAll(t *testing.T) {
	tests := []struct {
		name     string
		articles []model.ManualArticle
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns article list",
			articles: []model.ManualArticle{
				{ID: 1, Category: model.ManualCategoryScreens, Slug: "top"},
				{ID: 2, Category: model.ManualCategoryWorkflows, Slug: "checkin"},
			},
			wantLen: 2,
		},
		{
			name:     "returns empty list when none exist",
			articles: []model.ManualArticle{},
			wantLen:  0,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockManualArticleRepository{
				findAllFn: func(_ context.Context) ([]model.ManualArticle, error) {
					return tt.articles, tt.repoErr
				},
			}
			svc := NewManualArticleService(repo)

			articles, err := svc.FindAll(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, articles, tt.wantLen)
			}
		})
	}
}

func TestManualArticleService_FindByCategoryAndSlug(t *testing.T) {
	tests := []struct {
		name     string
		category model.ManualCategory
		slug     string
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "returns article for valid category",
			category: model.ManualCategoryScreens,
			slug:     "top",
		},
		{
			name:     "returns error for invalid category",
			category: model.ManualCategory("invalid"),
			slug:     "top",
			wantErr:  true,
		},
		{
			name:     "propagates repository error",
			category: model.ManualCategoryWorkflows,
			slug:     "missing",
			repoErr:  apperrors.WrapNotFound("manual_article", "missing"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockManualArticleRepository{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.ManualArticle{Category: category, Slug: slug}, nil
				},
			}
			svc := NewManualArticleService(repo)

			article, err := svc.FindByCategoryAndSlug(context.Background(), tt.category, tt.slug)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, article)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, article)
			}
		})
	}
}

func TestManualArticleService_Upsert(t *testing.T) {
	staffID := uint64(7)

	tests := []struct {
		name         string
		input        *UpsertManualArticleInput
		existing     *model.ManualArticle
		existingErr  error
		upsertErr    error
		wantErr      bool
		wantOrderVal float64
	}{
		{
			name: "creates new article with explicit order value",
			input: &UpsertManualArticleInput{
				Category:     model.ManualCategoryScreens,
				Slug:         "top",
				Title:        "トップ画面",
				OrderValue:   1,
				Section:      "基本操作",
				BodyMarkdown: "# トップ",
			},
			wantOrderVal: 1,
		},
		{
			name: "returns error for invalid category",
			input: &UpsertManualArticleInput{
				Category: model.ManualCategory("invalid"),
				Slug:     "top",
				Title:    "タイトル",
				Section:  "セクション",
			},
			wantErr: true,
		},
		{
			name: "returns error when slug is empty",
			input: &UpsertManualArticleInput{
				Category: model.ManualCategoryScreens,
				Slug:     "",
				Title:    "タイトル",
				Section:  "セクション",
			},
			wantErr: true,
		},
		{
			name: "returns error when title is empty",
			input: &UpsertManualArticleInput{
				Category: model.ManualCategoryScreens,
				Slug:     "top",
				Title:    "",
				Section:  "セクション",
			},
			wantErr: true,
		},
		{
			name: "returns error when section is empty",
			input: &UpsertManualArticleInput{
				Category: model.ManualCategoryScreens,
				Slug:     "top",
				Title:    "タイトル",
				Section:  "",
			},
			wantErr: true,
		},
		{
			name: "falls back to existing order value when zero and article exists",
			input: &UpsertManualArticleInput{
				Category:   model.ManualCategoryScreens,
				Slug:       "top",
				Title:      "タイトル",
				OrderValue: 0,
				Section:    "セクション",
			},
			existing:     &model.ManualArticle{OrderValue: 42},
			wantOrderVal: 42,
		},
		{
			name: "falls back to default 9999 when zero and article does not exist",
			input: &UpsertManualArticleInput{
				Category:   model.ManualCategoryScreens,
				Slug:       "new-slug",
				Title:      "タイトル",
				OrderValue: 0,
				Section:    "セクション",
			},
			existingErr:  apperrors.WrapNotFound("manual_article", "new-slug"),
			wantOrderVal: 9999,
		},
		{
			name: "returns error when existing lookup fails with non-notfound error",
			input: &UpsertManualArticleInput{
				Category:   model.ManualCategoryScreens,
				Slug:       "top",
				Title:      "タイトル",
				OrderValue: 0,
				Section:    "セクション",
			},
			existingErr: errors.New("db error"),
			wantErr:     true,
		},
		{
			name: "propagates repository upsert error",
			input: &UpsertManualArticleInput{
				Category:   model.ManualCategoryScreens,
				Slug:       "top",
				Title:      "タイトル",
				OrderValue: 1,
				Section:    "セクション",
			},
			upsertErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedOrder float64
			repo := &mockManualArticleRepository{
				findByCategoryAndSlugFn: func(_ context.Context, _ model.ManualCategory, _ string) (*model.ManualArticle, error) {
					if tt.existingErr != nil {
						return nil, tt.existingErr
					}
					if tt.existing != nil {
						return tt.existing, nil
					}
					return &model.ManualArticle{}, nil
				},
				upsertFn: func(_ context.Context, article *model.ManualArticle, _ *uint64) (*model.ManualArticle, error) {
					capturedOrder = article.OrderValue
					if tt.upsertErr != nil {
						return nil, tt.upsertErr
					}
					return article, nil
				},
			}
			svc := NewManualArticleService(repo)

			saved, err := svc.Upsert(context.Background(), tt.input, &staffID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, saved)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, saved)
				assert.Equal(t, tt.wantOrderVal, capturedOrder)
			}
		})
	}
}

func TestManualArticleService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		category  model.ManualCategory
		slug      string
		findErr   error
		deleteErr error
		wantErr   bool
	}{
		{
			name:     "deletes existing article",
			category: model.ManualCategoryScreens,
			slug:     "top",
		},
		{
			name:     "returns error for invalid category",
			category: model.ManualCategory("invalid"),
			slug:     "top",
			wantErr:  true,
		},
		{
			name:     "returns error when article not found",
			category: model.ManualCategoryScreens,
			slug:     "missing",
			findErr:  apperrors.WrapNotFound("manual_article", "missing"),
			wantErr:  true,
		},
		{
			name:      "propagates repository delete error",
			category:  model.ManualCategoryScreens,
			slug:      "top",
			deleteErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockManualArticleRepository{
				findByCategoryAndSlugFn: func(_ context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
					if tt.findErr != nil {
						return nil, tt.findErr
					}
					return &model.ManualArticle{Category: category, Slug: slug}, nil
				},
				deleteFn: func(_ context.Context, _ model.ManualCategory, _ string) error {
					return tt.deleteErr
				},
			}
			svc := NewManualArticleService(repo)

			err := svc.Delete(context.Background(), tt.category, tt.slug)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManualArticleService_FindVersionsByArticleID(t *testing.T) {
	tests := []struct {
		name     string
		versions []model.ManualArticleVersion
		repoErr  error
		wantLen  int
		wantErr  bool
	}{
		{
			name: "returns versions list",
			versions: []model.ManualArticleVersion{
				{ID: 1, ArticleID: 5, Title: "v1"},
				{ID: 2, ArticleID: 5, Title: "v2"},
			},
			wantLen: 2,
		},
		{
			name:     "returns empty list when no versions exist",
			versions: []model.ManualArticleVersion{},
			wantLen:  0,
		},
		{
			name:    "propagates repository error",
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockManualArticleRepository{
				findVersionsByArticleIDFn: func(_ context.Context, _ uint64) ([]model.ManualArticleVersion, error) {
					return tt.versions, tt.repoErr
				},
			}
			svc := NewManualArticleService(repo)

			versions, err := svc.FindVersionsByArticleID(context.Background(), 5)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, versions, tt.wantLen)
			}
		})
	}
}
