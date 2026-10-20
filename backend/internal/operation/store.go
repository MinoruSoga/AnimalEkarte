package operation

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// gormStore is the operation_bindings / operation_results /
// operation_attempt_outcomes adapter. Every query carries clinic_id.
type gormStore struct {
	db *gorm.DB
}

// applySavepointName is the savepoint created before Participant.Apply so the
// owner writes can be undone when Apply reports an indeterminate effect.
const applySavepointName = "operation_apply_savepoint"

// newGormStore returns the storage used by NewRunner.
func newGormStore(db *gorm.DB) store {
	return &gormStore{db: db}
}

// ClaimBinding inserts the binding when this exact scope and key has none, then
// reads back the stored row. A concurrent claim of the same key yields the
// winner's fingerprint, which the runner compares with the requested content.
func (s *gormStore) ClaimBinding(
	ctx context.Context,
	scope Scope,
	key uuid.UUID,
	sum []byte,
	keyVersion string,
) (storedBinding, error) {
	if persistence.TxFromContext(ctx) != nil {
		return storedBinding{}, apperrors.WrapInternalServerError("operation claim must not run inside an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, s.db)

	candidate := model.OperationBinding{
		ClinicID:              scope.ClinicID,
		OwnerKind:             scope.OwnerKind,
		ActorKind:             string(scope.ActorKind),
		ActorRef:              scope.ActorRef,
		Intent:                scope.Intent,
		OperationKey:          key,
		Fingerprint:           sum,
		FingerprintKeyVersion: keyVersion,
	}
	// UNIQUE (clinic_id,owner_kind,actor_kind,actor_ref,intent,operation_key)
	// is the only conflict target; DO NOTHING keeps the first content binding.
	if err := db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&candidate).Error; err != nil {
		return storedBinding{}, apperrors.FromGORM(err, "operation_binding", "")
	}

	var row model.OperationBinding
	if err := db.
		Where("clinic_id = ? AND owner_kind = ? AND actor_kind = ? AND actor_ref = ? AND intent = ? AND operation_key = ?",
			scope.ClinicID, scope.OwnerKind, string(scope.ActorKind), scope.ActorRef, scope.Intent, key).
		Take(&row).Error; err != nil {
		return storedBinding{}, apperrors.FromGORM(err, "operation_binding", "")
	}
	return storedBinding{
		Binding: Binding{
			ID:       row.ID,
			ClinicID: row.ClinicID,
			Key:      row.OperationKey,
		},
		Fingerprint: row.Fingerprint,
		KeyVersion:  row.FingerprintKeyVersion,
	}, nil
}

// FindBinding returns the binding of this exact scope and key. A binding of
// another clinic, actor or intent is not visible: scope-not-found is answered
// without disclosing that a key exists elsewhere.
func (s *gormStore) FindBinding(
	ctx context.Context,
	scope Scope,
	key uuid.UUID,
) (storedBinding, bool, error) {
	var row model.OperationBinding
	err := persistence.DBOrTx(ctx, s.db).
		Where("clinic_id = ? AND owner_kind = ? AND actor_kind = ? AND actor_ref = ? AND intent = ? AND operation_key = ?",
			scope.ClinicID, scope.OwnerKind, string(scope.ActorKind), scope.ActorRef, scope.Intent, key).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return storedBinding{}, false, nil
	}
	if err != nil {
		return storedBinding{}, false, apperrors.FromGORM(err, "operation_binding", "")
	}
	return storedBinding{
		Binding: Binding{
			ID:       row.ID,
			ClinicID: row.ClinicID,
			Key:      row.OperationKey,
		},
		Fingerprint: row.Fingerprint,
		KeyVersion:  row.FingerprintKeyVersion,
	}, true, nil
}

// LockBindingNoWait locks the binding row for the owner transaction without
// waiting. A row held by another attempt maps to errBindingBusy.
func (s *gormStore) LockBindingNoWait(ctx context.Context, b Binding) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("operation binding lock requires an ambient transaction")
	}
	var row model.OperationBinding
	err := persistence.DBOrTx(ctx, s.db).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
		Select("id", "clinic_id").
		Where("clinic_id = ? AND id = ?", b.ClinicID, b.ID).
		Take(&row).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// 55P03 lock_not_available: the other attempt is inside the business
		// transaction, so the caller is told to query again, not to retry blind.
		if errors.As(err, &pgErr) && pgErr.Code == "55P03" {
			return errBindingBusy
		}
		return apperrors.FromGORM(err, "operation_binding", "")
	}
	return nil
}

// LoadResult returns the success receipt of the binding, or (nil, nil) when no
// attempt has succeeded.
func (s *gormStore) LoadResult(ctx context.Context, b Binding) (*Result, error) {
	var row model.OperationResult
	err := persistence.DBOrTx(ctx, s.db).
		Select("clinic_id", "binding_id", "result_refs", "committed_versions", "effect_summary", "recorded_at").
		Where("clinic_id = ? AND binding_id = ?", b.ClinicID, b.ID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.FromGORM(err, "operation_result", "")
	}
	result, err := resultFromStorage(row)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// InsertResult writes the success receipt in the transaction that performed the
// owner write; the deferred success trigger of the owner tables depends on it.
func (s *gormStore) InsertResult(ctx context.Context, b Binding, res Result) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("operation result insert requires an ambient transaction")
	}
	resultRefs, err := encodeStringMap(res.ResultRefs)
	if err != nil {
		return err
	}
	committedVersions, err := encodeStringMap(res.CommittedVersions)
	if err != nil {
		return err
	}
	effectSummary, err := encodeStringMap(res.EffectSummary)
	if err != nil {
		return err
	}
	row := model.OperationResult{
		ClinicID:          b.ClinicID,
		BindingID:         b.ID,
		ResultRefs:        resultRefs,
		CommittedVersions: committedVersions,
		EffectSummary:     effectSummary,
	}
	if err := persistence.DBOrTx(ctx, s.db).Create(&row).Error; err != nil {
		return apperrors.FromGORM(err, "operation_result", "")
	}
	return nil
}

