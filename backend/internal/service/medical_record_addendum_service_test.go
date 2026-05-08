package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMedicalRecordAddendumRepository は MedicalRecordAddendumRepository のテスト用モック
type mockMedicalRecordAddendumRepository struct {
	createFn                func(ctx context.Context, addendum *model.MedicalRecordAddendum) error
	findByIDFn              func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error)
	findByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error)
}

func (m *mockMedicalRecordAddendumRepository) Create(ctx context.Context, addendum *model.MedicalRecordAddendum) error {
	if m.createFn != nil {
		return m.createFn(ctx, addendum)
	}
	return nil
}

func (m *mockMedicalRecordAddendumRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockMedicalRecordAddendumRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	if m.findByMedicalRecordIDFn != nil {
		return m.findByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
	}
	return nil, nil
}

func TestMedicalRecordAddendumService_Create(t *testing.T) {
	finalizedRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Status: model.MedicalRecordStatusFinalized}
	draftRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Status: model.MedicalRecordStatusDraft}

	tests := []struct {
		name            string
		input           CreateMedicalRecordAddendumInput
		medicalRecord   *model.MedicalRecord
		medicalRecordErr error
		repoErr         error
		wantErr         bool
		wantInvalid     bool
	}{
		{
			name: "creates addendum successfully",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 1,
				AuthorUserID:    10,
				AfterText:       "updated soap note text",
				Reason:          "correction of typo",
			},
			medicalRecord: finalizedRecord,
			wantErr:       false,
		},
		{
			name: "rejects empty after_text",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 1,
				AuthorUserID:    10,
				AfterText:       "",
				Reason:          "some reason",
			},
			medicalRecord: finalizedRecord,
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "rejects empty reason",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 1,
				AuthorUserID:    10,
				AfterText:       "some text",
				Reason:          "",
			},
			medicalRecord: finalizedRecord,
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "rejects addendum on draft medical record",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 1,
				AuthorUserID:    10,
				AfterText:       "updated text",
				Reason:          "some reason",
			},
			medicalRecord: draftRecord,
			wantErr:       true,
			wantInvalid:   true,
		},
		{
			name: "returns error when medical record not found",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 999,
				AuthorUserID:    10,
				AfterText:       "updated text",
				Reason:          "some reason",
			},
			medicalRecordErr: apperrors.WrapNotFound("medical_record", "999"),
			wantErr:          true,
		},
		{
			name: "returns error on repository create failure",
			input: CreateMedicalRecordAddendumInput{
				MedicalRecordID: 1,
				AuthorUserID:    10,
				AfterText:       "updated text",
				Reason:          "some reason",
			},
			medicalRecord: finalizedRecord,
			repoErr:       errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.medicalRecordErr != nil {
						return nil, tt.medicalRecordErr
					}
					return tt.medicalRecord, nil
				},
			}
			addendumRepo := &mockMedicalRecordAddendumRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecordAddendum) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordAddendumService(addendumRepo, mrRepo)

			addendum, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, addendum)
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, addendum)
				assert.Equal(t, tt.input.MedicalRecordID, addendum.MedicalRecordID)
				assert.Equal(t, tt.input.AuthorUserID, addendum.AuthorUserID)
				assert.Equal(t, tt.input.AfterText, addendum.AfterText)
				assert.Equal(t, tt.input.Reason, addendum.Reason)
			}
		})
	}
}

func TestMedicalRecordAddendumService_FindByMedicalRecordID(t *testing.T) {
	finalizedRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Status: model.MedicalRecordStatusFinalized}

	t.Run("returns addenda successfully", func(t *testing.T) {
		expected := []*model.MedicalRecordAddendum{
			{ID: 1, MedicalRecordID: 1, AfterText: "text 1"},
			{ID: 2, MedicalRecordID: 1, AfterText: "text 2"},
		}
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return finalizedRecord, nil
			},
		}
		addendumRepo := &mockMedicalRecordAddendumRepository{
			findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]*model.MedicalRecordAddendum, error) {
				return expected, nil
			},
		}
		svc := NewMedicalRecordAddendumService(addendumRepo, mrRepo)

		addenda, err := svc.FindByMedicalRecordID(context.Background(), 1, 1)

		assert.NoError(t, err)
		assert.Len(t, addenda, 2)
	})

	t.Run("returns error when medical record not found", func(t *testing.T) {
		mrRepo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, apperrors.WrapNotFound("medical_record", "999")
			},
		}
		svc := NewMedicalRecordAddendumService(&mockMedicalRecordAddendumRepository{}, mrRepo)

		addenda, err := svc.FindByMedicalRecordID(context.Background(), 1, 999)

		assert.Error(t, err)
		assert.Nil(t, addenda)
	})
}
