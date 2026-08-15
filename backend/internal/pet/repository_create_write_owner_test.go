package pet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestPetOwnerRegistrationWriter_RequiresAmbientTransactionAndClinicOwnerMatch(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
	)

	owner := makeTestOwner(t, db, clinicA, "pet capability clinic guard")
	species := &model.AnimalSpecies{Name: "pet capability clinic guard species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	writer := NewOwnerRegistrationWriter()
	intent := OwnerRegistrationIntent{
		ClinicID: clinicB,
		OwnerID:  owner.ID,
		Pets: []CreatePetDraft{{
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
		txErr := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
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
	db := setupPetRepositoryIsolationTestDB(t)
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

	require.NoError(t, NewRepository(db).Create(ctx, pet))
	require.NotZero(t, pet.ID)
	require.NotNil(t, pet.Owner)
	require.NotNil(t, pet.AnimalSpecies)
	assert.Equal(t, owner.ID, pet.Owner.ID)
	assert.Equal(t, species.ID, pet.AnimalSpecies.ID)
	assert.Equal(t, fmt.Sprintf("%d-1", owner.ID), pet.PetNumber)
}

func TestPetRepository_Create_RejectsCrossClinicInsuranceAndRollsBackPet(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
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

	err := NewRepository(db).Create(ctx, pet)
	require.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err), "cross-clinic master FK must be rejected: %v", err)

	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", pet.Name).
		Count(&petCount).Error)
	assert.Zero(t, petCount, "rejected pet must not survive the transaction")
}

func TestPetRepository_Create_SerializesSameOwnerNumberAllocation(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
	ctx := context.Background()
	const (
		clinicID = uint64(1)
		workers  = 6
	)

	owner := makeTestOwner(t, db, clinicID, "same owner concurrent pet create")
	species := &model.AnimalSpecies{Name: "same owner concurrent species"}
	require.NoError(t, db.WithContext(ctx).Create(species).Error)
	repo := NewRepository(db)

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

func TestPetRepository_Create_AmbientTransactionNeverEscapesBaseDB(t *testing.T) {
	db := setupPetRepositoryIsolationTestDB(t)
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

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := persistence.TxFromContext(txCtx)
		require.NotNil(t, ambientTx)
		ambientConnPool = ambientTx.Statement.ConnPool

		require.NoError(t, NewRepository(db).Create(txCtx, pet))
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
	db := setupPetRepositoryIsolationTestDB(t)
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

	err := NewRepository(db).Create(ctx, pet)
	require.ErrorIs(t, err, sentinel)

	require.NoError(t, db.Callback().Query().Remove(callbackName))
	callbackRemoved = true
	var petCount int64
	require.NoError(t, db.Model(&model.Pet{}).
		Where("name = ?", "pet reload failure").
		Count(&petCount).Error)
	assert.Zero(t, petCount, "pet reload failure must roll back the pet write")
}
