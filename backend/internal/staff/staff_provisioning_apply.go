package staff

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

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
