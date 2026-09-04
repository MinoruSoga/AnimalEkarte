package medicalrecord

// clinical_master_delete_tx_atomicity_test.go — clinic-scoped clinical master Delete
// methods that join ambient tx via DBOrTx. Unused rows roll back when a later
// step in the same WithTx fails.

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/model"
)

var errClinicalMasterDeleteAmbientTxSentinel = errors.New("simulated post-delete failure in ambient tx")

func assertClinicalMasterDeleteRollsBack(
	t *testing.T,
	db *gorm.DB,
	id uint64,
	deleteFn func(ctx context.Context) error,
	findFn func(ctx context.Context) (uint64, error),
) {
	t.Helper()
	ctx := context.Background()
	txErr := withTx(ctx, db, func(txCtx context.Context) error {
		if err := deleteFn(txCtx); err != nil {
			return err
		}
		return errClinicalMasterDeleteAmbientTxSentinel
	})
	require.ErrorIs(t, txErr, errClinicalMasterDeleteAmbientTxSentinel)

	gotID, err := findFn(ctx)
	require.NoError(t, err, "Delete must join ambient tx so rollback restores the row")
	assert.Equal(t, id, gotID)
}

func TestConsultationRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupConsultationTestDB(t)
	repo := NewConsultationRepository(db)
	const clinicA = uint64(1)
	c := makeConsultation(t, db, clinicA, "原子削除ロールバック対象問診", nil)
	assertClinicalMasterDeleteRollsBack(t, db, c.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, c.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, c.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestCheckupTypeRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupCheckupTypeTestDB(t)
	repo := NewCheckupTypeRepository(db)
	ctx := context.Background()
	const clinicA = uint64(1)
	ct := &model.CheckupType{ClinicID: clinicA, Name: "原子削除ロールバック対象検査種別"}
	require.NoError(t, db.WithContext(ctx).Create(ct).Error)
	assertClinicalMasterDeleteRollsBack(t, db, ct.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, ct.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, ct.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestHospitalizationPlanRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupHospitalizationPlanRepoTestDB(t)
	repo := NewHospitalizationPlanRepository(db)
	const clinicA = uint64(1)
	plan := makeHospitalizationPlanFixture(t, db, clinicA, "原子削除ロールバック対象入院プラン")
	assertClinicalMasterDeleteRollsBack(t, db, plan.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, plan.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, plan.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestChiefComplaintTypeRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupChiefComplaintTypeRepositoryTestDB(t)
	repo := NewChiefComplaintTypeRepository(db)
	const clinicA = uint64(1)
	c := makeChiefComplaintType(t, db, clinicA, "原子削除ロールバック対象主訴")
	assertClinicalMasterDeleteRollsBack(t, db, c.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, c.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, c.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestDiagnosisNameRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupDiagnosisNameTestDB(t)
	repo := NewDiagnosisNameRepository(db)
	const clinicA = uint64(1)
	dt := makeDiagnosisTypeMaster(t, db, clinicA, "原子削除診断カテゴリ")
	dn := makeDiagnosisNameRec(t, db, clinicA, dt.ID, "原子削除ロールバック対象診断名")
	assertClinicalMasterDeleteRollsBack(t, db, dn.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, dn.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, dn.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestDiagnosisTypeRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupDiagnosisTypeTestDB(t)
	repo := NewDiagnosisTypeRepository(db)
	const clinicA = uint64(1)
	dt := makeDiagnosisTypeMaster(t, db, clinicA, "原子削除ロールバック対象診断カテゴリ")
	assertClinicalMasterDeleteRollsBack(t, db, dt.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, dt.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, dt.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}

func TestVaccineRepository_Delete_RollsBackWhenAmbientTxFails(t *testing.T) {
	db := setupVaccineRepositoryTestDB(t)
	repo := NewVaccineRepository(db)
	const clinicA = uint64(1)
	v := makeVaccineMaster(t, db, clinicA, "原子削除ロールバック対象ワクチン")
	assertClinicalMasterDeleteRollsBack(t, db, v.ID,
		func(ctx context.Context) error { return repo.Delete(ctx, clinicA, v.ID) },
		func(ctx context.Context) (uint64, error) {
			got, err := repo.FindByID(ctx, clinicA, v.ID)
			if err != nil {
				return 0, err
			}
			return got.ID, nil
		},
	)
}
