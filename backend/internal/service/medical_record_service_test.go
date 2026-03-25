package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// mockMedicalRecordRepository は MedicalRecordRepository のテスト用モック実装
type mockMedicalRecordRepository struct {
	findAllFn        func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	findByIDFn       func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	findByRecordNoFn func(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error)
	createFn         func(ctx context.Context, record *model.MedicalRecord) error
	updateFn         func(ctx context.Context, record *model.MedicalRecord) error
	deleteFn         func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockMedicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordRepository) FindByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error) {
	return m.findByRecordNoFn(ctx, clinicID, recordNo)
}

func (m *mockMedicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	return m.createFn(ctx, record)
}

func (m *mockMedicalRecordRepository) Update(ctx context.Context, record *model.MedicalRecord) error {
	return m.updateFn(ctx, record)
}

func (m *mockMedicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

// mrMockOwnerRepo は MedicalRecord テスト用 OwnerRepository モック（FindByID のみ）
type mrMockOwnerRepo struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

func (m *mrMockOwnerRepo) FindAll(_ context.Context, _ uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
	return nil, 0, nil
}
func (m *mrMockOwnerRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return m.findByIDFn(ctx, clinicID, id)
}
func (m *mrMockOwnerRepo) CreateWithPets(_ context.Context, _ *model.Owner, _ []model.Pet) error {
	return nil
}
func (m *mrMockOwnerRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mrMockOwnerRepo) Delete(_ context.Context, _, _ uint64) error { return nil }

// mrMockPetRepo は MedicalRecord テスト用 PetRepository モック（FindByID のみ）
type mrMockPetRepo struct {
	findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

func (m *mrMockPetRepo) FindAll(_ context.Context, _ uint64, _ *uint64, _, _ int, _ string) ([]model.Pet, int64, error) {
	return nil, 0, nil
}
func (m *mrMockPetRepo) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	return m.findByIDFn(ctx, clinicID, id)
}
func (m *mrMockPetRepo) CountByOwner(_ context.Context, _, _ uint64) (int64, error) { return 0, nil }
func (m *mrMockPetRepo) Create(_ context.Context, _ *model.Pet) error               { return nil }
func (m *mrMockPetRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
	return nil
}
func (m *mrMockPetRepo) Delete(_ context.Context, _, _ uint64) error { return nil }

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
				findAllFn: func(_ context.Context, _ uint64, petID *uint64, ownerID *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
					capturedPetID = petID
					capturedOwnerID = ownerID
					return tt.repoRecords, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil)

			records, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, tt.page, tt.limit)

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
			svc := NewMedicalRecordService(repo, nil, nil)

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
	svc := NewMedicalRecordService(repo, nil, nil)

	record, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, record)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestMedicalRecordService_GetByRecordNo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		clinicID   uint64
		recordNo   string
		repoRecord *model.MedicalRecord
		repoErr    error
		wantErr    bool
		wantNF     bool
	}{
		{
			name:       "returns record when found by record_no",
			clinicID:   1,
			recordNo:   "R001",
			repoRecord: &model.MedicalRecord{ID: 1, ClinicID: 1, RecordNo: "R001", Date: now},
			repoErr:    nil,
			wantErr:    false,
			wantNF:     false,
		},
		{
			name:       "returns not found error when record_no does not exist",
			clinicID:   1,
			recordNo:   "RXXX",
			repoRecord: nil,
			repoErr:    apperrors.WrapNotFound("medical_record", "RXXX"),
			wantErr:    true,
			wantNF:     true,
		},
		{
			name:       "returns error on repository failure",
			clinicID:   1,
			recordNo:   "R001",
			repoRecord: nil,
			repoErr:    errors.New("db error"),
			wantErr:    true,
			wantNF:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				findByRecordNoFn: func(_ context.Context, _ uint64, _ string) (*model.MedicalRecord, error) {
					return tt.repoRecord, tt.repoErr
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil)

			record, err := svc.GetByRecordNo(context.Background(), tt.clinicID, tt.recordNo)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoRecord, record)
			}
		})
	}
}

