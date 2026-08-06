package medicalrecord

// medical_record_owner_pet_clinic_isolation_test.go — AUD-008
// Appointment なし Create/Update の Owner/Pet clinic 所有確認・Owner-Pet 整合・tx 参加を検証する。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type medicalRecordTxContextKey struct{}

type medicalRecordTxMarker struct{}

func (medicalRecordTxMarker) WithTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, medicalRecordTxContextKey{}, true))
}

func assertMedicalRecordTxContext(ctx context.Context, t *testing.T) {
	t.Helper()
	assert.Equal(t, true, ctx.Value(medicalRecordTxContextKey{}), "Owner/Pet checks and medical record writes must share the transaction context")
}

func newMedicalRecordClinicIsolationService(
	recordRepo *mockMedicalRecordRepository,
	reservationRepo *mockReservationRepoForMedicalRecord,
) MedicalRecordService {
	return NewMedicalRecordServiceWithTxAudit(
		recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})
}

func TestMedicalRecordService_Create_RejectsCrossClinicOwnerPetWithoutAppointment(t *testing.T) {
	tests := []struct {
		name          string
		assertOwnerFn func(context.Context, uint64, uint64) error
		findPetFn     func(context.Context, uint64, uint64) (uint64, error)
	}{
		{
			name: "rejects_cross_clinic_owner",
			assertOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(201), ownerID)
				return apperrors.WrapNotFound("owner", "201")
			},
		},
		{
			name: "rejects_cross_clinic_pet",
			findPetFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
				assertMedicalRecordTxContext(ctx, t)
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(202), petID)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name: "rejects_pet_belonging_to_different_owner",
			findPetFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
				assertMedicalRecordTxContext(ctx, t)
				return 999, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ownerID := uint64(201)
			petID := uint64(202)
			recordRepo := &mockMedicalRecordRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					t.Fatal("medical record must not be created with invalid Owner/Pet links")
					return nil
				},
			}
			reservationRepo := &mockReservationRepoForMedicalRecord{
				assertOwnerFn:  tt.assertOwnerFn,
				findPetOwnerFn: tt.findPetFn,
			}
			svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

			got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
				Date:    time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
				OwnerID: &ownerID,
				PetID:   &petID,
			})

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
		})
	}
}

func TestMedicalRecordService_Create_RejectsCrossClinicDoctorWithoutAppointment(t *testing.T) {
	doctorID := uint64(303)
	recordRepo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, _ *model.MedicalRecord) error {
			t.Fatal("medical record must not be created with a doctor outside the clinic")
			return nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		assertDoctorFn: func(ctx context.Context, clinicID, gotDoctorID uint64) error {
			assertMedicalRecordTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, doctorID, gotDoctorID)
			return apperrors.WrapNotFound("staff", "303")
		},
	}
	svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

	record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
		Date:     time.Date(2026, 7, 22, 10, 0, 0, 0, time.UTC),
		DoctorID: &doctorID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, record)
}

func TestMedicalRecordService_Create_AcceptsSameClinicOwnerPetAndNilContract(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)

	tests := []struct {
		name    string
		ownerID *uint64
		petID   *uint64
	}{
		{name: "accepts_same_clinic_matching_owner_and_pet", ownerID: &ownerID, petID: &petID},
		{name: "accepts_nil_owner_and_nil_pet", ownerID: nil, petID: nil},
		{name: "accepts_owner_only", ownerID: &ownerID, petID: nil},
		{name: "accepts_pet_only", ownerID: nil, petID: &petID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			recordRepo := &mockMedicalRecordRepository{
				createFn: func(ctx context.Context, record *model.MedicalRecord) error {
					assertMedicalRecordTxContext(ctx, t)
					created = true
					record.ID = 55
					return nil
				},
			}
			reservationRepo := &mockReservationRepoForMedicalRecord{
				assertOwnerFn: func(ctx context.Context, clinicID, id uint64) error {
					assertMedicalRecordTxContext(ctx, t)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, ownerID, id)
					return nil
				},
				findPetOwnerFn: func(ctx context.Context, clinicID, id uint64) (uint64, error) {
					assertMedicalRecordTxContext(ctx, t)
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, petID, id)
					return ownerID, nil
				},
			}
			svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

			got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
				Date:    time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
				OwnerID: tt.ownerID,
				PetID:   tt.petID,
			})

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.True(t, created)
		})
	}
}

