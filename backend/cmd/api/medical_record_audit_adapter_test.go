package main

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

type capturingMedicalRecordAuditService struct {
	audit.Service
	entry *audit.Entry
	err   error
}

func (c *capturingMedicalRecordAuditService) LogEntry(_ context.Context, entry *audit.Entry) error {
	c.entry = entry
	return c.err
}

func TestMedicalRecordAuditAdapter_LogEntry(t *testing.T) {
	clinicID := uint64(3)
	reservationID := uint64(77)
	wantErr := errors.New("audit failed")
	inner := &capturingMedicalRecordAuditService{err: wantErr}
	adapter := medicalRecordAuditBridge{logger: inner}
	metadata := map[string]any{"failure_category": "internal_error"}

	err := adapter.LogEntry(context.Background(), &medicalrecord.AuditEntry{
		ClinicID:   &clinicID,
		ActorType:  "system",
		Action:     "reservation.draft_cleanup_failed",
		Resource:   "reservation",
		ResourceID: &reservationID,
		Metadata:   metadata,
	})

	require.ErrorIs(t, err, wantErr)
	require.NotNil(t, inner.entry)
	assert.Equal(t, &clinicID, inner.entry.ClinicID)
	assert.Equal(t, "system", inner.entry.ActorType)
	assert.Equal(t, "reservation.draft_cleanup_failed", inner.entry.Action)
	assert.Equal(t, "reservation", inner.entry.Resource)
	assert.Equal(t, &reservationID, inner.entry.ResourceID)
	assert.Equal(t, metadata, inner.entry.Metadata)
}

func TestRuntimeCompositionWiresMedicalRecordCleanupIntoReservationServices(t *testing.T) {
	file, err := parser.ParseFile(
		token.NewFileSet(),
		"composition_runtime.go",
		nil,
		0,
	)
	require.NoError(t, err)
	assert.True(
		t,
		hasReservationMedicalRecordWiring(file),
		"runtime composition must inject the target medical-record service into reservation",
	)
}

func hasReservationMedicalRecordWiring(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		identifier, ok := literal.Type.(*ast.Ident)
		if !ok || identifier.Name != "reservationServiceDependencies" {
			return true
		}
		for _, element := range literal.Elts {
			field, ok := element.(*ast.KeyValueExpr)
			if !ok || !isIdentifier(field.Key, "MedicalRecords") {
				continue
			}
			selector, ok := field.Value.(*ast.SelectorExpr)
			if ok &&
				isIdentifier(selector.X, "medicalRecordComposition") &&
				selector.Sel.Name == "MedicalRecords" {
				found = true
			}
		}
		return true
	})
	return found
}

func isIdentifier(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == name
}
