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
	findAllFn                         func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error)
	findByIDFn                        func(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	createFn                          func(ctx context.Context, record *model.MedicalRecord) error
	updateFieldsFn                    func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error)
	deleteFn                          func(ctx context.Context, clinicID, id uint64) error
	countByPetIDFn                    func(ctx context.Context, clinicID, petID uint64) (int64, error)
	countEstimatesByMedicalRecordIDFn func(ctx context.Context, medicalRecordID uint64) (int64, error)
}

func (m *mockMedicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.MedicalRecord, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockMedicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	return m.createFn(ctx, record)
}

func (m *mockMedicalRecordRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicalRecord, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.MedicalRecord{}, nil
}

func (m *mockMedicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockMedicalRecordRepository) CountByPetID(ctx context.Context, clinicID, petID uint64) (int64, error) {
	if m.countByPetIDFn != nil {
		return m.countByPetIDFn(ctx, clinicID, petID)
	}
	return 0, nil
}

func (m *mockMedicalRecordRepository) CountEstimatesByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (int64, error) {
	if m.countEstimatesByMedicalRecordIDFn != nil {
		return m.countEstimatesByMedicalRecordIDFn(ctx, medicalRecordID)
	}
	return 0, nil
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
func (m *mrMockOwnerRepo) FindByEmail(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mrMockOwnerRepo) FindByPhone(_ context.Context, _ uint64, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mrMockOwnerRepo) FindByNameAndPhone(_ context.Context, _ uint64, _, _ string) (*model.Owner, error) {
	return nil, nil
}
func (m *mrMockOwnerRepo) CountPetsByOwnerID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

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
func (m *mrMockPetRepo) CountUsageByAnimalSpeciesID(_ context.Context, _ uint64) (int64, error) {
	return 0, nil
}
func (m *mrMockPetRepo) Create(_ context.Context, _ *model.Pet) error { return nil }
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
			svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

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
			svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

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
	svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

	record, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, record)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
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
			svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

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
	statusFinalized := model.MedicalRecordStatusFinalized
	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0}
	tests := []struct {
		name        string
		input       UpdateMedicalRecordInput
		findByIDErr error // FindByID のエラー（nil = 正常レコード返却）
		updateErr   error // Update のエラー
		wantErr     bool
		wantNF      bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return validRecord, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.MedicalRecord, error) {
					if tt.updateErr != nil {
						return nil, tt.updateErr
					}
					return &model.MedicalRecord{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

			record, err := svc.Update(context.Background(), 1, 1, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, record)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, record)
			}
		})
	}
}

func TestMedicalRecordService_Update_OwnerValidation(t *testing.T) {
	ownerID := uint64(100)
	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0}

	t.Run("rejects owner_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
		}
		ownerRepo := &mrMockOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return nil, apperrors.WrapNotFound("owner", "100")
			},
		}
		svc := NewMedicalRecordService(repo, ownerRepo, nil, nil, nil, nil)

		_, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			OwnerID: &ownerID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
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
		ownerRepo := &mrMockOwnerRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
				return &model.Owner{ID: ownerID, ClinicID: 1}, nil
			},
		}
		svc := NewMedicalRecordService(repo, ownerRepo, nil, nil, nil, nil)

		record, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			OwnerID: &ownerID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, record)
	})
}

func TestMedicalRecordService_Update_PetValidation(t *testing.T) {
	petID := uint64(200)

	validRecord := &model.MedicalRecord{ID: 1, ClinicID: 1, Version: 0}

	t.Run("rejects pet_id not in clinic", func(t *testing.T) {
		repo := &mockMedicalRecordRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
				return validRecord, nil
			},
		}
		petRepo := &mrMockPetRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, apperrors.WrapNotFound("pet", "200")
			},
		}
		svc := NewMedicalRecordService(repo, nil, petRepo, nil, nil, nil)

		_, err := svc.Update(context.Background(), 1, 1, UpdateMedicalRecordInput{
			PetID: &petID,
		})

		assert.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
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
		petRepo := &mrMockPetRepo{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return &model.Pet{ID: petID, ClinicID: 1}, nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, petRepo, nil, nil, nil)

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
				countEstimatesByMedicalRecordIDFn: func(_ context.Context, _ uint64) (int64, error) {
					return tt.estimateCount, tt.estimateErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)

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

// noopInquiryRepo は CreateSubRecords テスト用の no-op InquiryRepository
type noopInquiryRepo struct{}

func (n *noopInquiryRepo) UpsertByMedicalRecordID(_ context.Context, _ uint64, inquiry *model.Inquiry) (*model.Inquiry, error) {
	return inquiry, nil
}
func (n *noopInquiryRepo) CountByChiefComplaintTypeID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

// noopClinicalPlanRepo は CreateSubRecords テスト用の no-op ClinicalPlanRepository
type noopClinicalPlanRepo struct{}

func (n *noopClinicalPlanRepo) FindByMedicalRecordID(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
	return nil, apperrors.WrapNotFound("clinical_plan", "0")
}
func (n *noopClinicalPlanRepo) Create(_ context.Context, _ *model.ClinicalPlan) error { return nil }
func (n *noopClinicalPlanRepo) Update(_ context.Context, _, _ uint64, _ map[string]any) error {
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
			findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
				return nil, 0, nil
			},
		}
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, nil)
		appt := &model.Reservation{
			ID:        1,
			ClinicID:  1,
			StartTime: now,
		}
		// エラーなく静かにスキップするだけ（パニックしない）
		svc.AutoCreateFromReservation(context.Background(), 1, appt)
	})

	t.Run("resolves owner_id from line_customer and creates medical record (BUG-386)", func(t *testing.T) {
		var createdRecord *model.MedicalRecord
		repo := &mockMedicalRecordRepository{
			findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
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
		svc := NewMedicalRecordService(repo, nil, nil, &noopInquiryRepo{}, &noopClinicalPlanRepo{}, lineCustomerRepo)

		appt := &model.Reservation{
			ID:             1,
			ClinicID:       1,
			StartTime:      now,
			LineCustomerID: &lineCustomerID,
			// owner_id / pet_id は未設定（LINE予約で自動紐付けが未完了の状態）
		}

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
			findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.MedicalRecord, int64, error) {
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
		svc := NewMedicalRecordService(repo, nil, nil, nil, nil, lineCustomerRepo)

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
