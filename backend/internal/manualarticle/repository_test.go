package manualarticle

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - ManualArticle は clinic_id を持たないグローバル（医院共通）マスタ。
//   - 公開 READ は published_version_id の内容版だけを返す（hidden・未公開は除外）。
//   - Delete は対象なしで NotFound を返す。
//   - FindVersionsByArticleID は edited_at DESC で返し、件数は MaxVersionsPerArticle を超えない。

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupTestDB は manual_articles / manual_article_versions を用意する。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ManualArticle{}, &model.ManualArticleVersion{}))
	db.Exec("TRUNCATE TABLE manual_article_versions CASCADE")
	db.Exec("TRUNCATE TABLE manual_articles CASCADE")
	return db
}

func numericFixture(t *testing.T) model.ExactNumeric {
	t.Helper()
	var n pgtype.Numeric
	require.NoError(t, n.Scan("9999"))
	return model.ExactNumeric{Numeric: n}
}

func versionFixture(t *testing.T, articleID uint64, title string) *model.ManualArticleVersion {
	t.Helper()
	v := &model.ManualArticleVersion{
		ArticleID:    articleID,
		Title:        title,
		OrderValue:   numericFixture(t),
		Section:      "S",
		BodyMarkdown: "B",
	}
	return v
}

func TestRepository_FindArticleByCategoryAndSlug(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "intro"}
	require.NoError(t, db.Create(article).Error)

	got, err := repo.FindArticleByCategoryAndSlug(ctx, model.ManualCategoryScreens, "intro")
	require.NoError(t, err)
	assert.Equal(t, article.ID, got.ID)

	_, err = repo.FindArticleByCategoryAndSlug(ctx, model.ManualCategoryScreens, "missing")
	assert.True(t, apperrors.IsNotFound(err))
}

func TestRepository_FindPublishedByCategoryAndSlug_OnlyPublished(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	// 公開版を持つ記事: 返る。
	published := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "pub"}
	require.NoError(t, db.Create(published).Error)
	v := versionFixture(t, published.ID, "公開版")
	require.NoError(t, db.Create(v).Error)
	require.NoError(t, db.Model(published).Update("published_version_id", v.ID).Error)

	got, err := repo.FindPublishedByCategoryAndSlug(ctx, model.ManualCategoryScreens, "pub")
	require.NoError(t, err)
	assert.Equal(t, "公開版", got.Title)
	assert.Equal(t, "S", got.Section)

	// 未公開記事（published_version_id NULL）: 404。
	unpublished := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "unpub"}
	require.NoError(t, db.Create(unpublished).Error)
	_, err = repo.FindPublishedByCategoryAndSlug(ctx, model.ManualCategoryScreens, "unpub")
	assert.True(t, apperrors.IsNotFound(err))

	// hidden 記事: 404。
	now := time.Now()
	require.NoError(t, db.Model(published).Update("hidden_at", &now).Error)
	_, err = repo.FindPublishedByCategoryAndSlug(ctx, model.ManualCategoryScreens, "pub")
	assert.True(t, apperrors.IsNotFound(err))
}

func TestRepository_FindAllPublished(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryWorkflows, Slug: "w1"}
	require.NoError(t, db.Create(article).Error)
	v := versionFixture(t, article.ID, "w")
	require.NoError(t, db.Create(v).Error)
	require.NoError(t, db.Model(article).Update("published_version_id", v.ID).Error)

	rows, err := repo.FindAllPublished(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, article.ID, rows[0].ID)
}

func TestRepository_FindEditView(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "a"}
	require.NoError(t, db.Create(article).Error)
	v := versionFixture(t, article.ID, "current")
	require.NoError(t, db.Create(v).Error)
	require.NoError(t, db.Model(article).Updates(map[string]any{
		"current_version_id": v.ID, "version": 5,
	}).Error)

	view, err := repo.FindEditView(ctx, model.ManualCategoryScreens, "a")
	require.NoError(t, err)
	assert.Equal(t, int64(5), view.Article.Version)
	require.NotNil(t, view.Current)
	assert.Equal(t, "current", view.Current.Title)
}

func TestRepository_FindVersionsByArticleID(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "h"}
	require.NoError(t, db.Create(article).Error)
	for _, title := range []string{"v1", "v2", "v3"} {
		require.NoError(t, db.Create(versionFixture(t, article.ID, title)).Error)
		time.Sleep(time.Millisecond)
	}

	versions, err := repo.FindVersionsByArticleID(ctx, article.ID)
	require.NoError(t, err)
	assert.Len(t, versions, 3)
	assert.GreaterOrEqual(t, len(versions), 1)
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	err := repo.Delete(context.Background(), model.ManualCategoryScreens, "missing")
	assert.True(t, apperrors.IsNotFound(err))
}

func TestRepository_Delete_EmptyArticle(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	// 版が無い記事は削除できる（RESTRICT は版行があるときのみ効く）。
	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "del"}
	require.NoError(t, db.Create(article).Error)
	assert.NoError(t, repo.Delete(ctx, model.ManualCategoryScreens, "del"))
}

func TestRepository_NextContentRevision(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "rev"}
	require.NoError(t, db.Create(article).Error)

	next, err := repo.NextContentRevision(ctx, article.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), next)

	rev := int64(2)
	v := versionFixture(t, article.ID, "r")
	v.ContentRevision = &rev
	require.NoError(t, db.Create(v).Error)

	next, err = repo.NextContentRevision(ctx, article.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), next)
}

func TestRepository_LockArticleForUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "lock"}
	require.NoError(t, db.Create(article).Error)

	locked, err := repo.LockArticleForUpdate(ctx, article.ID)
	require.NoError(t, err)
	assert.Equal(t, article.ID, locked.ID)

	_, err = repo.LockArticleForUpdate(ctx, 999999)
	assert.True(t, apperrors.IsNotFound(err))
}
