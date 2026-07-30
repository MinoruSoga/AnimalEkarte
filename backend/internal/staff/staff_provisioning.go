package staff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sys/unix"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
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

// StaffProvisioningRepository is the persistence boundary for batch provisioning.
// Implementations must honor ambient transaction participation.
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
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("manifest decode failed: %v", err))
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
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets decode failed: %v", err))
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

// ValidateSecureInputPath requires an absolute realpath, mode 0600 regular file,
// and a location outside every configured repository root.
func ValidateSecureInputPath(path string, repoRoots []string) (string, error) {
	if path == "" {
		return "", apperrors.WrapInvalidInput("input path is required")
	}
	if !filepath.IsAbs(path) {
		return "", apperrors.WrapInvalidInput("input path must be absolute")
	}
	clean := filepath.Clean(path)
	info, err := os.Lstat(clean)
	if err != nil {
		return "", apperrors.WrapInvalidInput(fmt.Sprintf("input path is not accessible: %v", err))
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", apperrors.WrapInvalidInput("input path must not be a symlink")
	}
	if !info.Mode().IsRegular() {
		return "", apperrors.WrapInvalidInput("input path must be a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		return "", apperrors.WrapInvalidInput("input path mode must be 0600")
	}
	realPath, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return "", apperrors.WrapInvalidInput(fmt.Sprintf("input path realpath resolution failed: %v", err))
	}
	realPath = filepath.Clean(realPath)
	for _, root := range repoRoots {
		if root == "" {
			continue
		}
		realRoot, rootErr := filepath.EvalSymlinks(filepath.Clean(root))
		if rootErr != nil {
			realRoot = filepath.Clean(root)
		}
		if pathIsInsideRoot(realPath, realRoot) {
			return "", apperrors.WrapInvalidInput("input path must be outside the repository")
		}
	}
	return realPath, nil
}

func pathIsInsideRoot(path, root string) bool {
	if path == root {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(path, strings.TrimRight(root, sep)+sep)
}

// OpenSecureInputFile validates path policy and opens the file read-only without following symlinks.
func OpenSecureInputFile(path string, repoRoots []string) (*os.File, string, error) {
	realPath, err := ValidateSecureInputPath(path, repoRoots)
	if err != nil {
		return nil, "", err
	}
	fd, err := unix.Open(realPath, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, "", apperrors.WrapInvalidInput(fmt.Sprintf("failed to open input path: %v", err))
	}
	return os.NewFile(uintptr(fd), realPath), realPath, nil
}

// LoadStaffProvisionInputs loads and structurally validates both operator files.
// It performs no database writes.
func LoadStaffProvisionInputs(
	manifestPath, secretsPath string,
	repoRoots []string,
) (*StaffProvisionManifest, map[string]string, string, error) {
	manifestFile, _, err := OpenSecureInputFile(manifestPath, repoRoots)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err, "manifest input rejected")
	}
	defer func() { _ = manifestFile.Close() }()

	secretsFile, _, err := OpenSecureInputFile(secretsPath, repoRoots)
	if err != nil {
		return nil, nil, "", apperrors.Wrap(err, "secrets input rejected")
	}
	defer func() { _ = secretsFile.Close() }()

	manifest, err := DecodeStaffProvisionManifest(manifestFile)
	if err != nil {
		return nil, nil, "", err
	}
	secrets, err := DecodeStaffProvisionSecrets(secretsFile)
	if err != nil {
		return nil, nil, "", err
	}
	if err := ValidateStaffProvisionManifestStructure(manifest); err != nil {
		return nil, nil, "", err
	}
	secretMap, err := ValidateStaffProvisionSecrets(manifest, secrets)
	if err != nil {
		return nil, nil, "", err
	}
	digest, err := ComputeStaffProvisionDigest(manifest)
	if err != nil {
		return nil, nil, "", err
	}
	return manifest, secretMap, digest, nil
}

