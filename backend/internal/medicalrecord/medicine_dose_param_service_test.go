package medicalrecord

// medicine_dose_param_service_test.go — #201 B-2c: 種軸 dose パラメータ CRUD の service 検証。
//
// 固定する不変条件:
//   - 親 medicine の所有権/存在（clinicScope FindByID）が前提。別 clinic → NotFound(IDOR 遮断)。
//   - 医療安全ガード: per_weight でない薬剤・互換しない単位への dose param 設定を拒否（誤 per_weight 保存防止）。
//   - 行検証（default-deny / range / 上限必須）を ValidateMedicineDoseParamInput で二重化。
//   - 作成/更新/削除を audit_logs に before/after で記録（upsert/delete アクション）。

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMedicineDoseParamRepository / mockTreatmentAuditTxLogger は treatment_dose_save_test.go の共有モックを再利用する（DRY）。

// ---- helpers ----

const doseClinicID = uint64(1)

// perWeightMedicineRepo は per_weight 適格な親 medicine を返す medRepo モック。
func perWeightMedicineRepo() *mockMedicineRepository {
	return &mockMedicineRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Medicine, error) {
			unit := model.MedicineUnitPerTablet
			return &model.Medicine{
				ID: id, ClinicID: clinicID, Name: "メロキシカム",
				CalculationType: model.MedicineCalculationTypePerWeight,
				MedicineUnit:    &unit, Strength: fptr(1.5),
			}, nil
		},
	}
}

func validDoseParamInput() *MedicineDoseParamInput {
	return &MedicineDoseParamInput{
		Species:    model.MedicineDoseSpeciesDog,
		DoseBasis:  model.MedicineDoseBasisPerAdministration,
		DosePerKg:  5,
		MaxMgPerKg: fptr(10), // per_weight には上限必須
	}
}

// ---- Upsert: 親 medicine 所有権 / 医療安全ガード ----

func TestMedicineDoseParamService_Upsert_OwnershipAndGuards(t *testing.T) {
	tests := []struct {
		name      string
		medRepo   *mockMedicineRepository
		input     *MedicineDoseParamInput
		wantErrFn func(error) bool
	}{
		{
			name: "別clinicのmedicineはNotFound(IDOR遮断)",
			medRepo: &mockMedicineRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
					return nil, apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
				},
			},
			input:     validDoseParamInput(),
			wantErrFn: apperrors.IsNotFound,
		},
		{
			name: "calculation_type=noneの薬剤への設定は拒否",
			medRepo: &mockMedicineRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Medicine, error) {
					return &model.Medicine{ID: id, ClinicID: clinicID, CalculationType: model.MedicineCalculationTypeNone}, nil
				},
			},
			input:     validDoseParamInput(),
			wantErrFn: apperrors.IsInvalidInput,
		},
		{
			name: "per_weightだが互換しない単位(per_dose)は拒否(strength単位不整合)",
			medRepo: &mockMedicineRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Medicine, error) {
					unit := model.MedicineUnitPerDose
					return &model.Medicine{
						ID: id, ClinicID: clinicID, CalculationType: model.MedicineCalculationTypePerWeight,
						MedicineUnit: &unit, Strength: fptr(40),
					}, nil
				},
			},
			input:     validDoseParamInput(),
			wantErrFn: apperrors.IsInvalidInput,
		},
		{
			name:      "dose_per_kg<=0は拒否",
			medRepo:   perWeightMedicineRepo(),
			input:     &MedicineDoseParamInput{Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 0, MaxMgPerKg: fptr(10)},
			wantErrFn: apperrors.IsInvalidInput,
		},
		{
			name:      "min>maxは拒否",
			medRepo:   perWeightMedicineRepo(),
			input:     &MedicineDoseParamInput{Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 5, MinMgPerKg: fptr(20), MaxMgPerKg: fptr(10)},
			wantErrFn: apperrors.IsInvalidInput,
		},
		{
			name:      "上限(max/absolute)が無いと拒否(過量防止必須)",
			medRepo:   perWeightMedicineRepo(),
			input:     &MedicineDoseParamInput{Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 5},
			wantErrFn: apperrors.IsInvalidInput,
		},
		{
			name:      "不正なspeciesは拒否",
			medRepo:   perWeightMedicineRepo(),
			input:     &MedicineDoseParamInput{Species: model.MedicineDoseSpecies("rabbit"), DoseBasis: model.MedicineDoseBasisPerAdministration, DosePerKg: 5, MaxMgPerKg: fptr(10)},
			wantErrFn: apperrors.IsInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit := &mockTreatmentAuditTxLogger{}
			svc := NewMedicineDoseParamService(&mockMedicineDoseParamRepository{}, tt.medRepo, &mockTransactor{}, audit)

			_, err := svc.Upsert(context.Background(), doseClinicID, 50, tt.input, nil)
			require.Error(t, err)
			assert.True(t, tt.wantErrFn(err), "想定外のエラー種別: %v", err)
			assert.Empty(t, audit.entries, "拒否時は audit しない")
		})
	}
}

