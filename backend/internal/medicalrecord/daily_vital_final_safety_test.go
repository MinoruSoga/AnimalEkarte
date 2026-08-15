package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type finalSafetyTxContextKey struct{}

type finalSafetyTransactor struct {
	callbackErr error
}

func (t *finalSafetyTransactor) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	txCtx := context.WithValue(ctx, finalSafetyTxContextKey{}, true)
	t.callbackErr = fn(txCtx)
	return t.callbackErr
}

func isFinalSafetyTx(ctx context.Context) bool {
	inTx, _ := ctx.Value(finalSafetyTxContextKey{}).(bool)
	return inTx
}

func TestDailyRecordServiceRejectsPollutedHospitalizationOwnerPetBeforeChildWrites(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
		ownerID           = uint64(20)
		petID             = uint64(30)
	)
	date := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		call func(DailyRecordService) (*model.DailyRecord, error)
	}{
		{
			name: "vital",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddVitalRecord(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateVitalRecordInput{Time: date.Add(9 * time.Hour)},
				)
			},
		},
		{
			name: "care log",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddCareLog(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateCareLogInput{
						Time: "09:00:00",
						Type: string(model.CareLogTypeFood),
					},
				)
			},
		},
		{
			name: "staff note",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddStaffNote(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateStaffNoteInput{Time: "09:00:00", Content: "handoff"},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findOrCreateCalls := 0
			childWriteCalls := 0
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(
					context.Context,
					uint64,
					uint64,
					time.Time,
				) (*model.DailyRecord, error) {
					findOrCreateCalls++
					return &model.DailyRecord{}, nil
				},
				createVitalRecordFn: func(context.Context, *model.VitalRecord) error {
					childWriteCalls++
					return nil
				},
				createCareLogFn: func(context.Context, *model.CareLog) error {
					childWriteCalls++
					return nil
				},
				createStaffNoteFn: func(context.Context, *model.StaffNote) error {
					childWriteCalls++
					return nil
				},
			}
			hospitalizations := &mockHospitalizationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID:       hospitalizationID,
						ClinicID: clinicID,
						OwnerID:  ownerID,
						PetID:    petID,
					}, nil
				},
			}
			service := NewDailyRecordServiceWithRelationValidation(
				repo,
				hospitalizations,
				&dailyRecordOwnerPetVerifierStub{ownerID: ownerID + 1, err: errors.New("foreign pet relation")},
				nil,
				nil,
				&finalSafetyTransactor{},
			)

			got, err := tt.call(service)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, findOrCreateCalls)
			assert.Zero(t, childWriteCalls)
		})
	}
}

func TestDailyRecordServiceReadsFailClosedForPollutedNestedStaff(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
		ownerID           = uint64(20)
		petID             = uint64(30)
		staffID           = uint64(40)
		dailyID           = uint64(50)
	)
	date := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		record model.DailyRecord
	}{
		{
			name: "vital",
			record: model.DailyRecord{
				ID: dailyID, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date,
				VitalRecords: []model.VitalRecord{{
					ID: 1, ClinicID: clinicID, PetID: petID, DailyRecordID: ptrUint64(dailyID),
					StaffID: ptrUint64(staffID),
					Staff:   &model.Staff{ID: staffID, Name: "foreign secret", IsActive: true},
				}},
			},
		},
		{
			name: "care log",
			record: model.DailyRecord{
				ID: dailyID, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date,
				CareLogs: []model.CareLog{{
					ID: 1, ClinicID: clinicID, DailyRecordID: dailyID,
					StaffID: ptrUint64(staffID),
					Staff:   &model.Staff{ID: staffID, Name: "foreign secret", IsActive: true},
				}},
			},
		},
		{
			name: "staff note",
			record: model.DailyRecord{
				ID: dailyID, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date,
				StaffNotes: []model.StaffNote{{
					ID: 1, DailyRecordID: dailyID,
					StaffID: ptrUint64(staffID),
					Staff:   &model.Staff{ID: staffID, Name: "foreign secret", IsActive: true},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				listByHospitalizationIDFn: func(
					context.Context,
					uint64,
					uint64,
				) ([]model.DailyRecord, error) {
					return []model.DailyRecord{tt.record}, nil
				},
				findByHospitalizationIDAndDateFn: func(
					context.Context,
					uint64,
					uint64,
					time.Time,
				) (*model.DailyRecord, error) {
					record := tt.record
					return &record, nil
				},
			}
			hospitalizations := &mockHospitalizationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID:       hospitalizationID,
						ClinicID: clinicID,
						OwnerID:  ownerID,
						PetID:    petID,
					}, nil
				},
			}
			service := NewDailyRecordServiceWithRelationValidation(
				repo,
				hospitalizations,
				&dailyRecordOwnerPetVerifierStub{ownerID: ownerID},
				&clinicalStaffLockerStub{
					staff: &model.Staff{ID: staffID, IsActive: true},
				},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{
						StaffID:  staffID,
						ClinicID: clinicID + 1,
					},
				},
				&finalSafetyTransactor{},
			)

			records, err := service.List(context.Background(), clinicID, hospitalizationID)
			require.Error(t, err)
			assert.Nil(t, records, "polluted nested staff data must not be returned")

			record, err := service.GetByDate(
				context.Background(),
				clinicID,
				hospitalizationID,
				date,
			)
			require.Error(t, err)
			assert.Nil(t, record, "polluted nested staff data must not be returned")
		})
	}
}

