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
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// mockHospitalizationRepository は HospitalizationRepository のテスト用モック実装
type mockHospitalizationRepository struct {
	findAllFn                                func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error)
	findByIDFn                               func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn                                 func(ctx context.Context, hospitalization *model.Hospitalization) error
	updateFieldsFn                           func(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error)
	createStayCaseFn                         func(ctx context.Context, stay *model.StayCase) error
	findStayCasesByIDsFn                     func(ctx context.Context, clinicID uint64, ids []uint64) (map[uint64]model.StayCase, error)
	deleteFn                                 func(ctx context.Context, clinicID, id uint64) error
	countCarePlanItemsByHospitalizationIDFn  func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countDailyRecordsByHospitalizationIDFn   func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countTreatmentPlansByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

func (m *mockHospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, plannedEntryFrom, plannedEntryTo *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, plannedEntryFrom, plannedEntryTo, page, limit)
}

func (m *mockHospitalizationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

// LockByIDForUpdate は Discharge Q2-C 用。既存テストの findByIDFn フックをそのまま使えるよう FindByID に委譲する。
func (m *mockHospitalizationRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
	return m.FindByID(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) Create(ctx context.Context, hospitalization *model.Hospitalization) error {
	return m.createFn(ctx, hospitalization)
}

func (m *mockHospitalizationRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
	return m.updateFieldsFn(ctx, clinicID, id, cmd)
}

func (m *mockHospitalizationRepository) CreateStayCase(ctx context.Context, stay *model.StayCase) error {
	if m.createStayCaseFn != nil {
		return m.createStayCaseFn(ctx, stay)
	}
	stay.ID = 900 + stay.OwnerID
	return nil
}

func (m *mockHospitalizationRepository) FindStayCasesByIDs(ctx context.Context, clinicID uint64, ids []uint64) (map[uint64]model.StayCase, error) {
	if m.findStayCasesByIDsFn != nil {
		return m.findStayCasesByIDsFn(ctx, clinicID, ids)
	}
	return map[uint64]model.StayCase{}, nil
}

func (m *mockHospitalizationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockHospitalizationRepository) CountByCageID(_ context.Context, _, _ uint64) (int64, error) {
	return 0, nil
}

func (m *mockHospitalizationRepository) CountCarePlanItemsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countCarePlanItemsByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countCarePlanItemsByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockHospitalizationRepository) CountDailyRecordsByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countDailyRecordsByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countDailyRecordsByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func (m *mockHospitalizationRepository) CountTreatmentPlansByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error) {
	if m.countTreatmentPlansByHospitalizationIDFn == nil {
		return 0, nil
	}
	return m.countTreatmentPlansByHospitalizationIDFn(ctx, clinicID, hospitalizationID)
}

func TestHospitalizationService_List(t *testing.T) {
	petID := uint64(5)
	ownerID := uint64(2)
	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		page      int
		limit     int
		repoItems []model.Hospitalization
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns hospitalization list with total count",
			clinicID: 1,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2},
				{ID: 2, ClinicID: 1, PetID: 6, OwnerID: 3},
			},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no hospitalizations exist",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: []model.Hospitalization{},
			repoTotal: 0,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by pet_id",
			clinicID: 1,
			petID:    &petID,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:     "filters by owner_id",
			clinicID: 1,
			ownerID:  &ownerID,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2},
			},
			repoTotal: 1,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			page:      1,
			limit:     20,
			repoItems: nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _ *string, _, _ int) ([]model.Hospitalization, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}
		})
	}
}

func TestHospitalizationService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoItem *model.Hospitalization
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "returns hospitalization when found",
			clinicID: 1,
			id:       10,
			repoItem: &model.Hospitalization{
				ID:       10,
				ClinicID: 1,
				PetID:    5,
				OwnerID:  2,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns not found error when hospitalization does not exist",
			clinicID: 1,
			id:       999,
			repoItem: nil,
			repoErr:  apperrors.WrapNotFound("hospitalization", "999"),
			wantErr:  true,
			wantNF:   true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoItem: nil,
			repoErr:  errors.New("db error"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
					return tt.repoItem, tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			item, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoItem, item)
			}
		})
	}
}

