package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- Checkup モック ----

type mockCheckupRepository struct {
	listByMedicalRecordIDFn func(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	listByClinicFn          func(ctx context.Context, clinicID uint64, filters CheckupFilters, page, limit int) ([]model.Checkup, int64, error)
	findByOwnerIDFn         func(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
	findByIDFn              func(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error)
	lockByIDForUpdateFn     func(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error)
	createFn                func(ctx context.Context, checkup *model.Checkup) error
	updateFn                func(ctx context.Context, clinicID, checkupID uint64, cmd UpdateCheckupInput) error
	deleteFn                func(ctx context.Context, clinicID, checkupID uint64) error
}

func (m *mockCheckupRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	return m.listByMedicalRecordIDFn(ctx, clinicID, medicalRecordID)
}

func (m *mockCheckupRepository) FindByClinicID(ctx context.Context, clinicID uint64, filters CheckupFilters, page, limit int) ([]model.Checkup, int64, error) {
	if m.listByClinicFn != nil {
		return m.listByClinicFn(ctx, clinicID, filters, page, limit)
	}
	return nil, 0, nil
}

func (m *mockCheckupRepository) FindByID(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, checkupID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) LockByIDForUpdate(ctx context.Context, clinicID, checkupID uint64) (*model.Checkup, error) {
	if m.lockByIDForUpdateFn != nil {
		return m.lockByIDForUpdateFn(ctx, clinicID, checkupID)
	}
	return m.FindByID(ctx, clinicID, checkupID)
}

func (m *mockCheckupRepository) FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error) {
	if m.findByOwnerIDFn != nil {
		return m.findByOwnerIDFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockCheckupRepository) Create(ctx context.Context, checkup *model.Checkup) error {
	return m.createFn(ctx, checkup)
}

func (m *mockCheckupRepository) Update(ctx context.Context, clinicID, checkupID uint64, cmd UpdateCheckupInput) error {
	return m.updateFn(ctx, clinicID, checkupID, cmd)
}

func (m *mockCheckupRepository) Delete(ctx context.Context, clinicID, checkupID uint64) error {
	return m.deleteFn(ctx, clinicID, checkupID)
}

// ---- Tests ----

func TestCheckupService_List(t *testing.T) {
	tests := []struct {
		name            string
		medicalRecordID uint64
		repoCheckups    []model.Checkup
		repoErr         error
		wantLen         int
		wantErr         bool
	}{
		{
			name:            "returns checkups for medical record",
			medicalRecordID: 1,
			repoCheckups: []model.Checkup{
				{ID: 1, MedicalRecordID: 1, CheckupTypeID: 1, Result: "Normal"},
				{ID: 2, MedicalRecordID: 1, CheckupTypeID: 2, Result: "Abnormal"},
			},
			repoErr: nil,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:            "returns empty list when no checkups exist",
			medicalRecordID: 999,
			repoCheckups:    []model.Checkup{},
			repoErr:         nil,
			wantLen:         0,
			wantErr:         false,
		},
		{
			name:            "propagates repository error",
			medicalRecordID: 1,
			repoCheckups:    nil,
			repoErr:         errors.New("db error"),
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				listByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) ([]model.Checkup, error) {
					return tt.repoCheckups, tt.repoErr
				},
			}
			svc := NewCheckupService(repo, &mockMedicalRecordRepository{}, okCheckupTypeRepo(), nil, nil)

			checkups, err := svc.List(context.Background(), 1, tt.medicalRecordID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, checkups, tt.wantLen)
			}
		})
	}
}

