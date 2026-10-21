package medicalrecord

// Package master persistence for treatment_packages / treatment_package_items /
// treatment_package_catalog_versions (EMR-518 §3/§4, EMR-521).
// Lock order for catalog writes: clinic row -> catalog row -> package IDs asc -> component.
// Reads never create rows: a missing catalog row is a setup/invariant error.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errPackageCatalogNotInitialized = apperrors.WrapInternalServerError("treatment package catalog is not initialized")

type TreatmentPackageRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.TreatmentPackage, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPackage, error)
	// FindLiveItemsByPackageIDs returns removed_at IS NULL items ordered by package position.
	FindLiveItemsByPackageIDs(ctx context.Context, clinicID uint64, packageIDs []uint64) ([]model.TreatmentPackageItem, error)
	// FindAllItemsForPackage returns every item row incl. removed (replacement merge needs keys).
	FindAllItemsForPackage(ctx context.Context, clinicID, packageID uint64) ([]model.TreatmentPackageItem, error)
	LockClinicRow(ctx context.Context, clinicID uint64) error
	// LockCatalogRow takes FOR UPDATE on the catalog row and returns its revision.
	// Missing row is an invariant error, never INSERTed here.
	LockCatalogRow(ctx context.Context, clinicID uint64) (int64, error)
	GetCatalogRevision(ctx context.Context, clinicID uint64) (int64, error)
	IncrementCatalogRevision(ctx context.Context, clinicID uint64, expected int64) error
	// LockPackagesByIDAsc takes FOR UPDATE on nondeleted packages in ascending id order.
	LockPackagesByIDAsc(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackage, error)
	MaxSortOrder(ctx context.Context, clinicID uint64) (int, error)
	Create(ctx context.Context, pkg *model.TreatmentPackage, items []model.TreatmentPackageItem) error
	UpdateFieldsCAS(ctx context.Context, clinicID, id uint64, expectedVersion int64, fields map[string]any) (*model.TreatmentPackage, error)
	// ReplaceItems merges under the locked parent: existing keys update, new keys insert,
	// missing keys keep their row with removed_at/position=NULL. No physical delete.
	ReplaceItems(ctx context.Context, clinicID, packageID uint64, keep []model.TreatmentPackageItem, removeIDs []uint64, now time.Time) error
	SoftDeleteCAS(ctx context.Context, clinicID, id uint64, expectedVersion int64, actorID uint64, now time.Time) error
	UpdateSortOrderBumpVersion(ctx context.Context, clinicID, id uint64, sortOrder int, actorID uint64) error
}

type treatmentPackageRepositoryImpl struct{ db *gorm.DB }

func NewTreatmentPackageRepository(db *gorm.DB) TreatmentPackageRepository {
	return &treatmentPackageRepositoryImpl{db: db}
}

func (r *treatmentPackageRepositoryImpl) FindAll(ctx context.Context, clinicID uint64) ([]model.TreatmentPackage, error) {
	pkgs := make([]model.TreatmentPackage, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("deleted_at IS NULL").
		Order("sort_order ASC, id ASC").
		Limit(persistence.MaxMasterListRows).
		Find(&pkgs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_package", "")
	}
	return pkgs, nil
}

func (r *treatmentPackageRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.TreatmentPackage, error) {
	var pkg model.TreatmentPackage
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&pkg).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_package", fmt.Sprintf("%d", id))
	}
	return &pkg, nil
}

func (r *treatmentPackageRepositoryImpl) FindLiveItemsByPackageIDs(ctx context.Context, clinicID uint64, packageIDs []uint64) ([]model.TreatmentPackageItem, error) {
	items := make([]model.TreatmentPackageItem, 0)
	if len(packageIDs) == 0 {
		return items, nil
	}
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("package_id IN ? AND removed_at IS NULL", packageIDs).
		Order("package_id ASC, position ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_package_item", "")
	}
	return items, nil
}