func TestMedicalRecordService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		record  *model.MedicalRecord
		repoErr error
		wantErr bool
	}{
		{
			name: "creates record successfully",
			record: &model.MedicalRecord{
				ClinicID: 1,
				RecordNo: "R100",
				Date:     now,
				Status:   model.MedicalRecordStatusDraft,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name: "returns already exists error on duplicate record_no",
			record: &model.MedicalRecord{
				ClinicID: 1,
				RecordNo: "R001",
				Date:     now,
			},
			repoErr: apperrors.WrapAlreadyExists("medical_record", "R001"),
			wantErr: true,
		},
		{
			name: "returns error on repository failure",
			record: &model.MedicalRecord{
				ClinicID: 1,
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
			svc := NewMedicalRecordService(repo, nil, nil)

			err := svc.Create(context.Background(), tt.record)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicalRecordService_Update(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		record  *model.MedicalRecord
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates record successfully",
			record: &model.MedicalRecord{
				ID:       1,
				ClinicID: 1,
				RecordNo: "R001",
				Date:     now,
				Status:   model.MedicalRecordStatusFinalized,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name: "returns not found error when record does not exist",
			record: &model.MedicalRecord{
				ID:       999,
				ClinicID: 1,
				RecordNo: "R999",
				Date:     now,
			},
			repoErr: apperrors.WrapNotFound("medical_record", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			record: &model.MedicalRecord{
				ID:       1,
				ClinicID: 1,
				RecordNo: "R001",
				Date:     now,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				updateFn: func(_ context.Context, _ *model.MedicalRecord) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil)

			err := svc.Update(context.Background(), tt.record)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicalRecordService_Update_OwnerValidation(t *testing.T) {
	now := time.Now()
	ownerID := uint64(100)

	t.Run("rejects owner_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			updateFn: func(_ context.Context, _ *model.MedicalRecord) error { return nil },
		}
		ownerRepo := &mrMockOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "100")
			},
		}
		svc := NewMedicalRecordService(repo, ownerRepo, nil)

		err := svc.Update(context.Background(), &model.MedicalRecord{
			ID: 1, ClinicID: 1, RecordNo: "R001", Date: now, OwnerID: &ownerID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("accepts valid owner_id in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			updateFn: func(_ context.Context, _ *model.MedicalRecord) error { return nil },
		}
		ownerRepo := &mrMockOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: ownerID, ClinicID: 1}, nil
			},
		}
		svc := NewMedicalRecordService(repo, ownerRepo, nil)

		err := svc.Update(context.Background(), &model.MedicalRecord{
			ID: 1, ClinicID: 1, RecordNo: "R001", Date: now, OwnerID: &ownerID,
		})

		assert.NoError(t, err)
	})
}

func TestMedicalRecordService_Update_PetValidation(t *testing.T) {
	now := time.Now()
	petID := uint64(200)

	t.Run("rejects pet_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			updateFn: func(_ context.Context, _ *model.MedicalRecord) error { return nil },
		}
		petRepo := &mrMockPetRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, apperrors.WrapNotFound("pet", "200")
			},
		}
		svc := NewMedicalRecordService(repo, nil, petRepo)

		err := svc.Update(context.Background(), &model.MedicalRecord{
			ID: 1, ClinicID: 1, RecordNo: "R001", Date: now, PetID: &petID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("accepts valid pet_id in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			updateFn: func(_ context.Context, _ *model.MedicalRecord) error { return nil },
		}
		petRepo := &mrMockPetRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: petID, ClinicID: 1}, nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, petRepo)

		err := svc.Update(context.Background(), &model.MedicalRecord{
			ID: 1, ClinicID: 1, RecordNo: "R001", Date: now, PetID: &petID,
		})

		assert.NoError(t, err)
	})
}

func TestMedicalRecordService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes record successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when record does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("medical_record", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoErr:  errors.New("db error"),
			wantErr:  true,
			wantNF:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil)

			err := svc.Delete(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
