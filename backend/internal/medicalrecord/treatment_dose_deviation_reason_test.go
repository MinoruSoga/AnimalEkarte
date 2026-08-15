package medicalrecord

// treatment_dose_deviation_reason_test.go — TASK-377: 用量逸脱理由の必須契約。
// filter: Test.*Treatment.*Dose.*(DeviationReason|Audit|Rollback|Clinic)

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	sentinelDoseDeviationReason = "SENTINEL_DOSE_REASON_TASK377_NEVER_LEAK"
	clinicIDDoseReason          = uint64(1)
	medicalRecordIDDoseReason   = uint64(100)
)

func withWideMaxParam(f *doseSaveFixture) {
	maxMgPerKg := 10.0
	f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
		return &model.MedicineDoseParam{
			ID: 1, ClinicID: 1, MedicineID: 50,
			Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
			DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
		}, nil
	}
}

func withBelowMinParam(f *doseSaveFixture) {
	minMgPerKg := 4.5
	maxMgPerKg := 10.0
	f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
		return &model.MedicineDoseParam{
			ID: 1, ClinicID: 1, MedicineID: 50,
			Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
			DosePerKg: 5, MinMgPerKg: &minMgPerKg, MaxMgPerKg: &maxMgPerKg,
		}, nil
	}
}

func medicineCreateWithReason(qty float64, reason string) *CreateTreatmentInput {
	in := medicineCreateInput(qty)
	in.DoseDeviationReason = reason
	return in
}

// TestTreatmentDoseDeviationReason_CreateRejectsMissingBlankAndOverlong は理由欠落・空白・500超で
// treatment/inventory/audit の write がゼロであることを固定する。
func TestTreatmentDoseDeviationReason_CreateRejectsMissingBlankAndOverlong(t *testing.T) {
	overlong := strings.Repeat("あ", 501)
	require.Equal(t, 501, utf8.RuneCountInString(overlong))

	tests := []struct {
		name   string
		qty    float64
		setup  func(*doseSaveFixture)
		reason string
	}{
		{name: "DeviatesFromComputed_理由欠落", qty: 3, setup: withWideMaxParam, reason: ""},
		{name: "DeviatesFromComputed_空白のみ", qty: 3, setup: withWideMaxParam, reason: "   \t\n"},
		{name: "DeviatesFromComputed_500文字超過", qty: 3, setup: withWideMaxParam, reason: overlong},
		{name: "BelowMinSaved_理由欠落", qty: 1.7, setup: withBelowMinParam, reason: ""},
		{name: "BelowMinSaved_空白のみ", qty: 1.7, setup: withBelowMinParam, reason: "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
			tt.setup(f)
			invCalls := 0
			invID := uint64(77)
			f.invRepo.decreaseStockFn = func(_ context.Context, _, _ uint64, _ float64) error {
				invCalls++
				return nil
			}
			svc := f.newSvc()
			in := medicineCreateWithReason(tt.qty, tt.reason)
			in.InventoryID = &invID

			got, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, in)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, apperrors.IsInvalidInput(err), "want InvalidInput: %v", err)
			assert.NotContains(t, err.Error(), sentinelDoseDeviationReason)
			assert.NotContains(t, err.Error(), overlong)
			assert.Zero(t, f.createCalls, "treatment write must be zero")
			assert.Zero(t, invCalls, "inventory write must be zero")
			assert.Empty(t, f.audit.entries, "audit write must be zero")
		})
	}
}

