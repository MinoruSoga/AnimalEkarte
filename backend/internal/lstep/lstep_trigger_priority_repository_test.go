package lstep

// lstep_trigger_priority_repository_test.go — LstepTriggerPriorityRepository 統合テスト (Q23)。
//
// 保護する不変条件:
//   - FindByClinicID は priority ASC 順で clinic_id 分離される。
//   - UpsertBatch は (clinic_id, trigger_type) の UNIQUE 制約に対する ON CONFLICT DO UPDATE で、
//     初回は新規作成、2回目以降は priority のみ更新する（重複行を作らない）。
//   - UpsertBatch は保存値の ClinicID を引数 clinicID で強制しつつ、入力を変更しない。
//   - FindPriorityByTriggerType は該当なしで NotFound を返し、clinic_id で分離される。

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

// setupLstepTriggerPriorityTestDB は lstep_trigger_priorities テーブルを用意する。
// LstepTriggerPriority.ClinicID/TriggerType には複合 uniqueIndex タグがあるため、
// AutoMigrate だけで UNIQUE(clinic_id, trigger_type) が再現され UpsertBatch の
// ON CONFLICT をそのまま検証できる。
func setupLstepTriggerPriorityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LstepTriggerPriority{}))
	db.Exec("TRUNCATE TABLE lstep_trigger_priorities CASCADE")
	return db
}

func TestLstepTriggerPriorityRepository_FindByClinicID(t *testing.T) {
	db := setupLstepTriggerPriorityTestDB(t)
	repo := NewLstepTriggerPriorityRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)
	require.NoError(t, repo.UpsertBatch(ctx, clinicA, []model.LstepTriggerPriority{
		{TriggerType: "trigger-low", Priority: 5},
		{TriggerType: "trigger-high", Priority: 1},
	}))
	require.NoError(t, repo.UpsertBatch(ctx, clinicB, []model.LstepTriggerPriority{
		{TriggerType: "trigger-other", Priority: 2},
	}))

	t.Run("priority ASC 順で返す", func(t *testing.T) {
		items, err := repo.FindByClinicID(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "trigger-high", items[0].TriggerType, "priority=1 が先頭")
		assert.Equal(t, "trigger-low", items[1].TriggerType)
	})

	t.Run("別クリニックの設定は含まれない（clinic_id分離）", func(t *testing.T) {
		items, err := repo.FindByClinicID(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "trigger-other", items[0].TriggerType)
	})

	t.Run("該当なしは空スライスを返す", func(t *testing.T) {
		items, err := repo.FindByClinicID(ctx, 999)
		require.NoError(t, err)
		assert.Empty(t, items)
	})
}

func TestLstepTriggerPriorityRepository_UpsertBatch(t *testing.T) {
	db := setupLstepTriggerPriorityTestDB(t)
	repo := NewLstepTriggerPriorityRepository(db)
	ctx := context.Background()

	t.Run("空スライスはno-op", func(t *testing.T) {
		require.NoError(t, repo.UpsertBatch(ctx, 1, []model.LstepTriggerPriority{}))
		items, err := repo.FindByClinicID(ctx, 1)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("同一(clinic_id, trigger_type)はUPSERTでpriorityのみ更新し重複行を作らない", func(t *testing.T) {
		require.NoError(t, repo.UpsertBatch(ctx, 2, []model.LstepTriggerPriority{
			{TriggerType: "dormant_365d", Priority: 5},
		}))
		require.NoError(t, repo.UpsertBatch(ctx, 2, []model.LstepTriggerPriority{
			{TriggerType: "dormant_365d", Priority: 1},
		}))

		var count int64
		require.NoError(t, db.Model(&model.LstepTriggerPriority{}).
			Where("clinic_id = ? AND trigger_type = ?", 2, "dormant_365d").Count(&count).Error)
		assert.Equal(t, int64(1), count, "重複行が作られてはならない")

		priority, err := repo.FindPriorityByTriggerType(ctx, 2, "dormant_365d")
		require.NoError(t, err)
		assert.Equal(t, 1, priority, "priorityは最新値に更新されるべき")
	})

	t.Run("保存時のClinicIDは引数clinicIDで強制し入力は変更しない", func(t *testing.T) {
		items := []model.LstepTriggerPriority{
			{ClinicID: 999, TriggerType: "override-check", Priority: 4}, // 意図的に別clinic_idを指定
		}
		require.NoError(t, repo.UpsertBatch(ctx, 3, items))
		assert.Equal(t, uint64(999), items[0].ClinicID, "呼び出し元のスライス要素を変更してはならない")

		var count int64
		require.NoError(t, db.Model(&model.LstepTriggerPriority{}).Where("clinic_id = ? AND trigger_type = ?", 3, "override-check").Count(&count).Error)
		assert.Equal(t, int64(1), count)

		require.NoError(t, db.Model(&model.LstepTriggerPriority{}).Where("clinic_id = ? AND trigger_type = ?", 999, "override-check").Count(&count).Error)
		assert.Equal(t, int64(0), count, "元々指定していた clinic_id=999 には作成されない")
	})
}

func TestLstepTriggerPriorityRepository_FindPriorityByTriggerType(t *testing.T) {
	db := setupLstepTriggerPriorityTestDB(t)
	repo := NewLstepTriggerPriorityRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.UpsertBatch(ctx, 1, []model.LstepTriggerPriority{
		{TriggerType: "checkup_followup", Priority: 3},
	}))

	t.Run("存在する優先度を返す", func(t *testing.T) {
		priority, err := repo.FindPriorityByTriggerType(ctx, 1, "checkup_followup")
		require.NoError(t, err)
		assert.Equal(t, 3, priority)
	})

	t.Run("存在しないtrigger_typeはNotFoundを返す", func(t *testing.T) {
		_, err := repo.FindPriorityByTriggerType(ctx, 1, "nonexistent-trigger")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックのtrigger_typeは見えない（clinic_id分離）", func(t *testing.T) {
		_, err := repo.FindPriorityByTriggerType(ctx, 2, "checkup_followup")
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
