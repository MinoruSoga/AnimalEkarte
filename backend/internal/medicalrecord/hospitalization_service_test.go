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

// mockHospitalizationRepository は HospitalizationRepository のテスト用モック実装
type mockHospitalizationRepository struct {
	findAllFn                                func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error)
	findByIDFn                               func(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
	createFn                                 func(ctx context.Context, hospitalization *model.Hospitalization) error
	updateFieldsFn                           func(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error)
	updateIfNotDischargedFn                  func(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error)
	deleteFn                                 func(ctx context.Context, clinicID, id uint64) error
	countCarePlanItemsByHospitalizationIDFn  func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countDailyRecordsByHospitalizationIDFn   func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
	countTreatmentPlansByHospitalizationIDFn func(ctx context.Context, clinicID, hospitalizationID uint64) (int64, error)
}

func (m *mockHospitalizationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Hospitalization, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, status, startDate, endDate, page, limit)
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

func (m *mockHospitalizationRepository) UpdateIfNotDischarged(ctx context.Context, clinicID, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
	if m.updateIfNotDischargedFn != nil {
		return m.updateIfNotDischargedFn(ctx, clinicID, id, cmd)
	}
	return nil, apperrors.WrapNotFound("hospitalization", "updateIfNotDischargedFn not stubbed")
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
	status := string(model.HospitalizationStatusAdmitted)

	tests := []struct {
		name      string
		clinicID  uint64
		petID     *uint64
		ownerID   *uint64
		status    *string
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
				{ID: 1, ClinicID: 1, PetID: 5, OwnerID: 2, Status: model.HospitalizationStatusAdmitted},
				{ID: 2, ClinicID: 1, PetID: 6, OwnerID: 3, Status: model.HospitalizationStatusReserved},
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
			name:     "filters by status",
			clinicID: 1,
			status:   &status,
			page:     1,
			limit:    20,
			repoItems: []model.Hospitalization{
				{ID: 1, ClinicID: 1, Status: model.HospitalizationStatusAdmitted},
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
				findAllFn: func(_ context.Context, _ uint64, _ *uint64, _ *uint64, _, _, _ *string, _, _ int) ([]model.Hospitalization, int64, error) {
					return tt.repoItems, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{})

			items, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, tt.status, nil, nil, tt.page, tt.limit)

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
	now := time.Now()
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
				ID:        10,
				ClinicID:  1,
				PetID:     5,
				OwnerID:   2,
				Status:    model.HospitalizationStatusAdmitted,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
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
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{})

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
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	futureStart := today.AddDate(0, 0, 3)
	tests := []struct {
		name       string
		clinicID   uint64
		input      *CreateHospitalizationInput
		repoErr    error
		wantErr    bool
		wantStatus model.HospitalizationStatus
	}{
		{
			name:     "creates hospitalization successfully",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				CageID:              func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
				Status:              model.HospitalizationStatusReserved,
			},
			repoErr:    nil,
			wantErr:    false,
			wantStatus: model.HospitalizationStatusReserved, // explicit status kept
		},
		{
			name:     "defaults status to admitted when empty and start_date is today (BUG-031)",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				CageID:              func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeHotel,
				StartDate:           today,
				EndDate:             today.AddDate(0, 0, 7),
			},
			repoErr:    nil,
			wantErr:    false,
			wantStatus: model.HospitalizationStatusAdmitted,
		},
		{
			name:     "defaults status to reserved when empty and start_date is future (BUG-031)",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				CageID:              func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           futureStart,
				EndDate:             futureStart.AddDate(0, 0, 7),
			},
			repoErr:    nil,
			wantErr:    false,
			wantStatus: model.HospitalizationStatusReserved,
		},
		{
			name:     "returns error when already exists",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				CageID:    func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:   2,
				PetID:     5,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			repoErr: apperrors.WrapAlreadyExists("hospitalization", now.String()),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				CageID:    func() *uint64 { v := uint64(10); return &v }(),
				OwnerID:   2,
				PetID:     5,
				StartDate: now,
				EndDate:   now.Add(24 * time.Hour),
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:     "rejects missing cage_id without persisting (BUG-037)",
			clinicID: 1,
			input: &CreateHospitalizationInput{
				OwnerID:             2,
				PetID:               5,
				HospitalizationType: model.HospitalizationTypeInpatient,
				StartDate:           now,
				EndDate:             now.Add(24 * time.Hour),
			},
			repoErr: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *model.Hospitalization
			createCalls := 0
			repo := &mockHospitalizationRepository{
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
			}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

			hosp, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hosp)
				if tt.name == "rejects missing cage_id without persisting (BUG-037)" {
					assert.Equal(t, 0, createCalls)
					assert.True(t, apperrors.IsInvalidInput(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, hosp)
				require.NotNil(t, created)
				assert.Equal(t, tt.wantStatus, created.Status)
				assert.Equal(t, tt.wantStatus, hosp.Status)
			}
		})
	}
}