// HasUnknownOutcome reports whether any unknown outcome was recorded for the
// binding. Unknown is sticky: a later failure row never releases it.
func (s *gormStore) HasUnknownOutcome(ctx context.Context, b Binding) (bool, error) {
	var exists bool
	err := persistence.DBOrTx(ctx, s.db).
		Raw(`SELECT EXISTS (SELECT 1 FROM operation_attempt_outcomes WHERE clinic_id = ? AND binding_id = ? AND outcome = ?)`,
			b.ClinicID, b.ID, string(OutcomeUnknown)).
		Scan(&exists).Error
	if err != nil {
		return false, apperrors.FromGORM(err, "operation_attempt_outcome", "")
	}
	return exists, nil
}

// RecordOutcome appends one attempt outcome. details carries only the status
// code: no content, fingerprint, PII or free text is stored.
func (s *gormStore) RecordOutcome(
	ctx context.Context,
	b Binding,
	attemptKey uuid.UUID,
	outcome AttemptOutcome,
	status Status,
) error {
	if persistence.TxFromContext(ctx) != nil {
		return apperrors.WrapInternalServerError("operation outcome must not run inside an ambient transaction")
	}
	details, err := encodeStringMap(map[string]string{"status": string(status)})
	if err != nil {
		return err
	}
	row := model.OperationAttemptOutcome{
		ClinicID:   b.ClinicID,
		BindingID:  b.ID,
		AttemptKey: attemptKey,
		Outcome:    string(outcome),
		Details:    details,
	}
	if err := persistence.DBOrTx(ctx, s.db).Create(&row).Error; err != nil {
		return apperrors.FromGORM(err, "operation_attempt_outcome", "")
	}
	return nil
}

// MarkApplySavepoint creates the savepoint that RecordUnknownUnderLock rolls
// back to. The runner calls it immediately before Participant.Apply: a
// savepoint created after the owner writes would undo nothing, so the writes
// could commit together with an unknown outcome.
func (s *gormStore) MarkApplySavepoint(ctx context.Context) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("operation apply savepoint requires an ambient transaction")
	}
	if err := persistence.DBOrTx(ctx, s.db).SavePoint(applySavepointName).Error; err != nil {
		return apperrors.FromGORM(err, "operation_attempt_outcome", "")
	}
	return nil
}

// RecordUnknownUnderLock records the unknown attempt outcome inside the
// caller's business transaction: it rolls back to the savepoint that
// MarkApplySavepoint established before the owner writes — undoing those
// writes — and inserts the unknown row. The caller's commit then makes the
// stop state durable while the binding row lock is still held, so no
// concurrent execution can run Apply in the window between the rollback and
// the unknown record. details carries only the status code: no content,
// fingerprint, PII or free text is stored.
func (s *gormStore) RecordUnknownUnderLock(
	ctx context.Context,
	b Binding,
	attemptKey uuid.UUID,
	status Status,
) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("operation unknown record requires an ambient transaction")
	}
	tx := persistence.DBOrTx(ctx, s.db)
	if err := tx.RollbackTo(applySavepointName).Error; err != nil {
		return apperrors.FromGORM(err, "operation_attempt_outcome", "")
	}
	details, err := encodeStringMap(map[string]string{"status": string(status)})
	if err != nil {
		return err
	}
	row := model.OperationAttemptOutcome{
		ClinicID:   b.ClinicID,
		BindingID:  b.ID,
		AttemptKey: attemptKey,
		Outcome:    string(OutcomeUnknown),
		Details:    details,
	}
	if err := tx.Create(&row).Error; err != nil {
		return apperrors.FromGORM(err, "operation_attempt_outcome", "")
	}
	return nil
}

// encodeStringMap stores a JSON object. A nil or empty map is stored as {},
// because the columns reject a JSON null.
func encodeStringMap(values map[string]string) (json.RawMessage, error) {
	if len(values) == 0 {
		return json.RawMessage("{}"), nil
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return nil, apperrors.WrapInternalServerError("operation receipt could not be encoded")
	}
	return encoded, nil
}

// decodeStringMap reads a JSON object of string values.
func decodeStringMap(raw json.RawMessage) (map[string]string, error) {
	if len(raw) == 0 {
		return map[string]string{}, nil
	}
	values := make(map[string]string)
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, apperrors.WrapInternalServerError("operation receipt could not be decoded")
	}
	return values, nil
}

func resultFromStorage(row model.OperationResult) (Result, error) {
	resultRefs, err := decodeStringMap(row.ResultRefs)
	if err != nil {
		return Result{}, err
	}
	committedVersions, err := decodeStringMap(row.CommittedVersions)
	if err != nil {
		return Result{}, err
	}
	effectSummary, err := decodeStringMap(row.EffectSummary)
	if err != nil {
		return Result{}, err
	}
	return Result{
		ResultRefs:        resultRefs,
		CommittedVersions: committedVersions,
		EffectSummary:     effectSummary,
		RecordedAt:        row.RecordedAt,
	}, nil
}