func TestHospitalizationService_Create(t *testing.T) {
	now := time.Now()
	entry, exit := now, now.Add(48*time.Hour)
	cageTen := func() *uint64 { v := uint64(10); return &v }
	tests := []struct {
		name        string
		input       *CreateHospitalizationInput
		stayErr     error
		repoErr     error
		wantErr     bool
		wantInvalid bool
		wantCageErr string
	}{
		{
			name: "creates stay case and hospitalization without cage",
			input: &CreateHospitalizationInput{
				OwnerID: 2, PetID: 5, StayKind: stayKindInpatient, PlannedEntryAt: &entry, PlannedExitAt: &exit,
			},
		},
		{
			name: "hotel stay with no planned interval is stored as given (nothing invented)",
			input: &CreateHospitalizationInput{
				OwnerID: 2, PetID: 5, StayKind: stayKindHotel,
			},
		},
		{
			// #734 guard kept: a non-NULL cage needs the acceptance proof that no writer produces yet.
			name: "non-NULL cage is refused as occupancy_unknown and nothing is persisted",
			input: &CreateHospitalizationInput{
				OwnerID: 2, PetID: 5, StayKind: stayKindInpatient, CageID: cageTen(),
			},
			wantErr: true, wantCageErr: codeOccupancyUnknown,
		},
		{
			name: "rejects cage_id 0",
			input: &CreateHospitalizationInput{
				OwnerID: 2, PetID: 5, StayKind: stayKindInpatient, CageID: func() *uint64 { v := uint64(0); return &v }(),
			},
			wantErr: true, wantInvalid: true,
		},
		{
			name:    "rejects unknown stay_kind",
			input:   &CreateHospitalizationInput{OwnerID: 2, PetID: 5, StayKind: "boarding"},
			wantErr: true, wantInvalid: true,
		},
		{
			name: "rejects planned exit not after entry",
			input: &CreateHospitalizationInput{
				OwnerID: 2, PetID: 5, StayKind: stayKindHotel, PlannedEntryAt: &exit, PlannedExitAt: &entry,
			},
			wantErr: true, wantInvalid: true,
		},
		{
			name:    "returns error when stay case insert fails and does not insert the hospitalization",
			input:   &CreateHospitalizationInput{OwnerID: 2, PetID: 5, StayKind: stayKindInpatient},
			stayErr: errors.New("db error"), wantErr: true,
		},
		{
			name:    "returns error on hospitalization insert failure",
			input:   &CreateHospitalizationInput{OwnerID: 2, PetID: 5, StayKind: stayKindInpatient},
			repoErr: errors.New("db error"), wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *model.Hospitalization
			var stay *model.StayCase
			createCalls := 0
			repo := &mockHospitalizationRepository{
				createStayCaseFn: func(_ context.Context, sc *model.StayCase) error {
					if tt.stayErr != nil {
						return tt.stayErr
					}
					sc.ID = 77
					stay = sc
					return nil
				},
				createFn: func(_ context.Context, h *model.Hospitalization) error {
					createCalls++
					created = h
					return tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo, &mockReservationRepository{
				assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return tt.input.OwnerID, nil
				},
			}, &mockPetRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
					return &model.Pet{ID: id}, nil
				},
			}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			hosp, err := svc.Create(context.Background(), 1, tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, hosp)
				if tt.repoErr == nil {
					assert.Equal(t, 0, createCalls, "no hospitalization row on a refused create")
				}
				if tt.wantInvalid {
					assert.True(t, apperrors.IsInvalidInput(err))
				}
				if tt.wantCageErr != "" {
					var appErr *apperrors.AppError
					require.True(t, errors.As(err, &appErr))
					assert.Equal(t, tt.wantCageErr, appErr.Code)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, hosp)
			require.NotNil(t, created)
			require.NotNil(t, stay)
			assert.Equal(t, tt.input.StayKind, stay.StayKind)
			assert.Equal(t, tt.input.PlannedEntryAt, stay.PlannedEntryAt)
			assert.Equal(t, tt.input.PlannedExitAt, stay.PlannedExitAt)
			assert.Equal(t, uint64(2), stay.OwnerID)
			assert.Equal(t, uint64(5), stay.PetID)
			require.NotNil(t, created.StayCaseID)
			assert.Equal(t, uint64(77), *created.StayCaseID)
			assert.Nil(t, created.CageID, "create records no cage assignment")
		})
	}
}