func TestDefaultHospitalizationStatus(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 8, 10, 15, 30, 0, 0, loc)
	today := time.Date(2026, 8, 10, 0, 0, 0, 0, loc)
	assert.Equal(t, model.HospitalizationStatusAdmitted, defaultHospitalizationStatus(today, now))
	assert.Equal(t, model.HospitalizationStatusAdmitted, defaultHospitalizationStatus(today.Add(-time.Hour), now))
	assert.Equal(t, model.HospitalizationStatusAdmitted, defaultHospitalizationStatus(today.AddDate(0, 0, -1), now))
	assert.Equal(t, model.HospitalizationStatusReserved, defaultHospitalizationStatus(today.AddDate(0, 0, 1), now))
	// UTC midnight that is still clinic-local "today" when offset is +9
	utcStillToday := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, model.HospitalizationStatusAdmitted, defaultHospitalizationStatus(utcStillToday, now))
}

func TestHospitalizationService_Update(t *testing.T) {
	now := time.Now()
	statusAdmitted := model.HospitalizationStatusAdmitted
	tests := []struct {
		name    string
		input   UpdateHospitalizationInput
		repoErr error
		wantErr bool
	}{
		{
			name: "updates hospitalization successfully",
			input: UpdateHospitalizationInput{
				Status:    &statusAdmitted,
				StartDate: &now,
			},
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
			input:   UpdateHospitalizationInput{Status: &statusAdmitted},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockHospitalizationRepository{
				findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
					return &model.Hospitalization{
						ID:        id,
						ClinicID:  clinicID,
						StartDate: now.Add(-24 * time.Hour),
						EndDate:   now.Add(24 * time.Hour),
						Status:    model.HospitalizationStatusAdmitted,
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Hospitalization{ID: 1, ClinicID: 1, Status: model.HospitalizationStatusAdmitted}, nil
				},
			}
			svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{})

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
	svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{})

	hosp, err := svc.Update(context.Background(), 1, 1, nil)

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, hosp)
}

func TestHospitalizationService_Update_FindByIDError(t *testing.T) {
	statusAdmitted := model.HospitalizationStatusAdmitted
	repo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "999")
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			t.Fatal("hospitalization must not be updated when the parent lookup fails")
			return nil, nil
		},
	}
	svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{})

	hosp, err := svc.Update(context.Background(), 1, 999, &UpdateHospitalizationInput{Status: &statusAdmitted})

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
					return &model.Hospitalization{ID: id, ClinicID: clinicID, Status: model.HospitalizationStatusAdmitted}, nil
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
			svc := NewHospitalizationServiceWithAudit(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, okCarePlanAuditTx{})

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
	svc := NewHospitalizationServiceWithAudit(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, okCarePlanAuditTx{})

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestHospitalizationService_Create_InsuranceFields(t *testing.T) {
	now := time.Now()
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
				HospitalizationType:  model.HospitalizationTypeInpatient,
				StartDate:            now,
				EndDate:              now.Add(24 * time.Hour),
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
				HospitalizationType:  model.HospitalizationTypeInpatient,
				StartDate:            now,
				EndDate:              now.Add(24 * time.Hour),
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
			}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

			hosp, err := svc.Create(context.Background(), 1, tt.input)

			assert.NoError(t, err)
			assert.NotNil(t, hosp)
			assert.Equal(t, tt.wantInsuranceCompany, captured.InsuranceCompanyName)
			assert.Equal(t, tt.wantInsuranceNumber, captured.InsuranceNumber)
		})
	}
}

