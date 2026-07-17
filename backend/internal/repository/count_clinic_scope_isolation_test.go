package repository

// count_clinic_scope_isolation_test.go — BE-refactor.md R2-5 (D12) の回帰防止。
//
// CountEstimatesByMedicalRecordID / CountMedicalRecordsByReservationID / CountItemsByEstimateID は
// これまで clinic_id 述語なしで id のみをキーに集計していた。呼び出し元（各 Delete）は
// FindByID(ctx, clinicID, id) で id の所有権を事前検証済みのため実害は低いが、
// 「repository の読取は clinic スコープ必須」の規約に無追跡の例外として残っていた。
// 本テストは他クリニックの子データが誤って集計に混入しないことを固定する
// （clinic_id 述語を削除すると必ず失敗するよう設計）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

// setupEstimateIsolationTestDB は setupTestDB に加え estimates/estimate_items を AutoMigrate する。
// setupTestDB の共有 AutoMigrate リストに Estimate/EstimateItem が含まれていないため
// （estimate_status/item_category ENUM は既に作成済み）、このテストファイル内で追加する。
func setupEstimateIsolationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.Estimate{}, &model.EstimateItem{}))
	db.Exec("TRUNCATE TABLE estimates CASCADE")
	return db
}

func TestMedicalRecordRepository_CountEstimatesByMedicalRecordID_ClinicIsolation(t *testing.T) {
	db := setupEstimateIsolationTestDB(t)
	repo := NewMedicalRecordRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-A-1", Date: time.Now()}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	est := &model.Estimate{ClinicID: clinicA, MedicalRecordID: &mr.ID}
	require.NoError(t, db.WithContext(ctx).Create(est).Error)

	t.Run("同一クリニックIDでは件数が見える", func(t *testing.T) {
		count, err := repo.CountEstimatesByMedicalRecordID(ctx, clinicA, mr.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件を返す（漏洩しない）", func(t *testing.T) {
		count, err := repo.CountEstimatesByMedicalRecordID(ctx, clinicB, mr.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count, "clinic_id 述語が無いと過去汚染データで別クリニックの件数が混入しうる")
	})
}

func TestReservationRepository_CountMedicalRecordsByReservationID_ClinicIsolation(t *testing.T) {
	db := setupReservationIsolationTestDB(t)
	repo := NewReservationRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	res := makeReservation(t, db, clinicA)
	mr := &model.MedicalRecord{ClinicID: clinicA, RecordNo: "MR-A-2", Date: time.Now(), AppointmentID: &res.ID}
	require.NoError(t, db.WithContext(ctx).Create(mr).Error)

	t.Run("同一クリニックIDでは件数が見える", func(t *testing.T) {
		count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicA, res.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件を返す（漏洩しない）", func(t *testing.T) {
		count, err := repo.CountMedicalRecordsByReservationID(ctx, clinicB, res.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)
	})
}

func TestEstimateRepository_CountItemsByEstimateID_ClinicIsolation(t *testing.T) {
	db := setupEstimateIsolationTestDB(t)
	repo := NewEstimateRepository(db)
	ctx := context.Background()

	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	est := &model.Estimate{ClinicID: clinicA}
	require.NoError(t, db.WithContext(ctx).Create(est).Error)
	item := &model.EstimateItem{EstimateID: est.ID, Name: "テスト項目", Category: model.ItemCategoryOther}
	require.NoError(t, db.WithContext(ctx).Create(item).Error)

	t.Run("同一クリニックIDでは件数が見える", func(t *testing.T) {
		count, err := repo.CountItemsByEstimateID(ctx, clinicA, est.ID)
		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})

	t.Run("別クリニックIDでは0件を返す（JOIN述語がないと漏洩しうる）", func(t *testing.T) {
		count, err := repo.CountItemsByEstimateID(ctx, clinicB, est.ID)
		require.NoError(t, err)
		require.Equal(t, int64(0), count)
	})
}
