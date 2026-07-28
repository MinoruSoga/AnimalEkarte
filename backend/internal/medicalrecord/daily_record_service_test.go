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

// ---- DailyRecord モック ----

type mockDailyRecordRepository struct {
	listByHospitalizationIDFn        func(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	getOrCreateByDateFn              func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	findByHospitalizationIDAndDateFn func(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	createVitalRecordFn              func(ctx context.Context, vr *model.VitalRecord) error
	createCareLogFn                  func(ctx context.Context, cr *model.CareLog) error
	createStaffNoteFn                func(ctx context.Context, sn *model.StaffNote) error
}

func (m *mockDailyRecordRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	return m.listByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockDailyRecordRepository) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.getOrCreateByDateFn(ctx, clinicID, hospitalizationID, date)
}

func (m *mockDailyRecordRepository) FindByHospitalizationIDAndDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.findByHospitalizationIDAndDateFn(ctx, clinicID, hospitalizationID, date)
}

func (m *mockDailyRecordRepository) CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error {
	return m.createVitalRecordFn(ctx, vr)
}

func (m *mockDailyRecordRepository) CreateCareLog(ctx context.Context, cr *model.CareLog) error {
	return m.createCareLogFn(ctx, cr)
}

func (m *mockDailyRecordRepository) CreateStaffNote(ctx context.Context, sn *model.StaffNote) error {
	return m.createStaffNoteFn(ctx, sn)
}

// okHospRepoForDailyRecord は親入院の所有権検証が成功する（同一クリニック）モックを返す。
func okHospRepoForDailyRecord() *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, hospitalizationID uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID: hospitalizationID, ClinicID: clinicID, PetID: 42,
			}, nil
		},
	}
}

type dailyRecordOwnerPetVerifierStub struct {
	ownerID uint64
	err     error
}

func (s *dailyRecordOwnerPetVerifierStub) AssertOwnerInClinic(
	context.Context,
	uint64,
	uint64,
) error {
	return s.err
}

func (s *dailyRecordOwnerPetVerifierStub) FindPetOwnerInClinic(
	context.Context,
	uint64,
	uint64,
) (uint64, error) {
	return s.ownerID, s.err
}

func newDailyRecordServiceForTest(
	repo DailyRecordRepository,
	hospRepo HospitalizationRepository,
	tx Transactor,
) DailyRecordService {
	return NewDailyRecordServiceWithRelationValidation(
		repo,
		hospRepo,
		&dailyRecordOwnerPetVerifierStub{},
		nil,
		nil,
		tx,
	)
}

// ---- Tests ----

func TestDailyRecordService_List(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	tests := []struct {
		name              string
		hospitalizationID uint64
		repoRecords       []model.DailyRecord
		repoErr           error
		wantLen           int
		wantErr           bool
	}{
		{
			name:              "returns daily records for hospitalization",
			hospitalizationID: 1,
			repoRecords: []model.DailyRecord{
				{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today},
				{ID: 2, ClinicID: 1, HospitalizationID: 1, Date: today.AddDate(0, 0, 1)},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:              "returns empty list when no records exist",
			hospitalizationID: 999,
			repoRecords:       []model.DailyRecord{},
			repoErr:           nil,
			wantLen:           0,
			wantErr:           false,
		},
		{
			name:              "propagates repository error",
			hospitalizationID: 1,
			repoRecords:       nil,
			repoErr:           errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.DailyRecord, error) {
					return tt.repoRecords, tt.repoErr
				},
			}
			svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

			records, err := svc.List(context.Background(), uint64(1), tt.hospitalizationID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, records, tt.wantLen)
			}
		})
	}
}

func TestDailyRecordService_FindOrCreateByDate(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "returns existing daily record",
			hospitalizationID: 1,
			date:              today,
			repoErr:           nil,
			wantErr:           false,
		},
		{
			name:              "creates new daily record when not found",
			hospitalizationID: 1,
			date:              today,
			repoErr:           nil,
			wantErr:           false,
		},
		{
			name:              "returns error when repository fails",
			hospitalizationID: 1,
			date:              today,
			repoErr:           errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

			record, err := svc.FindOrCreateByDate(context.Background(), uint64(1), tt.hospitalizationID, tt.date)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
				assert.Equal(t, tt.hospitalizationID, record.HospitalizationID)
			}
		})
	}
}

