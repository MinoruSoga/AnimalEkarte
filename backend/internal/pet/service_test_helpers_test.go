package pet

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func ptrString(value string) *string {
	return &value
}

// mockPetRepository は Repository のテスト用モック実装。
type mockPetRepository struct {
	findAllFn                     func(ctx context.Context, clinicIDs []uint64, filters PetListFilters, page, limit int) ([]model.Pet, int64, error)
	findByIDFn                    func(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
	findByIDForClinicsFn          func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Pet, error)
	countByOwnerFn                func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	countLivingByOwnerFn          func(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	countLivingByOwnerIDsFn       func(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error)
	countUsageByAnimalSpeciesIDFn func(ctx context.Context, speciesID uint64) (int64, error)
	createFn                      func(ctx context.Context, pet *model.Pet) error
	updateFn                      func(ctx context.Context, clinicID, id uint64, fields map[string]any) error
	deleteFn                      func(ctx context.Context, clinicID, id uint64) error
	findLivingByOwnerFn           func(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
}

func (m *mockPetRepository) FindAll(
	ctx context.Context,
	clinicIDs []uint64,
	filters PetListFilters,
	page, limit int,
) ([]model.Pet, int64, error) {
	return m.findAllFn(ctx, clinicIDs, filters, page, limit)
}

func (m *mockPetRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error) {
	if m.findByIDFn == nil {
		return nil, nil
	}
	return m.findByIDFn(ctx, clinicID, id)
}

func (m *mockPetRepository) FindByIDForClinics(
	ctx context.Context,
	clinicIDs []uint64,
	id uint64,
) (*model.Pet, error) {
	if m.findByIDForClinicsFn == nil {
		return nil, nil
	}
	return m.findByIDForClinicsFn(ctx, clinicIDs, id)
}

func (m *mockPetRepository) CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.countByOwnerFn == nil {
		return 0, nil
	}
	return m.countByOwnerFn(ctx, clinicID, ownerID)
}

func (m *mockPetRepository) CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error) {
	if m.countLivingByOwnerFn == nil {
		return 0, nil
	}
	return m.countLivingByOwnerFn(ctx, clinicID, ownerID)
}

func (m *mockPetRepository) CountLivingByOwnerIDs(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
) (map[uint64]int64, error) {
	if m.countLivingByOwnerIDsFn == nil {
		return map[uint64]int64{}, nil
	}
	return m.countLivingByOwnerIDsFn(ctx, clinicID, ownerIDs)
}

func (m *mockPetRepository) CountUsageByAnimalSpeciesID(
	ctx context.Context,
	speciesID uint64,
) (int64, error) {
	if m.countUsageByAnimalSpeciesIDFn == nil {
		return 0, nil
	}
	return m.countUsageByAnimalSpeciesIDFn(ctx, speciesID)
}

func (m *mockPetRepository) Create(ctx context.Context, pet *model.Pet) error {
	return m.createFn(ctx, pet)
}

func (m *mockPetRepository) Update(
	ctx context.Context,
	clinicID, id uint64,
	cmd UpdatePetInput,
) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, clinicID, id, buildPetUpdate(&cmd))
	}
	return nil
}

func (m *mockPetRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockPetRepository) FindLivingByOwner(
	ctx context.Context,
	clinicID, ownerID uint64,
) ([]model.Pet, error) {
	if m.findLivingByOwnerFn == nil {
		return nil, nil
	}
	return m.findLivingByOwnerFn(ctx, clinicID, ownerID)
}

func (*mockPetRepository) FindOwnersByPetBirthday(
	context.Context,
	uint64,
	int,
	int,
) ([]uint64, error) {
	return nil, nil
}

func defaultOwnerRepo() *mockOwnerRepository {
	return &mockOwnerRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Owner, error) {
			return &model.Owner{ID: 5}, nil
		},
	}
}

