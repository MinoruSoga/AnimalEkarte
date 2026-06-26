package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ------------------------------------
// Fake AuditService for audit logger tests
// ------------------------------------

type fakeAuditServiceForLab struct {
	entries []*AuditLogInput
	err     error
}

func (f *fakeAuditServiceForLab) Log(_ context.Context, _ *model.AuditLog) error {
	return f.err
}
func (f *fakeAuditServiceForLab) LogEntry(_ context.Context, input *AuditLogInput) error {
	if f.err != nil {
		return f.err
	}
	f.entries = append(f.entries, input)
	return nil
}
func (f *fakeAuditServiceForLab) LogAuthLogin(_ context.Context, _, _ *uint64, _, _, _ string) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogLstepOperation(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogMedicalRecordChange(_ context.Context, _ uint64, _ *uint64, _ string, _ uint64, _, _ map[string]any) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogVitalChange(_ context.Context, _ uint64, _ *uint64, _ string, _, _ uint64, _, _ map[string]any) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogAddendumCreate(_ context.Context, _ uint64, _ *uint64, _, _ uint64, _ *model.MedicalRecordAddendum) error {
	return nil
}
func (f *fakeAuditServiceForLab) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}

// ------------------------------------
// Tests: LogPreviewRequested
// ------------------------------------

func TestLabAuditLogger_LogPreviewRequested_RecordsEvent(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(42)

	logger.LogPreviewRequested(context.Background(), 1, &actorID, "fixture", 5)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportPreviewRequested, e.Action)
	assert.Equal(t, model.AuditResourceLabImport, e.Resource)
	assert.Equal(t, uint64(1), *e.ClinicID)
	assert.Equal(t, &actorID, e.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, e.ActorType)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok, "metadata should be map[string]any")
	assert.Equal(t, "fixture", meta["source_type"])
	assert.Equal(t, 5, meta["row_count"])
}

func TestLabAuditLogger_LogPreviewRequested_SystemActorWhenNoActorID(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogPreviewRequested(context.Background(), 1, nil, "fixture", 3)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActorTypeSystem, e.ActorType)
	assert.Nil(t, e.ActorID)
}

// ------------------------------------
// Tests: LogCommitRequested
// ------------------------------------

func TestLabAuditLogger_LogCommitRequested_RecordsEvent(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(7)

	logger.LogCommitRequested(context.Background(), 2, &actorID, "fixture", 10)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportCommitRequested, e.Action)
	assert.Equal(t, uint64(2), *e.ClinicID)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "fixture", meta["source_type"])
	assert.Equal(t, 10, meta["row_count"])
}

// ------------------------------------
// Tests: LogCommitSucceeded
// ------------------------------------

func TestLabAuditLogger_LogCommitSucceeded_RecordsEvent(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(9)
	jobID := uuid.New()

	logger.LogCommitSucceeded(context.Background(), 1, &actorID, jobID, CommitAuditCounts{
		RowCount: 3, PersistedCount: 2, DuplicateCount: 1, FailedCount: 0,
	})

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportCommitSucceeded, e.Action)
	assert.Equal(t, model.AuditResourceLabImport, e.Resource)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, jobID.String(), meta["job_id"])
	assert.Equal(t, 3, meta["row_count"])
	assert.Equal(t, 2, meta["persisted_count"])
	assert.Equal(t, 1, meta["duplicate_count"])
	assert.Equal(t, 0, meta["failed_count"])
}

func TestLabAuditLogger_LogCommitSucceeded_NoRawLabValues(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(9)
	jobID := uuid.New()

	logger.LogCommitSucceeded(context.Background(), 1, &actorID, jobID, CommitAuditCounts{
		RowCount: 2, PersistedCount: 2,
	})

	require.Len(t, fake.entries, 1)
	meta, ok := fake.entries[0].Metadata.(map[string]any)
	require.True(t, ok)
	// PII-safety: raw lab values, pet names, fingerprints MUST NOT appear
	for _, forbidden := range []string{
		"inspection_value", "display_value", "reference_value",
		"old_pet_key", "old_chart_key", "old_row_key",
		"source_fingerprint",
		"pet_name", "owner_name", "email", "phone",
	} {
		_, present := meta[forbidden]
		assert.False(t, present, "audit metadata must not contain %q", forbidden)
	}
}

// ------------------------------------
// Tests: LogCommitFailed
// ------------------------------------

func TestLabAuditLogger_LogCommitFailed_RecordsEvent(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogCommitFailed(context.Background(), 1, nil, "invalid_source_type")

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportCommitFailed, e.Action)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_source_type", meta["error_category"])
}

// ------------------------------------
// Tests: LogSourceBlocked
// ------------------------------------

func TestLabAuditLogger_LogSourceBlocked_RecordsEvent(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(5)

	logger.LogSourceBlocked(context.Background(), 1, &actorID, "drwan", "commit", model.LabBlockedReasonMDBSchemaUnconfirmed)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportSourceBlocked, e.Action)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "drwan", meta["source_type"])
	assert.Equal(t, "commit", meta["operation"])
	assert.Equal(t, string(model.LabBlockedReasonMDBSchemaUnconfirmed), meta["reason"])
}

