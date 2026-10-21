package manualarticle

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ManualArticleService は取扱説明書のビジネスロジック
//
// マニュアルは医院共通の情報のため clinic_id を持たない。
// 編集権限の判定は handler 層で実施。
// 公開 READ（FindAll / FindByCategoryAndSlug）は公開確定版のみを返す。
type ManualArticleService interface {
	FindAll(ctx context.Context) ([]PublishedArticle, error)
	FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error)
	FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error)
	FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
	Delete(ctx context.Context, category model.ManualCategory, slug string) error
}

type manualArticleService struct {
	repo Repository
}

// NewManualArticleService は ManualArticleService を初期化
func NewManualArticleService(repo Repository) ManualArticleService {
	return &manualArticleService{repo: repo}
}

func (s *manualArticleService) FindAll(ctx context.Context) ([]PublishedArticle, error) {
	articles, err := s.repo.FindAllPublished(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find manual articles")
	}
	return articles, nil
}

func (s *manualArticleService) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error) {
	if !model.IsValidManualCategory(string(category)) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	article, err := s.repo.FindPublishedByCategoryAndSlug(ctx, category, slug)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find manual article")
	}
	return article, nil
}

func (s *manualArticleService) FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	if !model.IsValidManualCategory(string(category)) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	article, err := s.repo.FindArticleByCategoryAndSlug(ctx, category, slug)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find manual article")
	}
	return article, nil
}

func (s *manualArticleService) FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error) {
	if !model.IsValidManualCategory(string(category)) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	if slug == "" {
		return nil, apperrors.WrapInvalidInput("slug is required")
	}
	view, err := s.repo.FindEditView(ctx, category, slug)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to load manual article edit view")
	}
	return view, nil
}

func (s *manualArticleService) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	if !model.IsValidManualCategory(string(category)) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	// P1: Delete 前の存在確認
	if _, err := s.repo.FindArticleByCategoryAndSlug(ctx, category, slug); err != nil {
		return apperrors.Wrap(err, "failed to find manual article")
	}
	if err := s.repo.Delete(ctx, category, slug); err != nil {
		return apperrors.Wrap(err, "failed to delete manual article")
	}
	return nil
}

func (s *manualArticleService) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	versions, err := s.repo.FindVersionsByArticleID(ctx, articleID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find manual article versions")
	}
	return versions, nil
}
