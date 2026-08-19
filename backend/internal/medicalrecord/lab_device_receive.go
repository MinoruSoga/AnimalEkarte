package medicalrecord

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

const (
	labDeviceDefaultWaitTTLSeconds = 1800
	labDeviceMinWaitTTLSeconds     = 60
	labDeviceMaxWaitTTLSeconds     = 86400
	labDeviceUnlinkedLimit         = 20
	labDeviceSavedLimit            = 10
	labDeviceReceivedLimit         = 200
	labDeviceReceivedLookbackDays  = 7
	labDeviceTodayVisitLimit       = 100
	labDeviceDefaultSlotsJSON      = `[{"key":"nx600","source_type":"fuji_nx600","device_hint":"NX600","baud":9600},{"key":"au10v","source_type":"fuji_au10v","device_hint":"AU10V","baud":9600},{"key":"pu4010","source_type":"arkray_pu4010","device_hint":"PU-4010","baud":9600}]`
)

// LabDevicePetFinder is the clinic-scoped pet read used to validate request pet_id.
type LabDevicePetFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// LabDeviceTodayVisitFinder lists today's in-progress charts for the wait board.
type LabDeviceTodayVisitFinder interface {
	ListTodayDraft(ctx context.Context, clinicID uint64, date string) ([]LabDeviceTodayVisit, error)
}

// LabDeviceExamPersister writes exams on attach and retracts them on detach.
type LabDeviceExamPersister interface {
	PersistLinkedJob(ctx context.Context, clinicID uint64, jobID uuid.UUID, petID uint64) error
	RetractLinkedJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) error
}

// LabDeviceReceiveService is the write-owner API for device wait / frames / attach.
type LabDeviceReceiveService interface {
	ReceiveFrames(ctx context.Context, clinicID uint64, payload []byte, hint string) (*LabDeviceReceiveResult, error)
	PutWait(ctx context.Context, clinicID, staffID, petID uint64) (*LabDeviceWaitView, error)
	ClearWait(ctx context.Context, clinicID uint64) error
	Board(ctx context.Context, clinicID uint64) (*LabDeviceBoard, error)
	Unlinked(ctx context.Context, clinicID uint64) ([]LabDeviceJobCard, error)
	Attach(ctx context.Context, clinicID uint64, jobID uuid.UUID, petID uint64) (*LabDeviceJobCard, error)
	Detach(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*LabDeviceJobCard, error)
	GetStation(ctx context.Context, clinicID uint64) (*LabDeviceStationView, error)
	PutStation(ctx context.Context, clinicID uint64, slotsJSON string) (*LabDeviceStationView, error)
}

// LabDeviceReceiveResult is one POST /frames response (one payload, one or more frames).
type LabDeviceReceiveResult struct {
	Results []LabDeviceFrameResult
}

// LabDeviceFrameResult is one decoded measurement after persist or fingerprint hit.
type LabDeviceFrameResult struct {
	Duplicate bool
	Job       LabDeviceJobCard
}

// LabDeviceBoard is the single receive page payload.
type LabDeviceBoard struct {
	Wait        *LabDeviceWaitView
	Unlinked    []LabDeviceJobCard
	Saved       []LabDeviceJobCard
	Received    []LabDeviceJobCard
	TodayVisits []LabDeviceTodayVisit
	Station     LabDeviceStationView
}

// LabDeviceTodayVisit is one selectable in-progress chart on the wait board.
type LabDeviceTodayVisit struct {
	RecordID      uint64
	PetID         uint64
	PetName       string
	OwnerName     string
	Species       string
	DoctorName    string
	VisitType     string
	PetIsDeceased bool
}

// LabDeviceWaitView is the active wait shown at page-max size.
type LabDeviceWaitView struct {
	PetID     uint64
	PetName   string
	StaffID   uint64
	ExpiresAt time.Time
}

// LabDeviceStationView is hospital-setup slots (TTL is not a product KPI).
type LabDeviceStationView struct {
	WaitTTLSeconds int
	SlotsJSON      string
}

// LabDeviceJobCard is a board / unlinked / attach card. Specimen ID is display-only.
type LabDeviceJobCard struct {
	JobID             uuid.UUID
	SourceType        model.LabImportSourceType
	DeviceHint        string
	Status            model.LabImportJobStatus
	PetID             *uint64
	PetName           string
	MeasuredAt        *time.Time
	ReceivedAt        *time.Time
	SpecimenIDRaw     string
	ItemCount         int
	UnmappedItemCount int
	Items             []LabDeviceJobItemView
}

// LabDeviceJobItemView is one decoded line on a card.
type LabDeviceJobItemView struct {
	DeviceItemCode  string
	ValueRaw        string
	Unit            string
	Flag            string
	ExamTypeFieldID *uint64
	NeedsReview     bool
	SortOrder       int
}