func TestLabAuditLogger_LogSourceBlocked_ReasonIsTypedConstant(t *testing.T) {
	// Compile-time guarantee: only model.LabBlockedReason values accepted — not arbitrary strings.
	// This test documents the taxonomy contract; if the signature reverts to plain string,
	// the typed-reason call sites below will fail to compile.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	for _, reason := range []model.LabBlockedReason{
		model.LabBlockedReasonMDBSchemaUnconfirmed,
		model.LabBlockedReasonSourceNotImplemented,
		model.LabBlockedReasonSourceTypeBlocked,
	} {
		fake.entries = nil
		logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "preview", reason)
		require.Len(t, fake.entries, 1)
		meta, ok := fake.entries[0].Metadata.(map[string]any)
		require.True(t, ok)
		// reason in payload must be the string form of the typed constant
		assert.Equal(t, string(reason), meta["reason"])
	}
}

func TestLabAuditLogger_LogSourceBlocked_ReasonStoredAsString(t *testing.T) {
	// Verify the typed constant is faithfully stored in the payload as its string value.
	// This pins the serialization contract: reason key holds the constant's string form.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "preview", model.LabBlockedReasonMDBSchemaUnconfirmed)

	require.Len(t, fake.entries, 1)
	meta, ok := fake.entries[0].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "mdb_schema_not_confirmed", meta["reason"], "reason must be the constant's string value")
}

// ------------------------------------
// Tests: failure policy (best-effort)
// ------------------------------------

func TestLabAuditLogger_AuditFailure_DoesNotPanic(t *testing.T) {
	// AuditService returns an error: lab import should NOT fail.
	// LabAuditLogger absorbs the error and logs a warning.
	fake := &fakeAuditServiceForLab{err: errAuditFail}
	logger := NewLabAuditLogger(fake)

	// None of these should panic or return an error to the caller.
	assert.NotPanics(t, func() {
		logger.LogPreviewRequested(context.Background(), 1, nil, "fixture", 1)
	})
	assert.NotPanics(t, func() {
		logger.LogCommitRequested(context.Background(), 1, nil, "fixture", 1)
	})
	assert.NotPanics(t, func() {
		logger.LogCommitSucceeded(context.Background(), 1, nil, uuid.New(), CommitAuditCounts{})
	})
	assert.NotPanics(t, func() {
		logger.LogCommitFailed(context.Background(), 1, nil, "context_cancelled")
	})
	assert.NotPanics(t, func() {
		logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "commit", model.LabBlockedReasonSourceTypeBlocked)
	})
}

// ------------------------------------
// Tests: clinic scope preserved
// ------------------------------------

func TestLabAuditLogger_ClinicAndActorPreserved(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(99)
	clinicID := uint64(3)

	logger.LogPreviewRequested(context.Background(), clinicID, &actorID, "fixture", 0)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, clinicID, *e.ClinicID, "clinic_id must be preserved in audit entry")
	assert.Equal(t, &actorID, e.ActorID, "actor_id must be preserved in audit entry")
}

// ------------------------------------
// Tests: Phase 4A.2 — runtime reason validation (fail-closed)
// ------------------------------------

func TestLabAuditLogger_LogSourceBlocked_InvalidReason_NotEmitted(t *testing.T) {
	// model.LabBlockedReason("arbitrary") must not reach the audit payload.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "commit", model.LabBlockedReason("arbitrary text"))

	assert.Empty(t, fake.entries, "invalid reason must not produce an audit entry")
}

func TestLabAuditLogger_LogSourceBlocked_EmptyReason_NotEmitted(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "commit", model.LabBlockedReason(""))

	assert.Empty(t, fake.entries, "empty reason must not produce an audit entry")
}

func TestLabAuditLogger_LogSourceBlocked_InvalidReasonDoesNotBreakAPI(t *testing.T) {
	// Fail-closed must remain best-effort: no panic, no error returned to caller.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	assert.NotPanics(t, func() {
		logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "preview", model.LabBlockedReason("arbitrary"))
	})
	// No audit entry produced.
	assert.Empty(t, fake.entries)
}

func TestLabAuditLogger_LogSourceBlocked_ValidReasonStillEmitted(t *testing.T) {
	// Regression guard: valid reasons must still reach the audit payload unchanged.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogSourceBlocked(context.Background(), 1, nil, "drwan", "commit", model.LabBlockedReasonSourceTypeBlocked)

	require.Len(t, fake.entries, 1)
	meta, ok := fake.entries[0].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "source_type_blocked", meta["reason"])
}

// errAuditFail is a sentinel error for best-effort tests.
var errAuditFail = &auditTestError{"audit write failed"}

type auditTestError struct{ msg string }

func (e *auditTestError) Error() string { return e.msg }
