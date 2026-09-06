package medicalrecord

// medicine_dose_config_test.go — #201 B-2: 薬マスタ per_weight 書込検証 + 有効化 audit。
// default-deny（誤 per_weight 拒否）と per_weight 有効化の監査記録を固定する。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func medicineServiceWithAudit(repo *mockMedicineRepository, audit AuditTxLogger) MedicineService {
	inv := &mockInventoryRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.InventoryItem) error { return nil },
	}
	return NewMedicineServiceWithAudit(repo, inv, &mockTransactor{}, audit)
}

func strptr(s string) *string { return &s }

func TestMedicineService_Create_DoseConfig(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("per_weight + 錠 + strength は成功し有効化を audit", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 50; return nil },
		}
		audit := &mockTreatmentAuditTxLogger{}
		svc := medicineServiceWithAudit(repo, audit)

		_, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{
			Name: "メロキシカム", CalculationType: strptr("per_weight"),
			MedicineUnit: strptr("per_tablet"), Strength: fptr(1.5),
		})
		require.NoError(t, err)
		require.Len(t, audit.entries, 1)
		assert.Equal(t, model.AuditActionMedicinePerWeightEnable, audit.entries[0].Action)
		assert.Equal(t, model.AuditResourceMedicine, audit.entries[0].Resource)
	})

	t.Run("per_weight + per_dose は拒否（誤 per_weight・audit なし）", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 51; return nil },
		}
		audit := &mockTreatmentAuditTxLogger{}
		svc := medicineServiceWithAudit(repo, audit)

		_, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{
			Name: "インスリン", CalculationType: strptr("per_weight"),
			MedicineUnit: strptr("per_dose"), Strength: fptr(40),
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err), "誤 per_weight は InvalidInput: %v", err)
		assert.Empty(t, audit.entries, "拒否時は audit しない")
	})

	t.Run("per_weight + strength 無しは拒否", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 52; return nil },
		}
		svc := medicineServiceWithAudit(repo, &mockTreatmentAuditTxLogger{})
		_, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{
			Name: "薬", CalculationType: strptr("per_weight"), MedicineUnit: strptr("per_tablet"),
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("後方互換: calculation_type 未指定は none・audit なし", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 53; return nil },
		}
		audit := &mockTreatmentAuditTxLogger{}
		svc := medicineServiceWithAudit(repo, audit)
		med, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{Name: "通常薬"})
		require.NoError(t, err)
		assert.Equal(t, model.MedicineCalculationTypeNone, med.CalculationType)
		assert.Empty(t, audit.entries)
	})

	// BE-refactor.md R1-2: per_weight 有効化監査（LogEntryTx）が失敗すると Create 全体が失敗する
	// （fail-closed。薬剤作成のみが成功し監査だけ欠落する部分コミットを許さない）。
	t.Run("per_weight 有効化 audit 失敗は Create 全体をロールバックする（fail-closed）", func(t *testing.T) {
		repo := &mockMedicineRepository{
			createFn: func(_ context.Context, m *model.Medicine) error { m.ID = 54; return nil },
		}
		audit := &mockTreatmentAuditTxLogger{logEntryTxErr: errAuditWriteFailed}
		svc := medicineServiceWithAudit(repo, audit)

		_, err := svc.Create(context.Background(), clinicID, &CreateMedicineInput{
			Name: "メロキシカム2", CalculationType: strptr("per_weight"),
			MedicineUnit: strptr("per_tablet"), Strength: fptr(1.5),
		})
		require.Error(t, err, "audit 失敗は作成全体を失敗させる")
	})
}

func TestMedicineService_Update_DoseConfig(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("none→per_weight 有効化を audit", func(t *testing.T) {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
				unit := model.MedicineUnitPerTablet
				return &model.Medicine{ID: 60, ClinicID: clinicID, Name: "薬", CalculationType: model.MedicineCalculationTypeNone, MedicineUnit: &unit}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				return &model.Medicine{ID: 60, CalculationType: model.MedicineCalculationTypePerWeight, Strength: fptr(5)}, nil
			},
		}
		audit := &mockTreatmentAuditTxLogger{}
		svc := medicineServiceWithAudit(repo, audit)

		_, err := svc.Update(context.Background(), clinicID, 60, &UpdateMedicineInput{
			CalculationType: strptr("per_weight"), Strength: fptr(5),
		})
		require.NoError(t, err)
		require.Len(t, audit.entries, 1)
		assert.Equal(t, model.AuditActionMedicinePerWeightEnable, audit.entries[0].Action)
	})

	t.Run("既存 per_weight の strength クリアは拒否（含量必須）", func(t *testing.T) {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
				unit := model.MedicineUnitPerTablet
				return &model.Medicine{ID: 61, ClinicID: clinicID, CalculationType: model.MedicineCalculationTypePerWeight, MedicineUnit: &unit, Strength: fptr(5)}, nil
			},
		}
		svc := medicineServiceWithAudit(repo, &mockTreatmentAuditTxLogger{})
		_, err := svc.Update(context.Background(), clinicID, 61, &UpdateMedicineInput{ClearStrength: true})
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	// BE-refactor.md R1-2: per_weight 有効化監査（LogEntryTx）が失敗すると Update 全体が失敗する
	// （fail-closed。fields 更新のみが成功し監査だけ欠落する部分コミットを許さない）。
	t.Run("none→per_weight 有効化 audit 失敗は Update 全体をロールバックする（fail-closed）", func(t *testing.T) {
		repo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Medicine, error) {
				unit := model.MedicineUnitPerTablet
				return &model.Medicine{ID: 62, ClinicID: clinicID, Name: "薬", CalculationType: model.MedicineCalculationTypeNone, MedicineUnit: &unit}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateMedicineInput) (*model.Medicine, error) {
				return &model.Medicine{ID: 62, CalculationType: model.MedicineCalculationTypePerWeight, Strength: fptr(5)}, nil
			},
		}
		audit := &mockTreatmentAuditTxLogger{logEntryTxErr: errAuditWriteFailed}
		svc := medicineServiceWithAudit(repo, audit)

		_, err := svc.Update(context.Background(), clinicID, 62, &UpdateMedicineInput{
			CalculationType: strptr("per_weight"), Strength: fptr(5),
		})
		require.Error(t, err, "audit 失敗は更新全体を失敗させる")
	})
}
