package staff_test

// repository_test.go — Repository 統合テスト。
//
// 保護する不変条件:
//   - FindAll / FindByID は clinic_id でテナント隔離される。
//   - FindAll はソフトデリート済みテンプレートを除外し sort_order/name 順で返す。
//   - Delete はソフトデリートであり、以後 FindByID/FindAll から除外される。
//   - Reorder は指定順に sort_order=1..n を割り当て、他クリニックの id を含むと失敗する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	. "github.com/animal-ekarte/backend/internal/staff"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupShiftTemplateTestDB は shift_templates / shift_template_breaks を整備する。
func setupShiftTemplateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ShiftTemplate{}, &model.ShiftTemplateBreak{}))
	// shift_template_breaks.break_start/break_end can be left as "timestamp with time zone" in a
	// pre-existing ekarte_db_test if the table was ever created before model.ShiftTemplateBreak's
	// `gorm:"type:time"` tag existed — AutoMigrate never ALTERs an existing column's type, only
	// adds missing ones. Force the correct type every run (no-op cast if already "time").
	db.Exec(`ALTER TABLE shift_template_breaks ALTER COLUMN break_start TYPE time USING break_start::time`)
	db.Exec(`ALTER TABLE shift_template_breaks ALTER COLUMN break_end TYPE time USING break_end::time`)
	db.Exec("TRUNCATE TABLE shift_template_breaks CASCADE")
	db.Exec("TRUNCATE TABLE shift_templates CASCADE")
	return db
}

func makeShiftTemplate(t *testing.T, db *gorm.DB, clinicID uint64, name string, sortOrder int) *model.ShiftTemplate {
	t.Helper()
	tpl := &model.ShiftTemplate{
		ClinicID:  clinicID,
		Name:      name,
		ShiftType: model.ShiftTypeFull,
		SortOrder: sortOrder,
	}
	require.NoError(t, db.WithContext(context.Background()).Omit("Breaks").Create(tpl).Error)
	return tpl
}

func TestShiftTemplateRepository_FindAll(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	tplB := makeShiftTemplate(t, db, clinicA, "早番", 2)
	tplA := makeShiftTemplate(t, db, clinicA, "遅番", 1)
	deleted := makeShiftTemplate(t, db, clinicA, "廃止済み", 3)
	require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))
	makeShiftTemplate(t, db, clinicB, "別クリニック", 1)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "ソフトデリート済み・別クリニックは除外される")
	assert.Equal(t, tplA.ID, got[0].ID, "sort_order ASC で先頭は sort_order=1")
	assert.Equal(t, tplB.ID, got[1].ID)
}