func TestHospitalizationService_Update(t *testing.T) {
	memo := "memo"
	tests := []struct {
		name    string
		input   UpdateHospitalizationInput
		repoErr error
		wantErr bool
	}{
		{
			name:    "updates hospitalization successfully",
			input:   UpdateHospitalizationInput{Memo: &memo},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateHospitalizationInput{},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:    "returns error on repository failure",
			input:   UpdateHospitalizationInput{Memo: &memo},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID:       id,
						ClinicID: clinicID,
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Hospitalization{ID: 1, ClinicID: 1}, nil
				},
			}
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			hosp, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hosp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hosp)
			}
		})
	}
}

func TestHospitalizationService_Update_InputNil(t *testing.T) {
	repo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			t.Fatal("hospitalization must not be looked up when input is nil")
			return nil, nil
		},
	}
	svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Update(context.Background(), 1, 1, nil)

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, hosp)
}

func TestHospitalizationService_Update_FindByIDError(t *testing.T) {
	memo := "memo"
	repo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "999")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			t.Fatal("hospitalization must not be updated when the parent lookup fails")
			return nil, nil
		},
	}
	svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Update(context.Background(), 1, 999, &UpdateHospitalizationInput{Memo: &memo})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, hosp)
}

func TestHospitalizationService_Delete(t *testing.T) {
	tests := []struct {
		name                 string
		clinicID             uint64
		id                   uint64
		dailyRecordCount     int64
		countDailyRecordErr  error
		treatmentPlanCount   int64
		countTreatmentErr    error
		carePlanItemCount    int64
		countCarePlanItemErr error
		repoErr              error
		wantErr              bool
		wantNF               bool
		wantConflict         bool
	}{
		{
			name:                 "deletes hospitalization successfully when no care plan items exist",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              nil,
			wantErr:              false,
		},
		{
			name:                 "returns conflict error when hospitalization has care plan items",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    5,
			countCarePlanItemErr: nil,
			repoErr:              nil,
			wantErr:              true,
			wantConflict:         true,
		},
		{
			name:                 "returns error when care plan item count check fails",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: errors.New("db error"),
			repoErr:              nil,
			wantErr:              true,
		},
		{
			name:                 "returns not found error when hospitalization does not exist",
			clinicID:             1,
			id:                   999,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              apperrors.WrapNotFound("hospitalization", "999"),
			wantErr:              true,
			wantNF:               true,
		},
		{
			name:                 "returns error on repository failure",
			clinicID:             1,
			id:                   10,
			carePlanItemCount:    0,
			countCarePlanItemErr: nil,
			repoErr:              errors.New("db error"),
			wantErr:              true,
		},
		{
			name:             "returns conflict error when hospitalization has daily records",
			clinicID:         1,
			id:               10,
			dailyRecordCount: 2,
			wantErr:          true,
			wantConflict:     true,
		},
		{
			name:                "returns error when daily record count check fails",
			clinicID:            1,
			id:                  10,
			countDailyRecordErr: errors.New("db error"),
			wantErr:             true,
		},
		{
			name:               "returns conflict error when hospitalization has treatment plans",
			clinicID:           1,
			id:                 10,
			treatmentPlanCount: 1,
			wantErr:            true,
			wantConflict:       true,
		},
		{
			name:              "returns error when treatment plan count check fails",
			clinicID:          1,
			id:                10,
			countTreatmentErr: errors.New("db error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
					// "not found" case is modeled as delete returning NotFound after a successful pre-check.
					return &model.Hospitalization{ID: id, ClinicID: clinicID}, nil
				},
				countDailyRecordsByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.dailyRecordCount, tt.countDailyRecordErr
				},
				countTreatmentPlansByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.treatmentPlanCount, tt.countTreatmentErr
				},
				countCarePlanItemsByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					return tt.carePlanItemCount, tt.countCarePlanItemErr
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			// MRB-05: Delete requires auditTx (fail-closed).
			svc := NewHospitalizationServiceWithAudit(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, okCarePlanAuditTx{}, WithPetLifeStateReader(stubPetLifeReader{}))

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

func TestHospitalizationService_Delete_FindByIDError(t *testing.T) {
	repo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "999")
		},
		countDailyRecordsByHospitalizationIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
			t.Fatal("dependency checks must not run when the parent lookup fails")
			return 0, nil
		},
	}
	svc := NewHospitalizationServiceWithAudit(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, okCarePlanAuditTx{}, WithPetLifeStateReader(stubPetLifeReader{}))

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestHospitalizationService_Create_InsuranceFields(t *testing.T) {
	companyName := "アニマル保険"
	insuranceNo := "INS-001"

	tests := []struct {
		name                 string
		input                *CreateHospitalizationInput
		wantInsuranceCompany *string
		wantInsuranceNumber  *string
	}{
		{
			name: "is_insurance=true の場合は保険フィールドを保存する",
			input: &CreateHospitalizationInput{
				CageID:               func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:              2,
				PetID:                5,
				StayKind:             stayKindInpatient,
				IsInsurance:          true,
				InsuranceCompanyName: &companyName,
				InsuranceNumber:      &insuranceNo,
			},
			wantInsuranceCompany: &companyName,
			wantInsuranceNumber:  &insuranceNo,
		},
		{
			name: "is_insurance=false の場合は保険フィールドを NULL にする",
			input: &CreateHospitalizationInput{
				CageID:               func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:              2,
				PetID:                5,
				StayKind:             stayKindInpatient,
				IsInsurance:          false,
				InsuranceCompanyName: &companyName,
				InsuranceNumber:      &insuranceNo,
			},
			wantInsuranceCompany: nil,
			wantInsuranceNumber:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *model.Hospitalization
			repo := &mockHospitalizationRepository{
				createFn: func(_ context.Context, h *model.Hospitalization) error {
					captured = h
					return nil
				},
			}
			svc := NewHospitalizationService(repo, &mockReservationRepository{
				assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return tt.input.OwnerID, nil
				},
			}, &mockPetRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
					return &model.Pet{ID: id}, nil
				},
			}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			hosp, err := svc.Create(context.Background(), 1, tt.input)

			assert.NoError(t, err)
			assert.NotNil(t, hosp)
			assert.Equal(t, tt.wantInsuranceCompany, captured.InsuranceCompanyName)
			assert.Equal(t, tt.wantInsuranceNumber, captured.InsuranceNumber)
		})
	}
}

