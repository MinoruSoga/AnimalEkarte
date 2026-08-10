package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestMedicalRecordService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		clinicID    uint64
		petID       *uint64
		ownerID     *uint64
		page        int
		limit       int
		repoRecords []model.MedicalRecord
		repoTotal   int64
		repoErr     error
		wantLen     int
		wantTotal   int64
		wantErr     bool
	}{
		{
			name:     "returns all records for clinic",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoRecords: []model.MedicalRecord{
				{ID: 1, ClinicID: 1, RecordNo: "R001", Date: now},
				{ID: 2, ClinicID: 1, RecordNo: "R002", Date: now},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    ptrUint64(10),
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoRecords: []model.MedicalRecord{
				{ID: 1, ClinicID: 1, RecordNo: "R001", Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			petID:    nil,
			ownerID:  ptrUint64(5),
			page:     1,
			limit:    20,
			repoRecords: []model.MedicalRecord{
				{ID: 2, ClinicID: 1, RecordNo: "R002", Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:        "returns empty list when no records exist",
			clinicID:    1,
			petID:       nil,
			ownerID:     nil,
			page:        1,
			limit:       20,
			repoRecords: []model.MedicalRecord{},
			repoTotal:   0,
			repoErr:     nil,
			wantLen:     0,
			wantTotal:   0,
			wantErr:     false,
		},
		{
			name:        "propagates repository error",
			clinicID:    1,
			petID:       nil,
			ownerID:     nil,
			page:        1,
			limit:       20,
			repoRecords: nil,
			repoTotal:   0,
			repoErr:     errors.New("db connection error"),
			wantLen:     0,
			wantTotal:   0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPetID := (*uint64)(nil)
			capturedOwnerID := (*uint64)(nil)
			repo := &mockMedicalRecordRepository{
				findAllFn: func(_ context.Context, _ []uint64, filters MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
					capturedPetID = filters.PetID
					capturedOwnerID = filters.OwnerID
					return tt.repoRecords, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			records, total, err := svc.List(context.Background(), []uint64{tt.clinicID}, MedicalRecordListFilters{PetID: tt.petID, OwnerID: tt.ownerID}, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, records, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.petID, capturedPetID)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestMedicalRecordService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		clinicID   uint64
		id         uint64
		repoRecord *model.MedicalRecord
		repoErr    error
		wantRecord *model.MedicalRecord
		wantErr    error
	}{
		{
			name:       "returns record when found",
			clinicID:   1,
			id:         10,
			repoRecord: &model.MedicalRecord{ID: 10, ClinicID: 1, RecordNo: "R010", Date: now},
			repoErr:    nil,
			wantRecord: &model.MedicalRecord{ID: 10, ClinicID: 1, RecordNo: "R010", Date: now},
			wantErr:    nil,
		},
		{
			name:       "returns not found error when record does not exist",
			clinicID:   1,
			id:         999,
			repoRecord: nil,
			repoErr:    apperrors.WrapNotFound("medical_record", "999"),
			wantRecord: nil,
			wantErr:    apperrors.ErrNotFound,
		},
		{
			name:       "returns error on repository failure",
			clinicID:   1,
			id:         10,
			repoRecord: nil,
			repoErr:    errors.New("db error"),
			wantRecord: nil,
			wantErr:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return tt.repoRecord, tt.repoErr
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			record, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRecord, record)
			}
		})
	}
}

func TestMedicalRecordService_GetByID_NotFound(t *testing.T) {
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, apperrors.WrapNotFound("medical_record", "999")
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

	record, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, record)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func createMedicalRecordForTest(ctx context.Context, svc MedicalRecordService, record *model.MedicalRecord) (*model.MedicalRecord, error) {
	var status *model.MedicalRecordStatus
	if record.Status != "" {
		v := record.Status
		status = &v
	}
	visitType := model.VisitTypeRevisit
	if record.VisitType != nil {
		visitType = *record.VisitType
	}
	return svc.Create(ctx, record.ClinicID, &CreateMedicalRecordInput{
		RecordNo:                 record.RecordNo,
		Date:                     record.Date,
		OwnerID:                  record.OwnerID,
		PetID:                    record.PetID,
		DoctorID:                 record.DoctorID,
		AppointmentID:            record.AppointmentID,
		Status:                   status,
		VisitType:                visitType,
		NextVisitRecommendedDate: record.NextVisitRecommendedDate,
		RecommendationReason:     record.RecommendationReason,
		EnteredBy:                record.EnteredBy,
	})
}

func TestMedicalRecordService_Create(t *testing.T) {
	now := time.Now()
	statusDraft := model.MedicalRecordStatusDraft
	tests := []struct {
		name    string
		input   CreateMedicalRecordInput
		repoErr error
		wantErr bool
	}{
		{
			name: "creates record successfully",
			input: CreateMedicalRecordInput{
				RecordNo: "R100",
				Date:     now,
				Status:   &statusDraft,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns already exists error on duplicate record_no",
			input: CreateMedicalRecordInput{
				RecordNo: "R001",
				Date:     now,
			},
			repoErr: apperrors.WrapAlreadyExists("medical_record", "R001"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			input: CreateMedicalRecordInput{
				RecordNo: "R200",
				Date:     now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				createFn: func(_ context.Context, _ *model.MedicalRecord) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			record, err := svc.Create(context.Background(), 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, record)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}

	// FEAT-382-2 supplement: recommendation_reason whitelist validation on Create
	t.Run("valid recommendation_reason is accepted on Create", func(t *testing.T) {
		reason := "checkup"
		var capturedRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, r *model.MedicalRecord) error {
				capturedRecord = r
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:                 time.Now(),
			RecommendationReason: &reason,
		})
		assert.NoError(t, err)
		assert.NotNil(t, record)
		if assert.NotNil(t, capturedRecord) && assert.NotNil(t, capturedRecord.RecommendationReason) {
			assert.Equal(t, "checkup", *capturedRecord.RecommendationReason)
		}
	})

	t.Run("invalid recommendation_reason returns InvalidInput", func(t *testing.T) {
		reason := "invalid_value"
		createCalled := false
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:                 time.Now(),
			RecommendationReason: &reason,
		})
		assert.Error(t, err)
		assert.Nil(t, record)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), "revisit, checkup, prevention, exam")
		assert.False(t, createCalled, "repo.Create must not be called on invalid reason")
	})

	t.Run("nil recommendation_reason bypasses whitelist (existing behavior)", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date: time.Now(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, record)
	})

	t.Run("returns existing record when appointment already has one", func(t *testing.T) {
		appointmentID := uint64(77)
		petID := uint64(10)
		createCalled := false
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, clinicID, gotAppointmentID uint64) (*model.MedicalRecord, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, gotAppointmentID)
				return &model.MedicalRecord{
					ID:            123,
					ClinicID:      1,
					Date:          time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
					PetID:         &petID,
					AppointmentID: &appointmentID,
				}, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		appointment := &model.Reservation{
			ID:        appointmentID,
			ClinicID:  1,
			PetID:     &petID,
			StartTime: time.Date(2026, 5, 29, 9, 0, 0, 0, time.UTC),
			VisitType: model.VisitTypeRevisit,
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Reservation, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				return appointment, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, &mockTransactor{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:          time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
			PetID:         &petID,
			AppointmentID: &appointmentID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.Equal(t, uint64(123), record.ID)
		assert.False(t, createCalled)
	})

	t.Run("fails closed when appointment duplicate lookup fails", func(t *testing.T) {
		appointmentID := uint64(78)
		petID := uint64(11)
		createCalled := false
		repo := &mockMedicalRecordRepository{
			findByAppointmentIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, errors.New("duplicate lookup failed")
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, PetID: &petID, StartTime: time.Now()}, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:          time.Now(),
			PetID:         &petID,
			AppointmentID: &appointmentID,
		})

		assert.Error(t, err)
		assert.Nil(t, record)
		assert.False(t, createCalled)
	})

	t.Run("normalizes appointment-linked record date to the appointment JST day", func(t *testing.T) {
		appointmentID := uint64(79)
		petID := uint64(12)
		appointmentStart := time.Date(2026, 5, 29, 23, 30, 0, 0, time.UTC)
		var capturedDate time.Time
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, record *model.MedicalRecord) error {
				capturedDate = record.Date
				return nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
				return &model.Reservation{ID: id, PetID: &petID, StartTime: appointmentStart}, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		_, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:          time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			PetID:         &petID,
			AppointmentID: &appointmentID,
		})

		assert.NoError(t, err)
		jst := time.FixedZone("JST", 9*60*60)
		assert.Equal(t, "2026-05-30", capturedDate.In(jst).Format(time.DateOnly))
	})

	t.Run("rejects pet mismatch when appointment already has pet", func(t *testing.T) {
		appointmentID := uint64(77)
		appointmentPetID := uint64(10)
		requestPetID := uint64(99)
		createCalled := false
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				createCalled = true
				return nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _ uint64, id uint64) (*model.Reservation, error) {
				assert.Equal(t, appointmentID, id)
				return &model.Reservation{ID: id, PetID: &appointmentPetID}, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:          time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
			PetID:         &requestPetID,
			AppointmentID: &appointmentID,
		})

		assert.Error(t, err)
		assert.Nil(t, record)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.False(t, createCalled)
	})

	t.Run("fills missing appointment owner and pet before creating record", func(t *testing.T) {
		appointmentID := uint64(77)
		petID := uint64(10)
		ownerID := uint64(20)
		createCalled := false
		var updatedFields map[string]any
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, record *model.MedicalRecord) error {
				createCalled = true
				record.ID = 555
				assert.Equal(t, petID, *record.PetID)
				assert.Equal(t, ownerID, *record.OwnerID)
				return nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _ uint64, id uint64) (*model.Reservation, error) {
				assert.Equal(t, appointmentID, id)
				return &model.Reservation{ID: id}, nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
			backfillFn: func(_ context.Context, _ uint64, id uint64, backfillOwnerID, backfillPetID, backfillDoctorID *uint64) error {
				assert.Equal(t, appointmentID, id)
				updatedFields = make(map[string]any, 3)
				if backfillOwnerID != nil {
					updatedFields["owner_id"] = *backfillOwnerID
				}
				if backfillPetID != nil {
					updatedFields["pet_id"] = *backfillPetID
				}
				if backfillDoctorID != nil {
					updatedFields["doctor_id"] = *backfillDoctorID
				}
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			Date:          time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
			OwnerID:       &ownerID,
			PetID:         &petID,
			AppointmentID: &appointmentID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.True(t, createCalled)
		assert.Equal(t, map[string]any{"owner_id": ownerID, "pet_id": petID}, updatedFields)
	})
}

