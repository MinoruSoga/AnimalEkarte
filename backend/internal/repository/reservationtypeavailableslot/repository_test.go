package reservationtypeavailableslot

// repository_test.go — Repository の CRUD・clinic_id 隔離・NotFound を実 DB で検証する。
// makeReservationTypeLinked はフラット package の同名ヘルパーの複製（BE8-4 batch2）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repotest"
)

// setupTestDB は予約可能開始時刻テスト用に DB を整備する。
// reservation_types CASCADE で reservation_type_available_slots も連鎖クリアされる（FK ON DELETE CASCADE）。
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := repotest.SetupTestDB(t)
	require.NoError(t, repotest.EnsureAutoMigrated(db, &model.ReservationType{}, &model.ReservationTypeAvailableSlot{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
}

func makeReservationTypeLinked(t *testing.T, db *gorm.DB, clinicID uint64, name string, groupID, parentID *uint64) *model.ReservationType {
	t.Helper()
	rt := &model.ReservationType{
		ClinicID: clinicID,
		Name:     name,
		Category: model.ReservationTypeCategoryGeneral,
		GroupID:  groupID,
		ParentID: parentID,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(rt).Error)
	return rt
}

func makeAvailableSlot(t *testing.T, db *gorm.DB, clinicID, reservationTypeID uint64, dayOfWeek int8, startTime string) *model.ReservationTypeAvailableSlot {
	t.Helper()
	dow := dayOfWeek
	slot := &model.ReservationTypeAvailableSlot{
		ClinicID:          clinicID,
		ReservationTypeID: reservationTypeID,
		AvailableType:     model.AvailableSlotTypeWeekly,
		DayOfWeek:         &dow,
		StartTime:         startTime,
		IsActive:          true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(slot).Error)
	return slot
}

func TestRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "枠設定区分A", nil, nil)
	rtB := makeReservationTypeLinked(t, db, clinicB, "枠設定区分B", nil, nil)

	slotA := makeAvailableSlot(t, db, clinicA, rtA.ID, 1, "09:00")
	makeAvailableSlot(t, db, clinicB, rtB.ID, 1, "09:00")

	t.Run("同一クリニックのみ取得できる", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicA, rtA.ID)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, slotA.ID, got[0].ID)
	})

	t.Run("別クリニックIDでは0件（clinic_id 隔離）", func(t *testing.T) {
		got, err := repo.FindAll(ctx, clinicB, rtA.ID)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "枠単体取得区分", nil, nil)
	slotA := makeAvailableSlot(t, db, clinicA, rtA.ID, 2, "10:00")

	t.Run("同一クリニックIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicA, slotA.ID)
		require.NoError(t, err)
		assert.Equal(t, slotA.ID, got.ID)
	})

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, clinicB, slotA.ID)
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

func TestRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rtA := makeReservationTypeLinked(t, db, clinicA, "枠新規作成区分", nil, nil)
	dow := int8(3)
	slot := &model.ReservationTypeAvailableSlot{
		ClinicID:          clinicA,
		ReservationTypeID: rtA.ID,
		AvailableType:     model.AvailableSlotTypeWeekly,
		DayOfWeek:         &dow,
		StartTime:         "13:30",
		IsActive:          true,
	}
	require.NoError(t, repo.Create(ctx, slot))
	assert.NotZero(t, slot.ID)

	got, err := repo.FindByID(ctx, clinicA, slot.ID)
	require.NoError(t, err)
	assert.Equal(t, "13:30", got.StartTime)
}

func TestRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)

	rtA := makeReservationTypeLinked(t, db, clinicA, "枠削除対象区分", nil, nil)
	slotA := makeAvailableSlot(t, db, clinicA, rtA.ID, 4, "14:00")

	t.Run("別クリニックIDでは NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicB, slotA.ID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("正しいクリニックIDで削除できる", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, clinicA, slotA.ID))
		_, err := repo.FindByID(ctx, clinicA, slotA.ID)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("存在しないIDの削除は NotFound", func(t *testing.T) {
		err := repo.Delete(ctx, clinicA, uint64(999999))
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})
}
