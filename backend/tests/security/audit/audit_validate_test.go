// Package audit — EMR-332 C18 evidence tests
// (docs/work/emr-332-security/workflow-audit/c18-logging-audit.md).
//
// Scope: deterministic, single-threaded, in-memory verification of the audit
// package service layer. The persistence boundary (audit.Repository) is a
// counting fake — no audit row is ever written to a database. The one real
// repository call (CreateTx without an ambient transaction) returns before
// any gorm.DB interaction, proving the fail-closed ambient-tx contract.
//
// All payload values are synthetic markers (SECCHK20261009*) — no PII.
package audit

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

// countingRepo is the in-memory audit persistence boundary. A repository write
// here is the counted "audit side effect" — no row ever reaches a database.
type countingRepo struct {
	created   []*model.AuditLog
	createdTx []*model.AuditLog
	createErr error
	txErr     error
}

func (r *countingRepo) Create(_ context.Context, log *model.AuditLog) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, log)
	return nil
}

func (r *countingRepo) CreateTx(_ context.Context, log *model.AuditLog) error {
	if r.txErr != nil {
		return r.txErr
	}
	r.createdTx = append(r.createdTx, log)
	return nil
}

func u64(v uint64) *uint64 { return &v }

// ── C18: ValidateLog enforces the audit_logs DDL invariants ──────────────────

