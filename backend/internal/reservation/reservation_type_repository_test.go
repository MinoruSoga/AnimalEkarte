package reservation

// reservation_type_repository_test.go
// ReservationTypeRepository の CRUD・カウント・並び替えメソッドを実 DB で検証する。
// Group/Parent/Children の Preload clinic_id 隔離は reservation_type_preload_clinic_isolation_test.go が
// 別途カバーしているため、本ファイルは CRUD/Count/Reorder/NotFound/clinic_id 分離に焦点を当てる。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/testdb"
)

// setupReservationTypeRepoTestDB は ReservationTypeRepository の CRUD テスト用に DB を整備する。
// reservation_types CASCADE で group / appointments (FK 経由) も連鎖クリアされる。
func setupReservationTypeRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationTypeGroup{}, &model.ReservationType{}, &model.Reservation{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	return db
}

// makeUsageAppointment は CountUsageByReservationTypeID 検証用の appointments 行を1件作成する。
func makeUsageAppointment(t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64, deleted bool) *model.Reservation {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: reservationTypeID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	if deleted {
		require.NoError(t, db.WithContext(context.Background()).Delete(res).Error)
	}
	return res
}

func TestReservationTypeRepository_FindAll(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "A区分", nil, nil)
	_ = makeReservationTypeLinked(t, db, clinicB, "B区分", nil, nil)

	t.Run("同一クリニックのみ取得できる", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, rtA.ID, got[0].ID)
	})

	t.Run("別クリニックIDでは見えない", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.NotEqual(t, rtA.ID, got[0].ID)
	})

	t.Run("ソフトデリート済みは除外される", func(t *testing.T) {
		deleted := makeReservationTypeLinked(t, db, clinicA, "削除予定区分", nil, nil)
		require.NoError(t, repo.Delete(ctx, clinicA, deleted.ID))

		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, rt := range got {
			assert.NotEqual(t, deleted.ID, rt.ID, "ソフトデリート済みの区分が含まれてはならない")
		}
	})
}

func TestReservationTypeRepository_FindAllWithChildren(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	parent := makeReservationTypeLinked(t, db, clinicA, "親区分", nil, nil)
	child := makeReservationTypeLinked(t, db, clinicA, "子区分", nil, &parent.ID)

	got, err := repo.FindAllWithChildren(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 1, "parent_id IS NULL のトップレベルのみ返るべき")
	assert.Equal(t, parent.ID, got[0].ID)
	require.Len(t, got[0].Children, 1)
	assert.Equal(t, child.ID, got[0].Children[0].ID)
}

func TestReservationTypeRepository_FindByID_NotFound(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()

	got, err := repo.FindByID(ctx, uint64(1), uint64(999999))
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err), "存在しないIDは NotFound であるべき: %v", err)
}

func TestReservationTypeRepository_FindByIDWithChildren_NotFound(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()

	got, err := repo.FindByIDWithChildren(ctx, uint64(1), uint64(999999))
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsNotFound(err))
}

// BUG-455-S6: gorm default:true omits zero bools from INSERT.
func TestReservationTypeRepository_Create_BoolDefaultsFalsePersist(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	rt := &model.ReservationType{
		ClinicID:           clinicID,
		Name:               "inactive invisible type",
		Category:           model.ReservationTypeCategoryGeneral,
		IsActive:           false,
		ReservationVisible: false,
	}
	require.NoError(t, repo.Create(ctx, rt))
	require.NotZero(t, rt.ID)
	assert.False(t, rt.IsActive)
	assert.False(t, rt.ReservationVisible)

	got, err := repo.FindByID(ctx, clinicID, rt.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)
	assert.False(t, got.ReservationVisible)

	var rawActive, rawVisible bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Select("is_active").
		Where("id = ?", rt.ID).
		Scan(&rawActive).Error)
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Select("reservation_visible").
		Where("id = ?", rt.ID).
		Scan(&rawVisible).Error)
	assert.False(t, rawActive, "raw is_active must be false")
	assert.False(t, rawVisible, "raw reservation_visible must be false")
}

