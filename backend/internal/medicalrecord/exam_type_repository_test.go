package medicalrecord

// exam_type_repository_test.go — ExamTypeRepository の統合テスト（内部カバレッジ向上）。
//
// 移動元: internal/repository/exam_type_repository_test.go — BE9-2C roll-up。
//
// makeExamTypeMaster/makeExaminationRec は internal/repository/examination_repository_test.go
// / examination_repository_tx_atomicity_test.go からも使う共有ヘルパーだったため、旧
// パッケージ側にそのまま残置（examination は BE9-2D 対象、本バッチ対象外）。ここでは
// medicalrecord 自身のテスト専用ローカルコピーとして再定義する（挙動は同一、export はしない
// ため2パッケージ間の重複は問題にならない）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / Reorder /
//
//	CountUsageByExamTypeID / CountChildrenByParentID
//
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。

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
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupExamTypeTestDB は exam_types / exam_type_fields / exams を整備する。
// CountUsageByExamTypeID は exams.exam_type_id を直接参照する（Examination は独自の clinic_id を持つため
// checkup_type と異なり medical_record/pet/owner の同時整備は不要）。
func setupExamTypeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.ExaminationType{},
		&model.ExamTypeField{},
		&model.Examination{},
		&model.ExamResult{},
		&model.ExamReferenceRange{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE exam_reference_ranges CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE exam_results CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE exams CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE exam_type_fields CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE exam_types CASCADE").Error)
	return db
}

// makeExamTypeMaster はテスト用の ExaminationType を作成して返す。
func makeExamTypeMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.ExaminationType {
	t.Helper()
	et := &model.ExaminationType{ClinicID: clinicID, Name: name, IsActive: true}
	require.NoError(t, db.WithContext(context.Background()).Create(et).Error)
	return et
}

// makeExaminationRec はテスト用の Examination を作成して返す。
func makeExaminationRec(t *testing.T, db *gorm.DB, ex *model.Examination) *model.Examination {
	t.Helper()
	require.NoError(t, db.WithContext(context.Background()).Create(ex).Error)
	return ex
}

func TestExamTypeRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	etB := &model.ExaminationType{ClinicID: clinicB, Name: "医院Bの検査種別", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(etB).Error)
	etSecond := &model.ExaminationType{ClinicID: clinicA, Name: "Bタイプ", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(etSecond).Error)
	etFirst := &model.ExaminationType{ClinicID: clinicA, Name: "Aタイプ", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(etFirst).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A の検査種別のみ返る")
	assert.Equal(t, etFirst.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, etSecond.ID, got[1].ID)
	for _, et := range got {
		assert.NotEqual(t, etB.ID, et.ID, "別クリニックの検査種別が混入してはならない")
	}
}

// seedExamTypesForBoundTest bulk-inserts clinic-scoped examination types with
// deterministic sort_order/name (bound-et-00001..) for MaxMasterListRows regressions.
func seedExamTypesForBoundTest(t *testing.T, db *gorm.DB, clinicID uint64, count int) {
	t.Helper()
	rows := make([]model.ExaminationType, 0, count)
	for i := 1; i <= count; i++ {
		rows = append(rows, model.ExaminationType{
			ClinicID:  clinicID,
			Name:      fmt.Sprintf("bound-et-%05d", i),
			SortOrder: i,
			IsActive:  true,
		})
	}
	require.NoError(t, db.WithContext(context.Background()).CreateInBatches(rows, 500).Error)
}

func assertExamTypesOrderedBySortThenName(t *testing.T, got []model.ExaminationType) {
	t.Helper()
	for i := 1; i < len(got); i++ {
		prev, cur := got[i-1], got[i]
		if prev.SortOrder == cur.SortOrder {
			assert.LessOrEqual(t, prev.Name, cur.Name, "same sort_order must order by name ASC")
			continue
		}
		assert.Less(t, prev.SortOrder, cur.SortOrder, "must order by sort_order ASC")
	}
}

