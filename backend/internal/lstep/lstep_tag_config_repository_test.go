package lstep

// lstep_tag_config_repository_test.go — LstepTagConfigRepository 統合テスト。
//
// 保護する不変条件:
//   - FindAllAutoManagedPrefixes は category, prefix 昇順で全件返す。
//   - FindAllConditionTagMappings は condition_code 昇順で全件返す。
//   - FindAllSendPurposeTagPrefixes は purpose 昇順で全件返す。
//   - 各 Delete* は成功時に行を削除し、存在しない ID では NotFound を返す。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupLstepTagConfigTestDB は lstep_auto_managed_prefixes / lstep_condition_tag_mappings /
// lstep_send_purpose_tag_prefixes テーブルを用意する（いずれも clinic_id を持たないグローバルマスタ）。
func setupLstepTagConfigTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.LstepAutoManagedPrefix{},
		&model.LstepConditionTagMapping{},
		&model.LstepSendPurposeTagPrefix{},
	))
	db.Exec("TRUNCATE TABLE lstep_auto_managed_prefixes CASCADE")
	db.Exec("TRUNCATE TABLE lstep_condition_tag_mappings CASCADE")
	db.Exec("TRUNCATE TABLE lstep_send_purpose_tag_prefixes CASCADE")
	return db
}

func TestLstepTagConfigRepository_AutoManagedPrefix(t *testing.T) {
	db := setupLstepTagConfigTestDB(t)
	repo := NewLstepTagConfigRepository(db)
	ctx := context.Background()

	t.Run("Create + FindAll は category, prefix 昇順で返す", func(t *testing.T) {
		descB := "B category prefix"
		require.NoError(t, repo.CreateAutoManagedPrefix(ctx, &model.LstepAutoManagedPrefix{Prefix: "zzz_b", Category: "B", Description: &descB}))
		require.NoError(t, repo.CreateAutoManagedPrefix(ctx, &model.LstepAutoManagedPrefix{Prefix: "aaa_a", Category: "A"}))

		prefixes, err := repo.FindAllAutoManagedPrefixes(ctx)
		require.NoError(t, err)
		require.Len(t, prefixes, 2)
		assert.Equal(t, "A", prefixes[0].Category, "category 昇順であること")
		assert.Equal(t, "aaa_a", prefixes[0].Prefix)
		assert.Equal(t, "B", prefixes[1].Category)
		require.NotNil(t, prefixes[1].Description)
		assert.Equal(t, descB, *prefixes[1].Description)
	})

	t.Run("DeleteAutoManagedPrefix は行を削除する", func(t *testing.T) {
		p := &model.LstepAutoManagedPrefix{Prefix: "delete_me_prefix", Category: "C1"}
		require.NoError(t, repo.CreateAutoManagedPrefix(ctx, p))
		require.NoError(t, repo.DeleteAutoManagedPrefix(ctx, p.ID))

		var count int64
		require.NoError(t, db.Model(&model.LstepAutoManagedPrefix{}).Where("id = ?", p.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("存在しない ID の DeleteAutoManagedPrefix は NotFound を返す", func(t *testing.T) {
		err := repo.DeleteAutoManagedPrefix(ctx, 9_999_999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLstepTagConfigRepository_ConditionTagMapping(t *testing.T) {
	db := setupLstepTagConfigTestDB(t)
	repo := NewLstepTagConfigRepository(db)
	ctx := context.Background()

	t.Run("Create + FindAll は condition_code 昇順で返す", func(t *testing.T) {
		require.NoError(t, repo.CreateConditionTagMapping(ctx, &model.LstepConditionTagMapping{ConditionCode: "Z_COND", TagName: "tag-z"}))
		require.NoError(t, repo.CreateConditionTagMapping(ctx, &model.LstepConditionTagMapping{ConditionCode: "A_COND", TagName: "tag-a"}))

		mappings, err := repo.FindAllConditionTagMappings(ctx)
		require.NoError(t, err)
		require.Len(t, mappings, 2)
		assert.Equal(t, "A_COND", mappings[0].ConditionCode)
		assert.Equal(t, "Z_COND", mappings[1].ConditionCode)
	})

	t.Run("DeleteConditionTagMapping は行を削除する", func(t *testing.T) {
		m := &model.LstepConditionTagMapping{ConditionCode: "DELETE_COND", TagName: "delete-tag"}
		require.NoError(t, repo.CreateConditionTagMapping(ctx, m))
		require.NoError(t, repo.DeleteConditionTagMapping(ctx, m.ID))

		var count int64
		require.NoError(t, db.Model(&model.LstepConditionTagMapping{}).Where("id = ?", m.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("存在しない ID の DeleteConditionTagMapping は NotFound を返す", func(t *testing.T) {
		err := repo.DeleteConditionTagMapping(ctx, 9_999_998)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestLstepTagConfigRepository_SendPurposeTagPrefix(t *testing.T) {
	db := setupLstepTagConfigTestDB(t)
	repo := NewLstepTagConfigRepository(db)
	ctx := context.Background()

	t.Run("Create + FindAll は purpose 昇順で返す", func(t *testing.T) {
		require.NoError(t, repo.CreateSendPurposeTagPrefix(ctx, &model.LstepSendPurposeTagPrefix{Purpose: "z_purpose", TagPrefix: "zp_"}))
		require.NoError(t, repo.CreateSendPurposeTagPrefix(ctx, &model.LstepSendPurposeTagPrefix{Purpose: "a_purpose", TagPrefix: "ap_"}))

		prefixes, err := repo.FindAllSendPurposeTagPrefixes(ctx)
		require.NoError(t, err)
		require.Len(t, prefixes, 2)
		assert.Equal(t, "a_purpose", prefixes[0].Purpose)
		assert.Equal(t, "z_purpose", prefixes[1].Purpose)
	})

	t.Run("DeleteSendPurposeTagPrefix は行を削除する", func(t *testing.T) {
		p := &model.LstepSendPurposeTagPrefix{Purpose: "delete_purpose", TagPrefix: "dp_"}
		require.NoError(t, repo.CreateSendPurposeTagPrefix(ctx, p))
		require.NoError(t, repo.DeleteSendPurposeTagPrefix(ctx, p.ID))

		var count int64
		require.NoError(t, db.Model(&model.LstepSendPurposeTagPrefix{}).Where("id = ?", p.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("存在しない ID の DeleteSendPurposeTagPrefix は NotFound を返す", func(t *testing.T) {
		err := repo.DeleteSendPurposeTagPrefix(ctx, 9_999_997)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