func TestReservationTypeRepository_CountUsageByReservationTypeID(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "使用状況確認用区分", nil, nil)

	t.Run("未使用は0件", func(t *testing.T) {
		count, err := repo.CountUsageByReservationTypeID(ctx, clinicA, rtA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	makeUsageAppointment(t, db, clinicA, rtA.ID, false)
	makeUsageAppointment(t, db, clinicA, rtA.ID, true) // 削除済みは含まれない

	t.Run("有効な予約のみカウントされる（ソフトデリート除外）", func(t *testing.T) {
		count, err := repo.CountUsageByReservationTypeID(ctx, clinicA, rtA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		count, err := repo.CountUsageByReservationTypeID(ctx, clinicB, rtA.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestReservationTypeRepository_CountChildrenByParentID(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	parent := makeReservationTypeLinked(t, db, clinicA, "親区分カウント用", nil, nil)

	t.Run("子なしは0件", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	child1 := makeReservationTypeLinked(t, db, clinicA, "子1", nil, &parent.ID)
	_ = makeReservationTypeLinked(t, db, clinicA, "子2", nil, &parent.ID)

	t.Run("子2件をカウントする", func(t *testing.T) {
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("ソフトデリート済みの子は除外される", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, child1.ID))
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestReservationTypeRepository_Create(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()

	rt := &model.ReservationType{
		ClinicID: 1,
		Name:     "新規作成区分",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, repo.Create(ctx, rt))
	assert.NotZero(t, rt.ID)

	got, err := repo.FindByID(ctx, 1, rt.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成区分", got.Name)
}

func TestReservationTypeRepository_Update(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt := makeReservationTypeLinked(t, db, clinicA, "更新前区分", nil, nil)

	t.Run("正しいクリニックIDで更新できる", func(t *testing.T) {
		updated, err := repo.Update(ctx, clinicA, rt.ID, map[string]any{"name": "更新後区分"})
		require.NoError(t, err)
		assert.Equal(t, "更新後区分", updated.Name)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		updated, err := repo.Update(ctx, clinicB, rt.ID, map[string]any{"name": "不正更新"})
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeRepository_Delete(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt := makeReservationTypeLinked(t, db, clinicA, "削除対象区分", nil, nil)

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, rt.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正しいクリニックIDで削除できる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, rt.ID))
		_, err := repo.FindByID(ctx, clinicA, rt.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("既に削除済みの再削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, rt.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("CountUsage==0 直後に予約が紐づいても削除は失敗する", func(t *testing.T) {
		target := makeReservationTypeLinked(t, db, clinicA, "TOCTOU使用区分", nil, nil)
		count, err := repo.CountUsageByReservationTypeID(ctx, clinicA, target.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		makeUsageAppointment(t, db, clinicA, target.ID, false)

		err = repo.Delete(ctx, clinicA, target.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この項目は予約データで使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, target.ID)
		require.NoError(t, findErr)
		assert.Equal(t, target.ID, got.ID)
	})

	t.Run("CountChildren==0 直後に子区分が付いても削除は失敗する", func(t *testing.T) {
		parent := makeReservationTypeLinked(t, db, clinicA, "TOCTOU親区分", nil, nil)
		count, err := repo.CountChildrenByParentID(ctx, clinicA, parent.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		_ = makeReservationTypeLinked(t, db, clinicA, "TOCTOU子区分", nil, &parent.ID)

		err = repo.Delete(ctx, clinicA, parent.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この予約区分には子予約区分が登録されているため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, parent.ID)
		require.NoError(t, findErr)
		assert.Equal(t, parent.ID, got.ID)
	})
}

func TestReservationTypeRepository_Reorder(t *testing.T) {
	db := setupReservationTypeRepoTestDB(t)
	repo := NewReservationTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt1 := makeReservationTypeLinked(t, db, clinicA, "並び1", nil, nil)
	rt2 := makeReservationTypeLinked(t, db, clinicA, "並び2", nil, nil)
	rt3 := makeReservationTypeLinked(t, db, clinicA, "並び3", nil, nil)

	t.Run("指定順に sort_order が振り直される", func(t *testing.T) {
		require.NoError(t, repo.Reorder(ctx, clinicA, []uint64{rt3.ID, rt1.ID, rt2.ID}))

		got3, err := repo.FindByID(ctx, clinicA, rt3.ID)
		require.NoError(t, err)
		got1, err := repo.FindByID(ctx, clinicA, rt1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, rt2.ID)
		require.NoError(t, err)

		assert.Equal(t, 1, got3.SortOrder)
		assert.Equal(t, 2, got1.SortOrder)
		assert.Equal(t, 3, got2.SortOrder)
	})

	t.Run("他クリニックの id を含むと失敗する", func(t *testing.T) {
		otherClinicRT := makeReservationTypeLinked(t, db, uint64(2), "別医院区分", nil, nil)
		err := repo.Reorder(ctx, clinicA, []uint64{rt1.ID, otherClinicRT.ID})
		assert.Error(t, err, "clinic A のスコープに存在しない id を含む Reorder は失敗するべき")
	})
}