// EMR-491 D6: deceased は業務拒否（InvalidInput）、unresolved / reader 失敗は生存扱いにせずエラーで拒否する。
func TestHospitalizationService_Create_RejectsDeceasedOrUnresolvedPet(t *testing.T) {
	for _, tc := range []struct {
		name        string
		state       sharedkernel.PetLifeState
		readErr     error
		wantInvalid bool
	}{
		{name: "deceased", state: sharedkernel.PetLifeStateDeceased, wantInvalid: true},
		{name: "unresolved", readErr: apperrors.Wrap(sharedkernel.ErrPetLifeStateUnresolved, "pet status is not recorded")},
		{name: "reader failure", readErr: errors.New("db down")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				createFn: func(_ context.Context, _ *model.Hospitalization) error {
					t.Fatal("hospitalization must not be created unless the pet is confirmed alive")
					return nil
				},
			}
			svc := NewHospitalizationService(repo, &mockReservationRepository{
				assertOwnerInClinicFn:  func(_ context.Context, _, _ uint64) error { return nil },
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil },
			}, &mockPetRepository{}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{},
				WithPetLifeStateReader(stubPetLifeState(tc.state, tc.readErr)))

			hosp, err := svc.Create(context.Background(), 1, &CreateHospitalizationInput{
				CageID:   func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:  2,
				PetID:    5,
				StayKind: stayKindInpatient,
			})

			assert.Error(t, err)
			assert.Equal(t, tc.wantInvalid, apperrors.IsInvalidInput(err), "unexpected error class: %v", err)
			if tc.readErr != nil {
				assert.ErrorIs(t, err, tc.readErr)
			}
			assert.Nil(t, hosp)
		})
	}
}

