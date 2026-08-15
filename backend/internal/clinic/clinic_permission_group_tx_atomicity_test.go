package clinic

// clinic_permission_group_tx_atomicity_test.go — BE-refactor.md X-7 (tx-clinic-create-nonparticipation)
//
// clinic_service.CreateClinic wraps clinic + default-permission-group creation in
// Transactor.WithTx, but clinic_repository.Create / permission_group_repository.Create must
// use dbOrTx(ctx, r.db) to actually participate in that ambient tx. Before the fix both used
// r.db.WithContext(ctx) directly, so a failure partway through the group-creation loop left an
// orphan clinic (and possibly one orphan permission group) committed, because WithTx's rollback
// has nothing to roll back — the rows were already autocommitted outside the tx.
//
// This test replicates the exact write sequence of clinicService.CreateClinic's WithTx body
// (clinic Create → permission group Create ×2) using the real repositories and the real
// Transactor, and forces the 2nd permission-group Create to fail. It asserts that NEITHER the
// clinic row NOR the first permission-group row survive — i.e. the whole operation rolled back
// atomically.

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestPermissionGroupRepository_ClinicCreateTxRollback_OnSecondGroupCreateFailure(t *testing.T) {
	db := setupClinicTestDB(t)
	ctx := context.Background()

	company := &model.Company{Name: "X-7ロールバック検証法人"}
	require.NoError(t, db.WithContext(ctx).Create(company).Error)

	clinicRepo := NewClinicRepository(db)
	pgRepo := auth.NewPermissionGroupRepository(db)
	transactor := persistence.NewTransactor(db)

	clinic := &model.Clinic{CompanyID: company.ID, Name: "X-7ロールバック検証院"}

	forcedErr := errors.New("forced failure on 2nd permission group create")
	txErr := transactor.WithTx(ctx, func(ctx context.Context) error {
		if err := clinicRepo.Create(ctx, clinic); err != nil {
			return err
		}

		// 1st default permission group ("執行") — real write, must participate in the ambient tx.
		group1 := &model.PermissionGroup{
			ClinicID: clinic.ID, Name: "執行", Description: "執行権限", IsActive: true, SortOrder: 1,
		}
		if err := pgRepo.Create(ctx, group1); err != nil {
			return err
		}

		// 2nd default permission group ("一般") fails — simulates the real bug trigger
		// (e.g. transient DB error, unique-constraint collision, etc.) partway through the loop.
		return forcedErr
	})

	require.Error(t, txErr)

	var clinicCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.Clinic{}).
		Where("id = ?", clinic.ID).Count(&clinicCount).Error)
	assert.Equal(t, int64(0), clinicCount,
		"clinic row must not survive when the 2nd default permission group Create fails inside WithTx (orphan clinic bug)")

	var groupCount int64
	require.NoError(t, db.WithContext(ctx).Model(&model.PermissionGroup{}).
		Where("clinic_id = ?", clinic.ID).Count(&groupCount).Error)
	assert.Equal(t, int64(0), groupCount,
		"the 1st permission group created before the failure must also roll back (no partial group set)")
}