func TestHospitalizationService_Create_RejectsDeceasedPet(t *testing.T) {
	now := time.Now()
	deceasedAt := now.Add(-24 * time.Hour)
	repo := &mockHospitalizationRepository{
		createFn: func(_ context.Context, _ *model.Hospitalization) error {
			t.Fatal("hospitalization must not be created for a deceased pet")
			return nil
		},
	}
	svc := NewHospitalizationService(repo, &mockReservationRepository{
		assertOwnerInClinicFn:  func(_ context.Context, _, _ uint64) error { return nil },
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) { return 2, nil },
	}, &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
		},
	}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{})

	hosp, err := svc.Create(context.Background(), 1, &CreateHospitalizationInput{
		CageID:              func() *uint64 { v := uint64(10); return &v }(),
		OwnerID:             2,
		PetID:               5,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now.Add(24 * time.Hour),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, hosp)
}

func TestHospitalizationService_Update_RejectsDeceasedPetReplacement(t *testing.T) {
	deceasedAt := time.Now().Add(-24 * time.Hour)
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
	}, &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id, DeceasedAt: &deceasedAt}, nil
		},
	}, nil, nil, nil, nil, &mockTransactor{})

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
	hospType := model.HospitalizationTypeInpatient
	start := time.Now()
	end := start.Add(24 * time.Hour)
	status := model.HospitalizationStatusAdmitted
	cageID := uint64(3)
	doctorID := uint64(7)
	memo := "メモ"
	ownerRequest := "希望事項"
	staffNotes := "スタッフメモ"

	fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{
		OwnerID:             &ownerID,
		PetID:               &petID,
		HospitalizationType: &hospType,
		StartDate:           &start,
		EndDate:             &end,
		Status:              &status,
		CageID:              &cageID,
		DoctorID:            &doctorID,
		Memo:                &memo,
		OwnerRequest:        &ownerRequest,
		StaffNotes:          &staffNotes,
	})

	assert.Equal(t, ownerID, fields["owner_id"])
	assert.Equal(t, petID, fields["pet_id"])
	assert.Equal(t, hospType, fields["hospitalization_type"])
	assert.Equal(t, start, fields["start_date"])
	assert.Equal(t, end, fields["end_date"])
	assert.Equal(t, status, fields["status"])
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
	return NewHospitalizationServiceWithAudit(d.hosp, d.reservation, nil, nil, d.carePlan, d.accounting, d.billingItem, &mockTransactor{}, d.audit)
}

func TestHospitalizationService_DischargeWithBilling_NotFound(t *testing.T) {
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
			return nil, apperrors.WrapNotFound("hospitalization", "999")
		},
	}
	svc := newDischargeTestDeps(hospRepo, nil, nil, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 999, DischargeWithBillingInput{})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_AlreadyDischarged(t *testing.T) {
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusDischarged}, nil
		},
	}
	svc := newDischargeTestDeps(hospRepo, nil, nil, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{})

	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_UpdateFails(t *testing.T) {
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newDischargeTestDeps(hospRepo, nil, nil, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now()})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_WithoutAccounting(t *testing.T) {
	var updated UpdateHospitalizationInput
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
			updated = cmd
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			t.Fatal("care plan items must not be fetched when CreateAccounting is false")
			return nil, nil
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, nil, nil).svc()

	dischargeDate := time.Now()
	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: dischargeDate, CreateAccounting: false})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.AccountingID)
	assert.Equal(t, string(model.HospitalizationStatusDischarged), result.Status)
	require.NotNil(t, updated.Status)
	require.NotNil(t, updated.EndDate)
	assert.Equal(t, model.HospitalizationStatusDischarged, *updated.Status)
	assert.Equal(t, dischargeDate, *updated.EndDate)
}

