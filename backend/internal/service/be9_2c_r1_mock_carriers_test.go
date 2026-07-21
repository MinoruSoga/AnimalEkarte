package service

// be9_2c_r1_mock_carriers_test.go — BE9-2C R①でreservationへ移動したtestが定義していた共有mockの
// carrier複製（残留consumer: appointment/liff/cross_tenant系test。liff(R⑤)/appointment(R④)移動時に解消）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockUnavailableTimeRepository は ReservationTypeUnavailableTimeRepository のテスト用モック
type mockUnavailableTimeRepository struct {
	findAllFn  func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error)
	createFn   func(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
	deleteFn   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockUnavailableTimeRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeUnavailableTime{}, nil
}

func (m *mockUnavailableTimeRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeUnavailableTime{ID: id}, nil
}

func (m *mockUnavailableTimeRepository) Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}

func (m *mockUnavailableTimeRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

// mockAvailableSlotRepository は ReservationTypeAvailableSlotRepository のテスト用モック実装
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
	return []model.ReservationTypeAvailableSlot{}, nil
}

func (m *mockAvailableSlotRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeAvailableSlot, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.ReservationTypeAvailableSlot{ID: id, ClinicID: clinicID}, nil
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
