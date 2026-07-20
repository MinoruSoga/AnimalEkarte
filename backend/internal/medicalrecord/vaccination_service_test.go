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

// mockVaccinationRepository は VaccinationRepository のテスト用モック実装
type mockVaccinationRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	findByOwnerFn  func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
	createFn       func(ctx context.Context, vaccination *model.Vaccination) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
}

func (m *mockVaccinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Vaccination, int64, error) {
	return m.findAllFn(ctx, clinicID, petID, ownerID, startDate, endDate, page, limit)
}

func (m *mockVaccinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
	if m.findByOwnerFn != nil {
		return m.findByOwnerFn(ctx, clinicID, ownerID)
	}
	return nil, nil
}

func (m *mockVaccinationRepository) Create(ctx context.Context, vaccination *model.Vaccination) error {
	return m.createFn(ctx, vaccination)
}

func (m *mockVaccinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccination, error) {
	return m.updateFieldsFn(ctx, clinicID, id, fields)
}

func (m *mockVaccinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockVaccinationRepository) FindOwnersByVaccineDeadline(_ context.Context, _ uint64, _ time.Time) ([]uint64, error) {
	return nil, nil
}

func TestBuildVaccinationUpdate(t *testing.T) {
	medicalRecordID := uint64(1)
	petID := uint64(2)
	vaccineID := uint64(3)
	date := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	doctorID := uint64(4)
	nextDate := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	nextScheduleType := model.NextScheduleType("fixed")
	supplemental := "追記"
	lot1, lot2, lot3, lot4 := "L1", "L2", "L3", "L4"
	remarks := "remarks"

	tests := []struct {
		name  string
		input *UpdateVaccinationInput
		want  map[string]any
	}{
		{
			name: "all fields set",
			input: &UpdateVaccinationInput{
				MedicalRecordID: &medicalRecordID, PetID: &petID, VaccineID: &vaccineID, Date: &date,
				DoctorID: &doctorID, NextDate: &nextDate, NextScheduleType: &nextScheduleType,
				Supplemental: &supplemental, Lot1: &lot1, Lot2: &lot2, Lot3: &lot3, Lot4: &lot4, Remarks: &remarks,
			},
			want: map[string]any{
				"medical_record_id":  medicalRecordID,
				"pet_id":             petID,
				"vaccine_id":         vaccineID,
				"date":               date,
				"doctor_id":          doctorID,
				"next_date":          nextDate,
				"next_schedule_type": nextScheduleType,
				"supplemental":       supplemental,
				"lot1":               lot1,
				"lot2":               lot2,
				"lot3":               lot3,
				"lot4":               lot4,
				"remarks":            remarks,
			},
		},
		{
			name:  "no fields set returns empty map",
			input: &UpdateVaccinationInput{},
			want:  map[string]any{},
		},
		{
			name:  "only pet_id set",
			input: &UpdateVaccinationInput{PetID: &petID},
			want:  map[string]any{"pet_id": petID},
		},
		{
			name:  "only doctor_id set",
			input: &UpdateVaccinationInput{DoctorID: &doctorID},
			want:  map[string]any{"doctor_id": doctorID},
		},
		{
			name:  "only next_date set",
			input: &UpdateVaccinationInput{NextDate: &nextDate},
			want:  map[string]any{"next_date": nextDate},
		},
		{
			name:  "only next_schedule_type set",
			input: &UpdateVaccinationInput{NextScheduleType: &nextScheduleType},
			want:  map[string]any{"next_schedule_type": nextScheduleType},
		},
		{
			name:  "only lot fields set",
			input: &UpdateVaccinationInput{Lot1: &lot1, Lot3: &lot3},
			want:  map[string]any{"lot1": lot1, "lot3": lot3},
		},
		{
			name:  "only remarks set",
			input: &UpdateVaccinationInput{Remarks: &remarks},
			want:  map[string]any{"remarks": remarks},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildVaccinationUpdate(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVaccinationService_List(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name             string
		clinicID         uint64
		petID            *uint64
		ownerID          *uint64
		page             int
		limit            int
		repoVaccinations []model.Vaccination
		repoTotal        int64
		repoErr          error
		wantLen          int
		wantTotal        int64
		wantErr          bool
	}{
		{
			name:     "returns all vaccinations without filter",
			clinicID: 1,
			petID:    nil,
			ownerID:  nil,
			page:     1,
			limit:    20,
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
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
			repoVaccinations: []model.Vaccination{
				{ID: 1, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
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
			repoVaccinations: []model.Vaccination{
				{ID: 2, MedicalRecordID: ptrUint64(2), VaccineID: 2, Date: now},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:             "returns empty list when no vaccinations exist",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: []model.Vaccination{},
			repoTotal:        0,
			repoErr:          nil,
			wantLen:          0,
			wantTotal:        0,
			wantErr:          false,
		},
		{
			name:             "propagates repository error",
			clinicID:         1,
			petID:            nil,
			ownerID:          nil,
			page:             1,
			limit:            20,
			repoVaccinations: nil,
			repoTotal:        0,
			repoErr:          errors.New("db connection error"),
			wantLen:          0,
			wantTotal:        0,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedPetID := (*uint64)(nil)
			capturedOwnerID := (*uint64)(nil)
			repo := &mockVaccinationRepository{
				findAllFn: func(_ context.Context, _ uint64, petID *uint64, ownerID *uint64, _, _ *string, _, _ int) ([]model.Vaccination, int64, error) {
					capturedPetID = petID
					capturedOwnerID = ownerID
					return tt.repoVaccinations, tt.repoTotal, tt.repoErr
				},
			}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil)

			vaccinations, total, err := svc.List(context.Background(), tt.clinicID, tt.petID, tt.ownerID, nil, nil, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vaccinations, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.petID, capturedPetID)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestVaccinationService_GetByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name            string
		clinicID        uint64
		id              uint64
		repoVaccination *model.Vaccination
		repoErr         error
		wantVaccination *model.Vaccination
		wantErr         error
	}{
		{
			name:            "returns vaccination when found",
			clinicID:        1,
			id:              10,
			repoVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			repoErr:         nil,
			wantVaccination: &model.Vaccination{ID: 10, MedicalRecordID: ptrUint64(1), VaccineID: 1, Date: now},
			wantErr:         nil,
		},
		{
			name:            "returns not found error when vaccination does not exist",
			clinicID:        1,
			id:              999,
			repoVaccination: nil,
			repoErr:         apperrors.WrapNotFound("vaccination", "999"),
			wantVaccination: nil,
			wantErr:         apperrors.ErrNotFound,
		},
		{
			name:            "returns error on repository failure",
			clinicID:        1,
			id:              10,
			repoVaccination: nil,
			repoErr:         errors.New("db error"),
			wantVaccination: nil,
			wantErr:         errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
					return tt.repoVaccination, tt.repoErr
				},
			}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVaccination, vaccination)
			}
		})
	}
}

func TestVaccinationService_GetByID_NotFound(t *testing.T) {
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "999")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, vaccination)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Create(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		clinicID uint64
		input    *CreateVaccinationInput
		repoErr  error
		wantErr  bool
	}{
		{
			name:     "creates vaccination successfully",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       1,
				Date:            now,
			},
			repoErr: nil,
			wantErr: false,
		},
		{
			name:     "returns error when vaccine_id is zero",
			clinicID: 1,
			input:    &CreateVaccinationInput{VaccineID: 0, Date: now},
			repoErr:  nil,
			wantErr:  true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: &CreateVaccinationInput{
				MedicalRecordID: ptrUint64(1),
				VaccineID:       2,
				Date:            now,
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				createFn: func(_ context.Context, vaccination *model.Vaccination) error {
					vaccination.ID = 10
					return tt.repoErr
				},
				findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
					return &model.Vaccination{ID: id, MedicalRecordID: tt.input.MedicalRecordID, VaccineID: tt.input.VaccineID, Date: tt.input.Date}, nil
				},
			}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.Create(context.Background(), tt.clinicID, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Create_PostCreateFindByIDError(t *testing.T) {
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 10
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{VaccineID: 1, Date: time.Now()})

	assert.NoError(t, err, "post-create FindByID failure falls back to created record instead of erroring")
	assert.NotNil(t, vaccination)
	assert.Equal(t, uint64(10), vaccination.ID)
}