func TestHospitalizationService_DischargeWithBilling_CarePlanItemsFetchError(t *testing.T) {
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return nil, errors.New("db error")
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, nil, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now(), CreateAccounting: true, ActorID: uint64PtrHosp(1)})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_BillingCreateError(t *testing.T) {
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return nil, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			return errors.New("db error")
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now(), CreateAccounting: true, ActorID: uint64PtrHosp(1)})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_WithCarePlanItems(t *testing.T) {
	items := []model.CarePlanItem{
		{ID: 1, Type: model.CarePlanTypeFood, Name: "食事介助", UnitPrice: 1000},
		{
			ID:        2,
			Type:      model.CarePlanTypeTreatment,
			Name:      "手術",
			UnitPrice: 2000,
			Procedure: &model.Procedure{IsSurgery: true},
		},
	}
	var createdItems []*model.BillingItem
	totalsUpdated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID:                  id,
				Status:              model.HospitalizationStatusAdmitted,
				HospitalizationType: model.HospitalizationTypeInpatient,
				PetID:               5,
				OwnerID:             2,
			}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return items, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			billing.ID = 55
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, item *model.BillingItem) error {
			createdItems = append(createdItems, item)
			return nil
		},
		updateBillingTotals: func(_ context.Context, _, billingID uint64, subtotal, taxTotal, totalAmount int64) error {
			totalsUpdated = true
			assert.Equal(t, uint64(55), billingID)
			assert.Equal(t, int64(3000), subtotal)
			assert.Equal(t, int64(300), taxTotal)
			assert.Equal(t, int64(3300), totalAmount)
			return nil
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now(), CreateAccounting: true, ActorID: uint64PtrHosp(1)})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.AccountingID)
	assert.Equal(t, uint64(55), *result.AccountingID)
	assert.Len(t, createdItems, 2)
	assert.Equal(t, model.ItemCategoryFood, createdItems[0].Category)
	assert.Equal(t, model.ItemCategorySurgery, createdItems[1].Category)
	assert.True(t, totalsUpdated)
}

func TestHospitalizationService_DischargeWithBilling_BillingItemCreateError(t *testing.T) {
	items := []model.CarePlanItem{{ID: 1, Name: "食事介助", UnitPrice: 1000}}
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return items, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			billing.ID = 55
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, _ *model.BillingItem) error {
			return errors.New("db error")
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now(), CreateAccounting: true, ActorID: uint64PtrHosp(1)})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_UpdateBillingTotalsError(t *testing.T) {
	items := []model.CarePlanItem{{ID: 1, Name: "食事介助", UnitPrice: 1000}}
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
			return &model.Hospitalization{ID: 10}, nil
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			return items, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, billing *model.Billing) error {
			billing.ID = 55
			return nil
		},
	}
	billingItemRepo := &mockBillingItemRepository{
		createFn: func(_ context.Context, _ *model.BillingItem) error {
			return nil
		},
		updateBillingTotals: func(_ context.Context, _, _ uint64, _, _, _ int64) error {
			return errors.New("db error")
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, billingItemRepo).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{DischargeDate: time.Now(), CreateAccounting: true, ActorID: uint64PtrHosp(1)})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestHospitalizationService_DischargeWithBilling_ConcurrentDoubleDischarge_ReturnsNotFoundWithoutAccounting(t *testing.T) {
	accountingCreated := false
	hospRepo := &mockHospitalizationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
			return &model.Hospitalization{
				ID: id, Status: model.HospitalizationStatusAdmitted, PetID: 5, OwnerID: 2,
			}, nil
		},
		updateIfNotDischargedFn: func(_ context.Context, _, id uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
			assert.Equal(t, uint64(10), id)
			require.NotNil(t, cmd.Status)
			assert.Equal(t, model.HospitalizationStatusDischarged, *cmd.Status)
			return nil, apperrors.WrapNotFound("hospitalization", "10")
		},
	}
	carePlanRepo := &mockCarePlanItemRepository{
		listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
			t.Fatal("care plan items must not be fetched when conditional discharge update loses the race")
			return nil, nil
		},
	}
	accountingRepo := &mockAccountingRepository{
		createFn: func(_ context.Context, _ uint64, _ *model.Billing) error {
			accountingCreated = true
			t.Fatal("Accounting.Create must not run when UpdateIfNotDischarged returns NotFound")
			return nil
		},
	}
	svc := newDischargeTestDeps(hospRepo, carePlanRepo, accountingRepo, nil).svc()

	result, err := svc.DischargeWithBilling(context.Background(), 1, 10, DischargeWithBillingInput{
		DischargeDate:    time.Now(),
		CreateAccounting: true,
		ActorID:          uint64PtrHosp(1),
	})

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, result)
	assert.False(t, accountingCreated)
}

