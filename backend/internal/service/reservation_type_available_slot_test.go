package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockAvailableSlotRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error)
	createFn   func(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockAvailableSlotRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeAvailableSlot, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return nil, nil
}

func (m *mockAvailableSlotRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeAvailableSlot{ID: id}, nil
}

func (m *mockAvailableSlotRepository) Create(ctx context.Context, slot *model.ReservationTypeAvailableSlot) error {
	if m.createFn != nil {
		return m.createFn(ctx, slot)
	}
	return nil
}

func (m *mockAvailableSlotRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func TestFilterApplicableAvailableSlots_WeeklyAndSpecific(t *testing.T) {
	monday := int8(1)
	tuesday := int8(2)
	specificDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, jstLocation)

	result := filterApplicableAvailableSlots([]model.ReservationTypeAvailableSlot{
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
		{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &tuesday, StartTime: "12:30", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "14:00", IsActive: true},
		{AvailableType: model.AvailableSlotTypeSpecific, SpecificDate: &specificDate, StartTime: "15:00", IsActive: false},
	}, date)

	assert.Len(t, result, 2)
	assert.Equal(t, "09:45", result[0].StartTime)
	assert.Equal(t, "14:00", result[1].StartTime)
}

// TestMergeAvailableTimeSlots: 予約可能枠は加算モード（ホワイトリストなし）を検証する。
// 営業時間内の登録時刻は重複追加されず、営業時間外の登録時刻は追加される。
func TestMergeAvailableTimeSlots(t *testing.T) {
	monday := int8(1)
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, jstLocation) // Monday

	t.Run("営業時間外の時刻のみ追加される", func(t *testing.T) {
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
			{StartTime: "0945", EndTime: "1045"},
		}
		availableSlots := []model.ReservationTypeAvailableSlot{
			// 0945 は既存スロットと重複 → 追加されない
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "09:45", IsActive: true},
			// 0800 は営業時間外 → 追加される
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "08:00", IsActive: true},
		}
		result := mergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Len(t, result, 3)
		assert.Equal(t, "0800", result[0].StartTime)
		assert.Equal(t, "0900", result[1].StartTime) // end = 0800 + 60min = 0900
		assert.Equal(t, "0900", result[0].EndTime)
		assert.Equal(t, "0945", result[2].StartTime)
	})

	t.Run("該当日に枠なしでもスロットはブロックされない", func(t *testing.T) {
		tuesday := int8(2)
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
			{StartTime: "0945", EndTime: "1045"},
		}
		// 火曜日登録の枠を月曜日に適用 → 該当なし → slots がそのまま返る
		availableSlots := []model.ReservationTypeAvailableSlot{
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &tuesday, StartTime: "10:00", IsActive: true},
		}
		result := mergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Equal(t, slots, result)
	})

	t.Run("有効でない枠は無視される", func(t *testing.T) {
		slots := []TimeSlot{
			{StartTime: "0900", EndTime: "1000"},
		}
		availableSlots := []model.ReservationTypeAvailableSlot{
			{AvailableType: model.AvailableSlotTypeWeekly, DayOfWeek: &monday, StartTime: "08:00", IsActive: false},
		}
		result := mergeAvailableTimeSlots(slots, availableSlots, date, 60)

		assert.Equal(t, slots, result)
	})
}
