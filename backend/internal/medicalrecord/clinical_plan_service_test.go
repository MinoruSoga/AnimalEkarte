package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ClinicalPlan モック ----

// updateFn は既存呼び出し（cross_tenant_master_fk_write_test.go を含む）との互換のため
// 4引数のまま維持する（BUG-416③: そのファイルは他セッションが編集中のため変更不可）。
// expectedVersion を検証したい新規テストは updateWithVersionFn を設定する（下記 Update メソッド参照）。
type mockClinicalPlanRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	createFn                func(ctx context.Context, plan *model.ClinicalPlan) error
	updateFn                func(ctx context.Context, clinicID, planID uint64, cmd UpdateClinicalPlanInput) error
	updateWithVersionFn     func(ctx context.Context, clinicID, planID uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error
	deleteFn                func(ctx context.Context, clinicID, planID uint64) error
	listDiagnosesFn         func(ctx context.Context, clinicID, planID uint64) ([]model.ClinicalPlanDiagnosis, error)
	replaceDiagnosesFn      func(ctx context.Context, clinicID, planID uint64, rows []ClinicalPlanDiagnosisInput) error
	deleteDiagnosesFn       func(ctx context.Context, clinicID, planID uint64) error
}

// ListDiagnoses は listDiagnosesFn 未設定のとき診断行なしを返す。
func (m *mockClinicalPlanRepository) ListDiagnoses(ctx context.Context, clinicID, planID uint64) ([]model.ClinicalPlanDiagnosis, error) {
	if m.listDiagnosesFn != nil {
		return m.listDiagnosesFn(ctx, clinicID, planID)
	}
	return nil, nil
}

// ReplaceDiagnoses は replaceDiagnosesFn 未設定のとき成功する（書込なし）。
func (m *mockClinicalPlanRepository) ReplaceDiagnoses(ctx context.Context, clinicID, planID uint64, rows []ClinicalPlanDiagnosisInput) error {
	if m.replaceDiagnosesFn != nil {
		return m.replaceDiagnosesFn(ctx, clinicID, planID, rows)
	}
	return nil
}

// DeleteDiagnoses は deleteDiagnosesFn 未設定のとき成功する（書込なし）。
func (m *mockClinicalPlanRepository) DeleteDiagnoses(ctx context.Context, clinicID, planID uint64) error {
	if m.deleteDiagnosesFn != nil {
		return m.deleteDiagnosesFn(ctx, clinicID, planID)
	}
	return nil
}

func (m *mockClinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockClinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockClinicalPlanRepository) Update(ctx context.Context, clinicID, planID uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
	if m.updateWithVersionFn != nil {
		return m.updateWithVersionFn(ctx, clinicID, planID, cmd, expectedVersion)
	}
	return m.updateFn(ctx, clinicID, planID, cmd)
}

func (m *mockClinicalPlanRepository) Delete(ctx context.Context, clinicID, planID uint64) error {
	return m.deleteFn(ctx, clinicID, planID)
}

// okMedRecForPlan は親カルテの所有権検証が成功する（同一医院）モックを返す。
func okMedRecForPlan() *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{}, nil
		},
	}
}

// ---- Tests ----

func TestBuildClinicalPlanUpdate(t *testing.T) {
	physicalExam := "Normal"
	diagnosisDetails := "details"
	treatmentPolicy := "policy"
	nextVersion := 4

	tests := []struct {
		name       string
		input      *UpdateClinicalPlanInput
		wantFields map[string]any
	}{
		{
			name: "empty input returns empty map",
			input: &UpdateClinicalPlanInput{
				ActorID: clinicalPlanTestActorID(),
			},
			wantFields: map[string]any{},
		},
		{
			name: "all scalar fields set are reflected with real column names",
			input: &UpdateClinicalPlanInput{
				PhysicalExam:     &physicalExam,
				DiagnosisDetails: &diagnosisDetails,
				TreatmentPolicy:  &treatmentPolicy,
				persistVersion:   &nextVersion,
				ActorID:          clinicalPlanTestActorID(),
			},
			wantFields: map[string]any{
				"physical_exam":     physicalExam,
				"diagnosis_details": diagnosisDetails,
				"treatment_policy":  treatmentPolicy,
				"version":           nextVersion,
			},
		},
		{
			// 診断は clinical_plan_diagnoses の行で保持する（EMR-495 A）。plan 本体の列更新には含めない。
			name: "diagnosis rows are not part of the plan column update",
			input: &UpdateClinicalPlanInput{
				Diagnoses: &[]ClinicalPlanDiagnosisInput{{Slot: 1}},
				ActorID:   clinicalPlanTestActorID(),
			},
			wantFields: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildClinicalPlanUpdate(tt.input)
			assert.Equal(t, tt.wantFields, fields)
		})
	}
}

