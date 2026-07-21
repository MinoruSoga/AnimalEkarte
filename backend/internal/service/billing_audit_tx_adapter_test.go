package service

// billing_audit_tx_adapter_test.go — B③レビューMEDIUM対応: adapter の 9 field 写像を固定する
// （R④ smtpSendAdapter の field-mapping テストと同型・fake 注入テストばかりになる closure/adapter
// パターンでは写像そのものの検証を必ず 1 本残す）。

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturingAuditTxLogger struct {
	got *AuditLogInput
}

func (c *capturingAuditTxLogger) LogEntryTx(_ context.Context, input *AuditLogInput) error {
	c.got = input
	return nil
}

func TestBillingAuditTxAdapter_FieldMapping(t *testing.T) {
	clinicID, actorID, resourceID := uint64(1), uint64(2), uint64(3)
	inner := &capturingAuditTxLogger{}
	a := billingAuditTxAdapter{inner: inner}

	err := a.LogEntryTx(context.Background(), &billingAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  "staff",
		Action:     "create",
		Resource:   "billing_refund",
		ResourceID: &resourceID,
		OldValue:   map[string]any{"o": 1},
		NewValue:   map[string]any{"n": 2},
		Metadata:   map[string]any{"m": 3},
	})
	require.NoError(t, err)
	require.NotNil(t, inner.got)
	assert.Equal(t, &clinicID, inner.got.ClinicID)
	assert.Equal(t, &actorID, inner.got.ActorID)
	assert.Equal(t, "staff", inner.got.ActorType)
	assert.Equal(t, "create", inner.got.Action)
	assert.Equal(t, "billing_refund", inner.got.Resource)
	assert.Equal(t, &resourceID, inner.got.ResourceID)
	assert.Equal(t, map[string]any{"o": 1}, inner.got.OldValue)
	assert.Equal(t, map[string]any{"n": 2}, inner.got.NewValue)
}
