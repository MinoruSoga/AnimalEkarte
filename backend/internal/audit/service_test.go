package audit

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

type recordingRepository struct {
	createErr   error
	createTxErr error
	createCalls int
	txCalls     int
	lastLog     *model.AuditLog
}

func (r *recordingRepository) Create(_ context.Context, log *model.AuditLog) error {
	r.createCalls++
	r.lastLog = log
	return r.createErr
}

func (r *recordingRepository) CreateTx(_ context.Context, log *model.AuditLog) error {
	r.txCalls++
	r.lastLog = log
	return r.createTxErr
}

func TestServiceLogEntryBuildsCanonicalLog(t *testing.T) {
	repo := &recordingRepository{}
	svc := NewService(repo)
	clinicID := uint64(1)
	actorID := uint64(2)
	resourceID := uint64(3)

	err := svc.LogEntry(context.Background(), &Entry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     "update",
		Resource:   "permission_group",
		ResourceID: &resourceID,
		OldValue:   map[string]any{"name": "old"},
		NewValue:   map[string]any{"name": "new"},
		IPAddress:  "",
		UserAgent:  "audit-test",
	})

	require.NoError(t, err)
	require.Equal(t, 1, repo.createCalls)
	require.NotNil(t, repo.lastLog)
	assert.Nil(t, repo.lastLog.IPAddress, "empty IP addresses must persist as SQL NULL")
	assert.JSONEq(t, `{"name":"old"}`, string(repo.lastLog.OldValue))
	assert.JSONEq(t, `{"name":"new"}`, string(repo.lastLog.NewValue))
	assert.Equal(t, "audit-test", repo.lastLog.UserAgent)
}

func TestServiceLogEntryNormalizesAndBoundsUserAgent(t *testing.T) {
	repo := &recordingRepository{}
	svc := NewService(repo)
	clinicID := uint64(1)
	longUserAgent := strings.Repeat("a", maxAuditUserAgentBytes-1) + "界\xff"

	err := svc.LogEntry(context.Background(), &Entry{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    "test.user_agent",
		Resource:  "auth",
		UserAgent: longUserAgent,
	})

	require.NoError(t, err)
	require.NotNil(t, repo.lastLog)
	assert.LessOrEqual(t, len(repo.lastLog.UserAgent), maxAuditUserAgentBytes)
	assert.True(t, utf8.ValidString(repo.lastLog.UserAgent))
	assert.NotContains(t, repo.lastLog.UserAgent, "\xff")
}

func TestServiceLogEntryTxUsesTransactionRepositoryPath(t *testing.T) {
	repo := &recordingRepository{}
	svc := NewService(repo)
	clinicID := uint64(1)

	err := svc.LogEntryTx(context.Background(), &Entry{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    "replace",
		Resource:  "checkup_field_result",
	})

	require.NoError(t, err)
	assert.Zero(t, repo.createCalls)
	assert.Equal(t, 1, repo.txCalls)
}

func TestServiceMedicalRecordChangePreservesNilJSONColumns(t *testing.T) {
	repo := &recordingRepository{}
	svc := NewService(repo)

	err := svc.LogMedicalRecordChange(
		context.Background(),
		1,
		uint64Pointer(2),
		"create",
		3,
		nil,
		map[string]any{"status": "draft"},
	)

	require.NoError(t, err)
	require.NotNil(t, repo.lastLog)
	assert.Nil(t, repo.lastLog.OldValue)
	assert.JSONEq(t, `{"status":"draft"}`, string(repo.lastLog.NewValue))
}

func TestServiceRejectsInvalidEntriesBeforePersistence(t *testing.T) {
	tests := []struct {
		name  string
		entry *Entry
	}{
		{name: "nil entry", entry: nil},
		{
			name: "missing clinic",
			entry: &Entry{
				ActorType: model.AuditActorTypeSystem,
				Action:    "batch",
				Resource:  "clinic",
			},
		},
		{
			name: "staff without actor",
			entry: &Entry{
				ClinicID:  uint64Pointer(1),
				ActorType: model.AuditActorTypeStaff,
				Action:    "update",
				Resource:  "owner",
			},
		},
		{
			name: "system with actor",
			entry: &Entry{
				ClinicID:  uint64Pointer(1),
				ActorID:   uint64Pointer(2),
				ActorType: model.AuditActorTypeSystem,
				Action:    "batch",
				Resource:  "owner",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &recordingRepository{}

			err := NewService(repo).LogEntry(context.Background(), tt.entry)

			assert.Error(t, err)
			assert.Zero(t, repo.createCalls)
			assert.Zero(t, repo.txCalls)
		})
	}
}

func TestServiceRecordsWriteFailureWithoutLeakingAuditPayload(t *testing.T) {
	var output bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelError})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	repo := &recordingRepository{createErr: errors.New("database unavailable")}
	clinicID := uint64(7)
	err := NewService(repo).LogEntry(context.Background(), &Entry{
		ClinicID:  &clinicID,
		ActorType: model.AuditActorTypeSystem,
		Action:    "batch_dormant_detect",
		Resource:  "clinic",
		NewValue:  map[string]any{"private": "must-not-be-logged"},
	})

	require.Error(t, err)
	assert.Contains(t, output.String(), "audit_write_failed")
	assert.NotContains(t, output.String(), "must-not-be-logged")
}

func TestMarshalJSONIsBestEffort(t *testing.T) {
	assert.Nil(t, MarshalJSON(nil))
	assert.Nil(t, MarshalJSON(make(chan int)))
	assert.JSONEq(t, `{"key":"value"}`, string(MarshalJSON(map[string]string{"key": "value"})))
}

func uint64Pointer(value uint64) *uint64 {
	return &value
}
