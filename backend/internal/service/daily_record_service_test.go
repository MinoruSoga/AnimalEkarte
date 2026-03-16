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
	listByHospitalizationIDFn       func(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error)
	getOrCreateByDateFn             func(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	findByHospitalizationIDAndDateFn func(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	createVitalRecordFn             func(ctx context.Context, vr *model.VitalRecord) error
	createCareLogRecordFn           func(ctx context.Context, cr *model.CareLogRecord) error
	createStaffNoteRecordFn         func(ctx context.Context, sn *model.StaffNoteRecord) error
}

func (m *mockDailyRecordRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error) {
	return m.listByHospitalizationIDFn(ctx, hospitalizationID)
}

func (m *mockDailyRecordRepository) GetOrCreateByDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.getOrCreateByDateFn(ctx, hospitalizationID, date)
}

func (m *mockDailyRecordRepository) FindByHospitalizationIDAndDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	return m.findByHospitalizationIDAndDateFn(ctx, hospitalizationID, date)
}

func (m *mockDailyRecordRepository) CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error {
	return m.createVitalRecordFn(ctx, vr)
}

func (m *mockDailyRecordRepository) CreateCareLogRecord(ctx context.Context, cr *model.CareLogRecord) error {
	return m.createCareLogRecordFn(ctx, cr)
}

func (m *mockDailyRecordRepository) CreateStaffNoteRecord(ctx context.Context, sn *model.StaffNoteRecord) error {
	return m.createStaffNoteRecordFn(ctx, sn)
}

// ---- Tests ----

func TestDailyRecordService_List(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	tests := []struct {
		name                string
		hospitalizationID   uint64
		repoRecords         []model.DailyRecord
		repoErr             error
		wantLen             int
		wantErr             bool
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
				listByHospitalizationIDFn: func(_ context.Context, _ uint64) ([]model.DailyRecord, error) {
					return tt.repoRecords, tt.repoErr
				},
			}
			svc := NewDailyRecordService(repo)

			records, err := svc.List(context.Background(), tt.hospitalizationID)

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
				getOrCreateByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
			}
			svc := NewDailyRecordService(repo)

			record, err := svc.GetOrCreateByDate(context.Background(), tt.hospitalizationID, tt.date)

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
				getOrCreateByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createVitalRecordFn: func(_ context.Context, _ *model.VitalRecord) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

			record, err := svc.AddVitalRecord(context.Background(), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestDailyRecordService_AddCareLogRecord(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	recordTime := today.Add(12 * time.Hour)
	staffID := uint64(10)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		input             *CreateCareLogRecordInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "adds care log record with food type",
			hospitalizationID: 1,
			date:              today,
			input: &CreateCareLogRecordInput{
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
			input: &CreateCareLogRecordInput{
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
			input: &CreateCareLogRecordInput{
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
			input: &CreateCareLogRecordInput{
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
			input: &CreateCareLogRecordInput{
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
				getOrCreateByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createCareLogRecordFn: func(_ context.Context, _ *model.CareLogRecord) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

			record, err := svc.AddCareLogRecord(context.Background(), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestDailyRecordService_AddStaffNoteRecord(t *testing.T) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	staffID := uint64(10)

	tests := []struct {
		name              string
		hospitalizationID uint64
		date              time.Time
		input             *CreateStaffNoteRecordInput
		repoErr           error
		wantErr           bool
	}{
		{
			name:              "adds staff note record successfully",
			hospitalizationID: 1,
			date:              today,
			input: &CreateStaffNoteRecordInput{
				Time:    "09:00",
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
			input: &CreateStaffNoteRecordInput{
				Time:    "14:30",
				Content: "Brief note",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:              "returns error when get_or_create fails",
			hospitalizationID: 1,
			date:              today,
			input: &CreateStaffNoteRecordInput{
				Time:    "09:00",
				Content: "Test note",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockDailyRecordRepository{
				getOrCreateByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID, Date: tt.date}, nil
				},
				createStaffNoteRecordFn: func(_ context.Context, _ *model.StaffNoteRecord) error {
					return nil
				},
				findByHospitalizationIDAndDateFn: func(_ context.Context, _ uint64, _ time.Time) (*model.DailyRecord, error) {
					return &model.DailyRecord{ID: 1, HospitalizationID: tt.hospitalizationID}, nil
				},
			}
			svc := NewDailyRecordService(repo)

			record, err := svc.AddStaffNoteRecord(context.Background(), tt.hospitalizationID, tt.date, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}
