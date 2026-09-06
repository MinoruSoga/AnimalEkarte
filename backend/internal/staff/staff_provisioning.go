package staff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const (
	// StaffProvisionSchemaVersion is the only accepted manifest schema.
	StaffProvisionSchemaVersion = "staff-provision-v1"
	// StaffProvisionBatchIDPrefix is the namespace prefix for batch_id values.
	StaffProvisionBatchIDPrefix = "staff-provision:"
	// staffProvisionUserAgent identifies CLI-originated audit rows.
	staffProvisionUserAgent = "staff-provision-cli"
)

// StaffProvisionManifest is the strict-decoded operator manifest.
// Unknown fields are rejected by the decoder.
type StaffProvisionManifest struct {
	SchemaVersion  string                     `json:"schema_version"`
	BatchID        string                     `json:"batch_id"`
	ClinicScope    []uint64                   `json:"clinic_scope"`
	ActorAccountID uint64                     `json:"actor_account_id"`
	Staff          []StaffProvisionStaffEntry `json:"staff"`
}

// StaffProvisionStaffEntry is one staff row in the provisioning manifest.
// Values are explicit: role→permission inference is intentionally unsupported.
type StaffProvisionStaffEntry struct {
	ExternalStaffID    string   `json:"external_staff_id"`
	Name               string   `json:"name"`
	Email              string   `json:"email"`
	MainClinicID       uint64   `json:"main_clinic_id"`
	ClinicIDs          []uint64 `json:"clinic_ids"`
	PermissionGroupIDs []uint64 `json:"permission_group_ids"`
	OccupationID       *uint64  `json:"occupation_id"`
	StaffType          string   `json:"staff_type"`
	IsActive           bool     `json:"is_active"`
	ReservationVisible bool     `json:"reservation_visible"`
	SecretRef          string   `json:"secret_ref"`
}

// StaffProvisionSecretsFile is the strict-decoded companion secret file.
// It contains only secret_ref → initial password pairs.
type StaffProvisionSecretsFile struct {
	Secrets []StaffProvisionSecretEntry `json:"secrets"`
}

// StaffProvisionSecretEntry binds one secret_ref to an initial password.
type StaffProvisionSecretEntry struct {
	SecretRef string `json:"secret_ref"`
	Password  string `json:"password"`
}

// StaffProvisionReceipt is a clinic-scoped, PII-free apply receipt.
type StaffProvisionReceipt struct {
	ClinicID uint64
	BatchID  string
	Digest   string
	Count    int
}

// StaffProvisionPreflightResult is the PII-free preflight outcome.
type StaffProvisionPreflightResult struct {
	BatchID        string   `json:"batch_id"`
	Digest         string   `json:"digest"`
	StaffCount     int      `json:"staff_count"`
	ClinicScope    []uint64 `json:"clinic_scope"`
	ActorAccountID uint64   `json:"actor_account_id"`
}

// StaffProvisionApplyResult is the PII-free apply outcome.
type StaffProvisionApplyResult struct {
	Status      string   `json:"status"` // applied | noop
	BatchID     string   `json:"batch_id"`
	Digest      string   `json:"digest"`
	StaffCount  int      `json:"staff_count"`
	ClinicScope []uint64 `json:"clinic_scope"`
}

// StaffProvisionActor is the resolved actor used for authorization and audit.
type StaffProvisionActor struct {
	AccountID     uint64
	IsSystemAdmin bool
	StaffID       *uint64
}

// StaffProvisioningRepository is StaffProvisioner's persistence port.
// The GORM store is concrete; this interface stays on the use-case side.
type StaffProvisioningRepository interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	AcquireBatchLock(ctx context.Context, batchID string) error
	FindReceiptsInScope(ctx context.Context, clinicIDs []uint64, batchID string) ([]StaffProvisionReceipt, error)
	FindAccountByID(ctx context.Context, accountID uint64) (*model.Account, error)
	FindStaffByAccountID(ctx context.Context, accountID uint64) (*model.Staff, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	ClinicExists(ctx context.Context, clinicID uint64) (bool, error)
	OccupationBelongsToClinic(ctx context.Context, clinicID, occupationID uint64) (bool, error)
	PermissionGroupsBelongToClinic(ctx context.Context, clinicID uint64, groupIDs []uint64) error
	StaffAssignedToClinic(ctx context.Context, staffID, clinicID uint64) (bool, error)
	HasMasterStaffCreate(ctx context.Context, staffID, clinicID uint64) (bool, error)
	CreateAccount(ctx context.Context, account *model.Account) error
	CreateStaff(ctx context.Context, staff *model.Staff) error
	CreateAssignment(ctx context.Context, assignment *model.StaffClinicAssignment) error
	AssignPermissionGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
	LockOccupationForShare(ctx context.Context, clinicID, occupationID uint64) error
	WriteAudit(ctx context.Context, entry *model.AuditLog) error
}

// StaffProvisioner runs preflight and apply for secret-managed staff batches.
type StaffProvisioner struct {
	repo      StaffProvisioningRepository
	repoRoots []string
}

