// Command stg-uat-staff-attach links synthetic UAT accounts onto existing staffs
// rows without inserting staffs. Operator output is digest/count/ids only.
package main

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func decodeStrictJSON(r io.Reader, dest any, label string) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s decode failed: %v", label, err))
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%s must contain exactly one JSON value", label))
	}
	return nil
}

func validateRoster(roster *rosterFile) error {
	if roster == nil {
		return apperrors.WrapInvalidInput("roster is required")
	}
	if roster.SchemaVersion != rosterSchemaVersion {
		return apperrors.WrapInvalidInput("unsupported schema_version")
	}
	if len(roster.Staff) == 0 {
		return apperrors.WrapInvalidInput("staff must not be empty")
	}

	seenID := make(map[uint64]struct{}, len(roster.Staff))
	for i := range roster.Staff {
		entry := &roster.Staff[i]
		prefix := fmt.Sprintf("staff[%d]", i)
		if entry.StaffID == 0 {
			return apperrors.WrapInvalidInput(prefix + ": staff_id is required")
		}
		if _, dup := seenID[entry.StaffID]; dup {
			return apperrors.WrapInvalidInput("duplicate staff_id in roster")
		}
		seenID[entry.StaffID] = struct{}{}

		entry.Email = normalizeEmail(entry.Email)
		if err := validateStaffEmail(entry.StaffID, entry.Email); err != nil {
			return apperrors.Wrap(err, prefix)
		}
		entry.SecretRef = strings.TrimSpace(entry.SecretRef)
		if entry.SecretRef == "" {
			return apperrors.WrapInvalidInput(prefix + ": secret_ref is required")
		}
		if err := validatePermissionGroupIDs(entry.PermissionGroupIDs); err != nil {
			return apperrors.Wrap(err, prefix)
		}
		if entry.ClinicID == 0 && len(entry.ClinicIDs) == 0 {
			return apperrors.WrapInvalidInput(prefix + ": clinic_id or clinic_ids is required")
		}
		for _, clinicID := range entry.ClinicIDs {
			if clinicID == 0 {
				return apperrors.WrapInvalidInput(prefix + ": clinic_ids must not contain zero")
			}
		}
	}

	slices.SortFunc(roster.Staff, func(a, b rosterStaff) int {
		return cmp.Compare(a.StaffID, b.StaffID)
	})
	return nil
}

func validateSecrets(roster *rosterFile, secrets *secretsFile) (map[string]string, error) {
	if secrets == nil {
		return nil, apperrors.WrapInvalidInput("secrets are required")
	}
	if len(secrets.Secrets) == 0 {
		return nil, apperrors.WrapInvalidInput("secrets must not be empty")
	}
	required := make(map[string]struct{}, len(roster.Staff))
	for _, entry := range roster.Staff {
		required[entry.SecretRef] = struct{}{}
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
		if strings.TrimSpace(secret.Password) == "" {
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("secrets[%d]: password is required", i))
		}
		out[ref] = secret.Password
	}
	if len(out) != len(required) {
		return nil, apperrors.WrapInvalidInput("secrets file must cover every staff secret_ref")
	}
	return out, nil
}

func validatePermissionGroupIDs(ids []uint64) error {
	if len(ids) == 0 {
		return apperrors.WrapInvalidInput("permission_group_ids must not be empty")
	}
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return apperrors.WrapInvalidInput("permission_group_ids must not contain zero")
		}
		if _, dup := seen[id]; dup {
			return apperrors.WrapInvalidInput("permission_group_ids must not contain duplicates")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func expectedStaffEmail(staffID uint64) string {
	return fmt.Sprintf("stg-staff-%d@example.test", staffID)
}

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func validateStaffEmail(staffID uint64, email string) error {
	if normalizeEmail(email) != expectedStaffEmail(staffID) {
		return apperrors.WrapInvalidInput("email must match stg-staff-{id}@example.test")
	}
	return nil
}

func resolveClinic(entry rosterStaff, staffRow *model.Staff) (uint64, error) {
	if staffRow == nil {
		return 0, apperrors.WrapInvalidInput("staff is required")
	}
	if entry.ClinicID != 0 {
		if staffRow.ClinicID != entry.ClinicID {
			return 0, apperrors.WrapInvalidInput("staff clinic does not match roster clinic")
		}
		return entry.ClinicID, nil
	}
	if len(entry.ClinicIDs) == 0 {
		return 0, apperrors.WrapInvalidInput("clinic_id or clinic_ids is required")
	}
	if !slices.Contains(entry.ClinicIDs, staffRow.ClinicID) {
		return 0, apperrors.WrapInvalidInput("staff clinic is not in roster clinic_ids")
	}
	return staffRow.ClinicID, nil
}

func computeAttachDigests(roster *rosterFile) (string, map[uint64]string, error) {
	entries := make([]digestStaff, 0, len(roster.Staff))
	staffDigests := make(map[uint64]string, len(roster.Staff))
	for _, entry := range roster.Staff {
		clinicIDs := append([]uint64(nil), entry.ClinicIDs...)
		slices.Sort(clinicIDs)
		groupIDs := append([]uint64(nil), entry.PermissionGroupIDs...)
		slices.Sort(groupIDs)
		identitySum := sha256.Sum256([]byte(entry.Email))
		item := digestStaff{
			StaffID:             entry.StaffID,
			ClinicID:            entry.ClinicID,
			ClinicIDs:           clinicIDs,
			IdentityFingerprint: hex.EncodeToString(identitySum[:]),
			PermissionGroupIDs:  groupIDs,
			SetActive:           entry.SetActive,
			SecretRef:           entry.SecretRef,
		}
		staffPayload, err := json.Marshal(item)
		if err != nil {
			return "", nil, apperrors.Wrap(err, "failed to marshal staff attach digest")
		}
		staffSum := sha256.Sum256(staffPayload)
		staffDigests[entry.StaffID] = hex.EncodeToString(staffSum[:])
		entries = append(entries, item)
	}
	payload, err := json.Marshal(digestRoot{
		SchemaVersion: roster.SchemaVersion,
		Staff:         entries,
	})
	if err != nil {
		return "", nil, apperrors.Wrap(err, "failed to marshal attach digest")
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), staffDigests, nil
}

func newAttachResult(status string, inputs *loadedInputs) *attachResult {
	ids := make([]uint64, 0, len(inputs.roster.Staff))
	for _, entry := range inputs.roster.Staff {
		ids = append(ids, entry.StaffID)
	}
	return &attachResult{
		Status:     status,
		Digest:     inputs.digest,
		StaffCount: len(ids),
		StaffIDs:   ids,
	}
}
