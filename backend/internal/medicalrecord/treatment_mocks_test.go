package medicalrecord

// treatment_mocks_test.go — BE9-2D sub-batch④b: treatment 系テストの共有 test double 群。
// 移動元 internal/service の共有 mock（mocks_shared_test.go / trimming_service_test.go /
// cross_tenant_master_fk_write_test.go の ok*/reject* builder 等）は package 外から import
// できないため、moved tests が実際に使う部分集合を service_deps.go の narrow view
// （medicineFinder/procedureFinder/consultationFinder/treatmentInventoryRepo/doseParamFinder）と
// in-package 具象 interface（TreatmentRepository）に対して再宣言する（service_deps_mock_test.go 同方針）。

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Treatment repository mock（移動元 treatment_service_test.go・全メソッド） ----

type mockTreatmentRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	findByIDFn              func(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error)
	lockByIDForUpdateFn     func(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error)
	createFn                func(ctx context.Context, treatment *model.Treatment) error
	updateFn                func(ctx context.Context, clinicID, treatmentID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, treatmentID uint64) error
	bulkUpdateSortOrderFn   func(ctx context.Context, updates []TreatmentSortUpdate) error
	findHistoryByPetIDFn    func(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
}

func (m *mockTreatmentRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockTreatmentRepository) FindByID(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, treatmentID)
	}
	return nil, nil
}

func (m *mockTreatmentRepository) LockByIDForUpdate(ctx context.Context, clinicID, treatmentID uint64) (*model.Treatment, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, treatmentID)
	}
	// Default: same snapshot as FindByID so existing unit tests keep working.
	return m.FindByID(ctx, clinicID, treatmentID)
}

func (m *mockTreatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	return m.createFn(ctx, treatment)
}

func (m *mockTreatmentRepository) Update(ctx context.Context, clinicID, treatmentID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, treatmentID, fields)
}

func (m *mockTreatmentRepository) Delete(ctx context.Context, clinicID, treatmentID uint64) error {
	return m.deleteFn(ctx, clinicID, treatmentID)
}

func (m *mockTreatmentRepository) BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error {
	return m.bulkUpdateSortOrderFn(ctx, updates)
}

func (m *mockTreatmentRepository) FindUnbilledByPetID(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
	return nil, nil
}

func (m *mockTreatmentRepository) FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	if m.findHistoryByPetIDFn != nil {
		return m.findHistoryByPetIDFn(ctx, clinicID, petID, filter, page, limit)
	}
	return nil, 0, nil
}

func (m *mockTreatmentRepository) CountFinalizedUnconfirmedByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, nil
}

// ---- Transactor mock（trimming_service_test.go の mockTransactor と同型・WithTx 素通し） ----

type mockTransactor struct {
	withTxErr error
	withTxFn  func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	if m.withTxErr != nil {
		return m.withTxErr
	}
	return fn(ctx)
}

// ---- AuditTxLogger mock（逸脱 audit の捕捉 entries + 失敗注入 logEntryTxErr） ----

type mockTreatmentAuditTxLogger struct {
	entries       []*AuditEntry
	logEntryTxErr error
}

func (m *mockTreatmentAuditTxLogger) LogEntryTx(_ context.Context, input *AuditEntry) error {
	m.entries = append(m.entries, input)
	return m.logEntryTxErr
}

// ---- narrow view mocks（medicineFinder/procedureFinder/consultationFinder/
// treatmentInventoryRepo/doseParamFinder） ----

// mockMedicineRepository は ⑥ で移動してきた各 service test の full 定義を使用（④b minimal 版は撤去）。

// mockProcedureRepository は ⑥ で移動してきた各 service test の full 定義を使用（④b minimal 版は撤去）。

// mockConsultationRepository は ⑥ で移動してきた各 service test の full 定義を使用（④b minimal 版は撤去）。

