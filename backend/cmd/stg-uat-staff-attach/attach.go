// Command stg-uat-staff-attach links synthetic UAT accounts onto existing staffs
// rows without inserting staffs. Operator output is digest/count/ids only.
package main

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
)

func newStaffAttacher(repo attachRepository, repoRoots []string) *attacher {
	roots := make([]string, 0, len(repoRoots))
	for _, root := range repoRoots {
		if root == "" {
			continue
		}
		roots = append(roots, root)
	}
	return &attacher{repo: repo, repoRoots: roots}
}

func (a *attacher) Preflight(ctx context.Context, rosterPath, secretsPath string) (*attachResult, error) {
	if a == nil || a.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff attacher is not configured")
	}
	inputs, err := a.loadInputs(rosterPath, secretsPath)
	if err != nil {
		return nil, err
	}
	if err := a.validateRefs(ctx, inputs.roster); err != nil {
		return nil, err
	}
	return newAttachResult("preflight", inputs), nil
}

func (a *attacher) Apply(ctx context.Context, rosterPath, secretsPath string) (*attachResult, error) {
	if a == nil || a.repo == nil {
		return nil, apperrors.WrapInternalServerError("staff attacher is not configured")
	}
	inputs, err := a.loadInputs(rosterPath, secretsPath)
	if err != nil {
		return nil, err
	}

	applied := 0
	if err := a.repo.WithTx(ctx, func(txCtx context.Context) error {
		applied = 0
		for _, entry := range inputs.roster.Staff {
			staffRow, findErr := a.repo.FindStaffByID(txCtx, entry.StaffID)
			if findErr != nil {
				return findErr
			}
			clinicID, clinicErr := resolveClinic(entry, staffRow)
			if clinicErr != nil {
				return clinicErr
			}
			if emailErr := validateStaffEmail(entry.StaffID, entry.Email); emailErr != nil {
				return emailErr
			}
			if groupErr := validatePermissionGroupIDs(entry.PermissionGroupIDs); groupErr != nil {
				return groupErr
			}
			if belongErr := a.repo.PermissionGroupsBelongToClinic(txCtx, clinicID, entry.PermissionGroupIDs); belongErr != nil {
				return belongErr
			}

			staffDigest := inputs.staffDigests[entry.StaffID]
			if staffRow.AccountID != nil {
				linked, linkErr := a.linkedAccountDecision(txCtx, staffRow, entry.StaffID, staffDigest)
				if linkErr != nil {
					return linkErr
				}
				if linked {
					continue
				}
			}

			password := inputs.secrets[entry.SecretRef]
			hashed, hashErr := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
			password = ""
			if hashErr != nil {
				return apperrors.Wrap(hashErr, "failed to hash credential")
			}

			account := &model.Account{
				Email:         expectedStaffEmail(entry.StaffID),
				PasswordHash:  string(hashed),
				IsActive:      true,
				IsSystemAdmin: false,
			}
			if createErr := a.repo.CreateAccount(txCtx, account); createErr != nil {
				return createErr
			}
			if updateErr := a.repo.UpdateStaffAccount(txCtx, staffRow.ID, clinicID, account.ID, entry.SetActive); updateErr != nil {
				return updateErr
			}
			if assignErr := a.repo.EnsureClinicAssignment(txCtx, staffRow.ID, clinicID); assignErr != nil {
				return assignErr
			}
			if groupErr := a.repo.AssignPermissionGroups(txCtx, clinicID, staffRow.ID, entry.PermissionGroupIDs); groupErr != nil {
				return groupErr
			}
			if saveErr := a.repo.SaveAttachDigest(txCtx, staffRow.ID, staffDigest); saveErr != nil {
				return saveErr
			}
			applied++
		}
		return nil
	}); err != nil {
		return nil, err
	}

	status := "applied"
	if applied == 0 {
		status = "noop"
	}
	return newAttachResult(status, inputs), nil
}

func (a *attacher) linkedAccountDecision(
	ctx context.Context,
	staffRow *model.Staff,
	staffID uint64,
	staffDigest string,
) (bool, error) {
	last, err := a.repo.LastAttachDigest(ctx, staffID)
	if err != nil {
		return false, err
	}
	if last == staffDigest {
		return true, nil
	}
	if last != "" {
		return false, apperrors.WrapConflict("staff is already linked with a different attach digest")
	}
	account, err := a.repo.FindAccountByID(ctx, *staffRow.AccountID)
	if err != nil {
		return false, err
	}
	if normalizeEmail(account.Email) == expectedStaffEmail(staffID) {
		if saveErr := a.repo.SaveAttachDigest(ctx, staffID, staffDigest); saveErr != nil {
			return false, saveErr
		}
		return true, nil
	}
	return false, apperrors.WrapConflict("staff is already linked to a different account")
}

func (a *attacher) validateRefs(ctx context.Context, roster *rosterFile) error {
	for _, entry := range roster.Staff {
		staffRow, err := a.repo.FindStaffByID(ctx, entry.StaffID)
		if err != nil {
			return err
		}
		clinicID, err := resolveClinic(entry, staffRow)
		if err != nil {
			return err
		}
		if err := validateStaffEmail(entry.StaffID, entry.Email); err != nil {
			return err
		}
		if err := validatePermissionGroupIDs(entry.PermissionGroupIDs); err != nil {
			return err
		}
		if err := a.repo.PermissionGroupsBelongToClinic(ctx, clinicID, entry.PermissionGroupIDs); err != nil {
			return err
		}
	}
	return nil
}

func (a *attacher) loadInputs(rosterPath, secretsPath string) (*loadedInputs, error) {
	rosterHandle, _, err := staff.OpenSecureInputFile(rosterPath, a.repoRoots)
	if err != nil {
		return nil, apperrors.Wrap(err, "roster input rejected")
	}
	defer func() { _ = rosterHandle.Close() }()

	secretsHandle, _, err := staff.OpenSecureInputFile(secretsPath, a.repoRoots)
	if err != nil {
		return nil, apperrors.Wrap(err, "secrets input rejected")
	}
	defer func() { _ = secretsHandle.Close() }()

	var roster rosterFile
	if err := decodeStrictJSON(rosterHandle, &roster, "roster"); err != nil {
		return nil, err
	}
	var secrets secretsFile
	if err := decodeStrictJSON(secretsHandle, &secrets, "secrets"); err != nil {
		return nil, err
	}
	if err := validateRoster(&roster); err != nil {
		return nil, err
	}
	secretMap, err := validateSecrets(&roster, &secrets)
	if err != nil {
		return nil, err
	}
	digest, staffDigests, err := computeAttachDigests(&roster)
	if err != nil {
		return nil, err
	}
	return &loadedInputs{
		roster:       &roster,
		secrets:      secretMap,
		digest:       digest,
		staffDigests: staffDigests,
	}, nil
}