func TestClinicalPlanService_GetOrCreate(t *testing.T) {
	existingPlan := &model.ClinicalPlan{
		ID:              1,
		MedicalRecordID: 1,
		PhysicalExam:    "Normal",
	}

	tests := []struct {
		name              string
		medicalRecordID   uint64
		findByRecordIDErr error
		repoCreatErr      error
		repoCreatePlan    *model.ClinicalPlan
		wantErr           bool
		wantCreate        bool
	}{
		{
			name:              "returns existing plan when found",
			medicalRecordID:   1,
			findByRecordIDErr: nil,
			repoCreatErr:      nil,
			repoCreatePlan:    existingPlan,
			wantErr:           false,
			wantCreate:        false,
		},
		{
			name:              "creates plan when not found",
			medicalRecordID:   1,
			findByRecordIDErr: apperrors.WrapNotFound("clinical_plan", "1"),
			repoCreatErr:      nil,
			repoCreatePlan: &model.ClinicalPlan{
				ID:              2,
				MedicalRecordID: 1,
			},
			wantErr:    false,
			wantCreate: true,
		},
		{
			name:              "returns error on non-notfound error",
			medicalRecordID:   1,
			findByRecordIDErr: errors.New("db error"),
			repoCreatErr:      nil,
			wantErr:           true,
			wantCreate:        false,
		},
		{
			name:              "returns error when create fails",
			medicalRecordID:   1,
			findByRecordIDErr: apperrors.WrapNotFound("clinical_plan", "1"),
			repoCreatErr:      errors.New("db create error"),
			wantErr:           true,
			wantCreate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicalPlanRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
					if tt.findByRecordIDErr != nil {
						return nil, tt.findByRecordIDErr
					}
					return existingPlan, nil
				},
				createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
					return tt.repoCreatErr
				},
			}
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

			plan, err := svc.GetOrCreate(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.medicalRecordID, plan.Plan.MedicalRecordID)
			}
		})
	}
}

func TestClinicalPlanService_Update(t *testing.T) {
	physicalExam := "Abnormal findings"
	diagnosisDetails := "Suspected infection"
	treatmentPolicy := "Prescribe antibiotics"

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *UpdateClinicalPlanInput
		repoFindPlan    *model.ClinicalPlan
		repoFindErr     error
		repoUpdateErr   error
		repoReturnPlan  *model.ClinicalPlan
		wantErr         bool
	}{
		{
			name:            "updates clinical plan successfully",
			medicalRecordID: 1,
			input: &UpdateClinicalPlanInput{
				PhysicalExam: &physicalExam,
				ActorID:      clinicalPlanTestActorID(),
			},
			repoFindPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
				PhysicalExam:    physicalExam,
			},
			wantErr: false,
		},
		{
			name:            "returns current plan when no fields provided (no-op)",
			medicalRecordID: 1,
			input: &UpdateClinicalPlanInput{
				ActorID: clinicalPlanTestActorID(),
			},
			repoFindPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			wantErr: false,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			input: &UpdateClinicalPlanInput{
				DiagnosisDetails: &diagnosisDetails,
				ActorID:          clinicalPlanTestActorID(),
			},
			repoFindPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
		{
			name:            "updates multiple fields",
			medicalRecordID: 1,
			input: &UpdateClinicalPlanInput{
				PhysicalExam:     &physicalExam,
				DiagnosisDetails: &diagnosisDetails,
				TreatmentPolicy:  &treatmentPolicy,
				ActorID:          clinicalPlanTestActorID(),
			},
			repoFindPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnPlan: &model.ClinicalPlan{
				ID:               1,
				MedicalRecordID:  1,
				PhysicalExam:     physicalExam,
				DiagnosisDetails: diagnosisDetails,
				TreatmentPolicy:  treatmentPolicy,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicalPlanRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoFindPlan, nil
				},
				createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
					return nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

			plan, err := svc.Update(context.Background(), 1, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}
		})
	}
}

