package trimming

// repository_test.go — Repository の統合テスト（カバレッジ向上）。
//
// 移動元: trimming_course_repository_test.go（BE8-4 batch26）。makeReservationType/
// makeReservation はフラット package の同名ヘルパー（reservation_clinic_isolation_test.go）の
// 複製（BE8-4: import cycle を避けるための最小限の重複、移動時の型リネームはしない方針の対象外）。
//
// 対象: FindAll / FindByID / Create / Update / Delete / CountUsageByTrimmingCourseID / Reorder
// 検証観点: 正常系、clinic_id 隔離、ソフトデリート除外、NotFound ラップ。
//
// CountUsageByTrimmingCourseID は appointment_trimming_details を appointments に JOIN するため、
// setup で reservation_types / appointments (model.Reservation) も migrate する。

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

// setupTrimmingCourseTestDB は trimming_courses / appointment_trimming_details / appointments 周りを整備する。
func setupTrimmingCourseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	// TRUNCATE first: 他テストが残した orphan 行を除去してから AutoMigrate（FK 検証を通すため）。
	db.Exec("TRUNCATE TABLE appointment_trimming_options CASCADE")
	db.Exec("TRUNCATE TABLE appointment_trimming_details CASCADE")
	db.Exec("TRUNCATE TABLE trimming_courses CASCADE")
	db.Exec("TRUNCATE TABLE reservation_types CASCADE") // appointments も連鎖クリア
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.ReservationType{}, &model.Reservation{},
		&model.TrimmingCourse{}, &model.AppointmentTrimmingDetail{},
	))
	return db
}

// makeReservationType はテスト用予約区分を1件作成する（reservation_clinic_isolation_test.go の複製）。
func makeReservationType(t *testing.T, db *gorm.DB, clinicID uint64) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID,
		Name:     "テスト診療区分",
		Category: model.ReservationTypeCategoryGeneral,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

// makeReservation はテスト用予約を1件作成する（reservation_clinic_isolation_test.go の複製）。
func makeReservation(t *testing.T, db *gorm.DB, clinicID uint64) *model.Reservation {
	t.Helper()
	rt := makeReservationType(t, db, clinicID)
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         now,
		EndTime:           now.Add(15 * time.Minute),
		ReservationTypeID: rt.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    json.RawMessage(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

func TestTrimmingCourseRepository_FindAll_ClinicIsolationAndSortOrder(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	cB := &model.TrimmingCourse{ClinicID: clinicB, Name: "医院Bのコース", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(cB).Error)
	second := &model.TrimmingCourse{ClinicID: clinicA, Name: "Bコース", SortOrder: 2}
	require.NoError(t, db.WithContext(ctx).Create(second).Error)
	first := &model.TrimmingCourse{ClinicID: clinicA, Name: "Aコース", SortOrder: 1}
	require.NoError(t, db.WithContext(ctx).Create(first).Error)

	got, err := repo.FindAll(ctx, clinicA)
	require.NoError(t, err)
	require.Len(t, got, 2, "clinic A のコースのみ返る")
	assert.Equal(t, first.ID, got[0].ID, "sort_order 昇順で先頭に来る")
	assert.Equal(t, second.ID, got[1].ID)
	for _, x := range got {
		assert.NotEqual(t, cB.ID, x.ID, "別クリニックのコースが混入してはならない")
	}
}

func TestTrimmingCourseRepository_FindByID(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	price := int64(5000)
	c := &model.TrimmingCourse{ClinicID: clinicA, Name: "全身シャンプーコース", Price: &price}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("同一クリニックで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "全身シャンプーコース", got.Name)
		require.NotNil(t, got.Price)
		assert.Equal(t, price, *got.Price)
	})

	t.Run("別クリニックからは NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("存在しない ID は NotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingCourseRepository_Create(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	c := &model.TrimmingCourse{ClinicID: clinicA, Name: "新規コース"}
	require.NoError(t, repo.Create(ctx, c))
	assert.NotZero(t, c.ID)

	got, err := repo.FindByID(ctx, clinicA, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "新規コース", got.Name)
}

// BUG-455-S8: gorm default:true omits zero bools from INSERT.
func TestTrimmingCourseRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	c := &model.TrimmingCourse{ClinicID: clinicA, Name: "inactive course", IsActive: false}
	require.NoError(t, repo.Create(ctx, c))
	assert.False(t, c.IsActive)

	got, err := repo.FindByID(ctx, clinicA, c.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	var raw bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.TrimmingCourse{}).
		Select("is_active").
		Where("id = ?", c.ID).
		Scan(&raw).Error)
	assert.False(t, raw, "raw is_active must be false")
}

func TestTrimmingCourseRepository_Update(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := &model.TrimmingCourse{ClinicID: clinicA, Name: "旧名称"}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("同一クリニックで更新できる", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, c.ID, map[string]any{"name": "新名称"})
		require.NoError(t, err)
		assert.Equal(t, "新名称", got.Name)
	})

	t.Run("別クリニックからの更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, c.ID, map[string]any{"name": "乗っ取り"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の更新は NotFound", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicA, 999999, map[string]any{"name": "x"})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}