// ValidateStaffProvisionManifestStructure validates structural invariants before any DB access.
func ValidateStaffProvisionManifestStructure(manifest *StaffProvisionManifest) error {
	if manifest == nil {
		return apperrors.WrapInvalidInput("manifest is required")
	}
	if manifest.SchemaVersion != StaffProvisionSchemaVersion {
		return apperrors.WrapInvalidInput("unsupported schema_version")
	}
	if manifest.ActorAccountID == 0 {
		return apperrors.WrapInvalidInput("actor_account_id is required")
	}
	if len(manifest.ClinicScope) == 0 {
		return apperrors.WrapInvalidInput("clinic_scope must not be empty")
	}
	if !slices.IsSorted(manifest.ClinicScope) {
		return apperrors.WrapInvalidInput("clinic_scope must be sorted ascending")
	}
	seenClinic := make(map[uint64]struct{}, len(manifest.ClinicScope))
	for _, id := range manifest.ClinicScope {
		if id == 0 {
			return apperrors.WrapInvalidInput("clinic_scope must not contain zero")
		}
		if _, dup := seenClinic[id]; dup {
			return apperrors.WrapInvalidInput("clinic_scope must not contain duplicates")
		}
		seenClinic[id] = struct{}{}
	}
	expectedBatchID := ClinicScopeBatchID(manifest.ClinicScope)
	if manifest.BatchID != expectedBatchID {
		return apperrors.WrapInvalidInput("batch_id does not match clinic_scope digest namespace")
	}
	if len(manifest.Staff) == 0 {
		return apperrors.WrapInvalidInput("staff must not be empty")
	}

	union := make(map[uint64]struct{})
	seenExternal := make(map[string]struct{}, len(manifest.Staff))
	seenEmail := make(map[string]struct{}, len(manifest.Staff))
	seenSecretRef := make(map[string]struct{}, len(manifest.Staff))

	for i, entry := range manifest.Staff {
		prefix := fmt.Sprintf("staff[%d]", i)
		externalID := strings.TrimSpace(entry.ExternalStaffID)
		if externalID == "" {
			return apperrors.WrapInvalidInput(prefix + ": external_staff_id is required")
		}
		if _, dup := seenExternal[externalID]; dup {
			return apperrors.WrapInvalidInput("duplicate external_staff_id in batch")
		}
		seenExternal[externalID] = struct{}{}

		name := strings.TrimSpace(entry.Name)
		if name == "" {
			return apperrors.WrapInvalidInput(prefix + ": name is required")
		}
		if utf8.RuneCountInString(name) > 100 {
			return apperrors.WrapInvalidInput(prefix + ": name is too long")
		}

		email := strings.TrimSpace(strings.ToLower(entry.Email))
		if email == "" || !strings.Contains(email, "@") {
			return apperrors.WrapInvalidInput(prefix + ": email is invalid")
		}
		if _, dup := seenEmail[email]; dup {
			return apperrors.WrapInvalidInput("duplicate email in batch")
		}
		seenEmail[email] = struct{}{}

		if entry.MainClinicID == 0 {
			return apperrors.WrapInvalidInput(prefix + ": main_clinic_id is required")
		}
		if len(entry.ClinicIDs) == 0 {
			return apperrors.WrapInvalidInput(prefix + ": clinic_ids must not be empty")
		}
		seenAssignment := make(map[uint64]struct{}, len(entry.ClinicIDs))
		mainFound := false
		for _, clinicID := range entry.ClinicIDs {
			if clinicID == 0 {
				return apperrors.WrapInvalidInput(prefix + ": clinic_ids must not contain zero")
			}
			if _, dup := seenAssignment[clinicID]; dup {
				return apperrors.WrapInvalidInput(prefix + ": clinic_ids must not contain duplicates")
			}
			seenAssignment[clinicID] = struct{}{}
			if _, inScope := seenClinic[clinicID]; !inScope {
				return apperrors.WrapInvalidInput(prefix + ": clinic_ids must be subset of clinic_scope")
			}
			if clinicID == entry.MainClinicID {
				mainFound = true
			}
			union[clinicID] = struct{}{}
		}
		if !mainFound {
			return apperrors.WrapInvalidInput(prefix + ": main_clinic_id must be included in clinic_ids")
		}
		if _, inScope := seenClinic[entry.MainClinicID]; !inScope {
			return apperrors.WrapInvalidInput(prefix + ": main_clinic_id must be in clinic_scope")
		}

		seenGroup := make(map[uint64]struct{}, len(entry.PermissionGroupIDs))
		for _, groupID := range entry.PermissionGroupIDs {
			if groupID == 0 {
				return apperrors.WrapInvalidInput(prefix + ": permission_group_ids must not contain zero")
			}
			if _, dup := seenGroup[groupID]; dup {
				return apperrors.WrapInvalidInput(prefix + ": permission_group_ids must not contain duplicates")
			}
			seenGroup[groupID] = struct{}{}
		}
		if entry.OccupationID != nil && *entry.OccupationID == 0 {
			return apperrors.WrapInvalidInput(prefix + ": occupation_id must not be zero when set")
		}
		if err := validateStaffType(entry.StaffType); err != nil {
			return apperrors.Wrap(err, prefix)
		}
		secretRef := strings.TrimSpace(entry.SecretRef)
		if secretRef == "" {
			return apperrors.WrapInvalidInput(prefix + ": secret_ref is required")
		}
		if _, dup := seenSecretRef[secretRef]; dup {
			return apperrors.WrapInvalidInput("duplicate secret_ref in batch")
		}
		seenSecretRef[secretRef] = struct{}{}
	}

	if len(union) != len(manifest.ClinicScope) {
		return apperrors.WrapInvalidInput("clinic_scope must equal the union of all staff main/assignment clinics")
	}
	for clinicID := range seenClinic {
		if _, ok := union[clinicID]; !ok {
			return apperrors.WrapInvalidInput("clinic_scope must equal the union of all staff main/assignment clinics")
		}
	}
	return nil
}

