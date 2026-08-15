package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/auth"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/staff"
)

type credentialAuditAccountReaderStub struct {
	account *model.Account
	err     error
}

func (s credentialAuditAccountReaderStub) GetByID(
	context.Context,
	uint64,
) (*model.Account, error) {
	return s.account, s.err
}

type credentialAuditStaffReaderStub struct {
	staff *model.Staff
	err   error
}

func (s credentialAuditStaffReaderStub) FindByAccountID(
	context.Context,
	uint64,
) (*model.Staff, error) {
	return s.staff, s.err
}

type credentialAuditAssignmentReaderStub struct {
	assignments []model.StaffClinicAssignment
	err         error
}

func (s credentialAuditAssignmentReaderStub) FindAllByStaffID(
	context.Context,
	uint64,
) ([]model.StaffClinicAssignment, error) {
	return append([]model.StaffClinicAssignment(nil), s.assignments...), s.err
}

type credentialAuditClinicReaderStub struct {
	clinics []model.Clinic
	err     error
}

func (s credentialAuditClinicReaderStub) ListClinics(
	context.Context,
) ([]model.Clinic, error) {
	return append([]model.Clinic(nil), s.clinics...), s.err
}

type credentialAuditServiceCapture struct {
	authClinicID *uint64
	authStaffID  *uint64
	authAction   string
	authIP       string
	authAgent    string
	entry        *audit.Entry
	txEntry      *audit.Entry
	txContext    context.Context
}

func (*credentialAuditServiceCapture) Log(
	context.Context,
	*model.AuditLog,
) error {
	return nil
}

func (a *credentialAuditServiceCapture) LogEntry(
	_ context.Context,
	entry *audit.Entry,
) error {
	a.entry = entry
	return nil
}

func (a *credentialAuditServiceCapture) LogEntryTx(
	ctx context.Context,
	entry *audit.Entry,
) error {
	a.txContext = ctx
	a.txEntry = entry
	return nil
}

func (a *credentialAuditServiceCapture) LogAuthLogin(
	_ context.Context,
	clinicID, staffID *uint64,
	action, ipAddress, userAgent string,
) error {
	a.authClinicID = clinicID
	a.authStaffID = staffID
	a.authAction = action
	a.authIP = ipAddress
	a.authAgent = userAgent
	return nil
}

func (*credentialAuditServiceCapture) LogLstepOperation(
	context.Context,
	uint64,
	*uint64,
	string,
	string,
	*uint64,
) error {
	return nil
}

func (*credentialAuditServiceCapture) LogLstepOperationWithMetadata(
	context.Context,
	uint64,
	*uint64,
	string,
	string,
	*uint64,
	any,
) error {
	return nil
}

func (*credentialAuditServiceCapture) LogMedicalRecordChange(
	context.Context,
	uint64,
	*uint64,
	string,
	uint64,
	map[string]any,
	map[string]any,
) error {
	return nil
}

func (*credentialAuditServiceCapture) LogVitalChange(
	context.Context,
	uint64,
	*uint64,
	string,
	uint64,
	uint64,
	map[string]any,
	map[string]any,
) error {
	return nil
}

func (*credentialAuditServiceCapture) LogAddendumCreate(
	context.Context,
	uint64,
	*uint64,
	uint64,
	uint64,
	*model.MedicalRecordAddendum,
) error {
	return nil
}

func (*credentialAuditServiceCapture) LogClinicSwitch(
	context.Context,
	*uint64,
	uint64,
	uint64,
	string,
	string,
) error {
	return nil
}

func TestAuthHTTPAuditAdapterMapsCredentialEntry(t *testing.T) {
	clinicID := uint64(23)
	actorID := uint64(17)
	accountID := uint64(41)
	logger := &credentialAuditServiceCapture{}
	adapter := authHTTPAuditAdapter{logger: logger}

	err := adapter.LogEntry(context.Background(), auth.AuthAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionAuthPasswordChange,
		Resource:   model.AuditResourceAccount,
		ResourceID: &accountID,
		NewValue:   map[string]any{"staff_id": uint64(17)},
		IPAddress:  "192.0.2.1",
		UserAgent:  "auth-adapter-test",
	})

	require.NoError(t, err)
	require.NotNil(t, logger.entry)
	assert.Equal(t, &clinicID, logger.entry.ClinicID)
	assert.Equal(t, &actorID, logger.entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, logger.entry.ActorType)
	assert.Equal(t, model.AuditActionAuthPasswordChange, logger.entry.Action)
	assert.Equal(t, model.AuditResourceAccount, logger.entry.Resource)
	assert.Equal(t, &accountID, logger.entry.ResourceID)
	assert.Equal(t, map[string]any{"staff_id": uint64(17)}, logger.entry.NewValue)
	assert.Equal(t, "192.0.2.1", logger.entry.IPAddress)
	assert.Equal(t, "auth-adapter-test", logger.entry.UserAgent)
}