type mockInventoryRepository struct {
	findByIDFn      func(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	decreaseStockFn func(ctx context.Context, clinicID, id uint64, quantity float64) error
	createFn        func(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
}

func (m *mockInventoryRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockInventoryRepository) DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error {
	if m.decreaseStockFn != nil {
		return m.decreaseStockFn(ctx, clinicID, id, quantity)
	}
	return nil
}

// 以下3メソッドは medicineInventoryRepo view（⑥ medicine 連携在庫同期）用の nil-guard 実装。
func (m *mockInventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {
	if m.createFn != nil {
		return m.createFn(ctx, clinicID, item)
	}
	return nil
}

func (m *mockInventoryRepository) UpdateNameByMedicineCategory(_ context.Context, _ uint64, _, _ string) error {
	return nil
}

func (m *mockInventoryRepository) DeleteByNameAndMedicineCategory(_ context.Context, _ uint64, _ string) error {
	return nil
}

func (m *mockInventoryRepository) Delete(_ context.Context, _, _ uint64) error {
	return nil
}

// mockMedicineDoseParamRepository は MedicineDoseParamRepository の full test double
// （④b では doseParamFinder 部分集合だったが、⑥で dose_param service test が本 package へ
// 移動したため full 実装へ拡張。nil fn の NotFound デフォルト維持）。
type mockMedicineDoseParamRepository struct {
	findByMedicineIDFn         func(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error)
	findByMedicineAndSpeciesFn func(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error)
	createFn                   func(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error
	updateFn                   func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicineDoseParam, error)
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockMedicineDoseParamRepository) FindByMedicineID(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	if m.findByMedicineIDFn == nil {
		return nil, nil
	}
	return m.findByMedicineIDFn(ctx, clinicID, medicineID)
}

func (m *mockMedicineDoseParamRepository) FindByMedicineAndSpecies(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
	if m.findByMedicineAndSpeciesFn == nil {
		return nil, apperrors.WrapNotFound("medicine_dose_param", "")
	}
	return m.findByMedicineAndSpeciesFn(ctx, clinicID, medicineID, species)
}

func (m *mockMedicineDoseParamRepository) Create(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, clinicID, param)
}

func (m *mockMedicineDoseParamRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicineDoseParam, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, clinicID, id, fields)
}

func (m *mockMedicineDoseParamRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, clinicID, id)
}

// ---- builders（移動元 internal/service cross_tenant_master_fk_write_test.go /
// treatment_service_test.go の同名 helper の medicalrecord narrow-view 版） ----

func okMedicineRepo() medicineFinder {
	return &mockMedicineRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
		return &model.Medicine{ID: id}, nil
	}}
}

func okProcedureRepo() procedureFinder {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		return &model.Procedure{ID: id}, nil
	}}
}

func okConsultationRepo() consultationFinder {
	return &mockConsultationRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Consultation, error) {
		return &model.Consultation{ID: id}, nil
	}}
}

func rejectProcedureRepo(ownedID uint64) procedureFinder {
	return &mockProcedureRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.Procedure, error) {
		if id != ownedID {
			return nil, apperrors.WrapNotFound("procedure", "foreign")
		}
		return &model.Procedure{ID: id}, nil
	}}
}

func rejectInventoryRepo(ownedID uint64) *mockInventoryRepository {
	return &mockInventoryRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.InventoryItem, error) {
			if id != ownedID {
				return nil, apperrors.WrapNotFound("inventory_item", "foreign")
			}
			return &model.InventoryItem{ID: id}, nil
		},
	}
}

// draftMedicalRecordRepo は lockDraftMedicalRecord（mock の LockByIDForUpdate は FindByID へ
// fallback）を draft で通すための共通 mock。
func draftMedicalRecordRepo() *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
}

// benignVitalRepo は vital 未記録（dose 再検証は手動 fallback）を返す無害 mock。
func benignVitalRepo() *mockVitalRepository {
	return &mockVitalRepository{listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
		return nil, nil
	}}
}

// okInventoryRepo は service 側同名 builder の narrow 版（⑥ medicine InventoryFK テスト用・
// 具象を返し treatmentInventoryRepo/medicineInventoryRepo 両 view を満たす）。
func okInventoryRepo() *mockInventoryRepository {
	return &mockInventoryRepository{findByIDFn: func(_ context.Context, _, id uint64) (*model.InventoryItem, error) {
		return &model.InventoryItem{ID: id}, nil
	}}
}