// 入院 write は reader 未配線のとき生存扱いにせず内部エラーで拒否する（nil は alive ではない）。
func TestHospitalizationService_Create_FailsClosedWithoutPetLifeStateReader(t *testing.T) {
	repo := &mockHospitalizationRepository{
		createFn: func(_ context.Context, _ *model.Hospitalization) error {
			t.Fatal("hospitalization must not be created without a pet life state reader")
			return nil
		},
	}
	svc := NewHospitalizationService(repo, &mockReservationRepository{
		assertOwnerInClinicFn:  func(_ context.Context, _, _ uint64) error { return nil },
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil },
	}, &mockPetRepository{}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Create(context.Background(), 1, &CreateHospitalizationInput{
		CageID:   func() *uint64 { v := uint64(10); return &v }(),
		OwnerID:  2,
		PetID:    5,
		StayKind: stayKindInpatient,
	})

	assert.Error(t, err)
	assert.False(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, hosp)
}

func TestHospitalizationService_Update_RejectsDeceasedPetReplacement(t *testing.T) {
	newPetID := uint64(9)
	repo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, ClinicID: 1, OwnerID: 2, PetID: 5}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			t.Fatal("hospitalization must not be updated when the replacement pet is deceased")
			return nil, nil
		},
	}
	svc := NewHospitalizationService(repo, &mockReservationRepository{
		assertOwnerInClinicFn:  func(_ context.Context, _, _ uint64) error { return nil },
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil },
	}, &mockPetRepository{}, nil, nil, nil, nil, &mockTransactor{},
		WithPetLifeStateReader(stubPetLifeState(sharedkernel.PetLifeStateDeceased, nil)))

	hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{PetID: &newPetID})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, hosp)
}

func TestBuildHospitalizationUpdate_InsuranceFields(t *testing.T) {
	companyName := "アニマル保険"
	insuranceNo := "INS-001"
	isInsuranceTrue := true
	isInsuranceFalse := false

	tests := []struct {
		name                       string
		input                      UpdateHospitalizationInput
		wantInsuranceCompanyExists bool
		wantInsuranceCompanyValue  any
		wantInsuranceNumberExists  bool
		wantInsuranceNumberValue   any
	}{
		{
			name: "is_insurance=true の場合は保険フィールドをセットする",
			input: UpdateHospitalizationInput{
				IsInsurance:          &isInsuranceTrue,
				InsuranceCompanyName: &companyName,
				InsuranceNumber:      &insuranceNo,
			},
			wantInsuranceCompanyExists: true,
			wantInsuranceCompanyValue:  companyName,
			wantInsuranceNumberExists:  true,
			wantInsuranceNumberValue:   insuranceNo,
		},
		{
			name: "is_insurance=false の場合は保険フィールドを nil にする",
			input: UpdateHospitalizationInput{
				IsInsurance:          &isInsuranceFalse,
				InsuranceCompanyName: &companyName,
				InsuranceNumber:      &insuranceNo,
			},
			wantInsuranceCompanyExists: true,
			wantInsuranceCompanyValue:  nil,
			wantInsuranceNumberExists:  true,
			wantInsuranceNumberValue:   nil,
		},
		{
			name: "is_insurance が nil の場合でも保険フィールド単体を更新できる",
			input: UpdateHospitalizationInput{
				InsuranceCompanyName: &companyName,
			},
			wantInsuranceCompanyExists: true,
			wantInsuranceCompanyValue:  companyName,
			wantInsuranceNumberExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := buildHospitalizationUpdate(&tt.input)

			if tt.wantInsuranceCompanyExists {
				val, ok := fields["insurance_company_name"]
				assert.True(t, ok, "insurance_company_name should be in fields map")
				assert.Equal(t, tt.wantInsuranceCompanyValue, val)
			} else {
				_, ok := fields["insurance_company_name"]
				assert.False(t, ok, "insurance_company_name should NOT be in fields map")
			}

			if tt.wantInsuranceNumberExists {
				val, ok := fields["insurance_number"]
				assert.True(t, ok, "insurance_number should be in fields map")
				assert.Equal(t, tt.wantInsuranceNumberValue, val)
			} else {
				_, ok := fields["insurance_number"]
				assert.False(t, ok, "insurance_number should NOT be in fields map")
			}
		})
	}
}

