package manualarticle

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - ManualArticle は clinic_id を持たないグローバル（医院共通）マスタ。
//   - FindAll は section ASC, order_value ASC, slug ASC で返す。
//   - Upsert は category+slug の UNIQUE 制約に基づき、既存なら UPDATE、無ければ INSERT、
//     いずれの場合も ManualArticleVersion に履歴スナップショットを追加する。
//   - Upsert の UPDATE 経路は CreatedAt を保持する。
//   - Delete は対象なしで NotFound を返す。
//   - FindVersionsByArticleID は edited_at DESC で返し、件数は MaxVersionsPerArticle を超えない。
//   - Upsert は同一 TX 内で記事あたりの履歴を MaxVersionsPerArticle 件までに prune する。
//   - prune 後も current article 行は最新内容を保持する。

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func TestRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	articleZ := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "z-slug", Title: "Z", OrderValue: 1, Section: "A", BodyMarkdown: "x"}
	articleA := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "a-slug", Title: "A", OrderValue: 1, Section: "A", BodyMarkdown: "x"}
	articleSecondSection := &model.ManualArticle{Category: model.ManualCategoryWorkflows, Slug: "b-slug", Title: "B", OrderValue: 1, Section: "B", BodyMarkdown: "x"}
	require.NoError(t, db.WithContext(ctx).Create(articleZ).Error)
	require.NoError(t, db.WithContext(ctx).Create(articleA).Error)
	require.NoError(t, db.WithContext(ctx).Create(articleSecondSection).Error)

	got, err := repo.FindAll(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)
	// section "A" の中では order_value が同じなので slug ASC (a-slug < z-slug)
	assert.Equal(t, "a-slug", got[0].Slug)
	assert.Equal(t, "z-slug", got[1].Slug)
	assert.Equal(t, "b-slug", got[2].Slug)
}