// TestTreatmentDoseDeviationReason_CreateSuccessWithReason は理由付き保存だけが成功し、
// snapshot・audit に normalized reason・flags・actor が残ることを固定する。
func TestTreatmentDoseDeviationReason_CreateSuccessWithReason(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f)
	invCalls := 0
	invID := uint64(77)
	f.invRepo.decreaseStockFn = func(_ context.Context, clinicID, id uint64, qty float64) error {
		assert.Equal(t, clinicIDDoseReason, clinicID)
		assert.Equal(t, invID, id)
		assert.Equal(t, 3.0, qty)
		invCalls++
		return nil
	}
	svc := f.newSvc()
	in := medicineCreateWithReason(3, "  体重再計測のため  ")
	in.InventoryID = &invID

	got, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, in)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, f.createCalls)
	assert.Equal(t, 1, invCalls)
	require.Len(t, f.audit.entries, 1)

	entry := f.audit.entries[0]
	require.NotNil(t, entry.ActorID)
	assert.Equal(t, uint64(3), *entry.ActorID)
	assert.Equal(t, model.AuditActionTreatmentDoseDeviation, entry.Action)
	newValue, ok := entry.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, false, newValue["exceeds_max"])
	assert.Equal(t, false, newValue["below_min"])
	assert.Equal(t, true, newValue["deviates_from_computed"])
	assert.Equal(t, "体重再計測のため", newValue["dose_deviation_reason"])

	require.NotNil(t, f.created)
	var snap DoseSnapshot
	require.NoError(t, json.Unmarshal(f.created.DoseParamSnapshot, &snap))
	assert.True(t, snap.DeviatesFromComputed)
	assert.Equal(t, "体重再計測のため", snap.DoseDeviationReason)
	assert.False(t, snap.ExceedsMax)
	assert.False(t, snap.BelowMin)
}

// TestTreatmentDoseDeviationReason_UpdateRejectsMissingReason は Update 経路の理由必須を固定する。
func TestTreatmentDoseDeviationReason_UpdateRejectsMissingReason(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f)
	medID := uint64(50)
	treatmentID := uint64(200)
	existing := &model.Treatment{
		ID: treatmentID, MedicalRecordID: medicalRecordIDDoseReason,
		ItemType: model.TreatmentItemTypeMedicine, MedicineID: &medID, Quantity: 1,
	}
	f.treatRepo = &mockTreatmentRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) { return existing, nil },
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			f.updateCalls++
			return nil
		},
	}
	svc := f.newSvc()
	qty := 3.0
	got, err := svc.Update(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, treatmentID, &UpdateTreatmentInput{
		Quantity: &qty,
		// DoseDeviationReason 欠落
	})
	require.Error(t, err)
	assert.Nil(t, got)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Zero(t, f.updateCalls)
	assert.Empty(t, f.audit.entries)
}

// TestTreatmentDoseDeviationReason_UpdateSuccessWithReason は Update 成功と audit 内容を固定する。
func TestTreatmentDoseDeviationReason_UpdateSuccessWithReason(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withBelowMinParam(f)
	medID := uint64(50)
	treatmentID := uint64(200)
	var updatedFields map[string]any
	existing := &model.Treatment{
		ID: treatmentID, MedicalRecordID: medicalRecordIDDoseReason,
		ItemType: model.TreatmentItemTypeMedicine, MedicineID: &medID, Quantity: 2,
	}
	f.treatRepo = &mockTreatmentRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) { return existing, nil },
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
			f.updateCalls++
			updatedFields = fields
			return nil
		},
	}
	svc := f.newSvc()
	qty := 1.7
	reason := "低用量から漸増する計画"
	actor := uint64(9)
	got, err := svc.Update(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, treatmentID, &UpdateTreatmentInput{
		Quantity:             &qty,
		DoseDeviationReason:  &reason,
		ActorID:              &actor,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 1, f.updateCalls)
	require.Len(t, f.audit.entries, 1)
	assert.Equal(t, &actor, f.audit.entries[0].ActorID)
	newValue, ok := f.audit.entries[0].NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, newValue["below_min"])
	assert.Equal(t, reason, newValue["dose_deviation_reason"])

	snapRaw, ok := updatedFields["dose_param_snapshot"].(json.RawMessage)
	require.True(t, ok)
	var snap DoseSnapshot
	require.NoError(t, json.Unmarshal(snapRaw, &snap))
	assert.True(t, snap.BelowMin)
	assert.Equal(t, reason, snap.DoseDeviationReason)
}

// TestTreatmentDoseDeviationReason_DirectAPIBypassSameValidation は FE を経由しない service 直接呼び出しでも
// 同じ理由 validation が働くことを固定する（handler 非経由 = direct API 相当）。
func TestTreatmentDoseDeviationReason_DirectAPIBypassSameValidation(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f)
	svc := f.newSvc()
	// 直接 CreateTreatmentInput を組み立てる（HTTP request を経由しない）
	medID := uint64(50)
	actor := uint64(3)
	_, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, &CreateTreatmentInput{
		ItemType: model.TreatmentItemTypeMedicine, MedicineID: &medID, Quantity: 3, ActorID: &actor,
		// 理由なし
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Zero(t, f.createCalls)
	assert.Empty(t, f.audit.entries)
}

