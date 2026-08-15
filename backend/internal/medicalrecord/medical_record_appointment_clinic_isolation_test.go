package medicalrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordService_Create_RejectsForeignOwnerPetBeforeAppointmentBackfill(t *testing.T) {
	tests := []struct {
		name          string
		assertOwnerFn func(context.Context, uint64, uint64) error
		findPetFn     func(context.Context, uint64, uint64) (uint64, error)
	}{
		{
			name: "foreign owner",
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return apperrors.WrapNotFound("owner", "201")
			},
		},
		{
			name: "foreign pet",
			findPetFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name: "pet belongs to another owner",
			findPetFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 999, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appointmentID := uint64(77)
			ownerID := uint64(201)
			petID := uint64(202)
			recordRepo := &mockMedicalRecordRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					t.Fatal("medical record must not be created with invalid appointment links")
					return nil
				},
			}
			reservationRepo := &mockReservationRepoForMedicalRecord{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
					return &model.Reservation{ID: id, ClinicID: clinicID}, nil
				},
				assertOwnerFn:  tt.assertOwnerFn,
				findPetOwnerFn: tt.findPetFn,
				backfillFn: func(_ context.Context, _ uint64, _ uint64, _, _, _ *uint64) error {
					t.Fatal("appointment must not be backfilled with invalid Owner/Pet links")
					return nil
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, &mockTransactor{})

			got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
				Date:          time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
				OwnerID:       &ownerID,
				PetID:         &petID,
				AppointmentID: &appointmentID,
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
		})
	}
}