// BUG-010: 空文字ポインタは明示クリアとして repo.Update に渡り、未送信 nil は no-op になる。
func TestClinicalPlanService_Update_EmptyStringClearsFields(t *testing.T) {
	empty := ""
	var got UpdateClinicalPlanInput
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{
				ID:               1,
				MedicalRecordID:  1,
				PhysicalExam:     "残すべきでない所見",
				DiagnosisDetails: "残すべきでない診断",
				TreatmentPolicy:  "残すべきでない方針",
				Version:          2,
			}, nil
		},
		updateWithVersionFn: func(_ context.Context, _, _ uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
			got = cmd
			if expectedVersion == nil || *expectedVersion != 2 {
				t.Fatalf("expectedVersion = %v, want 2", expectedVersion)
			}
			return nil
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})
	version := 2
	_, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{
		PhysicalExam:     &empty,
		DiagnosisDetails: &empty,
		TreatmentPolicy:  &empty,
		Version:          &version,
		ActorID:          clinicalPlanTestActorID(),
	})
	require.NoError(t, err)
	require.NotNil(t, got.PhysicalExam)
	require.NotNil(t, got.DiagnosisDetails)
	require.NotNil(t, got.TreatmentPolicy)
	require.NotNil(t, got.persistVersion)
	assert.Equal(t, "", *got.PhysicalExam)
	assert.Equal(t, "", *got.DiagnosisDetails)
	assert.Equal(t, "", *got.TreatmentPolicy)
	assert.Equal(t, 3, *got.persistVersion)
}

// TestClinicalPlanService_Update_ValidateDiagnosisFKsError は、貼り替え先の診断マスタFKが
// 呼び出し元医院の所有でない場合に Update がエラーを返し、repo.Update が呼ばれない
// ことを検証する（クロステナント write 防止）。
func TestClinicalPlanService_Update_ValidateDiagnosisFKsError(t *testing.T) {
	foreignTypeID := uint64(999)
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
			t.Fatal("clinical plan must not be updated when diagnosis FK ownership check fails")
			return nil
		},
		replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
			t.Fatal("diagnosis rows must not be written when diagnosis FK ownership check fails")
			return nil
		},
	}
	diagTypeRepo := &mockDiagnosisTypeRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
			if id != foreignTypeID {
				return &model.DiagnosisType{ID: id}, nil
			}
			return nil, apperrors.WrapNotFound("diagnosis_type", "foreign")
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), diagTypeRepo, okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{
		Diagnoses: &[]ClinicalPlanDiagnosisInput{{Slot: 1, DiagnosisTypeID: &foreignTypeID}},
		ActorID:   clinicalPlanTestActorID(),
	})

	assert.Error(t, err)
	assert.Nil(t, plan)
}

// TestClinicalPlanService_Update_GetOrCreateError は、Update が内部で呼ぶ GetOrCreate が
// 失敗した場合にそのエラーがラップされて伝播することを検証する。
func TestClinicalPlanService_Update_GetOrCreateError(t *testing.T) {
	physicalExam := "Normal"
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return nil, apperrors.WrapNotFound("clinical_plan", "1")
		},
		createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
			return errors.New("db create error")
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})

	assert.Error(t, err)
	assert.Nil(t, plan)
}

// TestClinicalPlanService_Update_RefetchError は、更新自体は成功したが更新後の
// FindByMedicalRecordID による再取得が失敗した場合にエラーが伝播することを検証する。
func TestClinicalPlanService_Update_RefetchError(t *testing.T) {
	physicalExam := "Normal"
	findCalls := 0
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			findCalls++
			if findCalls == 1 {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
			}
			return nil, errors.New("db error")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
			return nil
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Equal(t, 2, findCalls, "更新後の再取得も呼ばれる")
}

