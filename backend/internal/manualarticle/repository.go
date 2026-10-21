// Package manualarticle owns manual_articles / manual_article_versions /
// manual_article_operations / manual_article_operation_targets data access
// (BE8-4 batch3 — fully independent leaf domain, no clinic_id on articles).
package manualarticle

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MaxVersionsPerArticle is the hard retention / list cap for version snapshots per article.
// Storage prune and FindVersionsByArticleID both fail-closed at this bound (SEC-CS-F07).
const MaxVersionsPerArticle = 50

// PublishedArticle は公開確定済みの内容版を重ねた記事の公開 READ projection。
// 公開版の内容だけを持ち、未公開 current・CAS 版・hidden 情報は含めない。
type PublishedArticle struct {
	ID              uint64
	Category        model.ManualCategory
	Slug            string
	Title           string
	OrderValue      model.ExactNumeric
	Section         string
	BodyMarkdown    string
	EditedByStaffID *uint64
	EditedAt        time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EditView は編集権限者向けの現在状態: 記事行（CAS 版・pointer・hidden）と
// 現在の編集内容版（current_version_id が NULL なら nil）。
type EditView struct {
	Article model.ManualArticle
	Current *model.ManualArticleVersion
}

// Repository は取扱説明書（マニュアル）のデータアクセスインターフェース
//
// P4 例外: マニュアルは医院共通の情報のため clinic_id を持たない。
//
//	clinic_repository.go, account_repository.go と同様の扱い。
//
// 公開 READ（FindAllPublished / FindPublishedByCategoryAndSlug）は
// published_version_id の内容版だけを返し、hidden・未公開は除外する。
type Repository interface {
	FindAllPublished(ctx context.Context) ([]PublishedArticle, error)
	FindPublishedByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error)
	// FindArticleByCategoryAndSlug は記事行だけを返す（公開状態・hidden に関わらない
	// 管理系 read。編集・履歴・削除の対象解決に使う）。
	FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error)
	// FindEditView は記事行と現在内容版を返す。記事不存在は error。
	FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error)
	FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error)
	Delete(ctx context.Context, category model.ManualCategory, slug string) error
	EditTxStore
}

type repository struct {
	db *gorm.DB
}

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

// publishedSelect は公開版 join の固定 SELECT list。
const publishedSelect = `a.id, a.category, a.slug,
	v.title, v.order_value, v.section, v.body_markdown, v.edited_by_staff_id, v.edited_at,
	a.created_at, a.updated_at`

func (r *repository) publishedBase(ctx context.Context) *gorm.DB {
	return persistence.DBOrTx(ctx, r.db).
		Table("manual_articles AS a").
		Select(publishedSelect).
		Joins("JOIN manual_article_versions v ON v.id = a.published_version_id AND v.article_id = a.id").
		Where("a.hidden_at IS NULL")
}

func (r *repository) FindAllPublished(ctx context.Context) ([]PublishedArticle, error) {
	rows := make([]PublishedArticle, 0)
	if err := r.publishedBase(ctx).
		Order("v.section ASC, v.order_value ASC, a.slug ASC").
		Scan(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", "")
	}
	return rows, nil
}

func (r *repository) FindPublishedByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*PublishedArticle, error) {
	var row PublishedArticle
	result := r.publishedBase(ctx).
		Where("a.category = ? AND a.slug = ?", category, slug).
		Take(&row)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	return &row, nil
}

func (r *repository) FindArticleByCategoryAndSlug(ctx context.Context, category model.ManualCategory, slug string) (*model.ManualArticle, error) {
	var article model.ManualArticle
	if err := persistence.DBOrTx(ctx, r.db).
		Where("category = ? AND slug = ?", category, slug).
		First(&article).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%s/%s", category, slug))
	}
	return &article, nil
}

func (r *repository) FindEditView(ctx context.Context, category model.ManualCategory, slug string) (*EditView, error) {
	article, err := r.FindArticleByCategoryAndSlug(ctx, category, slug)
	if err != nil {
		return nil, err
	}
	view := &EditView{Article: *article}
	if article.CurrentVersionID != nil {
		current, verr := r.FindVersionByID(ctx, article.ID, *article.CurrentVersionID)
		if verr != nil {
			return nil, apperrors.Wrap(verr, "failed to load manual article current version")
		}
		view.Current = current
	}
	return view, nil
}

