package medicalrecord

import (
	"context"
	"slices"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-499 U1. hospitalizations.hospitalization_type / start_date / end_date / status no longer exist
// (001 final DROP). The hospitalization read wire is derived from the stay/movement spec, not from the
// retired columns:
//
//   - stay kind and the PLANNED interval come from stay_cases (kind, planned_entry_at, planned_exit_at);
//   - the ACTUAL admission / discharge times and the lifecycle come from stay_movement_events through the
//     EMR-499 core (projectStayCageEffectiveState).
//
// Spec basis: docs/spec/ontology/hospitalization.md (入院／宿泊予定 vs 滞在実績: "予定区間と実際の開始/終了は別。
// 未退室は予定超過後も滞在中", 実退院日時 is distinct from 予定退院), docs/spec/domains/hospitalization/lifecycle.md
// FR-hospitalization-entry-and-start ("予約だけ・予定日の到来だけで開始しない"), FR-hospitalization-hotel-stay-capacity,
// docs/design/domains/hospitalization/technical-spec.md ("予定終了と実退院を同一視しない"). Nothing here writes a
// lifecycle fact and nothing infers one; unresolvable stays are "unknown".

// Stable 422 codes (EMR-549 C1): the storage exists but the requested writer/reader is not built yet.
// They never mean success and never stub a result.
const (
	// codeStayLifecycleWriteUnavailable rejects writes of stay kind / planned interval / status through PATCH
	// (planned changes belong to extension proposals and acceptance proof, not to a direct overwrite).
	codeStayLifecycleWriteUnavailable = "stay_lifecycle_write_unavailable"
	// codeStayDischargeWriterUnavailable: the discharge movement writer (operation receipt + proof) is not built.
	codeStayDischargeWriterUnavailable = "stay_discharge_writer_unavailable"
	// codeStayDischargeBillingUnavailable: create_accounting=true; the billing builder is owned by EMR-505/506/508.
	codeStayDischargeBillingUnavailable = "stay_discharge_billing_unavailable"
	// codeStayLifecycleUnknown: the stay has no resolvable effective state (no stay case, no events, malformed history).
	codeStayLifecycleUnknown = "stay_lifecycle_unknown"
)

// Stay kinds exactly as stored in stay_cases.stay_kind (001 CHECK).
const (
	stayKindInpatient = "inpatient"
	stayKindHotel     = "hotel"
)

func validStayKind(kind string) bool { return kind == stayKindInpatient || kind == stayKindHotel }

// Lifecycle states exposed as `lifecycle_state`. "released" is split by cause: a cancelled or no-show
// stay must never be presented as a discharge.
const (
	lifecycleReserved   = "reserved"
	lifecycleAdmitted   = "admitted"
	lifecycleDischarged = "discharged"
	lifecycleCancelled  = "cancelled"
	lifecycleNoShow     = "no_show"
	lifecycleUnknown    = "unknown"
)

func validLifecycleState(v string) bool {
	switch v {
	case lifecycleReserved, lifecycleAdmitted, lifecycleDischarged, lifecycleCancelled, lifecycleNoShow, lifecycleUnknown:
		return true
	}
	return false
}

// HospitalizationStayProjection is the read-only wire view of one hospitalization's stay.
// A nil pointer means "not known", never a default.
type HospitalizationStayProjection struct {
	StayKind       *string
	PlannedEntryAt *time.Time
	PlannedExitAt  *time.Time
	Lifecycle      string
	AdmittedAt     *time.Time
	DischargedAt   *time.Time
}

func unknownHospitalizationStay() HospitalizationStayProjection {
	return HospitalizationStayProjection{Lifecycle: lifecycleUnknown}
}

// lifecycleOf converts the core verdict to the wire value. Anything the core cannot resolve is
// "unknown"; it is never folded into reserved.
func lifecycleOf(state StayCageEffectiveState) string {
	switch state.State {
	case StayCageReserved:
		return lifecycleReserved
	case StayCageAdmitted:
		return lifecycleAdmitted
	case StayCageReleased:
		switch state.ReleaseKind {
		case StayCageReleasedByDischarge:
			return lifecycleDischarged
		case StayCageReleasedByCancel:
			return lifecycleCancelled
		case StayCageReleasedByNoShow:
			return lifecycleNoShow
		}
	}
	return lifecycleUnknown
}

// StayProjections returns the stay projection for each given hospitalization (keyed by hospitalization ID).
// Rows without a readable stay case, or when no effective-state reader is wired, are "unknown";
// a database failure is returned as an error, never turned into "unknown".
func (s *hospitalizationService) StayProjections(ctx context.Context, clinicID uint64, rows []model.Hospitalization) (map[uint64]HospitalizationStayProjection, error) {
	result := make(map[uint64]HospitalizationStayProjection, len(rows))
	stayIDs := make([]uint64, 0, len(rows))
	for i := range rows {
		result[rows[i].ID] = unknownHospitalizationStay()
		if rows[i].ClinicID == clinicID && rows[i].StayCaseID != nil && *rows[i].StayCaseID != 0 {
			stayIDs = append(stayIDs, *rows[i].StayCaseID)
		}
	}
	if len(stayIDs) == 0 {
		return result, nil
	}
	slices.Sort(stayIDs)
	stayIDs = slices.Compact(stayIDs)

	var (
		stays  map[uint64]model.StayCase
		states []StayCageEffectiveState
	)
	read := func(readCtx context.Context) error {
		var err error
		stays, err = s.hospRepo.FindStayCasesByIDs(readCtx, clinicID, stayIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to read stay cases")
		}
		reader, ok := s.cageRepo.(StayCageEffectiveStateReader)
		if !ok {
			return nil // no reader wired: lifecycle stays unknown
		}
		states, err = reader.ReadEffectiveStatesForCageAssignment(readCtx, clinicID, stayIDs)
		if err != nil {
			return apperrors.Wrap(err, "failed to read stay effective states")
		}
		return nil
	}
	// One snapshot for stay row + movement history when the cage repository can provide it.
	if runner, ok := s.cageRepo.(cageOccupancySnapshotRunner); ok {
		if err := runner.WithCageOccupancyReadSnapshot(ctx, read); err != nil {
			return nil, err
		}
	} else if err := read(ctx); err != nil {
		return nil, err
	}
	stateByStay := make(map[uint64]StayCageEffectiveState, len(states))
	for _, st := range states {
		if st.ClinicID == clinicID {
			stateByStay[st.StayCaseID] = st
		}
	}
	for i := range rows {
		row := rows[i]
		if row.StayCaseID == nil {
			continue
		}
		view := unknownHospitalizationStay()
		if stay, ok := stays[*row.StayCaseID]; ok && stay.ClinicID == clinicID && stay.DeletedAt == nil && validStayKind(stay.StayKind) {
			kind := stay.StayKind
			view.StayKind = &kind
			view.PlannedEntryAt = stay.PlannedEntryAt
			view.PlannedExitAt = stay.PlannedExitAt
		}
		if st, ok := stateByStay[*row.StayCaseID]; ok {
			view.Lifecycle = lifecycleOf(st)
			view.AdmittedAt = st.AdmittedAt
			view.DischargedAt = st.DischargedAt
		}
		result[row.ID] = view
	}
	return result, nil
}

// validatePlannedStayInterval mirrors stay_cases CHECK (planned_exit_at > planned_entry_at) at the boundary
// so the caller gets a field error instead of a bare constraint failure. Either side may be absent: the
// planned interval is optional in the stored contract and nothing here invents a missing side.
func validatePlannedStayInterval(entry, exit *time.Time) error {
	if entry != nil && exit != nil && !exit.After(*entry) {
		return apperrors.WrapInvalidInput("planned_exit_at は planned_entry_at より後である必要があります")
	}
	return nil
}

// filterByLifecycleState keeps exactly the rows whose derived lifecycle equals state (unknown rows only
// match lifecycle_state=unknown). A row without a projection matches nothing and is never inferred into a
// state. The list tabs and the snapshot read share it so the row set and the total cannot diverge.
func filterByLifecycleState(rows []model.Hospitalization, projections map[uint64]HospitalizationStayProjection, state string) []model.Hospitalization {
	matched := make([]model.Hospitalization, 0, len(rows))
	for i := range rows {
		if projections[rows[i].ID].Lifecycle == state {
			matched = append(matched, rows[i])
		}
	}
	return matched
}

// listByLifecycleState serves the list tabs (予約 / 入院・預かり中 / 退院済, docs/spec/screens/07-hospitalization-list.md)
// from the DERIVED effective state, since no SQL column holds it any more. It reads the bounded candidate set,
// projects every row, keeps exactly the rows whose lifecycle equals the requested state (unknown rows only
// match lifecycle_state=unknown) and paginates in memory. Over the cap it refuses rather than truncating.
func (s *hospitalizationService) listByLifecycleState(ctx context.Context, clinicID uint64, petID, ownerID *uint64, state string, from, to *string, page, limit int) ([]model.Hospitalization, int64, error) {
	if !validLifecycleState(state) {
		return nil, 0, apperrors.WrapInvalidInput("lifecycle_state is invalid")
	}
	candidates, total, err := s.hospRepo.FindAll(ctx, clinicID, petID, ownerID, from, to, 1, hospitalizationSnapshotMaxRows)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list hospitalizations")
	}
	if total > hospitalizationSnapshotMaxRows {
		return nil, 0, apperrors.WrapInvalidInput("入院が10000件を超えています。条件を絞り込んでください")
	}
	projections, err := s.StayProjections(ctx, clinicID, candidates)
	if err != nil {
		return nil, 0, err
	}
	matched := filterByLifecycleState(candidates, projections, state)
	matchedTotal := int64(len(matched))
	// page は上限なしで渡されるため、(page-1)*limit を先に計算すると int が溢れて start が負になり
	// matched[start:end] で panic する。乗算の前に page-1 を len/limit と比較して範囲外を弾く。
	if page < 1 || limit < 1 || page-1 > len(matched)/limit {
		return []model.Hospitalization{}, matchedTotal, nil
	}
	start := (page - 1) * limit
	if start >= len(matched) {
		return []model.Hospitalization{}, matchedTotal, nil
	}
	end := min(start+limit, len(matched))
	return matched[start:end], matchedTotal, nil
}
