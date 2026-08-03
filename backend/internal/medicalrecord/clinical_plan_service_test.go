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
	updateFn                func(ctx context.Context, clinicID, planID uint64, fields map[string]any) error
	updateWithVersionFn     func(ctx context.Context, clinicID, planID uint64, fields map[string]any, expectedVersion *int) error
	deleteFn                func(ctx context.Context, clinicID, planID uint64) error
}

func (m *mockClinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockClinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockClinicalPlanRepository) Update(ctx context.Context, clinicID, planID uint64, fields map[string]any, expectedVersion *int) error {
	if m.updateWithVersionFn != nil {
		return m.updateWithVersionFn(ctx, clinicID, planID, fields, expectedVersion)
	}
	return m.updateFn(ctx, clinicID, planID, fields)
}

func (m *mockClinicalPlanRepository) Delete(ctx context.Context, clinicID, planID uint64) error {
	return m.deleteFn(ctx, clinicID, planID)
}

// okMedRecForPlan は親カルテの所有権検証が成功する（同一クリニック）モックを返す。
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
	diagnosisTypeID := uint64(1)
	diagnosisNameID := uint64(2)
	diagnosis2TypeID := uint64(3)
	diagnosis2NameID := uint64(4)
	diagnosisDetails := "details"
	treatmentPolicy := "policy"
	diagnosis2TypePtr := &diagnosis2TypeID
	diagnosis2NamePtr := &diagnosis2NameID
	var nilDiagnosis2Type *uint64
	var nilDiagnosis2Name *uint64

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
			name: "all fields set are reflected with real column names",
			input: &UpdateClinicalPlanInput{
				PhysicalExam:     &physicalExam,
				DiagnosisTypeID:  &diagnosisTypeID,
				DiagnosisNameID:  &diagnosisNameID,
				Diagnosis2TypeID: &diagnosis2TypePtr,
				Diagnosis2NameID: &diagnosis2NamePtr,
				DiagnosisDetails: &diagnosisDetails,
				TreatmentPolicy:  &treatmentPolicy,
				ActorID:          clinicalPlanTestActorID(),
			},
			wantFields: map[string]any{
				"physical_exam":       physicalExam,
				"diagnosis_type_id":   diagnosisTypeID,
				"diagnosis_name_id":   diagnosisNameID,
				"diagnosis_2_type_id": diagnosis2TypePtr, // *uint64（NULLクリアと同形）
				"diagnosis_2_name_id": diagnosis2NamePtr,
				"diagnosis_details":   diagnosisDetails,
				"treatment_policy":    treatmentPolicy,
			},
		},
		{
			name: "null diagnosis_2 fields clear to NULL",
			input: &UpdateClinicalPlanInput{
				Diagnosis2TypeID: &nilDiagnosis2Type,
				Diagnosis2NameID: &nilDiagnosis2Name,
				ActorID:          clinicalPlanTestActorID(),
			},
			wantFields: map[string]any{
				"diagnosis_2_type_id": nilDiagnosis2Type,
				"diagnosis_2_name_id": nilDiagnosis2Name,
			},
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
				assert.Equal(t, tt.medicalRecordID, plan.MedicalRecordID)
			}
		})
	}
}

func TestClinicalPlanService_Update(t *testing.T) {
	physicalExam := "Abnormal findings"
	diagnosisTypeID := uint64(1)
	diagnosisNameID := uint64(5)
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
				PhysicalExam:    &physicalExam,
				DiagnosisTypeID: &diagnosisTypeID,
				ActorID:         clinicalPlanTestActorID(),
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
				DiagnosisTypeID: &diagnosisTypeID,
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
				DiagnosisTypeID:  &diagnosisTypeID,
				DiagnosisNameID:  &diagnosisNameID,
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
				DiagnosisTypeID:  &diagnosisTypeID,
				DiagnosisNameID:  &diagnosisNameID,
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
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
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
	var gotFields map[string]any
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
		updateWithVersionFn: func(_ context.Context, _, _ uint64, fields map[string]any, expectedVersion *int) error {
			gotFields = fields
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
	require.NotNil(t, gotFields)
	assert.Equal(t, "", gotFields["physical_exam"])
	assert.Equal(t, "", gotFields["diagnosis_details"])
	assert.Equal(t, "", gotFields["treatment_policy"])
	assert.Equal(t, 3, gotFields["version"])
}

// BUG-010: 確定済み親カルテへの clinical plan 更新は Conflict で拒否し repo.Update を呼ばない。
func TestClinicalPlanService_Update_FinalizedParentRejected(t *testing.T) {
	physicalExam := "後から書いた所見"
	updateCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "確定前"}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, plan)
	assert.False(t, updateCalled)
}

