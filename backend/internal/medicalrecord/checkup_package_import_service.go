package medicalrecord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
		receipt, err := s.applyCheckupPackageInTx(txCtx, clinicID, actorID, canonical)
		if err != nil {
			return err
		}
		operator = receipt
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
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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
		if !errors.Is(err, gorm.ErrRecordNotFound) {
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
