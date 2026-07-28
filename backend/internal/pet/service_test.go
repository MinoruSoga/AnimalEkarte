package pet

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type postUpdatePetOwnerReader struct {
	delegate PetOwnerReader
	calls    int
}

func (r *postUpdatePetOwnerReader) FindByPetID(
	ctx context.Context,
	clinicID, petID uint64,
) ([]model.PetOwner, error) {
	r.calls++
	if r.calls == 1 {
		return []model.PetOwner{}, nil
	}
	return r.delegate.FindByPetID(ctx, clinicID, petID)
}

func TestPetService_List(t *testing.T) {
	tests := []struct {
		name      string
		clinicID  uint64
		ownerID   *uint64
		page      int
		limit     int
		search    string
		repoPets  []model.Pet
		repoTotal int64
		repoErr   error
		wantLen   int
		wantTotal int64
		wantErr   bool
	}{
		{
			name:     "returns all pets for clinic",
			clinicID: 1,
			ownerID:  nil,
			page:     1,
			limit:    20,
			search:   "",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
				{ID: 2, ClinicID: 1, OwnerID: 11, Name: "タマ"},
			},
			repoTotal: 2,
			repoErr:   nil,
			wantLen:   2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:     "filters pets by owner_id",
			clinicID: 1,
			ownerID:  ptrUint64(10),
			page:     1,
			limit:    20,
			search:   "",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "returns empty list when no pets exist",
			clinicID:  1,
			ownerID:   nil,
			page:      1,
			limit:     20,
			search:    "",
			repoPets:  []model.Pet{},
			repoTotal: 0,
			repoErr:   nil,
			wantLen:   0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:     "filters by search keyword",
			clinicID: 1,
			ownerID:  nil,
			page:     1,
			limit:    20,
			search:   "ポチ",
			repoPets: []model.Pet{
				{ID: 1, ClinicID: 1, OwnerID: 10, Name: "ポチ"},
			},
			repoTotal: 1,
			repoErr:   nil,
			wantLen:   1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:      "propagates repository error",
			clinicID:  1,
			ownerID:   nil,
			page:      1,
			limit:     20,
			search:    "",
			repoPets:  nil,
			repoTotal: 0,
			repoErr:   errors.New("db connection error"),
			wantLen:   0,
			wantTotal: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedOwnerID := (*uint64)(nil)
			repo := &mockPetRepository{
				findAllFn: func(_ context.Context, _ []uint64, filters PetListFilters, _, _ int) ([]model.Pet, int64, error) {
					capturedOwnerID = filters.OwnerID
					return tt.repoPets, tt.repoTotal, tt.repoErr
				},
			}
			svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(tt.clinicID), defaultMedicalRecordRepo())

			pets, total, err := svc.List(context.Background(), []uint64{tt.clinicID}, PetListFilters{OwnerID: tt.ownerID, Search: tt.search}, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, pets, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
				assert.Equal(t, tt.ownerID, capturedOwnerID)
			}
		})
	}
}

