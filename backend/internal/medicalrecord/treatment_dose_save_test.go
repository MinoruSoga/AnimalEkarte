package medicalrecord

// treatment_dose_save_test.go — #201 B-2: treatment 保存時 BE 再検証の統合テスト。
// species 解決→param 取得→保存時再検証→スナップショット永続化→逸脱 audit→fail-closed を mock で検証する。

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// errAuditWriteFailed は BE-refactor.md R1-2 の fail-closed 回帰テスト用の失敗注入センチネル
// （mockTreatmentAuditTxLogger.logEntryTxErr にセットして使う。internal/service 側の medicine/
// dose-param テストは carrier の同名センチネルを使い続ける）。
var errAuditWriteFailed = errors.New("audit write failed")

// ---- helpers ----

type doseSaveFixture struct {
	treatRepo *mockTreatmentRepository
	medRepo   *mockMedicineRepository
	mrRepo    *mockMedicalRecordRepository
	vitalRepo *mockVitalRepository
	paramRepo *mockMedicineDoseParamRepository
	invRepo   *mockInventoryRepository
	// audit は非nilフラグ + 逸脱 audit の捕捉（entries []*AuditEntry）/失敗注入（logEntryTxErr）。
	audit   *mockTreatmentAuditTxLogger
	created *model.Treatment
}

// newSvc は個別依存注入コンストラクタ（BE9-2D ④b）で fixture の mock を配線する。
// 旧 harness の repos.TransactionFn インライン実行は mockTransactor の WithTx 素通しが等価。
func (f *doseSaveFixture) newSvc() TreatmentService {
	return NewTreatmentServiceWithAudit(
		f.treatRepo, f.mrRepo, f.medRepo, okProcedureRepo(), okConsultationRepo(), f.invRepo,
		f.vitalRepo, f.paramRepo, &mockTransactor{}, f.audit)
}

// newDoseSaveFixture は per_weight 医薬 1 種・犬・体重 4kg・dose 5mg/kg・max 5・strength 10mg/錠 の保存環境を構築する。
// calcType=none を指定すると後方互換（再検証なし）を再現する。paramSpecies で取得 param の種を上書きできる（mismatch 検証）。
func newDoseSaveFixture(t *testing.T, calcType model.MedicineCalculationType, paramSpecies model.MedicineDoseSpecies) *doseSaveFixture {
	t.Helper()
	f := &doseSaveFixture{audit: &mockTreatmentAuditTxLogger{}, invRepo: &mockInventoryRepository{}}

	f.medRepo = &mockMedicineRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
			strength := 10.0
			unit := model.MedicineUnitPerTablet
			return &model.Medicine{
				ID: 50, ClinicID: 1, Name: "テスト薬",
				CalculationType: calcType, MedicineUnit: &unit, Strength: &strength,
			}, nil
		},
	}
	f.mrRepo = &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			petID := uint64(7)
			return &model.MedicalRecord{
				Status: model.MedicalRecordStatusDraft,
				PetID:  &petID,
				Pet:    &model.Pet{AnimalSpecies: &model.AnimalSpecies{Name: "犬"}},
			}, nil
		},
	}
	f.vitalRepo = &mockVitalRepository{
		listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
			w := 4.0
			return []model.VitalRecord{
				{ID: 11, Weight: &w, WeightUnit: model.BodyWeightUnitKg, RecordedAt: time.Date(2026, 6, 28, 9, 0, 0, 0, time.UTC)},
			}, nil
		},
	}
	f.paramRepo = &mockMedicineDoseParamRepository{
		findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{
				ID: 1, ClinicID: 1, MedicineID: 50,
				Species: paramSpecies, DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MaxMgPerKg: fptr(5),
			}, nil
		},
	}
	f.treatRepo = &mockTreatmentRepository{
		createFn: func(_ context.Context, tr *model.Treatment) error {
			tr.ID = 99
			f.created = tr
			return nil
		},
	}
	return f
}

func medicineCreateInput(qty float64) *CreateTreatmentInput {
	medID := uint64(50)
	actor := uint64(3)
	return &CreateTreatmentInput{
		ItemType:   model.TreatmentItemTypeMedicine,
		MedicineID: &medID,
		Quantity:   qty,
		ActorID:    &actor,
	}
}

