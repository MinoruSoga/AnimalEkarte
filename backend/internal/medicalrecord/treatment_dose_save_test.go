package medicalrecord

// treatment_dose_save_test.go — #201 B-2: treatment 保存時 BE 再検証の統合テスト。
// species 解決→param 取得→保存時再検証→スナップショット永続化→逸脱 audit→fail-closed を mock で検証する。

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
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

	createCalls    int
	updateCalls    int
	inventoryCalls int
}

// rollbackAwareTransactor は mock で tx rollback をシミュレートする。
// fn が error を返したら create/update/inventory/audit の side-effect カウンタを実行前に戻す。
type rollbackAwareTransactor struct {
	f *doseSaveFixture
}

func (t *rollbackAwareTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	createBefore := t.f.createCalls
	updateBefore := t.f.updateCalls
	invBefore := t.f.inventoryCalls
	createdBefore := t.f.created
	var auditBefore []*AuditEntry
	if t.f.audit != nil {
		auditBefore = append([]*AuditEntry(nil), t.f.audit.entries...)
	}
	err := fn(ctx)
	if err != nil {
		t.f.createCalls = createBefore
		t.f.updateCalls = updateBefore
		t.f.inventoryCalls = invBefore
		t.f.created = createdBefore
		if t.f.audit != nil {
			t.f.audit.entries = auditBefore
		}
	}
	return err
}