// TestExamTypeRepository_FindAll_AppliesMaxMasterListRows is G2F-11: FindAll applies
// persistence.MaxMasterListRows (same as consultation/vaccine/procedure), excludes
// other clinics and soft-deleted rows, and keeps sort_order ASC, name ASC.
func TestExamTypeRepository_FindAll_AppliesMaxMasterListRows(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	limit := persistence.MaxMasterListRows

	t.Run("exact MaxMasterListRows returns all rows in order", func(t *testing.T) {
		require.NoError(t, db.Exec("TRUNCATE TABLE exam_types CASCADE").Error)
		seedExamTypesForBoundTest(t, db, clinicA, limit)

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, limit, "exact bound must return every active clinic row")
		assert.Equal(t, 1, got[0].SortOrder)
		assert.Equal(t, limit, got[len(got)-1].SortOrder)
		assertExamTypesOrderedBySortThenName(t, got)
	})

	t.Run("limit+1 caps at MaxMasterListRows and excludes foreign/soft-deleted", func(t *testing.T) {
		require.NoError(t, db.Exec("TRUNCATE TABLE exam_types CASCADE").Error)
		// limit+2 active for clinic A so that soft-deleting one still leaves limit+1 active.
		seedExamTypesForBoundTest(t, db, clinicA, limit+2)

		// Cross-clinic rows that would sort first if clinic isolation regressed.
		foreign := []model.ExaminationType{
			{ClinicID: clinicB, Name: "aaa-other-clinic-1", SortOrder: 0, IsActive: true},
			{ClinicID: clinicB, Name: "aaa-other-clinic-2", SortOrder: 0, IsActive: true},
		}
		require.NoError(t, db.WithContext(ctx).Create(&foreign).Error)

		// Soft-delete a row that would otherwise appear in the capped window.
		var softDeleted model.ExaminationType
		require.NoError(t, db.WithContext(ctx).
			Where("clinic_id = ? AND sort_order = ?", clinicA, 2).
			First(&softDeleted).Error)
		require.NoError(t, repo.Delete(ctx, clinicA, softDeleted.ID))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, limit, "must cap at MaxMasterListRows when more active rows exist")
		assertExamTypesOrderedBySortThenName(t, got)

		// Soft-deleted must not appear; foreign clinic rows must not mix in.
		for _, et := range got {
			assert.NotEqual(t, softDeleted.ID, et.ID, "soft-deleted row must be excluded")
			assert.Equal(t, clinicA, et.ClinicID, "cross-clinic rows must not mix into result")
			assert.NotEqual(t, foreign[0].ID, et.ID)
			assert.NotEqual(t, foreign[1].ID, et.ID)
		}

		// After soft-deleting sort_order=2, ordered active starts at 1 then 3..
		assert.Equal(t, 1, got[0].SortOrder)
		assert.Equal(t, "bound-et-00001", got[0].Name)
		assert.Equal(t, 3, got[1].SortOrder)
		// First `limit` of remaining active ordered set ends at sort_order = limit+1.
		assert.Equal(t, limit+1, got[len(got)-1].SortOrder)
	})
}

func TestExamTypeRepository_ItemsPreload_ClinicIsolation(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	examType := makeExamTypeMaster(t, db, clinicA, "医院Aの検査種別")
	fieldA := &model.ExamTypeField{
		ExamTypeID: examType.ID,
		ClinicID:   clinicA,
		Name:       "医院Aの項目",
	}
	require.NoError(t, db.WithContext(ctx).Create(fieldA).Error)
	fieldB := &model.ExamTypeField{
		ExamTypeID: examType.ID,
		ClinicID:   clinicB,
		Name:       "医院Bの汚染項目",
	}
	require.NoError(t, db.WithContext(ctx).Create(fieldB).Error)

	all, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, all, 1)
	require.Len(t, all[0].Items, 1)
	assert.Equal(t, fieldA.ID, all[0].Items[0].ID)
	assert.NotEqual(t, fieldB.ID, all[0].Items[0].ID)

	one, err := repo.FindByID(ctx, clinicA, examType.ID)
	require.NoError(t, err)
	require.Len(t, one.Items, 1)
	assert.Equal(t, fieldA.ID, one.Items[0].ID)
	assert.Equal(t, clinicA, one.Items[0].ClinicID)
}