func TestTreatmentService_Create_DoseRevalidation(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("推奨どおり保存: スナップショット永続化・逸脱 audit なし", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2)) // computed=2錠
		require.NoError(t, err)
		require.NotNil(t, f.created)
		require.NotNil(t, f.created.DoseAmountMg, "スナップショット dose_amount_mg が永続化されること")
		assert.InDelta(t, 20.0, *f.created.DoseAmountMg, 1e-6)
		require.NotNil(t, f.created.DoseWeightKg)
		assert.InDelta(t, 4.0, *f.created.DoseWeightKg, 1e-6)
		assert.NotEmpty(t, f.created.DoseParamSnapshot, "dose_param_snapshot(jsonb) が値で固定されること")
		assert.Empty(t, f.audit.entries, "上限内・乖離なしでは逸脱 audit は発火しない")
	})

	t.Run("回帰: 範囲外 quantity で保存 → 逸脱 audit 記録（silent 過量保存の閉鎖）", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(5)) // 5錠=50mg > cap 20mg
		require.NoError(t, err, "上書きは拒否せず記録する")
		require.NotNil(t, f.created.DoseAmountMg)
		assert.InDelta(t, 50.0, *f.created.DoseAmountMg, 1e-6)
		require.Len(t, f.audit.entries, 1, "逸脱は audit に1件記録される")
		entry := f.audit.entries[0]
		assert.Equal(t, model.AuditActionTreatmentDoseDeviation, entry.Action)
		assert.Equal(t, model.AuditResourceTreatmentDose, entry.Resource)
		require.NotNil(t, entry.ActorID)
		assert.Equal(t, uint64(3), *entry.ActorID, "実施者が記録される")
		assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	})

	t.Run("species 不一致は fail-closed（保存中止）", func(t *testing.T) {
		// pet=犬 だが param が cat を返す → 防御アサートが発火し保存を中止する。
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesCat)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2))
		require.Error(t, err, "species 不一致では保存してはならない")
		assert.True(t, apperrors.IsConflict(err), "fail-closed は Conflict: %v", err)
	})

	t.Run("後方互換: calculation_type=none は再検証なし・スナップショットなし", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypeNone, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(7))
		require.NoError(t, err)
		require.NotNil(t, f.created)
		assert.Nil(t, f.created.DoseAmountMg, "none では dose スナップショットを書かない")
		assert.Empty(t, f.created.DoseParamSnapshot)
		assert.Empty(t, f.audit.entries)
	})

	// BE-refactor.md R1-2: 逸脱 audit（AuditTxLogger.LogEntryTx）が失敗すると Create 全体が失敗する
	// （fail-closed。treatment/在庫減算のみが成功し監査だけ欠落する部分コミットを許さない）。
	t.Run("逸脱 audit 失敗は Create 全体をロールバックする（fail-closed）", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		f.audit.logEntryTxErr = errAuditWriteFailed
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(5)) // 5錠=50mg > cap 20mg → 逸脱
		require.Error(t, err, "audit 失敗は treatment 作成全体を失敗させる")
	})
}

// TestTreatmentService_Update_DoseRevalidation は Update 経路の逸脱 audit（BE-refactor.md R1-2 で
// Create と同じ auditDoseDeviationTx 経路に統一）を固定する。quantity 変更で再評価対象になり、
// 逸脱時は audit 記録・audit 失敗時は Update 全体が fail-closed で失敗することを検証する。
func TestTreatmentService_Update_DoseRevalidation(t *testing.T) {
	const clinicID = uint64(1)
	const treatmentID = uint64(200)

	newUpdateFixture := func(t *testing.T) *doseSaveFixture {
		t.Helper()
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		medID := uint64(50)
		existing := &model.Treatment{
			ID: treatmentID, MedicalRecordID: 100, ItemType: model.TreatmentItemTypeMedicine,
			MedicineID: &medID, Quantity: 1,
		}
		f.treatRepo = &mockTreatmentRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) { return existing, nil },
			updateFn:   func(_ context.Context, _, _ uint64, _ map[string]any) error { return nil },
		}
		return f
	}

	t.Run("範囲外 quantity への更新 → 逸脱 audit 記録", func(t *testing.T) {
		f := newUpdateFixture(t)
		svc := f.newSvc()

		qty := 5.0 // 5錠=50mg > cap 20mg
		_, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{Quantity: &qty})
		require.NoError(t, err, "上書きは拒否せず記録する")
		require.Len(t, f.audit.entries, 1)
		assert.Equal(t, model.AuditActionTreatmentDoseDeviation, f.audit.entries[0].Action)
	})

	// BE-refactor.md R1-2: audit 失敗は Update 全体を fail-closed で失敗させる。
	t.Run("逸脱 audit 失敗は Update 全体をロールバックする（fail-closed）", func(t *testing.T) {
		f := newUpdateFixture(t)
		f.audit.logEntryTxErr = errAuditWriteFailed
		svc := f.newSvc()

		qty := 5.0
		_, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{Quantity: &qty})
		require.Error(t, err, "audit 失敗は treatment 更新全体を失敗させる")
	})
}
