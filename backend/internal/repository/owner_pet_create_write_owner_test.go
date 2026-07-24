package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
)

type ownerRegistrationWriterFunc func(
	ctx context.Context,
	intent petdomain.OwnerRegistrationIntent,
) ([]model.Pet, error)

func (f ownerRegistrationWriterFunc) CreateForOwnerRegistration(
	ctx context.Context,
	intent petdomain.OwnerRegistrationIntent,
) ([]model.Pet, error) {
	return f(ctx, intent)
}

// These tests pin the owner-registration graph contract before pets writes move
// behind the pet-owned capability. The owner and all nested pets must remain one
// atomic graph, while the returned value is the committed, reloaded graph.
func TestOwnerRepository_CreateWithPets_AssignsNumbersAndReloadsCommittedGraph(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	dog := &model.AnimalSpecies{Name: "飼主登録テスト犬"}
	cat := &model.AnimalSpecies{Name: "飼主登録テスト猫"}
	require.NoError(t, db.WithContext(ctx).Create(dog).Error)
	require.NoError(t, db.WithContext(ctx).Create(cat).Error)
	insurance := &model.Insurance{ClinicID: clinicID, Name: "飼主登録テスト保険", CoverageRate: 70}
	require.NoError(t, db.WithContext(ctx).Create(insurance).Error)
	birthDate := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC)
	neuteredDate := time.Date(2021, time.February, 3, 0, 0, 0, 0, time.UTC)
	bloodType := "DEA 1.1+"
	microchipNumber := "392140000123456"
	weight := 12.3
	acquisitionType := model.AcquisitionTypeRescued

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "飼主登録グラフ",
		Email:    "owner-registration-graph@example.com",
	}
	pets := []model.Pet{
		{
			AnimalSpeciesID: dog.ID,
			Name:            "ポチ",
			NameKana:        "ぽち",
			Gender:          model.PetGenderMale,
			Status:          model.PetStatusAlive,
			BirthDate:       &birthDate,
			Breed:           "柴犬",
			Color:           "茶",
			BloodType:       &bloodType,
			MicrochipNumber: &microchipNumber,
			Weight:          &weight,
			NeuteredDate:    &neuteredDate,
			AcquisitionType: &acquisitionType,
			DangerLevel:     model.DangerLevelMedium,
			Food:            "療法食",
			Environment:     "室内",
			Phone:           "090-0000-0000",
			InsuranceID:     &insurance.ID,
			Remarks:         "登録時備考",
		},
		{AnimalSpeciesID: cat.ID, Name: "タマ"},
	}

	require.NoError(t, NewOwnerRepository(db).CreateWithPets(ctx, owner, pets))
	require.NotZero(t, owner.ID)
	require.Len(t, owner.Pets, 2)

	byName := make(map[string]model.Pet, len(owner.Pets))
	for i := range owner.Pets {
		byName[owner.Pets[i].Name] = owner.Pets[i]
	}
	require.Contains(t, byName, "ポチ")
	require.Contains(t, byName, "タマ")
	assert.Equal(t, fmt.Sprintf("%d-1", owner.ID), byName["ポチ"].PetNumber)
	assert.Equal(t, fmt.Sprintf("%d-2", owner.ID), byName["タマ"].PetNumber)

	for _, created := range byName {
		assert.Equal(t, clinicID, created.ClinicID)
		assert.Equal(t, owner.ID, created.OwnerID)
		require.NotNil(t, created.AnimalSpecies)
	}

	// A separate pet-side reload must observe the same committed graph and
	// relations; this distinguishes post-commit reload from an in-memory echo.
	reloadedPet, err := NewPetRepository(db).FindByID(ctx, clinicID, byName["ポチ"].ID)
	require.NoError(t, err)
	require.NotNil(t, reloadedPet.Owner)
	require.NotNil(t, reloadedPet.AnimalSpecies)
	assert.Equal(t, owner.ID, reloadedPet.Owner.ID)
	assert.Equal(t, dog.ID, reloadedPet.AnimalSpecies.ID)
	assert.Equal(t, fmt.Sprintf("%d-1", owner.ID), reloadedPet.PetNumber)
	assert.Equal(t, "ぽち", reloadedPet.NameKana)
	assert.Equal(t, model.PetGenderMale, reloadedPet.Gender)
	assert.Equal(t, model.PetStatusAlive, reloadedPet.Status)
	require.NotNil(t, reloadedPet.BirthDate)
	require.NotNil(t, reloadedPet.BloodType)
	require.NotNil(t, reloadedPet.MicrochipNumber)
	require.NotNil(t, reloadedPet.Weight)
	require.NotNil(t, reloadedPet.NeuteredDate)
	require.NotNil(t, reloadedPet.AcquisitionType)
	require.NotNil(t, reloadedPet.InsuranceID)
	assert.Equal(t, birthDate, *reloadedPet.BirthDate)
	assert.Equal(t, "柴犬", reloadedPet.Breed)
	assert.Equal(t, "茶", reloadedPet.Color)
	assert.Equal(t, bloodType, *reloadedPet.BloodType)
	assert.Equal(t, microchipNumber, *reloadedPet.MicrochipNumber)
	assert.Equal(t, weight, *reloadedPet.Weight)
	assert.Equal(t, neuteredDate, *reloadedPet.NeuteredDate)
	assert.Equal(t, acquisitionType, *reloadedPet.AcquisitionType)
	assert.Equal(t, model.DangerLevelMedium, reloadedPet.DangerLevel)
	assert.Equal(t, "療法食", reloadedPet.Food)
	assert.Equal(t, "室内", reloadedPet.Environment)
	assert.Equal(t, "090-0000-0000", reloadedPet.Phone)
	assert.Equal(t, insurance.ID, *reloadedPet.InsuranceID)
	assert.Equal(t, "登録時備考", reloadedPet.Remarks)
}