// TestTreatmentDoseRollback_MissingActorNilAuditMarshal は reason-required 経路の失敗注入で
// treatment/inventory/audit write が全てゼロになることを固定する。
func TestTreatmentDoseRollback_MissingActorNilAuditMarshal(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*doseSaveFixture, *CreateTreatmentInput)
	}{
		{
			name: "missing authenticated actor",
			setup: func(_ *doseSaveFixture, in *CreateTreatmentInput) {
				in.ActorID = nil
			},
		},
		{
			name: "nil audit dependency",
			setup: func(f *doseSaveFixture, _ *CreateTreatmentInput) {
				f.audit = nil
			},
		},
		{
			name: "audit write failure",
			setup: func(f *doseSaveFixture, _ *CreateTreatmentInput) {
				f.audit.logEntryTxErr = errAuditWriteFailed
			},
		},
		{
			name: "forced snapshot marshal failure",
			setup: func(_ *doseSaveFixture, _ *CreateTreatmentInput) {
				prev := doseSnapshotMarshal
				doseSnapshotMarshal = func(any) ([]byte, error) {
					return nil, errAuditWriteFailed
				}
				t.Cleanup(func() { doseSnapshotMarshal = prev })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
			withWideMaxParam(f)
			invID := uint64(77)
			f.invRepo.decreaseStockFn = func(_ context.Context, _, _ uint64, _ float64) error {
				f.inventoryCalls++
				return nil
			}
			in := medicineCreateWithReason(3, "臨床上の判断")
			in.InventoryID = &invID
			tt.setup(f, in)
			svc := f.newSvc()

			got, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, in)
			require.Error(t, err, tt.name)
			assert.Nil(t, got)
			assert.Zero(t, f.createCalls, "treatment write zero: %s", tt.name)
			assert.Zero(t, f.inventoryCalls, "inventory write zero: %s", tt.name)
			if f.audit != nil {
				assert.Empty(t, f.audit.entries, "audit committed write zero: %s", tt.name)
			}
			assert.NotContains(t, err.Error(), "臨床上の判断")
		})
	}
}

// TestTreatmentDoseDeviationReason_SafeReevaluationClearsStaleReason は safe dose 再評価で
// snapshot から stale な理由が除去されることを固定する。
func TestTreatmentDoseDeviationReason_SafeReevaluationClearsStaleReason(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	// max を十分大きくし、qty=2 が推奨ちょうど（乖離なし・下限なし）になる
	maxMgPerKg := 100.0
	f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
		return &model.MedicineDoseParam{
			ID: 1, ClinicID: 1, MedicineID: 50,
			Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
			DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
		}, nil
	}
	medID := uint64(50)
	treatmentID := uint64(200)
	staleSnap, err := json.Marshal(DoseSnapshot{
		Species:              model.MedicineDoseSpeciesDog,
		DoseDeviationReason:  "古い逸脱理由",
		DeviatesFromComputed: true,
	})
	require.NoError(t, err)
	existing := &model.Treatment{
		ID: treatmentID, MedicalRecordID: medicalRecordIDDoseReason,
		ItemType: model.TreatmentItemTypeMedicine, MedicineID: &medID, Quantity: 5,
		DoseParamSnapshot: staleSnap,
	}
	var updatedFields map[string]any
	f.treatRepo = &mockTreatmentRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Treatment, error) { return existing, nil },
		updateFn: func(_ context.Context, _, _ uint64, fields map[string]any) error {
			f.updateCalls++
			updatedFields = fields
			return nil
		},
	}
	svc := f.newSvc()
	qty := 2.0 // 推奨値 = 2（safe）
	staleReason := "古い逸脱理由"
	actor := uint64(3)
	_, err = svc.Update(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, treatmentID, &UpdateTreatmentInput{
		Quantity:            &qty,
		DoseDeviationReason: &staleReason, // クライアントが送っても safe なら捨てる
		ActorID:             &actor,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, f.updateCalls)
	assert.Empty(t, f.audit.entries, "safe dose では逸脱 audit を書かない")

	snapRaw, ok := updatedFields["dose_param_snapshot"].(json.RawMessage)
	require.True(t, ok)
	var snap DoseSnapshot
	require.NoError(t, json.Unmarshal(snapRaw, &snap))
	assert.Empty(t, snap.DoseDeviationReason, "stale reason must be cleared")
	assert.False(t, snap.DeviatesFromComputed)
}

