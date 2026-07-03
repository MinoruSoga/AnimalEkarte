package repository

// appointment_admin_repository_test.go — ReservationAdminRepository (管理者向け予約管理) の統合テスト。
//
// 対象: FindAllByMonth / FindAllByDay / Create / SoftDelete / FindAllByCustomerID /
//   CancelByID / FindByIDForNotify。
// 保護する不変条件: clinic_id 隔離（別クリニックの予約は見えない・更新/削除できない）と
//   ソフトデリート除外・NotFound ラップ。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupReservationAdminTestDB は ReservationAdminRepository 用に
// reservation_types/appointments に加えて Owner/Pet/Staff/LineCustomer preload 検証用の
// テーブルも整備する。
func setupReservationAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.ReservationType{}, &model.Reservation{},
		&model.AnimalSpecies{}, &model.Pet{}, &model.Staff{}, &model.LineCustomer{},
	))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE") // appointments も連鎖クリア
	db.Exec("TRUNCATE TABLE pets CASCADE")
	db.Exec("TRUNCATE TABLE animal_species CASCADE")
	db.Exec("TRUNCATE TABLE staffs CASCADE")
	db.Exec("TRUNCATE TABLE line_customers CASCADE")
	return db
}

// makeLineCustomerForAdmin はテスト用 LineCustomer を作成して返す。
func makeLineCustomerForAdmin(t *testing.T, db *gorm.DB, clinicID uint64, lineUserID string) *model.LineCustomer {
	t.Helper()
	lc := &model.LineCustomer{
		ClinicID:         clinicID,
		LineUserID:       lineUserID,
		AdditionalFields: []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(lc).Error)
	return lc
}

// makeAdminReservationAt は指定日時に紐づく予約を1件作成する。Owner/Pet/Doctor/LineCustomer は任意。
func makeAdminReservationAt(t *testing.T, db *gorm.DB, clinicID uint64, start time.Time, ownerID, petID, doctorID, lineCustomerID *uint64) *model.Reservation {
	t.Helper()
	rt := makeReservationType(t, db, clinicID)
	res := &model.Reservation{
		ClinicID:          clinicID,
		StartTime:         start,
		EndTime:           start.Add(30 * time.Minute),
		OwnerID:           ownerID,
		PetID:             petID,
		DoctorID:          doctorID,
		LineCustomerID:    lineCustomerID,
		ReservationTypeID: rt.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}
	require.NoError(t, db.WithContext(context.Background()).Create(res).Error)
	return res
}

func TestReservationAdminRepository_FindAllByMonth(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	inMonth := makeAdminReservationAt(t, db, clinicA, time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC), nil, nil, nil, nil)
	makeAdminReservationAt(t, db, clinicA, time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC), nil, nil, nil, nil)  // 対象月の外
	makeAdminReservationAt(t, db, clinicB, time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC), nil, nil, nil, nil) // 別クリニック

	t.Run("同一月・同一クリニックの予約のみ返す", func(t *testing.T) {
		got, err := repo.FindAllByMonth(ctx, clinicA, 2026, time.June)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, inMonth.ID, got[0].ID)
	})

	t.Run("該当月に予約が無ければ空スライスを返す", func(t *testing.T) {
		got, err := repo.FindAllByMonth(ctx, clinicA, 2025, time.January)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationAdminRepository_FindAllByDay(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()

	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "飼主・管理者予約")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "予約犬")
	doctor := makeDoctor(t, db, clinicA, "予約医師")

	targetDay := time.Date(2026, 6, 15, 3, 0, 0, 0, time.UTC) // JST 12:00 相当
	onDay := makeAdminReservationAt(t, db, clinicA, targetDay, &owner.ID, &pet.ID, &doctor.ID, nil)
	makeAdminReservationAt(t, db, clinicA, targetDay.AddDate(0, 0, 1), nil, nil, nil, nil) // 翌日
	makeAdminReservationAt(t, db, clinicB, targetDay, nil, nil, nil, nil)                  // 別クリニック

	t.Run("同一日・同一クリニックの予約のみ返し、Owner/Pet/Doctorをpreloadする", func(t *testing.T) {
		got, err := repo.FindAllByDay(ctx, clinicA, targetDay)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, onDay.ID, got[0].ID)
		require.NotNil(t, got[0].Owner)
		assert.Equal(t, owner.ID, got[0].Owner.ID)
		require.NotNil(t, got[0].Pet)
		assert.Equal(t, pet.ID, got[0].Pet.ID)
		require.NotNil(t, got[0].Doctor)
		assert.Equal(t, doctor.ID, got[0].Doctor.ID)
	})

	t.Run("該当日に予約が無ければ空スライスを返す", func(t *testing.T) {
		got, err := repo.FindAllByDay(ctx, clinicA, targetDay.AddDate(0, 1, 0))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationAdminRepository_Create(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rt := makeReservationType(t, db, clinicA)
	now := time.Now().UTC().Truncate(time.Minute)
	res := &model.Reservation{
		ClinicID:          clinicA,
		StartTime:         now,
		EndTime:           now.Add(30 * time.Minute),
		ReservationTypeID: rt.ID,
		VisitType:         model.VisitTypeRevisit,
		Status:            model.ReservationStatusPending,
		Source:            model.ReservationSourceManual,
		CustomerFields:    []byte(`{}`),
	}

	require.NoError(t, repo.Create(ctx, res))
	assert.NotZero(t, res.ID)

	var stored model.Reservation
	require.NoError(t, db.First(&stored, res.ID).Error)
	assert.Equal(t, clinicA, stored.ClinicID)
}

func TestReservationAdminRepository_SoftDelete(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("同一クリニックの予約はソフトデリートされる", func(t *testing.T) {
		res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), nil, nil, nil, nil)
		require.NoError(t, repo.SoftDelete(ctx, clinicA, res.ID))

		var normal model.Reservation
		err := db.First(&normal, res.ID).Error
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "ソフトデリート後は通常クエリで見えなくなるべき")

		var withDeleted model.Reservation
		require.NoError(t, db.Unscoped().First(&withDeleted, res.ID).Error)
		assert.True(t, withDeleted.DeletedAt.Valid, "deleted_at が設定されるべき")
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		err := repo.SoftDelete(ctx, clinicA, uint64(9999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックの予約は削除できない（clinic_id隔離）", func(t *testing.T) {
		res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), nil, nil, nil, nil)
		err := repo.SoftDelete(ctx, clinicB, res.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stillThere model.Reservation
		require.NoError(t, db.First(&stillThere, res.ID).Error, "別クリニックからの削除で消えてはならない")
	})
}