func TestOwnerRepository_CreateWithPets_IgnoresPoisonedOwnerAssociations(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	species := &model.AnimalSpecies{Name: "owner association poison species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "owner association poison",
		Email:    "owner-association-poison@example.com",
		Pets: []model.Pet{{
			AnimalSpeciesID: species.ID,
			PetNumber:       "must-not-be-persisted",
			Name:            "poisoned implicit association",
		}},
	}

	require.NoError(t, NewOwnerRepository(db).CreateWithPets(ctx, owner, []model.Pet{{
		AnimalSpeciesID: species.ID,
		Name:            "explicit pet intent",
	}}))

	require.Len(t, owner.Pets, 1)
	assert.Equal(t, "explicit pet intent", owner.Pets[0].Name)
	assert.Equal(t, fmt.Sprintf("%d-1", owner.ID), owner.Pets[0].PetNumber)

	var poisonedCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "poisoned implicit association").
		Count(&poisonedCount).Error)
	assert.Zero(t, poisonedCount, "owner create must never autosave owner.Pets")
}

func TestOwnerRepository_CreateWithPets_RejectsCrossClinicInsuranceAndRollsBackGraph(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	species := &model.AnimalSpecies{Name: "保険分離テスト種"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	foreignInsurance := &model.Insurance{
		ClinicID:     clinicB,
		Name:         "医院B保険",
		CoverageRate: 50,
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignInsurance).Error)

	owner := &model.Owner{
		ClinicID: clinicA,
		Name:     "他院保険拒否飼主",
		Email:    "cross-clinic-insurance-owner@example.com",
	}
	pets := []model.Pet{{
		AnimalSpeciesID: species.ID,
		InsuranceID:     &foreignInsurance.ID,
		Name:            "他院保険拒否ペット",
	}}

	err := NewOwnerRepository(db).CreateWithPets(ctx, owner, pets)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "cross-clinic master FK must be rejected as invalid input: %v", err)

	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", "cross-clinic-insurance-owner@example.com").
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount, "owner must roll back with the rejected nested pet")

	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "他院保険拒否ペット").
		Count(&petCount).Error)
	assert.Zero(t, petCount, "nested pet must not survive the rejected graph")
}

func TestOwnerRepository_CreateWithPets_InvalidSpeciesRollsBackOwnerAndEarlierPets(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	validSpecies := &model.AnimalSpecies{Name: "原子性テスト種"}
	require.NoError(t, db.WithContext(ctx).Create(validSpecies).Error)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "原子性テスト飼主",
		Email:    "owner-registration-atomicity@example.com",
	}
	pets := []model.Pet{
		{AnimalSpeciesID: validSpecies.ID, Name: "先行ペット"},
		{AnimalSpeciesID: validSpecies.ID + 999999, Name: "不正種ペット"},
	}

	err := NewOwnerRepository(db).CreateWithPets(ctx, owner, pets)
	require.Error(t, err)

	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", "owner-registration-atomicity@example.com").
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount)

	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name IN ?", []string{"先行ペット", "不正種ペット"}).
		Count(&petCount).Error)
	assert.Zero(t, petCount, "all earlier nested writes must roll back with the graph")
}

