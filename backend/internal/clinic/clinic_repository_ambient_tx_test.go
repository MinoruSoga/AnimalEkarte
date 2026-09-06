package clinic

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func TestClinicRepository_FindAllParticipatesInAmbientTransactionPoolOne(
	t *testing.T,
) {
	db := testdb.SetupIsolatedTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
	))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	repository := NewClinicRepository(db)
	transactor := persistence.NewTransactor(db)
	rollbackProbe := errors.New("rollback clinic inventory probe")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = transactor.WithTx(ctx, func(txCtx context.Context) error {
		company := &model.Company{
			Name: "clinic inventory ambient tx company",
		}
		if createErr := persistence.DBOrTx(txCtx, db).
			Create(company).Error; createErr != nil {
			return fmt.Errorf("create transaction probe company: %w", createErr)
		}
		clinic := &model.Clinic{
			CompanyID: company.ID,
			Name:      "clinic inventory ambient tx clinic",
			IsActive:  true,
		}
		if createErr := repository.Create(txCtx, clinic); createErr != nil {
			return fmt.Errorf("create transaction probe clinic: %w", createErr)
		}

		clinics, listErr := repository.FindAll(txCtx)
		if listErr != nil {
			return fmt.Errorf(
				"FindAll must reuse the sole transaction connection: %w",
				listErr,
			)
		}
		if !containsClinicIDForAmbientTxTest(clinics, clinic.ID) {
			return errors.New("ambient FindAll did not observe uncommitted clinic")
		}
		return rollbackProbe
	})
	require.ErrorIs(t, err, rollbackProbe)

	var persisted int64
	require.NoError(t, db.Model(&model.Clinic{}).
		Where("name = ?", "clinic inventory ambient tx clinic").
		Count(&persisted).Error)
	assert.Zero(t, persisted, "transaction probe must roll back")
}

func TestClinicRepository_FindByIDsParticipatesInAmbientTransactionPoolOne(
	t *testing.T,
) {
	db, repository, transactor, ctx := setupClinicInventoryAmbientTxProbe(t)
	rollbackProbe := errors.New("rollback clinic FindByIDs probe")

	var clinicID uint64
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		clinic, createErr := createClinicInventoryAmbientTxClinic(
			txCtx,
			db,
			repository,
			"clinic inventory ambient tx find-by-ids",
		)
		if createErr != nil {
			return createErr
		}
		clinicID = clinic.ID

		clinics, findErr := repository.FindByIDs(txCtx, []uint64{clinic.ID})
		if findErr != nil {
			return fmt.Errorf(
				"FindByIDs must reuse the sole transaction connection: %w",
				findErr,
			)
		}
		if !containsClinicIDForAmbientTxTest(clinics, clinic.ID) {
			return errors.New("ambient FindByIDs did not observe uncommitted clinic")
		}
		return rollbackProbe
	})
	require.ErrorIs(t, err, rollbackProbe)
	require.NotZero(t, clinicID)
	assertClinicInventoryAmbientTxRolledBack(t, db, clinicID)
}

func TestClinicRepository_FindActiveIDsParticipatesInAmbientTransactionPoolOne(
	t *testing.T,
) {
	db, repository, transactor, ctx := setupClinicInventoryAmbientTxProbe(t)
	rollbackProbe := errors.New("rollback clinic FindActiveIDs probe")

	var activeID, inactiveID uint64
	err := transactor.WithTx(ctx, func(txCtx context.Context) error {
		active, createErr := createClinicInventoryAmbientTxClinic(
			txCtx,
			db,
			repository,
			"clinic inventory ambient tx find-active",
		)
		if createErr != nil {
			return createErr
		}
		inactive, createErr := createClinicInventoryAmbientTxClinic(
			txCtx,
			db,
			repository,
			"clinic inventory ambient tx find-inactive",
		)
		if createErr != nil {
			return createErr
		}
		if updateErr := persistence.DBOrTx(txCtx, db).
			Model(&model.Clinic{}).
			Where("id = ?", inactive.ID).
			Update("is_active", false).Error; updateErr != nil {
			return fmt.Errorf("deactivate transaction probe clinic: %w", updateErr)
		}
		activeID = active.ID
		inactiveID = inactive.ID

		ids, findErr := repository.FindActiveIDs(txCtx, []uint64{active.ID, inactive.ID})
		if findErr != nil {
			return fmt.Errorf(
				"FindActiveIDs must reuse the sole transaction connection: %w",
				findErr,
			)
		}
		if !containsUint64(ids, active.ID) {
			return errors.New("ambient FindActiveIDs did not observe uncommitted active clinic")
		}
		if containsUint64(ids, inactive.ID) {
			return errors.New("ambient FindActiveIDs observed inactive clinic")
		}
		return rollbackProbe
	})
	require.ErrorIs(t, err, rollbackProbe)
	require.NotZero(t, activeID)
	require.NotZero(t, inactiveID)
	assertClinicInventoryAmbientTxRolledBack(t, db, activeID)
	assertClinicInventoryAmbientTxRolledBack(t, db, inactiveID)
}

func containsClinicIDForAmbientTxTest(
	clinics []model.Clinic,
	clinicID uint64,
) bool {
	for i := range clinics {
		if clinics[i].ID == clinicID {
			return true
		}
	}
	return false
}

func containsUint64(ids []uint64, want uint64) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func setupClinicInventoryAmbientTxProbe(t *testing.T) (
	*gorm.DB,
	ClinicRepository,
	persistence.Transactor,
	context.Context,
) {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(
		db,
		&model.Company{},
		&model.Clinic{},
	))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return db, NewClinicRepository(db), persistence.NewTransactor(db), ctx
}

func createClinicInventoryAmbientTxClinic(
	txCtx context.Context,
	db *gorm.DB,
	repository ClinicRepository,
	name string,
) (*model.Clinic, error) {
	company := &model.Company{
		Name: name + " company",
	}
	if createErr := persistence.DBOrTx(txCtx, db).
		Create(company).Error; createErr != nil {
		return nil, fmt.Errorf("create transaction probe company: %w", createErr)
	}
	clinic := &model.Clinic{
		CompanyID: company.ID,
		Name:      name,
		IsActive:  true,
	}
	if createErr := repository.Create(txCtx, clinic); createErr != nil {
		return nil, fmt.Errorf("create transaction probe clinic: %w", createErr)
	}
	return clinic, nil
}

func assertClinicInventoryAmbientTxRolledBack(
	t *testing.T,
	db *gorm.DB,
	clinicID uint64,
) {
	t.Helper()
	var persisted int64
	require.NoError(t, db.Model(&model.Clinic{}).
		Where("id = ?", clinicID).
		Count(&persisted).Error)
	assert.Zero(t, persisted, "transaction probe must roll back")
}
