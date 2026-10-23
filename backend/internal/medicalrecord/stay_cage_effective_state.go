package medicalrecord

import (
	"slices"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// projectStayCageEffectiveState is the minimal EMR-499 lifecycle subset used by
// EMR-461. It interprets the entire ordered sequence, never only max(event_kind).
// Unknown cases are scoped to this stay; other stays can still be projected.
// Correction effects and no-event reservation identity are deliberately not
// inferred from the available storage columns.
func projectStayCageEffectiveState(stay model.StayCase, movements []model.StayMovementEvent) StayCageEffectiveState {
	result := StayCageEffectiveState{
		ClinicID: stay.ClinicID, StayCaseID: stay.ID, State: StayCageUnknown,
		Evidence: StayCageEvidence{StayVersion: stay.Version, MovementHeadRevision: 0},
	}
	if stay.ID == 0 || stay.ClinicID == 0 || stay.DeletedAt != nil || stay.Version <= 0 ||
		(stay.StayKind != "inpatient" && stay.StayKind != "hotel") {
		return result
	}
	if len(movements) == 0 {
		// A valid stay has movement head 0 until its first event. Missing
		// acceptance proof leaves it unknown; it is not a malformed revision.
		return result
	}
	events := slices.Clone(movements)
	slices.SortFunc(events, func(a, b model.StayMovementEvent) int {
		if a.Revision < b.Revision {
			return -1
		}
		if a.Revision > b.Revision {
			return 1
		}
		return 0
	})
	seenIDs := make(map[uint64]struct{}, len(events))
	state := StayCageUnknown
	var releaseKind StayCageReleaseKind
	var dischargedAt *time.Time
	var admitted *model.StayMovementEvent
	var previousRevision int64
	for i := range events {
		event := &events[i]
		_, duplicateID := seenIDs[event.ID]
		if event.ID == 0 || duplicateID || event.ClinicID != stay.ClinicID || event.StayCaseID != stay.ID ||
			event.Revision <= previousRevision || event.CorrectedEventID != nil {
			return result
		}
		seenIDs[event.ID] = struct{}{}
		previousRevision = event.Revision
		// This is a history-evidence revision, not a CAS version and not a
		// rule that the last event kind alone defines the lifecycle.
		result.Evidence.MovementHeadRevision = event.Revision
		switch event.EventKind {
		case "admit":
			if i != 0 || event.ActualAt == nil || event.HospitalizationID == nil || *event.HospitalizationID == 0 {
				return result
			}
			if stay.StayKind == "inpatient" && (event.AdmissionRecordRef == nil || *event.AdmissionRecordRef <= 0 ||
				event.DoctorID == nil || *event.DoctorID == 0) {
				return result
			}
			state, admitted = StayCageAdmitted, event
			releaseKind = ""
		case "discharge":
			if state != StayCageAdmitted || admitted == nil || event.ActualAt == nil || event.ActualAt.Before(*admitted.ActualAt) {
				return result
			}
			state, releaseKind = StayCageReleased, StayCageReleasedByDischarge
			at := *event.ActualAt
			dischargedAt = &at
		case "cancel", "no_show":
			if i != 0 {
				return result
			}
			state = StayCageReleased
			releaseKind = StayCageReleasedByCancel
			if event.EventKind == "no_show" {
				releaseKind = StayCageReleasedByNoShow
			}
		default:
			// correction needs its after-effect/invalidation contract. Raw
			// correction, unsupported transitions and malformed histories do
			// not become vacant or inherit an older event speculatively.
			return result
		}
	}
	result.State = state
	if state == StayCageReleased {
		result.ReleaseKind = releaseKind
	}
	result.DischargedAt = dischargedAt
	if admitted != nil && admitted.ActualAt != nil {
		at := *admitted.ActualAt
		result.AdmittedAt = &at
	}
	return result
}
