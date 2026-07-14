package service

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// mockAuditService は service.AuditService + AuditTxLogger のテスト用共有モック（BE7-15 正本）。
type mockAuditService struct {
	logAuthLoginFn func(ctx context.Context, clinicID, staffID *uint64, action, ip, ua string) error
	loggedActions  []string
	lastLogEntry   *AuditLogInput

	logEntryCalled   bool
	logEntryInput    *AuditLogInput
	logEntryTxCalled bool
	logEntryTxInput  *AuditLogInput
	logEntryErr      error
	logEntryTxErr    error
	entries          []*AuditLogInput

	logLstepOperationErr error
	logLstepOperationFn  func(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error

	logMedicalRecordChangeFn func(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error
	logVitalChangeFn         func(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error
	logAddendumCreateFn      func(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error
	calls                    []string
}

func (m *mockAuditService) Log(_ context.Context, _ *model.AuditLog) error { return nil }

func (m *mockAuditService) LogEntry(_ context.Context, input *AuditLogInput) error {
	m.logEntryCalled = true
	m.logEntryInput = input
	m.lastLogEntry = input
	m.entries = append(m.entries, input)
	return m.logEntryErr
}

func (m *mockAuditService) LogEntryTx(_ context.Context, input *AuditLogInput) error {
	m.logEntryTxCalled = true
	m.logEntryTxInput = input
	m.entries = append(m.entries, input)
	return m.logEntryTxErr
}

func (m *mockAuditService) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ip, ua string) error {
	m.loggedActions = append(m.loggedActions, action)
	if m.logAuthLoginFn != nil {
		return m.logAuthLoginFn(ctx, clinicID, staffID, action, ip, ua)
	}
	return nil
}

func (m *mockAuditService) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
	if m.logLstepOperationFn != nil {
		return m.logLstepOperationFn(ctx, clinicID, actorID, action, resource, resourceID)
	}
	if m.logLstepOperationErr != nil {
		return m.logLstepOperationErr
	}
	return nil
}

func (m *mockAuditService) LogLstepOperationWithMetadata(_ context.Context, _ uint64, _ *uint64, _, _ string, _ *uint64, _ any) error {
	return nil
}

func (m *mockAuditService) LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error {
	m.calls = append(m.calls, action)
	if m.logMedicalRecordChangeFn != nil {
		return m.logMedicalRecordChangeFn(ctx, clinicID, actorID, action, recordID, oldValue, newValue)
	}
	return nil
}

func (m *mockAuditService) LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error {
	m.calls = append(m.calls, action)
	if m.logVitalChangeFn != nil {
		return m.logVitalChangeFn(ctx, clinicID, actorID, action, vitalID, medicalRecordID, oldValue, newValue)
	}
	return nil
}

func (m *mockAuditService) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	m.calls = append(m.calls, "create")
	if m.logAddendumCreateFn != nil {
		return m.logAddendumCreateFn(ctx, clinicID, actorID, addendumID, medicalRecordID, addendum)
	}
	return nil
}

func (m *mockAuditService) LogClinicSwitch(_ context.Context, _ *uint64, _, _ uint64, _, _ string) error {
	return nil
}

// mockAuditRepository は repository.AuditRepository のテスト用共有モック（BE7-15 正本）。
type mockAuditRepository struct {
	createFn    func(ctx context.Context, log *model.AuditLog) error
	lastLogged  *model.AuditLog
	entries     []*model.AuditLog
	createTxErr error
}

func (m *mockAuditRepository) recordLog(log *model.AuditLog) {
	m.lastLogged = log
	m.entries = append(m.entries, log)
}

func (m *mockAuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	m.recordLog(log)
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

func (m *mockAuditRepository) CreateTx(ctx context.Context, log *model.AuditLog) error {
	if m.createTxErr != nil {
		return m.createTxErr
	}
	m.recordLog(log)
	if m.createFn != nil {
		return m.createFn(ctx, log)
	}
	return nil
}

// mockPermissionGroupRepository は repository.PermissionGroupRepository のテスト用共有モック（BE7-15 正本）。
type mockPermissionGroupRepository struct {
	findAllFn      func(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error)
	findByIDFn     func(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error)
	createFn       func(ctx context.Context, group *model.PermissionGroup) error
	updateFieldsFn func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error)
	deleteFn       func(ctx context.Context, clinicID, id uint64) error
	setRulesFn     func(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error

	countStaffsByGroupIDFn func(ctx context.Context, clinicID, groupID uint64) (int64, error)
	countUsageByGroupIDFn  func(ctx context.Context, clinicID, groupID uint64) (int64, error)

	reorderErr error
	reorderFn  func(ctx context.Context, clinicID uint64, ids []uint64) error

	getEffectivePermissionsByStaffID func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
	getEffectivePermissionsFn        func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)

	getGroupIDsByStaffIDFn     func(ctx context.Context, staffID uint64) ([]uint64, error)
	findAllGroupIDsByStaffIDFn func(ctx context.Context, staffID uint64) ([]uint64, error)

	setStaffGroupsFn    func(ctx context.Context, staffID uint64, groupIDs []uint64) error
	updateStaffGroupsFn func(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error
}

func (m *mockPermissionGroupRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PermissionGroup, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.PermissionGroup, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {
	if m.createFn != nil {
		return m.createFn(ctx, group)
	}
	return nil
}

func (m *mockPermissionGroupRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.PermissionGroup, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.PermissionGroup{}, nil
}

func (m *mockPermissionGroupRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockPermissionGroupRepository) UpdateRules(ctx context.Context, groupID uint64, rules []model.PermissionGroupRule) error {
	if m.setRulesFn != nil {
		return m.setRulesFn(ctx, groupID, rules)
	}
	return nil
}

func (m *mockPermissionGroupRepository) CountUsageByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
	if m.countStaffsByGroupIDFn != nil {
		return m.countStaffsByGroupIDFn(ctx, clinicID, groupID)
	}
	if m.countUsageByGroupIDFn != nil {
		return m.countUsageByGroupIDFn(ctx, clinicID, groupID)
	}
	return 0, nil
}

func (m *mockPermissionGroupRepository) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, clinicID, ids)
	}
	return m.reorderErr
}