// TestTreatmentDoseDeviationReason_NegativeContractNoLeak は sentinel 理由が owner history・
// validation error・slog に出ないことを固定する。
func TestTreatmentDoseDeviationReason_NegativeContractNoLeak(t *testing.T) {
	// validation error に本文が出ない
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f)
	svc := f.newSvc()
	_, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateWithReason(3, ""))
	require.Error(t, err)
	assert.NotContains(t, err.Error(), sentinelDoseDeviationReason)

	// 成功時の owner-facing history DTO に理由が出ない
	f2 := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f2)
	svc2 := f2.newSvc()
	created, err := svc2.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateWithReason(3, sentinelDoseDeviationReason))
	require.NoError(t, err)
	require.NotNil(t, created)
	// pet history response は dose snapshot を載せない
	history := toPetTreatmentHistoryResponse(created)
	body, err := json.Marshal(history)
	require.NoError(t, err)
	assert.NotContains(t, string(body), sentinelDoseDeviationReason)
	assert.NotContains(t, string(body), "dose_deviation_reason")
	assert.NotContains(t, string(body), "dose_param_snapshot")

	// treatment response には snapshot 経由で載り得る（既存 contract）。validation/history 以外は本 task の露出面ではない。
	// slog に理由本文が出ない（Create 成功ログ）
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	f3 := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f3)
	svc3 := f3.newSvc()
	_, err = svc3.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateWithReason(3, sentinelDoseDeviationReason))
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), sentinelDoseDeviationReason)
}

// TestTreatmentDoseAudit_ReasonRequiredRegressionCaps は上限超過 hard reject と safe/missing の既存契約が
// 退行していないことを固定する。
func TestTreatmentDoseAudit_ReasonRequiredRegressionCaps(t *testing.T) {
	t.Run("上限超過は理由があっても hard reject・write ゼロ", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()
		qty := (20 + 2*safetyEpsilon) / 10
		got, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateWithReason(qty, "理由あっても不可"))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Zero(t, f.createCalls)
		assert.Empty(t, f.audit.entries)
	})

	t.Run("safe dose は理由不要で保存でき audit なし", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()
		_, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateInput(2))
		require.NoError(t, err)
		assert.Equal(t, 1, f.createCalls)
		assert.Empty(t, f.audit.entries)
		var snap DoseSnapshot
		require.NoError(t, json.Unmarshal(f.created.DoseParamSnapshot, &snap))
		assert.Empty(t, snap.DoseDeviationReason)
	})

	t.Run("missing data は理由なしで保存継続", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
			return nil, nil
		}
		svc := f.newSvc()
		_, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, medicineCreateInput(5))
		require.NoError(t, err)
		assert.Equal(t, 1, f.createCalls)
		assert.Nil(t, f.created.DoseAmountMg)
		assert.Empty(t, f.audit.entries)
	})
}

// TestTreatmentDoseClinic_IsolationPreservedOnReasonPath は reason-required 保存でも clinic スコープが
// 維持されることを、在庫減算 clinicID と audit clinicID で固定する。
func TestTreatmentDoseClinic_IsolationPreservedOnReasonPath(t *testing.T) {
	f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
	withWideMaxParam(f)
	var seenClinic uint64
	invID := uint64(77)
	f.invRepo.decreaseStockFn = func(_ context.Context, clinicID, _ uint64, _ float64) error {
		seenClinic = clinicID
		return nil
	}
	svc := f.newSvc()
	in := medicineCreateWithReason(3, "clinic-scoped reason")
	in.InventoryID = &invID
	_, err := svc.Create(context.Background(), clinicIDDoseReason, medicalRecordIDDoseReason, in)
	require.NoError(t, err)
	assert.Equal(t, clinicIDDoseReason, seenClinic)
	require.Len(t, f.audit.entries, 1)
	require.NotNil(t, f.audit.entries[0].ClinicID)
	assert.Equal(t, clinicIDDoseReason, *f.audit.entries[0].ClinicID)
}
