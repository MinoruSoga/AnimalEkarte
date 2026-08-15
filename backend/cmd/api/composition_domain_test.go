package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type capturingBillingAuditTxLogger struct {
	got *audit.Entry
}

func (l *capturingBillingAuditTxLogger) LogEntryTx(
	_ context.Context,
	entry *audit.Entry,
) error {
	l.got = entry
	return nil
}

func TestNewMedicalRecordCompositionBuildsDomainGraph(t *testing.T) {
	db := &gorm.DB{}
	repositories := newMedicalRecordRepositories(db)
	staffRepositories := newStaffRepositories(db)
	reservationRepositories := newReservationRepositories(
		db,
		staffRepositories.Staff,
		staffRepositories.ShiftEntries,
		staffRepositories.Occupations,
	)
	composition := newMedicalRecordComposition(
		repositories,
		medicalRecordCompositionDependencies{
			Transactor:       persistence.NewTransactor(db),
			Reservations:     reservationRepositories.Reservations,
			Staff:            staffRepositories.Staff,
			StaffAssignments: staffRepositories.Assignments,
		},
	)

	assert.NotNil(t, repositories.medicalRecords)
	assert.NotNil(t, repositories.hospitalizations)
	assert.NotNil(t, repositories.treatments)
	assert.NotNil(t, composition.MedicalRecords)
	assert.NotNil(t, composition.Checkups)
	assert.NotNil(t, composition.DrainCheckups)
	assert.NotNil(t, composition.newHandler(medicalRecordHTTPDependencies{}))

	requireServiceDependency(t, composition.services.preventive.checkups, "relationVerifier")
	requireServiceDependency(t, composition.services.preventive.checkups, "transactor")
	requireServiceDependency(t, composition.services.core.examinations, "relations")
	requireServiceDependency(t, composition.services.core.examinations, "transactor")
	requireServiceDependency(t, composition.services.clinical.vitals, "relations")
	requireServiceDependency(t, composition.services.clinical.vitals, "staffs")
	requireServiceDependency(t, composition.services.clinical.vitals, "staffAssignments")
	requireServiceDependency(t, composition.services.clinical.images, "examinations")
	requireServiceDependency(t, composition.services.clinical.images, "staffs")
	requireServiceDependency(t, composition.services.clinical.images, "staffAssignments")
	requireServiceDependency(t, composition.services.hospital.dailyRecords, "staffs")
	requireServiceDependency(t, composition.services.hospital.dailyRecords, "staffAssignments")
}

func requireServiceDependency(
	t *testing.T,
	service any,
	fieldName string,
) {
	t.Helper()
	value := reflect.ValueOf(service)
	require.Equal(t, reflect.Pointer, value.Kind())
	field := value.Elem().FieldByName(fieldName)
	require.True(t, field.IsValid(), "service dependency field %q must exist", fieldName)
	require.Contains(
		t,
		[]reflect.Kind{reflect.Interface, reflect.Pointer, reflect.Func, reflect.Map, reflect.Slice},
		field.Kind(),
		"service dependency field %q must be nil-able",
		fieldName,
	)
	require.False(t, field.IsNil(), "service dependency %q must be wired", fieldName)
}

func TestNewBillingCompositionBuildsDomainGraph(t *testing.T) {
	db := &gorm.DB{}
	repositories := newBillingRepositories(db)
	composition := newBillingComposition(
		repositories,
		billingCompositionDependencies{
			Transactor: persistence.NewTransactor(db),
		},
	)

	assert.NotNil(t, repositories.accounting)
	assert.NotNil(t, repositories.billingItems)
	assert.NotNil(t, repositories.insurance)
	assert.NotNil(t, composition.Accounting)
	assert.NotNil(t, composition.BillingItems)
	assert.NotNil(t, composition.Insurance)
	assert.NotNil(t, composition.CashRegister)
	assert.NotNil(t, composition.newHandler(nil, nil))
}

func TestBillingAuditTxBridge_PreservesDomainEntryFields(t *testing.T) {
	clinicID, actorID, resourceID := uint64(1), uint64(2), uint64(3)
	logger := &capturingBillingAuditTxLogger{}
	bridge := billingAuditTxBridge{logger: logger}

	err := bridge.LogEntryTx(context.Background(), &billing.AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  "staff",
		Action:     "create",
		Resource:   "billing_refund",
		ResourceID: &resourceID,
		OldValue:   map[string]any{"old": 1},
		NewValue:   map[string]any{"new": 2},
		Metadata:   map[string]any{"source": "test"},
	})

	require.NoError(t, err)
	require.NotNil(t, logger.got)
	assert.Equal(t, &clinicID, logger.got.ClinicID)
	assert.Equal(t, &actorID, logger.got.ActorID)
	assert.Equal(t, "staff", logger.got.ActorType)
	assert.Equal(t, "create", logger.got.Action)
	assert.Equal(t, "billing_refund", logger.got.Resource)
	assert.Equal(t, &resourceID, logger.got.ResourceID)
	assert.Equal(t, map[string]any{"old": 1}, logger.got.OldValue)
	assert.Equal(t, map[string]any{"new": 2}, logger.got.NewValue)
	assert.Equal(t, map[string]any{"source": "test"}, logger.got.Metadata)
}
