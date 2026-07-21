package service

// be9_2c_b1_mock_carriers_test.go — BE9-2C B①でbillingへ移動したtestが定義していた共有mockの
// carrier複製（残留consumer: accounting/cash_register系test。B④⑤移動時に解消）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockPaymentMethodMasterRepository struct {
	findAllFn                     func(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error)
	createFn                      func(ctx context.Context, m *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error)
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error)
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	countUsageByPaymentMethodIDFn func(ctx context.Context, clinicID, id uint64) (int64, error)
	reorderFn                     func(ctx context.Context, clinicID uint64, ids []uint64) error
}

func (m *mockPaymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PaymentMethodMaster, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Create(ctx context.Context, pm *model.PaymentMethodMaster) (*model.PaymentMethodMaster, error) {
	if m.createFn != nil {
		return m.createFn(ctx, pm)
	}
	return pm, nil
}

func (m *mockPaymentMethodMasterRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PaymentMethodMaster, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return nil, nil
}

func (m *mockPaymentMethodMasterRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPaymentMethodMasterRepository) CountUsageByPaymentMethodID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByPaymentMethodIDFn != nil {
		return m.countUsageByPaymentMethodIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

func (m *mockPaymentMethodMasterRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

// mockInsuranceRepository — B①移動test（insurance_service_test.go）由来のcarrier複製（pet_service_test用）。
type mockInsuranceRepository struct {
	findAllFn                 func(ctx context.Context, clinicID uint64) ([]model.Insurance, error)
	findByIDFn                func(ctx context.Context, clinicID, id uint64) (*model.Insurance, error)
	createFn                  func(ctx context.Context, insurance *model.Insurance) error
	updateFn                  func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	reorderErr                error
	countUsageByInsuranceIDFn func(ctx context.Context, clinicID, insuranceID uint64) (int64, error)
}

func (m *mockInsuranceRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Insurance, error) {
	return m.findAllFn(ctx, clinicID)
}

func (m *mockInsuranceRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Insurance, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Create(ctx context.Context, insurance *model.Insurance) error {
	return m.createFn(ctx, insurance)
}

func (m *mockInsuranceRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Insurance, error) {
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockInsuranceRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockInsuranceRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockInsuranceRepository) CountUsageByInsuranceID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByInsuranceIDFn != nil {
		return m.countUsageByInsuranceIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

// mockBillingItemRepository — B③移動test由来のcarrier複製（lstep_health_tag_sync_prevention用・lstep移行時に解消）。
type mockBillingItemRepository struct {
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error)
	findByBillingIDFn             func(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error)
	createFn                      func(ctx context.Context, item *model.BillingItem) error
	updateFieldsFn                func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	updateBillingTotals           func(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
	hasItemByOwnerSinceFn         func(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	hasFoodPurchaseByOwnerSinceFn func(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
}

func (m *mockBillingItemRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.BillingItem, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockBillingItemRepository) FindByBillingID(ctx context.Context, clinicID, billingID uint64) ([]model.BillingItem, error) {
	return m.findByBillingIDFn(ctx, clinicID, billingID)
}

func (m *mockBillingItemRepository) Create(ctx context.Context, item *model.BillingItem) error {
	return m.createFn(ctx, item)
}

func (m *mockBillingItemRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockBillingItemRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockBillingItemRepository) UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
	if m.updateBillingTotals != nil {
		return m.updateBillingTotals(ctx, clinicID, billingID, subtotal, taxTotal, totalAmount)
	}
	return nil
}

func (m *mockBillingItemRepository) HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasItemByOwnerSinceFn != nil {
		return m.hasItemByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}

func (m *mockBillingItemRepository) HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error) {
	if m.hasFoodPurchaseByOwnerSinceFn != nil {
		return m.hasFoodPurchaseByOwnerSinceFn(ctx, clinicID, ownerID, since, names)
	}
	return false, nil
}

func (m *mockBillingItemRepository) FindUnbilledTrimmingItemsByPetID(_ context.Context, _, _ uint64) ([]model.BillingItem, error) {
	return nil, nil
}

func (m *mockBillingItemRepository) CountNonAccountingTrimmingByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

func ptrString(v string) *string { return &v }

func ptrInt64(v int64) *int64 { return &v }