func TestPetService_Update_RejectsSubOwnerPromotion(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	const clinicID = uint64(1)
	originalOwner := makeTestOwner(t, db, clinicID, "昇格前主飼主")
	promotedOwner := makeTestOwner(t, db, clinicID, "昇格対象副飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, originalOwner.ID, "昇格拒否ペット")
	makePetOwnerLink(t, db, clinicID, pet.ID, promotedOwner.ID, "家族")
	svc := NewServiceWithPetOwnerReader(
		NewRepository(db),
		petOwnerServiceDBOwnerFinder{db: db},
		nil,
		defaultMedicalRecordRepo(),
		nil,
		NewPetOwnerRepository(db),
		persistence.NewTransactor(db),
	)

	_, err := svc.Update(context.Background(), clinicID, pet.ID, &UpdatePetInput{
		OwnerID: &promotedOwner.ID,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	var persisted model.Pet
	require.NoError(t, db.Where("clinic_id = ? AND id = ?", clinicID, pet.ID).First(&persisted).Error)
	assert.Equal(t, originalOwner.ID, persisted.OwnerID)
}

func TestPetService_Update_RollsBackWhenSubOwnerAppearsAfterPrecheck(t *testing.T) {
	db := setupPetOwnerRepositoryTestDB(t)
	const clinicID = uint64(1)
	originalOwner := makeTestOwner(t, db, clinicID, "競合前主飼主")
	promotedOwner := makeTestOwner(t, db, clinicID, "競合副飼主")
	pet := makeSpeciesAndPet(t, db, clinicID, originalOwner.ID, "競合rollbackペット")
	makePetOwnerLink(t, db, clinicID, pet.ID, promotedOwner.ID, "家族")
	petOwners := &postUpdatePetOwnerReader{delegate: NewPetOwnerRepository(db)}
	svc := NewServiceWithPetOwnerReader(
		NewRepository(db),
		petOwnerServiceDBOwnerFinder{db: db},
		nil,
		defaultMedicalRecordRepo(),
		nil,
		petOwners,
		persistence.NewTransactor(db),
	)

	_, err := svc.Update(context.Background(), clinicID, pet.ID, &UpdatePetInput{
		OwnerID: &promotedOwner.ID,
	})

	require.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Equal(t, 2, petOwners.calls)
	var persisted model.Pet
	require.NoError(t, db.Where("clinic_id = ? AND id = ?", clinicID, pet.ID).First(&persisted).Error)
	assert.Equal(t, originalOwner.ID, persisted.OwnerID)
}

func TestPetService_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		clinicID uint64
		id       uint64
		repoPet  *model.Pet
		repoErr  error
		wantPet  *model.Pet
		wantErr  error
	}{
		{
			name:     "returns pet when found",
			clinicID: 1,
			id:       10,
			repoPet:  &model.Pet{ID: 10, ClinicID: 1, OwnerID: 5, Name: "ポチ"},
			repoErr:  nil,
			wantPet:  &model.Pet{ID: 10, ClinicID: 1, OwnerID: 5, Name: "ポチ"},
			wantErr:  nil,
		},
		{
			name:     "returns not found error when pet does not exist",
			clinicID: 1,
			id:       999,
			repoPet:  nil,
			repoErr:  apperrors.WrapNotFound("pet", "999"),
			wantPet:  nil,
			wantErr:  apperrors.ErrNotFound,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       10,
			repoPet:  nil,
			repoErr:  errors.New("db error"),
			wantPet:  nil,
			wantErr:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPetRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					return tt.repoPet, tt.repoErr
				},
			}
			svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(tt.clinicID), defaultMedicalRecordRepo())

			pet, err := svc.GetByID(context.Background(), tt.clinicID, tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, apperrors.ErrNotFound) {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPet, pet)
			}
		})
	}
}

func TestPetService_GetByID_NotFound(t *testing.T) {
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return nil, apperrors.WrapNotFound("pet", "999")
		},
	}
	svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo())

	pet, err := svc.GetByID(context.Background(), 1, 999)

	assert.Nil(t, pet)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

// TestPetService_GetByIDForClinics は #86 拠点横断詳細画面取得の委譲を検証する。
func TestPetService_GetByIDForClinics(t *testing.T) {
	tests := []struct {
		name      string
		clinicIDs []uint64
		id        uint64
		repoPet   *model.Pet
		repoErr   error
		wantErr   bool
		wantNF    bool
	}{
		{
			name:      "returns pet when found in any of the clinics",
			clinicIDs: []uint64{1, 2},
			id:        10,
			repoPet:   &model.Pet{ID: 10, ClinicID: 2, OwnerID: 5, Name: "ポチ"},
			wantErr:   false,
		},
		{
			name:      "returns not found error when pet is not in any of the clinics",
			clinicIDs: []uint64{1, 2},
			id:        999,
			repoErr:   apperrors.WrapNotFound("pet", "999"),
			wantErr:   true,
			wantNF:    true,
		},
		{
			name:      "returns error on repository failure",
			clinicIDs: []uint64{1},
			id:        10,
			repoErr:   errors.New("db error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedClinicIDs []uint64
			repo := &mockPetRepository{
				findByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, _ uint64) (*model.Pet, error) {
					capturedClinicIDs = clinicIDs
					return tt.repoPet, tt.repoErr
				},
			}
			svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo())

			pet, err := svc.GetByIDForClinics(context.Background(), tt.clinicIDs, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pet)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.repoPet, pet)
				assert.Equal(t, tt.clinicIDs, capturedClinicIDs)
			}
		})
	}
}

