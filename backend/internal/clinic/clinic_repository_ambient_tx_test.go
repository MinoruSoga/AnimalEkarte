package clinic

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