func TestMedicalRecordService_Create_AppointmentLifecycleGuard(t *testing.T) {
	appointmentID := uint64(77)
	ownerID := uint64(20)
	petID := uint64(10)
	doctorID := uint64(30)
	appointment := &model.Reservation{
		ID:        appointmentID,
		ClinicID:  1,
		OwnerID:   &ownerID,
		PetID:     &petID,
		DoctorID:  &doctorID,
		StartTime: time.Date(2026, 7, 22, 9, 0, 0, 0, time.UTC),
		VisitType: model.VisitTypeRevisit,
	}

	t.Run("linked draft always locks and revalidates the appointment before insert", func(t *testing.T) {
		backfillCalled := false
		recordRepo := &mockMedicalRecordRepository{
			createFn: func(ctx context.Context, record *model.MedicalRecord) error {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, backfillCalled)
				record.ID = 1
				return nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, backfillCalled, "the stable appointment read must happen after the owner lock")
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				return appointment, nil
			},
			backfillFn: func(ctx context.Context, clinicID, id uint64, gotOwnerID, gotPetID, gotDoctorID *uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				backfillCalled = true
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				assert.Equal(t, ownerID, *gotOwnerID)
				assert.Equal(t, petID, *gotPetID)
				assert.Equal(t, doctorID, *gotDoctorID)
				return nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			OwnerID:       &ownerID,
			PetID:         &petID,
			DoctorID:      &doctorID,
			AppointmentID: &appointmentID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.True(t, backfillCalled)
	})

	t.Run("finalized create joins the no-show lifecycle lock before the appointment lock", func(t *testing.T) {
		finalized := model.MedicalRecordStatusFinalized
		prepareCalled := false
		backfillCalled := false
		recordRepo := &mockMedicalRecordRepository{
			createFn: func(ctx context.Context, record *model.MedicalRecord) error {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, prepareCalled)
				assert.True(t, backfillCalled)
				record.ID = 2
				return nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			prepareFinalizationFn: func(ctx context.Context, clinicID, id uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				prepareCalled = true
				assert.False(t, backfillCalled, "lifecycle lock must precede the appointment row lock")
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				return nil
			},
			backfillFn: func(ctx context.Context, _, _ uint64, _, _, _ *uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, prepareCalled)
				backfillCalled = true
				return nil
			},
			findByIDFn: func(ctx context.Context, _, _ uint64) (*model.Reservation, error) {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, backfillCalled)
				return appointment, nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(recordRepo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
			OwnerID:       &ownerID,
			PetID:         &petID,
			DoctorID:      &doctorID,
			AppointmentID: &appointmentID,
			Status:        &finalized,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.True(t, prepareCalled)
		assert.True(t, backfillCalled)
	})
}

func TestMedicalRecordService_Create_SyncsNextVisitTagBestEffort(t *testing.T) {
	ownerID := uint64(10)
	var syncedClinicID, syncedOwnerID uint64

	repo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, record *model.MedicalRecord) error {
			record.ID = 30
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncNextVisitTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	reservationVerifier := &mockReservationRepoForMedicalRecord{
		assertOwnerFn: func(_ context.Context, clinicID, gotOwnerID uint64) error {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, ownerID, gotOwnerID)
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationVerifier, nil, nil, nil, &mockTransactor{}, tagSync)

	record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{
		Date:    time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC),
		OwnerID: &ownerID,
	})

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestMedicalRecordService_Update(t *testing.T) {
	now := time.Now()
	statusFinalized := model.MedicalRecordStatusFinalized
	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0}
	tests := []struct {
		name         string
		input        UpdateMedicalRecordInput
		findByIDErr  error // FindByID のエラー（nil = 正常レコード返却）
		updateErr    error // Update のエラー
		wantErr      bool
		wantNF       bool
		wantConflict bool
	}{
		{
			name: "updates record successfully",
			input: UpdateMedicalRecordInput{
				Date:   &now,
				Status: &statusFinalized,
			},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     false,
			wantNF:      false,
		},
		{
			name:        "returns error when no fields provided",
			input:       UpdateMedicalRecordInput{},
			findByIDErr: nil,
			updateErr:   nil,
			wantErr:     true,
			wantNF:      false,
		},
		{
			name: "returns not found error when record does not exist",
			input: UpdateMedicalRecordInput{
				Status: &statusFinalized,
			},
			findByIDErr: apperrors.WrapNotFound("medical_record", "999"),
			updateErr:   nil,
			wantErr:     true,
			wantNF:      true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateMedicalRecordInput{
				Date: &now,
			},
			findByIDErr: nil,
			updateErr:   errors.New("db error"),
			wantErr:     true,
			wantNF:      false,
		},
		{
			name: "returns conflict when record is already finalized",
			input: UpdateMedicalRecordInput{
				Date: &now,
			},
			findByIDErr:  nil,
			updateErr:    nil,
			wantErr:      true,
			wantNF:       false,
			wantConflict: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isFinalized := tt.name == "returns conflict when record is already finalized"
			existingRecord := validRecord
			if isFinalized {
				existingRecord = &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0, Status: model.MedicalRecordStatusFinalized}
			}
			repo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return existingRecord, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return &model.MedicalRecord{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTreatmentAuditTxLogger{}, &mockTransactor{})

			record, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, record)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestMedicalRecordService_Update_RejectsAppointmentRelink(t *testing.T) {
	appointmentID := uint64(77)
	otherAppointmentID := uint64(88)
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:            1,
				ClinicID:      1,
				AppointmentID: &appointmentID,
				Status:        model.MedicalRecordStatusDraft,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
			t.Fatal("appointment_id must not be changed after medical record creation")
			return nil, nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{
		backfillFn: func(_ context.Context, _, _ uint64, _, _, _ *uint64) error {
			t.Fatal("relink must be rejected before reservation mutation")
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})

	record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{AppointmentID: &otherAppointmentID})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, record)
}

