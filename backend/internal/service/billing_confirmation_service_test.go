package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- BillingConfirmation モック ----

type mockBillingConfirmationRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error)
	createFn                func(ctx context.Context, review *model.BillingConfirmation) error
	updateFn                func(ctx context.Context, clinicID, reviewID uint64, fields map[string]any) error
}

func (m *mockBillingConfirmationRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) (*model.BillingConfirmation, error) {
	return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockBillingConfirmationRepository) Create(ctx context.Context, review *model.BillingConfirmation) error {
	return m.createFn(ctx, review)
}

func (m *mockBillingConfirmationRepository) Update(ctx context.Context, clinicID, reviewID uint64, fields map[string]any) error {
	return m.updateFn(ctx, clinicID, reviewID, fields)
}

// okMedRecForBilling は親カルテの所有権検証が成功する（同一クリニック）モックを返す。
func okMedRecForBilling() *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{}, nil
		},
	}
}

// ---- Tests ----

func TestBillingConfirmationService_GetOrCreate(t *testing.T) {
	existingReview := &model.BillingConfirmation{
		ID:              1,
		MedicalRecordID: 1,
		Status:          model.ConfirmationStatusPending,
	}

	tests := []struct {
		name              string
		medicalRecordID   uint64
		findByRecordIDErr error
		createErr         error
		wantErr           bool
		wantCreate        bool
	}{
		{
			name:              "returns existing review when found",
			medicalRecordID:   1,
			findByRecordIDErr: nil,
			createErr:         nil,
			wantErr:           false,
			wantCreate:        false,
		},
		{
			name:              "creates review when not found",
			medicalRecordID:   1,
			findByRecordIDErr: apperrors.WrapNotFound("billing_confirmation", "1"),
			createErr:         nil,
			wantErr:           false,
			wantCreate:        true,
		},
		{
			name:              "returns error on non-notfound error",
			medicalRecordID:   1,
			findByRecordIDErr: errors.New("db error"),
			createErr:         nil,
			wantErr:           true,
			wantCreate:        false,
		},
		{
			name:              "returns error when create fails",
			medicalRecordID:   1,
			findByRecordIDErr: apperrors.WrapNotFound("billing_confirmation", "1"),
			createErr:         errors.New("db create error"),
			wantErr:           true,
			wantCreate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingConfirmationRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.BillingConfirmation, error) {
					if tt.findByRecordIDErr != nil {
						return nil, tt.findByRecordIDErr
					}
					return existingReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingConfirmation) error {
					return tt.createErr
				},
			}
			svc := NewBillingConfirmationService(repo, okMedRecForBilling())

			review, err := svc.GetOrCreate(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
				assert.Equal(t, tt.medicalRecordID, review.MedicalRecordID)
			}
		})
	}
}

func TestBillingConfirmationService_Confirm(t *testing.T) {
	confirmedBy := uint64(3)
	memo := "Confirmed by physician"

	tests := []struct {
		name             string
		medicalRecordID  uint64
		input            *ConfirmBillingConfirmationInput
		repoReview       *model.BillingConfirmation
		repoFindErr      error
		repoUpdateErr    error
		repoReturnReview *model.BillingConfirmation
		wantErr          bool
	}{
		{
			name:            "confirms billing review successfully",
			medicalRecordID: 1,
			input: &ConfirmBillingConfirmationInput{
				ConfirmedBy: confirmedBy,
				Memo:        memo,
			},
			repoReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusConfirmed,
			},
			wantErr: false,
		},
		{
			name:            "returns error when already confirmed",
			medicalRecordID: 1,
			input: &ConfirmBillingConfirmationInput{
				ConfirmedBy: confirmedBy,
			},
			repoReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusConfirmed,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			wantErr:       true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			input: &ConfirmBillingConfirmationInput{
				ConfirmedBy: confirmedBy,
			},
			repoReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingConfirmationRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.BillingConfirmation, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingConfirmation) error {
					return nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewBillingConfirmationService(repo, okMedRecForBilling())

			review, err := svc.Confirm(context.Background(), 1, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
			}
		})
	}
}

func TestBillingConfirmationService_Return(t *testing.T) {
	returnedBy := uint64(2)
	returnReason := "Missing information"
	memo := "Please revise"

	tests := []struct {
		name             string
		medicalRecordID  uint64
		input            *ReturnBillingConfirmationInput
		repoReview       *model.BillingConfirmation
		repoFindErr      error
		repoUpdateErr    error
		repoReturnReview *model.BillingConfirmation
		wantErr          bool
	}{
		{
			name:            "returns billing review successfully",
			medicalRecordID: 1,
			input: &ReturnBillingConfirmationInput{
				ReturnedBy:   returnedBy,
				ReturnReason: returnReason,
				Memo:         memo,
			},
			repoReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusConfirmed,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusReturned,
			},
			wantErr: false,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			input: &ReturnBillingConfirmationInput{
				ReturnedBy:   returnedBy,
				ReturnReason: returnReason,
			},
			repoReview: &model.BillingConfirmation{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.ConfirmationStatusConfirmed,
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingConfirmationRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.BillingConfirmation, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingConfirmation) error {
					return nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewBillingConfirmationService(repo, okMedRecForBilling())

			review, err := svc.Return(context.Background(), 1, tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
			}
		})
	}
}

// TestBillingConfirmationService_GetOrCreate_CrossTenantParentRejected は
// clinic A の呼び出しが clinic B のカルテを参照する自動作成（GetOrCreate）を拒否し、
// billing_confirmation（自前 clinic_id を持たない）が clinic B のカルテ配下に
// 永続化されない（repo.Create が呼ばれない）ことを検証する。
// 親所有権検証（medRec.FindByID(clinicID, medRecID)）を削除すると必ず失敗する回帰テスト。
func TestBillingConfirmationService_GetOrCreate_CrossTenantParentRejected(t *testing.T) {
	const (
		clinicA     = uint64(1)
		clinicBMRID = uint64(99) // clinic B のカルテ ID（clinic A は所有しない）
	)
	createCalled := false
	repo := &mockBillingConfirmationRepository{
		// clinic A のスコープでは clinic B のカルテの確認は見えない → NotFound（作成分岐へ）。
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.BillingConfirmation, error) {
			return nil, apperrors.WrapNotFound("billing_confirmation", "99")
		},
		createFn: func(_ context.Context, _ *model.BillingConfirmation) error {
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
	svc := NewBillingConfirmationService(repo, medRec)

	review, err := svc.GetOrCreate(context.Background(), clinicA, clinicBMRID)

	assert.Error(t, err, "clinic A から clinic B のカルテへの billing_confirmation 自動作成は拒否されるべき")
	assert.True(t, apperrors.IsNotFound(err), "拒否は NotFound(404) にマップされるべき: %v", err)
	assert.Nil(t, review)
	assert.False(t, createCalled, "親カルテ所有権検証に失敗した場合、billing_confirmation を永続化してはならない")
}