func TestDailyRecordService_AddVitalRecord(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	recordTime := today.Add(9 * time.Hour)
	temp := 37.5
	hr := 80
	respRate := 20
	weight := 5.5
	staffID := uint64(10)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		input             *CreateVitalRecordInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "adds vital record successfully",
			hospitalizationID: 1,
			date:              today,
			input: &CreateVitalRecordInput{
				Time:            recordTime,
				Temperature:     &temp,
				HeartRate:       &hr,
				RespirationRate: &respRate,
				Weight:          &weight,
				StaffID:         &staffID,
				Notes:           "Normal vitals",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "adds vital record with minimal fields",
			hospitalizationID: 1,
			date:              today,
			input: &CreateVitalRecordInput{
				Time: recordTime,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error when get_or_create fails",
			hospitalizationID: 1,
			date:              today,
			input: &CreateVitalRecordInput{
				Time: recordTime,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createVitalRecordFn: func(_ context.Context, vital *model.VitalRecord) error {
					assert.Equal(t, uint64(1), vital.ClinicID)
					assert.Equal(t, uint64(42), vital.PetID, "PetID must be resolved from parent hospitalization")
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := NewDailyRecordServiceWithRelationValidation(
				repo,
				okHospRepoForDailyRecord(),
				&dailyRecordOwnerPetVerifierStub{},
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: 1},
				},
				&mockTransactor{},
			)

			record, err := svc.AddVitalRecord(context.Background(), uint64(1), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestDailyRecordService_AddVitalRecord_CreateError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
			return errors.New("db error creating vital record")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddVitalRecord(context.Background(), 1, 1, today, &CreateVitalRecordInput{Time: today})

	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDailyRecordService_AddVitalRecord_RefetchError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
			return nil
		},
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return nil, errors.New("db error on refetch")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddVitalRecord(context.Background(), 1, 1, today, &CreateVitalRecordInput{Time: today})

	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDailyRecordService_AddCareLog(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	staffID := uint64(10)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		input             *CreateCareLogInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "adds care log record with food type",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogInput{
				Time:    "10:00:00",
				Type:    string(model.CareLogTypeFood),
				Status:  string(model.CareLogStatusCompleted),
				Value:   "100ml",
				StaffID: &staffID,
				Notes:   "All consumed",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "adds care log record with default status",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogInput{
				Time:   "10:00:00",
				Type:   string(model.CareLogTypeMedicine),
				Status: "", // Will default to Completed
				Value:  "2 tablets",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error on invalid type",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogInput{
				Time:   "10:00:00",
				Type:   "invalid_type",
				Status: string(model.CareLogStatusCompleted),
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:              "returns error on invalid status",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogInput{
				Time:   "10:00:00",
				Type:   string(model.CareLogTypeFood),
				Status: "invalid_status",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:              "returns error when get_or_create fails",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogInput{
				Time:   "10:00:00",
				Type:   string(model.CareLogTypeFood),
				Status: string(model.CareLogStatusCompleted),
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createCareLogFn: func(_ context.Context, careLog *model.CareLog) error {
					assert.Equal(t, uint64(1), careLog.ClinicID)
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := NewDailyRecordServiceWithRelationValidation(
				repo,
				okHospRepoForDailyRecord(),
				&dailyRecordOwnerPetVerifierStub{},
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: 1},
				},
				&mockTransactor{},
			)

			record, err := svc.AddCareLog(context.Background(), uint64(1), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestDailyRecordService_AddCareLog_CreateError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createCareLogFn: func(_ context.Context, _ *model.CareLog) error {
			return errors.New("db error creating care log")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddCareLog(context.Background(), 1, 1, today, &CreateCareLogInput{
		Time: "10:00:00", Type: string(model.CareLogTypeFood), Status: string(model.CareLogStatusCompleted),
	})

	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDailyRecordService_AddCareLog_RefetchError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createCareLogFn: func(_ context.Context, _ *model.CareLog) error {
			return nil
		},
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return nil, errors.New("db error on refetch")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddCareLog(context.Background(), 1, 1, today, &CreateCareLogInput{
		Time: "10:00:00", Type: string(model.CareLogTypeFood), Status: string(model.CareLogStatusCompleted),
	})

	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDailyRecordService_AddStaffNote(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	staffID := uint64(10)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		input             *CreateStaffNoteInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "adds staff note record successfully",
			hospitalizationID: 1,
			date:              today,
			input: &CreateStaffNoteInput{
				Time:    "09:00:00",
				Content: "Patient showing improvement",
				StaffID: &staffID,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "adds staff note with minimal fields",
			hospitalizationID: 1,
			date:              today,
			input: &CreateStaffNoteInput{
				Time:    "14:30:00",
				Content: "Brief note",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error when get_or_create fails",
			hospitalizationID: 1,
			date:              today,
			input: &CreateStaffNoteInput{
				Time:    "09:00:00",
				Content: "Test note",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := NewDailyRecordServiceWithRelationValidation(
				repo,
				okHospRepoForDailyRecord(),
				&dailyRecordOwnerPetVerifierStub{},
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: 1},
				},
				&mockTransactor{},
			)

			record, err := svc.AddStaffNote(context.Background(), uint64(1), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestDailyRecordService_AddStaffNote_CreateError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
			return errors.New("db error creating staff note")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddStaffNote(context.Background(), 1, 1, today, &CreateStaffNoteInput{Time: "10:00:00", Content: "note"})

	assert.Error(t, err)
	assert.Nil(t, record)
}

func TestDailyRecordService_AddStaffNote_RefetchError(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
			return nil
		},
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return nil, errors.New("db error on refetch")
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.AddStaffNote(context.Background(), 1, 1, today, &CreateStaffNoteInput{Time: "10:00:00", Content: "note"})

	assert.Error(t, err)
	assert.Nil(t, record)
}

// ---- AUD-003: GetByDate read-only + parent clinic isolation ----

func TestDailyRecordService_GetByDate(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	findOrCreateCalled := false
	repo := &mockDailyRecordRepository{
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 7, ClinicID: 1, HospitalizationID: hospitalizationID, Date: date}, nil
		},
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			findOrCreateCalled = true
			return &model.DailyRecord{}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.GetByDate(context.Background(), 1, 1, today)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, uint64(7), record.ID)
	assert.False(t, findOrCreateCalled, "GetByDate must not call FindOrCreateByDate")
}

func TestDailyRecordService_GetByDate_NotFoundWhenMissing(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	findOrCreateCalled := false
	repo := &mockDailyRecordRepository{
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return nil, apperrors.WrapNotFound("daily_record", "1/2026-07-01")
		},
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			findOrCreateCalled = true
			return &model.DailyRecord{}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, okHospRepoForDailyRecord(), &mockTransactor{})

	record, err := svc.GetByDate(context.Background(), 1, 1, today)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "missing date must be NotFound: %v", err)
	assert.Nil(t, record)
	assert.False(t, findOrCreateCalled, "missing-date GET must not FirstOrCreate")
}

func crossTenantHospRepo() *mockHospitalizationRepository {
	return &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "99")
		},
	}
}

func TestDailyRecordService_List_CrossTenantParentRejected(t *testing.T) {
	listCalled := false
	repo := &mockDailyRecordRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.DailyRecord, error) {
			listCalled = true
			return []model.DailyRecord{}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	records, err := svc.List(context.Background(), 1, 99)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, records)
	assert.False(t, listCalled)
}

func TestDailyRecordService_GetByDate_CrossTenantParentRejected(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	findCalled := false
	findOrCreateCalled := false
	repo := &mockDailyRecordRepository{
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			findCalled = true
			return &model.DailyRecord{}, nil
		},
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			findOrCreateCalled = true
			return &model.DailyRecord{}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	record, err := svc.GetByDate(context.Background(), 1, 99, today)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, record)
	assert.False(t, findCalled)
	assert.False(t, findOrCreateCalled)
}

func TestDailyRecordService_FindOrCreateByDate_CrossTenantParentRejected(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	findOrCreateCalled := false
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			findOrCreateCalled = true
			return &model.DailyRecord{ID: 1}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	record, err := svc.FindOrCreateByDate(context.Background(), 1, 99, today)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, record)
	assert.False(t, findOrCreateCalled, "cross-clinic Create must not persist a row")
}

