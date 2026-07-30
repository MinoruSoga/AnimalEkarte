package staff

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// staffProvisioningRepository is the GORM-backed StaffProvisioningRepository.
type staffProvisioningRepository struct {
	db *gorm.DB
}

// NewStaffProvisioningRepository constructs the production provisioning store.
func NewStaffProvisioningRepository(db *gorm.DB) StaffProvisioningRepository {
	return &staffProvisioningRepository{db: db}
}

func (r *staffProvisioningRepository) WithTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "staff provision transaction failed")
	}
	return nil
}

// AcquireBatchLock takes a transaction-scoped advisory lock derived from batch_id.
// Ambient transaction absence fails closed.
func (r *staffProvisioningRepository) AcquireBatchLock(ctx context.Context, batchID string) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("staff provision lock requires an ambient transaction")
	}
	lockKey := "staff-provision:" + batchID
	if err := persistence.DBOrTx(ctx, r.db).
		Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", lockKey).Error; err != nil {
		return apperrors.Wrap(err, "failed to acquire staff provision batch lock")
	}
	return nil
}

// FindReceiptsInScope returns receipt audits only for the authorized clinic_scope.
// Out-of-scope clinics are never queried, so existence cannot leak through this path.
func (r *staffProvisioningRepository) FindReceiptsInScope(
	ctx context.Context,
	clinicIDs []uint64,
	batchID string,
) ([]StaffProvisionReceipt, error) {
	if len(clinicIDs) == 0 {
		return nil, nil
	}
	type row struct {
		ClinicID uint64          `gorm:"column:clinic_id"`
		NewValue json.RawMessage `gorm:"column:new_value"`
	}
	var rows []row
	if err := persistence.DBOrTx(ctx, r.db).
		Table("audit_logs").
		Select("clinic_id, new_value").
		Where(
			"clinic_id IN ? AND action = ? AND resource = ? AND new_value->>'batch_id' = ?",
			clinicIDs,
			model.AuditActionStaffProvisionReceipt,
			model.AuditResourceStaffProvisionBatch,
			batchID,
		).
		Find(&rows).Error; err != nil {
		return nil, apperrors.FromGORM(err, "audit_log", "staff_provision_receipt")
	}

	out := make([]StaffProvisionReceipt, 0, len(rows))
	for _, item := range rows {
		var payload struct {
			BatchID string `json:"batch_id"`
			Digest  string `json:"digest"`
			Count   int    `json:"count"`
		}
		if len(item.NewValue) > 0 {
			if err := json.Unmarshal(item.NewValue, &payload); err != nil {
				return nil, apperrors.Wrap(err, "failed to decode staff provision receipt")
			}
		}
		out = append(out, StaffProvisionReceipt{
			ClinicID: item.ClinicID,
			BatchID:  payload.BatchID,
			Digest:   payload.Digest,
			Count:    payload.Count,
		})
	}
	return out, nil
}

func (r *staffProvisioningRepository) FindAccountByID(
	ctx context.Context,
	accountID uint64,
) (*model.Account, error) {
	var account model.Account
	if err := persistence.DBOrTx(ctx, r.db).
		First(&account, "id = ? AND deleted_at IS NULL", accountID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", accountID))
	}
	return &account, nil
}

func (r *staffProvisioningRepository) FindStaffByAccountID(
	ctx context.Context,
	accountID uint64,
) (*model.Staff, error) {
	var staff model.Staff
	db := persistence.DBOrTx(ctx, r.db)
	// Not-found is a normal branch for system-admin actors without a staff row;
	// keep the lookup silent so expected misses do not flood test/operator logs.
	result := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Limit(1).
		Find(&staff)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("account_id=%d", accountID))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("staff", fmt.Sprintf("account_id=%d", accountID))
	}
	return &staff, nil
}

func (r *staffProvisioningRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	db := persistence.DBOrTx(ctx, r.db)
	// Email is authentication PII — suppress SQL value logging for this lookup.
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Model(&model.Account{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "account", "email_lookup")
	}
	return count > 0, nil
}

func (r *staffProvisioningRepository) ClinicExists(ctx context.Context, clinicID uint64) (bool, error) {
	var count int64
	// clinics has no soft-delete column; active flag is the operational gate.
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Clinic{}).
		Where("id = ? AND is_active = TRUE", clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", clinicID))
	}
	return count > 0, nil
}

func (r *staffProvisioningRepository) OccupationBelongsToClinic(
	ctx context.Context,
	clinicID, occupationID uint64,
) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Occupation{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL AND is_active = TRUE", occupationID, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", occupationID))
	}
	return count > 0, nil
}