func TestTrimmingCourseRepository_Delete(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c := &model.TrimmingCourse{ClinicID: clinicA, Name: "削除対象"}
	require.NoError(t, db.WithContext(ctx).Create(c).Error)

	t.Run("別クリニックからの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, c.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しない ID の削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, 999999)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("同一クリニックで削除でき、ソフトデリートされる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, c.ID))

		_, err := repo.FindByID(ctx, clinicA, c.ID)
		assert.True(t, apperrors.IsNotFound(err))

		all, err := repo.FindAll(ctx, clinicA)
		require.NoError(t, err)
		for _, x := range all {
			assert.NotEqual(t, c.ID, x.ID)
		}

		var raw model.TrimmingCourse
		require.NoError(t, db.WithContext(ctx).Unscoped().Where("id = ?", c.ID).First(&raw).Error)
		assert.True(t, raw.DeletedAt.Valid, "deleted_at が設定されているべき（物理行は残る）")
	})
}

func TestTrimmingCourseRepository_CountUsageByTrimmingCourseID(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	course := &model.TrimmingCourse{ClinicID: clinicA, Name: "使用中のコース"}
	require.NoError(t, db.WithContext(ctx).Create(course).Error)

	t.Run("使用実績が無ければ 0", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingCourseID(ctx, clinicA, course.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("detail.clinic_id が親 appointment と異なる破損行を数えない", func(t *testing.T) {
		corruptAppt := makeReservation(t, db, clinicA)
		corruptDetail := &model.AppointmentTrimmingDetail{
			ClinicID:      clinicB,
			AppointmentID: corruptAppt.ID,
			CourseID:      &course.ID,
		}
		require.NoError(t, db.WithContext(ctx).Create(corruptDetail).Error)
		t.Cleanup(func() {
			require.NoError(t, db.WithContext(ctx).Delete(corruptDetail).Error)
		})

		count, err := repo.CountUsageByTrimmingCourseID(ctx, clinicA, course.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	// appointment_trimming_details.appointment_id は appointments(=Reservation) への FK のため実予約を作る。
	appt := makeReservation(t, db, clinicA)
	detail := &model.AppointmentTrimmingDetail{ClinicID: clinicA, AppointmentID: appt.ID, CourseID: &course.ID}
	require.NoError(t, db.WithContext(ctx).Create(detail).Error)

	t.Run("生存する appointment が使用していれば 1", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingCourseID(ctx, clinicA, course.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは 0（クロステナント越境なし）", func(t *testing.T) {
		count, err := repo.CountUsageByTrimmingCourseID(ctx, clinicB, course.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("ソフトデリートされた appointment は除外される（P2）", func(t *testing.T) {
		require.NoError(t, db.WithContext(ctx).Delete(appt).Error)
		count, err := repo.CountUsageByTrimmingCourseID(ctx, clinicA, course.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTrimmingCourseRepository_Reorder(t *testing.T) {
	db := setupTrimmingCourseTestDB(t)
	repo := NewTrimmingCourseRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	c1 := &model.TrimmingCourse{ClinicID: clinicA, Name: "C1"}
	require.NoError(t, db.WithContext(ctx).Create(c1).Error)
	c2 := &model.TrimmingCourse{ClinicID: clinicA, Name: "C2"}
	require.NoError(t, db.WithContext(ctx).Create(c2).Error)
	c3 := &model.TrimmingCourse{ClinicID: clinicA, Name: "C3"}
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
		other := &model.TrimmingCourse{ClinicID: clinicB, Name: "他院"}
		require.NoError(t, db.WithContext(ctx).Create(other).Error)
		err := repo.Reorder(ctx, clinicA, []uint64{c1.ID, other.ID})
		require.Error(t, err)
	})
}
