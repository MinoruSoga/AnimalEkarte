package medicalrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CheckupPackageImportService is the clinic-scoped dry-run/apply boundary.
type CheckupPackageImportService interface {
	Preview(ctx context.Context, clinicID, actorID uint64, raw []byte) (*CheckupPackageImportOperatorReceipt, error)
	Apply(ctx context.Context, clinicID, actorID uint64, raw []byte) (*CheckupPackageImportOperatorReceipt, error)
}

// CheckupPackageImportOperatorReceipt is the external operator DTO (sink separation).
// Must not include actor/clinic/digest/resource mapping/before-after.
type CheckupPackageImportOperatorReceipt struct {
	ReceiptID     string `json:"receipt_id"`
	Result        string `json:"result"` // applied | noop | conflict | dry_run_ok
	TypesCreated  int    `json:"types_created"`
	FieldsCreated int    `json:"fields_created"`
}

type checkupPackageImportService struct {
	db         *gorm.DB
	transactor Transactor
	auditTx    AuditTxLogger
	// canImport is evaluated by the handler via permission middleware; service
	// re-checks actor belongs to clinic and is active inside the apply transaction.
}

func NewCheckupPackageImportService(db *gorm.DB, transactor Transactor, auditTx AuditTxLogger) CheckupPackageImportService {
	return &checkupPackageImportService{db: db, transactor: transactor, auditTx: auditTx}
}

func (s *checkupPackageImportService) Preview(ctx context.Context, clinicID, actorID uint64, raw []byte) (*CheckupPackageImportOperatorReceipt, error) {
	if clinicID == 0 || actorID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic and actor are required")
	}
	canonical, err := ParseAndCanonicalizeCheckupPackage(raw)
	if err != nil {
		return nil, err
	}
	if err := s.validateActorInClinic(ctx, clinicID, actorID); err != nil {
		return nil, err
	}
	// Dry-run: domain write zero. Collision checks are read-only.
	if err := s.preflightCollisions(ctx, clinicID, canonical); err != nil {
		return nil, err
	}
	receiptID := opaqueReceiptID()
	slog.InfoContext(ctx, "checkup package import dry-run ok",
		slog.String("receipt_id", receiptID),
		slog.String("result", "dry_run_ok"),
	)
	return &CheckupPackageImportOperatorReceipt{
		ReceiptID:     receiptID,
		Result:        "dry_run_ok",
		TypesCreated:  0,
		FieldsCreated: 0,
	}, nil
}

func (s *checkupPackageImportService) Apply(ctx context.Context, clinicID, actorID uint64, raw []byte) (*CheckupPackageImportOperatorReceipt, error) {
	if clinicID == 0 || actorID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic and actor are required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("checkup package import requires a transaction dependency")
	}
	canonical, err := ParseAndCanonicalizeCheckupPackage(raw)
	if err != nil {
		return nil, err
	}

	var operator *CheckupPackageImportOperatorReceipt
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		db := persistence.DBOrTx(txCtx, s.db)

		// Clinic row lock serializes concurrent apply for the same clinic.
		var clinic model.Clinic
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", clinicID).
			First(&clinic).Error; err != nil {
			return apperrors.FromGORM(err, "clinic", fmt.Sprintf("%d", clinicID))
		}
		if err := s.validateActorInClinic(txCtx, clinicID, actorID); err != nil {
			return err
		}

		// Existing receipt for same namespace/version?
		var existing model.CheckupPackageImportReceipt
		findErr := db.Where(
			"clinic_id = ? AND namespace = ? AND version = ?",
			clinicID, canonical.Manifest.Namespace, canonical.Manifest.Version,
		).First(&existing).Error
		if findErr == nil {
			if existing.ContentDigest == canonical.Digest {
				operator = &CheckupPackageImportOperatorReceipt{
					ReceiptID:     opaqueReceiptIDFromUint(existing.ID),
					Result:        "noop",
					TypesCreated:  0,
					FieldsCreated: 0,
				}
				slog.InfoContext(txCtx, "checkup package import noop",
					slog.String("receipt_id", operator.ReceiptID),
					slog.String("result", "noop"),
				)
				return nil
			}
			return apperrors.WrapConflict("checkup package version content conflict")
		}
		if findErr != nil && !apperrors.IsNotFound(apperrors.FromGORM(findErr, "checkup_package_import_receipt", "lookup")) {
			// distinguish not found
			if findErr != gorm.ErrRecordNotFound {
				return apperrors.FromGORM(findErr, "checkup_package_import_receipt", "lookup")
			}
		}

		if err := s.preflightCollisions(txCtx, clinicID, canonical); err != nil {
			return err
		}

		// Write types then fields in stable key order (already canonicalized).
		typeIDByKey := make(map[string]uint64, len(canonical.Manifest.Types))
		typesCreated := 0
		// Two-pass types: parents first (no parent_key), then children.
		var roots, children []CheckupPackageTypeDef
		for _, t := range canonical.Manifest.Types {
			if t.ParentKey == nil {
				roots = append(roots, t)
			} else {
				children = append(children, t)
			}
		}
		for _, pass := range [][]CheckupPackageTypeDef{roots, children} {
			for _, t := range pass {
				var parentID *uint64
				if t.ParentKey != nil {
					id, ok := typeIDByKey[*t.ParentKey]
					if !ok {
						return apperrors.WrapInvalidInput(fmt.Sprintf("parent type %q not resolved", *t.ParentKey))
					}
					parentID = &id
				}
				ns := canonical.Manifest.Namespace
				key := t.Key
				row := model.CheckupType{
					ClinicID:        clinicID,
					Name:            t.Name,
					IsActive:        t.IsActive,
					Description:     t.Description,
					Interval:        t.Interval,
					TargetAge:       t.TargetAge,
					ParentID:        parentID,
					ImportNamespace: &ns,
					ImportKey:       &key,
					SortOrder:       t.SortOrder,
				}
				if err := db.Create(&row).Error; err != nil {
					return apperrors.FromGORM(err, "checkup_type", t.Key)
				}
				typeIDByKey[t.Key] = row.ID
				typesCreated++
			}
		}

		fieldsCreated := 0
		fieldIDByKey := make(map[string]uint64, len(canonical.Manifest.Fields))
		for _, f := range canonical.Manifest.Fields {
			typeID, ok := typeIDByKey[f.TypeKey]
			if !ok {
				return apperrors.WrapInvalidInput(fmt.Sprintf("field type_key %q missing", f.TypeKey))
			}
			opts, err := json.Marshal(f.Options)
			if err != nil {
				return apperrors.Wrap(err, "marshal field options")
			}
			var minV, maxV *float64
			if f.MinValue != nil {
				v, err := strconv.ParseFloat(*f.MinValue, 64)
				if err != nil {
					return apperrors.WrapInvalidInput("invalid min_value")
				}
				minV = &v
			}
			if f.MaxValue != nil {
				v, err := strconv.ParseFloat(*f.MaxValue, 64)
				if err != nil {
					return apperrors.WrapInvalidInput("invalid max_value")
				}
				maxV = &v
			}
			ns := canonical.Manifest.Namespace
			key := f.Key
			row := model.CheckupTypeField{
				ClinicID:        clinicID,
				CheckupTypeID:   typeID,
				Name:            f.Name,
				FieldType:       model.CheckupFieldType(f.FieldType),
				Unit:            f.Unit,
				MinValue:        minV,
				MaxValue:        maxV,
				Options:         datatypes.JSON(opts),
				IsProvisional:   false,
				ImportNamespace: &ns,
				ImportKey:       &key,
				SortOrder:       f.SortOrder,
			}
			if err := db.Create(&row).Error; err != nil {
				return apperrors.FromGORM(err, "checkup_type_field", f.Key)
			}
			fieldIDByKey[f.Key] = row.ID
			fieldsCreated++
		}

		mapping := map[string]any{
			"types":  typeIDByKey,
			"fields": fieldIDByKey,
		}
		mappingJSON, err := json.Marshal(mapping)
		if err != nil {
			return apperrors.Wrap(err, "marshal resource mapping")
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
			return apperrors.FromGORM(err, "checkup_package_import_receipt", canonical.Manifest.Version)
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
				return err
			}
		}

		operator = &CheckupPackageImportOperatorReceipt{
			ReceiptID:     opaqueReceiptIDFromUint(receipt.ID),
			Result:        "applied",
			TypesCreated:  typesCreated,
			FieldsCreated: fieldsCreated,
		}
		slog.InfoContext(txCtx, "checkup package import applied",
			slog.String("receipt_id", operator.ReceiptID),
			slog.String("result", "applied"),
		)
		return nil
	}); err != nil {
		return nil, err
	}
	return operator, nil
}

