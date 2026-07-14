package repository

// estimate_repository_test.go
// estimate_repository.go の実 DB 結合テスト（#212: internal/repository カバレッジ向上）。
// makeTestOwner は db_setup_test.go を再利用する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupEstimateRepoTestDB は estimate_repository のテストに必要なテーブルを整備する。
func setupEstimateRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&model.Estimate{}, &model.EstimateItem{}))
	db.Exec("TRUNCATE TABLE estimate_items CASCADE")
	db.Exec("TRUNCATE TABLE estimates CASCADE")
	return db
}

// makeEstimate は repo.Create 経由で見積書を作成する（Create メソッド自体もカバー対象にする）。
func makeEstimate(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, status model.EstimateStatus) *model.Estimate {
	t.Helper()
	e := &model.Estimate{
		ClinicID: clinicID,
		OwnerID:  &ownerID,
		Status:   status,
		Title:    "見積テスト",
	}
	require.NoError(t, NewEstimateRepository(db).Create(context.Background(), e))
	return e
}

func makeEstimateItem(t *testing.T, db *gorm.DB, estimateID uint64, name string) *model.EstimateItem {
	t.Helper()
	item := &model.EstimateItem{
		EstimateID: estimateID,
		Name:       name,
		Category:   model.ItemCategoryOther,
		UnitPrice:  1000,
		Quantity:   1,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(item).Error)
	return item
}

func TestEstimateRepository_Create_FindByID(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "見積飼主")

	t.Run("作成した見積書をOwner/Items付きで取得できる", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)
		makeEstimateItem(t, db, e.ID, "健康診断パック")

		got, err := repo.FindByID(ctx, clinicA, e.ID)
		require.NoError(t, err)
		assert.Equal(t, model.EstimateStatusDraft, got.Status)
		require.NotNil(t, got.Owner)
		assert.Equal(t, owner.ID, got.Owner.ID)
		require.Len(t, got.Items, 1)
		assert.Equal(t, "健康診断パック", got.Items[0].Name)
	})

	t.Run("別クリニックからはNotFound", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		_, err := repo.FindByID(ctx, clinicB, e.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "expected NotFound, got %v", err)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestEstimateRepository_FindAll(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ownerA := makeTestOwner(t, db, clinicA, "A飼主")
	ownerB := makeTestOwner(t, db, clinicB, "B飼主")
	makeEstimate(t, db, clinicA, ownerA.ID, model.EstimateStatusDraft)
	makeEstimate(t, db, clinicA, ownerA.ID, model.EstimateStatusSent)
	makeEstimate(t, db, clinicB, ownerB.ID, model.EstimateStatusDraft)

	t.Run("clinic_id でスコープされ他院データが混入しない", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, got, 2)
	})

	t.Run("status フィルタが機能する", func(t *testing.T) {
		draft := string(model.EstimateStatusDraft)
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, &draft, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, got, 1)
		assert.Equal(t, model.EstimateStatusDraft, got[0].Status)
	})

	t.Run("owner_id フィルタが機能する", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, &ownerA.ID, nil, nil, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, got, 2)
	})

	t.Run("ページングが機能する", func(t *testing.T) {
		got, total, err := repo.FindAll(ctx, clinicA, nil, nil, nil, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, got, 1)
	})
}

func TestEstimateRepository_Update(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "更新飼主")

	t.Run("同一クリニックの更新は成功する", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		err := repo.Update(ctx, clinicA, e.ID, map[string]any{"status": model.EstimateStatusSent})
		require.NoError(t, err)

		got, err := repo.FindByID(ctx, clinicA, e.ID)
		require.NoError(t, err)
		assert.Equal(t, model.EstimateStatusSent, got.Status)
	})

	t.Run("他院からの更新はNotFound", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		err := repo.Update(ctx, clinicB, e.ID, map[string]any{"status": model.EstimateStatusSent})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Update(ctx, clinicA, 999999, map[string]any{"status": model.EstimateStatusSent})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// TestEstimateRepository_UpdateIfNotLocked は approved/rejected への TOCTOU を
// status NOT IN 述語で原子的に拒否することを検証する（0 行 → Conflict）。
func TestEstimateRepository_UpdateIfNotLocked(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "UpdateIfNotLocked飼主")

	t.Run("draft status は更新に成功する", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		got, err := repo.UpdateIfNotLocked(ctx, clinicA, e.ID, map[string]any{
			"title": "更新後タイトル",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "更新後タイトル", got.Title)
		assert.Equal(t, model.EstimateStatusDraft, got.Status)
	})

	t.Run("sent status は更新に成功する", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusSent)

		got, err := repo.UpdateIfNotLocked(ctx, clinicA, e.ID, map[string]any{
			"title": "sent更新",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "sent更新", got.Title)
	})

	t.Run("approved status は Conflict（行は変わらない）", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusApproved)
		originalTitle := e.Title

		got, err := repo.UpdateIfNotLocked(ctx, clinicA, e.ID, map[string]any{
			"title": "改ざん試行",
		})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		unchanged, findErr := repo.FindByID(ctx, clinicA, e.ID)
		require.NoError(t, findErr)
		assert.Equal(t, originalTitle, unchanged.Title)
		assert.Equal(t, model.EstimateStatusApproved, unchanged.Status)
	})

	t.Run("rejected status は Conflict（行は変わらない）", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusRejected)
		originalTitle := e.Title

		got, err := repo.UpdateIfNotLocked(ctx, clinicA, e.ID, map[string]any{
			"title": "改ざん試行",
		})
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		unchanged, findErr := repo.FindByID(ctx, clinicA, e.ID)
		require.NoError(t, findErr)
		assert.Equal(t, originalTitle, unchanged.Title)
		assert.Equal(t, model.EstimateStatusRejected, unchanged.Status)
	})
}

func TestEstimateRepository_Delete(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeTestOwner(t, db, clinicA, "削除飼主")

	t.Run("同一クリニックの削除は成功する", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		require.NoError(t, repo.Delete(ctx, clinicA, e.ID))

		_, err := repo.FindByID(ctx, clinicA, e.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("他院からの削除はNotFound（対象は削除されない）", func(t *testing.T) {
		e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)

		err := repo.Delete(ctx, clinicB, e.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		got, err := repo.FindByID(ctx, clinicA, e.ID)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("存在しないIDはNotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

// clinic isolation は count_clinic_scope_isolation_test.go の
// TestEstimateRepository_CountItemsByEstimateID_ClinicIsolation で既にカバー済みのため、
// ここでは論理削除フィルタ（重複していない観点）のみを検証する。
func TestEstimateRepository_CountItemsByEstimateID_ExcludesSoftDeleted(t *testing.T) {
	db := setupEstimateRepoTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	owner := makeTestOwner(t, db, clinicA, "件数飼主")
	e := makeEstimate(t, db, clinicA, owner.ID, model.EstimateStatusDraft)
	makeEstimateItem(t, db, e.ID, "項目1")
	makeEstimateItem(t, db, e.ID, "項目2")

	t.Run("論理削除済みitemはカウントされない", func(t *testing.T) {
		item := makeEstimateItem(t, db, e.ID, "削除予定項目")
		require.NoError(t, db.Delete(item).Error)

		count, err := repo.CountItemsByEstimateID(ctx, clinicA, e.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})
}