// ---- Upsert: create / update + audit ----

func TestMedicineDoseParamService_Upsert_CreatesWhenAbsent(t *testing.T) {
	audit := &mockTreatmentAuditTxLogger{}
	created := false
	repo := &mockMedicineDoseParamRepository{
		findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return nil, apperrors.WrapNotFound("medicine_dose_param", "") // 有効行なし → 作成
		},
		createFn: func(_ context.Context, clinicID uint64, p *model.MedicineDoseParam) error {
			created = true
			assert.Equal(t, doseClinicID, clinicID)
			assert.Equal(t, model.MedicineDoseSpeciesDog, p.Species)
			assert.Equal(t, 5.0, p.DosePerKg)
			p.ID = 100
			return nil
		},
	}
	svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)

	staffID := uint64(7)
	got, err := svc.Upsert(context.Background(), doseClinicID, 50, validDoseParamInput(), &staffID)
	require.NoError(t, err)
	assert.True(t, created, "有効行が無い時は Create が呼ばれる")
	require.NotNil(t, got)
	assert.Equal(t, uint64(100), got.ID)

	require.Len(t, audit.entries, 1)
	e := audit.entries[0]
	assert.Equal(t, model.AuditActionMedicineDoseParamUpsert, e.Action)
	assert.Equal(t, model.AuditResourceMedicineDoseParam, e.Resource)
	assert.Equal(t, model.AuditActorTypeStaff, e.ActorType)
	assert.Nil(t, e.OldValue, "新規作成は before なし")
	assert.NotNil(t, e.NewValue)
}

func TestMedicineDoseParamService_Upsert_UpdatesWhenPresent(t *testing.T) {
	audit := &mockTreatmentAuditTxLogger{}
	var capturedFields map[string]any
	repo := &mockMedicineDoseParamRepository{
		findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{ID: 200, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesDog, DosePerKg: 3, MaxMgPerKg: fptr(8)}, nil
		},
		updateFn: func(_ context.Context, clinicID, id uint64, cmd MedicineDoseParamInput) (*model.MedicineDoseParam, error) {
			capturedFields = buildDoseParamReplaceFields(&cmd)
			assert.Equal(t, doseClinicID, clinicID)
			assert.Equal(t, uint64(200), id)
			return &model.MedicineDoseParam{ID: 200, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesDog, DosePerKg: 5, MaxMgPerKg: fptr(10)}, nil
		},
	}
	svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)

	got, err := svc.Upsert(context.Background(), doseClinicID, 50, validDoseParamInput(), nil)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, 5.0, got.DosePerKg)

	// PUT 全列置換: 省略 optional は NULL 明示。
	assert.Equal(t, 5.0, capturedFields["dose_per_kg"])
	assert.Nil(t, capturedFields["min_mg_per_kg"], "未指定 optional は NULL 明示置換")
	assert.Nil(t, capturedFields["rounding_mode"])

	require.Len(t, audit.entries, 1)
	e := audit.entries[0]
	assert.Equal(t, model.AuditActionMedicineDoseParamUpsert, e.Action)
	assert.Equal(t, model.AuditActorTypeSystem, e.ActorType, "actorID nil は system")
	assert.NotNil(t, e.OldValue, "更新は before あり")
	assert.NotNil(t, e.NewValue)
}

// TestMedicineDoseParamService_Upsert_AuditFailureRollsBack は BE-refactor.md R1-2 の
// fail-closed 契約を固定する: 監査（LogEntryTx）が失敗すると Upsert（create/update 両分岐）全体が
// エラーを返す（dose param の書込のみが成功し監査だけ欠落する部分コミットを許さない）。
func TestMedicineDoseParamService_Upsert_AuditFailureRollsBack(t *testing.T) {
	t.Run("create分岐: audit失敗はUpsert全体を失敗させる", func(t *testing.T) {
		audit := &mockTreatmentAuditTxLogger{logEntryTxErr: errAuditWriteFailed}
		repo := &mockMedicineDoseParamRepository{
			findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
				return nil, apperrors.WrapNotFound("medicine_dose_param", "")
			},
			createFn: func(_ context.Context, _ uint64, p *model.MedicineDoseParam) error { p.ID = 400; return nil },
		}
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)
		_, err := svc.Upsert(context.Background(), doseClinicID, 50, validDoseParamInput(), nil)
		require.Error(t, err, "audit失敗はcreate全体を失敗させる")
	})

	t.Run("update分岐: audit失敗はUpsert全体を失敗させる", func(t *testing.T) {
		audit := &mockTreatmentAuditTxLogger{logEntryTxErr: errAuditWriteFailed}
		repo := &mockMedicineDoseParamRepository{
			findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
				return &model.MedicineDoseParam{ID: 401, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesDog, DosePerKg: 3, MaxMgPerKg: fptr(8)}, nil
			},
			updateFn: func(_ context.Context, _, _ uint64, _ MedicineDoseParamInput) (*model.MedicineDoseParam, error) {
				return &model.MedicineDoseParam{ID: 401, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesDog, DosePerKg: 5, MaxMgPerKg: fptr(10)}, nil
			},
		}
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)
		_, err := svc.Upsert(context.Background(), doseClinicID, 50, validDoseParamInput(), nil)
		require.Error(t, err, "audit失敗はupdate全体を失敗させる")
	})
}

