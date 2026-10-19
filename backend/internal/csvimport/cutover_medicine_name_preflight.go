package csvimport

import (
	"context"
)

// cutoverMedicineNameBatchSize bounds the names per target lookup so one
// query stays small; no per-row database queries and no source values in
// errors.
const cutoverMedicineNameBatchSize = 256

// validateCutoverMedicinesTargetNames implements the pre-Begin leg of
// Codex 4266290426 for user decision A: before candidate22's atomic
// transaction, the bundle's medicines CSV is
// re-read (row count and SHA-256 bound to the exact opened bytes; header
// permutation only under the existing rehearsal contract drift) and every
// exact source name is checked against the target clinic's undeleted
// medicines. Soft-deleted rows and other clinics never collide; inactive
// rows do. Legacy21 never calls this helper. Read-only: no insert, update,
// or delete.
func validateCutoverMedicinesTargetNames(ctx context.Context, session cutoverSession, bundle CutoverBundle, seeds CutoverSeedIDs) error {
	return validateCutoverMasterTargetNames(ctx, session, bundle, seeds, "medicines",
		`SELECT EXISTS (SELECT 1 FROM public.medicines WHERE clinic_id = $1 AND deleted_at IS NULL AND name = ANY($2::text[]))`)
}