func TestClinicalPlanService_Delete(t *testing.T) {
	tests := []struct {
		name              string
		medicalRecordID   uint64
		repoPlan          *model.ClinicalPlan
		findByRecordIDErr error
		deleteErr         error
		wantErr           bool
	}{
		{
			name:            "deletes clinical plan successfully",
			medicalRecordID: 1,
			repoPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByRecordIDErr: nil,
			deleteErr:         nil,
			wantErr:           false,
		},
		{
			name:              "returns error when plan not found",
			medicalRecordID:   1,
			repoPlan:          nil,
			findByRecordIDErr: errors.New("not found"),
			deleteErr:         nil,
			wantErr:           true,
		},
		{
			name:            "returns error when delete fails",
			medicalRecordID: 1,
			repoPlan: &model.ClinicalPlan{
				ID:              1,
				MedicalRecordID: 1,
			},
			findByRecordIDErr: nil,
			deleteErr:         errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClinicalPlanRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
					return tt.repoPlan, tt.findByRecordIDErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.deleteErr
				},
			}
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, clinicalPlanTestActorID())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClinicalPlanService_GetOrCreate_CrossTenantParentRejected は
// clinic A の呼び出しが clinic B のカルテを参照する自動作成（GetOrCreate）を拒否し、
// clinical_plan（自前 clinic_id を持たない）が clinic B のカルテ配下に
// 永続化されない（repo.Create が呼ばれない）ことを検証する。
// 親所有権検証（medRec.FindByID(clinicID, medRecID)）を削除すると必ず失敗する回帰テスト。
func TestClinicalPlanService_GetOrCreate_CrossTenantParentRejected(t *testing.T) {
	const (
		clinicA     = uint64(1)
		clinicBMRID = uint64(99) // clinic B のカルテ ID（clinic A は所有しない）
	)
	createCalled := false
	repo := &mockClinicalPlanRepository{
		// clinic A のスコープでは clinic B のカルテのプランは見えない → NotFound（作成分岐へ）。
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return nil, apperrors.WrapNotFound("clinical_plan", "99")
		},
		createFn: func(_ context.Context, _ *model.ClinicalPlan) error {
			createCalled = true
			return nil
		},
	}
	// 親カルテも clinic A の所有ではない → NotFound（クロステナント）。
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, apperrors.WrapNotFound("medical_record", "99")
		},
	}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	plan, err := svc.GetOrCreate(context.Background(), clinicA, clinicBMRID)

	assert.Error(t, err, "clinic A から clinic B のカルテへの clinical_plan 自動作成は拒否されるべき")
	assert.True(t, apperrors.IsNotFound(err), "拒否は NotFound(404) にマップされるべき: %v", err)
	assert.Nil(t, plan)
	assert.False(t, createCalled, "親カルテ所有権検証に失敗した場合、clinical_plan を永続化してはならない")
}

// EMR-495 P1（契約 §0.1）: 親カルテの旧 finalized 拒否（旧 Update_FinalizedParentRejected /
// Update_RejectsFinalizedParent / Delete_RejectsFinalizedParent）は廃止した。medical_records.status は
// 001 で廃止され、model.MedicalRecord に Status もない。親行ロック（lockDraftMedicalRecord）は維持され、
// 更新は plan の version CAS・監査・診断行 FK 検査で守られる（下記の各テスト）。

func TestClinicalPlanService_Update_RejectsMismatchedDiagnosisTypeName(t *testing.T) {
	typeID := uint64(10)
	nameID := uint64(20)
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateClinicalPlanInput) error {
			t.Fatal("must not update when diagnosis type/name mismatch")
			return nil
		},
		replaceDiagnosesFn: func(_ context.Context, _, _ uint64, _ []ClinicalPlanDiagnosisInput) error {
			t.Fatal("must not write diagnosis rows when diagnosis type/name mismatch")
			return nil
		},
	}
	diagTypeRepo := &mockDiagnosisTypeRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisType, error) {
			return &model.DiagnosisType{ID: id}, nil
		},
	}
	diagNameRepo := &mockDiagnosisNameRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
			return &model.DiagnosisName{ID: id, DiagnosisTypeID: 99}, nil
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), diagTypeRepo, diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})
	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{
		Diagnoses: &[]ClinicalPlanDiagnosisInput{{Slot: 2, DiagnosisTypeID: &typeID, DiagnosisNameID: &nameID}},
		ActorID:   clinicalPlanTestActorID(),
	})
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, plan)
}

