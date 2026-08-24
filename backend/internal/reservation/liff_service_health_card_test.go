package reservation

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// newHealthCardTestService は customerRepo/vaccinationRepo のみ実データを差し込み、
// 他の依存は空モックで最小配線した liffService を返す。
func newHealthCardTestService(customerRepo *mockLiffCustomerRepository, vaccinationRepo *mockVaccinationRepository) *liffService {
	svc := NewLiffServiceWithType(
		&mockLiffSettingRepository{},
		&mockLiffTypeRepository{},
		nil,
		&mockLiffStaffRepository{},
		&mockLiffScheduleRepository{},
		&mockLiffAdminRepository{},
		customerRepo,
		&mockLiffOwnerRepository{},
		&mockTransactor{},
		&mockLiffReservationRepository{},
		&mockLiffNotifier{},
		&mockLiffUnavailableTimeRepository{},
		&mockAvailableSlotRepository{},
		&mockReservationTypeOccupationRepository{},
		&mockTrimmingCourseRepository{},
		&mockTrimmingOptionRepository{},
		&mockTrimmingDetailRepository{},
		vaccinationRepo,
		openDayHolidayFinder(),
	)
	impl, ok := svc.(*liffService)
	if !ok {
		panic("NewLiffServiceWithType did not return *liffService")
	}
	return impl
}

