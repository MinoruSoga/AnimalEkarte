package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func (s *checkupPackageImportService) applyCheckupPackageInTx(
	txCtx context.Context,
	clinicID, actorID uint64,
	canonical *CanonicalCheckupPackage,
) (*CheckupPackageImportOperatorReceipt, error) {
	db := persistence.DBOrTx(txCtx, s.db)

	var clinic model.Clinic
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", clinicID).
		First(&clinic).Error; err != nil {
		return nil, apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", clinicID))
	}
	if err := s.validateActorInClinic(txCtx, clinicID, actorID); err != nil {
		return nil, err
	}

	var existing model.CheckupPackageImportReceipt
	findErr := db.Where(
		"clinic_id = ? AND namespace = ? AND version = ?",
		clinicID, canonical.Manifest.Namespace, canonical.Manifest.Version,
	).First(&existing).Error
	if findErr == nil {
		if existing.ContentDigest == canonical.Digest {
			operator := &CheckupPackageImportOperatorReceipt{
				ReceiptID:     opaqueReceiptIDFromUint(existing.ID),
				Result:        "noop",
				TypesCreated:  0,
				FieldsCreated: 0,
			}
			slog.InfoContext(txCtx, "checkup package import noop",
				slog.String("receipt_id", operator.ReceiptID),
				slog.String("result", "noop"),
			)
			return operator, nil
		}
		return nil, apperrors.WrapConflict("checkup package version content conflict")
	}
	if findErr != nil && !apperrors.IsNotFound(apperrors.FromGORM(findErr, "checkup_package_import_receipt", "lookup")) {
		if !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, apperrors.FromGORM(findErr, "checkup_package_import_receipt", "lookup")
		}
	}

	if err := s.preflightCollisions(txCtx, clinicID, canonical); err != nil {
		return nil, err
	}

	typeIDByKey, typesCreated, err := s.importCheckupTypes(db, clinicID, canonical)
	if err != nil {
		return nil, err
	}
	fieldIDByKey, fieldsCreated, err := s.importCheckupFields(db, clinicID, canonical, typeIDByKey)
	if err != nil {
		return nil, err
	}

	mapping := map[string]any{
		"types":  typeIDByKey,
		"fields": fieldIDByKey,
	}
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return nil, apperrors.Wrap(err, "marshal resource mapping")
	}

	receipt := model.CheckupPackageImportReceipt{
		ClinicID:            clinicID,
		Namespace:           canonical.Manifest.Namespace,
		Version:             canonical.Manifest.Version,
		ContentDigest:       canonical.Digest,
		Status:              model.CheckupPackageImportStatusApplied,
		ActorID:             actorID,
		TypesCreated:        typesCreated,
		FieldsCreated:       fieldsCreated,
		ResourceMapping:     mappingJSON,
		ClinicalApprovalRef: canonical.Manifest.ClinicalApprovalRef,
	}
	if err := db.Create(&receipt).Error; err != nil {
		return nil, apperrors.FromGORM(err, "checkup_package_import_receipt", canonical.Manifest.Version)
	}

	if s.auditTx != nil {
		clinicIDCopy := clinicID
		receiptID := receipt.ID
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicIDCopy,
			ActorID:    &actorID,
			Action:     model.AuditActionCheckupPackageImportApply,
			Resource:   model.AuditResourceCheckupPackageImport,
			ResourceID: &receiptID,
			NewValue: map[string]any{
				"namespace":      receipt.Namespace,
				"version":        receipt.Version,
				"types_created":  typesCreated,
				"fields_created": fieldsCreated,
				"status":         string(receipt.Status),
			},
		}); err != nil {
			return nil, err
		}
	}

	operator := &CheckupPackageImportOperatorReceipt{
		ReceiptID:     opaqueReceiptIDFromUint(receipt.ID),
		Result:        "applied",
		TypesCreated:  typesCreated,
		FieldsCreated: fieldsCreated,
	}
	slog.InfoContext(txCtx, "checkup package import applied",
		slog.String("receipt_id", operator.ReceiptID),
		slog.String("result", "applied"),
	)
	return operator, nil
}