func TestPetService_Create(t *testing.T) {
	tests := []struct {
		name             string
		clinicID         uint64
		input            CreatePetInput
		repoErr          error
		ownerRepoErr     error
		insuranceRepoErr error
		wantErr          bool
		wantPet          bool
	}{
		{
			name:     "creates pet successfully",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "新しいペット",
				Gender:          "male",
			},
			repoErr: nil,
			wantErr: false,
			wantPet: true,
		},
		{
			name:     "rejects invalid gender",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "ペット",
				Gender:          "invalid",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "rejects invalid status",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "ペット",
				Status:          "invalid",
			},
			repoErr: nil,
			wantErr: true,
		},
		{
			name:     "rejects owner not in clinic",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         999,
				AnimalSpeciesID: 1,
				Name:            "ペット",
			},
			ownerRepoErr: apperrors.WrapNotFound("owner", "999"),
			wantErr:      true,
		},
		{
			name:     "returns already exists error on duplicate",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "既存ペット",
			},
			repoErr: apperrors.WrapAlreadyExists("pet", "既存ペット"),
			wantErr: true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "エラーペット",
			},
			repoErr: errors.New("db connection error"),
			wantErr: true,
		},
		{
			name:     "rejects insurance not in clinic",
			clinicID: 1,
			input: CreatePetInput{
				OwnerID:         5,
				AnimalSpeciesID: 1,
				Name:            "ペット",
				InsuranceID:     ptrUint64(99),
			},
			insuranceRepoErr: apperrors.WrapNotFound("insurance", "99"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPetRepository{
				createFn: func(_ context.Context, _ *model.Pet) error {
					return tt.repoErr
				},
			}
			ownerRepo := &mockOwnerRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
					if tt.ownerRepoErr != nil {
						return nil, tt.ownerRepoErr
					}
					return &model.Owner{ID: 5}, nil
				},
			}
			insuranceRepo := defaultInsuranceRepo(tt.clinicID)
			if tt.insuranceRepoErr != nil {
				insuranceRepo = &mockInsuranceRepository{
					findByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
						return nil, tt.insuranceRepoErr
					},
				}
			}
			svc := newPetSvc(repo, ownerRepo, insuranceRepo, defaultMedicalRecordRepo())

			pet, err := svc.Create(context.Background(), tt.clinicID, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pet)
			} else {
				assert.NoError(t, err)
				if tt.wantPet {
					assert.NotNil(t, pet)
					assert.Equal(t, tt.clinicID, pet.ClinicID)
					assert.Equal(t, tt.input.Name, pet.Name)
					assert.Equal(t, tt.input.OwnerID, pet.OwnerID)
				}
			}
		})
	}
}

func TestPetService_Create_DelegatesNumberAllocationToPetWriteOwner(t *testing.T) {
	countCalled := false
	repo := &mockPetRepository{
		countByOwnerFn: func(_ context.Context, _, _ uint64) (int64, error) {
			countCalled = true
			return 0, errors.New("legacy count-before-create must not be called")
		},
		createFn: func(_ context.Context, pet *model.Pet) error {
			pet.ID = 10
			pet.PetNumber = "5-1"
			return nil
		},
	}
	svc := newPetSvc(
		repo,
		defaultOwnerRepo(),
		defaultInsuranceRepo(1),
		defaultMedicalRecordRepo(),
	)

	created, err := svc.Create(context.Background(), 1, &CreatePetInput{
		OwnerID:         5,
		AnimalSpeciesID: 1,
		Name:            "採番委譲ペット",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.False(t, countCalled)
	assert.Equal(t, "5-1", created.PetNumber)
}

func TestPetService_Create_SyncLstepTagsBestEffort(t *testing.T) {
	animalSyncCalled := false
	basicSyncCalled := false
	repo := &mockPetRepository{
		createFn: func(_ context.Context, pet *model.Pet) error {
			pet.ID = 10
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			animalSyncCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), ownerID)
			return errors.New("classification sync failed")
		},
		syncPetBasicInfoTagsFn: func(_ context.Context, clinicID, ownerID uint64) error {
			basicSyncCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), ownerID)
			return errors.New("basic info sync failed")
		},
	}
	svc := NewService(repo, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo(), tagSync)

	pet, err := svc.Create(context.Background(), 1, &CreatePetInput{
		OwnerID:         5,
		AnimalSpeciesID: 1,
		Name:            "ポチ",
	})

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.True(t, animalSyncCalled)
	assert.True(t, basicSyncCalled)
}