func TestAuthHTTPAuditAdapterPreservesLoginAudit(t *testing.T) {
	clinicID := uint64(23)
	staffID := uint64(17)
	logger := &credentialAuditServiceCapture{}
	adapter := authHTTPAuditAdapter{logger: logger}

	err := adapter.LogAuthLogin(
		context.Background(),
		&clinicID,
		&staffID,
		model.AuditActionAuthLoginSuccess,
		"192.0.2.1",
		"auth-adapter-test",
	)

	require.NoError(t, err)
	assert.Equal(t, &clinicID, logger.authClinicID)
	assert.Equal(t, &staffID, logger.authStaffID)
	assert.Equal(t, model.AuditActionAuthLoginSuccess, logger.authAction)
	assert.Equal(t, "192.0.2.1", logger.authIP)
	assert.Equal(t, "auth-adapter-test", logger.authAgent)
}

func TestAuthCredentialAuditTxAdapterMapsEntryInCallerContext(t *testing.T) {
	type txMarker struct{}
	ctx := context.WithValue(context.Background(), txMarker{}, true)
	clinicID := uint64(23)
	actorID := uint64(17)
	accountID := uint64(41)
	logger := &credentialAuditServiceCapture{}
	adapter := authCredentialAuditTxAdapter{logger: logger}

	err := adapter.LogEntryTx(ctx, auth.AuthAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionAuthPasswordChange,
		Resource:   model.AuditResourceAccount,
		ResourceID: &accountID,
		NewValue:   map[string]any{"staff_id": uint64(17)},
		IPAddress:  "192.0.2.1",
		UserAgent:  "auth-tx-adapter-test",
	})

	require.NoError(t, err)
	assert.Same(t, ctx, logger.txContext)
	require.NotNil(t, logger.txEntry)
	assert.Equal(t, &clinicID, logger.txEntry.ClinicID)
	assert.Equal(t, &actorID, logger.txEntry.ActorID)
	assert.Equal(t, model.AuditActionAuthPasswordChange, logger.txEntry.Action)
	assert.Equal(t, &accountID, logger.txEntry.ResourceID)
}

func TestStaffCredentialAuditAdapterMapsEntryInCallerContext(t *testing.T) {
	type txMarker struct{}
	ctx := context.WithValue(context.Background(), txMarker{}, true)
	clinicID := uint64(23)
	actorID := uint64(17)
	accountID := uint64(41)
	logger := &credentialAuditServiceCapture{}
	adapter := staffCredentialAuditAdapter{logger: logger}

	err := adapter.LogEntryTx(ctx, staff.CredentialAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionAuthPasswordAdminReplace,
		Resource:   model.AuditResourceAccount,
		ResourceID: &accountID,
		NewValue:   map[string]any{"staff_id": uint64(29)},
		IPAddress:  "192.0.2.29",
		UserAgent:  "staff-adapter-test",
	})

	require.NoError(t, err)
	assert.Same(t, ctx, logger.txContext)
	require.NotNil(t, logger.txEntry)
	assert.Equal(t, &clinicID, logger.txEntry.ClinicID)
	assert.Equal(t, &actorID, logger.txEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, logger.txEntry.ActorType)
	assert.Equal(t, model.AuditActionAuthPasswordAdminReplace, logger.txEntry.Action)
	assert.Equal(t, model.AuditResourceAccount, logger.txEntry.Resource)
	assert.Equal(t, &accountID, logger.txEntry.ResourceID)
	assert.Equal(t, map[string]any{"staff_id": uint64(29)}, logger.txEntry.NewValue)
	assert.Equal(t, "192.0.2.29", logger.txEntry.IPAddress)
	assert.Equal(t, "staff-adapter-test", logger.txEntry.UserAgent)
}

