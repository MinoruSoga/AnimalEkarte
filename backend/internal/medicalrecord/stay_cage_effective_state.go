package medicalrecord

import (
	"slices"

	"github.com/animal-ekarte/backend/internal/model"
)

// projectStayCageEffectiveState is the minimal EMR-499 lifecycle subset used by
// EMR-461. It interprets the entire ordered sequence, never only max(event_kind).
// Unknown cases are scoped to this stay; other stays can still be projected.
// Correction effects and no-event reservation identity are deliberately not
// inferred from the available storage columns.
func projectStayCageEffectiveState(stay model.StayCase, movements []model.StayMovementEvent) StayCageEffectiveState {
	result := StayCageEffectiveState{ClinicID: stay.ClinicID, StayCaseID: stay.ID, State: StayCageUnknown}
	if stay.ID == 0 || stay.ClinicID == 0 || stay.DeletedAt != nil || stay.Version <= 0 ||
		(stay.StayKind != "inpatient" && stay.StayKind != "hotel") || len(movements) == 0 {
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
		result.EvidenceVersion = event.Revision
		switch event.EventKind {
		case "admit":
			if i != 0 || event.ActualAt == nil || event.HospitalizationID == nil || *event.HospitalizationID == 0 ||
				event.AdmissionRecordRef == nil || (stay.StayKind == "inpatient" && event.DoctorID == nil) {
				return result
			}
			state, admitted = StayCageAdmitted, event
		case "discharge":
			if state != StayCageAdmitted || admitted == nil || event.ActualAt == nil || event.ActualAt.Before(*admitted.ActualAt) {
				return result
			}
			state = StayCageReleased
		case "cancel", "no_show":
			if i != 0 {
				return result
			}
			state = StayCageReleased
		default:
			// correction needs its after-effect/invalidation contract. Raw
			// correction, unsupported transitions and malformed histories do
			// not become vacant or inherit an older event speculatively.
			return result
		}
	}
	result.State = state
	return result
}