func TestBuildHospitalizationUpdate_BasicFields(t *testing.T) {
	ownerID := uint64(2)
	petID := uint64(5)
	cageID := uint64(3)
	doctorID := uint64(7)
	memo := "メモ"
	ownerRequest := "希望事項"
	staffNotes := "スタッフメモ"

	fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{
		OwnerID:      &ownerID,
		PetID:        &petID,
		CageID:       &cageID,
		DoctorID:     &doctorID,
		Memo:         &memo,
		OwnerRequest: &ownerRequest,
		StaffNotes:   &staffNotes,
	})

	assert.Equal(t, ownerID, fields["owner_id"])
	assert.Equal(t, petID, fields["pet_id"])
	for _, retired := range []string{"hospitalization_type", "start_date", "end_date", "status"} {
		_, has := fields[retired]
		assert.False(t, has, "%s left the hospitalizations row (EMR-499 U1)", retired)
	}
	assert.Equal(t, cageID, fields["cage_id"])
	assert.Equal(t, doctorID, fields["doctor_id"])
	assert.Equal(t, memo, fields["memo"])
	assert.Equal(t, ownerRequest, fields["owner_request"])
	assert.Equal(t, staffNotes, fields["staff_notes"])
}

func TestBuildHospitalizationUpdate_EmptyInput(t *testing.T) {
	fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{})
	assert.Empty(t, fields)
}

// EMR-316: 「選択なし」への更新は clear_cage_id で cage_id を NULL にする（省略は変更なし）。
func TestBuildHospitalizationUpdate_ClearCageID(t *testing.T) {
	fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{ClearCageID: true})

	val, ok := fields["cage_id"]
	assert.True(t, ok, "clear_cage_id must put cage_id into the update set")
	assert.Nil(t, val, "clear_cage_id must set cage_id to NULL")
}

func TestBuildHospitalizationUpdate_SetCageIDAndOmitKeepsCage(t *testing.T) {
	cageID := uint64(3)

	set := buildHospitalizationUpdate(&UpdateHospitalizationInput{CageID: &cageID})
	assert.Equal(t, cageID, set["cage_id"])

	omitted := buildHospitalizationUpdate(&UpdateHospitalizationInput{})
	_, has := omitted["cage_id"]
	assert.False(t, has, "omitting cage_id must leave the column untouched")
}