// SEC-DUR-01-MR-T1: 譲渡後の会計なし退院は、snapshot owner と current pet owner の差を許容し clinic 外は拒否する。
func TestHospitalizationService_DischargeWithBilling_WithoutAccounting_AllowsHistoricalOwnerAfterPetTransfer(t *testing.T) {
	const (
		clinicID        = uint64(1)
		hospID          = uint64(10)
		previousOwnerID = uint64(20)
		currentOwnerID  = uint64(21)
		petID           = uint64(30)
	)

	t.Run("same_clinic_transfer_succeeds", func(t *testing.T) {
		updated := false
		hospRepo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{
					ID: id, ClinicID: clinicID,
					OwnerID: previousOwnerID, PetID: petID,
					Status: model.HospitalizationStatusAdmitted,
				}, nil
			},
			updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
				updated = true
				require.NotNil(t, cmd.Status)
				assert.Equal(t, model.HospitalizationStatusDischarged, *cmd.Status)
				return &model.Hospitalization{ID: hospID, Status: model.HospitalizationStatusDischarged}, nil
			},
		}
		carePlanRepo := &mockCarePlanItemRepository{
			listByHospitalizationIDFn: func(_ context.Context, _, _ uint64) ([]model.CarePlanItem, error) {
				t.Fatal("care plan items must not be fetched when CreateAccounting is false")
				return nil, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		deps := newDischargeTestDeps(hospRepo, carePlanRepo, nil, nil)
		deps.reservation = resRepo
		svc := deps.svc()

		result, err := svc.DischargeWithBilling(context.Background(), clinicID, hospID, DischargeWithBillingInput{
			DischargeDate:    time.Now(),
			CreateAccounting: false,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, updated)
	})

	t.Run("rejects_foreign_snapshot_owner", func(t *testing.T) {
		updated := false
		hospRepo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{
					ID: id, ClinicID: clinicID,
					OwnerID: previousOwnerID, PetID: petID,
					Status: model.HospitalizationStatusAdmitted,
				}, nil
			},
			updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				updated = true
				return &model.Hospitalization{ID: hospID}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				assert.Equal(t, previousOwnerID, ownerID)
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, gotPetID uint64) (uint64, error) {
				if gotPetID == petID {
					return currentOwnerID, nil
				}
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		deps := newDischargeTestDeps(hospRepo, &mockCarePlanItemRepository{}, nil, nil)
		deps.reservation = resRepo
		svc := deps.svc()

		result, err := svc.DischargeWithBilling(context.Background(), clinicID, hospID, DischargeWithBillingInput{
			DischargeDate:    time.Now(),
			CreateAccounting: false,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, result)
		assert.False(t, updated)
	})

	t.Run("rejects_foreign_pet", func(t *testing.T) {
		updated := false
		hospRepo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{
					ID: id, ClinicID: clinicID,
					OwnerID: previousOwnerID, PetID: petID,
					Status: model.HospitalizationStatusAdmitted,
				}, nil
			},
			updateIfNotDischargedFn: func(_ context.Context, _, _ uint64, _ UpdateHospitalizationInput) (*model.Hospitalization, error) {
				updated = true
				return &model.Hospitalization{ID: hospID}, nil
			},
		}
		resRepo := &mockReservationRepository{
			assertOwnerInClinicFn: func(_ context.Context, _, ownerID uint64) error {
				if ownerID == previousOwnerID {
					return nil
				}
				return apperrors.WrapNotFound("owner", "scoped")
			},
			findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
				return 0, apperrors.WrapNotFound("pet", "scoped")
			},
		}
		deps := newDischargeTestDeps(hospRepo, &mockCarePlanItemRepository{}, nil, nil)
		deps.reservation = resRepo
		svc := deps.svc()

		result, err := svc.DischargeWithBilling(context.Background(), clinicID, hospID, DischargeWithBillingInput{
			DischargeDate:    time.Now(),
			CreateAccounting: false,
		})
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Nil(t, result)
		assert.False(t, updated)
	})
}
