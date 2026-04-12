package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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

func (m *mockDailyRecordRepository) ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	return m.listByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockDailyRecordRepository) GetOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
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
				{ID: 1, HospitalizationID: 1, Date: today},
				{ID: 2, HospitalizationID: 1, Date: today.AddDate(0, 0, 1)},
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
			svc := NewDailyRecordService(repo)

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

func TestDailyRecordService_GetOrCreateByDate(t *testing.T) {
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
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := NewDailyRecordService(repo)

			record, err := svc.GetOrCreateByDate(context.Background(), uint64(1), tt.hospitalizationID, tt.date)

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
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

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

func TestDailyRecordService_AddCareLog(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	recordTime := today.Add(12 * time.Hour)
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
				Time:    recordTime,
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
				Time:   recordTime,
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
				Time:   recordTime,
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
				Time:   recordTime,
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
				Time:   recordTime,
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
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createCareLogFn: func(_ context.Context, _ *model.CareLog) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

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
				Time:    today.Add(9 * time.Hour),
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
				Time:    today.Add(14*time.Hour + 30*time.Minute),
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
				Time:    today.Add(9 * time.Hour),
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
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createStaffNoteFn: func(_ context.Context, _ *model.StaffNote) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

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