func (s *checkupPackageImportService) validateActorInClinic(ctx context.Context, clinicID, actorID uint64) error {
	db := persistence.DBOrTx(ctx, s.db)
	var staff model.Staff
	if err := db.Where("id = ? AND clinic_id = ? AND is_active = ?", actorID, clinicID, true).
		First(&staff).Error; err != nil {
		// Same non-leaking surface for foreign/inactive actor.
		return apperrors.WrapNotFound("staff", fmt.Sprintf("%d", actorID))
	}
	return nil
}

func (s *checkupPackageImportService) preflightCollisions(ctx context.Context, clinicID uint64, canonical *CanonicalCheckupPackage) error {
	db := persistence.DBOrTx(ctx, s.db)
	ns := canonical.Manifest.Namespace

	// Stable-key collisions with different content → conflict
	for _, t := range canonical.Manifest.Types {
		var existing model.CheckupType
		err := db.Where(
			"clinic_id = ? AND import_namespace = ? AND import_key = ? AND deleted_at IS NULL",
			clinicID, ns, t.Key,
		).First(&existing).Error
		if err == nil {
			return apperrors.WrapConflict(fmt.Sprintf("type stable key %q already imported", t.Key))
		}
		if err != gorm.ErrRecordNotFound {
			return apperrors.FromGORM(err, "checkup_type", t.Key)
		}
		// Name collision (active)
		var byName model.CheckupType
		err = db.Where(
			"clinic_id = ? AND name = ? AND deleted_at IS NULL",
			clinicID, t.Name,
		).First(&byName).Error
		if err == nil {
			return apperrors.WrapConflict(fmt.Sprintf("type name %q already exists", t.Name))
		}
		if err != gorm.ErrRecordNotFound {
			return apperrors.FromGORM(err, "checkup_type", t.Name)
		}
	}
	for _, f := range canonical.Manifest.Fields {
		var existing model.CheckupTypeField
		err := db.Where(
			"clinic_id = ? AND import_namespace = ? AND import_key = ? AND deleted_at IS NULL",
			clinicID, ns, f.Key,
		).First(&existing).Error
		if err == nil {
			return apperrors.WrapConflict(fmt.Sprintf("field stable key %q already imported", f.Key))
		}
		if err != gorm.ErrRecordNotFound {
			return apperrors.FromGORM(err, "checkup_type_field", f.Key)
		}
	}
	return nil
}

func opaqueReceiptID() string {
	return "rcp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func opaqueReceiptIDFromUint(id uint64) string {
	// Stable opaque-ish token without embedding clinic/actor/digest.
	return fmt.Sprintf("rcp_%016x", id)
}
