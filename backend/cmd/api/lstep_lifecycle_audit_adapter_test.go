package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/lstep"
)

type capturingLstepLifecycleAuditTxLogger struct {
	ctx context.Context
	got *audit.Entry
	err error
}

func (c *capturingLstepLifecycleAuditTxLogger) LogEntryTx(ctx context.Context, input *audit.Entry) error {
	c.ctx = ctx
	c.got = input
	return c.err
}

func TestLstepLifecycleAuditTxAdapter_FieldMappingAndContext(t *testing.T) {
	type contextKey string
	const txKey contextKey = "tx"

	clinicID, actorID, resourceID := uint64(1), uint64(2), uint64(3)
	wantErr := errors.New("audit write failed")
	inner := &capturingLstepLifecycleAuditTxLogger{err: wantErr}
	adapter := lstepLifecycleAuditTxAdapter{inner: inner}
	ctx := context.WithValue(context.Background(), txKey, "ambient-transaction")

	err := adapter.LogEntryTx(ctx, &lstep.LifecycleAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  "staff",
		Action:     "pet_death",
		Resource:   "pet",
		ResourceID: &resourceID,
	})

	require.ErrorIs(t, err, wantErr)
	require.NotNil(t, inner.got)
	assert.Same(t, ctx, inner.ctx)
	assert.Equal(t, &clinicID, inner.got.ClinicID)
	assert.Equal(t, &actorID, inner.got.ActorID)
	assert.Equal(t, "staff", inner.got.ActorType)
	assert.Equal(t, "pet_death", inner.got.Action)
	assert.Equal(t, "pet", inner.got.Resource)
	assert.Equal(t, &resourceID, inner.got.ResourceID)
}