func TestCheckupService_ListByClinic(t *testing.T) {
	startDate := "2026-05-01"
	petID := uint64(42)
	endDate := "2026-05-31"

	tests := []struct {
		name           string
		input          ListCheckupsByClinicInput
		repoCheckups   []model.Checkup
		repoTotal      int64
		repoErr        error
		wantLen        int
		wantTotal      int64
		wantErr        bool
		checkedFilters func(t *testing.T, filters CheckupFilters)
		checkedPaging  func(t *testing.T, page, limit int)
	}{
		{
			name: "returns checkups filtered by date range",
			input: ListCheckupsByClinicInput{
				ClinicID:  1,
				PetID:     &petID,
				StartDate: &startDate,
				EndDate:   &endDate,
				Page:      1,
				Limit:     20,
			},
			repoCheckups: []model.Checkup{
				{ID: 1, MedicalRecordID: 1},
				{ID: 2, MedicalRecordID: 2},
			},
			repoTotal: 2,
			wantLen:   2,
			wantTotal: 2,
			checkedFilters: func(t *testing.T, filters CheckupFilters) {
				assert.Equal(t, &petID, filters.PetID)
				assert.Equal(t, &startDate, filters.StartDate)
				assert.Equal(t, &endDate, filters.EndDate)
			},
		},
		{
			name:  "returns empty list when no checkups exist",
			input: ListCheckupsByClinicInput{ClinicID: 1, Page: 1, Limit: 20},
		},
		{
			name:    "propagates repository error",
			input:   ListCheckupsByClinicInput{ClinicID: 1, Page: 1, Limit: 20},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "passes page/limit through to repository and returns total",
			input: ListCheckupsByClinicInput{
				ClinicID: 1,
				Page:     2,
				Limit:    5,
			},
			repoCheckups: []model.Checkup{
				{ID: 6, MedicalRecordID: 1},
			},
			repoTotal: 11,
			wantLen:   1,
			wantTotal: 11,
			checkedPaging: func(t *testing.T, page, limit int) {
				assert.Equal(t, 2, page)
				assert.Equal(t, 5, limit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotFilters CheckupFilters
			var gotPage, gotLimit int
			repo := &mockCheckupRepository{
				listByClinicFn: func(_ context.Context, _ uint64, filters CheckupFilters, page, limit int) ([]model.Checkup, int64, error) {
					gotFilters = filters
					gotPage = page
					gotLimit = limit
					return tt.repoCheckups, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewCheckupService(repo, &mockMedicalRecordRepository{}, okCheckupTypeRepo(), nil, nil)

			checkups, total, err := svc.ListByClinic(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checkups)
				assert.Zero(t, total)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, checkups, tt.wantLen)
			assert.Equal(t, tt.wantTotal, total)
			if tt.checkedFilters != nil {
				tt.checkedFilters(t, gotFilters)
			}
			if tt.checkedPaging != nil {
				tt.checkedPaging(t, gotPage, gotLimit)
			}
		})
	}
}

func TestCheckupService_Create(t *testing.T) {
	now := time.Now()
	petID := uint64(5)
	doctorID := uint64(10)

	tests := []struct {
		name            string
		medicalRecordID uint64
		input           *CreateCheckupInput
		repoErr         error
		wantErr         bool
	}{
		{
			name:            "creates checkup successfully",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
				PetID:         &petID,
				DoctorID:      &doctorID,
				Result:        "Normal findings",
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "creates checkup with minimal fields",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:            "returns error when repository fails",
			medicalRecordID: 1,
			input: &CreateCheckupInput{
				CheckupTypeID: 1,
				Date:          now,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				createFn: func(_ context.Context, _ *model.Checkup) error {
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{ID: 1, MedicalRecordID: tt.medicalRecordID}, nil
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					record := &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}
					if tt.input.PetID != nil {
						record.OwnerID = ptrUint64(1)
						record.PetID = tt.input.PetID
					}
					return record, nil
				},
			}
			svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

			checkup, err := svc.Create(context.Background(), tt.medicalRecordID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, checkup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Update(t *testing.T) {
	newResult := "Updated result"
	newCheckupTypeID := uint64(2)

	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		input                      *UpdateCheckupInput
		repoCheckupMedicalRecordID uint64
		repoUpdateErr              error
		repoReturnCheckup          *model.Checkup
		wantErr                    bool
	}{
		{
			name:            "updates checkup successfully",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result:        &newResult,
				CheckupTypeID: &newCheckupTypeID,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 1,
				Result:          newResult,
			},
			wantErr: false,
		},
		{
			name:                       "returns error when no fields provided",
			medicalRecordID:            1,
			checkupID:                  1,
			input:                      &UpdateCheckupInput{},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              nil,
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
		{
			name:            "returns error when checkup doesn't belong to medical record",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoUpdateErr:              nil,
			repoReturnCheckup: &model.Checkup{
				ID:              1,
				MedicalRecordID: 2,
			},
			wantErr: true,
		},
		{
			name:            "returns error when update fails",
			medicalRecordID: 1,
			checkupID:       1,
			input: &UpdateCheckupInput{
				Result: &newResult,
			},
			repoCheckupMedicalRecordID: 1,
			repoUpdateErr:              errors.New("db error"),
			repoReturnCheckup:          nil,
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				updateFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupInput) error {
					return tt.repoUpdateErr
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
				},
			}
			svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

			checkup, err := svc.Update(context.Background(), 1, tt.medicalRecordID, tt.checkupID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, checkup)
			}
		})
	}
}