// NewStaffProvisioner constructs a provisioner that rejects input paths under
// any of the given repository roots (absolute realpaths).
func NewStaffProvisioner(repo StaffProvisioningRepository, repoRoots []string) *StaffProvisioner {
	roots := make([]string, 0, len(repoRoots))
	for _, root := range repoRoots {
		if root == "" {
			continue
		}
		roots = append(roots, root)
	}
	return &StaffProvisioner{repo: repo, repoRoots: roots}
}

// DecodeStaffProvisionManifest strictly decodes a manifest and rejects unknown fields.
func DecodeStaffProvisionManifest(r io.Reader) (*StaffProvisionManifest, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	var manifest StaffProvisionManifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, apperrors.WrapInvalidInput("マニフェストの形式が正しくありません")
	}
	// Reject trailing content after the root object.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, apperrors.WrapInvalidInput("manifest must contain exactly one JSON value")
	}
	return &manifest, nil
}

// DecodeStaffProvisionSecrets strictly decodes a secrets file and rejects unknown fields.
func DecodeStaffProvisionSecrets(r io.Reader) (*StaffProvisionSecretsFile, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	var secrets StaffProvisionSecretsFile
	if err := decoder.Decode(&secrets); err != nil {
		return nil, apperrors.WrapInvalidInput("シークレットの形式が正しくありません")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, apperrors.WrapInvalidInput("secrets must contain exactly one JSON value")
	}
	return &secrets, nil
}

// ClinicScopeBatchID builds the only accepted batch_id for a sorted unique clinic_scope.
func ClinicScopeBatchID(clinicScope []uint64) string {
	return StaffProvisionBatchIDPrefix + clinicScopeDigestHex(clinicScope)
}

func clinicScopeDigestHex(clinicScope []uint64) string {
	parts := make([]string, len(clinicScope))
	for i, id := range clinicScope {
		parts[i] = strconv.FormatUint(id, 10)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(sum[:])
}

// ComputeStaffProvisionDigest returns a PII-free content digest for receipt idempotency.
// Passwords are never included; secret_ref is retained as a non-secret reference token.
func ComputeStaffProvisionDigest(manifest *StaffProvisionManifest) (string, error) {
	if manifest == nil {
		return "", apperrors.WrapInvalidInput("manifest is required")
	}
	type digestStaff struct {
		ExternalStaffID    string   `json:"external_staff_id"`
		MainClinicID       uint64   `json:"main_clinic_id"`
		ClinicIDs          []uint64 `json:"clinic_ids"`
		PermissionGroupIDs []uint64 `json:"permission_group_ids"`
		OccupationID       *uint64  `json:"occupation_id"`
		StaffType          string   `json:"staff_type"`
		IsActive           bool     `json:"is_active"`
		ReservationVisible bool     `json:"reservation_visible"`
		SecretRef          string   `json:"secret_ref"`
		// name/email intentionally omitted from digest payload surface in logs;
		// they are hashed so content changes still change the digest without
		// embedding raw PII into intermediate structures that might be logged.
		IdentityFingerprint string `json:"identity_fingerprint"`
	}
	type digestRoot struct {
		SchemaVersion  string        `json:"schema_version"`
		BatchID        string        `json:"batch_id"`
		ClinicScope    []uint64      `json:"clinic_scope"`
		ActorAccountID uint64        `json:"actor_account_id"`
		Staff          []digestStaff `json:"staff"`
	}
	staffCopy := append([]StaffProvisionStaffEntry(nil), manifest.Staff...)
	slices.SortFunc(staffCopy, func(a, b StaffProvisionStaffEntry) int {
		return strings.Compare(a.ExternalStaffID, b.ExternalStaffID)
	})
	entries := make([]digestStaff, 0, len(staffCopy))
	for _, s := range staffCopy {
		clinicIDs := append([]uint64(nil), s.ClinicIDs...)
		slices.Sort(clinicIDs)
		groupIDs := append([]uint64(nil), s.PermissionGroupIDs...)
		slices.Sort(groupIDs)
		identitySum := sha256.Sum256([]byte(s.Name + "\x00" + s.Email))
		entries = append(entries, digestStaff{
			ExternalStaffID:     s.ExternalStaffID,
			MainClinicID:        s.MainClinicID,
			ClinicIDs:           clinicIDs,
			PermissionGroupIDs:  groupIDs,
			OccupationID:        s.OccupationID,
			StaffType:           s.StaffType,
			IsActive:            s.IsActive,
			ReservationVisible:  s.ReservationVisible,
			SecretRef:           s.SecretRef,
			IdentityFingerprint: hex.EncodeToString(identitySum[:]),
		})
	}
	payload, err := json.Marshal(digestRoot{
		SchemaVersion:  manifest.SchemaVersion,
		BatchID:        manifest.BatchID,
		ClinicScope:    append([]uint64(nil), manifest.ClinicScope...),
		ActorAccountID: manifest.ActorAccountID,
		Staff:          entries,
	})
	if err != nil {
		return "", apperrors.Wrap(err, "failed to marshal provision digest")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

type receiptDecision string

const (
	receiptDecisionApply receiptDecision = "apply"
	receiptDecisionNoop  receiptDecision = "noop"
)
