package repository

// reservation_type_unavailable_time_repository_test.go
// ReservationTypeUnavailableTimeRepository の CRUD・clinic_id 隔離・NotFound を実 DB で検証する。
// reservation_type_available_slot_repository_test.go と対をなす構造（同型インターフェース）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// setupReservationTypeUnavailableTimeTestDB は予約不可時間テスト用に DB を整備する。
// reservation_types CASCADE で reservation_type_unavailable_times も連鎖クリアされる（FK ON DELETE CASCADE）。
func setupReservationTypeUnavailableTimeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, ensureAutoMigrated(db, &model.ReservationType{}, &model.ReservationTypeUnavailableTime{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

func makeUnavailableTime(t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64, dayOfWeek int8, startTime, endTime string) *model.ReservationTypeUnavailableTime {
	t.Helper()
	dow := dayOfWeek
	ut := &model.ReservationTypeUnavailableTime{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		UnavailableType:   model.UnavailableTypeWeekly,
		DayOfWeek:         &dow,
		StartTime:         startTime,
		EndTime:           endTime,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(ut).Error)
	return ut
}

func TestReservationTypeUnavailableTimeRepository_FindAll(t *testing.T) {
	db := setupReservationTypeUnavailableTimeTestDB(t)
	repo := NewReservationTypeUnavailableTimeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "不可時間設定区分A", nil, nil)
	rtB := makeReservationTypeLinked(t, db, clinicB, "不可時間設定区分B", nil, nil)

	utA := makeUnavailableTime(t, db, clinicA, rtA.ID, 1, "12:00", "13:00")
	makeUnavailableTime(t, db, clinicB, rtB.ID, 1, "12:00", "13:00")

	t.Run("同一クリニックのみ取得できる", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, rtA.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, utA.ID, got[0].ID)
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB, rtA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestReservationTypeUnavailableTimeRepository_FindByID(t *testing.T) {
	db := setupReservationTypeUnavailableTimeTestDB(t)
	repo := NewReservationTypeUnavailableTimeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "不可時間単体取得区分", nil, nil)
	utA := makeUnavailableTime(t, db, clinicA, rtA.ID, 2, "09:00", "10:00")

	t.Run("同一クリニックIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, utA.ID)
		require.NoError(t, err)
		assert.Equal(t, utA.ID, got.ID)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, utA.ID)
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

func TestReservationTypeUnavailableTimeRepository_Create(t *testing.T) {
	db := setupReservationTypeUnavailableTimeTestDB(t)
	repo := NewReservationTypeUnavailableTimeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rtA := makeReservationTypeLinked(t, db, clinicA, "不可時間新規作成区分", nil, nil)
	dow := int8(3)
	ut := &model.ReservationTypeUnavailableTime{
		ClinicID:          clinicA,
		ReservationTypeID: rtA.ID,
		UnavailableType:   model.UnavailableTypeWeekly,
		DayOfWeek:         &dow,
		StartTime:         "15:00",
		EndTime:           "16:00",
	}
	require.NoError(t, repo.Create(ctx, ut))
	assert.NotZero(t, ut.ID)

	got, err := repo.FindByID(ctx, clinicA, ut.ID)
	require.NoError(t, err)
	assert.Equal(t, "15:00", got.StartTime)
	assert.Equal(t, "16:00", got.EndTime)
}

func TestReservationTypeUnavailableTimeRepository_Delete(t *testing.T) {
	db := setupReservationTypeUnavailableTimeTestDB(t)
	repo := NewReservationTypeUnavailableTimeRepository(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "不可時間削除対象区分", nil, nil)
	utA := makeUnavailableTime(t, db, clinicA, rtA.ID, 4, "17:00", "18:00")

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, utA.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正しいクリニックIDで削除できる（物理削除）", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, utA.ID))
		_, err := repo.FindByID(ctx, clinicA, utA.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