func TestStaffPermissionAssignmentAuditAdapterMapsOldAndNewValues(t *testing.T) {
	type txMarker struct{}
	ctx := context.WithValue(context.Background(), txMarker{}, true)
	clinicID := uint64(23)
	actorID := uint64(17)
	targetStaffID := uint64(29)
	logger := &credentialAuditServiceCapture{}
	adapter := staffPermissionAssignmentAuditAdapter{logger: logger}
	oldValue := map[string]any{"staff_id": targetStaffID, "group_ids": []uint64{2}}
	newValue := map[string]any{"staff_id": targetStaffID, "group_ids": []uint64{3, 5}}

	err := adapter.LogEntryTx(ctx, &staff.PermissionAssignmentAuditEntry{
		ClinicID:   &clinicID,
		ActorID:    &actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     model.AuditActionStaffPermissionGroupsReplace,
		Resource:   model.AuditResourceStaff,
		ResourceID: &targetStaffID,
		OldValue:   oldValue,
		NewValue:   newValue,
		IPAddress:  "192.0.2.17",
		UserAgent:  "staff-permission-adapter-test",
	})

	require.NoError(t, err)
	assert.Same(t, ctx, logger.txContext)
	require.NotNil(t, logger.txEntry)
	assert.Equal(t, oldValue, logger.txEntry.OldValue)
	assert.Equal(t, newValue, logger.txEntry.NewValue)
	assert.Equal(t, model.AuditActionStaffPermissionGroupsReplace, logger.txEntry.Action)
	assert.Equal(t, model.AuditResourceStaff, logger.txEntry.Resource)
}

func TestAuthCredentialAuditSubjectResolverUsesActiveExistingClinic(
	t *testing.T,
) {
	accountID := uint64(41)
	resolver := authCredentialAuditSubjectResolver{
		accounts: credentialAuditAccountReaderStub{account: &model.Account{
			ID:        accountID,
			IsActive:  true,
			UpdatedAt: time.Now(),
		}},
		staff: credentialAuditStaffReaderStub{staff: &model.Staff{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		assignments: credentialAuditAssignmentReaderStub{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 22, IsMain: true},
				{StaffID: 17, ClinicID: 23},
			},
		},
		clinics: credentialAuditClinicReaderStub{clinics: []model.Clinic{
			{ID: 22, IsActive: false},
			{ID: 23, IsActive: true},
		}},
	}

	subject, err := resolver.ResolveCredentialAuditSubject(
		context.Background(),
		accountID,
	)

	require.NoError(t, err)
	assert.Equal(t, auth.CredentialAuditSubject{
		ClinicID: 23,
		StaffID:  17,
	}, subject)
}

func TestAuthCredentialAuditSubjectResolverSystemAdminFallsBackToActiveClinic(
	t *testing.T,
) {
	accountID := uint64(41)
	resolver := authCredentialAuditSubjectResolver{
		accounts: credentialAuditAccountReaderStub{account: &model.Account{
			ID:            accountID,
			IsActive:      true,
			IsSystemAdmin: true,
			UpdatedAt:     time.Now(),
		}},
		staff: credentialAuditStaffReaderStub{staff: &model.Staff{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		assignments: credentialAuditAssignmentReaderStub{},
		clinics: credentialAuditClinicReaderStub{clinics: []model.Clinic{
			{ID: 0, IsActive: true},
			{ID: 22, IsActive: false},
			{ID: 24, IsActive: true},
		}},
	}

	subject, err := resolver.ResolveCredentialAuditSubject(
		context.Background(),
		accountID,
	)

	require.NoError(t, err)
	assert.Equal(t, auth.CredentialAuditSubject{
		ClinicID: 24,
		StaffID:  17,
	}, subject)
}

func TestAuthCredentialAuditSubjectResolverFailsClosedWithoutActiveClinic(
	t *testing.T,
) {
	accountID := uint64(41)
	resolver := authCredentialAuditSubjectResolver{
		accounts: credentialAuditAccountReaderStub{account: &model.Account{
			ID:        accountID,
			IsActive:  true,
			UpdatedAt: time.Now(),
		}},
		staff: credentialAuditStaffReaderStub{staff: &model.Staff{
			ID:        17,
			AccountID: &accountID,
			IsActive:  true,
		}},
		assignments: credentialAuditAssignmentReaderStub{
			assignments: []model.StaffClinicAssignment{
				{StaffID: 17, ClinicID: 22, IsMain: true},
			},
		},
		clinics: credentialAuditClinicReaderStub{clinics: []model.Clinic{
			{ID: 22, IsActive: false},
		}},
	}

	subject, err := resolver.ResolveCredentialAuditSubject(
		context.Background(),
		accountID,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrForbidden)
	assert.Zero(t, subject)
}