func TestC18_ValidateLogInvariantMatrix(t *testing.T) {
	cases := []struct {
		name    string
		log     *model.AuditLog
		wantErr bool
	}{
		{"nil log", nil, true},
		{"missing clinic", &model.AuditLog{ActorType: model.AuditActorTypeSystem, Action: "a", Resource: "r"}, true},
		{"zero clinic", &model.AuditLog{ClinicID: u64(0), ActorType: model.AuditActorTypeSystem, Action: "a", Resource: "r"}, true},
		{"missing actor_type", &model.AuditLog{ClinicID: u64(1), Action: "a", Resource: "r"}, true},
		{"blank actor_type", &model.AuditLog{ClinicID: u64(1), ActorType: "  ", Action: "a", Resource: "r"}, true},
		{"missing action", &model.AuditLog{ClinicID: u64(1), ActorType: model.AuditActorTypeSystem, Resource: "r"}, true},
		{"missing resource", &model.AuditLog{ClinicID: u64(1), ActorType: model.AuditActorTypeSystem, Action: "a"}, true},
		{"staff without actor_id", &model.AuditLog{ClinicID: u64(1), ActorType: model.AuditActorTypeStaff, Action: "a", Resource: "r"}, true},
		{"staff with zero actor_id", &model.AuditLog{ClinicID: u64(1), ActorID: u64(0), ActorType: model.AuditActorTypeStaff, Action: "a", Resource: "r"}, true},
		{"staff valid", &model.AuditLog{ClinicID: u64(1), ActorID: u64(9), ActorType: model.AuditActorTypeStaff, Action: "a", Resource: "r"}, false},
		{"system valid", &model.AuditLog{ClinicID: u64(1), ActorType: model.AuditActorTypeSystem, Action: "a", Resource: "r"}, false},
		{"system must not carry actor_id", &model.AuditLog{ClinicID: u64(1), ActorID: u64(9), ActorType: model.AuditActorTypeSystem, Action: "a", Resource: "r"}, true},
		{"unknown actor_type", &model.AuditLog{ClinicID: u64(1), ActorType: "SECCHK20261009ROLE", Action: "a", Resource: "r"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := audit.ValidateLog(tc.log)
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, apperrors.IsInvalidInput(err), "validation failures must be InvalidInput-class, got %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ── C18: refused audit writes perform zero repository calls ──────────────────

func TestC18_RefusedAuditWriteHasZeroPersistenceCalls(t *testing.T) {
	repo := &countingRepo{}
	svc := audit.NewService(repo)

	// system actor carrying an actor_id violates the invariant.
	err := svc.Log(context.Background(), &model.AuditLog{
		ClinicID: u64(1), ActorID: u64(5), ActorType: model.AuditActorTypeSystem,
		Action: "SECCHK20261009ACTION", Resource: "SECCHK20261009RES",
	})
	require.Error(t, err)
	assert.Empty(t, repo.created, "rejected audit entry must not reach persistence")
	assert.Empty(t, repo.createdTx)

	// LogEntry on a nil input fails before any repository write.
	err = svc.LogEntry(context.Background(), nil)
	require.Error(t, err)
	assert.Empty(t, repo.created)
}

// ── C18: LogEntryTx goes through the tx channel; real repo fails closed ──────

func TestC18_LogEntryTxUsesAmbientTxChannel(t *testing.T) {
	repo := &countingRepo{}
	svc := audit.NewService(repo)

	err := svc.LogEntryTx(context.Background(), &audit.Entry{
		ClinicID:  u64(1),
		ActorType: model.AuditActorTypeSystem,
		Action:    "SECCHK20261009ACTION",
		Resource:  model.AuditResourceReservation,
	})
	require.NoError(t, err)
	assert.Empty(t, repo.created, "LogEntryTx must use the transaction-local channel")
	require.Len(t, repo.createdTx, 1)
	assert.Equal(t, "SECCHK20261009ACTION", repo.createdTx[0].Action)
}

func TestC18_RealRepositoryCreateTxFailsClosedWithoutAmbientTx(t *testing.T) {
	// repository.CreateTx checks persistence.TxFromContext BEFORE touching gorm:
	// with no ambient transaction the write is refused outright — the
	// zero-initialized *gorm.DB is never used, so no DB contact occurs.
	repo := audit.NewRepository(&gorm.DB{})
	err := repo.CreateTx(context.Background(), &model.AuditLog{
		ClinicID: u64(1), ActorType: model.AuditActorTypeSystem,
		Action: "SECCHK20261009ACTION", Resource: model.AuditResourceReservation,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, "INTERNAL", appErr.Code,
		"missing ambient transaction must be a fail-closed internal error")
}

// ── C18: user-agent is UTF-8 normalized and bounded at 1024 bytes ────────────

func TestC18_UserAgentBoundedAt1024Bytes(t *testing.T) {
	repo := &countingRepo{}
	svc := audit.NewService(repo)

	// 400 3-byte runes = 1200 bytes > 1024 — truncation must land on a rune boundary.
	long := strings.Repeat("あ", 400)
	err := svc.LogEntry(context.Background(), &audit.Entry{
		ClinicID: u64(1), ActorType: model.AuditActorTypeSystem,
		Action: "SECCHK20261009ACTION", Resource: model.AuditResourceReservation,
		UserAgent: long,
	})
	require.NoError(t, err)
	require.Len(t, repo.created, 1)
	got := repo.created[0].UserAgent
	assert.LessOrEqual(t, len(got), 1024)
	assert.True(t, utf8.ValidString(got), "truncated user-agent must remain valid UTF-8")
}

func TestC18_IPAddressNormalization(t *testing.T) {
	repo := &countingRepo{}
	svc := audit.NewService(repo)

	require.NoError(t, svc.LogEntry(context.Background(), &audit.Entry{
		ClinicID: u64(1), ActorType: model.AuditActorTypeSystem,
		Action: "a", Resource: "r", IPAddress: "",
	}))
	assert.Nil(t, repo.created[0].IPAddress, "empty IP must normalize to NULL (inet column)")

	require.NoError(t, svc.LogEntry(context.Background(), &audit.Entry{
		ClinicID: u64(1), ActorType: model.AuditActorTypeSystem,
		Action: "a", Resource: "r", IPAddress: "203.0.113.7",
	}))
	require.NotNil(t, repo.created[1].IPAddress)
	assert.Equal(t, "203.0.113.7", *repo.created[1].IPAddress)
}

// ── C18: structured old/new values are marshaled; actor type derives ─────────

func TestC18_LogMedicalRecordChangeMarshalsSyntheticValues(t *testing.T) {
	repo := &countingRepo{}
	svc := audit.NewService(repo)

	err := svc.LogMedicalRecordChange(context.Background(), 7, nil, "create", 9001,
		map[string]any{"marker": "SECCHK20261009OLD"},
		map[string]any{"marker": "SECCHK20261009NEW"},
	)
	require.NoError(t, err)
	require.Len(t, repo.created, 1)

	log := repo.created[0]
	assert.Equal(t, "create", log.Action)
	assert.Equal(t, "medical_record", log.Resource)
	assert.Equal(t, model.AuditActorTypeSystem, log.ActorType, "nil actorID must record as system actor")
	assert.Equal(t, u64(7), log.ClinicID)
	assert.Equal(t, u64(9001), log.ResourceID)

	var oldV, newV map[string]any
	require.NoError(t, json.Unmarshal(log.OldValue, &oldV))
	require.NoError(t, json.Unmarshal(log.NewValue, &newV))
	assert.Equal(t, "SECCHK20261009OLD", oldV["marker"])
	assert.Equal(t, "SECCHK20261009NEW", newV["marker"])
}

// ── C18: audit write failure surfaces a logged error, not a silent drop ──────

func TestC18_RepositoryFailurePropagates(t *testing.T) {
	repo := &countingRepo{createErr: apperrors.WrapInternalServerError("synthetic failure")}
	svc := audit.NewService(repo)

	err := svc.Log(context.Background(), &model.AuditLog{
		ClinicID: u64(1), ActorID: u64(3), ActorType: model.AuditActorTypeStaff,
		Action: "SECCHK20261009ACTION", Resource: model.AuditResourceReservation,
	})
	require.Error(t, err, "repository failure must propagate to the caller")
	assert.Empty(t, repo.created)
}
