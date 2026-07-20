package handler

// treatment_plan_hospitalization_mock_test.go — BE9-2D ⑤ carrier: 移動した
// hospitalization_handler_test.go の mockHospitalizationService の残置コピー
// （treatment_plan_handler_test（④外・残置）が入院所有権検証 harness に使用。
// 解消 = treatment-plan の domain 移行時）。medicalrecord 型で再宣言。

import (
	"context"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

type mockHospitalizationService struct {
	listFn func(
		ctx context.Context,
		clinicID uint64,
		petID, ownerID *uint64,
		status, startDate, endDate *string,
		page, limit int,
	) ([]model.Hospitalization, int64, error)
	getByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn               func(ctx context.Context, clinicID uint64, input *medicalrecord.CreateHospitalizationInput) (*model.Hospitalization, error)
	updateFn               func(ctx context.Context, clinicID, id uint64, input *medicalrecord.UpdateHospitalizationInput) (*model.Hospitalization, error)
	deleteFn               func(ctx context.Context, clinicID, id uint64) error
	dischargeWithBillingFn func(ctx context.Context, clinicID, id uint64, input medicalrecord.DischargeWithBillingInput) (*medicalrecord.DischargeWithBillingResult, error)
}

func (m *mockHospitalizationService) List(
	ctx context.Context,
	clinicID uint64,
	petID, ownerID *uint64,
	status, startDate, endDate *string,
	page, limit int,
) ([]model.Hospitalization, int64, error) {
	return m.listFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
}

func (m *mockHospitalizationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockHospitalizationService) Create(
	ctx context.Context, clinicID uint64, input *medicalrecord.CreateHospitalizationInput,
) (*model.Hospitalization, error) {
	return m.createFn(ctx, clinicID, input)
}

func (m *mockHospitalizationService) Update(
	ctx context.Context, clinicID, id uint64, input *medicalrecord.UpdateHospitalizationInput,
) (*model.Hospitalization, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockHospitalizationService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationService) DischargeWithBilling(
	ctx context.Context, clinicID, id uint64, input medicalrecord.DischargeWithBillingInput,
) (*medicalrecord.DischargeWithBillingResult, error) {
	return m.dischargeWithBillingFn(ctx, clinicID, id, input)
}
