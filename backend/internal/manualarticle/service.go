package manualarticle

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// UpsertManualArticleInput はマニュアル upsert の入力
type UpsertManualArticleInput struct {
	Category     model.ManualCategory
	Slug         string
	Title        string
	OrderValue   float64
	Section      string
	BodyMarkdown string
}

// ManualArticleService は取扱説明書のビジネスロジック
//
// マニュアルは医院共通の情報のため clinic_id を持たない。
// 編集権限の判定は handler 層で実施。
type ManualArticleService interface {
	FindAll(ctx context.Context) ([]model.ManualArticle, error)
	FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	Upsert(ctx context.Context, input *UpsertManualArticleInput, editorStaffID *uint64) (*model.ManualArticle, error)
	Delete(ctx context.Context, category model.ManualCategory, slug string) error
	FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

type manualArticleService struct {
	repo Repository
}

// NewManualArticleService は ManualArticleService を初期化
func NewManualArticleService(repo Repository) ManualArticleService {
	return &manualArticleService{repo: repo}
}

func (s *manualArticleService) FindAll(ctx context.Context) ([]model.ManualArticle, error) {
	articles, err := s.repo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find manual articles", "error", err)
		return nil, apperrors.Wrap(err, "failed to find manual articles")
	}
	return articles, nil
}

func (s *manualArticleService) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	if !model.IsValidManualCategory(string(category)) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	article, err := s.repo.FindByCategoryAndSlug(ctx, category, slug)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find manual article")
	}
	return article, nil
}

func (s *manualArticleService) Upsert(ctx context.Context, input *UpsertManualArticleInput, editorStaffID *uint64) (*model.ManualArticle, error) {
	if !model.IsValidManualCategory(string(input.Category)) {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", input.Category))
	}
	if input.Slug == "" {
		return nil, apperrors.WrapInvalidInput("slug is required")
	}
	if input.Title == "" {
		return nil, apperrors.WrapInvalidInput("title is required")
	}
	if input.Section == "" {
		return nil, apperrors.WrapInvalidInput("section is required")
	}

	// order_value 未指定（0）の場合は既存値 or 既定値（9999）を維持
	// 既存取得失敗時は NotFound（= 新規作成）なら 9999、それ以外のエラーは伝播
	order := input.OrderValue
	if order == 0 {
		existing, err := s.repo.FindByCategoryAndSlug(ctx, input.Category, input.Slug)
		switch {
		case err == nil:
			order = existing.OrderValue
		case apperrors.IsNotFound(err):
			order = 9999
		default:
			slog.ErrorContext(ctx, "failed to fetch existing manual article for order fallback", "error", err, "category", input.Category, "slug", input.Slug)
			return nil, apperrors.Wrap(err, "failed to fetch existing manual article")
		}
	}

	article := &model.ManualArticle{
		Category:     input.Category,
		Slug:         input.Slug,
		Title:        input.Title,
		OrderValue:   order,
		Section:      input.Section,
		BodyMarkdown: input.BodyMarkdown,
	}

	saved, err := s.repo.Upsert(ctx, article, editorStaffID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to upsert manual article", "error", err, "category", input.Category, "slug", input.Slug)
		return nil, apperrors.Wrap(err, "failed to upsert manual article")
	}
	return saved, nil
}

func (s *manualArticleService) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	if !model.IsValidManualCategory(string(category)) {
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid manual category: %s", category))
	}
	// P1: Delete 前の存在確認
	if _, err := s.repo.FindByCategoryAndSlug(ctx, category, slug); err != nil {
		return apperrors.Wrap(err, "failed to find manual article")
	}
	if err := s.repo.Delete(ctx, category, slug); err != nil {
		slog.ErrorContext(ctx, "failed to delete manual article", "error", err, "category", category, "slug", slug)
		return apperrors.Wrap(err, "failed to delete manual article")
	}
	return nil
}

func (s *manualArticleService) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	versions, err := s.repo.FindVersionsByArticleID(ctx, articleID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find manual article versions", "error", err, "article_id", articleID)
		return nil, apperrors.Wrap(err, "failed to find manual article versions")
	}
	return versions, nil
}