func TestMedicalRecordService_Update_RejectsDateChangeForAppointmentLinkedRecord(t *testing.T) {
	appointmentID := uint64(77)
	changedDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:            1,
				ClinicID:      1,
				AppointmentID: &appointmentID,
				Status:        model.MedicalRecordStatusDraft,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
			t.Fatal("appointment-linked medical record date must not be changed")
			return nil, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, &mockReservationRepoForMedicalRecord{}, nil, nil, nil, medicalRecordTxMarker{})

	record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{Date: &changedDate})

	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, record)
}

func TestMedicalRecordService_Create_FailsClosedWithoutRequiredDependencies(t *testing.T) {
	t.Run("owner or pet input requires reservation verifier", func(t *testing.T) {
		ownerID := uint64(20)
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				t.Fatal("record must not be created without clinic ownership verifier")
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{Date: time.Now(), OwnerID: &ownerID})

		assert.Error(t, err)
		assert.Nil(t, record)
	})

	t.Run("create requires transactor", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				t.Fatal("record must not be created without transaction")
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, &mockReservationRepoForMedicalRecord{}, nil, nil, nil, nil)

		record, err := svc.Create(context.Background(), 1, &CreateMedicalRecordInput{Date: time.Now()})

		assert.Error(t, err)
		assert.Nil(t, record)
	})
}

func TestMedicalRecordService_Update_LinkedContextUsesReservationGuard(t *testing.T) {
	appointmentID := uint64(77)
	ownerID := uint64(20)
	petID := uint64(10)
	doctorID := uint64(30)
	foreignDoctorID := uint64(99)
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:            1,
				ClinicID:      1,
				OwnerID:       &ownerID,
				PetID:         &petID,
				DoctorID:      &doctorID,
				AppointmentID: &appointmentID,
				Status:        model.MedicalRecordStatusDraft,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
			t.Fatal("medical record must not be updated when reservation context validation fails")
			return nil, nil
		},
	}
	guardCalled := false
	reservationRepo := &mockReservationRepoForMedicalRecord{
		backfillFn: func(ctx context.Context, clinicID, id uint64, gotOwnerID, gotPetID, gotDoctorID *uint64) error {
			guardCalled = true
			assertMedicalRecordTxContext(ctx, t)
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, appointmentID, id)
			assert.Equal(t, ownerID, *gotOwnerID)
			assert.Equal(t, petID, *gotPetID)
			assert.Equal(t, foreignDoctorID, *gotDoctorID)
			return apperrors.WrapNotFound("staff", "99")
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})

	record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{DoctorID: &foreignDoctorID})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, record)
	assert.True(t, guardCalled)
}