func (r *repository) FindVersionsByArticleID(ctx context.Context, articleID uint64) ([]model.ManualArticleVersion, error) {
	versions := make([]model.ManualArticleVersion, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Where("article_id = ?", articleID).
		Order("edited_at DESC, id DESC").
		Limit(MaxVersionsPerArticle).
		Find(&versions).Error; err != nil {
		return nil, apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return versions, nil
}

func (r *repository) Delete(ctx context.Context, category model.ManualCategory, slug string) error {
	result := persistence.DBOrTx(ctx, r.db).
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

// ─── EditTxStore（ambient tx 内実行） ────────────────────────────────

func (r *repository) InsertOperationTarget(ctx context.Context, clinicID, bindingID, articleID uint64) error {
	if err := persistence.DBOrTx(ctx, r.db).Exec(
		"INSERT INTO manual_article_operation_targets (clinic_id, binding_id, article_id) VALUES (?, ?, ?)",
		clinicID, bindingID, articleID,
	).Error; err != nil {
		return apperrors.FromGORM(err, "manual_article_operation_target", "")
	}
	return nil
}

func (r *repository) ResolveOperationTarget(ctx context.Context, clinicID, bindingID uint64) (uint64, bool, error) {
	var articleID uint64
	result := persistence.DBOrTx(ctx, r.db).Raw(
		`SELECT t.article_id
		   FROM manual_article_operation_targets t
		   JOIN operation_bindings b
		     ON b.id = t.binding_id AND b.clinic_id = t.clinic_id
		  WHERE t.clinic_id = ? AND t.binding_id = ?
		    AND b.owner_kind = ? AND b.intent = ?`,
		clinicID, bindingID, OwnerKind, IntentEdit,
	).Scan(&articleID)
	if result.Error != nil {
		return 0, false, apperrors.FromGORM(result.Error, "manual_article_operation_target", "")
	}
	if result.RowsAffected == 0 {
		return 0, false, nil
	}
	return articleID, true, nil
}

func (r *repository) FindArticleByID(ctx context.Context, id uint64) (*model.ManualArticle, error) {
	var article model.ManualArticle
	err := persistence.DBOrTx(ctx, r.db).Where("id = ?", id).First(&article).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%d", id))
	}
	return &article, nil
}

func (r *repository) FindVersionByID(ctx context.Context, articleID, versionID uint64) (*model.ManualArticleVersion, error) {
	var version model.ManualArticleVersion
	err := persistence.DBOrTx(ctx, r.db).
		Where("id = ? AND article_id = ?", versionID, articleID).
		First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return &version, nil
}

func (r *repository) LockArticleForUpdate(ctx context.Context, id uint64) (*model.ManualArticle, error) {
	var article model.ManualArticle
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", id).
		First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("manual_article", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.FromGORM(err, "manual_article", fmt.Sprintf("%d", id))
	}
	return &article, nil
}

func (r *repository) NextContentRevision(ctx context.Context, articleID uint64) (int64, error) {
	var next int64
	if err := persistence.DBOrTx(ctx, r.db).Raw(
		"SELECT COALESCE(MAX(content_revision), 0) + 1 FROM manual_article_versions WHERE article_id = ?",
		articleID,
	).Scan(&next).Error; err != nil {
		return 0, apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", articleID))
	}
	return next, nil
}

func (r *repository) InsertVersion(ctx context.Context, v *model.ManualArticleVersion) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(v).Error; err != nil {
		return apperrors.FromGORM(err, "manual_article_version", fmt.Sprintf("article=%d", v.ArticleID))
	}
	return nil
}

func (r *repository) UpdateCurrentVersionCAS(ctx context.Context, articleID uint64, expectedVersion int64, expectedCurrentVersionID, newCurrentVersionID uint64) (bool, error) {
	result := persistence.DBOrTx(ctx, r.db).Exec(
		`UPDATE manual_articles
		    SET current_version_id = ?
		  WHERE id = ? AND version = ? AND current_version_id = ?`,
		newCurrentVersionID, articleID, expectedVersion, expectedCurrentVersionID,
	)
	if result.Error != nil {
		return false, apperrors.FromGORM(result.Error, "manual_article", fmt.Sprintf("%d", articleID))
	}
	return result.RowsAffected == 1, nil
}

func (r *repository) InsertOperation(ctx context.Context, op *model.ManualArticleOperation) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(op).Error; err != nil {
		return apperrors.FromGORM(err, "manual_article_operation", fmt.Sprintf("article=%d", op.ArticleID))
	}
	return nil
}