func TestPetService_Update(t *testing.T) {
	updatedPet := &model.Pet{ID: 1, ClinicID: 1, Name: "更新後ペット名"}

	tests := []struct {
		name       string
		clinicID   uint64
		id         uint64
		input      UpdatePetInput
		updateErr  error
		findByIDFn func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
		wantErr    bool
		wantNF     bool
		wantPet    bool
	}{
		{
			name:     "updates pet successfully",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name:   ptrString("更新後ペット名"),
				Gender: ptrString("female"),
			},
			updateErr: nil,
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return updatedPet, nil
			},
			wantErr: false,
			wantPet: true,
		},
		{
			name:     "rejects invalid gender",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name:   ptrString("ペット"),
				Gender: ptrString("invalid"),
			},
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:      "returns error when no fields provided",
			clinicID:  1,
			id:        1,
			input:     UpdatePetInput{}, // 全 nil
			updateErr: nil,
			wantErr:   true,
		},
		{
			name:     "returns not found error when pet does not exist",
			clinicID: 1,
			id:       999,
			input: UpdatePetInput{
				Name: ptrString("存在しないペット"),
			},
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
				return nil, apperrors.WrapNotFound("pet", "999")
			},
			updateErr: apperrors.WrapNotFound("pet", "999"),
			wantErr:   true,
			wantNF:    true,
		},
		{
			name:     "returns error on repository failure",
			clinicID: 1,
			id:       1,
			input: UpdatePetInput{
				Name: ptrString("エラーケース"),
			},
			updateErr: errors.New("db error"),
			wantErr:   true,
			wantNF:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findByIDFn := tt.findByIDFn
			if findByIDFn == nil {
				findByIDFn = func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					return nil, errors.New("findByID should not be called")
				}
			}
			repo := &mockPetRepository{
				updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
					return tt.updateErr
				},
				findByIDFn: findByIDFn,
			}
			svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(tt.clinicID), defaultMedicalRecordRepo())

			pet, err := svc.Update(context.Background(), tt.clinicID, tt.id, &tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pet)
				if tt.wantNF {
					assert.True(t, apperrors.IsNotFound(err))
				}
			} else {
				assert.NoError(t, err)
				if tt.wantPet {
					assert.NotNil(t, pet)
				}
			}
		})
	}
}

// TestPetService_Update_OwnerValidation は owner_id 変更時の clinic 所属確認の失敗パスを検証する。
func TestPetService_Update_OwnerValidation(t *testing.T) {
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, ClinicID: 1}, nil
		},
	}
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "999")
		},
	}
	svc := newPetSvc(repo, ownerRepo, defaultInsuranceRepo(1), defaultMedicalRecordRepo())

	pet, err := svc.Update(context.Background(), 1, 1, &UpdatePetInput{OwnerID: ptrUint64(999)})

	assert.Error(t, err)
	assert.Nil(t, pet)
}

// TestPetService_Update_InsuranceValidation は insurance_id 変更時の clinic 所属確認の失敗パスを検証する。
func TestPetService_Update_InsuranceValidation(t *testing.T) {
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 1, ClinicID: 1}, nil
		},
	}
	insuranceRepo := &mockInsuranceRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Insurance, error) {
			return nil, apperrors.WrapNotFound("insurance", "999")
		},
	}
	invalidInsuranceID := ptrUint64(999)
	svc := newPetSvc(repo, defaultOwnerRepo(), insuranceRepo, defaultMedicalRecordRepo())

	pet, err := svc.Update(context.Background(), 1, 1, &UpdatePetInput{InsuranceID: &invalidInsuranceID})

	assert.Error(t, err)
	assert.Nil(t, pet)
}

