package medicalrecord

// hospitalization_mocks_test.go — BE9-2D ⑤: 移動した hospitalization 系テストが使う test double のうち、
// 定義が internal/service に残るもの（mockReservationRepository=appointment tests /
// mockAccountingRepository=accounting tests / mockBillingItemRepository=billing tests /
// mockCageRepository=cage tests が原本を使い続ける）の narrow-view 最小再宣言
// （treatment_mocks_test.go / service_deps_mock_test.go と同方針）。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockReservationRepository struct {
	assertOwnerInClinicFn  func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerInClinicFn func(ctx context.Context, clinicID, petID uint64) (uint64, error)
}

func (m *mockReservationRepository) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerInClinicFn != nil {
		return m.assertOwnerInClinicFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockReservationRepository) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerInClinicFn != nil {
		return m.findPetOwnerInClinicFn(ctx, clinicID, petID)
	}
	return 0, nil
}

type mockAccountingRepository struct {
	createFn func(ctx context.Context, clinicID uint64, billing *model.Billing) error
}

func (m *mockAccountingRepository) Create(ctx context.Context, clinicID uint64, billing *model.Billing) error {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, billing)
	}
	return nil
}

type mockBillingItemRepository struct {
	createFn            func(ctx context.Context, item *model.BillingItem) error
	updateBillingTotals func(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
}

func (m *mockBillingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	if m.createFn != nil {
		return m.createFn(ctx, item)
	}
	return nil
}

func (m *mockBillingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	if m.updateBillingTotals != nil {
		return m.updateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	}
	return nil
}

type mockCageRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
}

func (m *mockCageRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Cage{ID: id}, nil
}

// rejectCageRepo は internal/service cross_tenant_master_fk_write_test.go の同名 builder の
// narrow-view 版（CageFK guard テスト用）。
func rejectCageRepo(ownedID uint64) cageFinder {
	return &mockCageRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("cage", "foreign")
		}
		return &model.Cage{ID: id}, nil
	}}
}
