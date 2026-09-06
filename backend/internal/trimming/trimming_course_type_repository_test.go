package trimming

// trimming_course_type_repository_test.go — TrimmingCourseTypeRepository の統合テスト（カバレッジ向上）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / CountUsageByCourseTypeID / Reorder
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupTrimmingCourseTypeTestDB は trimming_course_types / trimming_courses を整備する。
// CountUsageByCourseTypeID は trimming_courses.course_type_id を直接参照するため両方を揃える。
func setupTrimmingCourseTypeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.TrimmingCourseType{}, &model.TrimmingCourse{}))
	db.Exec("TRUNCATE TABLE trimming_courses CASCADE")
	db.Exec("TRUNCATE TABLE trimming_course_types CASCADE")
	return db
}

func TestTrimmingCourseTypeRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ctB := &model.TrimmingCourseType{ClinicID: clinicB, Name: "医院Bの種別", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(ctB).Error)
	second := &model.TrimmingCourseType{ClinicID: clinicA, Name: "Bタイプ", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	first := &model.TrimmingCourseType{ClinicID: clinicA, Name: "Aタイプ", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A の種別のみ返る")
	assert.Equal(t, first.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)
	for _, x := range got {
		assert.NotEqual(t, ctB.ID, x.ID, "別クリニックの種別が混入してはならない")
	}
}

func TestTrimmingCourseTypeRepository_FindByID(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.TrimmingCourseType{ClinicID: clinicA, Name: "全身トリミング"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, "全身トリミング", got.Name)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, ct.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingCourseTypeRepository_Create(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	ct := &model.TrimmingCourseType{ClinicID: clinicA, Name: "新規種別"}
	created, err := repo.Create(ctx, ct)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)

	got, err := repo.FindByID(ctx, clinicA, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規種別", got.Name)
}

func TestTrimmingCourseTypeRepository_Update(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.TrimmingCourseType{ClinicID: clinicA, Name: "旧名称"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		name := "新名称"
		got, err := repo.Update(ctx, clinicA, ct.ID, UpdateTrimmingCourseTypeInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		name := "乗っ取り"
		_, err := repo.Update(ctx, clinicB, ct.ID, UpdateTrimmingCourseTypeInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		name := "x"
		_, err := repo.Update(ctx, clinicA, 999999, UpdateTrimmingCourseTypeInput{Name: &name})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingCourseTypeRepository_Delete(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.TrimmingCourseType{ClinicID: clinicA, Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, ct.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, ct.ID))

		_, err := repo.FindByID(ctx, clinicA, ct.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, ct.ID, x.ID)
		}

		var raw model.TrimmingCourseType
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", ct.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき（物理行は残る）")
	})

	t.Run("コースが参照していれば Conflict で行は残る", func(t *testing.T) {
		used := &model.TrimmingCourseType{ClinicID: clinicA, Name: "使用中削除対象"}
		require.NoError(t, db.WithContext(ctx).Create(used).Error)
		course := &model.TrimmingCourse{ClinicID: clinicA, Name: "参照コース", CourseTypeID: &used.ID}
		require.NoError(t, db.WithContext(ctx).Create(course).Error)

		err := repo.Delete(ctx, clinicA, used.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)

		got, findErr := repo.FindByID(ctx, clinicA, used.ID)
		require.NoError(t, findErr)
		assert.Equal(t, used.ID, got.ID)
	})
}

func TestTrimmingCourseTypeRepository_CountUsageByCourseTypeID(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	ct := &model.TrimmingCourseType{ClinicID: clinicA, Name: "使用中の種別"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)

	t.Run("使用実績が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByCourseTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	course := &model.TrimmingCourse{ClinicID: clinicA, Name: "全身コース", CourseTypeID: &ct.ID}
	require.NoError(t, db.WithContext(ctx).Create(course).Error)

	t.Run("生存する trimming_course が参照していれば 1", func(t *testing.T) {
		count, err := repo.CountUsageByCourseTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByCourseTypeID(ctx, clinicB, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた trimming_course は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(course).Error)
		count, err := repo.CountUsageByCourseTypeID(ctx, clinicA, ct.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTrimmingCourseTypeRepository_Reorder(t *testing.T) {
	db := setupTrimmingCourseTypeTestDB(t)
	repo := NewTrimmingCourseTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := &model.TrimmingCourseType{ClinicID: clinicA, Name: "C1"}
	require.NoError(t, db.WithContext(ctx).Create(c1).Error)
	c2 := &model.TrimmingCourseType{ClinicID: clinicA, Name: "C2"}
	require.NoError(t, db.WithContext(ctx).Create(c2).Error)
	c3 := &model.TrimmingCourseType{ClinicID: clinicA, Name: "C3"}
	require.NoError(t, db.WithContext(ctx).Create(c3).Error)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{c3.ID, c1.ID, c2.ID}))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 3)
		assert.Equal(t, c3.ID, got[0].ID)
		assert.Equal(t, c1.ID, got[1].ID)
		assert.Equal(t, c2.ID, got[2].ID)
	})

	t.Run("別クリニックの ID を含むと失敗する", func(t *testing.T) {
		other := &model.TrimmingCourseType{ClinicID: clinicB, Name: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err)
	})
}