func TestMedicalRecordService_Update_FinalizationUsesAppointmentLifecycleGuard(t *testing.T) {
	appointmentID := uint64(77)
	finalized := model.MedicalRecordStatusFinalized

	t.Run("guard and update share the transaction context", func(t *testing.T) {
		guardCalled := false
		backfillCalled := false
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID:            1,
					ClinicID:      1,
					AppointmentID: &appointmentID,
					Status:        model.MedicalRecordStatusDraft,
				}, nil
			},
			updateFieldsFn: func(ctx context.Context, _, id uint64, _ map[string]any) (*model.MedicalRecord, error) {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, guardCalled)
				assert.True(t, backfillCalled)
				return &model.MedicalRecord{ID: id, ClinicID: 1, AppointmentID: &appointmentID, Status: finalized}, nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			prepareFinalizationFn: func(ctx context.Context, clinicID, id uint64) error {
				guardCalled = true
				assertMedicalRecordTxContext(ctx, t)
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				return nil
			},
			backfillFn: func(ctx context.Context, clinicID, id uint64, _, _, _ *uint64) error {
				assertMedicalRecordTxContext(ctx, t)
				assert.True(t, guardCalled, "lifecycle lock must precede the appointment context row lock")
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, appointmentID, id)
				backfillCalled = true
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, &mockTreatmentAuditTxLogger{}, medicalRecordTxMarker{})

		record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{Status: &finalized})

		assert.NoError(t, err)
		assert.NotNil(t, record)
		assert.True(t, guardCalled)
		assert.True(t, backfillCalled)
	})

	t.Run("guard conflict prevents finalization", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID:            1,
					ClinicID:      1,
					AppointmentID: &appointmentID,
					Status:        model.MedicalRecordStatusDraft,
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				t.Fatal("medical record must not be finalized when the appointment lifecycle guard rejects it")
				return nil, nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			prepareFinalizationFn: func(_ context.Context, _, _ uint64) error {
				return apperrors.WrapConflict("appointment is no-show")
			},
			backfillFn: func(_ context.Context, _, _ uint64, _, _, _ *uint64) error {
				t.Fatal("appointment context lock must not run when lifecycle guard rejects finalization")
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, nil, medicalRecordTxMarker{})

		record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{Status: &finalized})

		assert.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Nil(t, record)
	})
}

func TestMedicalRecordService_Update_SyncsVisitCompletionAndNextVisitTags(t *testing.T) {
	ownerID := uint64(10)
	statusFinalized := model.MedicalRecordStatusFinalized
	var visitCompletionOwnerID, nextVisitOwnerID uint64

	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 30, ClinicID: 1, Version: 0, Status: model.MedicalRecordStatusDraft}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: 1, OwnerID: &ownerID, Status: statusFinalized}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncVisitCompletionTagsFn: func(_ context.Context, _, ownerID uint64) error {
			visitCompletionOwnerID = ownerID
			return errors.New("sync failed")
		},
		syncNextVisitTagFn: func(_ context.Context, _, ownerID uint64) error {
			nextVisitOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTreatmentAuditTxLogger{}, &mockTransactor{}, tagSync)

	record, err := svc.Update(context.Background(), 1, 30, UpdateMedicalRecordInput{Status: &statusFinalized})

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, ownerID, visitCompletionOwnerID)
	assert.Equal(t, ownerID, nextVisitOwnerID)
}

func TestMedicalRecordService_Update_OwnerValidation(t *testing.T) {
	ownerID := uint64(100)
	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0, Status: model.MedicalRecordStatusDraft}

	t.Run("rejects owner_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				t.Fatal("update must not persist invalid owner")
				return nil, nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			assertOwnerFn: func(_ context.Context, _, _ uint64) error {
				return apperrors.WrapNotFound("owner", "100")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		_, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			OwnerID: &ownerID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("accepts valid owner_id in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: 1, ClinicID: 1}, nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			assertOwnerFn: func(_ context.Context, _, id uint64) error {
				assert.Equal(t, ownerID, id)
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			OwnerID: &ownerID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
	})
}

func TestMedicalRecordService_Update_PetValidation(t *testing.T) {
	petID := uint64(200)

	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0, Status: model.MedicalRecordStatusDraft}

	t.Run("rejects pet_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				t.Fatal("update must not persist invalid pet")
				return nil, nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 0, apperrors.WrapNotFound("pet", "200")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		_, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			PetID: &petID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("accepts valid pet_id in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: 1, ClinicID: 1}, nil
			},
		}
		resRepo := &mockReservationRepoForMedicalRecord{
			findPetOwnerFn: func(_ context.Context, _, id uint64) (uint64, error) {
				assert.Equal(t, petID, id)
				return 1, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, nil, nil, nil, &mockTransactor{})

		_, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			PetID: &petID,
		})

		assert.NoError(t, err)
	})
}

func TestMedicalRecordService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		clinicID      uint64
		id            uint64
		estimateCount int64
		estimateErr   error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:          "deletes record successfully",
			clinicID:      1,
			id:            10,
			estimateCount: 0,
			repoErr:       nil,
			wantErr:       false,
			wantNF:        false,
		},
		{
			name:          "returns conflict error when estimates reference the record",
			clinicID:      1,
			id:            10,
			estimateCount: 2,
			wantErr:       true,
			wantConflict:  true,
		},
		{
			name:        "returns error when estimate count check fails",
			clinicID:    1,
			id:          10,
			estimateErr: errors.New("db error"),
			wantErr:     true,
		},
		{
			name:          "returns not found error when record does not exist",
			clinicID:      1,
			id:            999,
			estimateCount: 0,
			repoErr:       apperrors.WrapNotFound("medical_record", "999"),
			wantErr:       true,
			wantNF:        true,
		},
		{
			name:          "returns error on repository failure",
			clinicID:      1,
			id:            10,
			estimateCount: 0,
			repoErr:       errors.New("db error"),
			wantErr:       true,
			wantNF:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.MedicalRecord{
						ID:       1,
						ClinicID: 1,
						Status:   model.MedicalRecordStatusDraft,
					}, nil
				},
				countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.estimateCount, tt.estimateErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
				if tt.wantConflict {
					assert.True(t, apperrors.IsConflict(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicalRecordService_Delete_SyncsVisitCompletionAndNextVisitTagsAfterDelete(t *testing.T) {
	ownerID := uint64(10)
	deleted := false
	visitCompletionAfterDelete := false
	nextVisitAfterDelete := false

	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       id,
				ClinicID: 1,
				OwnerID:  &ownerID,
				Status:   model.MedicalRecordStatusDraft,
			}, nil
		},
		countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleted = true
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncVisitCompletionTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			visitCompletionAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
		syncNextVisitTagFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			nextVisitAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{}, tagSync)

	err := svc.Delete(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.True(t, visitCompletionAfterDelete)
	assert.True(t, nextVisitAfterDelete)
}

func TestMedicalRecordService_CreateSubRecords_SkipsEmptyInquiryInput(t *testing.T) {
	inquiryRepo := &spyInquiryRepo{}
	svc := NewMedicalRecordServiceWithTxAudit(nil, inquiryRepo, &noopClinicalPlanRepo{}, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})

	_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{})

	assert.False(t, inquiryRepo.called)
}

