package reservation

// reservation_type_liff_repository_test.go
// ReservationTypeLiffRepository（reservation_types への LIFF 予約コース用ラッパー）の
// CRUD・隣接スワップ並び替え（UpdateSortOrder）を実 DB で検証する。

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

// setupReservationTypeLiffTestDB は ReservationTypeLiffRepository のテスト用に DB を整備する。
// setupTestDB が reservation_type_category ENUM を DROP CASCADE するため reservation_types.category 列が
// 消える。AutoMigrate で列・テーブルを再整備してから TRUNCATE でクリーンな状態にする。
func setupReservationTypeLiffTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationTypeGroup{}, &model.ReservationType{}, &model.Reservation{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	db.Exec("TRUNCATE TABLE reservation_type_groups CASCADE")
	return db
}

// makeLiffTypeWithSortOrder は sort_order を明示指定した予約コースを1件作成する。
// UpdateSortOrder のスワップ挙動検証には既定値0が全レコード共通だと隣接判定できないため、
// 明示的な sort_order 指定が必要。
func makeLiffTypeWithSortOrder(t *testing.T, db *gorm.DB, clinicID uint64, name string, sortOrder int) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID:  clinicID,
		Name:      name,
		Category:  model.ReservationTypeCategoryGeneral,
		SortOrder: sortOrder,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

// BUG-455-S6: LIFF create path must persist explicit reservation_visible=false.
func TestReservationTypeLiffRepository_Create_ReservationVisibleFalsePersists(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicID = uint64(1)

	st := &model.ReservationType{
		ClinicID:           clinicID,
		Name:               "liff hidden course",
		Category:           model.ReservationTypeCategoryGeneral,
		IsActive:           true,
		ReservationVisible: false,
	}
	require.NoError(t, repo.Create(ctx, st))
	assert.False(t, st.ReservationVisible)

	got, err := repo.FindByID(ctx, clinicID, st.ID)
	require.NoError(t, err)
	assert.False(t, got.ReservationVisible)

	var rawVisible bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationType{}).
		Select("reservation_visible").
		Where("id = ?", st.ID).
		Scan(&rawVisible).Error)
	assert.False(t, rawVisible, "raw reservation_visible must be false")
}

func TestReservationTypeLiffRepository_FindAll(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt2 := makeLiffTypeWithSortOrder(t, db, clinicA, "コース2", 2)
	rt1 := makeLiffTypeWithSortOrder(t, db, clinicA, "コース1", 1)
	_ = makeLiffTypeWithSortOrder(t, db, clinicB, "医院Bコース", 1)

	t.Run("同一クリニックのみ sort_order 昇順で取得できる", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, rt1.ID, got[0].ID)
		assert.Equal(t, rt2.ID, got[1].ID)
	})

	t.Run("別クリニックIDでは見えない", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.NotEqual(t, rt1.ID, got[0].ID)
		assert.NotEqual(t, rt2.ID, got[0].ID)
	})
}