// TestClinicalPlanService_Update_VersionPassthrough は BUG-416③ の楽観的ロックについて、
// (a) input.Version がそのまま（nextVersion ではなく）repo.Update の expectedVersion に渡ること、
// (b) fields["version"] が input.Version==nil のとき plan.Version+1、非nilのとき *input.Version+1
// になること（medical_record_crud.go の Update と同型の二分岐）を検証する。
func TestClinicalPlanService_Update_VersionPassthrough(t *testing.T) {
	physicalExam := "test"

	t.Run("input.Version が nil のとき plan.Version+1 を使い、expectedVersion は nil のまま渡る", func(t *testing.T) {
		var gotExpectedVersion *int
		var gotVersionCaptured bool
		var got UpdateClinicalPlanInput
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, Version: 3}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
				gotExpectedVersion = expectedVersion
				gotVersionCaptured = true
				got = cmd
				return nil
			},
		}
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

		plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
			ActorID: clinicalPlanTestActorID(),
		})

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		require.True(t, gotVersionCaptured, "repo.Update must be called")
		assert.Nil(t, gotExpectedVersion, "expectedVersion must be nil when input.Version is nil")
		require.NotNil(t, got.persistVersion)
		assert.Equal(t, 4, *got.persistVersion, "persistVersion must be plan.Version+1")
	})

	t.Run("input.Version が非nilのとき、その値が expectedVersion にそのまま渡り fields[version] は *input.Version+1", func(t *testing.T) {
		claimedVersion := 5
		var gotExpectedVersion *int
		var got UpdateClinicalPlanInput
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				// plan.Version はサーバ側の実際の値（claimedVersion と異なっていても
				// nextVersion 計算は input.Version 起点になることを示すため意図的にずらす）。
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, Version: 99}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, cmd UpdateClinicalPlanInput, expectedVersion *int) error {
				gotExpectedVersion = expectedVersion
				got = cmd
				return nil
			},
		}
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

		plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{
			PhysicalExam: &physicalExam,
			Version:      &claimedVersion,
			ActorID:      clinicalPlanTestActorID(),
		})

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		require.NotNil(t, gotExpectedVersion)
		assert.Equal(t, claimedVersion, *gotExpectedVersion, "expectedVersion must be input.Version itself, not nextVersion")
		require.NotNil(t, got.persistVersion)
		assert.Equal(t, claimedVersion+1, *got.persistVersion, "persistVersion must be *input.Version+1")
	})
}