func TestHospitalizationService_Update_RejectsClearAndSetCageID(t *testing.T) {
	cageID := uint64(3)
	svc := NewHospitalizationService(&mockHospitalizationRepository{}, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{ClearCageID: true, CageID: &cageID})

	assert.Nil(t, hosp)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestHospitalizationService_Update_ClearCageRequiresExpectedCageID(t *testing.T) {
	svc := NewHospitalizationService(&mockHospitalizationRepository{}, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{ClearCageID: true})
	assert.Nil(t, hosp)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))

	zero := uint64(0)
	hosp, err = svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{ClearCageID: true, ExpectedCageID: &zero})
	assert.Nil(t, hosp)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestHospitalizationService_Update_ClearCageStaleExpectationConflicts(t *testing.T) {
	expected := uint64(10)
	tests := []struct {
		name        string
		lockedCage  *uint64
		wantUpdated bool
	}{
		{name: "locked cage reassigned since the form was displayed", lockedCage: ptrUint64(11)},
		{name: "locked cage already cleared", lockedCage: nil},
		{name: "locked cage matches the displayed cage", lockedCage: ptrUint64(10), wantUpdated: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := &model.Hospitalization{
				ID: 1, ClinicID: 1,
				CageID: tt.lockedCage,
			}
			updated := false
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) { return existing, nil },
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
					updated = true
					result := *existing
					result.CageID = nil
					return &result, nil
				},
			}
			svc := NewHospitalizationService(hospRepo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			got, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{ClearCageID: true, ExpectedCageID: &expected})

			if tt.wantUpdated {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, updated)
				assert.Nil(t, got.CageID)
			} else {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.True(t, apperrors.IsConflict(err), "stale expected_cage_id must be a conflict: %v", err)
				assert.False(t, updated, "conflict must not write")
			}
		})
	}
}

func TestHospitalizationService_Update_CageCompareAndSet(t *testing.T) {
	cage10, cage11, cage12 := uint64(10), uint64(11), uint64(12)
	memo := "memo"
	tests := []struct {
		name             string
		lockedCage       *uint64
		input            *UpdateHospitalizationInput
		wantUpdated      bool
		wantConflict     bool
		wantInvalidInput bool
	}{
		{name: "null to cage12 succeeds with expected_null", lockedCage: nil,
			input: &UpdateHospitalizationInput{CageID: &cage12, ExpectedCageIsNull: true}, wantUpdated: true},
		{name: "cage10 to cage12 succeeds with matching expected", lockedCage: &cage10,
			input: &UpdateHospitalizationInput{CageID: &cage12, ExpectedCageID: &cage10}, wantUpdated: true},
		{name: "stale expected cage10 vs locked cage11 rejects cage12", lockedCage: &cage11,
			input: &UpdateHospitalizationInput{CageID: &cage12, ExpectedCageID: &cage10}, wantConflict: true},
		{name: "expected_null against locked cage10 rejects cage12", lockedCage: &cage10,
			input: &UpdateHospitalizationInput{CageID: &cage12, ExpectedCageIsNull: true}, wantConflict: true},
		{name: "idempotent null clear succeeds", lockedCage: nil,
			input: &UpdateHospitalizationInput{ClearCageID: true, ExpectedCageIsNull: true}, wantUpdated: true},
		{name: "stale clear expected cage10 vs locked nil rejects", lockedCage: nil,
			input: &UpdateHospitalizationInput{ClearCageID: true, ExpectedCageID: &cage10}, wantConflict: true},
		{name: "cage change without expected is invalid", lockedCage: nil,
			input: &UpdateHospitalizationInput{CageID: &cage12}, wantInvalidInput: true},
		{name: "expected without cage operation is invalid", lockedCage: &cage10,
			input: &UpdateHospitalizationInput{ExpectedCageID: &cage10}, wantInvalidInput: true},
		{name: "expected_null without cage operation is invalid", lockedCage: nil,
			input: &UpdateHospitalizationInput{ExpectedCageIsNull: true}, wantInvalidInput: true},
		{name: "both expected guards on cage change is invalid", lockedCage: nil,
			input: &UpdateHospitalizationInput{CageID: &cage12, ExpectedCageID: &cage10, ExpectedCageIsNull: true}, wantInvalidInput: true},
		{name: "memo only needs no expected", lockedCage: &cage10,
			input: &UpdateHospitalizationInput{Memo: &memo}, wantUpdated: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existing := &model.Hospitalization{
				ID: 1, ClinicID: 1,
				CageID: tt.lockedCage,
			}
			updated := false
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) { return existing, nil },
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
					updated = true
					result := *existing
					result.CageID = cmd.CageID
					return &result, nil
				},
			}
			cageRepo := &mockCageRepository{
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Cage, error) {
					return &model.Cage{ID: id}, nil
				},
			}
			svc := NewHospitalizationService(hospRepo, nil, nil, cageRepo, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

			got, err := svc.Update(context.Background(), 1, 1, tt.input)

			switch {
			case tt.wantInvalidInput:
				require.Error(t, err)
				assert.Nil(t, got)
				assert.True(t, apperrors.IsInvalidInput(err), "want invalid input: %v", err)
				assert.False(t, updated, "invalid input must not write")
			case tt.wantConflict:
				require.Error(t, err)
				assert.Nil(t, got)
				assert.True(t, apperrors.IsConflict(err), "want conflict: %v", err)
				assert.False(t, updated, "conflict must not write")
			default:
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, updated)
			}
		})
	}
}

