package audit

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

const maxAuditUserAgentBytes = 1024

// Service is the regular audit-writing contract.
type Service interface {
	Log(ctx context.Context, log *model.AuditLog) error
	LogEntry(ctx context.Context, input *Entry) error
	LogAuthLogin(ctx context.Context, clinicID *uint64, staffID *uint64, action string, ipAddress string, userAgent string) error
	LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
	LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
	LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error
	LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error
	LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error
	LogClinicSwitch(ctx context.Context, actorID *uint64, fromClinicID, toClinicID uint64, ipAddress, userAgent string) error
}

// TxLogger is the fail-closed, ambient-transaction audit contract.
type TxLogger interface {
	LogEntryTx(ctx context.Context, input *Entry) error
}

// Kernel combines regular and transaction-local audit writes.
type Kernel interface {
	Service
	TxLogger
}

// Entry is the transport-neutral input used to construct an audit log.
type Entry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
	IPAddress  string
	UserAgent  string
}

type service struct {
	repository Repository
}

var _ Kernel = (*service)(nil)

// NewService constructs the audit kernel.
func NewService(repository Repository) Kernel {
	return &service{repository: repository}
}

// ValidateLog validates the invariant required by the audit_logs DDL.
func ValidateLog(log *model.AuditLog) error {
	if log == nil {
		return apperrors.WrapInvalidInput("audit log is required")
	}
	if log.ClinicID == nil || *log.ClinicID == 0 {
		return apperrors.WrapInvalidInput("audit log clinic_id is required")
	}
	if strings.TrimSpace(log.ActorType) == "" {
		return apperrors.WrapInvalidInput("audit log actor_type is required")
	}
	if strings.TrimSpace(log.Action) == "" {
		return apperrors.WrapInvalidInput("audit log action is required")
	}
	if strings.TrimSpace(log.Resource) == "" {
		return apperrors.WrapInvalidInput("audit log resource is required")
	}
	switch log.ActorType {
	case model.AuditActorTypeStaff:
		if log.ActorID == nil || *log.ActorID == 0 {
			return apperrors.WrapInvalidInput("audit log actor_id is required for staff actor")
		}
	case model.AuditActorTypeSystem:
		if log.ActorID != nil {
			return apperrors.WrapInvalidInput("audit log actor_id must be empty for system actor")
		}
	default:
		return apperrors.WrapInvalidInput("audit log actor_type is invalid")
	}
	return nil
}

func normalizeIPAddress(ip string) *string {
	if ip == "" {
		return nil
	}
	return &ip
}

func normalizeAuditUserAgent(userAgent string) string {
	normalized := strings.ToValidUTF8(userAgent, "\uFFFD")
	if len(normalized) <= maxAuditUserAgentBytes {
		return normalized
	}
	end := maxAuditUserAgentBytes
	for end > 0 && !utf8.RuneStart(normalized[end]) {
		end--
	}
	return normalized[:end]
}

func buildLog(input *Entry) (*model.AuditLog, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("audit log input is required")
	}
	return &model.AuditLog{
		ClinicID:   input.ClinicID,
		ActorID:    input.ActorID,
		ActorType:  input.ActorType,
		Action:     input.Action,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		OldValue:   MarshalJSON(input.OldValue),
		NewValue:   MarshalJSON(input.NewValue),
		Metadata:   MarshalJSON(input.Metadata),
		IPAddress:  normalizeIPAddress(input.IPAddress),
		UserAgent:  normalizeAuditUserAgent(input.UserAgent),
	}, nil
}

func marshalOptionalMap(value map[string]any) []byte {
	if value == nil {
		return nil
	}
	return MarshalJSON(value)
}

// Log validates and writes a regular audit event.
func (s *service) Log(ctx context.Context, log *model.AuditLog) error {
	if err := ValidateLog(log); err != nil {
		return err
	}
	normalizedLog := *log
	normalizedLog.UserAgent = normalizeAuditUserAgent(log.UserAgent)
	if err := s.repository.Create(ctx, &normalizedLog); err != nil {
		slog.ErrorContext(ctx, "audit_write_failed",
			"action", normalizedLog.Action,
			"resource", normalizedLog.Resource,
			"clinic_id", *normalizedLog.ClinicID,
			"error", err,
		)
		return apperrors.Wrap(err, "failed to create audit log")
	}
	return nil
}