func TestVaccinationService_Create_PostCreateFindByIDNil(t *testing.T) {
	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 11
			return nil
		},
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, nil
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{VaccineID: 1, Date: time.Now()})

	assert.NoError(t, err)
	assert.NotNil(t, vaccination)
	assert.Equal(t, uint64(11), vaccination.ID)
}

func TestVaccinationService_Create_SyncsVaccineTagBestEffort(t *testing.T) {
	ownerID := uint64(10)
	petID := uint64(20)
	var syncedClinicID, syncedOwnerID, syncedVaccinationID uint64

	repo := &mockVaccinationRepository{
		createFn: func(_ context.Context, vaccination *model.Vaccination) error {
			vaccination.ID = 30
			return nil
		},
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:        id,
				PetID:     &petID,
				VaccineID: 3,
				Date:      time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
				Pet:       &model.Pet{OwnerID: ownerID},
			}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncVaccineTagFn: func(_ context.Context, clinicID, ownerID, vaccinationID uint64) error {
			syncedClinicID = clinicID
			syncedOwnerID = ownerID
			syncedVaccinationID = vaccinationID
			return errors.New("sync failed")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync)

	vaccination, err := svc.Create(context.Background(), 1, &CreateVaccinationInput{
		PetID:     &petID,
		VaccineID: 3,
		Date:      time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.NotNil(t, vaccination)
	assert.Equal(t, uint64(1), syncedClinicID)
	assert.Equal(t, ownerID, syncedOwnerID)
	assert.Equal(t, uint64(30), syncedVaccinationID)
}

func TestVaccinationService_Update(t *testing.T) {
	now := time.Now()
	supplemental := "追記情報"
	tests := []struct {
		name    string
		input   UpdateVaccinationInput
		repoErr error
		wantErr bool
		wantNF  bool
	}{
		{
			name: "updates vaccination successfully",
			input: UpdateVaccinationInput{
				Date:         &now,
				Supplemental: &supplemental,
			},
			repoErr: nil,
			wantErr: false,
			wantNF:  false,
		},
		{
			name:    "returns error when no fields provided",
			input:   UpdateVaccinationInput{},
			repoErr: nil,
			wantErr: true,
			wantNF:  false,
		},
		{
			name: "returns not found error when vaccination does not exist",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: apperrors.WrapNotFound("vaccination", "999"),
			wantErr: true,
			wantNF:  true,
		},
		{
			name: "returns error on repository failure",
			input: UpdateVaccinationInput{
				Supplemental: &supplemental,
			},
			repoErr: errors.New("db error"),
			wantErr: true,
			wantNF:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockVaccinationRepository{
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Vaccination, error) {
					if tt.repoErr != nil {
						return nil, tt.repoErr
					}
					return &model.Vaccination{ID: 1}, nil
				},
			}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil)

			vaccination, err := svc.Update(context.Background(), 1, 1, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, vaccination)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, vaccination)
			}
		})
	}
}