func (r *staffProvisioningRepository) PermissionGroupsBelongToClinic(
	ctx context.Context,
	clinicID uint64,
	groupIDs []uint64,
) error {
	if len(groupIDs) == 0 {
		return nil
	}
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.PermissionGroup{}).
		Where("clinic_id = ? AND deleted_at IS NULL AND id IN ?", clinicID, groupIDs).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, "permission_group", "")
	}
	if int(count) != len(groupIDs) {
		return apperrors.WrapInvalidInput("permission_group_ids contains invalid permission group")
	}
	return nil
}

func (r *staffProvisioningRepository) StaffAssignedToClinic(
	ctx context.Context,
	staffID, clinicID uint64,
) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	return count > 0, nil
}

// HasMasterStaffCreate reports whether the staff has master-staff create via
// any active permission group in the given clinic. Explicit rules only — no
// role inference.
func (r *staffProvisioningRepository) HasMasterStaffCreate(
	ctx context.Context,
	staffID, clinicID uint64,
) (bool, error) {
	var allowed bool
	err := persistence.DBOrTx(ctx, r.db).Raw(`
		SELECT COALESCE(bool_or(pgr.can_create), FALSE)
		FROM staff_permission_groups spg
		INNER JOIN permission_groups pg
			ON pg.id = spg.group_id
			AND pg.clinic_id = ?
			AND pg.deleted_at IS NULL
			AND pg.is_active = TRUE
		INNER JOIN permission_group_rules pgr
			ON pgr.group_id = pg.id
			AND pgr.deleted_at IS NULL
			AND pgr.resource = ?
		WHERE spg.staff_id = ?
	`, clinicID, string(model.ResourceMasterStaff), staffID).Scan(&allowed).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "permission_group_rule", "master-staff")
	}
	return allowed, nil
}

func (r *staffProvisioningRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "")
	}
	return nil
}

func (r *staffProvisioningRepository) CreateStaff(ctx context.Context, staff *model.Staff) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(staff).Error; err != nil {
		return apperrors.FromGORM(err, "staff", "")
	}
	return nil
}

func (r *staffProvisioningRepository) CreateAssignment(
	ctx context.Context,
	assignment *model.StaffClinicAssignment,
) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", "")
	}
	return nil
}

func (r *staffProvisioningRepository) AssignPermissionGroups(
	ctx context.Context,
	clinicID, staffID uint64,
	groupIDs []uint64,
) error {
	if len(groupIDs) == 0 {
		return nil
	}
	db := persistence.DBOrTx(ctx, r.db)
	if err := r.PermissionGroupsBelongToClinic(ctx, clinicID, groupIDs); err != nil {
		return err
	}
	// Ensure the main clinic assignment still exists inside the same transaction.
	var assignment model.StaffClinicAssignment
	if err := db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&assignment).Error; err != nil {
		return apperrors.FromGORM(
			err,
			"staff_clinic_assignment",
			fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID),
		)
	}
	rows := make([]model.StaffPermissionGroup, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		rows = append(rows, model.StaffPermissionGroup{StaffID: staffID, GroupID: groupID})
	}
	if err := db.Create(&rows).Error; err != nil {
		return apperrors.FromGORM(err, "staff_permission_group", fmt.Sprintf("staff:%d", staffID))
	}
	return nil
}

func (r *staffProvisioningRepository) LockOccupationForShare(
	ctx context.Context,
	clinicID, occupationID uint64,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("occupation lock requires an ambient transaction")
	}
	var occupation model.Occupation
	if err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "SHARE"}).
		Where(
			"id = ? AND clinic_id = ? AND deleted_at IS NULL AND is_active = TRUE",
			occupationID,
			clinicID,
		).
		First(&occupation).Error; err != nil {
		return apperrors.FromGORM(err, "occupation", fmt.Sprintf("%d", occupationID))
	}
	return nil
}

func (r *staffProvisioningRepository) WriteAudit(ctx context.Context, entry *model.AuditLog) error {
	tx := persistence.TxFromContext(ctx)
	if tx == nil {
		return apperrors.WrapInternalServerError("staff provision audit requires an ambient transaction")
	}
	if entry == nil {
		return apperrors.WrapInvalidInput("audit log is required")
	}
	if entry.ClinicID == nil || *entry.ClinicID == 0 {
		return apperrors.WrapInvalidInput("audit log clinic_id is required")
	}
	if entry.ActorType == "" || entry.Action == "" || entry.Resource == "" {
		return apperrors.WrapInvalidInput("audit log fields are incomplete")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	if err := tx.WithContext(ctx).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", "")
	}
	return nil
}
