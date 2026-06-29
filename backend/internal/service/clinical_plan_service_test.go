package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- ClinicalPlan モック ----

type mockClinicalPlanRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error)
	createFn                func(ctx context.Context, plan *model.ClinicalPlan) error
	updateFn                func(ctx context.Context, clinicID, planID uint64, fields map[string]any) error
	deleteFn                func(ctx context.Context, clinicID, planID uint64) error
}

func (m *mockClinicalPlanRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.ClinicalPlan, error) {
	return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockClinicalPlanRepository) Create(ctx context.Context, plan *model.ClinicalPlan) error {
	return m.createFn(ctx, plan)
}

func (m *mockClinicalPlanRepository) Update(ctx context.Context, clinicID, planID uint64, fields map[string]any) error {
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
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo())

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
			input:           &UpdateClinicalPlanInput{},
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
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo())

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
			svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo())

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
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo())

	plan, err := svc.GetOrCreate(context.Background(), clinicA, clinicBMRID)

	assert.Error(t, err, "clinic A から clinic B のカルテへの clinical_plan 自動作成は拒否されるべき")
	assert.True(t, apperrors.IsNotFound(err), "拒否は NotFound(404) にマップされるべき: %v", err)
	assert.Nil(t, plan)
	assert.False(t, createCalled, "親カルテ所有権検証に失敗した場合、clinical_plan を永続化してはならない")
}
