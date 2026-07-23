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

	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/service"
)

type capturingMedicalRecordAuditService struct {
	service.AuditService
	entry *service.AuditLogInput
	err   error
}

func (c *capturingMedicalRecordAuditService) LogEntry(_ context.Context, entry *service.AuditLogInput) error {
	c.entry = entry
	return c.err
}

func TestMedicalRecordAuditAdapter_LogEntry(t *testing.T) {
	clinicID := uint64(3)
	reservationID := uint64(77)
	wantErr := errors.New("audit failed")
	inner := &capturingMedicalRecordAuditService{err: wantErr}
	adapter := medicalRecordAuditAdapter{AuditService: inner}
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

func TestMainWiresMedicalRecordCleanupIntoReservationCancellationServices(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "main.go", nil, 0)
	require.NoError(t, err)

	contracts := []struct {
		serviceField string
		constructor  string
	}{
		{
			serviceField: "ReservationAdmin",
			constructor:  "NewReservationAdminServiceWithMedicalRecord",
		},
		{
			serviceField: "Liff",
			constructor:  "NewLiffServiceWithType",
		},
	}
	for _, contract := range contracts {
		t.Run(contract.serviceField, func(t *testing.T) {
			assert.True(
				t,
				hasReservationMedicalRecordWiring(file, contract.serviceField, contract.constructor),
				"main must assign svcs.%s using reservation.%s with medicalRecordSvc",
				contract.serviceField,
				contract.constructor,
			)
		})
	}
}

func hasReservationMedicalRecordWiring(file *ast.File, serviceField, constructor string) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		lhs, ok := assign.Lhs[0].(*ast.SelectorExpr)
		if !ok || !isIdentifier(lhs.X, "svcs") || lhs.Sel.Name != serviceField {
			return true
		}
		call, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		constructorSelector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !isIdentifier(constructorSelector.X, "reservation") ||
			constructorSelector.Sel.Name != constructor {
			return true
		}
		for _, argument := range call.Args {
			if isIdentifier(argument, "medicalRecordSvc") {
				found = true
				return false
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
