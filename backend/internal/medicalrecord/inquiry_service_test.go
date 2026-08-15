package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockInquiryRepository struct {
	upsertFn func(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error)
}

func (m *mockInquiryRepository) SaveByMedicalRecordID(ctx context.Context, clinicID uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	return m.upsertFn(ctx, clinicID, inquiry)
}

func (m *mockInquiryRepository) CountByChiefComplaintTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func TestInquiryService_Save(t *testing.T) {
	complaint := "食欲不振"
	notes := "2日前から"
	typeID := uint64(3)

	tests := []struct {
		name      string
		input     UpsertInquiryInput
		upsertErr error
		wantErr   bool
	}{
		{
			name: "upserts inquiry successfully",
			input: UpsertInquiryInput{
				ClinicID:             1,
				MedicalRecordID:      10,
				ChiefComplaintTypeID: &typeID,
				ChiefComplaint:       &complaint,
				Notes:                &notes,
			},
			wantErr: false,
		},
		{
			name: "upserts without optional fields",
			input: UpsertInquiryInput{
				ClinicID:        1,
				MedicalRecordID: 10,
			},
			wantErr: false,
		},
		{
			name: "propagates repository error",
			input: UpsertInquiryInput{
				ClinicID:        1,
				MedicalRecordID: 10,
				ChiefComplaint:  &complaint,
			},
			upsertErr: errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returned := &model.Inquiry{ID: 1, MedicalRecordID: tt.input.MedicalRecordID}
			repo := &mockInquiryRepository{
				upsertFn: func(_ context.Context, _ uint64, _ *model.Inquiry) (*model.Inquiry, error) {
					if tt.upsertErr != nil {
						return nil, tt.upsertErr
					}
					return returned, nil
				},
			}
			svc := NewInquiryService(repo, okChiefComplaintTypeRepo())
			result, err := svc.Save(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.MedicalRecordID, result.MedicalRecordID)
			}
		})
	}
}