func TestOwnerRepository_CreateWithPets_DelegatesPetWriteInsideSameTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	sentinel := errors.New("simulated pet owner failure")
	capabilityCalled := false
	writer := ownerRegistrationWriterFunc(func(
		txCtx context.Context,
		intent petdomain.OwnerRegistrationIntent,
	) ([]model.Pet, error) {
		capabilityCalled = true
		require.NotNil(t, txFromContext(txCtx), "pet capability must receive the owner transaction")
		assert.Equal(t, clinicID, intent.ClinicID)
		assert.NotZero(t, intent.OwnerID)
		require.Len(t, intent.Pets, 1)

		var ownerCount int64
		require.NoError(t, persistence.DBOrTx(txCtx, db).Model(&model.Owner{}).
			Where("id = ? AND clinic_id = ?", intent.OwnerID, intent.ClinicID).
			Count(&ownerCount).Error)
		assert.Equal(t, int64(1), ownerCount, "owner insert must be visible inside the same transaction")
		return nil, sentinel
	})

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "委譲ロールバック飼主",
		Email:    "owner-capability-rollback@example.com",
	}
	err := NewOwnerRepositoryWithPetWriter(db, writer).CreateWithPets(ctx, owner, []model.Pet{{
		AnimalSpeciesID: 1,
		Name:            "委譲ロールバックペット",
	}})
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	assert.True(t, capabilityCalled)

	var committedOwnerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", "owner-capability-rollback@example.com").
		Count(&committedOwnerCount).Error)
	assert.Zero(t, committedOwnerCount, "pet capability failure must roll back the owner insert")
}

func TestPetOwnerRegistrationWriter_RequiresAmbientTransactionAndClinicOwnerMatch(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "pet capability clinic guard")
	species := &model.AnimalSpecies{Name: "pet capability clinic guard species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	writer := petdomain.NewOwnerRegistrationWriter()
	intent := petdomain.OwnerRegistrationIntent{
		ClinicID: clinicB,
		OwnerID:  owner.ID,
		Pets: []petdomain.CreatePetDraft{{
			AnimalSpeciesID: species.ID,
			Name:            "cross-clinic capability pet",
		}},
	}

	t.Run("ambient transaction is mandatory", func(t *testing.T) {
		created, err := writer.CreateForOwnerRegistration(ctx, intent)
		assert.Nil(t, created)
		require.Error(t, err)
		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr))
		assert.Equal(t, "INTERNAL", appErr.Code, "missing ambient transaction must fail closed: %v", err)
	})

	t.Run("owner must belong to the intent clinic", func(t *testing.T) {
		txErr := NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
			created, err := writer.CreateForOwnerRegistration(txCtx, intent)
			assert.Nil(t, created)
			return err
		})
		require.Error(t, txErr)
		assert.True(t, apperrors.IsInvalidInput(txErr), "cross-clinic owner must be rejected: %v", txErr)

		var petCount int64
		require.NoError(t, db.Model(&model.Pet{}).
			Where("name = ?", "cross-clinic capability pet").
			Count(&petCount).Error)
		assert.Zero(t, petCount)
	})
}

func TestPetRepository_Create_ReturnsCommittedReloadedPet(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "単体ペット再読込飼主")
	species := &model.AnimalSpecies{Name: "単体ペット再読込種"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		PetNumber:       fmt.Sprintf("%d-1", owner.ID),
		Name:            "単体ペット再読込",
	}

	require.NoError(t, NewPetRepository(db).Create(ctx, pet))
	require.NotZero(t, pet.ID)
	require.NotNil(t, pet.Owner)
	require.NotNil(t, pet.AnimalSpecies)
	assert.Equal(t, owner.ID, pet.Owner.ID)
	assert.Equal(t, species.ID, pet.AnimalSpecies.ID)
	assert.Equal(t, fmt.Sprintf("%d-1", owner.ID), pet.PetNumber)
}