func TestMedicalRecordService_Update_RejectsCrossClinicOwnerPetAndMismatch(t *testing.T) {
	existingOwner := uint64(10)
	existingPet := uint64(20)
	existing := &model.MedicalRecord{
		ID: 1, ClinicID: 1, Version: 0,
		OwnerID: &existingOwner, PetID: &existingPet,
		Status: model.MedicalRecordStatusDraft,
	}

	tests := []struct {
		name          string
		input         UpdateMedicalRecordInput
		assertOwnerFn func(context.Context, uint64, uint64) error
		findPetFn     func(context.Context, uint64, uint64) (uint64, error)
	}{
		{
			name:  "rejects_cross_clinic_owner",
			input: UpdateMedicalRecordInput{OwnerID: uint64PtrMR(201)},
			assertOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				return apperrors.WrapNotFound("owner", "201")
			},
		},
		{
			name:  "rejects_cross_clinic_pet",
			input: UpdateMedicalRecordInput{PetID: uint64PtrMR(202)},
			findPetFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
				assertMedicalRecordTxContext(ctx, t)
				return 0, apperrors.WrapNotFound("pet", "202")
			},
		},
		{
			name:  "rejects_pet_only_change_that_breaks_final_owner_pet_consistency",
			input: UpdateMedicalRecordInput{PetID: uint64PtrMR(202)},
			findPetFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
				assertMedicalRecordTxContext(ctx, t)
				return 999, nil
			},
		},
		{
			name:  "rejects_owner_only_change_that_breaks_final_owner_pet_consistency",
			input: UpdateMedicalRecordInput{OwnerID: uint64PtrMR(201)},
			assertOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				return nil
			},
			findPetFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
				assertMedicalRecordTxContext(ctx, t)
				assert.Equal(t, existingPet, petID)
				return existingOwner, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated := false
			recordRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					cp := *existing
					return &cp, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
					updated = true
					t.Fatal("medical record must not be updated with invalid Owner/Pet links")
					return nil, nil
				},
			}
			reservationRepo := &mockReservationRepoForMedicalRecord{
				assertOwnerFn:  tt.assertOwnerFn,
				findPetOwnerFn: tt.findPetFn,
			}
			svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

			got, err := svc.Update(context.Background(), 1, 1, tt.input)

			assert.Error(t, err)
			assert.True(t, apperrors.IsNotFound(err))
			assert.Nil(t, got)
			assert.False(t, updated)
		})
	}
}

func TestMedicalRecordService_Update_AcceptsSameClinicFinalOwnerPet(t *testing.T) {
	existingOwner := uint64(10)
	existingPet := uint64(20)
	newOwner := uint64(11)
	newPet := uint64(21)
	existing := &model.MedicalRecord{
		ID: 1, ClinicID: 1, Version: 0,
		OwnerID: &existingOwner, PetID: &existingPet,
		Status: model.MedicalRecordStatusDraft,
	}

	updated := false
	recordRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			cp := *existing
			return &cp, nil
		},
		updateFieldsFn: func(ctx context.Context, _, _ uint64, fields map[string]any) (*model.MedicalRecord, error) {
			assertMedicalRecordTxContext(ctx, t)
			updated = true
			assert.Equal(t, newOwner, fields["owner_id"])
			assert.Equal(t, newPet, fields["pet_id"])
			return &model.MedicalRecord{ID: 1, ClinicID: 1, OwnerID: &newOwner, PetID: &newPet}, nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		assertOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) error {
			assertMedicalRecordTxContext(ctx, t)
			assert.Equal(t, newOwner, ownerID)
			return nil
		},
		findPetOwnerFn: func(ctx context.Context, clinicID, petID uint64) (uint64, error) {
			assertMedicalRecordTxContext(ctx, t)
			assert.Equal(t, newPet, petID)
			return newOwner, nil
		},
	}
	svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

	got, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
		OwnerID: &newOwner,
		PetID:   &newPet,
	})

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.True(t, updated)
}

func TestMedicalRecordService_Create_WithAppointment_RejectsForeignAndLeavesAppointmentUnchanged(t *testing.T) {
	appointmentID := uint64(77)
	ownerID := uint64(201)
	petID := uint64(202)
	appointmentUpdated := false
	recordRepo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, _ *model.MedicalRecord) error {
			t.Fatal("medical record must not be created")
			return nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
			return &model.Reservation{ID: id, ClinicID: clinicID}, nil
		},
		assertOwnerFn: func(ctx context.Context, _, _ uint64) error {
			assertMedicalRecordTxContext(ctx, t)
			return apperrors.WrapNotFound("owner", "201")
		},
		backfillFn: func(_ context.Context, _ uint64, _ uint64, _, _, _ *uint64) error {
			appointmentUpdated = true
			t.Fatal("appointment must not be backfilled with invalid Owner/Pet")
			return nil
		},
	}
	svc := newMedicalRecordClinicIsolationService(recordRepo, reservationRepo)

	got, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
		Date:          time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC),
		OwnerID:       &ownerID,
		PetID:         &petID,
		AppointmentID: &appointmentID,
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
	assert.False(t, appointmentUpdated)
}

func uint64PtrMR(v uint64) *uint64 { return &v }
