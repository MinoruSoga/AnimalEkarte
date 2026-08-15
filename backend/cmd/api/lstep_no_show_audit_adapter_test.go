package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestLstepNoShowAuditTxAdapter_FieldMappingAndContext(t *testing.T) {
	type contextKey string
	const txKey contextKey = "tx"

	fixedTime := time.Date(2026, 7, 22, 12, 34, 56, 0, time.UTC)
	wantErr := errors.New("audit write failed")
	inner := &capturingLstepLifecycleAuditTxLogger{err: wantErr}
	adapter := lstepNoShowAuditTxAdapter{inner: inner}
	ctx := context.WithValue(context.Background(), txKey, "ambient-transaction")

	err := adapter.LogNoShowTransitionTx(ctx, &lstep.NoShowAuditEntry{
		ClinicID:       11,
		AppointmentID:  29,
		PreviousStatus: model.ReservationStatusConfirmed,
		EvaluatedAt:    fixedTime,
		RuleVersion:    "rule-v1",
		BatchRunID:     "batch-123",
	})

	require.ErrorIs(t, err, wantErr)
	require.NotNil(t, inner.got)
	assert.Same(t, ctx, inner.ctx)
	assert.Equal(t, uint64(11), *inner.got.ClinicID)
	assert.Nil(t, inner.got.ActorID)
	assert.Equal(t, model.AuditActorTypeSystem, inner.got.ActorType)
	assert.Equal(t, model.AuditActionReservationNoShow, inner.got.Action)
	assert.Equal(t, model.AuditResourceReservation, inner.got.Resource)
	assert.Equal(t, uint64(29), *inner.got.ResourceID)
	assert.Equal(t, map[string]any{"status": model.ReservationStatusConfirmed}, inner.got.OldValue)
	assert.Equal(t, map[string]any{"status": model.ReservationStatusNoShow}, inner.got.NewValue)
	assert.Equal(t, map[string]any{
		"evaluated_at": fixedTime,
		"rule_version": "rule-v1",
		"batch_run_id": "batch-123",
	}, inner.got.Metadata)
}