// EMR-495 A: 診断行の書込（置換・version・件数 guard・clinic_id）。
func TestClinicalPlanService_Update_DiagnosisRows(t *testing.T) {
	typeID := uint64(10)
	nameID := uint64(20)
	snap := "snapshot-name"
	cat := "snapshot-category"
	rows := []ClinicalPlanDiagnosisInput{
		{Slot: 1, DiagnosisTypeID: &typeID, DiagnosisNameID: &nameID, DiagnosisNameSnapshot: &snap, CategoryNameSnapshot: &cat, SortOrder: 0},
		{Slot: 2, DiagnosisNameSnapshot: &snap},
	}
	diagNameRepo := &mockDiagnosisNameRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.DiagnosisName, error) {
			return &model.DiagnosisName{ID: id, DiagnosisTypeID: typeID}, nil
		},
	}
	newRepo := func(updateCmd *UpdateClinicalPlanInput, replaced *[][]ClinicalPlanDiagnosisInput) *mockClinicalPlanRepository {
		return &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 7, MedicalRecordID: 1, Version: 2}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, cmd UpdateClinicalPlanInput, _ *int) error {
				*updateCmd = cmd
				return nil
			},
			replaceDiagnosesFn: func(_ context.Context, clinicID, planID uint64, in []ClinicalPlanDiagnosisInput) error {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(7), planID)
				*replaced = append(*replaced, in)
				return nil
			},
		}
	}

	t.Run("diagnosis-only PATCH replaces rows and bumps the plan version", func(t *testing.T) {
		var cmd UpdateClinicalPlanInput
		var replaced [][]ClinicalPlanDiagnosisInput
		repo := newRepo(&cmd, &replaced)
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})

		view, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{Diagnoses: &rows, ActorID: clinicalPlanTestActorID()})

		require.NoError(t, err)
		require.NotNil(t, view)
		require.Len(t, replaced, 1)
		assert.Equal(t, rows, replaced[0], "snapshots are stored as sent, never derived from the current master")
		require.NotNil(t, cmd.persistVersion)
		assert.Equal(t, 3, *cmd.persistVersion, "Q9: a diagnosis-only PATCH advances the plan version")
	})

	t.Run("empty array clears all rows", func(t *testing.T) {
		var cmd UpdateClinicalPlanInput
		var replaced [][]ClinicalPlanDiagnosisInput
		repo := newRepo(&cmd, &replaced)
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})
		empty := []ClinicalPlanDiagnosisInput{}

		_, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{Diagnoses: &empty, ActorID: clinicalPlanTestActorID()})

		require.NoError(t, err)
		require.Len(t, replaced, 1)
		assert.Empty(t, replaced[0])
	})

	t.Run("omitted diagnoses leaves rows untouched", func(t *testing.T) {
		var cmd UpdateClinicalPlanInput
		var replaced [][]ClinicalPlanDiagnosisInput
		repo := newRepo(&cmd, &replaced)
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})
		exam := "x"

		_, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &exam, ActorID: clinicalPlanTestActorID()})

		require.NoError(t, err)
		assert.Empty(t, replaced)
	})

	t.Run("audit carries the diagnosis rows", func(t *testing.T) {
		var cmd UpdateClinicalPlanInput
		var replaced [][]ClinicalPlanDiagnosisInput
		repo := newRepo(&cmd, &replaced)
		repo.listDiagnosesFn = func(_ context.Context, _, _ uint64) ([]model.ClinicalPlanDiagnosis, error) {
			return []model.ClinicalPlanDiagnosis{{ID: 5, ClinicalPlanID: 7, DiagnosisSlot: 1, DiagnosisNameSnapshot: &snap}}, nil
		}
		var entry *AuditEntry
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, e *AuditEntry) error {
			entry = e
			return nil
		}}
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, audit)

		_, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{Diagnoses: &rows, ActorID: clinicalPlanTestActorID()})

		require.NoError(t, err)
		require.NotNil(t, entry)
		oldValue, ok := entry.OldValue.(map[string]any)
		require.True(t, ok)
		oldDiagnoses, ok := oldValue["diagnoses"].([]map[string]any)
		require.True(t, ok)
		require.Len(t, oldDiagnoses, 1)
		assert.Equal(t, uint64(5), oldDiagnoses[0]["id"])
		newValue, ok := entry.NewValue.(map[string]any)
		require.True(t, ok)
		_, ok = newValue["diagnoses"].([]map[string]any)
		assert.True(t, ok)
	})

	t.Run("more than the row cap is rejected before any repository call", func(t *testing.T) {
		repo := &mockClinicalPlanRepository{}
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})
		tooMany := make([]ClinicalPlanDiagnosisInput, maxClinicalPlanDiagnosisRows+1)
		for i := range tooMany {
			tooMany[i] = ClinicalPlanDiagnosisInput{Slot: 1, SortOrder: i}
		}

		view, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{Diagnoses: &tooMany, ActorID: clinicalPlanTestActorID()})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, view)
	})

	t.Run("slot outside 1..2 is rejected", func(t *testing.T) {
		repo := &mockClinicalPlanRepository{}
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), diagNameRepo, &mockCheckupTransactor{}, &mockAuditTxLogger{})
		bad := []ClinicalPlanDiagnosisInput{{Slot: 3}}

		_, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{Diagnoses: &bad, ActorID: clinicalPlanTestActorID()})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})
}

// EMR-495 Finding 5: clinical_plans.clinic_id は NOT NULL。GetOrCreate が信頼できる clinic を設定して作成する。
func TestClinicalPlanService_GetOrCreate_SetsClinicID(t *testing.T) {
	var created *model.ClinicalPlan
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return nil, apperrors.WrapNotFound("clinical_plan", "1")
		},
		createFn: func(_ context.Context, plan *model.ClinicalPlan) error {
			created = plan
			plan.ID = 9
			return nil
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	view, err := svc.GetOrCreate(context.Background(), 5, 1)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, uint64(5), created.ClinicID)
	assert.Equal(t, uint64(9), view.Plan.ID)
}

// Delete は診断行の論理削除を plan 削除と同じ経路で行う。
func TestClinicalPlanService_Delete_DeletesDiagnosisRows(t *testing.T) {
	var order []string
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 7, MedicalRecordID: 1}, nil
		},
		deleteDiagnosesFn: func(_ context.Context, _, planID uint64) error {
			assert.Equal(t, uint64(7), planID)
			order = append(order, "diagnoses")
			return nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			order = append(order, "plan")
			return nil
		},
	}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())

	require.NoError(t, err)
	assert.Equal(t, []string{"diagnoses", "plan"}, order)
}
