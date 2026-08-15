package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	staffdomain "github.com/animal-ekarte/backend/internal/staff"
)

type dailyRefetchFailureRepository struct {
	DailyRecordRepository
	refetchErr error
	wrote      bool
}

func (r *dailyRefetchFailureRepository) CreateVitalRecord(
	ctx context.Context,
	vital *model.VitalRecord,
) error {
	if err := r.DailyRecordRepository.CreateVitalRecord(ctx, vital); err != nil {
		return err
	}
	r.wrote = true
	return nil
}

func (r *dailyRefetchFailureRepository) CreateCareLog(
	ctx context.Context,
	careLog *model.CareLog,
) error {
	if err := r.DailyRecordRepository.CreateCareLog(ctx, careLog); err != nil {
		return err
	}
	r.wrote = true
	return nil
}

func (r *dailyRefetchFailureRepository) CreateStaffNote(
	ctx context.Context,
	note *model.StaffNote,
) error {
	if err := r.DailyRecordRepository.CreateStaffNote(ctx, note); err != nil {
		return err
	}
	r.wrote = true
	return nil
}

func (r *dailyRefetchFailureRepository) FindByHospitalizationIDAndDate(
	ctx context.Context,
	clinicID, hospitalizationID uint64,
	date time.Time,
) (*model.DailyRecord, error) {
	if r.wrote && persistence.TxFromContext(ctx) != nil {
		return nil, r.refetchErr
	}
	return r.DailyRecordRepository.FindByHospitalizationIDAndDate(
		ctx,
		clinicID,
		hospitalizationID,
		date,
	)
}

func TestDB_DailyRecordServiceRefetchFailureRollsBackChildWrites(t *testing.T) {
	refetchErr := errors.New("forced daily response refetch failure")
	date := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		childModel any
		call       func(DailyRecordService, *dailyRecordStaffRelationDBFixture) (*model.DailyRecord, error)
	}{
		{
			name:       "vital",
			childModel: &model.VitalRecord{},
			call: func(
				service DailyRecordService,
				fixture *dailyRecordStaffRelationDBFixture,
			) (*model.DailyRecord, error) {
				temperature := 38.1
				return service.AddVitalRecord(
					context.Background(),
					fixture.clinicA,
					fixture.hospitalization.ID,
					date,
					&CreateVitalRecordInput{
						Time:        date.Add(9 * time.Hour),
						Temperature: &temperature,
					},
				)
			},
		},
		{
			name:       "care log",
			childModel: &model.CareLog{},
			call: func(
				service DailyRecordService,
				fixture *dailyRecordStaffRelationDBFixture,
			) (*model.DailyRecord, error) {
				return service.AddCareLog(
					context.Background(),
					fixture.clinicA,
					fixture.hospitalization.ID,
					date,
					&CreateCareLogInput{
						Time:   "09:00:00",
						Type:   string(model.CareLogTypeFood),
						Status: string(model.CareLogStatusCompleted),
					},
				)
			},
		},
		{
			name:       "staff note",
			childModel: &model.StaffNote{},
			call: func(
				service DailyRecordService,
				fixture *dailyRecordStaffRelationDBFixture,
			) (*model.DailyRecord, error) {
				return service.AddStaffNote(
					context.Background(),
					fixture.clinicA,
					fixture.hospitalization.ID,
					date,
					&CreateStaffNoteInput{Time: "09:00:00", Content: "handoff"},
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupDailyRecordStaffRelationDBFixture(t)
			repo := &dailyRefetchFailureRepository{
				DailyRecordRepository: NewDailyRecordRepository(fixture.db),
				refetchErr:            refetchErr,
			}
			service := NewDailyRecordServiceWithRelationValidation(
				repo,
				NewHospitalizationRepository(fixture.db),
				reservation.NewReservationRepository(fixture.db),
				staffdomain.NewStaffRepository(fixture.db),
				staffdomain.NewStaffClinicAssignmentRepository(fixture.db),
				persistence.NewTransactor(fixture.db),
			)

			got, err := tt.call(service, fixture)

			require.ErrorIs(t, err, refetchErr)
			assert.Nil(t, got)
			assert.Zero(t, countDailyRelationRows(t, fixture.db, tt.childModel))
			assert.Zero(
				t,
				countDailyRelationRows(
					t,
					fixture.db,
					&model.DailyRecord{},
					"hospitalization_id = ? AND date = ?",
					fixture.hospitalization.ID,
					date,
				),
				"response refetch failure must roll back the parent and child writes",
			)
		})
	}
}

type vitalUpdateRefetchFailureRepository struct {
	VitalRepository
	refetchErr error
	updated    bool
}