func TestReservationAdminRepository_FindAllByCustomerID(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	customer := makeLineCustomerForAdmin(t, db, clinicA, "U-admin-001")
	older := makeAdminReservationAt(t, db, clinicA, time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC), nil, nil, nil, &customer.ID)
	newer := makeAdminReservationAt(t, db, clinicA, time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC), nil, nil, nil, &customer.ID)
	makeAdminReservationAt(t, db, clinicA, time.Date(2026, 6, 5, 9, 0, 0, 0, time.UTC), nil, nil, nil, nil)          // 別顧客(nil)
	makeAdminReservationAt(t, db, clinicB, time.Date(2026, 6, 5, 9, 0, 0, 0, time.UTC), nil, nil, nil, &customer.ID) // 別クリニック

	t.Run("同一クリニック・同一顧客の予約のみ新しい順で返す", func(t *testing.T) {
		got, err := repo.FindAllByCustomerID(ctx, clinicA, customer.ID)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, newer.ID, got[0].ID)
		assert.Equal(t, older.ID, got[1].ID)
	})

	t.Run("該当顧客が無ければ空スライスを返す", func(t *testing.T) {
		got, err := repo.FindAllByCustomerID(ctx, clinicA, uint64(9999999))
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationAdminRepository_CancelByID(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	t.Run("正しい顧客・クリニックからのキャンセルは成功しソフトデリートされる", func(t *testing.T) {
		customer := makeLineCustomerForAdmin(t, db, clinicA, "U-cancel-001")
		res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), nil, nil, nil, &customer.ID)

		require.NoError(t, repo.CancelByID(ctx, clinicA, customer.ID, res.ID))

		var withDeleted model.Reservation
		require.NoError(t, db.Unscoped().First(&withDeleted, res.ID).Error)
		assert.True(t, withDeleted.DeletedAt.Valid)
		assert.Equal(t, model.ReservationStatusCancelled, withDeleted.Status)
	})

	t.Run("存在しない予約IDはNotFoundを返す", func(t *testing.T) {
		customer := makeLineCustomerForAdmin(t, db, clinicA, "U-cancel-002")
		err := repo.CancelByID(ctx, clinicA, customer.ID, uint64(9999999))
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("既にキャンセル済みの予約は再キャンセルできない", func(t *testing.T) {
		customer := makeLineCustomerForAdmin(t, db, clinicA, "U-cancel-003")
		res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), nil, nil, nil, &customer.ID)
		require.NoError(t, repo.CancelByID(ctx, clinicA, customer.ID, res.ID))

		err := repo.CancelByID(ctx, clinicA, customer.ID, res.ID)
		require.Error(t, err, "status != cancelled 条件のため2回目はNotFoundになるべき")
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックからのキャンセルは失敗する（clinic_id隔離）", func(t *testing.T) {
		customer := makeLineCustomerForAdmin(t, db, clinicA, "U-cancel-004")
		res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), nil, nil, nil, &customer.ID)

		err := repo.CancelByID(ctx, clinicB, customer.ID, res.ID)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))

		var stillActive model.Reservation
		require.NoError(t, db.First(&stillActive, res.ID).Error)
		assert.Equal(t, model.ReservationStatusPending, stillActive.Status)
	})
}

func TestReservationAdminRepository_FindByIDForNotify(t *testing.T) {
	db := setupReservationAdminTestDB(t)
	repo := NewReservationAdminRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	owner := makeOwner(t, db, clinicA, "飼主・通知")
	pet := makeSpeciesAndPet(t, db, clinicA, owner.ID, "通知犬")
	doctor := makeDoctor(t, db, clinicA, "通知医師")
	res := makeAdminReservationAt(t, db, clinicA, time.Now().UTC(), &owner.ID, &pet.ID, &doctor.ID, nil)

	t.Run("同一クリニックでは関連エンティティ込みで取得できる", func(t *testing.T) {
		got, err := repo.FindByIDForNotify(ctx, clinicA, res.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Owner)
		assert.Equal(t, owner.ID, got.Owner.ID)
		require.NotNil(t, got.Pet)
		assert.Equal(t, pet.ID, got.Pet.ID)
		require.NotNil(t, got.Doctor)
		assert.Equal(t, doctor.ID, got.Doctor.ID)
		require.NotNil(t, got.ReservationType)
	})

	t.Run("存在しないIDはNotFoundを返す", func(t *testing.T) {
		got, err := repo.FindByIDForNotify(ctx, clinicA, uint64(9999999))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("別クリニックIDでは取得できない（clinic_id隔離）", func(t *testing.T) {
		got, err := repo.FindByIDForNotify(ctx, clinicB, res.ID)
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