func TestDailyRecordServiceChildRefetchFailureIsInsideTransaction(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
		ownerID           = uint64(20)
		petID             = uint64(30)
		dailyID           = uint64(40)
	)
	date := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	refetchErr := errors.New("forced daily response refetch failure")
	outsideTxErr := errors.New("daily response refetch ran after commit")

	tests := []struct {
		name string
		call func(DailyRecordService) (*model.DailyRecord, error)
	}{
		{
			name: "vital",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddVitalRecord(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateVitalRecordInput{Time: date.Add(9 * time.Hour)},
				)
			},
		},
		{
			name: "care log",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddCareLog(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateCareLogInput{
						Time: "09:00:00",
						Type: string(model.CareLogTypeFood),
					},
				)
			},
		},
		{
			name: "staff note",
			call: func(service DailyRecordService) (*model.DailyRecord, error) {
				return service.AddStaffNote(
					context.Background(),
					clinicID,
					hospitalizationID,
					date,
					&CreateStaffNoteInput{Time: "09:00:00", Content: "handoff"},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &finalSafetyTransactor{}
			writeCalls := 0
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(
					context.Context,
					uint64,
					uint64,
					time.Time,
				) (*model.DailyRecord, error) {
					return &model.DailyRecord{
						ID: dailyID, ClinicID: clinicID,
						HospitalizationID: hospitalizationID, Date: date,
					}, nil
				},
				createVitalRecordFn: func(_ context.Context, vital *model.VitalRecord) error {
					writeCalls++
					vital.ID = 1
					return nil
				},
				createCareLogFn: func(_ context.Context, careLog *model.CareLog) error {
					writeCalls++
					careLog.ID = 1
					return nil
				},
				createStaffNoteFn: func(_ context.Context, note *model.StaffNote) error {
					writeCalls++
					note.ID = 1
					return nil
				},
				findByHospitalizationIDAndDateFn: func(
					ctx context.Context,
					_ uint64,
					_ uint64,
					_ time.Time,
				) (*model.DailyRecord, error) {
					if !isFinalSafetyTx(ctx) {
						return nil, outsideTxErr
					}
					return nil, refetchErr
				},
			}
			hospitalizations := &mockHospitalizationRepository{
				findByIDFn: func(context.Context, uint64, uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID: hospitalizationID, ClinicID: clinicID,
						OwnerID: ownerID, PetID: petID,
					}, nil
				},
			}
			service := NewDailyRecordServiceWithRelationValidation(
				repo,
				hospitalizations,
				&dailyRecordOwnerPetVerifierStub{ownerID: ownerID},
				nil,
				nil,
				tx,
			)

			got, err := tt.call(service)

			require.ErrorIs(t, err, refetchErr)
			assert.Nil(t, got)
			require.ErrorIs(t, tx.callbackErr, refetchErr)
			assert.Equal(t, 1, writeCalls)
		})
	}
}

func TestVitalServiceUpdateRefetchFailureIsInsideTransaction(t *testing.T) {
	const (
		clinicID        = uint64(1)
		medicalRecordID = uint64(10)
		vitalID         = uint64(20)
		ownerID         = uint64(30)
		petID           = uint64(40)
	)
	refetchErr := errors.New("forced vital response refetch failure")
	outsideTxErr := errors.New("vital response refetch ran after commit")
	findCalls := 0
	updateCalls := 0
	repo := &mockVitalRepository{
		findByIDFn: func(ctx context.Context, _ uint64, _ uint64) (*model.VitalRecord, error) {
			findCalls++
			if findCalls == 1 {
				return &model.VitalRecord{
					ID: vitalID, ClinicID: clinicID, PetID: petID,
					MedicalRecordID: ptrUint64(medicalRecordID),
				}, nil
			}
			if !isFinalSafetyTx(ctx) {
				return nil, outsideTxErr
			}
			return nil, refetchErr
		},
		updateFn: func(context.Context, uint64, uint64, map[string]any) error {
			updateCalls++
			return nil
		},
	}
	medicalRecords := &mockMedicalRecordRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: medicalRecordID, ClinicID: clinicID,
				OwnerID: ptrUint64(ownerID), PetID: ptrUint64(petID),
				Status: model.MedicalRecordStatusDraft,
			}, nil
		},
	}
	tx := &finalSafetyTransactor{}
	service := NewVitalServiceWithRelationValidation(
		repo,
		medicalRecords,
		okVitalAudit(),
		validVitalRelations(petID, ownerID),
		nil,
		nil,
		tx,
	)
	notes := "updated"

	got, err := service.Update(
		context.Background(),
		clinicID,
		medicalRecordID,
		vitalID,
		&UpdateVitalInput{Notes: &notes},
	)

	require.ErrorIs(t, err, refetchErr)
	assert.Nil(t, got)
	require.ErrorIs(t, tx.callbackErr, refetchErr)
	assert.Equal(t, 1, updateCalls)
	assert.Equal(t, 2, findCalls)
}