func TestMedicalRecordService_CreateSubRecords_SavesInquiryWhenInputProvided(t *testing.T) {
	inquiryRepo := &spyInquiryRepo{}
	svc := NewMedicalRecordServiceWithTxAudit(nil, inquiryRepo, &noopClinicalPlanRepo{}, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
	chiefComplaint := "食欲不振"

	_ = svc.CreateSubRecords(context.Background(), 1, 10, CreateSubRecordsInput{
		ChiefComplaint: &chiefComplaint,
	})

	assert.True(t, inquiryRepo.called)
	assert.Equal(t, "食欲不振", inquiryRepo.saved.ChiefComplaint)
}

// noopInquiryRepo は CreateSubRecords テスト用の no-op InquiryRepository
type noopInquiryRepo struct{}

func (n *noopInquiryRepo) SaveByMedicalRecordID(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	return inquiry, nil
}
func (n *noopInquiryRepo) CountByChiefComplaintTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

type spyInquiryRepo struct {
	called bool
	saved  model.Inquiry
}

func (s *spyInquiryRepo) SaveByMedicalRecordID(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	s.called = true
	s.saved = *inquiry
	return inquiry, nil
}
func (s *spyInquiryRepo) CountByChiefComplaintTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// noopClinicalPlanRepo は CreateSubRecords テスト用の no-op ClinicalPlanRepository
type noopClinicalPlanRepo struct{}

func (n *noopClinicalPlanRepo) FindByMedicalRecordID(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
	return nil, apperrors.WrapNotFound("clinical_plan", "0")
}
func (n *noopClinicalPlanRepo) Create(_ context.Context, _ *model.ClinicalPlan) error { return nil }
func (n *noopClinicalPlanRepo) Update(_ context.Context, _, _ uint64, _ map[string]any, _ *int) error {
	return nil
}
func (n *noopClinicalPlanRepo) Delete(_ context.Context, _, _ uint64) error { return nil }

// mockLineCustomerRepository は AutoCreateFromReservation テスト用 LineCustomerRepository モック
type mockLineCustomerRepository struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
}

func (m *mockLineCustomerRepository) FindAll(_ context.Context, _ uint64) ([]model.LineCustomer, error) {
	return nil, nil
}
func (m *mockLineCustomerRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}
func (m *mockLineCustomerRepository) UpdateOwnerLink(_ context.Context, _, _ uint64, _ *uint64) error {
	return nil
}
func (m *mockLineCustomerRepository) FindOrCreateByLineUserID(_ context.Context, _ uint64, _, _ string) (*model.LineCustomer, error) {
	return nil, nil
}
func (m *mockLineCustomerRepository) UpdateAdditionalFields(_ context.Context, _, _ uint64, _ []byte) error {
	return nil
}

// TestAutoCreateFromReservation_BUG386 は LINE予約で owner_id/pet_id が nil のとき
// line_customer から補完してカルテ自動作成できることを検証する（BUG-386 回帰テスト）。
func TestAutoCreateFromReservation_BUG386(t *testing.T) {
	now := time.Now()
	ownerID := uint64(10)
	petID := uint64(20)
	lineCustomerID := uint64(5)

	t.Run("skips when owner_id and pet_id are nil and no line_customer_id", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{
			ID:        1,
			ClinicID:  1,
			StartTime: now,
		}
		// エラーなく静かにスキップするだけ（パニックしない）
		svc.AutoCreateFromReservation(context.Background(), 1, appt)
	})

	t.Run("skips trimming reservation", func(t *testing.T) {
		created := false
		findAllCalled := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				findAllCalled = true
				return nil, 0, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{
			ID:              1,
			ClinicID:        1,
			StartTime:       now,
			OwnerID:         &ownerID,
			PetID:           &petID,
			ReservationType: &model.ReservationType{Category: model.ReservationTypeCategoryTrimming},
		}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)

		assert.False(t, findAllCalled, "トリミング予約では通常カルテ重複チェックに進まない")
		assert.False(t, created, "トリミング予約では通常カルテを作成しない")
	})

	t.Run("skips hotel reservation", func(t *testing.T) {
		created := false
		findAllCalled := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				findAllCalled = true
				return nil, 0, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
		appt := &model.Reservation{
			ID:        1,
			ClinicID:  1,
			StartTime: now,
			OwnerID:   &ownerID,
			PetID:     &petID,
			ReservationType: &model.ReservationType{
				Category: model.ReservationTypeCategoryGeneral,
				Name:     "ペットホテル",
			},
		}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)

		assert.False(t, findAllCalled, "ホテル予約では通常カルテ重複チェックに進まない")
		assert.False(t, created, "ホテル予約では通常カルテを作成しない")
	})

	t.Run("resolves owner_id from line_customer and creates medical record (BUG-386)", func(t *testing.T) {
		var createdRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
			createFn: func(_ context.Context, record *model.MedicalRecord) error {
				createdRecord = record
				return nil
			},
		}
		lineCustomerRepo := &mockLineCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:      lineCustomerID,
					OwnerID: &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Name: "田中太郎",
						Pets: []model.Pet{
							{ID: petID, Name: "ポチ"},
						},
					},
				}, nil
			},
		}
		appt := &model.Reservation{
			ID:             1,
			ClinicID:       1,
			StartTime:      now,
			LineCustomerID: &lineCustomerID,
			// owner_id / pet_id は未設定（LINE予約で自動紐付けが未完了の状態）
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return appt, nil
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return ownerID, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, &noopInquiryRepo{}, &noopClinicalPlanRepo{}, nil, nil, nil, lineCustomerRepo, reservationRepo, nil, nil, nil, &mockTransactor{})

		svc.AutoCreateFromReservation(context.Background(), 1, appt)

		assert.NotNil(t, createdRecord, "カルテが作成されるべき")
		if createdRecord != nil {
			assert.Equal(t, &ownerID, createdRecord.OwnerID)
			assert.Equal(t, &petID, createdRecord.PetID)
		}
	})

	t.Run("skips when line_customer has no owner_id", func(t *testing.T) {
		created := false
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ []uint64, _ MedicalRecordListFilters, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				created = true
				return nil
			},
		}
		lineCustomerRepo := &mockLineCustomerRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:      lineCustomerID,
					OwnerID: nil, // 未紐付け
				}, nil
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, lineCustomerRepo, nil, nil, nil, nil, &mockTransactor{})

		appt := &model.Reservation{
			ID:             1,
			ClinicID:       1,
			StartTime:      now,
			LineCustomerID: &lineCustomerID,
		}

		svc.AutoCreateFromReservation(context.Background(), 1, appt)
		assert.False(t, created, "オーナー未紐付けのためカルテ作成はスキップされるべき")
	})
}

