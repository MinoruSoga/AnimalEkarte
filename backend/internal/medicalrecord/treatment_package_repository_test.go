package medicalrecord

// EMR-521 repository tests (NOT_RUN — real test DB, requires the package
// tables' migration to be applied; written for post-integration verification).

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/testdb"
)

func setupPackageTestDB(t *testing.T) (*gorm.DB, uint64) {
	t.Helper()
	db := testdb.SetupTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db,
		&model.Clinic{}, &model.TreatmentPackage{},
		&model.TreatmentPackageItem{}, &model.TreatmentPackageCatalogVersion{}))
	clinic := model.Clinic{Name: "package test clinic", CompanyID: 1}
	require.NoError(t, db.Create(&clinic).Error)
	return db, clinic.ID
}

func packageWithTx(t *testing.T, db *gorm.DB, fn func(ctx context.Context) error) {
	t.Helper()
	require.NoError(t, persistence.NewTransactor(db).WithTx(context.Background(), fn))
}

func TestTreatmentPackageRepository_CatalogMissingRowNeverWrittenByGet(t *testing.T) {
	db, clinicID := setupPackageTestDB(t)
	repo := NewTreatmentPackageRepository(db)

	_, err := repo.GetCatalogRevision(context.Background(), clinicID)
	assert.Error(t, err) // setup/invariant error; no fake version 1, no INSERT.

	var count int64
	require.NoError(t, db.Model(&model.TreatmentPackageCatalogVersion{}).
		Where("clinic_id = ?", clinicID).Count(&count).Error)
	assert.EqualValues(t, 0, count)
}

func TestTreatmentPackageRepository_ReplaceItemsKeepsRemovedKeys(t *testing.T) {
	db, clinicID := setupPackageTestDB(t)
	repo := NewTreatmentPackageRepository(db)
	now := time.Now()

	var pkg model.TreatmentPackage
	packageWithTx(t, db, func(txCtx context.Context) error {
		pkg = model.TreatmentPackage{ClinicID: clinicID, Name: "p", SortOrder: 0, Version: 1,
			CreatedBy: 1, UpdatedBy: 1, CreatedAt: now, UpdatedAt: now}
		var q model.ExactNumeric
		require.NoError(t, q.Scan("0.25"))
		key := uuid.MustParse("a0000000-0000-4000-8000-000000000011")
		cid := uint64(101)
		return repo.Create(txCtx, &pkg, []model.TreatmentPackageItem{{
			ClinicID: clinicID, ComponentKey: key, ItemKind: model.TreatmentPackageItemKindConsultation,
			ConsultationID: &cid, Quantity: q, CreatedAt: now, UpdatedAt: now,
		}})
	})

	// Replace with an empty desired set: the row is kept with removed_at + position=NULL.
	packageWithTx(t, db, func(txCtx context.Context) error {
		items, err := repo.FindAllItemsForPackage(txCtx, clinicID, pkg.ID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		return repo.ReplaceItems(txCtx, clinicID, pkg.ID, nil, []uint64{items[0].ID}, now)
	})

	items, err := repo.FindAllItemsForPackage(context.Background(), clinicID, pkg.ID)
	require.NoError(t, err)
	require.Len(t, items, 1) // 物理削除なし。
	assert.NotNil(t, items[0].RemovedAt)
	assert.Nil(t, items[0].Position) // (removed_at IS NULL)=(position IS NOT NULL) pair.

	live, err := repo.FindLiveItemsByPackageIDs(context.Background(), clinicID, []uint64{pkg.ID})
	require.NoError(t, err)
	assert.Empty(t, live) // removed は read/preview から除外。
}

func TestTreatmentPackageRepository_SoftDeleteCASHistoryPreserved(t *testing.T) {
	db, clinicID := setupPackageTestDB(t)
	repo := NewTreatmentPackageRepository(db)
	now := time.Now()

	var pkg model.TreatmentPackage
	packageWithTx(t, db, func(txCtx context.Context) error {
		pkg = model.TreatmentPackage{ClinicID: clinicID, Name: "p", SortOrder: 0, Version: 1,
			CreatedBy: 1, UpdatedBy: 1, CreatedAt: now, UpdatedAt: now}
		return repo.Create(txCtx, &pkg, nil)
	})

	packageWithTx(t, db, func(txCtx context.Context) error {
		return repo.SoftDeleteCAS(txCtx, clinicID, pkg.ID, 1, 9, now)
	})
	// Stale version must not delete again (zero-write conflict).
	err := persistence.NewTransactor(db).WithTx(context.Background(), func(txCtx context.Context) error {
		return repo.SoftDeleteCAS(txCtx, clinicID, pkg.ID, 1, 9, now)
	})
	assert.ErrorIs(t, err, apperrors.ErrConflict)

	_, err = repo.FindByID(context.Background(), clinicID, pkg.ID)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
	var stored model.TreatmentPackage
	require.NoError(t, db.Unscoped().First(&stored, "id = ?", pkg.ID).Error)
	assert.NotNil(t, stored.DeletedAt)
	assert.NotNil(t, stored.DeletedBy)
}