func TestDailyRecordService_FindOrCreateByDate_OwnerClinicSucceedsAfterCrossTenantReject(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	const clinicA, clinicB, hospB = uint64(1), uint64(2), uint64(99)

	findOrCreateCalls := 0
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
			findOrCreateCalls++
			return &model.DailyRecord{ID: 42, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date}, nil
		},
	}
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
			if clinicID == clinicB && id == hospB {
				return &model.Hospitalization{ID: hospB, ClinicID: clinicB, PetID: 100}, nil
			}
			return nil, apperrors.WrapNotFound("hospitalization", "99")
		},
	}
	svc := newDailyRecordServiceForTest(repo, hospRepo, &mockTransactor{})

	rejected, err := svc.FindOrCreateByDate(context.Background(), clinicA, hospB, today)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, rejected)
	assert.Equal(t, 0, findOrCreateCalls)

	created, err := svc.FindOrCreateByDate(context.Background(), clinicB, hospB, today)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, uint64(42), created.ID)
	assert.Equal(t, 1, findOrCreateCalls)
}

func TestDailyRecordService_AddVitalRecord_CrossTenantParentRejected(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	createCalled := false
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1}, nil
		},
		createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
			createCalled = true
			return nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	record, err := svc.AddVitalRecord(context.Background(), 1, 99, today, &CreateVitalRecordInput{Time: today})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, record)
	assert.False(t, createCalled)
}