func defaultInsuranceRepo(clinicID uint64) *mockInsuranceRepository {
	return &mockInsuranceRepository{
		findByIDFn: func(context.Context, uint64, uint64) (*model.Insurance, error) {
			return &model.Insurance{ID: 1, ClinicID: clinicID}, nil
		},
	}
}

func defaultMedicalRecordRepo() *mockMedicalRecordRepository {
	return &mockMedicalRecordRepository{
		countByPetIDFn: func(context.Context, uint64, uint64) (int64, error) {
			return 0, nil
		},
	}
}

func ptrUint64(value uint64) *uint64 {
	return &value
}

func newPetSvc(
	repo *mockPetRepository,
	ownerRepo *mockOwnerRepository,
	insuranceRepo *mockInsuranceRepository,
	medicalRecordRepo *mockMedicalRecordRepository,
) Service {
	return NewService(repo, ownerRepo, insuranceRepo, medicalRecordRepo, nil)
}

type mockOwnerRepository struct {
	findByIDFn func(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

func (m *mockOwnerRepository) FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	if m.findByIDFn == nil {
		return nil, nil
	}
	return m.findByIDFn(ctx, clinicID, ownerID)
}

type mockInsuranceRepository struct {
	findByIDFn func(ctx context.Context, clinicID, insuranceID uint64) (*model.Insurance, error)
}

func (m *mockInsuranceRepository) FindByID(ctx context.Context, clinicID, insuranceID uint64) (*model.Insurance, error) {
	if m.findByIDFn == nil {
		return nil, nil
	}
	return m.findByIDFn(ctx, clinicID, insuranceID)
}

type mockMedicalRecordRepository struct {
	findFirstVisitDateByPetIDFn func(ctx context.Context, clinicID, petID uint64) (*time.Time, error)
	countByPetIDFn              func(ctx context.Context, clinicID, petID uint64) (int64, error)
}

func (m *mockMedicalRecordRepository) FindFirstVisitDateByPetID(
	ctx context.Context,
	clinicID uint64,
	petID uint64,
) (*time.Time, error) {
	if m.findFirstVisitDateByPetIDFn == nil {
		return nil, nil
	}
	return m.findFirstVisitDateByPetIDFn(ctx, clinicID, petID)
}

func (m *mockMedicalRecordRepository) CountByPetID(
	ctx context.Context,
	clinicID uint64,
	petID uint64,
) (int64, error) {
	if m.countByPetIDFn == nil {
		return 0, nil
	}
	return m.countByPetIDFn(ctx, clinicID, petID)
}

type mockLstepTagSyncService struct {
	syncOwnerAnimalClassificationTagFn func(ctx context.Context, clinicID, ownerID uint64) error
	syncPetBasicInfoTagsFn             func(ctx context.Context, clinicID, ownerID uint64) error
	syncChronicConditionTagsFn         func(ctx context.Context, clinicID, ownerID uint64, codes []string) error
}

func (m *mockLstepTagSyncService) SyncOwnerAnimalClassificationTags(
	ctx context.Context,
	clinicID uint64,
	ownerID uint64,
) error {
	if m.syncOwnerAnimalClassificationTagFn == nil {
		return nil
	}
	return m.syncOwnerAnimalClassificationTagFn(ctx, clinicID, ownerID)
}

func (m *mockLstepTagSyncService) SyncPetBasicInfoTags(
	ctx context.Context,
	clinicID uint64,
	ownerID uint64,
) error {
	if m.syncPetBasicInfoTagsFn == nil {
		return nil
	}
	return m.syncPetBasicInfoTagsFn(ctx, clinicID, ownerID)
}

func (m *mockLstepTagSyncService) SyncChronicConditionTags(
	ctx context.Context,
	clinicID uint64,
	ownerID uint64,
	codes []string,
) error {
	if m.syncChronicConditionTagsFn == nil {
		return nil
	}
	return m.syncChronicConditionTagsFn(ctx, clinicID, ownerID, codes)
}