// mockReservationRepoForMedicalRecord は MedicalRecord が消費する予約read/backfillだけを実装する。
type mockReservationRepoForMedicalRecord struct {
	findByIDFn            func(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	assertDoctorFn        func(ctx context.Context, clinicID, doctorID uint64) error
	backfillFn            func(ctx context.Context, clinicID, id uint64, ownerID, petID, doctorID *uint64) error
	prepareFinalizationFn func(ctx context.Context, clinicID, id uint64) error
	assertOwnerFn         func(ctx context.Context, clinicID, ownerID uint64) error
	findPetOwnerFn        func(ctx context.Context, clinicID, petID uint64) (uint64, error)
	findPetByIDFn         func(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	findByIDCallCount     int
}

func (m *mockReservationRepoForMedicalRecord) FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	m.findByIDCallCount++
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockReservationRepoForMedicalRecord) AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error {
	if m.assertDoctorFn != nil {
		return m.assertDoctorFn(ctx, clinicID, doctorID)
	}
	return nil
}

func (m *mockReservationRepoForMedicalRecord) BackfillForMedicalRecord(ctx context.Context, clinicID, id uint64, ownerID, petID, doctorID *uint64) error {
	if m.backfillFn != nil {
		return m.backfillFn(ctx, clinicID, id, ownerID, petID, doctorID)
	}
	return nil
}

func (m *mockReservationRepoForMedicalRecord) PrepareForMedicalRecordFinalization(ctx context.Context, clinicID, id uint64) error {
	if m.prepareFinalizationFn != nil {
		return m.prepareFinalizationFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockReservationRepoForMedicalRecord) AssertOwnerInClinic(ctx context.Context, clinicID, ownerID uint64) error {
	if m.assertOwnerFn != nil {
		return m.assertOwnerFn(ctx, clinicID, ownerID)
	}
	return nil
}

func (m *mockReservationRepoForMedicalRecord) FindPetOwnerInClinic(ctx context.Context, clinicID, petID uint64) (uint64, error) {
	if m.findPetOwnerFn != nil {
		return m.findPetOwnerFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockReservationRepoForMedicalRecord) FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error) {
	if m.findPetByIDFn != nil {
		return m.findPetByIDFn(ctx, clinicID, petID)
	}
	return &model.Pet{ID: petID, Status: model.PetStatusAlive, DeceasedAt: nil}, nil
}


// mockLstepDeliveryTriggerForMR は TriggerFirstVisitWelcome の呼び出し検証用モック
type mockLstepDeliveryTriggerForMR struct {
	triggerFirstVisitWelcomeFn    func(ctx context.Context, clinicID, ownerID uint64) error
	triggerFirstVisitWelcomeCount int
}

func (m *mockLstepDeliveryTriggerForMR) TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID uint64) error {
	m.triggerFirstVisitWelcomeCount++
	if m.triggerFirstVisitWelcomeFn != nil {
		return m.triggerFirstVisitWelcomeFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerCheckupFollowUp(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerFirstVisitFollowUp3D(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerFirstVisitFollowUp7D(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerNextVisitReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerVaccineDeadline60(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerVaccineDeadline30(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerBirthdayMessage(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerDormantPrevention180(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerDormantPrevention210(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerDormantPrevention240(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerDormantPrevention365(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerFilariaAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerFleaTickAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForMR) TriggerFoodRefillReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func TestMedicalRecordService_Create_FirstVisitTrigger(t *testing.T) {
	ownerID := uint64(42)
	clinicID := uint64(1)
	appointmentID := uint64(99)

	newRepo := func(countFn func(ctx context.Context, clinicID, ownerID uint64) (int64, error)) *mockMedicalRecordRepository {
		return &mockMedicalRecordRepository{
			createFn: func(_ context.Context, _ *model.MedicalRecord) error {
				return nil
			},
			countByOwnerIDFn: countFn,
		}
	}

	newRecord := func(apptID *uint64) *model.MedicalRecord {
		return &model.MedicalRecord{
			ClinicID:      clinicID,
			OwnerID:       &ownerID,
			AppointmentID: apptID,
			Date:          time.Now(),
		}
	}

	t.Run("A_reservation_VisitType_first_triggers_welcome_no_count", func(t *testing.T) {
		repo := newRepo(nil)
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{VisitType: model.VisitTypeFirst}, nil
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.NoError(t, err)
		assert.Equal(t, 1, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must be called once")
		assert.Equal(t, 2, resRepo.findByIDCallCount, "FindByID must perform the stable context read and first-visit classification read")
		assert.Equal(t, 0, repo.countByOwnerIDCallCount, "CountByOwnerID must NOT be called when VisitType=first")
	})

	t.Run("B_reservation_VisitType_first_triggers_welcome_ignoring_existing_records", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 5, nil // 既存カルテあり — COUNT が呼ばれた場合は false になるはず
		})
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{VisitType: model.VisitTypeFirst}, nil
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.NoError(t, err)
		assert.Equal(t, 1, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must be called even when owner has prior records")
		assert.Equal(t, 0, repo.countByOwnerIDCallCount, "CountByOwnerID must NOT be called when VisitType=first (performance guarantee)")
	})

	t.Run("C_reservation_VisitType_revisit_does_not_trigger", func(t *testing.T) {
		repo := newRepo(nil)
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{VisitType: model.VisitTypeRevisit}, nil
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.NoError(t, err)
		assert.Equal(t, 0, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must NOT be called for revisit")
		assert.Equal(t, 0, repo.countByOwnerIDCallCount, "CountByOwnerID must NOT be called when VisitType=revisit")
	})

	t.Run("D_appointment_with_empty_VisitType_falls_back_to_COUNT_zero_records", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil // カルテ未作成 → 初診
		})
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{VisitType: ""}, nil // VisitType 空
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.NoError(t, err)
		assert.Equal(t, 1, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must be called when COUNT=0")
		assert.Equal(t, 1, repo.countByOwnerIDCallCount, "CountByOwnerID must be called for empty VisitType fallback")
	})

	t.Run("E_appointment_with_empty_VisitType_falls_back_to_COUNT_existing_records", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 3, nil // 既存カルテあり → 再診
		})
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return &model.Reservation{VisitType: ""}, nil
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.NoError(t, err)
		assert.Equal(t, 0, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must NOT be called when COUNT>0")
		assert.Equal(t, 1, repo.countByOwnerIDCallCount, "CountByOwnerID must be called for empty VisitType fallback")
	})

	t.Run("F_no_AppointmentID_falls_back_to_COUNT_zero_records", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil
		})
		resRepo := &mockReservationRepoForMedicalRecord{}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(nil)) // no AppointmentID

		assert.NoError(t, err)
		assert.Equal(t, 1, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must be called for walk-in with COUNT=0")
		assert.Equal(t, 0, resRepo.findByIDCallCount, "FindByID must NOT be called when AppointmentID is nil")
		assert.Equal(t, 1, repo.countByOwnerIDCallCount, "CountByOwnerID must be called for walk-in fallback")
	})

	t.Run("G_no_AppointmentID_falls_back_to_COUNT_existing_records", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 2, nil
		})
		resRepo := &mockReservationRepoForMedicalRecord{}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(nil))

		assert.NoError(t, err)
		assert.Equal(t, 0, trigger.triggerFirstVisitWelcomeCount, "TriggerFirstVisitWelcome must NOT be called for walk-in with COUNT>0")
		assert.Equal(t, 0, resRepo.findByIDCallCount, "FindByID must NOT be called when AppointmentID is nil")
		assert.Equal(t, 1, repo.countByOwnerIDCallCount, "CountByOwnerID must be called for walk-in fallback")
	})

	t.Run("H_reservation_lookup_error_fails_closed_before_create", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil
		})
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return nil, errors.New("db error")
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.Error(t, err, "appointment-linked medical record creation must fail closed when the stable appointment read fails")
		assert.Equal(t, 0, trigger.triggerFirstVisitWelcomeCount)
		assert.Equal(t, 1, resRepo.findByIDCallCount, "the stable appointment read must be attempted")
		assert.Equal(t, 0, repo.countByOwnerIDCallCount, "first-visit fallback must not run after lifecycle validation fails")
	})

	t.Run("I_reservation_nil_fails_closed_before_create", func(t *testing.T) {
		repo := newRepo(func(_ context.Context, _, _ uint64) (int64, error) {
			return 0, nil
		})
		resRepo := &mockReservationRepoForMedicalRecord{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Reservation, error) {
				return nil, nil // nil reservation（予約が見つからない）
			},
		}
		trigger := &mockLstepDeliveryTriggerForMR{}

		svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, resRepo, trigger, nil, nil, &mockTransactor{})
		_, err := createMedicalRecordForTest(context.Background(), svc, newRecord(&appointmentID))

		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Equal(t, 0, trigger.triggerFirstVisitWelcomeCount)
		assert.Equal(t, 1, resRepo.findByIDCallCount, "the stable appointment read must be attempted")
		assert.Equal(t, 0, repo.countByOwnerIDCallCount, "first-visit fallback must not run without a stable appointment")
	})
}

