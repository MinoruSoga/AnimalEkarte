package manualarticle

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockManualArticleRepository は Repository のテスト用モック実装
type mockManualArticleRepository struct {
	findAllPublishedFn  func(ctx context.Context) ([]PublishedArticle, error)
	findPublishedFn     func(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error)
	findArticleFn       func(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	findEditViewFn      func(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error)
	deleteFn            func(ctx context.Context, category model.ManualCategory, slug string) error
	findVersionsByIDFn  func(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

func (m *mockManualArticleRepository) FindAllPublished(ctx context.Context) ([]PublishedArticle, error) {
	return m.findAllPublishedFn(ctx)
}

func (m *mockManualArticleRepository) FindPublishedByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error) {
	return m.findPublishedFn(ctx, category, slug)
}

func (m *mockManualArticleRepository) FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	return m.findArticleFn(ctx, category, slug)
}

func (m *mockManualArticleRepository) FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error) {
	return m.findEditViewFn(ctx, category, slug)
}

func (m *mockManualArticleRepository) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	if m.findVersionsByIDFn != nil {
		return m.findVersionsByIDFn(ctx, articleID)
	}
	return nil, nil
}

func (m *mockManualArticleRepository) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	return m.deleteFn(ctx, category, slug)
}

// EditTxStore は mock 対象外 — 空の stub 実装で Repository を満たす。
func (m *mockManualArticleRepository) InsertOperationTarget(context.Context, uint64, uint64, uint64) error {
	return nil
}

func (m *mockManualArticleRepository) ResolveOperationTarget(context.Context, uint64, uint64) (uint64, bool, error) {
	return 0, false, nil
}

func (m *mockManualArticleRepository) FindArticleByID(context.Context, uint64) (*model.ManualArticle, error) {
	return nil, nil
}

func (m *mockManualArticleRepository) FindVersionByID(context.Context, uint64, uint64) (*model.ManualArticleVersion, error) {
	return nil, nil
}

func (m *mockManualArticleRepository) LockArticleForUpdate(context.Context, uint64) (*model.ManualArticle, error) {
	return nil, nil
}

func (m *mockManualArticleRepository) NextContentRevision(context.Context, uint64) (int64, error) {
	return 0, nil
}

func (m *mockManualArticleRepository) InsertVersion(context.Context, *model.ManualArticleVersion) error {
	return nil
}

func (m *mockManualArticleRepository) UpdateCurrentVersionCAS(context.Context, uint64, int64, uint64, uint64) (bool, error) {
	return false, nil
}

func (m *mockManualArticleRepository) InsertOperation(context.Context, *model.ManualArticleOperation) error {
	return nil
}

var _ Repository = (*mockManualArticleRepository)(nil)

func TestManualArticleService_FindAll(t *testing.T) {
	want := []PublishedArticle{{ID: 1}}
	repo := &mockManualArticleRepository{
		findAllPublishedFn: func(context.Context) ([]PublishedArticle, error) { return want, nil },
	}
	svc := NewManualArticleService(repo)
	got, err := svc.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestManualArticleService_FindByCategoryAndSlug_InvalidCategory(t *testing.T) {
	repo := &mockManualArticleRepository{
		findPublishedFn: func(context.Context, model.ManualCategory, string) (*PublishedArticle, error) {
			t.Fatal("repository must not be called for invalid category")
			return nil, nil
		},
	}
	svc := NewManualArticleService(repo)
	_, err := svc.FindByCategoryAndSlug(context.Background(), "invalid", "slug")
	assert.Error(t, err)
}

func TestManualArticleService_FindByCategoryAndSlug_PublishedOnly(t *testing.T) {
	// 公開 READ は published_version_id の版のみ — repository の
	// FindPublishedByCategoryAndSlug に委譲される。
	repo := &mockManualArticleRepository{
		findPublishedFn: func(context.Context, model.ManualCategory, string) (*PublishedArticle, error) {
			return nil, apperrors.WrapNotFound("manual_article", "screens/x")
		},
	}
	svc := NewManualArticleService(repo)
	_, err := svc.FindByCategoryAndSlug(context.Background(), model.ManualCategoryScreens, "x")
	assert.Error(t, err)
}

func TestManualArticleService_FindEditView_InvalidCategory(t *testing.T) {
	repo := &mockManualArticleRepository{}
	svc := NewManualArticleService(repo)
	_, err := svc.FindEditView(context.Background(), "invalid", "s")
	assert.Error(t, err)
	_, err = svc.FindEditView(context.Background(), model.ManualCategoryScreens, "")
	assert.Error(t, err)
}

func TestManualArticleService_Delete_NotFoundBeforeDelete(t *testing.T) {
	repo := &mockManualArticleRepository{
		findArticleFn: func(context.Context, model.ManualCategory, string) (*model.ManualArticle, error) {
			return nil, apperrors.WrapNotFound("manual_article", "screens/x")
		},
		deleteFn: func(context.Context, model.ManualCategory, string) error {
			t.Fatal("Delete must not run when the article is missing")
			return nil
		},
	}
	svc := NewManualArticleService(repo)
	err := svc.Delete(context.Background(), model.ManualCategoryScreens, "x")
	assert.Error(t, err)
}

func TestManualArticleService_Delete_OK(t *testing.T) {
	called := false
	svc := NewManualArticleService(&mockManualArticleRepository{
		findArticleFn: func(context.Context, model.ManualCategory, string) (*model.ManualArticle, error) {
			return &model.ManualArticle{ID: 3}, nil
		},
		deleteFn: func(context.Context, model.ManualCategory, string) error {
			called = true
			return nil
		},
	})
	assert.NoError(t, svc.Delete(context.Background(), model.ManualCategoryScreens, "a"))
	assert.True(t, called)
}

func TestManualArticleService_FindVersionsByArticleID_PropagatesError(t *testing.T) {
	repo := &mockManualArticleRepository{
		findVersionsByIDFn: func(context.Context, uint64) ([]model.ManualArticleVersion, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewManualArticleService(repo)
	_, err := svc.FindVersionsByArticleID(context.Background(), 1)
	assert.Error(t, err)
}
