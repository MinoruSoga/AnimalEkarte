package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrimmingAuditTxAdapter_FieldMapping(t *testing.T) {
	clinicID, actorID, resourceID := uint64(1), uint64(2), uint64(3)
	inner := &capturingAuditTxLogger{}
	adapter := trimmingAuditTxAdapter{inner: inner}
	oldValue := map[string]any{"old": true}
	newValue := map[string]any{"new": true}
	metadata := map[string]any{"mutation_path": "delete"}

	err := adapter.LogEntryTx(context.Background(), &trimmingAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  "staff",
		Action:     "trimming.delete",
		Resource:   "trimming",
		ResourceID: &resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Metadata:   metadata,
	})

	require.NoError(t, err)
	require.NotNil(t, inner.got)
	assert.Equal(t, &clinicID, inner.got.ClinicID)
	assert.Equal(t, &actorID, inner.got.ActorID)
	assert.Equal(t, "staff", inner.got.ActorType)
	assert.Equal(t, "trimming.delete", inner.got.Action)
	assert.Equal(t, "trimming", inner.got.Resource)
	assert.Equal(t, &resourceID, inner.got.ResourceID)
	assert.Equal(t, oldValue, inner.got.OldValue)
	assert.Equal(t, newValue, inner.got.NewValue)
	assert.Equal(t, metadata, inner.got.Metadata)
}
