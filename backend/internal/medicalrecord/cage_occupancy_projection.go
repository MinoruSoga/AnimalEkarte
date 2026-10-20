package medicalrecord

import (
	"context"
	"slices"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// cageOccupancyUnknown is an explicit unresolved verdict, distinct from a
// reader/database failure. Existing save guards remain fail-closed until the
// writer contract is integrated; the assignment read exposes it as 200 unknown.
type cageOccupancyUnknown struct{ reason string }

func (e *cageOccupancyUnknown) Error() string { return e.reason }
func (e *cageOccupancyUnknown) Unwrap() error { return apperrors.ErrServiceUnavailable }

// StayCageOccupancyState is the effective lifecycle decision supplied by EMR-499,
// not a raw stay_movement_events.event_kind. The cage adapter never interprets
// correction chains, planned dates, or the removed hospitalizations.status.
type StayCageOccupancyState string

const (
	StayCageReserved StayCageOccupancyState = "reserved"
	StayCageAdmitted StayCageOccupancyState = "admitted"
	StayCageReleased StayCageOccupancyState = "released"
	StayCageUnknown  StayCageOccupancyState = "unknown"
)

// StayCageEvidence keeps the stay version separate from the movement head.
// Head 0 is a valid empty history. Appointment proof and leaf expectations are
// to be connected by the follow-up 499 writer; this subset is not a write CAS.
type StayCageEvidence struct {
	StayVersion          int64
	MovementHeadRevision int64
}

// StayCageEffectiveState is the minimum 499 -> 461 handoff.
type StayCageEffectiveState struct {
	ClinicID   uint64
	StayCaseID uint64
	State      StayCageOccupancyState
	Evidence   StayCageEvidence
}

// StayCageEffectiveStateReader is provided by the one EMR-499 lifecycle core.
// It must return one current, clinic-scoped decision per requested stay case.
// Under an ambient transaction it must use that transaction and the agreed
// writer/reader lock protocol, so a lifecycle correction cannot pass the cage
// check and commit a conflicting occupancy. That protocol is an integration
// prerequisite; this interface does not invent an event interpretation/lock.
// Missing, unresolvable and legacy cases must be unknown, never released.
type StayCageEffectiveStateReader interface {
	ReadEffectiveStatesForCageAssignment(ctx context.Context, clinicID uint64, stayCaseIDs []uint64) ([]StayCageEffectiveState, error)
}

type stayCageAssignment struct {
	HospitalizationID uint64
	ClinicID          uint64
	StayCaseID        *uint64
	CageID            uint64
}

// countEffectiveStayCages is shared by candidate read and every save-time count.
// Self exclusion is by hospitalization identity, never by cage or patient.
func countEffectiveStayCages(ctx context.Context, clinicID uint64, assignments []stayCageAssignment, excludeHospitalizationIDs []uint64, reader StayCageEffectiveStateReader) (map[uint64]int64, error) {
	if clinicID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic is required")
	}
	if reader == nil {
		return nil, apperrors.WrapServiceUnavailable("stay effective-state reader is not configured")
	}
	relevant := make([]stayCageAssignment, 0, len(assignments))
	stayIDs := make([]uint64, 0, len(assignments))
	seenHospitalizations := make(map[uint64]struct{}, len(assignments))
	seenStays := make(map[uint64]struct{}, len(assignments))
	for _, row := range assignments {
		if row.ClinicID != clinicID || row.HospitalizationID == 0 || row.CageID == 0 {
			return nil, &cageOccupancyUnknown{reason: "cage assignment scope is unresolved"}
		}
		if slices.Contains(excludeHospitalizationIDs, row.HospitalizationID) {
			continue
		}
		if row.StayCaseID == nil || *row.StayCaseID == 0 {
			return nil, &cageOccupancyUnknown{reason: "cage assignment has no resolved stay"}
		}
		if _, exists := seenHospitalizations[row.HospitalizationID]; exists {
			return nil, &cageOccupancyUnknown{reason: "duplicate cage assignment identity"}
		}
		if _, exists := seenStays[*row.StayCaseID]; exists {
			return nil, &cageOccupancyUnknown{reason: "duplicate stay assignment identity"}
		}
		seenHospitalizations[row.HospitalizationID] = struct{}{}
		seenStays[*row.StayCaseID] = struct{}{}
		stayIDs = append(stayIDs, *row.StayCaseID)
		relevant = append(relevant, row)
	}
	counts := make(map[uint64]int64)
	if len(stayIDs) == 0 {
		return counts, nil
	}
	slices.Sort(stayIDs)
	states, err := reader.ReadEffectiveStatesForCageAssignment(ctx, clinicID, stayIDs)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read effective stay occupancy")
	}
	byStay := make(map[uint64]StayCageEffectiveState, len(states))
	for _, state := range states {
		_, requested := seenStays[state.StayCaseID]
		_, duplicate := byStay[state.StayCaseID]
		if state.ClinicID != clinicID || !requested || duplicate || state.Evidence.StayVersion <= 0 || state.Evidence.MovementHeadRevision < 0 {
			return nil, &cageOccupancyUnknown{reason: "stay occupancy evidence is unresolved"}
		}
		byStay[state.StayCaseID] = state
	}
	for _, row := range relevant {
		state, ok := byStay[*row.StayCaseID]
		if !ok {
			return nil, &cageOccupancyUnknown{reason: "stay occupancy evidence is missing"}
		}
		// Empty history is legitimate. A known lifecycle still needs its
		// movement or the accepted reservation proof that 499 has yet to connect.
		if state.State != StayCageUnknown && state.Evidence.MovementHeadRevision == 0 {
			return nil, &cageOccupancyUnknown{reason: "stay occupancy proof is not connected"}
		}
		switch state.State {
		case StayCageReserved, StayCageAdmitted:
			counts[row.CageID]++
		case StayCageReleased:
			// Only the core's explicit resolved decision releases occupancy.
		default:
			return nil, &cageOccupancyUnknown{reason: "stay occupancy is unknown"}
		}
	}
	return counts, nil
}