func TestDailyRecordService_AddCareLog_CrossTenantParentRejected(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	createCalled := false
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1}, nil
		},
		createCareLogFn: func(_ context.Context, _ *model.CareLog) error {
			createCalled = true
			return nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	record, err := svc.AddCareLog(context.Background(), 1, 99, today, &CreateCareLogInput{
		Time: "10:00:00",
		Type: string(model.CareLogTypeFood),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, record)
	assert.False(t, createCalled)
}

func TestDailyRecordService_AddStaffNote_CrossTenantParentRejected(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	createCalled := false
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 1}, nil
		},
		createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
			createCalled = true
			return nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, crossTenantHospRepo(), &mockTransactor{})

	record, err := svc.AddStaffNote(context.Background(), 1, 99, today, &CreateStaffNoteInput{Time: "10:00:00", Content: "x"})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "%v", err)
	assert.Nil(t, record)
	assert.False(t, createCalled)
}

func TestDailyRecordService_AddVitalRecord_SetsPetIDFromHospitalization(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	var gotPetID uint64
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 7, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
		createVitalRecordFn: func(_ context.Context, vital *model.VitalRecord) error {
			gotPetID = vital.PetID
			return nil
		},
		findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{ID: 7, ClinicID: 1, HospitalizationID: 1, Date: today}, nil
		},
	}
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, ClinicID: clinicID, PetID: 99}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, hospRepo, &mockTransactor{})
	_, err := svc.AddVitalRecord(context.Background(), 1, 1, today, &CreateVitalRecordInput{Time: today})
	require.NoError(t, err)
	assert.Equal(t, uint64(99), gotPetID)
}

