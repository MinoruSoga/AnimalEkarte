// Package manualarticle owns manual_articles / manual_article_versions data access
// (BE8-4 batch3 — fully independent leaf domain, no clinic_id).
package manualarticle

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MaxVersionsPerArticle is the hard retention / list cap for version snapshots per article.
// Storage prune and FindVersionsByArticleID both fail-closed at this bound (SEC-CS-F07).
const MaxVersionsPerArticle = 50

// Repository は取扱説明書（マニュアル）のデータアクセスインターフェース
//
// P4 例外: マニュアルは医院共通の情報のため clinic_id を持たない。
//
//	clinic_repository.go, account_repository.go と同様の扱い。
type Repository interface {
	FindAll(ctx context.Context) ([]model.ManualArticle, error)
	FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	Upsert(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error)
	Delete(ctx context.Context, category model.ManualCategory, slug string) error
	FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
}

type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]model.ManualArticle, error) {
	articles := make([]model.ManualArticle, 0)
	if err := r.db.WithContext(ctx).
		Order("section ASC, order_value ASC, slug ASC").
		Find(&articles).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", "")
	}
	return articles, nil
}

func (r *repository) FindByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
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
// 履歴テーブルへの記録と最新行の再取得も同一トランザクション内で行う (TRM-01 / X-01)。
func (r *repository) Upsert(ctx context.Context, article *model.ManualArticle, editorStaffID *uint64) (*model.ManualArticle, error) {
	var saved model.ManualArticle
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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

		// SEC-CS-F07: prune oldest snapshots beyond retention in the same TX.
		if err := pruneVersionsBeyondRetention(tx, article.ID); err != nil {
			return err
		}

		// TRM-01: re-load before commit so post-commit read cannot flip success to failure.
		if err := tx.Where("category = ? AND slug = ?", article.Category, article.Slug).
			First(&saved).Error; err != nil {
			return apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%s/%s", article.Category, article.Slug))
		}
		return nil
	}); err != nil {
		return nil, apperrors.Wrap(err, "failed to upsert manual article")
	}

	return &saved, nil
}

// pruneVersionsBeyondRetention keeps the newest MaxVersionsPerArticle rows per article.
// Must run inside the same transaction as version create (fail-closed retention).
func pruneVersionsBeyondRetention(tx *gorm.DB, articleID uint64) error {
	var oldIDs []uint64
	if err := tx.Model(&model.ManualArticleVersion{}).
		Where("article_id = ?", articleID).
		Order("edited_at DESC, id DESC").
		Offset(MaxVersionsPerArticle).
		Pluck("id", &oldIDs).Error; err != nil {
		return apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	if len(oldIDs) == 0 {
		return nil
	}
	if err := tx.Where("id IN ?", oldIDs).Delete(&model.ManualArticleVersion{}).Error; err != nil {
		return apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	result := r.db.WithContext(ctx).
		Where("category = ? AND slug = ?", category, slug).
		Delete(&model.ManualArticle{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	return nil
}

func (r *repository) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	versions := make([]model.ManualArticleVersion, 0)
	if err := r.db.WithContext(ctx).
		Where("article_id = ?", articleID).
		Order("edited_at DESC, id DESC").
		Limit(MaxVersionsPerArticle).
		Find(&versions).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return versions, nil
}