func TestCheckupService_Delete(t *testing.T) {
	tests := []struct {
		name                       string
		medicalRecordID            uint64
		checkupID                  uint64
		repoCheckupMedicalRecordID uint64
		repoDeleteErr              error
		wantErr                    bool
	}{
		{
			name:                       "deletes checkup successfully",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              nil,
			wantErr:                    false,
		},
		{
			name:                       "returns error when checkup doesn't belong to medical record",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 2, // Different medical record
			repoDeleteErr:              nil,
			wantErr:                    true,
		},
		{
			name:                       "returns error when delete fails",
			medicalRecordID:            1,
			checkupID:                  1,
			repoCheckupMedicalRecordID: 1,
			repoDeleteErr:              errors.New("db error"),
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockCheckupRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{
						ID:              tt.checkupID,
						MedicalRecordID: tt.repoCheckupMedicalRecordID,
					}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoDeleteErr
				},
			}
			mrRepo := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
				},
			}
			svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

			err := svc.Delete(context.Background(), 1, tt.medicalRecordID, tt.checkupID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckupService_DeleteLocksParentBeforeCheckup(t *testing.T) {
	order := make([]string, 0, 2)
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: checkupID, MedicalRecordID: 10}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
			order = append(order, "checkup")
			return &model.Checkup{ID: checkupID, MedicalRecordID: 10}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	records := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
		},
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			order = append(order, "medical_record")
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewCheckupService(repo, records, okCheckupTypeRepo(), nil, nil)

	err := svc.Delete(context.Background(), 1, 10, 20)

	assert.NoError(t, err)
	assert.Equal(t, []string{"medical_record", "checkup"}, order)
}

func TestCheckupService_DeleteRequiresTransactionDependency(t *testing.T) {
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, checkupID uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: checkupID, MedicalRecordID: 10}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
	}
	records := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{ID: id, ClinicID: clinicID, Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewCheckupService(
		repo,
		records,
		okCheckupTypeRepo(),
		nil,
		nil,
		CheckupWriteDependencies{Transactor: nil},
	)

	err := svc.Delete(context.Background(), 1, 10, 20)

	assert.Error(t, err)
}

