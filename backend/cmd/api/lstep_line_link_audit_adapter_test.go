package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
)

type capturingLineLinkAuditTxLogger struct {
	ctx   context.Context
	entry *audit.Entry
	err   error
}

func (c *capturingLineLinkAuditTxLogger) LogEntryTx(ctx context.Context, entry *audit.Entry) error {
	c.ctx = ctx
	c.entry = entry
	return c.err
}

func TestLstepLineLinkAuditTxAdapter_MapsSafeSuccessEventAndContext(t *testing.T) {
	type contextKey string
	const txKey contextKey = "tx"
	wantErr := errors.New("audit failed")
	inner := &capturingLineLinkAuditTxLogger{err: wantErr}
	adapter := lstepLineLinkAuditTxAdapter{inner: inner}
	ctx := context.WithValue(context.Background(), txKey, "ambient")

	err := adapter.LogOwnerLineLinkTx(ctx, 11, 29)

	require.ErrorIs(t, err, wantErr)
	assert.Same(t, ctx, inner.ctx)
	require.NotNil(t, inner.entry)
	require.NotNil(t, inner.entry.ClinicID)
	require.NotNil(t, inner.entry.ResourceID)
	assert.Equal(t, uint64(11), *inner.entry.ClinicID)
	assert.Nil(t, inner.entry.ActorID)
	assert.Equal(t, model.AuditActorTypeSystem, inner.entry.ActorType)
	assert.Equal(t, model.AuditActionOwnerLineUserIDUpdate, inner.entry.Action)
	assert.Equal(t, "owner", inner.entry.Resource)
	assert.Equal(t, uint64(29), *inner.entry.ResourceID)
	assert.Equal(t, map[string]any{"linked": false}, inner.entry.OldValue)
	assert.Equal(t, map[string]any{"linked": true}, inner.entry.NewValue)
}