// ValidateStaffProvisionSecrets ensures one-to-one secret_ref coverage and password policy.
func ValidateStaffProvisionSecrets(
	manifest *StaffProvisionManifest,
	secrets *StaffProvisionSecretsFile,
) (map[string]string, error) {
	if secrets == nil {
		return nil, apperrors.WrapInvalidInput("secrets are required")
	}
	if len(secrets.Secrets) == 0 {
		return nil, apperrors.WrapInvalidInput("secrets must not be empty")
	}
	required := make(map[string]struct{}, len(manifest.Staff))
	for _, entry := range manifest.Staff {
		required[strings.TrimSpace(entry.SecretRef)] = struct{}{}
	}
	out := make(map[string]string, len(secrets.Secrets))
	for i, secret := range secrets.Secrets {
		ref := strings.TrimSpace(secret.SecretRef)
		if ref == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets[%d]: secret_ref is required", i))
		}
		if _, exists := out[ref]; exists {
			return nil, apperrors.WrapInvalidInput("duplicate secret_ref in secrets file")
		}
		if _, ok := required[ref]; !ok {
			return nil, apperrors.WrapInvalidInput("secrets file contains unreferenced secret_ref")
		}
		if err := validatePassword(secret.Password); err != nil {
			// Do not echo password content; only structural failure.
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets[%d]: password policy failed", i))
		}
		out[ref] = secret.Password
	}
	if len(out) != len(required) {
		return nil, apperrors.WrapInvalidInput("secrets file must cover every staff secret_ref exactly once")
	}
	return out, nil
}

// Preflight validates inputs and DB references with zero writes.
// Receipt comparison is intentionally NOT performed here — it runs inside apply
// under the batch advisory lock after authorization succeeds.
func (p *StaffProvisioner) Preflight(
	ctx context.Context,
	manifestPath, secretsPath string,
) (*StaffProvisionPreflightResult, error) {
	if p == nil || p.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff provisioner is not configured")
	}
	manifest, secrets, digest, err := LoadStaffProvisionInputs(manifestPath, secretsPath, p.repoRoots)
	if err != nil {
		return nil, err
	}
	if _, err := p.authorizeAndValidateRefs(ctx, manifest); err != nil {
		return nil, err
	}
	// Standalone preflight also checks email availability so operators learn about
	// collisions before apply. Apply itself re-checks under the batch lock after
	// receipt comparison so complete same-digest re-runs can no-op safely.
	if err := p.validateEmailsAvailable(ctx, manifest); err != nil {
		return nil, err
	}
	_ = secrets
	return &StaffProvisionPreflightResult{
		BatchID:        manifest.BatchID,
		Digest:         digest,
		StaffCount:     len(manifest.Staff),
		ClinicScope:    append([]uint64(nil), manifest.ClinicScope...),
		ActorAccountID: manifest.ActorAccountID,
	}, nil
}