func TestMedicalRecordService_UpdateRecommendationReason(t *testing.T) {
	const (
		clinicID uint64 = 1
		recordID uint64 = 10
	)

	newSvc := func(repo *mockMedicalRecordRepository) MedicalRecordService {
		return NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, &mockTransactor{})
	}

	t.Run("valid_reason_revisit", func(t *testing.T) {
		reason := "revisit"
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID, RecommendationReason: &reason}, nil
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "revisit"})
		assert.NoError(t, err)
		assert.Equal(t, "revisit", *rec.RecommendationReason)
	})

	t.Run("valid_reason_checkup", func(t *testing.T) {
		reason := "checkup"
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID, RecommendationReason: &reason}, nil
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "checkup"})
		assert.NoError(t, err)
		assert.Equal(t, "checkup", *rec.RecommendationReason)
	})

	t.Run("valid_reason_prevention", func(t *testing.T) {
		reason := "prevention"
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID, RecommendationReason: &reason}, nil
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "prevention"})
		assert.NoError(t, err)
		assert.Equal(t, "prevention", *rec.RecommendationReason)
	})

	t.Run("valid_reason_exam", func(t *testing.T) {
		reason := "exam"
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID, RecommendationReason: &reason}, nil
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "exam"})
		assert.NoError(t, err)
		assert.Equal(t, "exam", *rec.RecommendationReason)
	})

	t.Run("rejects_finalized_parent", func(t *testing.T) {
		updateCalled := false
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID, Status: model.MedicalRecordStatusFinalized}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				updateCalled = true
				return nil, errors.New("must not be called")
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "revisit"})
		assert.Error(t, err)
		assert.Nil(t, rec)
		assert.True(t, apperrors.IsConflict(err), "確定済みカルテへの推奨理由更新は Conflict(409) であるべき: %v", err)
		assert.False(t, updateCalled, "確定済みカルテの推奨理由は更新されてはならない")
	})

	t.Run("empty_reason_clears_to_nil", func(t *testing.T) {
		var capturedFields map[string]any
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{ID: recordID}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.MedicalRecord, error) {
				capturedFields = fields
				return &model.MedicalRecord{ID: recordID, RecommendationReason: nil}, nil
			},
		}
		rec, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: ""})
		assert.NoError(t, err)
		assert.Nil(t, rec.RecommendationReason)
		assert.Nil(t, capturedFields["recommendation_reason"], "empty reason must write nil to DB")
	})

	t.Run("invalid_reason_returns_error", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{}
		_, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "invalid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recommendation_reason must be one of")
	})

	t.Run("uppercase_REVISIT_is_rejected", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{}
		_, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "REVISIT"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recommendation_reason must be one of")
	})

	t.Run("not_found_returns_error", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return nil, apperrors.WrapNotFound("medical_record", "10")
			},
		}
		_, err := newSvc(repo).UpdateRecommendationReason(context.Background(), clinicID, recordID, UpdateRecommendationReasonInput{Reason: "revisit"})
		assert.Error(t, err)
	})

	t.Run("wrong_clinic_id_returns_not_found", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, cID, _ uint64) (*model.MedicalRecord, error) {
				if cID != clinicID {
					return nil, apperrors.WrapNotFound("medical_record", "10")
				}
				return &model.MedicalRecord{ID: recordID}, nil
			},
		}
		_, err := newSvc(repo).UpdateRecommendationReason(context.Background(), 999, recordID, UpdateRecommendationReasonInput{Reason: "revisit"})
		assert.Error(t, err)
	})
}

// TestMedicalRecordService_Create_AuditLog はカルテ作成時に audit "create" が記録されることを確認する。
func TestMedicalRecordService_Create_AuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	repo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, r *model.MedicalRecord) error {
			r.ID = 50
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, nil, &mockTransactor{})

	enteredBy := uint64(7)
	record := &model.MedicalRecord{
		ClinicID:  1,
		Date:      time.Now(),
		Status:    model.MedicalRecordStatusDraft,
		EnteredBy: &enteredBy,
	}
	created, err := createMedicalRecordForTest(context.Background(), svc, record)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Contains(t, auditSvc.calls, "create", "create 操作が audit に記録されること")
}

// TestMedicalRecordService_Update_AuditLog はカルテ更新時に audit "update" が記録されることを確認する。
func TestMedicalRecordService_Update_AuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	draftStatus := model.MedicalRecordStatusDraft
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       10,
				ClinicID: 1,
				Status:   model.MedicalRecordStatusDraft,
				DoctorID: ptrUint64(5),
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       10,
				ClinicID: 1,
				Status:   model.MedicalRecordStatusDraft,
				DoctorID: ptrUint64(99),
			}, nil
		},
	}
	reservationRepo := &mockReservationRepoForMedicalRecord{}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, auditSvc, nil, &mockTransactor{})

	staffID := uint64(7)
	_, err := svc.Update(context.Background(), 1, 10, UpdateMedicalRecordInput{
		Status:   &draftStatus,
		DoctorID: ptrUint64(99),
		ActorID:  &staffID,
	})
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "update", "update 操作が audit に記録されること")
}