func TestShiftTemplateRepository_FindByID(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	tpl := makeShiftTemplate(t, db, clinicA, "早番", 1)

	t.Run("found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, tpl.ID)
		require.NoError(t, err)
		assert.Equal(t, tpl.ID, got.ID)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, tpl.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deleted template is not found", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, tpl.ID))
		got, err := repo.FindByID(ctx, clinicA, tpl.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftTemplateRepository_Create(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	tpl := &model.ShiftTemplate{ClinicID: 1, Name: "新規テンプレ", ShiftType: model.ShiftTypeFull}
	require.NoError(t, repo.Create(ctx, tpl))
	assert.NotZero(t, tpl.ID)
}

// BUG-455-S7: gorm default:true omits zero bools from INSERT.
func TestShiftTemplateRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	tpl := &model.ShiftTemplate{
		ClinicID:  1,
		Name:      "inactive template",
		ShiftType: model.ShiftTypeFull,
		IsActive:  false,
	}
	require.NoError(t, repo.Create(ctx, tpl))
	assert.False(t, tpl.IsActive)

	got, err := repo.FindByID(ctx, 1, tpl.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	var raw bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ShiftTemplate{}).
		Select("is_active").
		Where("id = ?", tpl.ID).
		Scan(&raw).Error)
	assert.False(t, raw, "raw is_active must be false")
}

func TestShiftTemplateRepository_Update(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	tpl := makeShiftTemplate(t, db, clinicA, "早番", 1)

	t.Run("updates and returns the refreshed entity", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, tpl.ID, map[string]any{"name": "改名後"})
		require.NoError(t, err)
		assert.Equal(t, "改名後", got.Name)
	})

	t.Run("not found for nonexistent id", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, uint64(999999), map[string]any{"name": "x"})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("clinic isolation", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicB, tpl.ID, map[string]any{"name": "乗っ取り"})
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftTemplateRepository_Delete(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	tpl := makeShiftTemplate(t, db, clinicA, "早番", 1)

	t.Run("clinic isolation: wrong clinic cannot delete", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, tpl.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("soft-deletes successfully", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, tpl.ID))

		var count int64
		require.NoError(t, db.Unscoped().Model(&model.ShiftTemplate{}).Where("id = ?", tpl.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count, "物理的には行が残る（ソフトデリート）")

		var scopedCount int64
		require.NoError(t, db.Model(&model.ShiftTemplate{}).Where("id = ?", tpl.ID).Count(&scopedCount).Error)
		assert.Equal(t, int64(0), scopedCount, "デフォルトスコープでは除外される")
	})

	t.Run("not found for already-deleted id", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, tpl.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestShiftTemplateRepository_UpdateBreaks(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const clinicA = uint64(1)
	tpl := makeShiftTemplate(t, db, clinicA, "早番", 1)

	t.Run("creates breaks", func(t *testing.T) {
		breaks := []model.ShiftTemplateBreak{{BreakStart: "12:00", BreakEnd: "13:00"}}
		require.NoError(t, repo.UpdateBreaks(ctx, tpl.ID, breaks))

		var count int64
		require.NoError(t, db.Model(&model.ShiftTemplateBreak{}).Where("shift_template_id = ?", tpl.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("replaces existing breaks", func(t *testing.T) {
		breaks := []model.ShiftTemplateBreak{
			{BreakStart: "10:00", BreakEnd: "10:15"},
			{BreakStart: "15:00", BreakEnd: "15:15"},
		}
		require.NoError(t, repo.UpdateBreaks(ctx, tpl.ID, breaks))

		var got []model.ShiftTemplateBreak
		require.NoError(t, db.Where("shift_template_id = ?", tpl.ID).Order("break_start ASC").Find(&got).Error)
		require.Len(t, got, 2)
		// Postgres の time 型カラムは文字列スキャン時 "HH:MM:SS" 形式で返る（入力が "HH:MM" でも同様）。
		assert.Equal(t, "10:00:00", got[0].BreakStart)
	})

	t.Run("empty slice removes all breaks", func(t *testing.T) {
		require.NoError(t, repo.UpdateBreaks(ctx, tpl.ID, []model.ShiftTemplateBreak{}))

		var count int64
		require.NoError(t, db.Model(&model.ShiftTemplateBreak{}).Where("shift_template_id = ?", tpl.ID).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}

func TestShiftTemplateRepository_Reorder(t *testing.T) {
	db := setupShiftTemplateTestDB(t)
	repo := NewShiftTemplateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)
	tpl1 := makeShiftTemplate(t, db, clinicA, "A", 1)
	tpl2 := makeShiftTemplate(t, db, clinicA, "B", 2)
	otherClinicTpl := makeShiftTemplate(t, db, clinicB, "別クリニック", 1)

	t.Run("reassigns sort_order in the given sequence", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{tpl2.ID, tpl1.ID}))

		got1, err := repo.FindByID(ctx, clinicA, tpl1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, tpl2.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, got1.SortOrder)
		assert.Equal(t, 1, got2.SortOrder)
	})

	t.Run("fails when an id belongs to another clinic", func(t *testing.T) {
		err := repo.Reorder(ctx, clinicA, []uint64{otherClinicTpl.ID})
		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "他クリニックのIDはInvalidInputで拒否される: %v", err)
	})
}


