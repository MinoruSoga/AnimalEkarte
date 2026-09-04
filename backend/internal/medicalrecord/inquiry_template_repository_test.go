package medicalrecord

// repository_test.go — Repository の統合テスト（内部カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / Reorder / CountUsageByInquiryTemplateID
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
// CountUsageByInquiryTemplateID は現スキーマで常に 0 を返す定数実装（inquiry_answers 未実装のため）。

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupInquiryTemplateTestDB は inquiry_templates テーブルを整備する。
func setupInquiryTemplateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.InquiryTemplate{}))
	db.Exec("TRUNCATE TABLE inquiry_templates CASCADE")
	return db
}

func TestInquiryTemplateRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tplB := &model.InquiryTemplate{ClinicID: clinicB, Title: "医院Bの定型文", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(tplB).Error)
	second := &model.InquiryTemplate{ClinicID: clinicA, Title: "Bタイトル", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	first := &model.InquiryTemplate{ClinicID: clinicA, Title: "Aタイトル", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A の定型文のみ返る")
	assert.Equal(t, first.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)
	for _, tpl := range got {
		assert.NotEqual(t, tplB.ID, tpl.ID, "別クリニックの定型文が混入してはならない")
	}
}

func TestInquiryTemplateRepository_FindByID(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tpl := &model.InquiryTemplate{ClinicID: clinicA, Title: "問診定型文A"}
	require.NoError(t, db.WithContext(ctx).Create(tpl).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, tpl.ID)
		require.NoError(t, err)
		assert.Equal(t, "問診定型文A", got.Title)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, tpl.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestInquiryTemplateRepository_Create(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	tpl := &model.InquiryTemplate{ClinicID: clinicA, Title: "新規定型文"}
	require.NoError(t, repo.Create(ctx, tpl))
	assert.NotZero(t, tpl.ID)

	got, err := repo.FindByID(ctx, clinicA, tpl.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規定型文", got.Title)
}

func TestInquiryTemplateRepository_Update(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tpl := &model.InquiryTemplate{ClinicID: clinicA, Title: "旧タイトル"}
	require.NoError(t, db.WithContext(ctx).Create(tpl).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, tpl.ID, map[string]any{"title": "新タイトル"})
		require.NoError(t, err)
		assert.Equal(t, "新タイトル", got.Title)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, tpl.ID, map[string]any{"title": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"title": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestInquiryTemplateRepository_Delete(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	tpl := &model.InquiryTemplate{ClinicID: clinicA, Title: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(tpl).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, tpl.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされ FindAll から除外される", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, tpl.ID))

		_, err := repo.FindByID(ctx, clinicA, tpl.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, tpl.ID, x.ID)
		}

		var raw model.InquiryTemplate
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", tpl.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestInquiryTemplateRepository_Delete_JoinsAmbientTransaction(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	sentinel := errors.New("simulated downstream failure")

	tpl := &model.InquiryTemplate{ClinicID: clinicA, Title: "原子削除ロールバック対象"}
	require.NoError(t, db.WithContext(ctx).Create(tpl).Error)

	txErr := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := repo.Delete(txCtx, clinicA, tpl.ID); err != nil {
			return err
		}
		return sentinel
	})
	require.ErrorIs(t, txErr, sentinel)

	got, err := repo.FindByID(ctx, clinicA, tpl.ID)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, tpl.ID, got.ID)
}

func TestInquiryTemplateRepository_Reorder(t *testing.T) {
	db := setupInquiryTemplateTestDB(t)
	repo := NewInquiryTemplateRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t1 := &model.InquiryTemplate{ClinicID: clinicA, Title: "T1"}
	require.NoError(t, db.WithContext(ctx).Create(t1).Error)
	t2 := &model.InquiryTemplate{ClinicID: clinicA, Title: "T2"}
	require.NoError(t, db.WithContext(ctx).Create(t2).Error)
	t3 := &model.InquiryTemplate{ClinicID: clinicA, Title: "T3"}
	require.NoError(t, db.WithContext(ctx).Create(t3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{t3.ID, t1.ID, t2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, t3.ID, got[0].ID)
		assert.Equal(t, t1.ID, got[1].ID)
		assert.Equal(t, t2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.InquiryTemplate{ClinicID: clinicB, Title: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{t1.ID, other.ID})
		require.Error(t, err)
	})
}

func TestInquiryTemplateRepository_CountUsageByInquiryTemplateID(t *testing.T) {
	repo := NewInquiryTemplateRepository(setupInquiryTemplateTestDB(t))
	ctx := context.Background()

	// 現スキーマに inquiry_template_id を参照する FK テーブルが存在しないため常に 0 を返す
	// （PO判断 2026-05-25: inquiry_answers は当面実装しない）。
	count, err := repo.CountUsageByInquiryTemplateID(ctx, uint64(1), uint64(1))
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