func TestDailyRecordService_FindOrCreateRejectsInconsistentParentLinks(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
	)
	today := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		hospitalization *model.Hospitalization
		daily           *model.DailyRecord
	}{
		{
			name: "hospitalization snapshot belongs to another clinic",
			hospitalization: &model.Hospitalization{
				ID: hospitalizationID, ClinicID: clinicID + 1, PetID: 20,
			},
			daily: &model.DailyRecord{
				ID: 1, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: today,
			},
		},
		{
			name: "repository returns daily record for another hospitalization",
			hospitalization: &model.Hospitalization{
				ID: hospitalizationID, ClinicID: clinicID, PetID: 20,
			},
			daily: &model.DailyRecord{
				ID: 1, ClinicID: clinicID, HospitalizationID: hospitalizationID + 1, Date: today,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findOrCreateCalls := 0
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					findOrCreateCalls++
					return tt.daily, nil
				},
			}
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return tt.hospitalization, nil
				},
			}
			svc := newDailyRecordServiceForTest(repo, hospRepo, &mockTransactor{})

			got, err := svc.FindOrCreateByDate(context.Background(), clinicID, hospitalizationID, today)

			assert.Error(t, err)
			assert.Nil(t, got)
			if tt.hospitalization.ClinicID != clinicID {
				assert.Zero(t, findOrCreateCalls)
			}
		})
	}
}

func TestDailyRecordService_AddEntriesRejectsUnassignedStaffBeforeWrite(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
		petID             = uint64(20)
		staffID           = uint64(30)
	)
	today := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		call func(DailyRecordService) (*model.DailyRecord, error)
	}{
		{
			name: "vital",
			call: func(svc DailyRecordService) (*model.DailyRecord, error) {
				return svc.AddVitalRecord(context.Background(), clinicID, hospitalizationID, today, &CreateVitalRecordInput{
					Time: today.Add(9 * time.Hour), StaffID: ptrUint64(staffID),
				})
			},
		},
		{
			name: "care log",
			call: func(svc DailyRecordService) (*model.DailyRecord, error) {
				return svc.AddCareLog(context.Background(), clinicID, hospitalizationID, today, &CreateCareLogInput{
					Time: "09:00:00", Type: string(model.CareLogTypeFood), StaffID: ptrUint64(staffID),
				})
			},
		},
		{
			name: "staff note",
			call: func(svc DailyRecordService) (*model.DailyRecord, error) {
				return svc.AddStaffNote(context.Background(), clinicID, hospitalizationID, today, &CreateStaffNoteInput{
					Time: "09:00:00", Content: "handoff", StaffID: ptrUint64(staffID),
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeCalls := 0
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{
						ID: 1, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: today,
					}, nil
				},
				createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
					writeCalls++
					return nil
				},
				createCareLogFn: func(_ context.Context, _ *model.CareLog) error {
					writeCalls++
					return nil
				},
				createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
					writeCalls++
					return nil
				},
			}
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID: hospitalizationID, ClinicID: clinicID, PetID: petID,
					}, nil
				},
			}
			svc := NewDailyRecordServiceWithRelationValidation(
				repo,
				hospRepo,
				&dailyRecordOwnerPetVerifierStub{},
				&clinicalStaffLockerStub{staff: &model.Staff{ID: staffID, IsActive: true}},
				&clinicalStaffAssignmentLockerStub{
					assignment: &model.StaffClinicAssignment{StaffID: staffID, ClinicID: clinicID + 1},
				},
				&mockTransactor{},
			)

			got, err := tt.call(svc)

			assert.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, writeCalls)
		})
	}
}

func TestDailyRecordService_AddStaffNoteFailsClosedWithoutStaffDependencies(t *testing.T) {
	const (
		clinicID          = uint64(1)
		hospitalizationID = uint64(10)
	)
	today := time.Date(2026, 7, 24, 0, 0, 0, 0, time.UTC)
	writeCalls := 0
	repo := &mockDailyRecordRepository{
		getOrCreateByDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
			return &model.DailyRecord{
				ID: 1, ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: today,
			}, nil
		},
		createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
			writeCalls++
			return nil
		},
	}
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID: hospitalizationID, ClinicID: clinicID, PetID: 20,
			}, nil
		},
	}
	svc := newDailyRecordServiceForTest(repo, hospRepo, &mockTransactor{})

	got, err := svc.AddStaffNote(context.Background(), clinicID, hospitalizationID, today, &CreateStaffNoteInput{
		Time: "09:00:00", Content: "handoff", StaffID: ptrUint64(30),
	})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, writeCalls)
}