// TestClinicalPlanService_Update_ValidateDiagnosisFKsError は、貼り替え先の診断マスタFKが
// 呼び出し元クリニックの所有でない場合に Update がエラーを返し、repo.Update が呼ばれない
// ことを検証する（クロステナント write 防止）。
func TestClinicalPlanService_Update_ValidateDiagnosisFKsError(t *testing.T) {
	foreignTypeID := uint64(999)
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			t.Fatal("clinical plan must not be updated when diagnosis FK ownership check fails")
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

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{DiagnosisTypeID: &foreignTypeID,
		ActorID: clinicalPlanTestActorID(),
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
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
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

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID)

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

// TestClinicalPlanService_Update_RejectsFinalizedParent は確定済みカルテの clinical_plan
// 更新が Conflict(409) で拒否され、repo.Update が呼ばれないことを検証する（SD-2 系ガード監査）。
func TestClinicalPlanService_Update_RejectsFinalizedParent(t *testing.T) {
	updateCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})
	physicalExam := "test"

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの clinical_plan 更新は Conflict(409) であるべき: %v", err)
	assert.False(t, updateCalled, "確定済みカルテの所見・診断は更新されてはならない")
}

// TestClinicalPlanService_Delete_RejectsFinalizedParent は確定済みカルテの clinical_plan
// 削除が Conflict(409) で拒否され、repo.Delete が呼ばれないことを検証する。
func TestClinicalPlanService_Delete_RejectsFinalizedParent(t *testing.T) {
	deleteCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})

	err := svc.Delete(context.Background(), 1, 1)

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err), "確定済みカルテの clinical_plan 削除は Conflict(409) であるべき: %v", err)
	assert.False(t, deleteCalled, "確定済みカルテの所見・診断は削除されてはならない")
}

func TestClinicalPlanService_Update_RejectsMismatchedDiagnosis2TypeName(t *testing.T) {
	typeID := uint64(10)
	nameID := uint64(20)
	typePtr := &typeID
	namePtr := &nameID
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			t.Fatal("must not update when diagnosis_2 type/name mismatch")
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
		Diagnosis2TypeID: &typePtr,
		Diagnosis2NameID: &namePtr,
		ActorID:          clinicalPlanTestActorID(),
	})
	assert.Error(t, err)
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
		var gotFields map[string]any
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, Version: 3}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, fields map[string]any, expectedVersion *int) error {
				gotExpectedVersion = expectedVersion
				gotVersionCaptured = true
				gotFields = fields
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
		assert.Equal(t, 4, gotFields["version"], "fields[version] must be plan.Version+1")
	})

	t.Run("input.Version が非nilのとき、その値が expectedVersion にそのまま渡り fields[version] は *input.Version+1", func(t *testing.T) {
		claimedVersion := 5
		var gotExpectedVersion *int
		var gotFields map[string]any
		repo := &mockClinicalPlanRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
				// plan.Version はサーバ側の実際の値（claimedVersion と異なっていても
				// nextVersion 計算は input.Version 起点になることを示すため意図的にずらす）。
				return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, Version: 99}, nil
			},
			updateWithVersionFn: func(_ context.Context, _, _ uint64, fields map[string]any, expectedVersion *int) error {
				gotExpectedVersion = expectedVersion
				gotFields = fields
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
		assert.Equal(t, claimedVersion+1, gotFields["version"], "fields[version] must be *input.Version+1")
	})
}