func (m *mockPermissionGroupRepository) FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsByStaffID != nil {
		return m.getEffectivePermissionsByStaffID(ctx, staffID, clinicID)
	}
	if m.getEffectivePermissionsFn != nil {
		return m.getEffectivePermissionsFn(ctx, staffID, clinicID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) FindAllGroupIDsByStaffID(ctx context.Context, staffID uint64) ([]uint64, error) {
	if m.findAllGroupIDsByStaffIDFn != nil {
		return m.findAllGroupIDsByStaffIDFn(ctx, staffID)
	}
	if m.getGroupIDsByStaffIDFn != nil {
		return m.getGroupIDsByStaffIDFn(ctx, staffID)
	}
	return nil, nil
}

func (m *mockPermissionGroupRepository) UpdateStaffGroups(ctx context.Context, clinicID, staffID uint64, groupIDs []uint64) error {
	if m.updateStaffGroupsFn != nil {
		return m.updateStaffGroupsFn(ctx, clinicID, staffID, groupIDs)
	}
	if m.setStaffGroupsFn != nil {
		return m.setStaffGroupsFn(ctx, staffID, groupIDs)
	}
	return nil
}

// mockOccupationRepository は repository.OccupationRepository のテスト用共有モック（BE7-15 正本）。
type mockOccupationRepository struct {
	findAllFn                  func(ctx context.Context, clinicID uint64) ([]model.Occupation, error)
	findByIDFn                 func(ctx context.Context, clinicID, id uint64) (*model.Occupation, error)
	createFn                   func(ctx context.Context, occupation *model.Occupation) error
	updateFieldsFn             func(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error)
	deleteFn                   func(ctx context.Context, clinicID, id uint64) error
	reorderErr                 error
	countUsageByOccupationIDFn func(ctx context.Context, clinicID, occupationID uint64) (int64, error)
}

func (m *mockOccupationRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.Occupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID)
	}
	return []model.Occupation{}, nil
}

func (m *mockOccupationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Occupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, id)
	}
	return &model.Occupation{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOccupationRepository) Create(ctx context.Context, occupation *model.Occupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, occupation)
	}
	return nil
}

func (m *mockOccupationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Occupation, error) {
	if m.updateFieldsFn != nil {
		return m.updateFieldsFn(ctx, clinicID, id, fields)
	}
	return &model.Occupation{}, nil
}

func (m *mockOccupationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, id)
	}
	return nil
}

func (m *mockOccupationRepository) Reorder(_ context.Context, _ uint64, _ []uint64) error {
	return m.reorderErr
}

func (m *mockOccupationRepository) CountUsageByOccupationID(ctx context.Context, clinicID, id uint64) (int64, error) {
	if m.countUsageByOccupationIDFn != nil {
		return m.countUsageByOccupationIDFn(ctx, clinicID, id)
	}
	return 0, nil
}

// mockReservationTypeOccupationRepository は repository.ReservationTypeOccupationRepository のテスト用共有モック（BE7-15 正本）。
type mockReservationTypeOccupationRepository struct {
	findAllFn         func(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	findByIDFn        func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	createFn          func(ctx context.Context, o *model.ReservationTypeOccupation) error
	deleteFn          func(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
	countByStaffIDsFn func(ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time) (map[string]int64, error)
}

func (m *mockReservationTypeOccupationRepository) FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx, clinicID, reservationTypeID)
	}
	return []model.ReservationTypeOccupation{}, nil
}

func (m *mockReservationTypeOccupationRepository) FindByID(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return &model.ReservationTypeOccupation{ID: 1, ClinicID: clinicID, ReservationTypeID: reservationTypeID, OccupationID: occupationID}, nil
}

func (m *mockReservationTypeOccupationRepository) Create(ctx context.Context, o *model.ReservationTypeOccupation) error {
	if m.createFn != nil {
		return m.createFn(ctx, o)
	}
	return nil
}

func (m *mockReservationTypeOccupationRepository) Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, clinicID, reservationTypeID, occupationID)
	}
	return nil
}

func (m *mockReservationTypeOccupationRepository) CountWorkingStaffByReservationTypeIDs(ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time) (map[string]int64, error) {
	if m.countByStaffIDsFn != nil {
		return m.countByStaffIDsFn(ctx, clinicID, reservationTypeID, dates)
	}
	result := make(map[string]int64, len(dates))
	for _, d := range dates {
		result[d.Format("2006-01-02")] = 1
	}
	return result, nil
}