func TestPetRepository_Create_RejectsCrossClinicInsuranceAndRollsBackPet(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "direct pet cross-clinic insurance")
	species := &model.AnimalSpecies{Name: "direct pet cross-clinic insurance species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	foreignInsurance := &model.Insurance{
		ClinicID:     clinicB,
		Name:         "direct pet foreign insurance",
		CoverageRate: 50,
	}
	require.NoError(t, db.WithContext(ctx).Create(foreignInsurance).Error)
	pet := &model.Pet{
		ClinicID:        clinicA,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		InsuranceID:     &foreignInsurance.ID,
		Name:            "direct pet rejected insurance",
	}

	err := NewPetRepository(db).Create(ctx, pet)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "cross-clinic master FK must be rejected: %v", err)

	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", pet.Name).
		Count(&petCount).Error)
	assert.Zero(t, petCount, "rejected pet must not survive the transaction")
}

func TestPetRepository_Create_SerializesSameOwnerNumberAllocation(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicID = uint64(1)
		workers  = 6
	)

	owner := makeTestOwner(t, db, clinicID, "same owner concurrent pet create")
	species := &model.AnimalSpecies{Name: "same owner concurrent species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	repo := NewPetRepository(db)

	start := make(chan struct{})
	results := make(chan *model.Pet, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(index int) {
			defer wg.Done()
			pet := &model.Pet{
				ClinicID:        clinicID,
				OwnerID:         owner.ID,
				AnimalSpeciesID: species.ID,
				Name:            fmt.Sprintf("concurrent pet %d", index),
			}
			<-start
			if err := repo.Create(ctx, pet); err != nil {
				errs <- err
				return
			}
			results <- pet
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	close(results)

	for err := range errs {
		require.NoError(t, err)
	}
	numbers := make(map[string]struct{}, workers)
	for created := range results {
		numbers[created.PetNumber] = struct{}{}
		require.NotNil(t, created.Owner)
		require.NotNil(t, created.AnimalSpecies)
	}
	require.Len(t, numbers, workers)
	for sequence := 1; sequence <= workers; sequence++ {
		assert.Contains(t, numbers, fmt.Sprintf("%d-%d", owner.ID, sequence))
	}
}

func TestOwnerRepository_CreateWithPets_ReloadFailureRollsBackGraph(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	species := &model.AnimalSpecies{Name: "owner reload failure species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "owner reload failure",
		Email:    "owner-reload-failure@example.com",
	}
	sentinel := errors.New("simulated owner response reload failure")
	callbackName := "owner_graph_reload_failure_rollback"
	callbackRemoved := false
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement.Schema != nil &&
			query.Statement.Schema.Table == "owners" &&
			len(query.Statement.Preloads) > 0 {
			query.AddError(sentinel)
		}
	}))
	t.Cleanup(func() {
		if !callbackRemoved {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	err := NewOwnerRepository(db).CreateWithPets(ctx, owner, []model.Pet{{
		AnimalSpeciesID: species.ID,
		Name:            "owner reload failure pet",
	}})
	require.ErrorIs(t, err, sentinel)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRemoved = true
	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", "owner-reload-failure@example.com").
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount, "owner reload failure must roll back the graph")
	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "owner reload failure pet").
		Count(&petCount).Error)
	assert.Zero(t, petCount, "nested pet must roll back with owner reload failure")
}

func TestOwnerRepository_FindByID_UsesAmbientTransaction(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "owner ambient find",
		Email:    "owner-ambient-find@example.com",
	}
	rollbackOuter := errors.New("rollback outer transaction after ambient find")

	err := NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := txFromContext(txCtx)
		require.NotNil(t, ambientTx)
		require.NoError(t, ambientTx.WithContext(txCtx).Create(owner).Error)

		loaded, findErr := NewOwnerRepository(db).FindByID(txCtx, clinicID, owner.ID)
		require.NoError(t, findErr)
		require.Equal(t, owner.ID, loaded.ID)
		require.Equal(t, owner.Email, loaded.Email)
		return rollbackOuter
	})
	require.ErrorIs(t, err, rollbackOuter)

	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", owner.Email).
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount, "outer rollback must remove the owner observed through FindByID")
}