// TestMedicalRecordService_Update_FinalizeAuditLog はカルテ確定時に "finalize" 監査ログが追加記録されることを確認する。
func TestMedicalRecordService_Update_FinalizeAuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	auditTx := &mockTreatmentAuditTxLogger{}
	finalizedStatus := model.MedicalRecordStatusFinalized
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       10,
				ClinicID: 1,
				Status:   model.MedicalRecordStatusDraft,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID:       10,
				ClinicID: 1,
				Status:   model.MedicalRecordStatusFinalized,
			}, nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, auditTx, &mockTransactor{})

	staffID := uint64(7)
	_, err := svc.Update(context.Background(), 1, 10, UpdateMedicalRecordInput{
		Status:  &finalizedStatus,
		ActorID: &staffID,
	})
	assert.NoError(t, err)
	require.Len(t, auditTx.entries, 1)
	entry := auditTx.entries[0]
	require.NotNil(t, entry.ClinicID)
	require.NotNil(t, entry.ActorID)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, uint64(1), *entry.ClinicID)
	assert.Equal(t, staffID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, "finalize", entry.Action)
	assert.Equal(t, "medical_record", entry.Resource)
	assert.Equal(t, uint64(10), *entry.ResourceID)
	assert.Nil(t, entry.OldValue)
	assert.Equal(t, "finalized", entry.NewValue.(map[string]any)["status"])
	assert.Nil(t, entry.Metadata)
}

// TestMedicalRecordService_Delete_AuditLog はカルテ削除時に audit "delete" が記録されることを確認する。
func TestMedicalRecordService_Delete_AuditLog(t *testing.T) {
	auditSvc := &mockAuditService{}
	repo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: 10, ClinicID: 1, Status: model.MedicalRecordStatusDraft}, nil
		},
		countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
			return 0, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, nil, &mockTransactor{})

	err := svc.Delete(context.Background(), 1, 10)
	assert.NoError(t, err)
	assert.Contains(t, auditSvc.calls, "delete", "delete 操作が audit に記録されること")
}

// TestMedicalRecordService_AuditFailure_NonFatal は監査ログ書き込み失敗がメイン処理を止めないことを確認する。
func TestMedicalRecordService_AuditFailure_NonFatal(t *testing.T) {
	auditSvc := &mockAuditService{
		logMedicalRecordChangeFn: func(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
			return errors.New("audit db down")
		},
	}
	repo := &mockMedicalRecordRepository{
		createFn: func(_ context.Context, r *model.MedicalRecord) error {
			r.ID = 1
			return nil
		},
	}
	svc := NewMedicalRecordServiceWithTxAudit(repo, nil, nil, nil, nil, nil, nil, nil, nil, auditSvc, nil, &mockTransactor{})

	record := &model.MedicalRecord{ClinicID: 1, Date: time.Now(), Status: model.MedicalRecordStatusDraft}
	created, err := createMedicalRecordForTest(context.Background(), svc, record)
	assert.NoError(t, err, "監査ログ失敗はメイン処理のエラーを返さない（best-effort）")
	assert.NotNil(t, created)
}

// SEC-DUR-01-MR-T1: 譲渡後のカルテ確定は snapshot owner と current pet owner の差を許容し clinic 外は拒否する。
func TestMedicalRecordService_Update_Finalize_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	const (
		clinicID        = uint64(1)
		recordID        = uint64(10)
		previousOwnerID = uint64(20)
		currentOwnerID  = uint64(21)
		petID           = uint64(30)
	)
	finalizedStatus := model.MedicalRecordStatusFinalized
	staffID := uint64(7)

	t.Run("same_clinic_transfer_succeeds", func(t *testing.T) {
		updateCalls := 0
		auditTx := &mockTreatmentAuditTxLogger{}
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: recordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft, Version: 1,
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.MedicalRecord, error) {
				updateCalls++
				assert.Equal(t, model.MedicalRecordStatusFinalized, fields["status"])
				return &model.MedicalRecord{
					ID: recordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusFinalized, Version: 2,
				}, nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			assertOwnerFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(
			repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, auditTx, &mockTransactor{},
		)

		got, err := svc.Update(context.Background(), clinicID, recordID, UpdateMedicalRecordInput{
			Status:  &finalizedStatus,
			ActorID: &staffID,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, 1, updateCalls)
		require.Len(t, auditTx.entries, 1)
		assert.Equal(t, "finalize", auditTx.entries[0].Action)
	})

	t.Run("rejects_foreign_snapshot_owner", func(t *testing.T) {
		updateCalls := 0
		auditTx := &mockTreatmentAuditTxLogger{}
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: recordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft, Version: 1,
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				updateCalls++
				return &model.MedicalRecord{ID: recordID, Status: model.MedicalRecordStatusFinalized}, nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			assertOwnerFn: func(_ context.Context, _, ownerID uint64) error {
				assert.Equal(t, previousOwnerID, ownerID)
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(
			repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, auditTx, &mockTransactor{},
		)

		got, err := svc.Update(context.Background(), clinicID, recordID, UpdateMedicalRecordInput{
			Status:  &finalizedStatus,
			ActorID: &staffID,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
		assert.Empty(t, auditTx.entries)
	})

	t.Run("rejects_foreign_pet", func(t *testing.T) {
		updateCalls := 0
		auditTx := &mockTreatmentAuditTxLogger{}
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return &model.MedicalRecord{
					ID: recordID, ClinicID: clinicID,
					OwnerID: ptrUint64(previousOwnerID), PetID: ptrUint64(petID),
					Status: model.MedicalRecordStatusDraft, Version: 1,
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
				updateCalls++
				return &model.MedicalRecord{ID: recordID, Status: model.MedicalRecordStatusFinalized}, nil
			},
		}
		reservationRepo := &mockReservationRepoForMedicalRecord{
			assertOwnerFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		svc := NewMedicalRecordServiceWithTxAudit(
			repo, nil, nil, nil, nil, nil, nil, reservationRepo, nil, nil, auditTx, &mockTransactor{},
		)

		got, err := svc.Update(context.Background(), clinicID, recordID, UpdateMedicalRecordInput{
			Status:  &finalizedStatus,
			ActorID: &staffID,
		})
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, updateCalls)
		assert.Empty(t, auditTx.entries)
	})
}
