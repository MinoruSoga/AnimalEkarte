package repository

// reservation_clinic_isolation_test.go — #196 clinic_id テナント隔離回帰テスト
//
// テスト対象: ReservationRepository の clinic_id 境界
// 保護する不変条件: clinic A のスコープで clinic B の予約を
//   読み取れない、更新できない、削除できない。
//
// このテストは clinicScope を削除すると必ず失敗するよう設計されている。

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupReservationIsolationTestDB は予約 clinic_id 隔離テスト用の DB を整備する。
// setupTestDB が owners/billings 系をクリアし、ここで reservation_types/appointments を追加する。
// ekarte_db_test は本番 migration 適用済みのため fk_appointments_reservation_type FK 制約が存在する。
// TRUNCATE reservation_types CASCADE で appointments も連鎖クリアされる。
func setupReservationIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	// setupTestDB は per-call で ENUM を DROP しない。ensureAutoMigrated で reservation 系テーブルを揃える。
	require.NoError(t, ensureAutoMigrated(db, &model.ReservationType{}, &model.Reservation{}))
	// reservation_types CASCADE により fk_appointments_reservation_type 経由で
	// appointments も連鎖クリアされる。
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

// makeReservationType はテスト用予約区分を1件作成する。
// ekarte_db_test は本番 migration 適用済みのため reservation_types テーブルが存在する。
// clinic_id・name のみ DB default がない必須フィールド。category は ENUM のため明示的に設定する。
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

// makeReservation はテスト用予約を1件作成する。
// 先に makeReservationType で有効な区分を用意し、FK 制約を満たす。
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

// TestReservationRepository_FindByID_ClinicIsolation は
// clinic A の予約を clinic B の clinicID で取得できないことを検証する。
// clinicScope を reservation_repository.go から削除すると「別クリニックIDでは取得できない」が失敗する。
func TestReservationRepository_FindByID_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("同一クリニックIDでは取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, resA.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, resA.ID, got.ID)
		assert.Equal(t, clinicA, got.ClinicID)
	})

	t.Run("別クリニックIDでは取得できない（clinic_id 隔離）", func(t *testing.T) {
		// clinicScope が有効なら clinic B から clinic A の予約は見えない。
		got, err := repo.FindByID(ctx, clinicB, resA.ID)
		assert.Error(t, err, "clinic B から clinic A の予約を取得できてはならない")
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})
}

// TestReservationRepository_Update_ClinicIsolation は
// 別クリニックIDからの Update が NotFound を返し、行が変更されないことを検証する。
// clinicScope を削除すると「行が変更されていない」が失敗する。
func TestReservationRepository_Update_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("別クリニックIDからの Update は NotFound を返す", func(t *testing.T) {
		_, err := repo.Update(ctx, clinicB, resA.ID, map[string]any{"notes": "不正書き換え"})
		require.Error(t, err, "clinic B から clinic A の予約を更新できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("行が変更されていない（データ改ざん防止）", func(t *testing.T) {
		// プリロードを使わず直接 DB 読み取りで notes が変化していないことを確認する。
		var r model.Reservation
		require.NoError(t, db.Where("id = ? AND clinic_id = ?", resA.ID, clinicA).First(&r).Error)
		assert.Equal(t, "", r.Notes, "別クリニックからの Update で notes が変わってはならない")
	})

	t.Run("正しいクリニックIDからの Update は成功する", func(t *testing.T) {
		got, err := repo.Update(ctx, clinicA, resA.ID, map[string]any{"notes": "正常更新"})
		require.NoError(t, err)
		assert.Equal(t, "正常更新", got.Notes)
	})
}

// TestReservationRepository_Delete_ClinicIsolation は
// 別クリニックIDからの Delete が NotFound を返し、予約が削除されないことを検証する。
// clinicScope を削除すると「予約はまだ存在する」が失敗する。
func TestReservationRepository_Delete_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	resA := makeReservation(t, db, clinicA)

	t.Run("別クリニックIDからの Delete は NotFound を返す", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, resA.ID)
		require.Error(t, err, "clinic B から clinic A の予約を削除できてはならない")
		assert.True(t, apperrors.IsNotFound(err), "エラーは NotFound であるべき: %v", err)
	})

	t.Run("予約はまだ存在する（不正削除防止）", func(t *testing.T) {
		// プリロードを使わず直接 DB 読み取りで行が残存することを確認する。
		var r model.Reservation
		require.NoError(t, db.Where("id = ? AND clinic_id = ?", resA.ID, clinicA).First(&r).Error)
		assert.Equal(t, resA.ID, r.ID, "clinic A の予約は別クリニックの Delete で消えてはならない")
	})
}
