package medicalrecord

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

// fakeAuditServiceForLab is a minimal double for the consumer-side AuditLogger view
// (non-tx LogEntry over *AuditEntry) that labAuditLogger writes through after the BE9-2D move.
// Before the move it implemented the full internal/service.AuditService; only LogEntry was ever
// exercised, so the migrated copy narrows to exactly that method.
type fakeAuditServiceForLab struct {
	entries []*AuditEntry
	err     error
}

func (f *fakeAuditServiceForLab) LogEntry(_ context.Context, input *AuditEntry) error {
	if f.err != nil {
		return f.err
	}
	f.entries = append(f.entries, input)
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

	logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategoryInvalidInput)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, model.AuditActionLabImportCommitFailed, e.Action)
	meta, ok := e.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_input", meta["error_category"])
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
		logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategoryInternal)
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

// ------------------------------------
// Tests: Phase 4A.3 — LabAuditErrorCategory runtime validation (fail-closed)
// ------------------------------------

func TestLabAuditLogger_LogCommitFailed_InvalidCategory_NotEmitted(t *testing.T) {
	// model.LabAuditErrorCategory("arbitrary text") must not reach the audit payload.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategory("arbitrary text"))

	assert.Empty(t, fake.entries, "invalid category must not produce an audit entry")
}

func TestLabAuditLogger_LogCommitFailed_EmptyCategory_NotEmitted(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategory(""))

	assert.Empty(t, fake.entries, "empty category must not produce an audit entry")
}

func TestLabAuditLogger_LogCommitFailed_InvalidCategoryDoesNotBreakAPI(t *testing.T) {
	// Fail-closed must remain best-effort: no panic, no error returned to caller.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	assert.NotPanics(t, func() {
		logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategory("patient name / arbitrary"))
	})
	assert.Empty(t, fake.entries)
}

func TestLabAuditLogger_LogCommitFailed_ValidCategoryStillEmitted(t *testing.T) {
	// Regression guard: valid categories must still reach the audit payload unchanged.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategoryInvalidInput)

	require.Len(t, fake.entries, 1)
	meta, ok := fake.entries[0].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "invalid_input", meta["error_category"])
}

func TestLabAuditLogger_LogCommitFailed_AllValidCategories_Emitted(t *testing.T) {
	// Each declared constant must produce exactly one audit entry with the correct string value.
	cases := []struct {
		category model.LabAuditErrorCategory
		want     string
	}{
		{model.LabAuditErrorCategoryInvalidInput, "invalid_input"},
		{model.LabAuditErrorCategoryNotFound, "not_found"},
		{model.LabAuditErrorCategoryForbidden, "forbidden"},
		{model.LabAuditErrorCategoryUnauthorized, "unauthorized"},
		{model.LabAuditErrorCategoryInternal, "internal"},
	}
	for _, tc := range cases {
		fake := &fakeAuditServiceForLab{}
		logger := NewLabAuditLogger(fake)

		logger.LogCommitFailed(context.Background(), 1, nil, tc.category)

		require.Len(t, fake.entries, 1, "category %q must produce one audit entry", tc.category)
		meta, ok := fake.entries[0].Metadata.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, tc.want, meta["error_category"], "category %q must serialize to %q", tc.category, tc.want)
	}
}

func TestLabAuditLogger_LogCommitFailed_ClinicAndActorPreserved(t *testing.T) {
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)
	actorID := uint64(77)
	clinicID := uint64(5)

	logger.LogCommitFailed(context.Background(), clinicID, &actorID, model.LabAuditErrorCategoryInternal)

	require.Len(t, fake.entries, 1)
	e := fake.entries[0]
	assert.Equal(t, clinicID, *e.ClinicID)
	assert.Equal(t, &actorID, e.ActorID)
	assert.Equal(t, model.AuditActionLabImportCommitFailed, e.Action)
}

func TestLabAuditLogger_LogCommitFailed_NoPIIInPayload(t *testing.T) {
	// The audit payload must not contain raw lab values, patient data, or credential-like fields.
	fake := &fakeAuditServiceForLab{}
	logger := NewLabAuditLogger(fake)

	logger.LogCommitFailed(context.Background(), 1, nil, model.LabAuditErrorCategoryInternal)

	require.Len(t, fake.entries, 1)
	meta, ok := fake.entries[0].Metadata.(map[string]any)
	require.True(t, ok)
	for _, forbidden := range []string{
		"inspection_value", "display_value", "reference_value",
		"old_pet_key", "old_chart_key", "old_row_key",
		"source_fingerprint", "pet_name", "owner_name",
		"email", "phone", "password", "token",
	} {
		_, present := meta[forbidden]
		assert.False(t, present, "audit metadata must not contain %q", forbidden)
	}
}

// errAuditFail is a sentinel error for best-effort tests.
var errAuditFail = &auditTestError{"audit write failed"}

type auditTestError struct{ msg string }

func (e *auditTestError) Error() string { return e.msg }