func TestLiffService_GetHealthCard(t *testing.T) {
	t.Run("clinic-scoped fetch returns owner name and pets with grouped vaccines", func(t *testing.T) {
		var capturedClinicID, capturedCustomerID uint64
		ownerID := uint64(50)
		speciesDog := &model.AnimalSpecies{ID: 1, Name: "犬"}
		nextDue := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
		vaccinatedAt := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				capturedClinicID = clinicID
				capturedCustomerID = id
				return &model.LineCustomer{
					ID:       id,
					ClinicID: clinicID,
					OwnerID:  &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Name: "山田太郎",
						Pets: []model.Pet{
							{
								ID:              10,
								OwnerID:         ownerID,
								Name:            "ポチ",
								Breed:           "柴犬",
								AnimalSpeciesID: 1,
								AnimalSpecies:   speciesDog,
							},
						},
					},
				}, nil
			},
		}

		petID := uint64(10)
		vaccinationRepo := &mockVaccinationRepository{
			findByOwnerFn: func(ctx context.Context, clinicID, requestedOwnerID uint64) ([]model.Vaccination, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, ownerID, requestedOwnerID)
				return []model.Vaccination{
					{
						PetID:    &petID,
						Date:     vaccinatedAt,
						NextDate: &nextDue,
						Vaccine:  &model.Vaccine{ID: 5, Name: "混合ワクチン"},
					},
				}, nil
			},
		}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(1), capturedClinicID)
		assert.Equal(t, uint64(100), capturedCustomerID)
		assert.Equal(t, "山田太郎", result.OwnerName)
		require.Len(t, result.Pets, 1)
		assert.Equal(t, uint64(10), result.Pets[0].PetID)
		assert.Equal(t, "ポチ", result.Pets[0].PetName)
		assert.Equal(t, "犬", result.Pets[0].Species)
		assert.Equal(t, "柴犬", result.Pets[0].Breed)
		require.Len(t, result.Pets[0].Vaccines, 1)
		assert.Equal(t, "混合ワクチン", result.Pets[0].Vaccines[0].VaccineName)
		assert.True(t, vaccinatedAt.Equal(result.Pets[0].Vaccines[0].VaccinatedAt))
		require.NotNil(t, result.Pets[0].Vaccines[0].NextDueAt)
		assert.True(t, nextDue.Equal(*result.Pets[0].Vaccines[0].NextDueAt))
	})

	t.Run("owner not linked returns display name fallback and empty pets without querying vaccinations", func(t *testing.T) {
		findByOwnerCalled := false
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:          id,
					ClinicID:    clinicID,
					DisplayName: "LINE表示名",
					RealName:    "",
					OwnerID:     nil,
				}, nil
			},
		}
		vaccinationRepo := &mockVaccinationRepository{
			findByOwnerFn: func(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error) {
				findByOwnerCalled = true
				return nil, nil
			},
		}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "LINE表示名", result.OwnerName)
		assert.NotNil(t, result.Pets, "Pets はnilではなく空スライスであること")
		assert.Empty(t, result.Pets)
		assert.False(t, findByOwnerCalled, "Owner未紐付の場合はvaccinationRepo.FindByOwnerを呼び出してはならない")
	})

	t.Run("owner not linked falls back to real name when display name is empty", func(t *testing.T) {
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:          id,
					ClinicID:    clinicID,
					DisplayName: "",
					RealName:    "本名太郎",
					OwnerID:     nil,
				}, nil
			},
		}
		vaccinationRepo := &mockVaccinationRepository{}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.NoError(t, err)
		assert.Equal(t, "本名太郎", result.OwnerName)
	})

	t.Run("customerRepo.FindByID error propagates wrapped", func(t *testing.T) {
		underlying := errors.New("db down")
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				return nil, underlying
			},
		}
		vaccinationRepo := &mockVaccinationRepository{}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, underlying), "エラーはラップされていても元エラーを保持していること")
		assert.NotEqual(t, underlying, err, "エラーはapperrors.Wrapでラップされていること（生のまま返さない）")
	})

	t.Run("vaccination belonging to unrelated pet_id is excluded (defensive clinic isolation)", func(t *testing.T) {
		ownerID := uint64(50)
		speciesCat := &model.AnimalSpecies{ID: 2, Name: "猫"}
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:       id,
					ClinicID: clinicID,
					OwnerID:  &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Name: "鈴木花子",
						Pets: []model.Pet{
							{ID: 20, OwnerID: ownerID, Name: "タマ", AnimalSpeciesID: 2, AnimalSpecies: speciesCat},
						},
					},
				}, nil
			},
		}

		foreignPetID := uint64(999)
		vaccinationRepo := &mockVaccinationRepository{
			findByOwnerFn: func(ctx context.Context, clinicID, requestedOwnerID uint64) ([]model.Vaccination, error) {
				return []model.Vaccination{
					{
						PetID:   &foreignPetID,
						Date:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
						Vaccine: &model.Vaccine{ID: 9, Name: "他人のペットのワクチン"},
					},
				}, nil
			},
		}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.NoError(t, err)
		require.Len(t, result.Pets, 1)
		assert.Empty(t, result.Pets[0].Vaccines, "所有ペットに属さないワクチンデータは混入してはならない")
	})

	t.Run("deceased pets are excluded from the health card", func(t *testing.T) {
		ownerID := uint64(50)
		deceasedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		customerRepo := &mockLiffCustomerRepository{
			findByIDFn: func(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error) {
				return &model.LineCustomer{
					ID:       id,
					ClinicID: clinicID,
					OwnerID:  &ownerID,
					Owner: &model.Owner{
						ID:   ownerID,
						Name: "佐藤次郎",
						Pets: []model.Pet{
							{ID: 30, OwnerID: ownerID, Name: "モモ", DeceasedAt: &deceasedAt},
							{ID: 31, OwnerID: ownerID, Name: "クロ"},
						},
					},
				}, nil
			},
		}
		vaccinationRepo := &mockVaccinationRepository{
			findByOwnerFn: func(ctx context.Context, clinicID, requestedOwnerID uint64) ([]model.Vaccination, error) {
				return nil, nil
			},
		}

		svc := newHealthCardTestService(customerRepo, vaccinationRepo)

		result, err := svc.GetHealthCard(context.Background(), 1, 100)

		require.NoError(t, err)
		require.Len(t, result.Pets, 1)
		assert.Equal(t, uint64(31), result.Pets[0].PetID)
	})
}
