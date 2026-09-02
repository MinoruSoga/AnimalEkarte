// Command stg-uat-staff-attach links synthetic UAT accounts onto existing staffs
// rows without inserting staffs. Operator output is digest/count/ids only.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (r *gormAttachRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(persistence.WithTxValue(ctx, tx))
	}); err != nil {
		return apperrors.Wrap(err, "staff attach transaction failed")
	}
	return nil
}

func (r *gormAttachRepository) FindStaffByID(ctx context.Context, id uint64) (*model.Staff, error) {
	var row model.Staff
	if err := persistence.DBOrTx(ctx, r.db).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&row).Error; err != nil {
		return nil, apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", id))
	}
	return &row, nil
}

func (r *gormAttachRepository) FindAccountByID(ctx context.Context, accountID uint64) (*model.Account, error) {
	var account model.Account
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormlogger.Silent)}).
		First(&account, "id = ? AND deleted_at IS NULL", accountID).Error; err != nil {
		return nil, apperrors.FromGORM(err, "account", fmt.Sprintf("%d", accountID))
	}
	return &account, nil
}

func (r *gormAttachRepository) PermissionGroupsBelongToClinic(
	ctx context.Context,
	clinicID uint64,
	groupIDs []uint64,
) error {
	if len(groupIDs) == 0 {
		return apperrors.WrapInvalidInput("permission_group_ids must not be empty")
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

func (r *gormAttachRepository) CreateAccount(ctx context.Context, account *model.Account) error {
	db := persistence.DBOrTx(ctx, r.db)
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormlogger.Silent)}).
		Create(account).Error; err != nil {
		return apperrors.FromGORM(err, "account", "")
	}
	return nil
}

func (r *gormAttachRepository) UpdateStaffAccount(
	ctx context.Context,
	staffID, clinicID, accountID uint64,
	setActive bool,
) error {
	db := persistence.DBOrTx(ctx, r.db)
	var row model.Staff
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		First(&row).Error; err != nil {
		return apperrors.FromGORM(err, "staff", fmt.Sprintf("%d", staffID))
	}
	if row.AccountID != nil {
		return apperrors.WrapConflict("staff is already linked")
	}
	updates := map[string]any{"account_id": accountID}
	if setActive {
		updates["is_active"] = true
	}
	result := db.Model(&model.Staff{}).
		Where("id = ? AND clinic_id = ? AND deleted_at IS NULL AND account_id IS NULL", staffID, clinicID).
		Updates(updates)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "staff", fmt.Sprintf("%d", staffID))
	}
	if result.RowsAffected != 1 {
		return apperrors.WrapConflict("staff account attach did not update exactly one row")
	}
	return nil
}

func (r *gormAttachRepository) EnsureClinicAssignment(ctx context.Context, staffID, clinicID uint64) error {
	db := persistence.DBOrTx(ctx, r.db)
	var count int64
	if err := db.Model(&model.StaffClinicAssignment{}).
		Where("staff_id = ? AND clinic_id = ? AND deleted_at IS NULL", staffID, clinicID).
		Count(&count).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	if count > 0 {
		return nil
	}
	assignment := &model.StaffClinicAssignment{
		StaffID:  staffID,
		ClinicID: clinicID,
		IsMain:   true,
	}
	if err := db.Create(assignment).Error; err != nil {
		return apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("staff=%d,clinic=%d", staffID, clinicID))
	}
	return nil
}

func (r *gormAttachRepository) AssignPermissionGroups(
	ctx context.Context,
	clinicID, staffID uint64,
	groupIDs []uint64,
) error {
	if err := r.PermissionGroupsBelongToClinic(ctx, clinicID, groupIDs); err != nil {
		return err
	}
	db := persistence.DBOrTx(ctx, r.db)
	var assignment model.StaffClinicAssignment
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
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

func (r *gormAttachRepository) LastAttachDigest(ctx context.Context, staffID uint64) (string, error) {
	var entry model.AuditLog
	err := persistence.DBOrTx(ctx, r.db).
		Where("resource = ? AND resource_id = ? AND action = ?", model.AuditResourceStaff, staffID, attachAuditAction).
		Order("id DESC").
		First(&entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", apperrors.FromGORM(err, "audit_log", fmt.Sprintf("staff:%d", staffID))
	}
	return parseAttachDigestReceipt(entry.NewValue)
}

func (r *gormAttachRepository) SaveAttachDigest(ctx context.Context, staffID uint64, digest string) error {
	staffRow, err := r.FindStaffByID(ctx, staffID)
	if err != nil {
		return err
	}
	clinicID := staffRow.ClinicID
	payload, err := json.Marshal(attachDigestReceipt{Digest: digest})
	if err != nil {
		return apperrors.Wrap(err, "failed to marshal attach digest receipt")
	}
	entry := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorType:  model.AuditActorTypeSystem,
		Action:     attachAuditAction,
		Resource:   model.AuditResourceStaff,
		ResourceID: &staffID,
		NewValue:   payload,
		UserAgent:  attachAuditUserAgent,
	}
	if err := persistence.DBOrTx(ctx, r.db).Create(entry).Error; err != nil {
		return apperrors.FromGORM(err, "audit_log", fmt.Sprintf("staff:%d", staffID))
	}
	return nil
}

func parseAttachDigestReceipt(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "", nil
	}
	var receipt attachDigestReceipt
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return "", apperrors.WrapInvalidInput("attach digest receipt is unreadable")
	}
	return receipt.Digest, nil
}
