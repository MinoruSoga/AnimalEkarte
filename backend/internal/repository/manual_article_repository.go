package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ManualArticleRepository は取扱説明書（マニュアル）のデータアクセスインターフェース
//
// P4 例外: マニュアルは医院共通の情報のため clinic_id を持たない。
//
//	clinic_repository.go, account_repository.go と同様の扱い。
type ManualArticleRepository interface {
	FindAll(ctx context.Context) ([]model.ManualArticle, error)
	FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	Upsert(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error)
	Delete(ctx context.Context, category model.ManualCategory, slug string) error
	FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

type manualArticleRepository struct {
	db *gorm.DB
}

// NewManualArticleRepository は ManualArticleRepository を初期化して返す
func NewManualArticleRepository(db *gorm.DB) ManualArticleRepository {
	return &manualArticleRepository{db: db}
}

func (r *manualArticleRepository) FindAll(ctx context.Context) ([]model.ManualArticle, error) {
	articles := make([]model.ManualArticle, 0)
	if err := r.db.WithContext(ctx).
		Order("section ASC, order_value ASC, slug ASC").
		Find(&articles).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", "")
	}
	return articles, nil
}

func (r *manualArticleRepository) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	var article model.ManualArticle
	err := r.db.WithContext(ctx).
		Where("category = ? AND slug = ?", category, slug).
		First(&article).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	return &article, nil
}

// Upsert は category+slug で UNIQUE 制約を利用した upsert を実行する。
// 履歴テーブルへの記録も同一トランザクション内で行う。
func (r *manualArticleRepository) Upsert(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error) {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 既存レコード取得
		var existing model.ManualArticle
		findErr := tx.Where("category = ? AND slug = ?", article.Category, article.Slug).First(&existing).Error

		switch {
		case findErr == nil:
			// 既存あり: UPDATE
			updates := map[string]any{
				"title":               article.Title,
				"order_value":         article.OrderValue,
				"section":             article.Section,
				"body_markdown":       article.BodyMarkdown,
				"updated_by_staff_id": editorStaffID,
			}
			if err := tx.Model(&model.ManualArticle{}).
				Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%d", existing.ID))
			}
			article.ID = existing.ID
			article.CreatedAt = existing.CreatedAt
		case errors.Is(findErr, gorm.ErrRecordNotFound):
			// 既存なし: INSERT
			article.UpdatedByStaffID = editorStaffID
			if err := tx.Create(article).Error; err != nil {
				return apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%s/%s", article.Category, article.Slug))
			}
		default:
			// DB 接続失敗等: 伝播
			return apperrors.FromGORM(findErr, "manual_article", fmt.Sprintf("%s/%s", article.Category, article.Slug))
		}

		// 履歴記録（毎回スナップショット）
		version := model.ManualArticleVersion{
			ArticleID:       article.ID,
			Title:           article.Title,
			OrderValue:      article.OrderValue,
			Section:         article.Section,
			BodyMarkdown:    article.BodyMarkdown,
			EditedByStaffID: editorStaffID,
		}
		if err := tx.Create(&version).Error; err != nil {
			return apperrors.FromGORM(err, "manual_article_version", "")
		}

		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to upsert manual article")
	}

	// 最新レコードを返す
	return r.FindByCategoryAndSlug(ctx, article.Category, article.Slug)
}

func (r *manualArticleRepository) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	result := r.db.WithContext(ctx).
		Where("category = ? AND slug = ?", category, slug).
		Delete(&model.ManualArticle{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	if result.RowsAffected == 0 {
		return apperrors.FromGORM(gorm.ErrRecordNotFound, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	return nil
}

func (r *manualArticleRepository) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	versions := make([]model.ManualArticleVersion, 0)
	if err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("edited_at DESC").
		Find(&versions).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return versions, nil
}