func TestCheckupService_Create_SyncsCheckupTagBestEffort(t *testing.T) {
	ownerID := uint64(10)
	nextDate := time.Date(2026, 6, 27, 0, 0, 0, 0, time.UTC)
	var syncedClinicID, syncedOwnerID, syncedCheckupTypeID uint64
	var syncedDate time.Time
	var syncedNextDate *time.Time

	repo := &mockCheckupRepository{
		createFn: func(_ context.Context, checkup *model.Checkup) error {
			checkup.ID = 30
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{
				ID:              30,
				MedicalRecordID: 2,
				CheckupTypeID:   3,
				Date:            time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
				NextDate:        &nextDate,
				MedicalRecord:   &model.MedicalRecord{OwnerID: &ownerID},
			}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncCheckupTagFn: func(_ context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			syncedCheckupTypeID = checkupTypeID
			syncedDate = checkupDate
			syncedNextDate = nextDate
			return errors.New("sync failed")
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, tagSync)

	result, err := svc.Create(context.Background(), 2, &CreateCheckupInput{
		ClinicID:      1,
		CheckupTypeID: 3,
		Date:          time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		NextDate:      &nextDate,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
	assert.Equal(t, uint64(3), syncedCheckupTypeID)
	assert.Equal(t, result.Date, syncedDate)
	assert.Equal(t, result.NextDate, syncedNextDate)
}

func TestCheckupService_Update_ResyncsOwnerCheckupTags(t *testing.T) {
	ownerID := uint64(10)
	medicalRecordID := uint64(2)
	findCount := 0
	var syncedOwnerID uint64

	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			findCount++
			if findCount == 1 {
				return &model.Checkup{ID: 30, MedicalRecordID: medicalRecordID}, nil
			}
			return &model.Checkup{
				ID:              30,
				MedicalRecordID: medicalRecordID,
				MedicalRecord:   &model.MedicalRecord{OwnerID: &ownerID},
			}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupInput) error {
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerCheckupTagsFn: func(_ context.Context, _, ownerID uint64) error {
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	resultText := "updated"
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, tagSync)

	result, err := svc.Update(context.Background(), 1, medicalRecordID, 30, &UpdateCheckupInput{Result: &resultText})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestCheckupService_Delete_ResyncsOwnerCheckupTagsAfterDelete(t *testing.T) {
	ownerID := uint64(10)
	medicalRecordID := uint64(2)
	deleted := false
	syncedAfterDelete := false

	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{
				ID:              30,
				MedicalRecordID: medicalRecordID,
				MedicalRecord:   &model.MedicalRecord{OwnerID: &ownerID},
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleted = true
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerCheckupTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			syncedAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, tagSync)

	err := svc.Delete(context.Background(), 1, medicalRecordID, 30)

	assert.NoError(t, err)
	assert.True(t, syncedAfterDelete)
}

func TestCheckupService_Update_FindByIDError(t *testing.T) {
	newResult := "updated"
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(repo, &mockMedicalRecordRepository{}, okCheckupTypeRepo(), nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCheckupInput{Result: &newResult})

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCheckupService_Delete_FindByIDError(t *testing.T) {
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(repo, &mockMedicalRecordRepository{}, okCheckupTypeRepo(), nil, nil)

	err := svc.Delete(context.Background(), 1, 1, 1)

	assert.Error(t, err)
}

// TestBuildCheckupUpdate_AllFields は buildCheckupUpdate の全フィールド分岐を直接検証する。
func TestBuildCheckupUpdate_AllFields(t *testing.T) {
	t.Run("フィールド未指定時は空map", func(t *testing.T) {
		fields := buildCheckupUpdate(&UpdateCheckupInput{})
		assert.Empty(t, fields)
	})

	t.Run("CheckupTypeID/PetID/Date/NextDate/DoctorID/Result すべて反映される", func(t *testing.T) {
		checkupTypeID := uint64(2)
		petID := uint64(5)
		date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		nextDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		doctorID := uint64(9)
		result := "所見あり"

		fields := buildCheckupUpdate(&UpdateCheckupInput{
			CheckupTypeID: &checkupTypeID,
			PetID:         &petID,
			Date:          &date,
			NextDate:      &nextDate,
			DoctorID:      &doctorID,
			Result:        &result,
		})

		assert.Equal(t, checkupTypeID, fields["checkup_type_id"])
		assert.Equal(t, petID, fields["pet_id"])
		assert.Equal(t, date, fields["date"])
		assert.Equal(t, nextDate, fields["next_date"])
		assert.Equal(t, doctorID, fields["doctor_id"])
		assert.Equal(t, result, fields["result"])
	})

	t.Run("DoctorIDClear=false かつ DoctorID 指定時は DoctorID が採用される", func(t *testing.T) {
		falseVal := false
		doctorID := uint64(7)
		fields := buildCheckupUpdate(&UpdateCheckupInput{DoctorIDClear: &falseVal, DoctorID: &doctorID})
		assert.Equal(t, doctorID, fields["doctor_id"])
	})
}

func TestBuildCheckupUpdate_DoctorIDClear(t *testing.T) {
	trueVal := true

	t.Run("DoctorIDClear=true sets doctor_id to nil in fields", func(t *testing.T) {
		fields := buildCheckupUpdate(&UpdateCheckupInput{DoctorIDClear: &trueVal})
		doctorID, ok := fields["doctor_id"]
		assert.True(t, ok, "doctor_id key must be present when DoctorIDClear=true")
		assert.Nil(t, doctorID, "doctor_id value must be nil when DoctorIDClear=true")
	})

	t.Run("DoctorIDClear=true takes precedence over DoctorID value", func(t *testing.T) {
		id := uint64(10)
		fields := buildCheckupUpdate(&UpdateCheckupInput{DoctorIDClear: &trueVal, DoctorID: &id})
		doctorID, ok := fields["doctor_id"]
		assert.True(t, ok)
		assert.Nil(t, doctorID, "DoctorIDClear must win over DoctorID")
	})
}

func TestCheckupService_Create_FinalizedRejection(t *testing.T) {
	now := time.Now()
	repo := &mockCheckupRepository{}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	_, err := svc.Create(context.Background(), 1, &CreateCheckupInput{
		ClinicID: 1, CheckupTypeID: 1, Date: now,
	})
	assert.Error(t, err)
}

func TestCheckupService_Update_FinalizedRejection(t *testing.T) {
	newResult := "updated"
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	_, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCheckupInput{Result: &newResult})
	assert.Error(t, err)
}

func TestCheckupService_Delete_FinalizedRejection(t *testing.T) {
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	err := svc.Delete(context.Background(), 1, 1, 1)
	assert.Error(t, err)
}

// mockLstepDeliveryTriggerForCheckup は checkup_service の Create フォローアップトリガー
// （fire-and-forget goroutine）を検証するための最小限モック。
type mockLstepDeliveryTriggerForCheckup struct {
	triggerCheckupFollowUpFn func(ctx context.Context, clinicID, ownerID uint64) error
}

func (m *mockLstepDeliveryTriggerForCheckup) TriggerFirstVisitWelcome(_ context.Context, _, _ uint64) error {
	return nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error {
	if m.triggerCheckupFollowUpFn != nil {
		return m.triggerCheckupFollowUpFn(ctx, clinicID, ownerID)
	}
	return nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerFirstVisitFollowUp3D(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerFirstVisitFollowUp7D(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerNextVisitReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerVaccineDeadline60(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerVaccineDeadline30(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerBirthdayMessage(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerDormantPrevention180(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerDormantPrevention210(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerDormantPrevention240(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerDormantPrevention365(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerFilariaAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerFleaTickAlert(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}
func (m *mockLstepDeliveryTriggerForCheckup) TriggerFoodRefillReminder(_ context.Context, _ uint64, _ time.Time) (int, []error) {
	return 0, nil
}

// TestCheckupService_Create_TriggersCheckupFollowUpBestEffort は健診記録作成後に
// fire-and-forget goroutine で TriggerCheckupFollowUp が呼ばれることを検証する（非致命的トリガー）。
func TestCheckupService_Create_TriggersCheckupFollowUpBestEffort(t *testing.T) {
	ownerID := uint64(42)
	triggeredCh := make(chan struct{}, 1)
	var triggeredClinicID, triggeredOwnerID uint64

	repo := &mockCheckupRepository{
		createFn: func(_ context.Context, checkup *model.Checkup) error {
			checkup.ID = 5
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{
				ID:              5,
				MedicalRecordID: 2,
				MedicalRecord:   &model.MedicalRecord{OwnerID: &ownerID},
			}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	trigger := &mockLstepDeliveryTriggerForCheckup{
		triggerCheckupFollowUpFn: func(_ context.Context, clinicID, ownerID uint64) error {
			triggeredClinicID = clinicID
			triggeredOwnerID = ownerID
			triggeredCh <- struct{}{}
			return errors.New("trigger failed") // 非致命的: Create は成功する
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), trigger, nil)

	result, err := svc.Create(context.Background(), 2, &CreateCheckupInput{
		ClinicID:      3,
		CheckupTypeID: 1,
		Date:          time.Now(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	select {
	case <-triggeredCh:
		assert.Equal(t, uint64(3), triggeredClinicID)
		assert.Equal(t, ownerID, triggeredOwnerID)
	case <-time.After(2 * time.Second):
		t.Fatal("TriggerCheckupFollowUp was not called within timeout")
	}
}

// TestCheckupService_Create_CheckupTypeIDZero_SkipsOwnershipCheck は
// CheckupTypeID=0（未指定）の場合、checkup_type の所有権検証をスキップして作成が成功することを検証する。
func TestCheckupService_Create_CheckupTypeIDZero_SkipsOwnershipCheck(t *testing.T) {
	checkupTypeChecked := false
	repo := &mockCheckupRepository{
		createFn: func(_ context.Context, checkup *model.Checkup) error {
			checkup.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	checkupTypeRepo := &mockCheckupTypeRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.CheckupType, error) {
			checkupTypeChecked = true
			return &model.CheckupType{}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, checkupTypeRepo, nil, nil)

	_, err := svc.Create(context.Background(), 1, &CreateCheckupInput{
		ClinicID:      1,
		CheckupTypeID: 0,
		Date:          time.Now(),
	})

	assert.NoError(t, err)
	assert.False(t, checkupTypeChecked, "CheckupTypeID=0 の場合は checkup_type 所有権検証をスキップする")
}

// TestCheckupService_Create_CheckupTypeOwnershipError はクロステナント write 防止（checkup_type
// の所有権検証失敗）が Create のエラーとして伝播することを検証する。
func TestCheckupService_Create_CheckupTypeOwnershipError(t *testing.T) {
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	checkupTypeRepo := rejectCheckupTypeRepo(99)
	svc := NewCheckupService(&mockCheckupRepository{}, mrRepo, checkupTypeRepo, nil, nil)

	result, err := svc.Create(context.Background(), 1, &CreateCheckupInput{
		ClinicID:      1,
		CheckupTypeID: 1, // 99 以外は NotFound
		Date:          time.Now(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Create_MedicalRecordLookupError は親カルテ取得失敗の伝播を検証する。
func TestCheckupService_Create_MedicalRecordLookupError(t *testing.T) {
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(&mockCheckupRepository{}, mrRepo, okCheckupTypeRepo(), nil, nil)

	result, err := svc.Create(context.Background(), 1, &CreateCheckupInput{
		ClinicID: 1, CheckupTypeID: 1, Date: time.Now(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Create_FindByIDAfterCreateError は作成直後の再取得失敗の伝播を検証する。
func TestCheckupService_Create_FindByIDAfterCreateError(t *testing.T) {
	repo := &mockCheckupRepository{
		createFn: func(_ context.Context, checkup *model.Checkup) error {
			checkup.ID = 1
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return nil, errors.New("db error")
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	result, err := svc.Create(context.Background(), 1, &CreateCheckupInput{
		ClinicID: 1, CheckupTypeID: 1, Date: time.Now(),
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Update_CheckupTypeOwnershipError は貼り替え先 checkup_type の
// クロステナント所有権検証失敗が Update のエラーとして伝播することを検証する。
func TestCheckupService_Update_CheckupTypeOwnershipError(t *testing.T) {
	newTypeID := uint64(1)
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	checkupTypeRepo := rejectCheckupTypeRepo(99)
	svc := NewCheckupService(repo, mrRepo, checkupTypeRepo, nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCheckupInput{CheckupTypeID: &newTypeID})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Update_MedicalRecordLookupError は親カルテ取得失敗の伝播を検証する。
func TestCheckupService_Update_MedicalRecordLookupError(t *testing.T) {
	newResult := "updated"
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCheckupInput{Result: &newResult})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Update_FindByIDAfterUpdateError は更新後の再取得失敗の伝播を検証する。
func TestCheckupService_Update_FindByIDAfterUpdateError(t *testing.T) {
	newResult := "updated"
	findCount := 0
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			findCount++
			if findCount == 1 {
				return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
			}
			return nil, errors.New("db error")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupInput) error {
			return nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}, nil
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	result, err := svc.Update(context.Background(), 1, 1, 1, &UpdateCheckupInput{Result: &newResult})

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestCheckupService_Delete_MedicalRecordLookupError は親カルテ取得失敗の伝播を検証する。
func TestCheckupService_Delete_MedicalRecordLookupError(t *testing.T) {
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: 1}, nil
		},
	}
	mrRepo := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewCheckupService(repo, mrRepo, okCheckupTypeRepo(), nil, nil)

	err := svc.Delete(context.Background(), 1, 1, 1)

	assert.Error(t, err)
}

func TestCheckupService_Create_RejectsInvalidPatientAndDoctorRelations(t *testing.T) {
	const (
		clinicID       = uint64(1)
		medicalRecord  = uint64(10)
		recordOwnerID  = uint64(20)
		recordPetID    = uint64(30)
		otherPetID     = uint64(31)
		assignedDoctor = uint64(40)
	)

	tests := []struct {
		name      string
		petID     *uint64
		doctorID  *uint64
		configure func(repo *mockMedicalRecordRepository)
		wantErr   bool
	}{
		{
			name:  "rejects a pet outside the clinic",
			petID: ptrUint64(999),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.findPetOwnerInClinicFn = func(_ context.Context, _, _ uint64) (uint64, error) {
					return 0, apperrors.WrapNotFound("pet", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:  "rejects a different same-clinic patient",
			petID: ptrUint64(otherPetID),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.findPetOwnerInClinicFn = func(_ context.Context, _, petID uint64) (uint64, error) {
					if petID == recordPetID || petID == otherPetID {
						return recordOwnerID, nil
					}
					return 0, apperrors.WrapNotFound("pet", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:     "rejects an inactive or unassigned doctor",
			petID:    ptrUint64(recordPetID),
			doctorID: ptrUint64(999),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.findPetOwnerInClinicFn = func(_ context.Context, _, _ uint64) (uint64, error) {
					return recordOwnerID, nil
				}
				repo.assertDoctorInClinicFn = func(_ context.Context, _, _ uint64) error {
					return apperrors.WrapNotFound("staff", "scoped")
				}
			},
			wantErr: true,
		},
		{
			name:     "accepts the record patient and an active assigned doctor",
			petID:    ptrUint64(recordPetID),
			doctorID: ptrUint64(assignedDoctor),
			configure: func(repo *mockMedicalRecordRepository) {
				repo.findPetOwnerInClinicFn = func(_ context.Context, _, _ uint64) (uint64, error) {
					return recordOwnerID, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalls := 0
			txCalls := 0
			repo := &mockCheckupRepository{
				createFn: func(_ context.Context, checkup *model.Checkup) error {
					createCalls++
					checkup.ID = 1
					return nil
				},
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
					return &model.Checkup{ID: 1, MedicalRecordID: medicalRecord, PetID: tt.petID, DoctorID: tt.doctorID}, nil
				},
			}
			relations := &mockMedicalRecordRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
					return &model.MedicalRecord{
						ID:       medicalRecord,
						ClinicID: clinicID,
						OwnerID:  ptrUint64(recordOwnerID),
						PetID:    ptrUint64(recordPetID),
						Status:   model.MedicalRecordStatusDraft,
					}, nil
				},
				findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
					return recordOwnerID, nil
				},
				withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
					txCalls++
					return fn(ctx)
				},
			}
			if tt.configure != nil {
				tt.configure(relations)
			}
			svc := NewCheckupService(repo, relations, okCheckupTypeRepo(), nil, nil)

			got, err := svc.Create(context.Background(), medicalRecord, &CreateCheckupInput{
				ClinicID: clinicID, CheckupTypeID: 1, PetID: tt.petID, DoctorID: tt.doctorID, Date: time.Now(),
			})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Zero(t, createCalls)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, 1, createCalls)
			assert.Equal(t, 1, txCalls)
		})
	}
}

func TestCheckupService_Update_RevalidatesEffectiveRelations(t *testing.T) {
	const (
		clinicID      = uint64(1)
		medicalRecord = uint64(10)
		recordOwnerID = uint64(20)
		recordPetID   = uint64(30)
	)
	otherPetID := uint64(31)
	updateCalls := 0
	repo := &mockCheckupRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Checkup, error) {
			return &model.Checkup{ID: 1, MedicalRecordID: medicalRecord, PetID: ptrUint64(recordPetID)}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ UpdateCheckupInput) error {
			updateCalls++
			return nil
		},
	}
	relations := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{
				ID: medicalRecord, ClinicID: clinicID, OwnerID: ptrUint64(recordOwnerID),
				PetID: ptrUint64(recordPetID), Status: model.MedicalRecordStatusDraft,
			}, nil
		},
		findPetOwnerInClinicFn: func(_ context.Context, _, petID uint64) (uint64, error) {
			if petID == recordPetID || petID == otherPetID {
				return recordOwnerID, nil
			}
			return 0, apperrors.WrapNotFound("pet", "scoped")
		},
	}
	svc := NewCheckupService(repo, relations, okCheckupTypeRepo(), nil, nil)

	got, err := svc.Update(context.Background(), clinicID, medicalRecord, 1, &UpdateCheckupInput{PetID: &otherPetID})

	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Zero(t, updateCalls)
}