func TestOwnerRepository_CreateWithPets_AmbientTransactionNeverEscapesBaseDB(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	species := &model.AnimalSpecies{Name: "owner ambient transaction species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	owner := &model.Owner{
		ClinicID: clinicID,
		Name:     "owner ambient transaction",
		Email:    "owner-ambient-transaction@example.com",
	}
	escapedBaseDB := errors.New("owner graph query escaped ambient transaction")
	rollbackOuter := errors.New("rollback outer transaction after assertion")
	callbackName := "owner_graph_ambient_transaction_guard"
	var ambientConnPool gorm.ConnPool
	ownerQueries := 0
	callbackRemoved := false
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if ambientConnPool == nil ||
			query.Statement.Schema == nil ||
			query.Statement.Schema.Table != "owners" {
			return
		}
		ownerQueries++
		if query.Statement.ConnPool != ambientConnPool {
			query.AddError(escapedBaseDB)
		}
	}))
	t.Cleanup(func() {
		if !callbackRemoved {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	err := NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := txFromContext(txCtx)
		require.NotNil(t, ambientTx)
		ambientConnPool = ambientTx.Statement.ConnPool

		require.NoError(t, NewOwnerRepository(db).CreateWithPets(txCtx, owner, []model.Pet{{
			AnimalSpeciesID: species.ID,
			Name:            "owner ambient transaction pet",
		}}))
		require.Len(t, owner.Pets, 1)
		return rollbackOuter
	})
	require.ErrorIs(t, err, rollbackOuter)
	require.NotErrorIs(t, err, escapedBaseDB)
	assert.GreaterOrEqual(t, ownerQueries, 2, "owner lock and graph reload must both use the ambient transaction")

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRemoved = true
	var ownerCount int64
	require.NoError(t, db.Model(&model.Owner{}).
		Where("email = ?", "owner-ambient-transaction@example.com").
		Count(&ownerCount).Error)
	assert.Zero(t, ownerCount, "outer rollback must include owner graph creation")
	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "owner ambient transaction pet").
		Count(&petCount).Error)
	assert.Zero(t, petCount, "outer rollback must include nested pet creation")
}

func TestPetRepository_Create_AmbientTransactionNeverEscapesBaseDB(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "pet ambient transaction owner")
	species := &model.AnimalSpecies{Name: "pet ambient transaction species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		Name:            "pet ambient transaction",
	}
	escapedBaseDB := errors.New("pet graph query escaped ambient transaction")
	rollbackOuter := errors.New("rollback outer transaction after pet assertion")
	callbackName := "pet_graph_ambient_transaction_guard"
	var ambientConnPool gorm.ConnPool
	graphQueries := 0
	callbackRemoved := false
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if ambientConnPool == nil || query.Statement.Schema == nil {
			return
		}
		table := query.Statement.Schema.Table
		if table != "owners" && table != "pets" {
			return
		}
		graphQueries++
		if query.Statement.ConnPool != ambientConnPool {
			query.AddError(escapedBaseDB)
		}
	}))
	t.Cleanup(func() {
		if !callbackRemoved {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	err := NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := txFromContext(txCtx)
		require.NotNil(t, ambientTx)
		ambientConnPool = ambientTx.Statement.ConnPool

		require.NoError(t, NewPetRepository(db).Create(txCtx, pet))
		require.NotZero(t, pet.ID)
		require.NotNil(t, pet.Owner)
		require.NotNil(t, pet.AnimalSpecies)
		return rollbackOuter
	})
	require.ErrorIs(t, err, rollbackOuter)
	require.NotErrorIs(t, err, escapedBaseDB)
	assert.GreaterOrEqual(t, graphQueries, 3, "owner lock, pet count, and graph reload must use the ambient transaction")

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRemoved = true
	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", pet.Name).
		Count(&petCount).Error)
	assert.Zero(t, petCount, "outer rollback must include direct pet creation")
}

func TestPetRepository_Create_ReloadFailureRollsBackPet(t *testing.T) {
	db := setupOwnerPetIsolationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	owner := makeTestOwner(t, db, clinicID, "pet reload failure owner")
	species := &model.AnimalSpecies{Name: "pet reload failure species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	pet := &model.Pet{
		ClinicID:        clinicID,
		OwnerID:         owner.ID,
		AnimalSpeciesID: species.ID,
		Name:            "pet reload failure",
	}
	sentinel := errors.New("simulated pet response reload failure")
	callbackName := "pet_reload_failure_rollback"
	callbackRemoved := false
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register(callbackName, func(query *gorm.DB) {
		if query.Statement.Schema != nil &&
			query.Statement.Schema.Table == "pets" &&
			len(query.Statement.Preloads) > 0 {
			query.AddError(sentinel)
		}
	}))
	t.Cleanup(func() {
		if !callbackRemoved {
			require.NoError(t, db.Callback().Query().Remove(callbackName))
		}
	})

	err := NewPetRepository(db).Create(ctx, pet)
	require.ErrorIs(t, err, sentinel)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRemoved = true
	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "pet reload failure").
		Count(&petCount).Error)
	assert.Zero(t, petCount, "pet reload failure must roll back the pet write")
}