func (r *treatmentPackageRepositoryImpl) FindAllItemsForPackage(ctx context.Context, clinicID, packageID uint64) ([]model.TreatmentPackageItem, error) {
	items := make([]model.TreatmentPackageItem, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("package_id = ?", packageID).
		Order("id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_package_item", fmt.Sprintf("%d", packageID))
	}
	return items, nil
}

// LockClinicRow takes FOR UPDATE on the clinics row so every catalog writer
// joins the same lock order (clinic -> catalog -> package -> component).
func (r *treatmentPackageRepositoryImpl) LockClinicRow(ctx context.Context, clinicID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("clinic lock requires an ambient transaction")
	}
	var id uint64
	err := persistence.DBOrTx(ctx, r.db).
		Table("clinics").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", clinicID).
		Select("id").
		Scan(&id).Error
	if err != nil {
		return apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", clinicID))
	}
	if id == 0 {
		return apperrors.WrapNotFound("clinic", fmt.Sprintf("%d", clinicID))
	}
	return nil
}

func (r *treatmentPackageRepositoryImpl) LockCatalogRow(ctx context.Context, clinicID uint64) (int64, error) {
	if persistence.TxFromContext(ctx) == nil {
		return 0, apperrors.WrapInternalServerError("catalog lock requires an ambient transaction")
	}
	var row model.TreatmentPackageCatalogVersion
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("clinic_id = ?", clinicID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || apperrors.IsNotFound(err) {
			return 0, errPackageCatalogNotInitialized
		}
		return 0, apperrors.FromGORM(err, "treatment_package_catalog_version", fmt.Sprintf("%d", clinicID))
	}
	return row.Revision, nil
}

// GetCatalogRevision reads the persisted catalog revision without locking or creating it.
func (r *treatmentPackageRepositoryImpl) GetCatalogRevision(ctx context.Context, clinicID uint64) (int64, error) {
	var row model.TreatmentPackageCatalogVersion
	err := persistence.DBOrTx(ctx, r.db).
		Where("clinic_id = ?", clinicID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || apperrors.IsNotFound(err) {
			return 0, errPackageCatalogNotInitialized
		}
		return 0, apperrors.FromGORM(err, "treatment_package_catalog_version", fmt.Sprintf("%d", clinicID))
	}
	return row.Revision, nil
}

func (r *treatmentPackageRepositoryImpl) IncrementCatalogRevision(ctx context.Context, clinicID uint64, expected int64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("catalog mutation requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPackageCatalogVersion{}).
		Where("clinic_id = ? AND revision = ?", clinicID, expected).
		Updates(map[string]any{
			"revision":   gorm.Expr("revision + 1"),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_package_catalog_version", fmt.Sprintf("%d", clinicID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict("カタログが更新されています。再読み込みしてから保存してください")
	}
	return nil
}

func (r *treatmentPackageRepositoryImpl) LockPackagesByIDAsc(ctx context.Context, clinicID uint64, ids []uint64) ([]model.TreatmentPackage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("package lock requires an ambient transaction")
	}
	ordered := append([]uint64(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	pkgs := make([]model.TreatmentPackage, 0, len(ordered))
	if len(ordered) == 0 {
		return pkgs, nil
	}
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id IN ? AND deleted_at IS NULL", ordered).
		Order("id ASC").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&pkgs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment_package", "")
	}
	return pkgs, nil
}

func (r *treatmentPackageRepositoryImpl) MaxSortOrder(ctx context.Context, clinicID uint64) (int, error) {
	var max sql.NullInt64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPackage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("deleted_at IS NULL").
		Select("MAX(sort_order)").
		Scan(&max).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "treatment_package", "")
	}
	if !max.Valid {
		return -1, nil
	}
	return int(max.Int64), nil
}

func (r *treatmentPackageRepositoryImpl) Create(ctx context.Context, pkg *model.TreatmentPackage, items []model.TreatmentPackageItem) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("package create requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Create(pkg).Error; err != nil {
		return apperrors.FromGORM(err, "treatment_package", "")
	}
	for i := range items {
		items[i].ClinicID = pkg.ClinicID
		items[i].PackageID = pkg.ID
	}
	if len(items) == 0 {
		return nil
	}
	if err := db.Create(&items).Error; err != nil {
		return apperrors.FromGORM(err, "treatment_package_item", "")
	}
	return nil
}

