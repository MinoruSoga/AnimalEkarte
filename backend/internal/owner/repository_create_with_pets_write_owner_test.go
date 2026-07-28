package owner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	ownerdomain "github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	petdomain "github.com/animal-ekarte/backend/internal/pet"
	"github.com/animal-ekarte/backend/internal/testdb"
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

func setupOwnerCreateWithPetsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.AnimalSpecies{},
		&model.Pet{},
		&model.Insurance{},
	))
	require.NoError(t, db.Exec("TRUNCATE TABLE insurances CASCADE").Error)
	require.NoError(t, db.Exec("TRUNCATE TABLE animal_species CASCADE").Error)
	return db
}

func newOwnerRepository(db *gorm.DB) ownerdomain.Repository {
	writer := petdomain.NewWriter(db)
	return ownerdomain.NewRepository(db, petdomain.NewOwnerRegistrationAdapter(writer))
}

func TestOwnerRepository_CreateWithPets_AssignsNumbersAndReloadsCommittedGraph(t *testing.T) {
	db := setupOwnerCreateWithPetsTestDB(t)
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

	require.NoError(t, newOwnerRepository(db).CreateWithPets(ctx, owner, pets))
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
	reloadedPet, err := petdomain.NewRepository(db).FindByID(ctx, clinicID, byName["ポチ"].ID)
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
	db := setupOwnerCreateWithPetsTestDB(t)
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

	require.NoError(t, newOwnerRepository(db).CreateWithPets(ctx, owner, []model.Pet{{
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
	db := setupOwnerCreateWithPetsTestDB(t)
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

	err := newOwnerRepository(db).CreateWithPets(ctx, owner, pets)
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
	db := setupOwnerCreateWithPetsTestDB(t)
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

	err := newOwnerRepository(db).CreateWithPets(ctx, owner, pets)
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
	db := setupOwnerCreateWithPetsTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	sentinel := errors.New("simulated pet owner failure")
	capabilityCalled := false
	writer := ownerRegistrationWriterFunc(func(
		txCtx context.Context,
		intent petdomain.OwnerRegistrationIntent,
	) ([]model.Pet, error) {
		capabilityCalled = true
		require.NotNil(t, persistence.TxFromContext(txCtx), "pet capability must receive the owner transaction")
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
	err := ownerdomain.NewRepository(db, petdomain.NewOwnerRegistrationAdapter(writer)).CreateWithPets(ctx, owner, []model.Pet{{
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

func TestOwnerRepository_CreateWithPets_ReloadFailureRollsBackGraph(t *testing.T) {
	db := setupOwnerCreateWithPetsTestDB(t)
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

	err := newOwnerRepository(db).CreateWithPets(ctx, owner, []model.Pet{{
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

func TestOwnerRepository_CreateWithPets_AmbientTransactionNeverEscapesBaseDB(t *testing.T) {
	db := setupOwnerCreateWithPetsTestDB(t)
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

	err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
		ambientTx := persistence.TxFromContext(txCtx)
		require.NotNil(t, ambientTx)
		ambientConnPool = ambientTx.Statement.ConnPool

		require.NoError(t, newOwnerRepository(db).CreateWithPets(txCtx, owner, []model.Pet{{
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