func (r *vitalUpdateRefetchFailureRepository) Update(
	ctx context.Context,
	clinicID, vitalID uint64,
	fields map[string]any,
) error {
	if err := r.VitalRepository.Update(ctx, clinicID, vitalID, fields); err != nil {
		return err
	}
	r.updated = true
	return nil
}

func (r *vitalUpdateRefetchFailureRepository) FindByID(
	ctx context.Context,
	clinicID, vitalID uint64,
) (*model.VitalRecord, error) {
	if r.updated && persistence.TxFromContext(ctx) != nil {
		return nil, r.refetchErr
	}
	return r.VitalRepository.FindByID(ctx, clinicID, vitalID)
}

func TestDB_VitalServiceUpdateRefetchFailureRollsBackMutation(t *testing.T) {
	fixture := setupImageVitalRelationDBFixture(t)
	medicalRecordID := fixture.recordA.ID
	vital := &model.VitalRecord{
		ClinicID:        fixture.clinicA,
		PetID:           fixture.petA.ID,
		MedicalRecordID: &medicalRecordID,
		RecordedAt:      time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC),
		Notes:           "before",
	}
	require.NoError(t, fixture.db.Create(vital).Error)

	refetchErr := errors.New("forced vital response refetch failure")
	baseRepo := NewVitalRepository(fixture.db)
	repo := &vitalUpdateRefetchFailureRepository{
		VitalRepository: baseRepo,
		refetchErr:      refetchErr,
	}
	service := NewVitalServiceWithRelationValidation(
		repo,
		NewMedicalRecordRepository(fixture.db),
		okVitalAudit(),
		reservation.NewReservationRepository(fixture.db),
		staffdomain.NewStaffRepository(fixture.db),
		staffdomain.NewStaffClinicAssignmentRepository(fixture.db),
		persistence.NewTransactor(fixture.db),
	)
	updatedNotes := "after"

	got, err := service.Update(
		context.Background(),
		fixture.clinicA,
		fixture.recordA.ID,
		vital.ID,
		&UpdateVitalInput{Notes: &updatedNotes},
	)

	require.ErrorIs(t, err, refetchErr)
	assert.Nil(t, got)
	persisted, findErr := baseRepo.FindByID(
		context.Background(),
		fixture.clinicA,
		vital.ID,
	)
	require.NoError(t, findErr)
	assert.Equal(t, "before", persisted.Notes)
}

func TestDB_DailyRecordServiceReadsRejectPollutedNestedStaff(t *testing.T) {
	date := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		create func(*testing.T, *dailyRecordStaffRelationDBFixture, *model.DailyRecord)
	}{
		{
			name: "vital",
			create: func(
				t *testing.T,
				fixture *dailyRecordStaffRelationDBFixture,
				daily *model.DailyRecord,
			) {
				t.Helper()
				require.NoError(t, fixture.db.Create(&model.VitalRecord{
					ClinicID:      fixture.clinicA,
					PetID:         fixture.hospitalization.PetID,
					DailyRecordID: &daily.ID,
					RecordedAt:    date.Add(9 * time.Hour),
					StaffID:       &fixture.foreignStaff.ID,
				}).Error)
			},
		},
		{
			name: "care log",
			create: func(
				t *testing.T,
				fixture *dailyRecordStaffRelationDBFixture,
				daily *model.DailyRecord,
			) {
				t.Helper()
				require.NoError(t, fixture.db.Create(&model.CareLog{
					ClinicID:      fixture.clinicA,
					DailyRecordID: daily.ID,
					Time:          "09:00:00",
					Type:          model.CareLogTypeFood,
					Status:        model.CareLogStatusCompleted,
					StaffID:       &fixture.foreignStaff.ID,
				}).Error)
			},
		},
		{
			name: "staff note",
			create: func(
				t *testing.T,
				fixture *dailyRecordStaffRelationDBFixture,
				daily *model.DailyRecord,
			) {
				t.Helper()
				require.NoError(t, fixture.db.Create(&model.StaffNote{
					DailyRecordID: daily.ID,
					Time:          "09:00:00",
					Content:       "must not leak",
					StaffID:       &fixture.foreignStaff.ID,
				}).Error)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupDailyRecordStaffRelationDBFixture(t)
			daily := &model.DailyRecord{
				ClinicID:          fixture.clinicA,
				HospitalizationID: fixture.hospitalization.ID,
				Date:              date,
			}
			require.NoError(t, fixture.db.Create(daily).Error)
			tt.create(t, fixture, daily)

			records, err := fixture.service.List(
				context.Background(),
				fixture.clinicA,
				fixture.hospitalization.ID,
			)
			require.Error(t, err)
			assert.Nil(t, records)

			record, err := fixture.service.GetByDate(
				context.Background(),
				fixture.clinicA,
				fixture.hospitalization.ID,
				date,
			)
			require.Error(t, err)
			assert.Nil(t, record)
		})
	}
}