// newSvc は個別依存注入コンストラクタ（BE9-2D ④b）で fixture の mock を配線する。
// WithTx は rollbackAwareTransactor で error 時 side-effect を巻き戻す。
// audit の typed-nil（*mockTreatmentAuditTxLogger(nil)）を interface の真の nil に正規化する。
func (f *doseSaveFixture) newSvc() TreatmentService {
	var audit AuditTxLogger
	if f.audit != nil {
		audit = f.audit
	}
	return NewTreatmentServiceWithAudit(
		f.treatRepo, f.mrRepo, f.medRepo, okProcedureRepo(), okConsultationRepo(), f.invRepo,
		f.vitalRepo, f.paramRepo, &rollbackAwareTransactor{f: f}, audit)
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
			f.createCalls++
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

func TestTreatmentDoseSave_Create(t *testing.T) {
	const clinicID = uint64(1)

	t.Run("上限ちょうどは保存できる", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2)) // 20mg == cap 20mg
		require.NoError(t, err)
		assert.Equal(t, 1, f.createCalls)
		require.NotNil(t, f.created)
		require.NotNil(t, f.created.DoseAmountMg, "スナップショット dose_amount_mg が永続化されること")
		assert.InDelta(t, 20.0, *f.created.DoseAmountMg, 1e-6)
		require.NotNil(t, f.created.DoseWeightKg)
		assert.InDelta(t, 4.0, *f.created.DoseWeightKg, 1e-6)
		assert.NotEmpty(t, f.created.DoseParamSnapshot, "dose_param_snapshot(jsonb) が値で固定されること")
		assert.Empty(t, f.audit.entries, "上限内・乖離なしでは逸脱 audit は発火しない")
	})

	t.Run("上限を safetyEpsilon より大きく超えると InvalidInput で拒否し永続化しない", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		svc := f.newSvc()
		qty := (20 + 2*safetyEpsilon) / 10

		got, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(qty))
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err), "上限超過は400へマップされる InvalidInput: %v", err)
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.ErrorContains(t, err, "上限")
		assert.Zero(t, f.createCalls, "拒否時は treatmentRepo.Create を呼ばない")
		assert.Nil(t, f.created)
		assert.Empty(t, f.audit.entries, "拒否された保存に逸脱 audit を書かない")
	})

	t.Run("BelowMinSaved のみは理由付きで保存でき audit を記録する", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		minMgPerKg := 4.5
		maxMgPerKg := 10.0
		f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{
				ID: 1, ClinicID: 1, MedicineID: 50,
				Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MinMgPerKg: &minMgPerKg, MaxMgPerKg: &maxMgPerKg,
			}, nil
		}
		svc := f.newSvc()

		in := medicineCreateInput(1.7)
		in.DoseDeviationReason = "漸増開始"
		_, err := svc.Create(context.Background(), clinicID, 100, in)
		require.NoError(t, err)
		assert.Equal(t, 1, f.createCalls)
		require.Len(t, f.audit.entries, 1)
		newValue, ok := f.audit.entries[0].NewValue.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, true, newValue["below_min"])
		assert.Equal(t, false, newValue["exceeds_max"])
		assert.Equal(t, "漸増開始", newValue["dose_deviation_reason"])
	})

	t.Run("DeviatesFromComputed のみは理由付きで保存でき audit を記録する", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		maxMgPerKg := 10.0
		f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{
				ID: 1, ClinicID: 1, MedicineID: 50,
				Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
			}, nil
		}
		svc := f.newSvc()

		in := medicineCreateInput(3)
		in.DoseDeviationReason = "臨床判断による増量"
		_, err := svc.Create(context.Background(), clinicID, 100, in)
		require.NoError(t, err)
		assert.Equal(t, 1, f.createCalls)
		require.Len(t, f.audit.entries, 1)
		newValue, ok := f.audit.entries[0].NewValue.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, false, newValue["below_min"])
		assert.Equal(t, false, newValue["exceeds_max"])
		assert.Equal(t, true, newValue["deviates_from_computed"])
		assert.Equal(t, "臨床判断による増量", newValue["dose_deviation_reason"])
	})

	t.Run("情報欠落時は従来どおり評価をスキップして保存する", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(*doseSaveFixture)
		}{
			{
				name: "体重なし",
				setup: func(f *doseSaveFixture) {
					f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
						return nil, nil
					}
				},
			},
			{
				name: "体重<=0 のみの vital は体重なしと同じ手動経路",
				setup: func(f *doseSaveFixture) {
					w := 0.0
					f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
						return []model.VitalRecord{
							{ID: 12, Weight: &w, WeightUnit: model.BodyWeightUnitKg, RecordedAt: time.Date(2026, 6, 28, 9, 0, 0, 0, time.UTC)},
						}, nil
					}
				},
			},
			{
				name: "当該種の dose param なし",
				setup: func(f *doseSaveFixture) {
					f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
						return nil, apperrors.WrapNotFound("medicine_dose_param", "50/dog")
					}
				},
			},
			{
				name: "種別を dose species へ正規化できない",
				setup: func(f *doseSaveFixture) {
					f.mrRepo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
						petID := uint64(7)
						return &model.MedicalRecord{
							Status: model.MedicalRecordStatusDraft,
							PetID:  &petID,
							Pet:    &model.Pet{AnimalSpecies: &model.AnimalSpecies{Name: "鳥"}},
						}, nil
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
				tt.setup(f)
				svc := f.newSvc()

				_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(5))
				require.NoError(t, err)
				assert.Equal(t, 1, f.createCalls)
				require.NotNil(t, f.created)
				assert.Nil(t, f.created.DoseAmountMg)
				assert.Empty(t, f.audit.entries)
			})
		}
	})

	t.Run("species 不一致は fail-closed（保存中止）", func(t *testing.T) {
		// pet=犬 だが param が cat を返す → 防御アサートが発火し保存を中止する。
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesCat)
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2))
		require.Error(t, err, "species 不一致では保存してはならない")
		assert.True(t, apperrors.IsConflict(err), "fail-closed は Conflict: %v", err)
		assert.Zero(t, f.createCalls)
	})

	t.Run("dose param 読込エラーは従来どおり fail-closed", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return nil, errors.New("dose param lookup failed")
		}
		svc := f.newSvc()

		_, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2))
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to load dose param")
		assert.Zero(t, f.createCalls)
	})

	// #201 P0: vital 読取のシステム障害は体重未記録（手動経路）と同一視せず、保存を中止する。
	t.Run("vital 読込エラーは fail-closed で保存中止する", func(t *testing.T) {
		f := newDoseSaveFixture(t, model.MedicineCalculationTypePerWeight, model.MedicineDoseSpeciesDog)
		f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
			return nil, errors.New("vital lookup failed")
		}
		svc := f.newSvc()

		got, err := svc.Create(context.Background(), clinicID, 100, medicineCreateInput(2))
		require.Error(t, err, "vital 読取エラーは dose 再検証スキップ（手動経路）にしてはならない")
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "failed to load vitals for dose weight")
		assert.Zero(t, f.createCalls, "失敗時は treatmentRepo.Create を呼ばない（tx rollback 相当）")
		assert.Empty(t, f.audit.entries)
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
		maxMgPerKg := 10.0
		f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{
				ID: 1, ClinicID: 1, MedicineID: 50,
				Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
			}, nil
		}
		f.audit.logEntryTxErr = errAuditWriteFailed
		svc := f.newSvc()

		in := medicineCreateInput(3) // 上限内の推奨値乖離
		in.DoseDeviationReason = "audit failure injection"
		_, err := svc.Create(context.Background(), clinicID, 100, in)
		require.Error(t, err, "audit 失敗は treatment 作成全体を失敗させる")
		assert.Zero(t, f.createCalls)
	})
}