func TestHospitalizationService_Update_ClearCageOnReactivation(t *testing.T) {
	tests := []struct {
		name      string
		oldMember bool
	}{
		{"former member", true},
		{"occupied representative", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldCageID, otherCageID := uint64(10), uint64(20)
			existing := &model.Hospitalization{
				ID: 1, ClinicID: 1,
				CageID: &oldCageID,
			}
			cageReads := 0
			cageRepo := &mockCageRepository{
				lockCageGroupForUpdateFn: func(_ context.Context, _, _ uint64) ([]model.Cage, error) {
					cageReads++
					if tt.oldMember {
						return []model.Cage{{ID: oldCageID, MergedIntoCageID: &otherCageID}}, nil
					}
					return []model.Cage{{ID: oldCageID}, {ID: otherCageID, MergedIntoCageID: &oldCageID}}, nil
				},
				countOccupiedCagesFn: func(_ context.Context, _ uint64, _ []uint64) (map[uint64]int64, error) {
					return map[uint64]int64{oldCageID: 1}, nil
				},
			}
			updated := false
			hospRepo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) { return existing, nil },
				updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
					updated = true
					fields := buildHospitalizationUpdate(&cmd)
					value, present := fields["cage_id"]
					require.True(t, present)
					require.Nil(t, value)
					result := *existing
					result.CageID = nil
					return &result, nil
				},
			}
			svc := NewHospitalizationService(hospRepo, nil, nil, cageRepo, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))
			input, err := (&updateHospitalizationRequest{ClearCageID: true, ExpectedCageID: &oldCageID}).toServiceInput()
			require.NoError(t, err)

			got, err := svc.Update(context.Background(), 1, 1, &input)

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.True(t, updated)
			assert.Nil(t, got.CageID)
			assert.Zero(t, cageReads, "the cleared former cage is not an assignment target")
		})
	}
}

// ---- DischargeWithBilling ----

// dischargeTestDeps は BE9-2D ⑤ Phase1 の個別注入コンストラクタ向け discharge テスト共通配線。
// 旧 harness の repos.TransactionFn インライン実行は mockTransactor の WithTx 素通しが等価。
type dischargeTestDeps struct {
	hosp        HospitalizationRepository
	carePlan    CarePlanItemRepository
	accounting  accountingCreator
	billingItem billingItemWriter
	reservation *mockReservationRepository
	audit       AuditTxLogger
}

func newDischargeTestDeps(hospRepo HospitalizationRepository, carePlanRepo CarePlanItemRepository, accountingRepo accountingCreator, billingItemRepo billingItemWriter) *dischargeTestDeps {
	return &dischargeTestDeps{
		hosp:        hospRepo,
		carePlan:    carePlanRepo,
		accounting:  accountingRepo,
		billingItem: billingItemRepo,
		audit:       &mockTreatmentAuditTxLogger{},
		reservation: &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
			findPetOwnerInClinicFn: func(_ context.Context, _, petID uint64) (uint64, error) {
				// Discharge fixtures use OwnerID=2 / PetID=5 by default.
				if petID == 5 {
					return 2, nil
				}
				return petID, nil
			},
		},
	}
}

func (d *dischargeTestDeps) svc() HospitalizationService {
	return NewHospitalizationServiceWithAudit(d.hosp, d.reservation, nil, nil, d.carePlan, d.accounting, d.billingItem, &mockTransactor{}, d.audit, WithPetLifeStateReader(stubPetLifeReader{}))
}
