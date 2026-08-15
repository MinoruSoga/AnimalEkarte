package reservation

// repository_test.go — ReservationTypeAvailableSlotRepository の CRUD・clinic_id 隔離・NotFound を実 DB で検証する。
// makeReservationTypeLinked はフラット package の同名ヘルパーの複製（BE8-4 batch2）。

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

// setupTestDB は予約可能開始時刻テスト用に DB を整備する。
// reservation_types CASCADE で reservation_type_available_slots も連鎖クリアされる（FK ON DELETE CASCADE）。
func setupAvailableSlotRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.ReservationType{}, &model.ReservationTypeAvailableSlot{}))
	db.Exec("TRUNCATE TABLE reservation_types CASCADE")
	return db
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

func TestReservationTypeAvailableSlotRepository_FindAll(t *testing.T) {
	db := setupAvailableSlotRepoTestDB(t)
	repo := NewReservationTypeAvailableSlotRepository(db)
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

func TestReservationTypeAvailableSlotRepository_FindByID(t *testing.T) {
	db := setupAvailableSlotRepoTestDB(t)
	repo := NewReservationTypeAvailableSlotRepository(db)
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

func TestReservationTypeAvailableSlotRepository_Create(t *testing.T) {
	db := setupAvailableSlotRepoTestDB(t)
	repo := NewReservationTypeAvailableSlotRepository(db)
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

// BUG-455-S6: gorm default:true omits zero bools from INSERT.
func TestReservationTypeAvailableSlotRepository_Create_IsActiveFalsePersists(t *testing.T) {
	db := setupAvailableSlotRepoTestDB(t)
	repo := NewReservationTypeAvailableSlotRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)

	rtA := makeReservationTypeLinked(t, db, clinicA, "枠inactive作成区分", nil, nil)
	dow := int8(2)
	slot := &model.ReservationTypeAvailableSlot{
		ClinicID:          clinicA,
		ReservationTypeID: rtA.ID,
		AvailableType:     model.AvailableSlotTypeWeekly,
		DayOfWeek:         &dow,
		StartTime:         "10:00",
		IsActive:          false,
	}
	require.NoError(t, repo.Create(ctx, slot))
	assert.False(t, slot.IsActive)

	got, err := repo.FindByID(ctx, clinicA, slot.ID)
	require.NoError(t, err)
	assert.False(t, got.IsActive)

	var rawActive bool
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ReservationTypeAvailableSlot{}).
		Select("is_active").
		Where("id = ?", slot.ID).
		Scan(&rawActive).Error)
	assert.False(t, rawActive, "raw is_active must be false")
}

func TestReservationTypeAvailableSlotRepository_Delete(t *testing.T) {
	db := setupAvailableSlotRepoTestDB(t)
	repo := NewReservationTypeAvailableSlotRepository(db)
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