func TestVaccinationService_Update_FindByIDError(t *testing.T) {
	supplemental := "追記"
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "1")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil)

	vaccination, err := svc.Update(context.Background(), 1, 1, &UpdateVaccinationInput{Supplemental: &supplemental})

	assert.Error(t, err)
	assert.Nil(t, vaccination)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoErr  error
		wantErr  bool
		wantNF   bool
	}{
		{
			name:     "deletes vaccination successfully",
			clinicID: 1,
			id:       10,
			repoErr:  nil,
			wantErr:  false,
			wantNF:   false,
		},
		{
			name:     "returns not found error when vaccination does not exist",
			clinicID: 1,
			id:       999,
			repoErr:  apperrors.WrapNotFound("vaccination", "999"),
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
			repo := &mockVaccinationRepository{
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := NewVaccinationService(repo, okVaccineRepo(), nil)

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

func TestVaccinationService_Delete_FindByIDError(t *testing.T) {
	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Vaccination, error) {
			return nil, apperrors.WrapNotFound("vaccination", "999")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), nil)

	err := svc.Delete(context.Background(), 1, 999)

	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestVaccinationService_Update_ResyncsOwnerVaccineTags(t *testing.T) {
	ownerID := uint64(10)
	var syncedOwnerID uint64
	supplemental := "updated"

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{ID: id}, nil
		},
		updateFieldsFn: func(_ context.Context, _, id uint64, _ map[string]any) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:  id,
				Pet: &model.Pet{OwnerID: ownerID},
			}, nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, ownerID uint64) error {
			syncedOwnerID = ownerID
			return errors.New("sync failed")
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync)

	vaccination, err := svc.Update(context.Background(), 1, 30, &UpdateVaccinationInput{Supplemental: &supplemental})

	assert.NoError(t, err)
	assert.NotNil(t, vaccination)
	assert.Equal(t, ownerID, syncedOwnerID)
}

func TestVaccinationService_Delete_ResyncsOwnerVaccineTagsAfterDelete(t *testing.T) {
	ownerID := uint64(10)
	deleted := false
	syncedAfterDelete := false

	repo := &mockVaccinationRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Vaccination, error) {
			return &model.Vaccination{
				ID:  id,
				Pet: &model.Pet{OwnerID: ownerID},
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleted = true
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		resyncOwnerVaccineTagsFn: func(_ context.Context, _, syncedOwnerID uint64) error {
			syncedAfterDelete = deleted
			assert.Equal(t, ownerID, syncedOwnerID)
			return nil
		},
	}
	svc := NewVaccinationService(repo, okVaccineRepo(), tagSync)

	err := svc.Delete(context.Background(), 1, 30)

	assert.NoError(t, err)
	assert.True(t, syncedAfterDelete)
}