// ---- Delete + audit ----

func TestMedicineDoseParamService_Delete(t *testing.T) {
	t.Run("別clinicはNotFound・audit無し", func(t *testing.T) {
		audit := &mockTreatmentAuditTxLogger{}
		medRepo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
				return nil, apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
			},
		}
		svc := NewMedicineDoseParamService(&mockMedicineDoseParamRepository{}, medRepo, &mockTransactor{}, audit)
		err := svc.Delete(context.Background(), doseClinicID, 50, model.MedicineDoseSpeciesDog, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Empty(t, audit.entries)
	})

	t.Run("不正speciesは拒否", func(t *testing.T) {
		svc := NewMedicineDoseParamService(&mockMedicineDoseParamRepository{}, perWeightMedicineRepo(), &mockTransactor{}, &mockTreatmentAuditTxLogger{})
		err := svc.Delete(context.Background(), doseClinicID, 50, model.MedicineDoseSpecies("rabbit"), nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("対象行なしはNotFound", func(t *testing.T) {
		repo := &mockMedicineDoseParamRepository{
			findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
				return nil, apperrors.WrapNotFound("medicine_dose_param", "")
			},
		}
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, &mockTreatmentAuditTxLogger{})
		err := svc.Delete(context.Background(), doseClinicID, 50, model.MedicineDoseSpeciesDog, nil)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("削除成功はsoft-delete+delete audit(before記録)", func(t *testing.T) {
		audit := &mockTreatmentAuditTxLogger{}
		deleted := false
		repo := &mockMedicineDoseParamRepository{
			findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
				return &model.MedicineDoseParam{ID: 300, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesCat, DosePerKg: 2, MaxMgPerKg: fptr(6)}, nil
			},
			deleteFn: func(_ context.Context, clinicID, id uint64) error {
				deleted = true
				assert.Equal(t, doseClinicID, clinicID)
				assert.Equal(t, uint64(300), id)
				return nil
			},
		}
		staffID := uint64(9)
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)
		err := svc.Delete(context.Background(), doseClinicID, 50, model.MedicineDoseSpeciesCat, &staffID)
		require.NoError(t, err)
		assert.True(t, deleted)
		require.Len(t, audit.entries, 1)
		e := audit.entries[0]
		assert.Equal(t, model.AuditActionMedicineDoseParamDelete, e.Action)
		assert.Equal(t, model.AuditResourceMedicineDoseParam, e.Resource)
		assert.NotNil(t, e.OldValue, "削除は before 記録")
		assert.Nil(t, e.NewValue, "削除は after なし")
	})

	// BE-refactor.md R1-2: audit（LogEntryTx）失敗は Delete 全体を失敗させる（fail-closed）。
	t.Run("audit失敗はDelete全体をロールバックする（fail-closed）", func(t *testing.T) {
		audit := &mockTreatmentAuditTxLogger{logEntryTxErr: errAuditWriteFailed}
		repo := &mockMedicineDoseParamRepository{
			findByMedicineAndSpeciesFn: func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
				return &model.MedicineDoseParam{ID: 402, ClinicID: doseClinicID, MedicineID: 50, Species: model.MedicineDoseSpeciesCat, DosePerKg: 2, MaxMgPerKg: fptr(6)}, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
		}
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, audit)
		err := svc.Delete(context.Background(), doseClinicID, 50, model.MedicineDoseSpeciesCat, nil)
		require.Error(t, err, "audit失敗は削除全体を失敗させる")
	})
}

// ---- List ----

func TestMedicineDoseParamService_List(t *testing.T) {
	t.Run("別clinicのmedicineはNotFound", func(t *testing.T) {
		medRepo := &mockMedicineRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Medicine, error) {
				return nil, apperrors.WrapNotFound("medicine", fmt.Sprintf("%d", id))
			},
		}
		svc := NewMedicineDoseParamService(&mockMedicineDoseParamRepository{}, medRepo, &mockTransactor{}, &mockTreatmentAuditTxLogger{})
		_, err := svc.List(context.Background(), doseClinicID, 50)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("所有medicineのparam一覧を返す", func(t *testing.T) {
		repo := &mockMedicineDoseParamRepository{
			findByMedicineIDFn: func(_ context.Context, _, _ uint64) ([]model.MedicineDoseParam, error) {
				return []model.MedicineDoseParam{
					{ID: 1, Species: model.MedicineDoseSpeciesDog},
					{ID: 2, Species: model.MedicineDoseSpeciesCat},
				}, nil
			},
		}
		svc := NewMedicineDoseParamService(repo, perWeightMedicineRepo(), &mockTransactor{}, &mockTreatmentAuditTxLogger{})
		got, err := svc.List(context.Background(), doseClinicID, 50)
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})
}
