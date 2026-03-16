package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- BillingReview モック ----

type mockBillingReviewRepository struct {
	findByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error)
	createFn                func(ctx context.Context, review *model.BillingReview) error
	updateFn                func(ctx context.Context, reviewID uint64, fields map[string]any) error
}

func (m *mockBillingReviewRepository) FindByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.BillingReview, error) {
	return m.findByMedicalRecordIDFn(ctx, medicalRecordID)
}

func (m *mockBillingReviewRepository) Create(ctx context.Context, review *model.BillingReview) error {
	return m.createFn(ctx, review)
}

func (m *mockBillingReviewRepository) Update(ctx context.Context, reviewID uint64, fields map[string]any) error {
	return m.updateFn(ctx, reviewID, fields)
}

// ---- Tests ----

func TestBillingReviewService_GetOrCreate(t *testing.T) {
	existingReview := &model.BillingReview{
		ID:              1,
		MedicalRecordID: 1,
		Status:          model.BillingReviewStatusPending,
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
			findByRecordIDErr: apperrors.WrapNotFound("billing_review", "1"),
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
			findByRecordIDErr: apperrors.WrapNotFound("billing_review", "1"),
			createErr:         errors.New("db create error"),
			wantErr:           true,
			wantCreate:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingReviewRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _ uint64) (*model.BillingReview, error) {
					if tt.findByRecordIDErr != nil {
						return nil, tt.findByRecordIDErr
					}
					return existingReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingReview) error {
					return tt.createErr
				},
			}
			svc := NewBillingReviewService(repo)

			review, err := svc.GetOrCreate(context.Background(), tt.medicalRecordID)

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

func TestBillingReviewService_Confirm(t *testing.T) {
	confirmedBy := uint64(3)
	memo := "Confirmed by physician"

	tests := []struct {
		name              string
		medicalRecordID   uint64
		input             *ConfirmBillingReviewInput
		repoReview        *model.BillingReview
		repoFindErr       error
		repoUpdateErr     error
		repoReturnReview  *model.BillingReview
		wantErr           bool
	}{
		{
			name:            "confirms billing review successfully",
			medicalRecordID: 1,
			input: &ConfirmBillingReviewInput{
				ConfirmedBy: confirmedBy,
				Memo:        memo,
			},
			repoReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusConfirmed,
			},
			wantErr: false,
		},
		{
			name:            "returns error when already confirmed",
			medicalRecordID: 1,
			input: &ConfirmBillingReviewInput{
				ConfirmedBy: confirmedBy,
			},
			repoReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusConfirmed,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			wantErr:       true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			input: &ConfirmBillingReviewInput{
				ConfirmedBy: confirmedBy,
			},
			repoReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingReviewRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _ uint64) (*model.BillingReview, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingReview) error {
					return nil
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewBillingReviewService(repo)

			review, err := svc.Confirm(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
			}
		})
	}
}

func TestBillingReviewService_Return(t *testing.T) {
	returnedBy := uint64(2)
	returnReason := "Missing information"
	memo := "Please revise"

	tests := []struct {
		name              string
		medicalRecordID   uint64
		input             *ReturnBillingReviewInput
		repoReview        *model.BillingReview
		repoFindErr       error
		repoUpdateErr     error
		repoReturnReview  *model.BillingReview
		wantErr           bool
	}{
		{
			name:            "returns billing review successfully",
			medicalRecordID: 1,
			input: &ReturnBillingReviewInput{
				ReturnedBy:   returnedBy,
				ReturnReason: returnReason,
				Memo:         memo,
			},
			repoReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: nil,
			repoReturnReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusReturned,
			},
			wantErr: false,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			input: &ReturnBillingReviewInput{
				ReturnedBy:   returnedBy,
				ReturnReason: returnReason,
			},
			repoReview: &model.BillingReview{
				ID:              1,
				MedicalRecordID: 1,
				Status:          model.BillingReviewStatusPending,
			},
			repoFindErr:   nil,
			repoUpdateErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBillingReviewRepository{
				findByMedicalRecordIDFn: func(_ context.Context, _ uint64) (*model.BillingReview, error) {
					if tt.repoFindErr != nil {
						return nil, tt.repoFindErr
					}
					return tt.repoReview, nil
				},
				createFn: func(_ context.Context, _ *model.BillingReview) error {
					return nil
				},
				updateFn: func(_ context.Context, _ uint64, _ map[string]any) error {
					return tt.repoUpdateErr
				},
			}
			svc := NewBillingReviewService(repo)

			review, err := svc.Return(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, review)
			}
		})
	}
}
