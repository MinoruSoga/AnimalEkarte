package service

// be9_2c_b1_mock_carriers_test.go — BE9-2C B①でbillingへ移動したtestが定義していた共有mockの
// carrier複製（残留consumer: accounting/cash_register系test。B④⑤移動時に解消）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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

// mockCampaignRepository — B①移動test（campaign_service_test.go）由来のcarrier複製（billing_item用・B③で解消）。
type mockCampaignRepository struct {
	repository.CampaignRepository
	findAllFn                  func(ctx context.Context, clinicID uint64) ([]model.Campaign, error)
	findByIDFn                 func(ctx context.Context, clinicID, id uint64) (*model.Campaign, error)
	createFn                   func(ctx context.Context, m *model.Campaign) (*model.Campaign, error)
	updateFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error)
	replaceTargetsFn           func(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
	reorderFn                  func(ctx context.Context, clinicID uint64, ids []uint64) error
	findApplicableForItemFn    func(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error)
	findAllApplicableForItemFn func(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error)
}

func (m *mockCampaignRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Campaign, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.Campaign{}, nil
}

func (m *mockCampaignRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Campaign, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Campaign{ID: id, ClinicID: clinicID}, nil
}

func (m *mockCampaignRepository) Create(ctx context.Context, campaign *model.Campaign) (*model.Campaign, error) {
	if m.createFn != nil {
		return m.createFn(ctx, campaign)
	}
	return campaign, nil
}

func (m *mockCampaignRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Campaign, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, fields)
	}
	return &model.Campaign{ID: id, ClinicID: clinicID}, nil
}

func (m *mockCampaignRepository) ReplaceTargets(ctx context.Context, campaignID uint64, categories []model.ItemCategory, itemIDs []uint64) error {
	if m.replaceTargetsFn != nil {
		return m.replaceTargetsFn(ctx, campaignID, categories, itemIDs)
	}
	return nil
}

func (m *mockCampaignRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockCampaignRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return nil
}

func (m *mockCampaignRepository) FindApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) (*model.Campaign, error) {
	if m.findApplicableForItemFn != nil {
		return m.findApplicableForItemFn(ctx, clinicID, date, category, merchandiseItemID)
	}
	return nil, nil
}

func (m *mockCampaignRepository) FindAllApplicableForItem(ctx context.Context, clinicID uint64, date time.Time, category model.ItemCategory, merchandiseItemID *uint64) ([]*model.Campaign, error) {
	if m.findAllApplicableForItemFn != nil {
		return m.findAllApplicableForItemFn(ctx, clinicID, date, category, merchandiseItemID)
	}
	return nil, nil
}