// LogEntry constructs and writes a regular audit event.
func (s *service) LogEntry(ctx context.Context, input *Entry) error {
	log, err := buildLog(input)
	if err != nil {
		return err
	}
	return s.Log(ctx, log)
}

// LogEntryTx writes through the ambient-transaction repository path.
func (s *service) LogEntryTx(ctx context.Context, input *Entry) error {
	log, err := buildLog(input)
	if err != nil {
		return err
	}
	if err := ValidateLog(log); err != nil {
		return err
	}
	log.UserAgent = normalizeAuditUserAgent(log.UserAgent)
	if err := s.repository.CreateTx(ctx, log); err != nil {
		return apperrors.Wrap(err, "failed to create audit log")
	}
	return nil
}

// LogAuthLogin records an authentication event.
func (s *service) LogAuthLogin(
	ctx context.Context,
	clinicID, staffID *uint64,
	action, ipAddress, userAgent string,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:  clinicID,
		ActorID:   staffID,
		ActorType: model.AuditActorTypeStaff,
		Action:    action,
		Resource:  "auth",
		IPAddress: normalizeIPAddress(ipAddress),
		UserAgent: normalizeAuditUserAgent(userAgent),
	})
}

// LogLstepOperation records an LSTEP operation without metadata.
func (s *service) LogLstepOperation(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action, resource string,
	resourceID *uint64,
) error {
	return s.LogLstepOperationWithMetadata(ctx, clinicID, actorID, action, resource, resourceID, nil)
}

// LogLstepOperationWithMetadata records an LSTEP operation.
func (s *service) LogLstepOperationWithMetadata(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action, resource string,
	resourceID *uint64,
	metadata any,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Metadata:   MarshalJSON(metadata),
	})
}

// LogMedicalRecordChange records a medical-record mutation.
func (s *service) LogMedicalRecordChange(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	recordID uint64,
	oldValue, newValue map[string]any,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     action,
		Resource:   "medical_record",
		ResourceID: &recordID,
		OldValue:   marshalOptionalMap(oldValue),
		NewValue:   marshalOptionalMap(newValue),
	})
}

// LogVitalChange records a vital-record mutation.
func (s *service) LogVitalChange(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	vitalID, medicalRecordID uint64,
	oldValue, newValue map[string]any,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     action,
		Resource:   "vital",
		ResourceID: &vitalID,
		OldValue:   marshalOptionalMap(oldValue),
		NewValue:   marshalOptionalMap(newValue),
		Metadata:   MarshalJSON(map[string]any{"medical_record_id": medicalRecordID}),
	})
}

// LogClinicSwitch records an active-clinic change.
func (s *service) LogClinicSwitch(
	ctx context.Context,
	actorID *uint64,
	fromClinicID, toClinicID uint64,
	ipAddress, userAgent string,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:  &toClinicID,
		ActorID:   actorID,
		ActorType: sharedkernel.AuditActorTypeFor(actorID),
		Action:    "switch_clinic",
		Resource:  "auth",
		OldValue:  MarshalJSON(map[string]any{"clinic_id": fromClinicID}),
		NewValue:  MarshalJSON(map[string]any{"clinic_id": toClinicID}),
		IPAddress: normalizeIPAddress(ipAddress),
		UserAgent: normalizeAuditUserAgent(userAgent),
	})
}

// LogAddendumCreate records a medical-record addendum.
func (s *service) LogAddendumCreate(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	addendumID, medicalRecordID uint64,
	addendum *model.MedicalRecordAddendum,
) error {
	return s.Log(ctx, &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     "create",
		Resource:   "medical_record_addendum",
		ResourceID: &addendumID,
		NewValue: MarshalJSON(map[string]any{
			"before_text":    addendum.BeforeText,
			"after_text":     addendum.AfterText,
			"reason":         addendum.Reason,
			"author_user_id": addendum.AuthorUserID,
		}),
		Metadata: MarshalJSON(map[string]any{"medical_record_id": medicalRecordID}),
	})
}