// Apply authorizes, then under a batch lock compares receipts and either no-ops
// or creates the full batch atomically. Email uniqueness is enforced only on the
// create path so complete same-digest re-apply is a pure no-op.
func (p *StaffProvisioner) Apply(
	ctx context.Context,
	manifestPath, secretsPath string,
) (*StaffProvisionApplyResult, error) {
	if p == nil || p.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff provisioner is not configured")
	}
	manifest, secrets, digest, err := LoadStaffProvisionInputs(manifestPath, secretsPath, p.repoRoots)
	if err != nil {
		return nil, err
	}
	// Authorization and FK checks happen BEFORE receipt comparison (packet contract).
	// Email availability is deferred until after receipt decision so idempotent
	// complete batches no-op even though accounts already exist.
	actor, err := p.authorizeAndValidateRefs(ctx, manifest)
	if err != nil {
		return nil, err
	}

	var result *StaffProvisionApplyResult
	if err := p.repo.WithTx(ctx, func(txCtx context.Context) error {
		if lockErr := p.repo.AcquireBatchLock(txCtx, manifest.BatchID); lockErr != nil {
			return lockErr
		}
		receipts, receiptErr := p.repo.FindReceiptsInScope(txCtx, manifest.ClinicScope, manifest.BatchID)
		if receiptErr != nil {
			return receiptErr
		}
		decision, decideErr := decideReceiptState(manifest.ClinicScope, manifest.BatchID, digest, receipts)
		if decideErr != nil {
			return decideErr
		}
		if decision == receiptDecisionNoop {
			result = &StaffProvisionApplyResult{
				Status:      "noop",
				BatchID:     manifest.BatchID,
				Digest:      digest,
				StaffCount:  len(manifest.Staff),
				ClinicScope: append([]uint64(nil), manifest.ClinicScope...),
			}
			return nil
		}

		if emailErr := p.validateEmailsAvailable(txCtx, manifest); emailErr != nil {
			return emailErr
		}

		createdCount, createErr := p.createAllStaff(txCtx, manifest, secrets, actor, digest)
		if createErr != nil {
			return createErr
		}
		if writeErr := p.writeReceipts(txCtx, manifest, digest, createdCount, actor); writeErr != nil {
			return writeErr
		}
		result = &StaffProvisionApplyResult{
			Status:      "applied",
			BatchID:     manifest.BatchID,
			Digest:      digest,
			StaffCount:  createdCount,
			ClinicScope: append([]uint64(nil), manifest.ClinicScope...),
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

type receiptDecision string

const (
	receiptDecisionApply receiptDecision = "apply"
	receiptDecisionNoop  receiptDecision = "noop"
)

func decideReceiptState(
	clinicScope []uint64,
	batchID, digest string,
	receipts []StaffProvisionReceipt,
) (receiptDecision, error) {
	byClinic := make(map[uint64][]StaffProvisionReceipt, len(clinicScope))
	for _, receipt := range receipts {
		// Defense in depth: never consider out-of-scope rows even if a buggy
		// repository returned them. Existence outside scope must not affect
		// decisions or error text.
		if !slices.Contains(clinicScope, receipt.ClinicID) {
			continue
		}
		if receipt.BatchID != "" && receipt.BatchID != batchID {
			continue
		}
		byClinic[receipt.ClinicID] = append(byClinic[receipt.ClinicID], receipt)
	}

	matched := 0
	for _, clinicID := range clinicScope {
		clinicReceipts := byClinic[clinicID]
		if len(clinicReceipts) == 0 {
			continue
		}
		digests := make(map[string]struct{}, len(clinicReceipts))
		for _, receipt := range clinicReceipts {
			if receipt.Digest == "" {
				return "", apperrors.WrapConflict("staff provision batch receipt is incomplete")
			}
			digests[receipt.Digest] = struct{}{}
		}
		if len(digests) != 1 {
			return "", apperrors.WrapConflict("staff provision batch receipt digest mismatch")
		}
		var only string
		for d := range digests {
			only = d
		}
		if only != digest {
			return "", apperrors.WrapConflict("staff provision batch receipt digest mismatch")
		}
		matched++
	}

	switch {
	case matched == 0:
		return receiptDecisionApply, nil
	case matched == len(clinicScope):
		return receiptDecisionNoop, nil
	default:
		// Partial completion is always a conflict — do not resume mid-batch.
		return "", apperrors.WrapConflict("staff provision batch receipt is incomplete")
	}
}

// authorizeAndValidateRefs validates actor authorization and FK references.
// It performs no writes and does not inspect receipts (existence must not leak
// to unauthorized callers — authorization runs first).
func (p *StaffProvisioner) authorizeAndValidateRefs(
	ctx context.Context,
	manifest *StaffProvisionManifest,
) (*StaffProvisionActor, error) {
	for _, clinicID := range manifest.ClinicScope {
		exists, err := p.repo.ClinicExists(ctx, clinicID)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to verify clinic")
		}
		if !exists {
			return nil, apperrors.WrapInvalidInput("clinic_scope references unknown clinic")
		}
	}

	actorAccount, err := p.repo.FindAccountByID(ctx, manifest.ActorAccountID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, apperrors.WrapInvalidInput("actor_account_id is invalid")
		}
		return nil, apperrors.Wrap(err, "failed to load actor account")
	}
	if !actorAccount.IsActive {
		return nil, apperrors.WrapForbidden("actor account is not active")
	}

	actor := &StaffProvisionActor{
		AccountID:     actorAccount.ID,
		IsSystemAdmin: actorAccount.IsSystemAdmin,
	}
	staff, staffErr := p.repo.FindStaffByAccountID(ctx, actorAccount.ID)
	if staffErr != nil && !apperrors.IsNotFound(staffErr) {
		return nil, apperrors.Wrap(staffErr, "failed to load actor staff")
	}
	if staff != nil {
		id := staff.ID
		actor.StaffID = &id
	}

	if !actor.IsSystemAdmin {
		if actor.StaffID == nil {
			// Do not reveal whether receipts exist for unauthorized callers.
			return nil, apperrors.WrapForbidden("actor is not authorized for clinic_scope")
		}
		for _, clinicID := range manifest.ClinicScope {
			assigned, assignErr := p.repo.StaffAssignedToClinic(ctx, *actor.StaffID, clinicID)
			if assignErr != nil {
				return nil, apperrors.Wrap(assignErr, "failed to verify actor clinic membership")
			}
			if !assigned {
				return nil, apperrors.WrapForbidden("actor is not authorized for clinic_scope")
			}
			allowed, permErr := p.repo.HasMasterStaffCreate(ctx, *actor.StaffID, clinicID)
			if permErr != nil {
				return nil, apperrors.Wrap(permErr, "failed to verify actor permissions")
			}
			if !allowed {
				return nil, apperrors.WrapForbidden("actor is not authorized for clinic_scope")
			}
		}
	}

	for i, entry := range manifest.Staff {
		prefix := fmt.Sprintf("staff[%d]", i)
		if entry.OccupationID != nil {
			ok, occErr := p.repo.OccupationBelongsToClinic(ctx, entry.MainClinicID, *entry.OccupationID)
			if occErr != nil {
				return nil, apperrors.Wrap(occErr, "failed to verify occupation")
			}
			if !ok {
				return nil, apperrors.WrapInvalidInput(prefix + ": occupation_id is invalid for main_clinic_id")
			}
		}
		if len(entry.PermissionGroupIDs) > 0 {
			if groupErr := p.repo.PermissionGroupsBelongToClinic(
				ctx,
				entry.MainClinicID,
				entry.PermissionGroupIDs,
			); groupErr != nil {
				return nil, apperrors.WrapInvalidInput(prefix + ": permission_group_ids are invalid for main_clinic_id")
			}
		}
	}
	return actor, nil
}

func (p *StaffProvisioner) validateEmailsAvailable(
	ctx context.Context,
	manifest *StaffProvisionManifest,
) error {
	for _, entry := range manifest.Staff {
		email := strings.TrimSpace(strings.ToLower(entry.Email))
		exists, emailErr := p.repo.EmailExists(ctx, email)
		if emailErr != nil {
			return apperrors.Wrap(emailErr, "failed to check email uniqueness")
		}
		if exists {
			// Do not echo the email address in the error surface.
			return apperrors.WrapAlreadyExists("account", "email")
		}
	}
	return nil
}

func (p *StaffProvisioner) createAllStaff(
	ctx context.Context,
	manifest *StaffProvisionManifest,
	secrets map[string]string,
	actor *StaffProvisionActor,
	digest string,
) (int, error) {
	// Sort for deterministic insert order (reduces deadlock risk on unique indexes).
	staffCopy := append([]StaffProvisionStaffEntry(nil), manifest.Staff...)
	slices.SortFunc(staffCopy, func(a, b StaffProvisionStaffEntry) int {
		return strings.Compare(a.ExternalStaffID, b.ExternalStaffID)
	})

	created := 0
	for _, entry := range staffCopy {
		password := secrets[strings.TrimSpace(entry.SecretRef)]
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
		if hashErr != nil {
			return 0, apperrors.Wrap(hashErr, "failed to hash password")
		}
		// Drop password material immediately after hashing.
		password = ""

		if entry.OccupationID != nil {
			if err := p.repo.LockOccupationForShare(ctx, entry.MainClinicID, *entry.OccupationID); err != nil {
				return 0, err
			}
		}

		account := &model.Account{
			Email:        strings.TrimSpace(strings.ToLower(entry.Email)),
			PasswordHash: string(hashed),
			IsActive:     true,
		}
		if err := p.repo.CreateAccount(ctx, account); err != nil {
			return 0, err
		}

		staffType := model.StaffType(entry.StaffType)
		staff := &model.Staff{
			ClinicID:           entry.MainClinicID,
			Name:               strings.TrimSpace(entry.Name),
			OccupationID:       entry.OccupationID,
			IsActive:           entry.IsActive,
			AccountID:          &account.ID,
			StaffType:          staffType,
			ReservationVisible: entry.ReservationVisible,
		}
		if err := p.repo.CreateStaff(ctx, staff); err != nil {
			return 0, err
		}

		// Main clinic first, then remaining assignments sorted.
		clinicIDs := append([]uint64(nil), entry.ClinicIDs...)
		slices.Sort(clinicIDs)
		// Ensure main is written with IsMain=true regardless of sort position.
		if err := p.repo.CreateAssignment(ctx, &model.StaffClinicAssignment{
			StaffID:  staff.ID,
			ClinicID: entry.MainClinicID,
			IsMain:   true,
		}); err != nil {
			return 0, err
		}
		for _, clinicID := range clinicIDs {
			if clinicID == entry.MainClinicID {
				continue
			}
			if err := p.repo.CreateAssignment(ctx, &model.StaffClinicAssignment{
				StaffID:  staff.ID,
				ClinicID: clinicID,
				IsMain:   false,
			}); err != nil {
				return 0, err
			}
		}

		if len(entry.PermissionGroupIDs) > 0 {
			if err := p.repo.AssignPermissionGroups(
				ctx,
				entry.MainClinicID,
				staff.ID,
				entry.PermissionGroupIDs,
			); err != nil {
				return 0, err
			}
		}

		if err := p.writeStaffCreateAudit(ctx, actor, staff, manifest.BatchID, entry.ExternalStaffID, digest); err != nil {
			return 0, err
		}
		created++
	}
	return created, nil
}

func (p *StaffProvisioner) writeStaffCreateAudit(
	ctx context.Context,
	actor *StaffProvisionActor,
	staff *model.Staff,
	batchID, externalStaffID, digest string,
) error {
	clinicID := staff.ClinicID
	resourceID := staff.ID
	entry := &model.AuditLog{
		ClinicID:   &clinicID,
		Action:     model.AuditActionStaffProvisionCreate,
		Resource:   model.AuditResourceStaff,
		ResourceID: &resourceID,
		// PII-free: no name/email/password.
		NewValue: mustJSON(map[string]any{
			"batch_id":          batchID,
			"digest":            digest,
			"external_staff_id": externalStaffID,
			"staff_id":          staff.ID,
		}),
		UserAgent: staffProvisionUserAgent,
	}
	applyActor(entry, actor)
	return p.repo.WriteAudit(ctx, entry)
}

func (p *StaffProvisioner) writeReceipts(
	ctx context.Context,
	manifest *StaffProvisionManifest,
	digest string,
	count int,
	actor *StaffProvisionActor,
) error {
	for _, clinicID := range manifest.ClinicScope {
		clinicID := clinicID
		entry := &model.AuditLog{
			ClinicID: &clinicID,
			Action:   model.AuditActionStaffProvisionReceipt,
			Resource: model.AuditResourceStaffProvisionBatch,
			// PII-free receipt payload only.
			NewValue: mustJSON(map[string]any{
				"batch_id": manifest.BatchID,
				"digest":   digest,
				"count":    count,
			}),
			UserAgent: staffProvisionUserAgent,
		}
		applyActor(entry, actor)
		if err := p.repo.WriteAudit(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

func applyActor(entry *model.AuditLog, actor *StaffProvisionActor) {
	if actor == nil {
		entry.ActorType = model.AuditActorTypeSystem
		entry.ActorID = nil
		return
	}
	if actor.StaffID != nil {
		entry.ActorType = model.AuditActorTypeStaff
		entry.ActorID = actor.StaffID
		return
	}
	// System-admin CLI actor without a staff row uses system actor type.
	entry.ActorType = model.AuditActorTypeSystem
	entry.ActorID = nil
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}