// TestPetService_Update_RefetchError は更新後の再取得（2回目の FindByID）失敗パスを検証する。
func TestPetService_Update_RefetchError(t *testing.T) {
	callCount := 0
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			callCount++
			if callCount == 1 {
				return &model.Pet{ID: 1, ClinicID: 1}, nil
			}
			return nil, errors.New("db error on refetch")
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo())

	pet, err := svc.Update(context.Background(), 1, 1, &UpdatePetInput{Name: ptrString("更新後")})

	assert.Error(t, err)
	assert.Nil(t, pet)
	assert.Equal(t, 2, callCount)
}

func TestPetService_Update_SyncLstepTagsBestEffort(t *testing.T) {
	animalSyncCalled := false
	basicSyncCalled := false
	repo := &mockPetRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
			return &model.Pet{ID: 10, ClinicID: 1, OwnerID: 5, Name: "ポチ"}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			return nil
		},
	}
	tagSync := &mockLstepTagSyncService{
		syncOwnerAnimalClassificationTagFn: func(_ context.Context, clinicID, ownerID uint64) error {
			animalSyncCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), ownerID)
			return errors.New("classification sync failed")
		},
		syncPetBasicInfoTagsFn: func(_ context.Context, clinicID, ownerID uint64) error {
			basicSyncCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(5), ownerID)
			return errors.New("basic info sync failed")
		},
	}
	svc := NewService(repo, defaultOwnerRepo(), defaultInsuranceRepo(1), defaultMedicalRecordRepo(), tagSync)

	pet, err := svc.Update(context.Background(), 1, 10, &UpdatePetInput{
		Name: ptrString("ポチ 更新"),
	})

	assert.NoError(t, err)
	assert.NotNil(t, pet)
	assert.True(t, animalSyncCalled)
	assert.True(t, basicSyncCalled)
}

func TestPetService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		clinicID      uint64
		id            uint64
		findByIDErr   error
		recordCount   int64
		countByPetErr error
		repoErr       error
		wantErr       bool
		wantNF        bool
		wantConflict  bool
	}{
		{
			name:         "deletes pet successfully",
			clinicID:     1,
			id:           10,
			recordCount:  0,
			repoErr:      nil,
			wantErr:      false,
			wantNF:       false,
			wantConflict: false,
		},
		{
			name:         "returns not found error when pet does not exist",
			clinicID:     1,
			id:           999,
			findByIDErr:  apperrors.WrapNotFound("pet", "999"),
			recordCount:  0,
			wantErr:      true,
			wantNF:       true,
			wantConflict: false,
		},
		{
			name:         "returns error on repository failure",
			clinicID:     1,
			id:           10,
			recordCount:  0,
			repoErr:      errors.New("db error"),
			wantErr:      true,
			wantNF:       false,
			wantConflict: false,
		},
		{
			name:         "returns conflict error when pet has medical records",
			clinicID:     1,
			id:           10,
			recordCount:  3,
			repoErr:      nil,
			wantErr:      true,
			wantNF:       false,
			wantConflict: true,
		},
		{
			name:          "returns error when medical record count check fails",
			clinicID:      1,
			id:            10,
			countByPetErr: errors.New("db error"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			medicalRecordRepo := &mockMedicalRecordRepository{
				countByPetIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
					if tt.countByPetErr != nil {
						return 0, tt.countByPetErr
					}
					return tt.recordCount, nil
				},
			}
			repo := &mockPetRepository{
				findByIDFn: func(_ context.Context, _, _ uint64) (*model.Pet, error) {
					if tt.findByIDErr != nil {
						return nil, tt.findByIDErr
					}
					return &model.Pet{ID: tt.id, ClinicID: tt.clinicID}, nil
				},
				deleteFn: func(_ context.Context, _, _ uint64) error {
					return tt.repoErr
				},
			}
			svc := newPetSvc(repo, defaultOwnerRepo(), defaultInsuranceRepo(tt.clinicID), medicalRecordRepo)

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