// TestTreatmentDoseSave_Update は Update 経路の逸脱 audit（BE-refactor.md R1-2 で
// Create と同じ auditDoseDeviationTx 経路に統一）を固定する。quantity 変更で再評価対象になり、
// 逸脱時は audit 記録・audit 失敗時は Update 全体が fail-closed で失敗することを検証する。
func TestTreatmentDoseSave_Update(t *testing.T) {
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
			updateFn: func(_ context.Context, _, _ uint64, _ UpdateTreatmentInput) error {
				f.updateCalls++
				return nil
			},
		}
		return f
	}

	t.Run("上限を safetyEpsilon より大きく超える更新は InvalidInput で拒否し既存行を変えない", func(t *testing.T) {
		f := newUpdateFixture(t)
		svc := f.newSvc()

		qty := (20 + 2*safetyEpsilon) / 10
		got, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{Quantity: &qty})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
		status, _, _ := httpapi.ResolveErrorResponse(err)
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Zero(t, f.updateCalls)
		assert.Empty(t, f.audit.entries)
	})

	t.Run("下限未満・推奨値乖離は理由付きで保存継続、評価情報なしは理由なしで継続", func(t *testing.T) {
		tests := []struct {
			name      string
			quantity  float64
			setup     func(*doseSaveFixture)
			reason    *string
			wantAudit bool
		}{
			{
				name:     "BelowMinSaved",
				quantity: 1.7,
				setup: func(f *doseSaveFixture) {
					minMgPerKg := 4.5
					maxMgPerKg := 10.0
					f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
						return &model.MedicineDoseParam{
							ID: 1, ClinicID: 1, MedicineID: 50,
							Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
							DosePerKg: 5, MinMgPerKg: &minMgPerKg, MaxMgPerKg: &maxMgPerKg,
						}, nil
					}
				},
				reason:    strPtr("漸増"),
				wantAudit: true,
			},
			{
				name:     "DeviatesFromComputed",
				quantity: 3,
				setup: func(f *doseSaveFixture) {
					maxMgPerKg := 10.0
					f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
						return &model.MedicineDoseParam{
							ID: 1, ClinicID: 1, MedicineID: 50,
							Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
							DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
						}, nil
					}
				},
				reason:    strPtr("臨床判断"),
				wantAudit: true,
			},
			{
				name:     "体重なし",
				quantity: 5,
				setup: func(f *doseSaveFixture) {
					f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
						return nil, nil
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				f := newUpdateFixture(t)
				tt.setup(f)
				svc := f.newSvc()

				qty := tt.quantity
				actor := uint64(3)
				_, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{
					Quantity:            &qty,
					DoseDeviationReason: tt.reason,
					ActorID:             &actor,
				})
				require.NoError(t, err)
				assert.Equal(t, 1, f.updateCalls)
				if tt.wantAudit {
					assert.Len(t, f.audit.entries, 1)
				} else {
					assert.Empty(t, f.audit.entries)
				}
			})
		}
	})

	// #201 P0: Update 経路でも vital 読取障害は手動スキップにせず fail-closed。
	t.Run("vital 読込エラーは fail-closed で更新中止する", func(t *testing.T) {
		f := newUpdateFixture(t)
		f.vitalRepo.listByMedicalRecordIDFn = func(_ context.Context, _, _ uint64) ([]model.VitalRecord, error) {
			return nil, errors.New("vital lookup failed")
		}
		svc := f.newSvc()

		qty := 2.0
		got, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{Quantity: &qty})
		require.Error(t, err, "vital 読取エラーは dose 再検証スキップ（手動経路）にしてはならない")
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "failed to load vitals for dose weight")
		assert.Zero(t, f.updateCalls, "失敗時は treatmentRepo.Update を呼ばない（tx rollback 相当）")
		assert.Empty(t, f.audit.entries)
	})

	t.Run("親行ロック後の最新 treatment と部分 PATCH を合成して上限を再評価する", func(t *testing.T) {
		f := newUpdateFixture(t)
		medID := uint64(50)
		beforeLock := &model.Treatment{
			ID: treatmentID, MedicalRecordID: 100, ItemType: model.TreatmentItemTypeMedicine,
			MedicineID: &medID, Quantity: 1,
		}
		afterConcurrentUpdate := &model.Treatment{
			ID: treatmentID, MedicalRecordID: 100, ItemType: model.TreatmentItemTypeMedicine,
			MedicineID: &medID, Quantity: 5,
		}
		findCalls := 0
		f.treatRepo.findByIDFn = func(_ context.Context, _, _ uint64) (*model.Treatment, error) {
			findCalls++
			if findCalls == 1 {
				return beforeLock, nil
			}
			return afterConcurrentUpdate, nil
		}
		svc := f.newSvc()

		newMedicineID := uint64(51)
		got, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{MedicineID: &newMedicineID})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Equal(t, 2, findCalls, "事前確認後、親行ロック内で最新 treatment を再取得する")
		assert.Zero(t, f.updateCalls)
	})

	// BE-refactor.md R1-2: audit 失敗は Update 全体を fail-closed で失敗させる。
	t.Run("逸脱 audit 失敗は Update 全体をロールバックする（fail-closed）", func(t *testing.T) {
		f := newUpdateFixture(t)
		maxMgPerKg := 10.0
		f.paramRepo.findByMedicineAndSpeciesFn = func(_ context.Context, _, _ uint64, _ model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
			return &model.MedicineDoseParam{
				ID: 1, ClinicID: 1, MedicineID: 50,
				Species: model.MedicineDoseSpeciesDog, DoseBasis: model.MedicineDoseBasisPerAdministration,
				DosePerKg: 5, MaxMgPerKg: &maxMgPerKg,
			}, nil
		}
		f.audit.logEntryTxErr = errAuditWriteFailed
		svc := f.newSvc()

		qty := 3.0
		reason := "audit failure injection"
		actor := uint64(3)
		_, err := svc.Update(context.Background(), clinicID, 100, treatmentID, &UpdateTreatmentInput{
			Quantity: &qty, DoseDeviationReason: &reason, ActorID: &actor,
		})
		require.Error(t, err, "audit 失敗は treatment 更新全体を失敗させる")
		assert.Zero(t, f.updateCalls)
	})
}