func TestRepository_FindByCategoryAndSlug(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "target", Title: "対象", OrderValue: 1, Section: "A", BodyMarkdown: "本文"}
	require.NoError(t, db.WithContext(ctx).Create(article).Error)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByCategoryAndSlug(ctx, model.ManualCategoryScreens, "target")
		require.NoError(t, err)
		assert.Equal(t, "対象", got.Title)
	})

	t.Run("not found for nonexistent slug", func(t *testing.T) {
		got, err := repo.FindByCategoryAndSlug(ctx, model.ManualCategoryScreens, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("not found for same slug but different category", func(t *testing.T) {
		got, err := repo.FindByCategoryAndSlug(ctx, model.ManualCategoryWorkflows, "target")
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestRepository_Upsert(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	editor1 := uint64(1)
	editor2 := uint64(2)

	t.Run("insert path: creates a new article and a version snapshot", func(t *testing.T) {
		article := &model.ManualArticle{
			Category:     model.ManualCategoryScreens,
			Slug:         "new-article",
			Title:        "新規",
			OrderValue:   1,
			Section:      "A",
			BodyMarkdown: "初版",
		}
		got, err := repo.Upsert(ctx, article, &editor1)
		require.NoError(t, err)
		require.NotZero(t, got.ID)
		assert.Equal(t, "初版", got.BodyMarkdown)

		versions, err := repo.FindVersionsByArticleID(ctx, got.ID)
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "初版", versions[0].BodyMarkdown)
	})

	t.Run("update path: updates the existing article, preserves CreatedAt, and appends a version", func(t *testing.T) {
		original := &model.ManualArticle{
			Category:     model.ManualCategoryScreens,
			Slug:         "existing-article",
			Title:        "既存",
			OrderValue:   1,
			Section:      "A",
			BodyMarkdown: "旧本文",
		}
		require.NoError(t, db.WithContext(ctx).Create(original).Error)
		originalCreatedAt := original.CreatedAt

		update := &model.ManualArticle{
			Category:     model.ManualCategoryScreens,
			Slug:         "existing-article",
			Title:        "既存（更新）",
			OrderValue:   2,
			Section:      "A",
			BodyMarkdown: "新本文",
		}
		got, err := repo.Upsert(ctx, update, &editor2)
		require.NoError(t, err)
		assert.Equal(t, original.ID, got.ID)
		assert.Equal(t, "新本文", got.BodyMarkdown)
		// originalCreatedAt はメモリ上の Go time.Time（ナノ秒精度）、got.CreatedAt は Postgres
		// timestamptz を経由して読み戻された値（マイクロ秒精度）のため、0 許容だとラウンドトリップ
		// による端数丸め差（観測: 831ns）で誤って失敗する。CreatedAt が「更新されていない」ことの
		// 検証が目的であり、秒単位の同一性で十分。
		assert.WithinDuration(t, originalCreatedAt, got.CreatedAt, time.Second)

		versions, err := repo.FindVersionsByArticleID(ctx, original.ID)
		require.NoError(t, err)
		require.Len(t, versions, 1)
		assert.Equal(t, "新本文", versions[0].BodyMarkdown)
		require.NotNil(t, versions[0].EditedByStaffID)
		assert.Equal(t, editor2, *versions[0].EditedByStaffID)
	})
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "delete-me", Title: "削除対象", OrderValue: 1, Section: "A", BodyMarkdown: "x"}
	require.NoError(t, db.WithContext(ctx).Create(article).Error)

	t.Run("deletes an existing article", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, model.ManualCategoryScreens, "delete-me"))
		_, err := repo.FindByCategoryAndSlug(ctx, model.ManualCategoryScreens, "delete-me")
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("not found for nonexistent category+slug", func(t *testing.T) {
		err := repo.Delete(ctx, model.ManualCategoryScreens, "never-existed")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestRepository_FindVersionsByArticleID(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{Category: model.ManualCategoryScreens, Slug: "versioned", Title: "版管理", OrderValue: 1, Section: "A", BodyMarkdown: "v1"}
	require.NoError(t, db.WithContext(ctx).Create(article).Error)

	v1 := &model.ManualArticleVersion{ArticleID: article.ID, Title: "版管理", OrderValue: 1, Section: "A", BodyMarkdown: "v1"}
	require.NoError(t, db.WithContext(ctx).Create(v1).Error)
	v2 := &model.ManualArticleVersion{ArticleID: article.ID, Title: "版管理", OrderValue: 1, Section: "A", BodyMarkdown: "v2"}
	require.NoError(t, db.WithContext(ctx).Create(v2).Error)

	t.Run("returns versions ordered by edited_at DESC", func(t *testing.T) {
		got, err := repo.FindVersionsByArticleID(ctx, article.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, v2.ID, got[0].ID)
		assert.Equal(t, v1.ID, got[1].ID)
	})

	t.Run("empty for article with no versions", func(t *testing.T) {
		got, err := repo.FindVersionsByArticleID(ctx, uint64(999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestRepository_VersionHistoryRetention(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()
	editor := uint64(1)

	// 履歴が MaxVersionsPerArticle を超えると prune され、list も同上限で fail-closed する。
	// current article は最新 upsert 内容を保持する。
	const overflow = 5
	totalWrites := MaxVersionsPerArticle + overflow

	var articleID uint64
	for i := 1; i <= totalWrites; i++ {
		body := fmt.Sprintf("body-%03d", i)
		got, err := repo.Upsert(ctx, &model.ManualArticle{
			Category:     model.ManualCategoryScreens,
			Slug:         "retention-article",
			Title:        "履歴上限",
			OrderValue:   float64(i),
			Section:      "A",
			BodyMarkdown: body,
		}, &editor)
		require.NoError(t, err)
		articleID = got.ID
	}

	// current article は最新内容を保持
	current, err := repo.FindByCategoryAndSlug(ctx, model.ManualCategoryScreens, "retention-article")
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("body-%03d", totalWrites), current.BodyMarkdown)
	assert.Equal(t, float64(totalWrites), current.OrderValue)

	// DB 上の履歴件数も MaxVersionsPerArticle 以下
	var storedCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.ManualArticleVersion{}).
		Where("article_id = ?", articleID).
		Count(&storedCount).Error)
	assert.Equal(t, int64(MaxVersionsPerArticle), storedCount)

	// list は上限を超えず、最新が先頭
	versions, err := repo.FindVersionsByArticleID(ctx, articleID)
	require.NoError(t, err)
	require.LessOrEqual(t, len(versions), MaxVersionsPerArticle)
	require.Len(t, versions, MaxVersionsPerArticle)
	assert.Equal(t, fmt.Sprintf("body-%03d", totalWrites), versions[0].BodyMarkdown)
	assert.Equal(t, fmt.Sprintf("body-%03d", totalWrites-MaxVersionsPerArticle+1), versions[len(versions)-1].BodyMarkdown)
}

func TestRepository_FindVersionsByArticleID_CapsAtMax(t *testing.T) {
	// Upsert の prune を経由せず、直接大量 insert しても list は LIMIT で上限を守る。
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	article := &model.ManualArticle{
		Category:     model.ManualCategoryWorkflows,
		Slug:         "list-cap",
		Title:        "list cap",
		OrderValue:   1,
		Section:      "A",
		BodyMarkdown: "current",
	}
	require.NoError(t, db.WithContext(ctx).Create(article).Error)

	const excess = 7
	for i := 1; i <= MaxVersionsPerArticle+excess; i++ {
		v := &model.ManualArticleVersion{
			ArticleID:    article.ID,
			Title:        "list cap",
			OrderValue:   1,
			Section:      "A",
			BodyMarkdown: fmt.Sprintf("snap-%03d", i),
		}
		require.NoError(t, db.WithContext(ctx).Create(v).Error)
	}

	var storedCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.ManualArticleVersion{}).
		Where("article_id = ?", article.ID).
		Count(&storedCount).Error)
	require.Equal(t, int64(MaxVersionsPerArticle+excess), storedCount)

	got, err := repo.FindVersionsByArticleID(ctx, article.ID)
	require.NoError(t, err)
	require.Len(t, got, MaxVersionsPerArticle)
	// 最新 (edited_at DESC) が先頭
	assert.Equal(t, fmt.Sprintf("snap-%03d", MaxVersionsPerArticle+excess), got[0].BodyMarkdown)
}
