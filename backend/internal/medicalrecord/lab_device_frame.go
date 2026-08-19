package medicalrecord

import (
	"errors"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// ErrInvalidLabDevicePayload is returned when a serial frame cannot be decoded.
var ErrInvalidLabDevicePayload = errors.New("invalid_payload")

// LabDeviceItem is one decoded instrument line. Values stay strings.
type LabDeviceItem struct {
	Code     string
	ValueRaw string
	Unit     string
	Flag     string
}

// LabDeviceFrame is one measurement (one fingerprint). Pet ID is empty.
type LabDeviceFrame struct {
	SourceType        model.LabImportSourceType
	SourceFingerprint string
	MeasuredAt        *time.Time
	SpecimenIDRaw     string
	DeviceHint        string
	Items             []LabDeviceItem
	Warnings          []string
	NeedsReview       bool
}