func TestExamTypeRepository_FindByID(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := &model.ExaminationType{ClinicID: clinicA, Name: "血液検査"}
	require.NoError(t, db.WithContext(ctx).Create(et).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, et.ID)
		require.NoError(t, err)
		assert.Equal(t, "血液検査", got.Name)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, et.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestExamTypeRepository_Create(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	et := &model.ExaminationType{ClinicID: clinicA, Name: "新規検査種別"}
	require.NoError(t, repo.Create(ctx, et))
	assert.NotZero(t, et.ID)

	got, err := repo.FindByID(ctx, clinicA, et.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規検査種別", got.Name)
}

func TestExamTypeRepository_Update(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := &model.ExaminationType{ClinicID: clinicA, Name: "旧名称"}
	require.NoError(t, db.WithContext(ctx).Create(et).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, et.ID, map[string]any{"name": "新名称"})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, et.ID, map[string]any{"name": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestExamTypeRepository_Delete(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := &model.ExaminationType{ClinicID: clinicA, Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(et).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, et.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, et.ID))

		_, err := repo.FindByID(ctx, clinicA, et.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, et.ID, x.ID)
		}

		var raw model.ExaminationType
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", et.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき")
	})
}

func TestExamTypeRepository_Reorder(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	e1 := &model.ExaminationType{ClinicID: clinicA, Name: "E1"}
	require.NoError(t, db.WithContext(ctx).Create(e1).Error)
	e2 := &model.ExaminationType{ClinicID: clinicA, Name: "E2"}
	require.NoError(t, db.WithContext(ctx).Create(e2).Error)
	e3 := &model.ExaminationType{ClinicID: clinicA, Name: "E3"}
	require.NoError(t, db.WithContext(ctx).Create(e3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{e3.ID, e1.ID, e2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, e3.ID, got[0].ID)
		assert.Equal(t, e1.ID, got[1].ID)
		assert.Equal(t, e2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.ExaminationType{ClinicID: clinicB, Name: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{e1.ID, other.ID})
		require.Error(t, err)
	})
}

func TestExamTypeRepository_CountUsageByExamTypeID(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	et := makeExamTypeMaster(t, db, clinicA, "使用中の検査種別")

	t.Run("使用実績が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByExamTypeID(ctx, clinicA, et.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	exam := makeExaminationRec(t, db, &model.Examination{ClinicID: clinicA, ExamTypeID: et.ID, Date: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)})

	t.Run("生存する exam が使用していれば 1", func(t *testing.T) {
		count, err := repo.CountUsageByExamTypeID(ctx, clinicA, et.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByExamTypeID(ctx, clinicB, et.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた exam は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(exam).Error)
		count, err := repo.CountUsageByExamTypeID(ctx, clinicA, et.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestExamTypeRepository_CountChildrenByParentID(t *testing.T) {
	db := setupExamTypeTestDB(t)
	repo := NewExamTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	parent := &model.ExaminationType{ClinicID: clinicA, Name: "親検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(parent).Error)

	t.Run("子が無ければ 0", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	child1 := &model.ExaminationType{ClinicID: clinicA, Name: "子1", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child1).Error)
	child2 := &model.ExaminationType{ClinicID: clinicA, Name: "子2", ParentID: &parent.ID}
	require.NoError(t, db.WithContext(ctx).Create(child2).Error)

	t.Run("子が2件あれば 2", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("別クリニックIDでは 0", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicB, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた子は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(child1).Error)
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}