// UpdateFieldsCAS increments version exactly once under the editing-start version CAS.
func (r *treatmentPackageRepositoryImpl) UpdateFieldsCAS(ctx context.Context, clinicID, id uint64, expectedVersion int64, fields map[string]any) (*model.TreatmentPackage, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("package update requires an ambient transaction")
	}
	fields["version"] = gorm.Expr("version + 1")
	fields["updated_at"] = time.Now()
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPackage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND version = ? AND deleted_at IS NULL", id, expectedVersion).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "treatment_package", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapConflict("編集後に変更されています。再読み込みしてから保存してください")
	}
	return r.FindByID(ctx, clinicID, id)
}

// ReplaceItems keeps removed rows (removed_at + position=NULL) so component keys
// are never reused with a different meaning. Reinserting a removed key is rejected upstream.
func (r *treatmentPackageRepositoryImpl) ReplaceItems(ctx context.Context, clinicID, packageID uint64, keep []model.TreatmentPackageItem, removeIDs []uint64, now time.Time) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("component replacement requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db)
	for _, id := range removeIDs {
		result := db.Model(&model.TreatmentPackageItem{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ? AND removed_at IS NULL", id).
			Updates(map[string]any{"removed_at": now, "position": nil, "updated_at": now})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "treatment_package_item", fmt.Sprintf("%d", id))
		}
	}
	for i := range keep {
		item := &keep[i]
		if item.ID == 0 {
			item.ClinicID = clinicID
			item.PackageID = packageID
			item.CreatedAt = now
			item.UpdatedAt = now
			if err := db.Create(item).Error; err != nil {
				return apperrors.FromGORM(err, "treatment_package_item", "")
			}
			continue
		}
		result := db.Model(&model.TreatmentPackageItem{}).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ? AND removed_at IS NULL", item.ID).
			Updates(map[string]any{
				"item_kind":           item.ItemKind,
				"consultation_id":     item.ConsultationID,
				"procedure_id":        item.ProcedureID,
				"medicine_id":         item.MedicineID,
				"exam_type_id":        item.ExamTypeID,
				"vaccine_id":          item.VaccineID,
				"checkup_type_id":     item.CheckupTypeID,
				"merchandise_item_id": item.MerchandiseItemID,
				"quantity":            item.Quantity,
				"manual_definition":   item.ManualDefinition,
				"position":            item.Position,
				"updated_at":          now,
			})
		if result.Error != nil {
			return apperrors.FromGORM(result.Error, "treatment_package_item", fmt.Sprintf("%d", item.ID))
		}
	}
	return nil
}

// SoftDeleteCAS sets the pair-null deleted_at/deleted_by and bumps version once
// under the requested version CAS. Children/history are preserved.
func (r *treatmentPackageRepositoryImpl) SoftDeleteCAS(ctx context.Context, clinicID, id uint64, expectedVersion int64, actorID uint64, now time.Time) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("package delete requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPackage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND version = ? AND deleted_at IS NULL", id, expectedVersion).
		Updates(map[string]any{
			"deleted_at": now,
			"deleted_by": actorID,
			"updated_by": actorID,
			"version":    gorm.Expr("version + 1"),
			"updated_at": now,
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_package", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapConflict("編集後に変更されています。再読み込みしてから削除してください")
	}
	return nil
}

// UpdateSortOrderBumpVersion applies one sort_order change and one version bump
// to a package already locked by the reorder caller.
func (r *treatmentPackageRepositoryImpl) UpdateSortOrderBumpVersion(ctx context.Context, clinicID, id uint64, sortOrder int, actorID uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("reorder requires an ambient transaction")
	}
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.TreatmentPackage{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{
			"sort_order": sortOrder,
			"updated_by": actorID,
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment_package", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment_package", fmt.Sprintf("%d", id))
	}
	return nil
}