func TestReservationTypeLiffRepository_FindByID(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt := makeReservationTypeLinked(t, db, clinicA, "単体取得コース", nil, nil)

	t.Run("同一クリニックIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, rt.ID)
		require.NoError(t, err)
		assert.Equal(t, rt.ID, got.ID)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, rt.ID)
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeLiffRepository_CountChildrenByParentID(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	parent := makeReservationTypeLinked(t, db, clinicA, "親コース", nil, nil)

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

func TestReservationTypeLiffRepository_Create(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()

	rt := &model.ReservationType{
		ClinicID: 1,
		Name:     "新規作成コース",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, repo.Create(ctx, rt))
	assert.NotZero(t, rt.ID)

	got, err := repo.FindByID(ctx, 1, rt.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規作成コース", got.Name)
}

func TestReservationTypeLiffRepository_Update(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt := makeReservationTypeLinked(t, db, clinicA, "更新前コース", nil, nil)

	t.Run("正しいクリニックIDで更新できる", func(t *testing.T) {
		name := "更新後コース"
		updated, err := repo.Update(ctx, clinicA, rt.ID, UpdateReservationTypeLiffInput{Name: &name})
		require.NoError(t, err)
		assert.Equal(t, "更新後コース", updated.Name)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		name := "不正更新"
		updated, err := repo.Update(ctx, clinicB, rt.ID, UpdateReservationTypeLiffInput{Name: &name})
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestReservationTypeLiffRepository_Delete(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rt := makeReservationTypeLinked(t, db, clinicA, "削除対象コース", nil, nil)

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
		typeRepo := NewReservationTypeRepository(db)
		target := makeReservationTypeLinked(t, db, clinicA, "TOCTOU LIFFコース", nil, nil)
		count, err := typeRepo.CountUsageByReservationTypeID(ctx, clinicA, target.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		makeUsageAppointment(t, db, clinicA, target.ID, false)

		err = repo.Delete(ctx, clinicA, target.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この予約コースは予約データで使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, target.ID)
		require.NoError(t, findErr)
		assert.Equal(t, target.ID, got.ID)
	})

	t.Run("DeleteWithDependencyChecks: CountUsage==0 直後の参照追加は Conflict で行が残る", func(t *testing.T) {
		typeRepo := NewReservationTypeRepository(db)
		target := makeReservationTypeLinked(t, db, clinicA, "TOCTOU LIFF依存チェック", nil, nil)
		count, err := typeRepo.CountUsageByReservationTypeID(ctx, clinicA, target.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)

		makeUsageAppointment(t, db, clinicA, target.ID, false)

		usage := liffUsageExistsByType{repo: typeRepo}
		err = repo.DeleteWithDependencyChecks(ctx, clinicA, target.ID, usage)
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err), "expected Conflict, got %v", err)
		assert.Contains(t, err.Error(), "この予約コースは予約データで使用中のため削除できません")

		got, findErr := repo.FindByID(ctx, clinicA, target.ID)
		require.NoError(t, findErr)
		assert.Equal(t, target.ID, got.ID)
	})
}

type liffUsageExistsByType struct {
	repo ReservationTypeRepository
}

func (u liffUsageExistsByType) ExistsByReservationTypeID(ctx context.Context, clinicID, reservationTypeID uint64) (bool, error) {
	count, err := u.repo.CountUsageByReservationTypeID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func TestReservationTypeLiffRepository_UpdateSortOrder(t *testing.T) {
	db := setupReservationTypeLiffTestDB(t)
	repo := NewReservationTypeLiffRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("down方向で隣接レコードとsort_orderがスワップされる", func(t *testing.T) {
		rt1 := makeLiffTypeWithSortOrder(t, db, clinicA, "down1", 10)
		rt2 := makeLiffTypeWithSortOrder(t, db, clinicA, "down2", 20)

		require.NoError(t, repo.UpdateSortOrder(ctx, clinicA, rt1.ID, "down"))

		got1, err := repo.FindByID(ctx, clinicA, rt1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, rt2.ID)
		require.NoError(t, err)
		assert.Equal(t, 20, got1.SortOrder, "rt1 は隣接する大きい方の sort_order を得るべき")
		assert.Equal(t, 10, got2.SortOrder, "rt2 は元の rt1 の sort_order を得るべき")
	})

	t.Run("up方向で隣接レコードとsort_orderがスワップされる", func(t *testing.T) {
		rt1 := makeLiffTypeWithSortOrder(t, db, clinicA, "up1", 30)
		rt2 := makeLiffTypeWithSortOrder(t, db, clinicA, "up2", 40)

		require.NoError(t, repo.UpdateSortOrder(ctx, clinicA, rt2.ID, "up"))

		got1, err := repo.FindByID(ctx, clinicA, rt1.ID)
		require.NoError(t, err)
		got2, err := repo.FindByID(ctx, clinicA, rt2.ID)
		require.NoError(t, err)
		assert.Equal(t, 40, got1.SortOrder)
		assert.Equal(t, 30, got2.SortOrder)
	})

	t.Run("隣接レコードが無い場合は変更なし（no-op）", func(t *testing.T) {
		// UpdateSortOrder の "adjacent" 判定は sort_order の大小関係のみで決まる（隣接クリニック内で
		// target.SortOrder より小さい最大値を探す）。この t.Run は同一 db/repo を共有する親テスト内で
		// 前の t.Run（down1=10→20, down2=20→10, up1=30→40, up2=40→30 のスワップ後）が既に
		// clinicA に sort_order={10,20,30,40} の行を残しているため、同じ clinicA を使うと
		// sort_order=100 の "up" は 40 を隣接として誤ってヒットしてしまう。
		// 真に隣接なしを検証するため、他の t.Run が触れない専用クリニックIDを使う。
		const soloClinic = uint64(999)
		solo := makeLiffTypeWithSortOrder(t, db, soloClinic, "隣接なし", 100)

		require.NoError(t, repo.UpdateSortOrder(ctx, soloClinic, solo.ID, "up"))

		got, err := repo.FindByID(ctx, soloClinic, solo.ID)
		require.NoError(t, err)
		assert.Equal(t, 100, got.SortOrder, "隣接レコードが無ければ sort_order は変わらないべき")
	})

	t.Run("別クリニックIDの対象は NotFound", func(t *testing.T) {
		rt := makeLiffTypeWithSortOrder(t, db, clinicA, "越境対象", 200)
		err := repo.UpdateSortOrder(ctx, clinicB, rt.ID, "down")
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "クリニック不一致は NotFound であるべき: %v", err)
	})
}
