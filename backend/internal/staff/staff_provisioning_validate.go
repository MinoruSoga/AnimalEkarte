package staff

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/unix"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

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
	return os.NewFile(uintptr(fd), realPath), realPath, nil //nolint:gosec // G115: fd comes from unix.Open
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
	seenClinic, err := validateStaffProvisionClinicScope(manifest)
	if err != nil {
		return err
	}
	if len(manifest.Staff) == 0 {
		return apperrors.WrapInvalidInput("staff must not be empty")
	}

	union := make(map[uint64]struct{})
	seenExternal := make(map[string]struct{}, len(manifest.Staff))
	seenEmail := make(map[string]struct{}, len(manifest.Staff))
	seenSecretRef := make(map[string]struct{}, len(manifest.Staff))

	for i, entry := range manifest.Staff {
		if err := validateStaffProvisionEntry(
			fmt.Sprintf("staff[%d]", i),
			entry,
			seenClinic,
			seenExternal,
			seenEmail,
			seenSecretRef,
			union,
		); err != nil {
			return err
		}
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

func validateStaffProvisionClinicScope(manifest *StaffProvisionManifest) (map[uint64]struct{}, error) {
	if len(manifest.ClinicScope) == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_scope must not be empty")
	}
	if !slices.IsSorted(manifest.ClinicScope) {
		return nil, apperrors.WrapInvalidInput("clinic_scope must be sorted ascending")
	}
	seenClinic := make(map[uint64]struct{}, len(manifest.ClinicScope))
	for _, id := range manifest.ClinicScope {
		if id == 0 {
			return nil, apperrors.WrapInvalidInput("clinic_scope must not contain zero")
		}
		if _, dup := seenClinic[id]; dup {
			return nil, apperrors.WrapInvalidInput("clinic_scope must not contain duplicates")
		}
		seenClinic[id] = struct{}{}
	}
	expectedBatchID := ClinicScopeBatchID(manifest.ClinicScope)
	if manifest.BatchID != expectedBatchID {
		return nil, apperrors.WrapInvalidInput("batch_id does not match clinic_scope digest namespace")
	}
	return seenClinic, nil
}

func validateStaffProvisionEntry(
	prefix string,
	entry StaffProvisionStaffEntry,
	seenClinic map[uint64]struct{},
	seenExternal map[string]struct{},
	seenEmail map[string]struct{},
	seenSecretRef map[string]struct{},
	union map[uint64]struct{},
) error {
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
